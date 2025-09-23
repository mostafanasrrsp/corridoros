package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// CorridorRequest represents a corridor allocation request
type CorridorRequest struct {
	CorridorType      string    `json:"corridor_type"`
	Lanes             int       `json:"lanes"`
	LambdaNm          []int     `json:"lambda_nm"`
	MinGbps           int       `json:"min_gbps"`
	LatencyBudgetNs   int       `json:"latency_budget_ns"`
	ReachMm           int       `json:"reach_mm"`
	Mode              string    `json:"mode"`
	QoS               QoSConfig `json:"qos"`
	AttestationRequired bool    `json:"attestation_required"`
}

// QoSConfig represents QoS settings
type QoSConfig struct {
	PFC      bool   `json:"pfc"`
	Priority string `json:"priority"`
}

// CorridorResponse represents a corridor allocation response
type CorridorResponse struct {
	ID              string    `json:"id"`
	CorridorType    string    `json:"corridor_type"`
	Lanes           int       `json:"lanes"`
	LambdaNm        []int     `json:"lambda_nm"`
	MinGbps         int       `json:"min_gbps"`
	LatencyBudgetNs int       `json:"latency_budget_ns"`
	ReachMm         int       `json:"reach_mm"`
	Mode            string    `json:"mode"`
	QoS             QoSConfig `json:"qos"`
	AttestationRequired bool  `json:"attestation_required"`
	AchievableGbps  int       `json:"achievable_gbps"`
	BER             float64   `json:"ber"`
	EyeMargin       string    `json:"eye_margin"`
	CreatedAt       time.Time `json:"created_at"`
	Status          string    `json:"status"`
}

// TelemetryData represents corridor telemetry
type TelemetryData struct {
	BER                float64 `json:"ber"`
	TempC              float64 `json:"temp_c"`
	PowerPjPerBit      float64 `json:"power_pj_per_bit"`
	Drift              string  `json:"drift"`
	UtilizationPercent float64 `json:"utilization_percent"`
	ErrorCount         int     `json:"error_count"`
}

// RecalibrateRequest represents a recalibration request
type RecalibrateRequest struct {
	TargetBER      float64 `json:"target_ber"`
	AmbientProfile string  `json:"ambient_profile"`
}

// RecalibrateResponse represents recalibration response
type RecalibrateResponse struct {
	Status            string    `json:"status"`
	Converged         bool      `json:"converged"`
	BiasVoltages      []float64 `json:"bias_voltages_mv"`
	LambdaShifts      []float64 `json:"lambda_shifts_nm"`
	LaserPowerAdjust  []float64 `json:"laser_power_adjust_db"`
	ConvergenceTimeMs int64     `json:"convergence_time_ms"`
	FinalBER          float64   `json:"final_ber"`
	FinalEyeMargin    float64   `json:"final_eye_margin"`
	PowerSavings      float64   `json:"power_savings_percent"`
}

const corrdURL = "http://localhost:8080"

