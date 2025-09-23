package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/rand"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

// HELIOPASSSimulator simulates the HELIOPASS calibration system
type HELIOPASSSimulator struct {
	// Simulation parameters
	BaseTemperature    float64
	BaseHumidity       float64
	BaseVibration      float64
	BaseEMI            float64
	DriftRate          float64
	NoiseLevel         float64
	ConvergenceRate    float64
	MaxIterations      int
}

// SimulationRequest represents a HELIOPASS simulation request
type SimulationRequest struct {
	CorridorID       string    `json:"corridor_id"`
	TargetBER        float64   `json:"target_ber"`
	AmbientProfile   string    `json:"ambient_profile"`
	LambdaCount      int       `json:"lambda_count"`
	InitialBER       float64   `json:"initial_ber,omitempty"`
	InitialEyeMargin float64   `json:"initial_eye_margin,omitempty"`
	Temperature      float64   `json:"temperature_c,omitempty"`
	Duration         int       `json:"duration_seconds,omitempty"`
}

// SimulationResponse represents the simulation results
type SimulationResponse struct {
	CorridorID         string                 `json:"corridor_id"`
	Status             string                 `json:"status"`
	Converged          bool                   `json:"converged"`
	FinalBER           float64                `json:"final_ber"`
	FinalEyeMargin     float64                `json:"final_eye_margin"`
	ConvergenceTime    float64                `json:"convergence_time_seconds"`
	Iterations         int                    `json:"iterations"`
	BiasVoltages       []float64              `json:"bias_voltages_mv"`
	LambdaShifts       []float64              `json:"lambda_shifts_nm"`
	LaserPowerAdjust   []float64              `json:"laser_power_adjust_db"`
	PowerSavings       float64                `json:"power_savings_percent"`
	TemperatureProfile []TemperaturePoint     `json:"temperature_profile"`
	BERProfile         []BERPoint             `json:"ber_profile"`
	EyeMarginProfile   []EyeMarginPoint       `json:"eye_margin_profile"`
	Error              string                 `json:"error,omitempty"`
}

// TemperaturePoint represents a temperature measurement
type TemperaturePoint struct {
	Time        float64 `json:"time_seconds"`
	Temperature float64 `json:"temperature_c"`
}

// BERPoint represents a BER measurement
type BERPoint struct {
	Time float64 `json:"time_seconds"`
	BER  float64 `json:"ber"`
}

// EyeMarginPoint represents an eye margin measurement
type EyeMarginPoint struct {
	Time      float64 `json:"time_seconds"`
	EyeMargin float64 `json:"eye_margin_ui"`
}

// AmbientProfile represents environmental conditions
type AmbientProfile struct {
	Name           string  `json:"name"`
	Temperature    float64 `json:"temperature_c"`
	Humidity       float64 `json:"humidity_percent"`
	VibrationRMS   float64 `json:"vibration_rms_um"`
	EMINoise       float64 `json:"emi_noise_db"`
	DriftRate      float64 `json:"drift_rate_nm_per_hour"`
	StabilityClass string  `json:"stability_class"`
	NoiseLevel     float64 `json:"noise_level"`
}

// NewHELIOPASSSimulator creates a new HELIOPASS simulator
func NewHELIOPASSSimulator() *HELIOPASSSimulator {
	return &HELIOPASSSimulator{
		BaseTemperature: 22.0,
		BaseHumidity:    45.0,
		BaseVibration:   0.1,
		BaseEMI:         -80.0,
		DriftRate:       0.001,
		NoiseLevel:      0.1,
		ConvergenceRate: 0.8,
		MaxIterations:   50,
	}
}

