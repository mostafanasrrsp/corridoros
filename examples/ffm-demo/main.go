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

// FFMHandle represents a Free-Form Memory allocation
type FFMHandle struct {
	ID               string    `json:"id"`
	Bytes            uint64    `json:"bytes"`
	LatencyClass     string    `json:"latency_class"`
	BandwidthFloor   uint64    `json:"bandwidth_floor_GBs"`
	Persistence      string    `json:"persistence"`
	Shareable        bool      `json:"shareable"`
	SecurityDomain   string    `json:"security_domain"`
	CreatedAt        time.Time `json:"created_at"`
	PolicyLeaseTTL   int       `json:"policy_lease_ttl_s"`
	FileDescriptors  []string  `json:"fds"`
	AchievedBandwidth uint64   `json:"achieved_GBs"`
	MovedPages       uint64    `json:"moved_pages"`
	TailP99Ms        float64   `json:"tail_p99_ms"`
}

// AllocationRequest represents a memory allocation request
type AllocationRequest struct {
	Bytes            uint64 `json:"bytes"`
	LatencyClass     string `json:"latency_class"`
	BandwidthFloor   uint64 `json:"bandwidth_floor_GBs"`
	Persistence      string `json:"persistence"`
	Shareable        bool   `json:"shareable"`
	SecurityDomain   string `json:"security_domain"`
}

// TelemetryResponse represents telemetry data
type TelemetryResponse struct {
	AchievedGBs  uint64  `json:"achieved_GBs"`
	MovedPages   uint64  `json:"moved_pages"`
	TailP99Ms    float64 `json:"tail_p99_ms"`
	Temperature  float64 `json:"temperature_c"`
	PowerW       float64 `json:"power_w"`
	Utilization  float64 `json:"utilization_percent"`
}

const memqosdURL = "http://localhost:8081"

