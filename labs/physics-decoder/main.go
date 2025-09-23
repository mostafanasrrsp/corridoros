package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

// PhysicsDecoderService provides physics calculations and dimensional analysis
type PhysicsDecoderService struct {
	// Constants
	SpeedOfLight     float64 // m/s
	PlanckConstant   float64 // J⋅s
	BoltzmannConstant float64 // J/K
	ElectronCharge   float64 // C
	AvogadroNumber   float64 // mol^-1
}

// DecoderRequest represents a physics calculation request
type DecoderRequest struct {
	Formula    string                 `json:"formula"`
	Variables  map[string]float64     `json:"variables"`
	Units      map[string]string      `json:"units"`
	Context    string                 `json:"context,omitempty"`
	Hypothesis bool                   `json:"hypothesis,omitempty"`
}

// DecoderResponse represents the calculation result
type DecoderResponse struct {
	Result      float64            `json:"result"`
	Unit        string             `json:"unit"`
	Formula     string             `json:"formula"`
	Steps       []CalculationStep  `json:"steps"`
	Valid       bool               `json:"valid"`
	Error       string             `json:"error,omitempty"`
	Warnings    []string           `json:"warnings,omitempty"`
	Dimensions  map[string]string  `json:"dimensions"`
	Context     string             `json:"context,omitempty"`
	Hypothesis  bool               `json:"hypothesis,omitempty"`
}

// CalculationStep represents a step in the calculation
type CalculationStep struct {
	Description string  `json:"description"`
	Value       float64 `json:"value"`
	Unit        string  `json:"unit"`
	Formula     string  `json:"formula,omitempty"`
}

// FormulaInfo represents information about a physics formula
type FormulaInfo struct {
	Name        string            `json:"name"`
	Formula     string            `json:"formula"`
	Description string            `json:"description"`
	Variables   map[string]string `json:"variables"`
	Units       map[string]string `json:"units"`
	Category    string            `json:"category"`
	Validated   bool              `json:"validated"`
}

// NewPhysicsDecoderService creates a new physics decoder service
func NewPhysicsDecoderService() *PhysicsDecoderService {
	return &PhysicsDecoderService{
		SpeedOfLight:     299792458.0,                    // m/s
		PlanckConstant:   6.62607015e-34,                 // J⋅s
		BoltzmannConstant: 1.380649e-23,                  // J/K
		ElectronCharge:   1.602176634e-19,                // C
		AvogadroNumber:   6.02214076e23,                  // mol^-1
	}
}

// Calculate performs physics calculations
func (p *PhysicsDecoderService) Calculate(req DecoderRequest) (*DecoderResponse, error) {
	response := &DecoderResponse{
		Formula:    req.Formula,
		Context:    req.Context,
		Hypothesis: req.Hypothesis,
		Steps:      []CalculationStep{},
		Dimensions: make(map[string]string),
		Warnings:   []string{},
	}

	// Parse and validate formula
	formula, err := p.parseFormula(req.Formula)
	if err != nil {
		response.Error = err.Error()
		response.Valid = false
		return response, nil
	}

	// Perform calculation based on formula type
	switch formula {
	case "energy_mass":
		result, steps, err := p.calculateEnergyMass(req.Variables, req.Units)
		if err != nil {
			response.Error = err.Error()
			response.Valid = false
			return response, nil
		}
		response.Result = result
		response.Unit = "J"
		response.Steps = steps
		response.Dimensions = map[string]string{"energy": "ML²T⁻²"}

	case "wavelength_frequency":
		result, steps, err := p.calculateWavelengthFrequency(req.Variables, req.Units)
		if err != nil {
			response.Error = err.Error()
			response.Valid = false
			return response, nil
		}
		response.Result = result
		response.Unit = "m"
		response.Steps = steps
		response.Dimensions = map[string]string{"wavelength": "L"}

	case "photon_energy":
		result, steps, err := p.calculatePhotonEnergy(req.Variables, req.Units)
		if err != nil {
			response.Error = err.Error()
			response.Valid = false
			return response, nil
		}
		response.Result = result
		response.Unit = "J"
		response.Steps = steps
		response.Dimensions = map[string]string{"energy": "ML²T⁻²"}

	case "thermal_energy":
		result, steps, err := p.calculateThermalEnergy(req.Variables, req.Units)
		if err != nil {
			response.Error = err.Error()
			response.Valid = false
			return response, nil
		}
		response.Result = result
		response.Unit = "J"
		response.Steps = steps
		response.Dimensions = map[string]string{"energy": "ML²T⁻²"}

	case "optical_power":
		result, steps, err := p.calculateOpticalPower(req.Variables, req.Units)
		if err != nil {
			response.Error = err.Error()
			response.Valid = false
			return response, nil
		}
		response.Result = result
		response.Unit = "W"
		response.Steps = steps
		response.Dimensions = map[string]string{"power": "ML²T⁻³"}

	default:
		response.Error = "Unknown formula: " + formula
		response.Valid = false
		return response, nil
	}

	response.Valid = true

	// Add warnings for hypothesis formulas
	if req.Hypothesis {
		response.Warnings = append(response.Warnings, "This calculation uses a hypothesis formula - verify results independently")
	}

	return response, nil
}