// GetAmbientProfiles returns available ambient profiles
func (h *HELIOPASSSimulator) GetAmbientProfiles() map[string]AmbientProfile {
	return map[string]AmbientProfile{
		"lab_default": {
			Name:           "Laboratory Default",
			Temperature:    22.0,
			Humidity:       45.0,
			VibrationRMS:   0.1,
			EMINoise:       -80.0,
			DriftRate:      0.001,
			StabilityClass: "excellent",
			NoiseLevel:     0.05,
		},
		"field_noise_low": {
			Name:           "Field Low Noise",
			Temperature:    25.0,
			Humidity:       60.0,
			VibrationRMS:   1.0,
			EMINoise:       -70.0,
			DriftRate:      0.01,
			StabilityClass: "good",
			NoiseLevel:     0.1,
		},
		"field_noise_high": {
			Name:           "Field High Noise",
			Temperature:    30.0,
			Humidity:       80.0,
			VibrationRMS:   5.0,
			EMINoise:       -60.0,
			DriftRate:      0.1,
			StabilityClass: "fair",
			NoiseLevel:     0.2,
		},
		"datacenter": {
			Name:           "Data Center",
			Temperature:    24.0,
			Humidity:       50.0,
			VibrationRMS:   0.5,
			EMINoise:       -75.0,
			DriftRate:      0.005,
			StabilityClass: "excellent",
			NoiseLevel:     0.08,
		},
		"space_sim": {
			Name:           "Space Simulation",
			Temperature:    -50.0,
			Humidity:       0.0,
			VibrationRMS:   0.01,
			EMINoise:       -90.0,
			DriftRate:      0.0001,
			StabilityClass: "excellent",
			NoiseLevel:     0.01,
		},
	}
}

// Simulate performs HELIOPASS simulation
func (h *HELIOPASSSimulator) Simulate(req SimulationRequest) (*SimulationResponse, error) {
	profiles := h.GetAmbientProfiles()
	profile, exists := profiles[req.AmbientProfile]
	if !exists {
		return nil, fmt.Errorf("unknown ambient profile: %s", req.AmbientProfile)
	}

	// Set defaults
	if req.LambdaCount == 0 {
		req.LambdaCount = 8
	}
	if req.InitialBER == 0 {
		req.InitialBER = 1e-9
	}
	if req.InitialEyeMargin == 0 {
		req.InitialEyeMargin = 0.5
	}
	if req.Temperature == 0 {
		req.Temperature = profile.Temperature
	}
	if req.Duration == 0 {
		req.Duration = 60
	}

	// Initialize simulation state
	currentBER := req.InitialBER
	currentEyeMargin := req.InitialEyeMargin
	targetBER := req.TargetBER

	// Initialize bias voltages and lambda shifts
	biasVoltages := make([]float64, req.LambdaCount)
	lambdaShifts := make([]float64, req.LambdaCount)
	laserPowerAdjust := make([]float64, req.LambdaCount)

	for i := range biasVoltages {
		biasVoltages[i] = 1.2 + (rand.Float64()-0.5)*0.2
		lambdaShifts[i] = (rand.Float64() - 0.5) * 0.02
		laserPowerAdjust[i] = (rand.Float64() - 0.5) * 0.5
	}

	// Simulation profiles
	temperatureProfile := []TemperaturePoint{}
	berProfile := []BERPoint{}
	eyeMarginProfile := []EyeMarginPoint{}

	// Run simulation
	converged := false
	iterations := 0
	dt := float64(req.Duration) / float64(h.MaxIterations)

	for i := 0; i < h.MaxIterations; i++ {
		iterations++
		time := float64(i) * dt

		// Update temperature with ambient profile and noise
		temperature := profile.Temperature + h.simulateTemperatureNoise(time, profile)
		temperatureProfile = append(temperatureProfile, TemperaturePoint{
			Time:        time,
			Temperature: temperature,
		})

		// Simulate BER improvement
		improvement := h.calculateImprovement(i, profile.NoiseLevel)
		currentBER = targetBER + (currentBER-targetBER)*improvement

		// Add noise
		berNoise := h.calculateBERNoise(time, profile)
		currentBER += berNoise
		currentBER = math.Max(currentBER, 1e-15) // Minimum BER

		berProfile = append(berProfile, BERPoint{
			Time: time,
			BER:  currentBER,
		})

		// Simulate eye margin improvement
		eyeImprovement := h.calculateEyeImprovement(i, profile.NoiseLevel)
		currentEyeMargin = 0.8 + (currentEyeMargin-0.8)*eyeImprovement

		// Add noise to eye margin
		eyeNoise := h.calculateEyeNoise(time, profile)
		currentEyeMargin += eyeNoise
		currentEyeMargin = math.Max(0.1, math.Min(1.5, currentEyeMargin))

		eyeMarginProfile = append(eyeMarginProfile, EyeMarginPoint{
			Time:      time,
			EyeMargin: currentEyeMargin,
		})

		// Update bias voltages and lambda shifts
		h.updateBiasVoltages(biasVoltages, time, profile)
		h.updateLambdaShifts(lambdaShifts, time, profile)
		h.updateLaserPower(laserPowerAdjust, time, profile)

		// Check convergence
		if currentBER <= targetBER*1.1 && currentEyeMargin >= 0.7 {
			converged = true
			break
		}
	}

	// Calculate final metrics
	convergenceTime := float64(iterations) * dt
	powerSavings := h.calculatePowerSavings(biasVoltages, laserPowerAdjust)

	status := "converged"
	if !converged {
		status = "partial_convergence"
	}

	return &SimulationResponse{
		CorridorID:         req.CorridorID,
		Status:             status,
		Converged:          converged,
		FinalBER:           currentBER,
		FinalEyeMargin:     currentEyeMargin,
		ConvergenceTime:    convergenceTime,
		Iterations:         iterations,
		BiasVoltages:       biasVoltages,
		LambdaShifts:       lambdaShifts,
		LaserPowerAdjust:   laserPowerAdjust,
		PowerSavings:       powerSavings,
		TemperatureProfile: temperatureProfile,
		BERProfile:         berProfile,
		EyeMarginProfile:   eyeMarginProfile,
	}, nil
}