func main() {
	fmt.Println("CorridorOS FFM Demo")
	fmt.Println("==================")

	// Test 1: Allocate different tiers of memory
	fmt.Println("\n1. Allocating memory across different tiers...")
	
	allocations := []AllocationRequest{
		{
			Bytes:          16 * 1024 * 1024 * 1024, // 16GB
			LatencyClass:   "T0", // HBM
			BandwidthFloor: 500,
			Persistence:    "none",
			Shareable:      true,
			SecurityDomain: "demo-tier0",
		},
		{
			Bytes:          32 * 1024 * 1024 * 1024, // 32GB
			LatencyClass:   "T1", // DRAM
			BandwidthFloor: 200,
			Persistence:    "none",
			Shareable:      true,
			SecurityDomain: "demo-tier1",
		},
		{
			Bytes:          64 * 1024 * 1024 * 1024, // 64GB
			LatencyClass:   "T2", // CXL
			BandwidthFloor: 100,
			Persistence:    "none",
			Shareable:      true,
			SecurityDomain: "demo-tier2",
		},
		{
			Bytes:          128 * 1024 * 1024 * 1024, // 128GB
			LatencyClass:   "T3", // Persistent
			BandwidthFloor: 50,
			Persistence:    "write-back",
			Shareable:      true,
			SecurityDomain: "demo-tier3",
		},
	}

	var handles []FFMHandle
	for _, req := range allocations {
		handle, err := allocateFFM(req)
		if err != nil {
			log.Printf("Error allocating %s: %v", req.LatencyClass, err)
			continue
		}
		handles = append(handles, *handle)
		fmt.Printf("  %s: %s allocated (ID: %s)\n", req.LatencyClass, formatBytes(req.Bytes), handle.ID)
	}

	// Test 2: Monitor telemetry
	fmt.Println("\n2. Monitoring telemetry for 30 seconds...")
	
	for i := 0; i < 30; i++ {
		fmt.Printf("\n--- Telemetry Update %d/30 ---\n", i+1)
		
		for _, handle := range handles {
			telemetry, err := getTelemetry(handle.ID)
			if err != nil {
				log.Printf("Error getting telemetry for %s: %v", handle.ID, err)
				continue
			}
			
			bandwidthRatio := float64(telemetry.AchievedGBs) / float64(handle.BandwidthFloor) * 100
			fmt.Printf("  %s: %d/%d Gbps (%.1f%%) | P99: %.2fms | Util: %.1f%%\n",
				handle.LatencyClass,
				telemetry.AchievedGBs,
				handle.BandwidthFloor,
				bandwidthRatio,
				telemetry.TailP99Ms,
				telemetry.Utilization)
		}
		
		time.Sleep(1 * time.Second)
	}

	// Test 3: Bandwidth adjustment
	fmt.Println("\n3. Testing bandwidth adjustment...")
	
	if len(handles) > 0 {
		handle := handles[0]
		fmt.Printf("Adjusting bandwidth for %s from %d to %d Gbps...\n",
			handle.LatencyClass, handle.BandwidthFloor, handle.BandwidthFloor+50)
		
		err := adjustBandwidth(handle.ID, handle.BandwidthFloor+50)
		if err != nil {
			log.Printf("Error adjusting bandwidth: %v", err)
		} else {
			fmt.Println("Bandwidth adjustment successful!")
		}
	}

	// Test 4: Tier migration
	fmt.Println("\n4. Testing tier migration...")
	
	if len(handles) > 1 {
		handle := handles[1]
		fmt.Printf("Migrating %s from %s to T2...\n", handle.ID, handle.LatencyClass)
		
		err := migrateTier(handle.ID, "T2")
		if err != nil {
			log.Printf("Error migrating tier: %v", err)
		} else {
			fmt.Println("Tier migration successful!")
		}
	}

	// Test 5: List all allocations
	fmt.Println("\n5. Current allocations:")
	
	allAllocations, err := listFFM()
	if err != nil {
		log.Printf("Error listing allocations: %v", err)
	} else {
		fmt.Printf("%-12s %-8s %-12s %-8s %-8s %-12s\n", 
			"ID", "Tier", "Size", "Bw Floor", "Achieved", "Domain")
		fmt.Println("------------------------------------------------------------")
		
		for _, alloc := range allAllocations {
			fmt.Printf("%-12s %-8s %-12s %-8d %-8d %-12s\n",
				alloc.ID,
				alloc.LatencyClass,
				formatBytes(alloc.Bytes),
				alloc.BandwidthFloor,
				alloc.AchievedBandwidth,
				alloc.SecurityDomain)
		}
	}

	fmt.Println("\nDemo completed!")
}

// HTTP client functions
func allocateFFM(req AllocationRequest) (*FFMHandle, error) {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(memqosdURL+"/v1/ffm/alloc", "application/json", bytes.NewBuffer(jsonData))
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

	var handle FFMHandle
	err = json.Unmarshal(body, &handle)
	return &handle, err
}

func getTelemetry(id string) (*TelemetryResponse, error) {
	resp, err := http.Get(memqosdURL + "/v1/ffm/" + id + "/telemetry")
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

	var telemetry TelemetryResponse
	err = json.Unmarshal(body, &telemetry)
	return &telemetry, err
}

func adjustBandwidth(id string, newBandwidth uint64) error {
	req := map[string]uint64{"floor_GBs": newBandwidth}
	jsonData, err := json.Marshal(req)
	if err != nil {
		return err
	}

	httpReq, err := http.NewRequest("PATCH", memqosdURL+"/v1/ffm/"+id+"/bandwidth", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

func migrateTier(id string, newTier string) error {
	req := map[string]string{"target": newTier}
	jsonData, err := json.Marshal(req)
	if err != nil {
		return err
	}

	httpReq, err := http.NewRequest("PATCH", memqosdURL+"/v1/ffm/"+id+"/latency_class", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

func listFFM() ([]FFMHandle, error) {
	resp, err := http.Get(memqosdURL + "/v1/ffm/")
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

	var allocations []FFMHandle
	err = json.Unmarshal(body, &allocations)
	return allocations, err
}

// Helper function
func formatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
