// memqosd_skeleton.go — Free-Form Memory daemon (skeleton)
package main

import (
	"encoding/json"
	"log"
	"net/http"
)

type FFMAllocRequest struct {
	Bytes              uint64 `json:"bytes"`
	LatencyClass       string `json:"latency_class"` // T0..T3
	BandwidthFloorGBs  uint32 `json:"bandwidth_floor_GBs"`
	Persistence        string `json:"persistence"`   // none|durable
	Shareable          bool   `json:"shareable"`
	SecurityDomain     string `json:"security_domain"`
}

type FFMAllocReply struct {
	Handle            string   `json:"ffm_handle"`
	FDs               []string `json:"fds"`
	PolicyLeaseTTLsec int      `json:"policy_lease_ttl_s"`
}

func ffmAlloc(w http.ResponseWriter, r *http.Request) {
	var req FFMAllocRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), 400); return
	}
	// TODO: build/choose CXL region, create DAX-backed file, mmap handle.
	reply := FFMAllocReply{ Handle: "ffm-9c2e", FDs: []string{"/proc/self/fd/37"}, PolicyLeaseTTLsec: 3600 }
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(reply)
}

func main() {
	http.HandleFunc("/v1/ffm/alloc", ffmAlloc)
	// TODO: PATCH /v1/ffm/{handle}/bandwidth
	// TODO: PATCH /v1/ffm/{handle}/latency_class
	// TODO: GET   /v1/ffm/{handle}/telemetry
	log.Println("memqosd skeleton listening on :7070")
	log.Fatal(http.ListenAndServe(":7070", nil))
}