// parseFormula determines the type of formula from the input
func (p *PhysicsDecoderService) parseFormula(formula string) (string, error) {
	formula = strings.ToLower(strings.TrimSpace(formula))
	
	if strings.Contains(formula, "e=mc²") || strings.Contains(formula, "e=mc^2") {
		return "energy_mass", nil
	}
	if strings.Contains(formula, "λ=c/f") || strings.Contains(formula, "wavelength") {
		return "wavelength_frequency", nil
	}
	if strings.Contains(formula, "e=hf") || strings.Contains(formula, "photon") {
		return "photon_energy", nil
	}
	if strings.Contains(formula, "e=kt") || strings.Contains(formula, "thermal") {
		return "thermal_energy", nil
	}
	if strings.Contains(formula, "p=") || strings.Contains(formula, "power") {
		return "optical_power", nil
	}
	
	return "", fmt.Errorf("unrecognized formula: %s", formula)
}

// calculateEnergyMass calculates E = mc²
func (p *PhysicsDecoderService) calculateEnergyMass(vars map[string]float64, units map[string]string) (float64, []CalculationStep, error) {
	mass, ok := vars["m"]
	if !ok {
		return 0, nil, fmt.Errorf("mass variable 'm' not provided")
	}
	
	// Convert mass to kg if needed
	if unit, exists := units["m"]; exists {
		switch unit {
		case "g":
			mass = mass / 1000.0
		case "kg":
			// already in kg
		default:
			return 0, nil, fmt.Errorf("unsupported mass unit: %s", unit)
		}
	}
	
	c := p.SpeedOfLight
	result := mass * c * c
	
	steps := []CalculationStep{
		{
			Description: "Mass in kg",
			Value:       mass,
			Unit:        "kg",
		},
		{
			Description: "Speed of light",
			Value:       c,
			Unit:        "m/s",
		},
		{
			Description: "Energy calculation",
			Value:       result,
			Unit:        "J",
			Formula:     "E = mc²",
		},
	}
	
	return result, steps, nil
}

// calculateWavelengthFrequency calculates λ = c/f
func (p *PhysicsDecoderService) calculateWavelengthFrequency(vars map[string]float64, units map[string]string) (float64, []CalculationStep, error) {
	frequency, ok := vars["f"]
	if !ok {
		return 0, nil, fmt.Errorf("frequency variable 'f' not provided")
	}
	
	// Convert frequency to Hz if needed
	if unit, exists := units["f"]; exists {
		switch unit {
		case "kHz":
			frequency = frequency * 1000
		case "MHz":
			frequency = frequency * 1000000
		case "GHz":
			frequency = frequency * 1000000000
		case "THz":
			frequency = frequency * 1000000000000
		case "Hz":
			// already in Hz
		default:
			return 0, nil, fmt.Errorf("unsupported frequency unit: %s", unit)
		}
	}
	
	c := p.SpeedOfLight
	result := c / frequency
	
	steps := []CalculationStep{
		{
			Description: "Frequency in Hz",
			Value:       frequency,
			Unit:        "Hz",
		},
		{
			Description: "Speed of light",
			Value:       c,
			Unit:        "m/s",
		},
		{
			Description: "Wavelength calculation",
			Value:       result,
			Unit:        "m",
			Formula:     "λ = c/f",
		},
	}
	
	return result, steps, nil
}