// Helper methods for simulation
func (h *HELIOPASSSimulator) simulateTemperatureNoise(time float64, profile AmbientProfile) float64 {
	// Simulate temperature drift and noise
	drift := math.Sin(time*0.1) * 0.5
	noise := (rand.Float64() - 0.5) * profile.NoiseLevel * 2
	return drift + noise
}

func (h *HELIOPASSSimulator) calculateImprovement(iteration int, noiseLevel float64) float64 {
	// Exponential improvement with noise
	baseImprovement := math.Exp(-float64(iteration) * h.ConvergenceRate)
	noise := (rand.Float64() - 0.5) * noiseLevel
	return baseImprovement + noise
}

func (h *HELIOPASSSimulator) calculateBERNoise(time float64, profile AmbientProfile) float64 {
	// BER noise based on environmental conditions
	baseNoise := profile.NoiseLevel * 1e-12
	timeNoise := math.Sin(time*0.5) * baseNoise * 0.5
	randomNoise := (rand.Float64() - 0.5) * baseNoise
	return timeNoise + randomNoise
}

func (h *HELIOPASSSimulator) calculateEyeImprovement(iteration int, noiseLevel float64) float64 {
	// Similar to BER improvement but for eye margin
	baseImprovement := math.Exp(-float64(iteration) * h.ConvergenceRate * 0.8)
	noise := (rand.Float64() - 0.5) * noiseLevel * 0.1
	return baseImprovement + noise
}

func (h *HELIOPASSSimulator) calculateEyeNoise(time float64, profile AmbientProfile) float64 {
	// Eye margin noise
	baseNoise := profile.NoiseLevel * 0.01
	timeNoise := math.Sin(time*0.3) * baseNoise * 0.5
	randomNoise := (rand.Float64() - 0.5) * baseNoise
	return timeNoise + randomNoise
}

