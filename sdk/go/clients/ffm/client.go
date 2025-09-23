package ffm

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
)

type AllocateRequest struct {
    Bytes              uint64 `json:"bytes"`
    LatencyClass       string `json:"latency_class"`
    BandwidthFloorGBs  uint64 `json:"bandwidth_floor_GBs"`
    Persistence        string `json:"persistence"`
    Shareable          bool   `json:"shareable"`
    SecurityDomain     string `json:"security_domain"`
    AttestationRequired bool  `json:"attestation_required,omitempty"`
    AttestationTicket   string `json:"attestation_ticket,omitempty"`
}

type Handle struct {
    ID    string `json:"id"`
    Bytes uint64 `json:"bytes"`
}

type Telemetry struct {
    AchievedGBs uint64 `json:"achieved_GBs"`
}

type Client struct { BaseURL string; HTTP *http.Client }

func New(base string) *Client { return &Client{BaseURL: base, HTTP: &http.Client{}} }

func (c *Client) Allocate(req AllocateRequest) (*Handle, error) {
    b, _ := json.Marshal(req)
    resp, err := c.HTTP.Post(c.BaseURL+"/v1/ffm/alloc", "application/json", bytes.NewBuffer(b))
    if err != nil { return nil, err }
    defer resp.Body.Close()
    if resp.StatusCode != http.StatusCreated { body,_ := io.ReadAll(resp.Body); return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body)) }
    var h Handle
    return &h, json.NewDecoder(resp.Body).Decode(&h)
}

func (c *Client) Get(id string) (*Handle, error) {
    resp, err := c.HTTP.Get(c.BaseURL+"/v1/ffm/"+id)
    if err != nil { return nil, err }
    defer resp.Body.Close()
    if resp.StatusCode != http.StatusOK { body,_ := io.ReadAll(resp.Body); return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body)) }
    var h Handle
    return &h, json.NewDecoder(resp.Body).Decode(&h)
}

func (c *Client) Telemetry(id string) (*Telemetry, error) {
    resp, err := c.HTTP.Get(c.BaseURL+"/v1/ffm/"+id+"/telemetry")
    if err != nil { return nil, err }
    defer resp.Body.Close()
    if resp.StatusCode != http.StatusOK { body,_ := io.ReadAll(resp.Body); return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body)) }
    var t Telemetry
    return &t, json.NewDecoder(resp.Body).Decode(&t)
}