// calculatePhotonEnergy calculates E = hf
func (p *PhysicsDecoderService) calculatePhotonEnergy(vars map[string]float64, units map[string]string) (float64, []CalculationStep, error) {
	frequency, ok := vars["f"]
	if !ok {
		return 0, nil, fmt.Errorf("frequency variable 'f' not provided")
	}
	
	// Convert frequency to Hz if needed
	if unit, exists := units["f"]; exists {
		switch unit {
		case "kHz":
			frequency = frequency * 1000
		case "MHz":
			frequency = frequency * 1000000
		case "GHz":
			frequency = frequency * 1000000000
		case "THz":
			frequency = frequency * 1000000000000
		case "Hz":
			// already in Hz
		default:
			return 0, nil, fmt.Errorf("unsupported frequency unit: %s", unit)
		}
	}
	
	h := p.PlanckConstant
	result := h * frequency
	
	steps := []CalculationStep{
		{
			Description: "Frequency in Hz",
			Value:       frequency,
			Unit:        "Hz",
		},
		{
			Description: "Planck constant",
			Value:       h,
			Unit:        "J⋅s",
		},
		{
			Description: "Photon energy calculation",
			Value:       result,
			Unit:        "J",
			Formula:     "E = hf",
		},
	}
	
	return result, steps, nil
}

// calculateThermalEnergy calculates E = kT
func (p *PhysicsDecoderService) calculateThermalEnergy(vars map[string]float64, units map[string]string) (float64, []CalculationStep, error) {
	temperature, ok := vars["T"]
	if !ok {
		return 0, nil, fmt.Errorf("temperature variable 'T' not provided")
	}
	
	// Convert temperature to K if needed
	if unit, exists := units["T"]; exists {
		switch unit {
		case "°C":
			temperature = temperature + 273.15
		case "°F":
			temperature = (temperature - 32) * 5/9 + 273.15
		case "K":
			// already in K
		default:
			return 0, nil, fmt.Errorf("unsupported temperature unit: %s", unit)
		}
	}
	
	k := p.BoltzmannConstant
	result := k * temperature
	
	steps := []CalculationStep{
		{
			Description: "Temperature in K",
			Value:       temperature,
			Unit:        "K",
		},
		{
			Description: "Boltzmann constant",
			Value:       k,
			Unit:        "J/K",
		},
		{
			Description: "Thermal energy calculation",
			Value:       result,
			Unit:        "J",
			Formula:     "E = kT",
		},
	}
	
	return result, steps, nil
}

// calculateOpticalPower calculates P = E/t or P = I*A
func (p *PhysicsDecoderService) calculateOpticalPower(vars map[string]float64, units map[string]string) (float64, []CalculationStep, error) {
	// Try P = E/t first
	if energy, ok := vars["E"]; ok {
		time, ok := vars["t"]
		if !ok {
			return 0, nil, fmt.Errorf("time variable 't' not provided for P = E/t")
		}
		
		result := energy / time
		
		steps := []CalculationStep{
			{
				Description: "Energy",
				Value:       energy,
				Unit:        "J",
			},
			{
				Description: "Time",
				Value:       time,
				Unit:        "s",
			},
			{
				Description: "Power calculation",
				Value:       result,
				Unit:        "W",
				Formula:     "P = E/t",
			},
		}
		
		return result, steps, nil
	}
	
	// Try P = I*A
	if intensity, ok := vars["I"]; ok {
		area, ok := vars["A"]
		if !ok {
			return 0, nil, fmt.Errorf("area variable 'A' not provided for P = I*A")
		}
		
		result := intensity * area
		
		steps := []CalculationStep{
			{
				Description: "Intensity",
				Value:       intensity,
				Unit:        "W/m²",
			},
			{
				Description: "Area",
				Value:       area,
				Unit:        "m²",
			},
			{
				Description: "Power calculation",
				Value:       result,
				Unit:        "W",
				Formula:     "P = I*A",
			},
		}
		
		return result, steps, nil
	}
	
	return 0, nil, fmt.Errorf("insufficient variables for power calculation")
}