func (h *HELIOPASSSimulator) updateBiasVoltages(voltages []float64, time float64, profile AmbientProfile) {
	for i := range voltages {
		// Temperature compensation
		tempFactor := 1.0 + (profile.Temperature-h.BaseTemperature)*0.001
		// Drift compensation
		driftFactor := 1.0 + math.Sin(time*0.2)*profile.DriftRate*0.1
		// Random adjustment
		randomAdjust := (rand.Float64() - 0.5) * 0.01
		
		voltages[i] = voltages[i] * tempFactor * driftFactor + randomAdjust
		voltages[i] = math.Max(0.8, math.Min(1.5, voltages[i])) // Clamp to valid range
	}
}

func (h *HELIOPASSSimulator) updateLambdaShifts(shifts []float64, time float64, profile AmbientProfile) {
	for i := range shifts {
		// Drift over time
		drift := math.Sin(time*0.15) * profile.DriftRate * 0.01
		// Random adjustment
		randomAdjust := (rand.Float64() - 0.5) * 0.001
		
		shifts[i] = shifts[i] + drift + randomAdjust
		shifts[i] = math.Max(-0.1, math.Min(0.1, shifts[i])) // Clamp to valid range
	}
}

func (h *HELIOPASSSimulator) updateLaserPower(powerAdjust []float64, time float64, profile AmbientProfile) {
	for i := range powerAdjust {
		// Temperature compensation
		tempFactor := 1.0 + (profile.Temperature-h.BaseTemperature)*0.0005
		// Random adjustment
		randomAdjust := (rand.Float64() - 0.5) * 0.1
		
		powerAdjust[i] = powerAdjust[i] * tempFactor + randomAdjust
		powerAdjust[i] = math.Max(-2.0, math.Min(2.0, powerAdjust[i])) // Clamp to valid range
	}
}

func (h *HELIOPASSSimulator) calculatePowerSavings(biasVoltages []float64, laserPowerAdjust []float64) float64 {
	// Calculate power savings based on optimized bias voltages and laser power
	voltageSavings := 0.0
	for _, v := range biasVoltages {
		// Lower voltages generally mean lower power
		voltageSavings += (1.2 - v) * 10.0 // 10% per 0.1V difference
	}
	voltageSavings = voltageSavings / float64(len(biasVoltages))

	laserSavings := 0.0
	for _, p := range laserPowerAdjust {
		// Negative adjustments mean power savings
		if p < 0 {
			laserSavings += math.Abs(p) * 5.0 // 5% per dB reduction
		}
	}
	laserSavings = laserSavings / float64(len(laserPowerAdjust))

	return math.Max(0, math.Min(20, voltageSavings+laserSavings)) // Cap at 20%
}

// HTTP handlers
func (h *HELIOPASSSimulator) handleSimulate(w http.ResponseWriter, r *http.Request) {
	var req SimulationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	response, err := h.Simulate(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *HELIOPASSSimulator) handleGetProfiles(w http.ResponseWriter, r *http.Request) {
	profiles := h.GetAmbientProfiles()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(profiles)
}

func (h *HELIOPASSSimulator) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func main() {
	// Initialize random seed
	rand.Seed(time.Now().UnixNano())

	// Create HELIOPASS simulator
	simulator := NewHELIOPASSSimulator()

	// Set up HTTP router
	router := mux.NewRouter()
	api := router.PathPrefix("/v1/helio-sim").Subrouter()

	// API endpoints
	api.HandleFunc("/simulate", simulator.handleSimulate).Methods("POST")
	api.HandleFunc("/profiles", simulator.handleGetProfiles).Methods("GET")
	api.HandleFunc("/health", simulator.handleHealth).Methods("GET")

	// Health check
	router.HandleFunc("/health", simulator.handleHealth).Methods("GET")

	// Start server
	log.Println("Starting HELIOPASS Simulator on :8086")
	log.Fatal(http.ListenAndServe(":8086", router))
}
