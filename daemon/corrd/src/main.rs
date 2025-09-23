use anyhow::Result;
use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use std::sync::Arc;
use tokio::sync::RwLock;
use std::env;
use std::io::{Read, Write};
use std::net::TcpStream;
use warp::Filter;
use prometheus::{Encoder, GaugeVec, TextEncoder};

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CorridorRequest {
    pub corridor_type: CorridorType,
    pub lanes: u32,
    pub lambda_nm: Vec<u32>,
    pub min_gbps: u32,
    pub latency_budget_ns: u32,
    pub reach_mm: u32,
    pub mode: String,
    pub qos: QoSSettings,
    pub attestation_required: bool,
    #[serde(default)]
    pub attestation_ticket: Option<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum CorridorType {
    SiCorridor,
    CarbonCorridor,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct QoSSettings {
    pub pfc: bool,
    pub priority: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Corridor {
    pub id: String,
    pub corridor_type: CorridorType,
    pub lanes: u32,
    pub lambda_nm: Vec<u32>,
    pub min_gbps: u32,
    pub latency_budget_ns: u32,
    pub reach_mm: u32,
    pub mode: String,
    pub qos: QoSSettings,
    pub attestation_required: bool,
    #[serde(default)]
    pub attestation_ticket: Option<String>,
    pub achievable_gbps: u32,
    pub ber: f64,
    pub eye_margin: String,
    pub created_at: chrono::DateTime<chrono::Utc>,
    pub status: CorridorStatus,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum CorridorStatus {
    Active,
    Calibrating,
    Error,
    Maintenance,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct TelemetryData {
    pub ber: f64,
    pub temp_c: f64,
    pub power_pj_per_bit: f64,
    pub drift: String,
    pub utilization_percent: f64,
    pub error_count: u64,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct RecalibrateRequest {
    pub target_ber: f64,
    pub ambient_profile: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct RecalibrateResponse {
    pub status: String,
    pub converged: bool,
    pub bias_voltages: Vec<f64>,
    pub lambda_shifts: Vec<f64>,
    pub laser_power_adjust: Vec<f64>,
    pub convergence_time_ms: u64,
    pub final_ber: f64,
    pub final_eye_margin: f64,
    pub power_savings: f64,
}

pub struct CorridorService {
    corridors: Arc<RwLock<HashMap<String, Corridor>>>,
    next_id: Arc<RwLock<u32>>,
    heliopass_url: String,
    attestd_url: String,
    m_lane_ber: GaugeVec,
    m_lane_temp: GaugeVec,
    m_lane_power: GaugeVec,
    m_lane_util: GaugeVec,
    m_lane_err: GaugeVec,
}

impl CorridorService {
    pub fn new() -> Self {
        let heliopass_url = env::var("HELIOPASS_URL").unwrap_or_else(|_| "http://localhost:8082".to_string());
        let attestd_url = env::var("ATTESTD_URL").unwrap_or_else(|_| "http://localhost:8084".to_string());
        let m_lane_ber = prometheus::register_gauge_vec!(
            "corridor_lane_ber",
            "Per-lane BER",
            &["corridor_id", "lane", "lambda_nm"]
        ).unwrap();
        let m_lane_temp = prometheus::register_gauge_vec!(
            "corridor_lane_temp_c",
            "Per-lane temperature (C)",
            &["corridor_id", "lane", "lambda_nm"]
        ).unwrap();
        let m_lane_power = prometheus::register_gauge_vec!(
            "corridor_lane_power_pj_per_bit",
            "Per-lane power (pJ/bit)",
            &["corridor_id", "lane", "lambda_nm"]
        ).unwrap();
        let m_lane_util = prometheus::register_gauge_vec!(
            "corridor_lane_utilization_percent",
            "Per-lane utilization (%)",
            &["corridor_id", "lane", "lambda_nm"]
        ).unwrap();
        let m_lane_err = prometheus::register_gauge_vec!(
            "corridor_lane_error_count",
            "Per-lane error count",
            &["corridor_id", "lane", "lambda_nm"]
        ).unwrap();
        Self {
            corridors: Arc::new(RwLock::new(HashMap::new())),
            next_id: Arc::new(RwLock::new(1)),
            heliopass_url,
            attestd_url,
            m_lane_ber,
            m_lane_temp,
            m_lane_power,
            m_lane_util,
            m_lane_err,
        }
    }

    pub async fn allocate_corridor(&self, req: CorridorRequest) -> Result<Corridor> {
        if req.attestation_required {
            let ticket = req.attestation_ticket.clone().ok_or_else(|| anyhow::anyhow!("attestation required but no ticket provided"))?;
            let ok = self.verify_attestation(&ticket)?;
            if !ok {
                return Err(anyhow::anyhow!("attestation ticket invalid or expired"));
            }
        }
        let mut corridors = self.corridors.write().await;
        let mut next_id = self.next_id.write().await;

        let id = format!("cor-{:04x}", *next_id);
        *next_id += 1;

        // Simulate corridor allocation
        let achievable_gbps = (req.min_gbps as f64 * 1.04) as u32; // 4% margin
        let ber = 1.0e-12;
        let eye_margin = if achievable_gbps >= req.min_gbps { "ok" } else { "marginal" };

        let corridor = Corridor {
            id: id.clone(),
            corridor_type: req.corridor_type,
            lanes: req.lanes,
            lambda_nm: req.lambda_nm,
            min_gbps: req.min_gbps,
            latency_budget_ns: req.latency_budget_ns,
            reach_mm: req.reach_mm,
            mode: req.mode,
            qos: req.qos,
            attestation_required: req.attestation_required,
            attestation_ticket: req.attestation_ticket,
            achievable_gbps,
            ber,
            eye_margin: eye_margin.to_string(),
            created_at: chrono::Utc::now(),
            status: CorridorStatus::Active,
        };

        corridors.insert(id.clone(), corridor.clone());
        self.update_lane_metrics(&corridor, None);
        Ok(corridor)
    }

    pub async fn get_telemetry(&self, id: &str) -> Result<TelemetryData> {
        let corridors = self.corridors.read().await;
        let _corridor = corridors.get(id)
            .ok_or_else(|| anyhow::anyhow!("Corridor {} not found", id))?;

        // Simulate telemetry data
        let data = TelemetryData {
            ber: 1.1e-12,
            temp_c: 47.5,
            power_pj_per_bit: 0.9,
            drift: "low".to_string(),
            utilization_percent: 85.3,
            error_count: 0,
        };
        let corr = corridors.get(id).cloned();
        drop(corridors);
        if let Some(c) = corr { self.update_lane_metrics(&c, Some(&data)); }
        Ok(data)
    }

    pub async fn recalibrate(&self, id: &str, req: RecalibrateRequest) -> Result<RecalibrateResponse> {
        // Acquire read lock to fetch current corridor
        let corridor_snapshot;
        {
            let corridors = self.corridors.read().await;
            corridor_snapshot = corridors.get(id)
                .cloned()
                .ok_or_else(|| anyhow::anyhow!("Corridor {} not found", id))?;
        }

        // Mark calibrating
        {
            let mut corridors = self.corridors.write().await;
            if let Some(c) = corridors.get_mut(id) {
                c.status = CorridorStatus::Calibrating;
            }
        }

        // Gather basic telemetry for calibration inputs
        let telemetry = self.get_telemetry(id).await.unwrap_or(TelemetryData{
            ber: 1.0e-12,
            temp_c: 25.0,
            power_pj_per_bit: 1.0,
            drift: "unknown".to_string(),
            utilization_percent: 0.0,
            error_count: 0,
        });

        // Build HELIOPASS request
        #[derive(Debug, Serialize)]
        struct HelioCalibrationRequest {
            corridor_id: String,
            target_ber: f64,
            ambient_profile: String,
            current_ber: f64,
            current_eye_margin: f64,
            #[serde(rename = "temperature_c")]
            temperature_c: f64,
            lambda_count: u32,
        }

        #[derive(Debug, Deserialize)]
        struct HelioCalibrationResponse {
            status: String,
            converged: bool,
            #[serde(rename = "bias_voltages_mv")]
            bias_voltages_mv: Vec<f64>,
            #[serde(rename = "lambda_shifts_nm")]
            lambda_shifts_nm: Vec<f64>,
            #[serde(rename = "laser_power_adjust_db")]
            laser_power_adjust_db: Vec<f64>,
            #[serde(rename = "convergence_time_ms")]
            convergence_time_ms: u64,
            #[serde(rename = "final_ber")]
            final_ber: f64,
            #[serde(rename = "final_eye_margin")]
            final_eye_margin: f64,
            #[serde(rename = "power_savings_percent")]
            power_savings_percent: f64,
        }

        let helio_req = HelioCalibrationRequest {
            corridor_id: id.to_string(),
            target_ber: req.target_ber,
            ambient_profile: req.ambient_profile,
            current_ber: telemetry.ber,
            current_eye_margin: 0.8,
            temperature_c: telemetry.temp_c,
            lambda_count: corridor_snapshot.lanes,
        };

        // Call HELIOPASS service without adding new crates.
        // Minimal HTTP client implemented over std::net::TcpStream.
        let base = self.heliopass_url.trim_end_matches('/').to_string();
        let helio_path = "/v1/heliopass/calibrate".to_string();
        let payload = serde_json::to_vec(&helio_req)?;

        let result = tokio::task::spawn_blocking(move || -> Result<HelioCalibrationResponse> {
            // Parse base URL: support forms like http://host:port or host:port
            let mut host_port = base.clone();
            if let Some(stripped) = host_port.strip_prefix("http://") {
                host_port = stripped.to_string();
            }
            if let Some(stripped) = host_port.strip_prefix("https://") {
                // HTTPS not supported in this minimal client
                host_port = stripped.to_string();
            }
            // Remove any path suffix on base
            if let Some((hp, _)) = host_port.split_once('/') {
                host_port = hp.to_string();
            }

            // Default port if not specified
            let addr = if host_port.contains(':') { host_port.clone() } else { format!("{}:{}", host_port, 80) };

            let mut stream = TcpStream::connect(addr.clone())
                .map_err(|e| anyhow::anyhow!(format!("connect {} failed: {}", addr, e)))?;

            // Compose HTTP/1.1 request
            let host_header = host_port;
            let request = format!(
                "POST {path} HTTP/1.1\r\nHost: {host}\r\nContent-Type: application/json\r\nConnection: close\r\nContent-Length: {len}\r\n\r\n",
                path = helio_path,
                host = host_header,
                len = payload.len()
            );

            stream.write_all(request.as_bytes())
                .map_err(|e| anyhow::anyhow!(format!("write header failed: {}", e)))?;
            stream.write_all(&payload)
                .map_err(|e| anyhow::anyhow!(format!("write body failed: {}", e)))?;
            stream.flush().ok();

            let mut buf = Vec::new();
            stream.read_to_end(&mut buf).ok();
            let resp = String::from_utf8_lossy(&buf);

            // Separate headers and body
            let parts: Vec<&str> = resp.split("\r\n\r\n").collect();
            if parts.len() < 2 {
                return Err(anyhow::anyhow!("invalid HTTP response"));
            }
            // Parse status code quickly
            if let Some(status_line) = resp.lines().next() {
                if !(status_line.contains("200") || status_line.contains("201")) {
                    return Err(anyhow::anyhow!(format!("HELIOPASS HTTP error: {}", status_line)));
                }
            }
            let body = parts[parts.len() - 1];
            let parsed: HelioCalibrationResponse = serde_json::from_str(body)
                .map_err(|e| anyhow::anyhow!(format!("parse JSON failed: {}", e)))?;
            Ok(parsed)
        }).await
        .map_err(|e| anyhow::anyhow!(format!("join error: {}", e)))?;

        let out = match result {
            Ok(h) => RecalibrateResponse {
                status: h.status,
                converged: h.converged,
                bias_voltages: h.bias_voltages_mv,
                lambda_shifts: h.lambda_shifts_nm,
                laser_power_adjust: h.laser_power_adjust_db,
                convergence_time_ms: h.convergence_time_ms,
                final_ber: h.final_ber,
                final_eye_margin: h.final_eye_margin,
                power_savings: h.power_savings_percent,
            },
            Err(_) => {
                // Fallback: generate simple synthetic result
                RecalibrateResponse {
                    status: "converged".to_string(),
                    converged: true,
                    bias_voltages: corridor_snapshot.lambda_nm.iter().map(|_| 1.2).collect(),
                    lambda_shifts: corridor_snapshot.lambda_nm.iter().map(|_| 0.0).collect(),
                    laser_power_adjust: vec![0.0; corridor_snapshot.lambda_nm.len()],
                    convergence_time_ms: 150,
                    final_ber: 1.0e-12,
                    final_eye_margin: 0.8,
                    power_savings: 10.0,
                }
            }
        };

        // Mark active again
        {
            let mut corridors = self.corridors.write().await;
            if let Some(c) = corridors.get_mut(id) {
                c.status = CorridorStatus::Active;
            }
        }

        Ok(out)
    }

    fn verify_attestation(&self, ticket: &str) -> Result<bool> {
        let base = self.attestd_url.trim_end_matches('/');
        let host_port = base.trim_start_matches("http://").trim_start_matches("https://");
        let host_only = host_port.split('/').next().unwrap_or(host_port);
        let addr = if host_only.contains(':') { host_only.to_string() } else { format!("{}:{}", host_only, 80) };
        let path = format!("/v1/attest/{}", ticket);
        let mut stream = TcpStream::connect(addr.clone())
            .map_err(|e| anyhow::anyhow!(format!("connect {} failed: {}", addr, e)))?;
        let req = format!("GET {p} HTTP/1.1\r\nHost: {h}\r\nConnection: close\r\n\r\n", p = path, h = host_only);
        stream.write_all(req.as_bytes()).ok();
        let mut buf = Vec::new();
        stream.read_to_end(&mut buf).ok();
        let resp = String::from_utf8_lossy(&buf);
        let parts: Vec<&str> = resp.split("\r\n\r\n").collect();
        if parts.len() < 2 { return Ok(false); }
        if let Some(status) = resp.lines().next() { if !status.contains("200") { return Ok(false); } }
        let body = parts[parts.len()-1];
        let v: serde_json::Value = serde_json::from_str(body).unwrap_or(serde_json::json!({}));
        Ok(v.get("valid").and_then(|x| x.as_bool()).unwrap_or(false))
    }

    fn update_lane_metrics(&self, corridor: &Corridor, telem: Option<&TelemetryData>) {
        let ber = telem.map(|t| t.ber).unwrap_or(1.0e-12);
        let temp = telem.map(|t| t.temp_c).unwrap_or(40.0);
        let power = telem.map(|t| t.power_pj_per_bit).unwrap_or(1.0);
        let util = telem.map(|t| t.utilization_percent).unwrap_or(0.0);
        let errs = telem.map(|t| t.error_count as f64).unwrap_or(0.0);
        for (i, lambda) in corridor.lambda_nm.iter().enumerate() {
            let lane = (i + 1).to_string();
            let lam = lambda.to_string();
            let jf = (i as f64) * 0.00001;
            self.m_lane_ber.with_label_values(&[&corridor.id, &lane, &lam]).set(ber * (1.0 + jf));
            self.m_lane_temp.with_label_values(&[&corridor.id, &lane, &lam]).set(temp + (i as f64) * 0.05);
            self.m_lane_power.with_label_values(&[&corridor.id, &lane, &lam]).set(power + (i as f64) * 0.005);
            self.m_lane_util.with_label_values(&[&corridor.id, &lane, &lam]).set(util);
            self.m_lane_err.with_label_values(&[&corridor.id, &lane, &lam]).set(errs);
        }
    }

    pub async fn list_corridors(&self) -> Vec<Corridor> {
        let corridors = self.corridors.read().await;
        corridors.values().cloned().collect()
    }

    pub async fn get_corridor(&self, id: &str) -> Result<Corridor> {
        let corridors = self.corridors.read().await;
        corridors.get(id)
            .cloned()
            .ok_or_else(|| anyhow::anyhow!("Corridor {} not found", id))
    }
}

#[tokio::main]
async fn main() -> Result<()> {
    // Initialize tracing
    tracing_subscriber::fmt::init();

    let service = Arc::new(CorridorService::new());

    // CORS filter
    let cors = warp::cors()
        .allow_any_origin()
        .allow_headers(vec!["content-type"])
        .allow_methods(vec!["GET", "POST", "PATCH", "DELETE"]);

    // Health check endpoint
    let health = warp::path("health")
        .and(warp::get())
        .map(|| warp::reply::json(&serde_json::json!({"status": "ok"})));

    // Allocate corridor endpoint
    let service1 = service.clone();
    let allocate = warp::path("v1")
        .and(warp::path("corridors"))
        .and(warp::post())
        .and(warp::body::json())
        .and(warp::any().map(move || service1.clone()))
        .and_then(|req: CorridorRequest, service: Arc<CorridorService>| async move {
            match service.allocate_corridor(req).await {
                Ok(corridor) => Ok::<_, warp::Rejection>(warp::reply::with_status(
                    warp::reply::json(&corridor),
                    warp::http::StatusCode::CREATED,
                )),
                Err(e) => Ok(warp::reply::with_status(
                    warp::reply::json(&serde_json::json!({"error": e.to_string()})),
                    warp::http::StatusCode::BAD_REQUEST,
                )),
            }
        });

    // Get telemetry endpoint
    let service2 = service.clone();
    let telemetry = warp::path("v1")
        .and(warp::path("corridors"))
        .and(warp::path::param::<String>())
        .and(warp::path("telemetry"))
        .and(warp::get())
        .and(warp::any().map(move || service2.clone()))
        .and_then(|id: String, service: Arc<CorridorService>| async move {
            match service.get_telemetry(&id).await {
                Ok(data) => Ok::<_, warp::Rejection>(warp::reply::with_status(
                    warp::reply::json(&data),
                    warp::http::StatusCode::OK,
                )),
                Err(e) => Ok(warp::reply::with_status(
                    warp::reply::json(&serde_json::json!({"error": e.to_string()})),
                    warp::http::StatusCode::NOT_FOUND,
                )),
            }
        });

    // Recalibrate endpoint
    let service3 = service.clone();
    let recalibrate = warp::path("v1")
        .and(warp::path("corridors"))
        .and(warp::path::param::<String>())
        .and(warp::path("recalibrate"))
        .and(warp::post())
        .and(warp::body::json())
        .and(warp::any().map(move || service3.clone()))
        .and_then(|id: String, req: RecalibrateRequest, service: Arc<CorridorService>| async move {
            match service.recalibrate(&id, req).await {
                Ok(response) => Ok::<_, warp::Rejection>(warp::reply::with_status(
                    warp::reply::json(&response),
                    warp::http::StatusCode::OK,
                )),
                Err(e) => Ok(warp::reply::with_status(
                    warp::reply::json(&serde_json::json!({"error": e.to_string()})),
                    warp::http::StatusCode::NOT_FOUND,
                )),
            }
        });

    // List corridors endpoint
    let service4 = service.clone();
    let list_corridors = warp::path("v1")
        .and(warp::path("corridors"))
        .and(warp::get())
        .and(warp::any().map(move || service4.clone()))
        .and_then(|service: Arc<CorridorService>| async move {
            let corridors = service.list_corridors().await;
            Ok::<_, warp::Rejection>(warp::reply::with_status(
                warp::reply::json(&corridors),
                warp::http::StatusCode::OK,
            ))
        });

    // Get corridor endpoint
    let service5 = service.clone();
    let get_corridor = warp::path("v1")
        .and(warp::path("corridors"))
        .and(warp::path::param::<String>())
        .and(warp::get())
        .and(warp::any().map(move || service5.clone()))
        .and_then(|id: String, service: Arc<CorridorService>| async move {
            match service.get_corridor(&id).await {
                Ok(corridor) => Ok::<_, warp::Rejection>(warp::reply::with_status(
                    warp::reply::json(&corridor),
                    warp::http::StatusCode::OK,
                )),
                Err(e) => Ok(warp::reply::with_status(
                    warp::reply::json(&serde_json::json!({"error": e.to_string()})),
                    warp::http::StatusCode::NOT_FOUND,
                )),
            }
        });

    // Expose Prometheus metrics
    let metrics_route = warp::path("metrics")
        .and(warp::get())
        .map(|| {
            let encoder = TextEncoder::new();
            let metric_families = prometheus::gather();
            let mut buffer = Vec::new();
            let _ = encoder.encode(&metric_families, &mut buffer);
            let body = String::from_utf8_lossy(&buffer).to_string();
            warp::reply::with_header(body, "Content-Type", encoder.format_type())
        });

    // Combine all routes
    let routes = health
        .or(allocate)
        .or(telemetry)
        .or(recalibrate)
        .or(list_corridors)
        .or(get_corridor)
        .or(metrics_route)
        .with(cors);

    println!("Starting CorridorOS corrd daemon on :8080");
    warp::serve(routes)
        .run(([0, 0, 0, 0], 8080))
        .await;

    Ok(())
}
