use serde::{Deserialize, Serialize};

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct QoSConfig { pub pfc: bool, pub priority: String }

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CorridorAllocateRequest {
    pub corridor_type: String,
    pub lanes: u32,
    pub lambda_nm: Vec<u32>,
    pub min_gbps: u32,
    pub latency_budget_ns: u32,
    pub reach_mm: u32,
    pub mode: String,
    pub qos: QoSConfig,
    pub attestation_required: bool,
    pub attestation_ticket: Option<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Corridor { pub id: String, pub status: String }

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct FfmAllocateRequest {
    pub bytes: u64,
    pub latency_class: String,
    pub bandwidth_floor_GBs: u64,
    pub persistence: String,
    pub shareable: bool,
    pub security_domain: String,
    pub attestation_required: Option<bool>,
    pub attestation_ticket: Option<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct FfmHandle { pub id: String, pub bytes: u64 }

#[derive(Debug, Clone)]
pub struct Client { pub base_url: String }

impl Client {
    pub fn new(base: impl Into<String>) -> Self { Self{ base_url: base.into() } }
    pub fn allocate_corridor(&self, _r: &CorridorAllocateRequest) -> Result<Corridor, String> {
        Err("not implemented in minimal offline SDK".to_string())
    }
    pub fn allocate_ffm(&self, _r: &FfmAllocateRequest) -> Result<FfmHandle, String> {
        Err("not implemented in minimal offline SDK".to_string())
    }
}