func main() {
	fmt.Println("CorridorOS Photonic Corridor Demo")
	fmt.Println("=================================")

	// Test 1: Allocate different types of corridors
	fmt.Println("\n1. Allocating different types of photonic corridors...")
	
	corridorRequests := []CorridorRequest{
		{
			CorridorType:      "SiCorridor",
			Lanes:             8,
			LambdaNm:          []int{1550, 1551, 1552, 1553, 1554, 1555, 1556, 1557},
			MinGbps:           400,
			LatencyBudgetNs:   250,
			ReachMm:           75,
			Mode:              "waveguide",
			QoS: QoSConfig{
				PFC:      true,
				Priority: "gold",
			},
			AttestationRequired: true,
		},
		{
			CorridorType:      "CarbonCorridor",
			Lanes:             4,
			LambdaNm:          []int{1310, 1311, 1312, 1313},
			MinGbps:           200,
			LatencyBudgetNs:   500,
			ReachMm:           50,
			Mode:              "free-space",
			QoS: QoSConfig{
				PFC:      false,
				Priority: "silver",
			},
			AttestationRequired: false,
		},
		{
			CorridorType:      "SiCorridor",
			Lanes:             16,
			LambdaNm:          []int{1530, 1531, 1532, 1533, 1534, 1535, 1536, 1537, 1538, 1539, 1540, 1541, 1542, 1543, 1544, 1545},
			MinGbps:           800,
			LatencyBudgetNs:   100,
			ReachMm:           100,
			Mode:              "waveguide",
			QoS: QoSConfig{
				PFC:      true,
				Priority: "gold",
			},
			AttestationRequired: true,
		},
	}

	var corridors []CorridorResponse
	for i, req := range corridorRequests {
		corridor, err := allocateCorridor(req)
		if err != nil {
			log.Printf("Error allocating corridor %d: %v", i+1, err)
			continue
		}
		corridors = append(corridors, *corridor)
		fmt.Printf("  %s: %d lanes, %d Gbps (ID: %s)\n", 
			req.CorridorType, req.Lanes, corridor.AchievableGbps, corridor.ID)
	}

	// Test 2: Monitor telemetry
	fmt.Println("\n2. Monitoring corridor telemetry for 20 seconds...")
	
	for i := 0; i < 20; i++ {
		fmt.Printf("\n--- Telemetry Update %d/20 ---\n", i+1)
		
		for _, corridor := range corridors {
			telemetry, err := getTelemetry(corridor.ID)
			if err != nil {
				log.Printf("Error getting telemetry for %s: %v", corridor.ID, err)
				continue
			}
			
			fmt.Printf("  %s: BER=%.2e | Temp=%.1f°C | Power=%.2f pJ/bit | Util=%.1f%% | Errors=%d\n",
				corridor.ID,
				telemetry.BER,
				telemetry.TempC,
				telemetry.PowerPjPerBit,
				telemetry.UtilizationPercent,
				telemetry.ErrorCount)
		}
		
		time.Sleep(1 * time.Second)
	}

	// Test 3: Calibration
	fmt.Println("\n3. Testing HELIOPASS calibration...")
	
	if len(corridors) > 0 {
		corridor := corridors[0]
		fmt.Printf("Calibrating corridor %s with target BER 1e-12...\n", corridor.ID)
		
		req := RecalibrateRequest{
			TargetBER:      1e-12,
			AmbientProfile: "lab_default",
		}
		
		resp, err := calibrateCorridor(corridor.ID, req)
		if err != nil {
			log.Printf("Error calibrating corridor: %v", err)
		} else {
			fmt.Printf("Calibration completed: %s (converged: %t)\n", resp.Status, resp.Converged)
			fmt.Printf("  Convergence time: %d ms\n", resp.ConvergenceTimeMs)
			fmt.Printf("  Final BER: %.2e\n", resp.FinalBER)
			fmt.Printf("  Final eye margin: %.2f UI\n", resp.FinalEyeMargin)
			fmt.Printf("  Power savings: %.1f%%\n", resp.PowerSavings)
			fmt.Printf("  Bias voltages: %v mV\n", resp.BiasVoltages)
			fmt.Printf("  Lambda shifts: %v nm\n", resp.LambdaShifts)
		}
	}

	// Test 4: List all corridors
	fmt.Println("\n4. Current corridors:")
	
	allCorridors, err := listCorridors()
	if err != nil {
		log.Printf("Error listing corridors: %v", err)
	} else {
		fmt.Printf("%-12s %-12s %-6s %-15s %-8s %-10s %-8s\n", 
			"ID", "Type", "Lanes", "Wavelengths", "Gbps", "Status", "Created")
		fmt.Println("----------------------------------------------------------------")
		
		for _, corridor := range allCorridors {
			lambdaStr := fmt.Sprintf("%d-%d", corridor.LambdaNm[0], corridor.LambdaNm[len(corridor.LambdaNm)-1])
			fmt.Printf("%-12s %-12s %-6d %-15s %-8d %-10s %-8s\n",
				corridor.ID,
				corridor.CorridorType,
				corridor.Lanes,
				lambdaStr,
				corridor.AchievableGbps,
				corridor.Status,
				corridor.CreatedAt.Format("15:04:05"))
		}
	}

	// Test 5: Performance comparison
	fmt.Println("\n5. Performance comparison:")
	
	for _, corridor := range corridors {
		telemetry, err := getTelemetry(corridor.ID)
		if err != nil {
			continue
		}
		
		efficiency := float64(corridor.AchievableGbps) / float64(corridor.Lanes) / 100.0 // Gbps per lane
		powerEfficiency := telemetry.PowerPjPerBit
		
		fmt.Printf("  %s: %.1f Gbps/lane, %.2f pJ/bit, %.2e BER\n",
			corridor.CorridorType,
			efficiency,
			powerEfficiency,
			telemetry.BER)
	}

	fmt.Println("\nDemo completed!")
}

// HTTP client functions
func allocateCorridor(req CorridorRequest) (*CorridorResponse, error) {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(corrdURL+"/v1/corridors", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	var corridor CorridorResponse
	err = json.Unmarshal(body, &corridor)
	return &corridor, err
}

func getTelemetry(id string) (*TelemetryData, error) {
	resp, err := http.Get(corrdURL + "/v1/corridors/" + id + "/telemetry")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	var telemetry TelemetryData
	err = json.Unmarshal(body, &telemetry)
	return &telemetry, err
}

func calibrateCorridor(id string, req RecalibrateRequest) (*RecalibrateResponse, error) {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(corrdURL+"/v1/corridors/"+id+"/recalibrate", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	var recalResp RecalibrateResponse
	err = json.Unmarshal(body, &recalResp)
	return &recalResp, err
}

func listCorridors() ([]CorridorResponse, error) {
	resp, err := http.Get(corrdURL + "/v1/corridors")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	var corridors []CorridorResponse
	err = json.Unmarshal(body, &corridors)
	return corridors, err
}