// GetFormulas returns available physics formulas
func (p *PhysicsDecoderService) GetFormulas() []FormulaInfo {
	return []FormulaInfo{
		{
			Name:        "Mass-Energy Equivalence",
			Formula:     "E = mc²",
			Description: "Einstein's mass-energy equivalence",
			Variables:   map[string]string{"m": "mass", "c": "speed of light"},
			Units:       map[string]string{"m": "kg", "c": "m/s"},
			Category:    "Relativity",
			Validated:   true,
		},
		{
			Name:        "Wavelength-Frequency Relationship",
			Formula:     "λ = c/f",
			Description: "Relationship between wavelength and frequency",
			Variables:   map[string]string{"λ": "wavelength", "c": "speed of light", "f": "frequency"},
			Units:       map[string]string{"λ": "m", "c": "m/s", "f": "Hz"},
			Category:    "Optics",
			Validated:   true,
		},
		{
			Name:        "Photon Energy",
			Formula:     "E = hf",
			Description: "Energy of a photon",
			Variables:   map[string]string{"E": "energy", "h": "Planck constant", "f": "frequency"},
			Units:       map[string]string{"E": "J", "h": "J⋅s", "f": "Hz"},
			Category:    "Quantum Mechanics",
			Validated:   true,
		},
		{
			Name:        "Thermal Energy",
			Formula:     "E = kT",
			Description: "Average thermal energy per degree of freedom",
			Variables:   map[string]string{"E": "energy", "k": "Boltzmann constant", "T": "temperature"},
			Units:       map[string]string{"E": "J", "k": "J/K", "T": "K"},
			Category:    "Thermodynamics",
			Validated:   true,
		},
		{
			Name:        "Optical Power",
			Formula:     "P = E/t or P = I*A",
			Description: "Power calculation from energy/time or intensity*area",
			Variables:   map[string]string{"P": "power", "E": "energy", "t": "time", "I": "intensity", "A": "area"},
			Units:       map[string]string{"P": "W", "E": "J", "t": "s", "I": "W/m²", "A": "m²"},
			Category:    "Optics",
			Validated:   true,
		},
	}
}

// HTTP handlers
func (p *PhysicsDecoderService) handleCalculate(w http.ResponseWriter, r *http.Request) {
	var req DecoderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	response, err := p.Calculate(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (p *PhysicsDecoderService) handleGetFormulas(w http.ResponseWriter, r *http.Request) {
	formulas := p.GetFormulas()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(formulas)
}

func (p *PhysicsDecoderService) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func main() {
	// Create physics decoder service
	service := NewPhysicsDecoderService()

	// Set up HTTP router
	router := mux.NewRouter()
	api := router.PathPrefix("/v1/physics").Subrouter()

	// API endpoints
	api.HandleFunc("/calculate", service.handleCalculate).Methods("POST")
	api.HandleFunc("/formulas", service.handleGetFormulas).Methods("GET")
	api.HandleFunc("/health", service.handleHealth).Methods("GET")

	// Health check
	router.HandleFunc("/health", service.handleHealth).Methods("GET")

	// Start server
	log.Println("Starting Physics Decoder service on :8085")
	log.Fatal(http.ListenAndServe(":8085", router))
}
