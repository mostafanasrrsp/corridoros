package corridor

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
)

type QoSConfig struct {
    PFC      bool   `json:"pfc"`
    Priority string `json:"priority"`
}

type AllocateRequest struct {
    CorridorType       string    `json:"corridor_type"`
    Lanes              int       `json:"lanes"`
    LambdaNm           []int     `json:"lambda_nm"`
    MinGbps            int       `json:"min_gbps"`
    LatencyBudgetNs    int       `json:"latency_budget_ns"`
    ReachMm            int       `json:"reach_mm"`
    Mode               string    `json:"mode"`
    QoS                QoSConfig `json:"qos"`
    AttestationRequired bool     `json:"attestation_required"`
    AttestationTicket   *string  `json:"attestation_ticket,omitempty"`
}

type Corridor struct {
    ID              string    `json:"id"`
    CorridorType    string    `json:"corridor_type"`
    Lanes           int       `json:"lanes"`
    LambdaNm        []int     `json:"lambda_nm"`
    AchievableGbps  int       `json:"achievable_gbps"`
    Status          string    `json:"status"`
}

type Telemetry struct {
    BER           float64 `json:"ber"`
    TempC         float64 `json:"temp_c"`
    PowerPjPerBit float64 `json:"power_pj_per_bit"`
}

type RecalRequest struct {
    TargetBER      float64 `json:"target_ber"`
    AmbientProfile string  `json:"ambient_profile"`
}

type RecalResponse struct {
    Status            string   `json:"status"`
    Converged         bool     `json:"converged"`
    BiasVoltages      []float64 `json:"bias_voltages_mv"`
}

type Client struct { BaseURL string; HTTP *http.Client }

func New(base string) *Client { return &Client{BaseURL: base, HTTP: &http.Client{}} }

func (c *Client) Allocate(req AllocateRequest) (*Corridor, error) {
    b, _ := json.Marshal(req)
    resp, err := c.HTTP.Post(c.BaseURL+"/v1/corridors", "application/json", bytes.NewBuffer(b))
    if err != nil { return nil, err }
    defer resp.Body.Close()
    if resp.StatusCode != http.StatusCreated { body,_ := io.ReadAll(resp.Body); return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body)) }
    var cor Corridor
    return &cor, json.NewDecoder(resp.Body).Decode(&cor)
}

func (c *Client) Telemetry(id string) (*Telemetry, error) {
    resp, err := c.HTTP.Get(c.BaseURL+"/v1/corridors/"+id+"/telemetry")
    if err != nil { return nil, err }
    defer resp.Body.Close()
    if resp.StatusCode != http.StatusOK { body,_ := io.ReadAll(resp.Body); return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body)) }
    var t Telemetry
    return &t, json.NewDecoder(resp.Body).Decode(&t)
}

func (c *Client) Recalibrate(id string, r RecalRequest) (*RecalResponse, error) {
    b, _ := json.Marshal(r)
    resp, err := c.HTTP.Post(c.BaseURL+"/v1/corridors/"+id+"/recalibrate", "application/json", bytes.NewBuffer(b))
    if err != nil { return nil, err }
    defer resp.Body.Close()
    if resp.StatusCode != http.StatusOK { body,_ := io.ReadAll(resp.Body); return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body)) }
    var out RecalResponse
    return &out, json.NewDecoder(resp.Body).Decode(&out)
}

