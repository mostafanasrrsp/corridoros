package main

import (
    "crypto/sha256"
    "encoding/hex"
    "encoding/json"
    "errors"
    "log"
    "math"
    "net/http"
    "sort"
    "strings"
    "sync"
    "time"
)

// Consent and governance
type Governance struct {
    WomenLed bool   `json:"women_led"`
    Contact  string `json:"contact"`
}

type Participant struct {
    Pseudonym     string   `json:"pseudonym"`
    Consent       bool     `json:"consent"`
    Scope         []string `json:"scope"`
    RetentionDays int      `json:"retention_days"`
}

type ConsentManifest struct {
    StudyID             string        `json:"study_id"`
    Version             string        `json:"version"`
    CommunityGovernance Governance    `json:"community_governance"`
    Participants        []Participant `json:"participants"`
    DataMinimization    bool          `json:"data_minimization"`
    CaptureMode         string        `json:"capture_mode"` // offline only supported
    Timestamp           time.Time     `json:"timestamp"`
}

// Synchrony session store
type Session struct {
    ID         string
    Manifest   ConsentManifest
    CreatedAt  time.Time
    Streams    map[string][]Series // key: stream type ("breath" or "rr")
}

type Series struct {
    Pseudonym string   `json:"pseudonym"`
    T         []float64 `json:"t"` // seconds
    V         []float64 `json:"v"`
}

// Requests / responses
type StartSessionRequest struct {
    Manifest ConsentManifest `json:"manifest"`
}

type StartSessionResponse struct {
    SessionID      string `json:"session_id"`
    AttestationID  string `json:"attestation_id"`
    ManifestHash   string `json:"manifest_hash"`
    Flags          []string `json:"flags"`
}

type IngestRequest struct {
    Stream       string   `json:"stream"` // "breath" or "rr"
    Participants []Series `json:"participants"`
}

type MetricsResponse struct {
    Stream              string             `json:"stream"`
    Participants        []string           `json:"participants"`
    WindowSeconds       float64            `json:"window_seconds"`
    PairwiseCorrelation map[string]float64 `json:"pairwise_correlation"`
    GroupSynchronyIndex float64            `json:"group_synchrony_index"`
    Notes               []string           `json:"notes"`
}

// Service implementation
type Service struct {
    mu       sync.RWMutex
    sessions map[string]*Session
}

func NewService() *Service {
    return &Service{sessions: make(map[string]*Session)}
}

// Handlers
func (s *Service) handleHealth(w http.ResponseWriter, r *http.Request) {
    writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Service) handleStartSession(w http.ResponseWriter, r *http.Request) {
    var req StartSessionRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "invalid request", http.StatusBadRequest)
        return
    }

    // Enforce governance and scope
    if !req.Manifest.CommunityGovernance.WomenLed {
        http.Error(w, "women_led governance required", http.StatusForbidden)
        return
    }
    if !req.Manifest.DataMinimization {
        http.Error(w, "data_minimization must be true", http.StatusForbidden)
        return
    }
    if strings.ToLower(req.Manifest.CaptureMode) != "offline" {
        http.Error(w, "only offline capture_mode supported", http.StatusForbidden)
        return
    }
    for _, p := range req.Manifest.Participants {
        if !p.Consent {
            http.Error(w, "all participants must consent", http.StatusForbidden)
            return
        }
    }

    // Generate session ID and manifest hash
    now := time.Now().UTC()
    manifestBytes, _ := json.Marshal(req.Manifest)
    h := sha256.Sum256(manifestBytes)
    manifestHash := hex.EncodeToString(h[:])
    sessionID := "sync-" + manifestHash[:8]
    attestationID := "eth-" + manifestHash[:12]

    s.mu.Lock()
    s.sessions[sessionID] = &Session{
        ID:        sessionID,
        Manifest:  req.Manifest,
        CreatedAt: now,
        Streams:   make(map[string][]Series),
    }
    s.mu.Unlock()

    resp := StartSessionResponse{
        SessionID:     sessionID,
        AttestationID: attestationID,
        ManifestHash:  manifestHash,
        Flags:         []string{"offline", "simulation"},
    }
    writeJSON(w, http.StatusCreated, resp)
}

func (s *Service) handleIngest(w http.ResponseWriter, r *http.Request) {
    sessionID := pathParam(r.URL.Path, 3) // /v1/synchrony/session/{id}/ingest
    if sessionID == "" {
        http.Error(w, "missing session id", http.StatusBadRequest)
        return
    }

    var req IngestRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "invalid request", http.StatusBadRequest)
        return
    }

    if req.Stream != "breath" && req.Stream != "rr" {
        http.Error(w, "unsupported stream (breath|rr)", http.StatusBadRequest)
        return
    }

    s.mu.Lock()
    defer s.mu.Unlock()
    sess, ok := s.sessions[sessionID]
    if !ok {
        http.Error(w, "session not found", http.StatusNotFound)
        return
    }
    // Store anonymized series (pseudonyms only)
    sess.Streams[req.Stream] = append(sess.Streams[req.Stream], req.Participants...)
    writeJSON(w, http.StatusAccepted, map[string]string{"status": "ingested"})
}

func (s *Service) handleMetrics(w http.ResponseWriter, r *http.Request) {
    sessionID := pathParam(r.URL.Path, 3) // /v1/synchrony/session/{id}/metrics
    stream := r.URL.Query().Get("stream")
    if stream == "" {
        stream = "breath"
    }

    s.mu.RLock()
    sess, ok := s.sessions[sessionID]
    s.mu.RUnlock()
    if !ok {
        http.Error(w, "session not found", http.StatusNotFound)
        return
    }

    series := sess.Streams[stream]
    if len(series) < 2 {
        http.Error(w, "need at least two participants", http.StatusBadRequest)
        return
    }

    // Compute pairwise Pearson correlations on a uniform grid
    step := 0.5 // seconds
    start, end := commonTimeBounds(series)
    if end-start < step*10 {
        http.Error(w, "insufficient overlap for analysis", http.StatusBadRequest)
        return
    }
    grid := makeGrid(start, end, step)
    resampled := make([][]float64, len(series))
    names := make([]string, len(series))
    for i, srs := range series {
        names[i] = srs.Pseudonym
        y, err := resample(grid, srs.T, srs.V)
        if err != nil {
            http.Error(w, "resampling error", http.StatusBadRequest)
            return
        }
        resampled[i] = zscore(y)
    }

    pairCorr := map[string]float64{}
    var sum float64
    var count int
    for i := 0; i < len(resampled); i++ {
        for j := i + 1; j < len(resampled); j++ {
            c := pearson(resampled[i], resampled[j])
            key := names[i] + "|" + names[j]
            pairCorr[key] = c
            sum += c
            count++
        }
    }
    gsi := sum / float64(count) // simple group synchrony index

    resp := MetricsResponse{
        Stream:              stream,
        Participants:        names,
        WindowSeconds:       end - start,
        PairwiseCorrelation: pairCorr,
        GroupSynchronyIndex: gsi,
        Notes:               []string{"offline", "anonymized", "women_led_required"},
    }
    writeJSON(w, http.StatusOK, resp)
}

// Utilities
func writeJSON(w http.ResponseWriter, code int, v any) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(code)
    _ = json.NewEncoder(w).Encode(v)
}

func pathParam(path string, idx int) string {
    parts := strings.Split(strings.Trim(path, "/"), "/")
    if len(parts) > idx {
        return parts[idx]
    }
    return ""
}

func commonTimeBounds(series []Series) (float64, float64) {
    start := -math.MaxFloat64
    end := math.MaxFloat64
    for _, s := range series {
        if len(s.T) == 0 { continue }
        if s.T[0] > start { start = s.T[0] }
        if s.T[len(s.T)-1] < end { end = s.T[len(s.T)-1] }
    }
    if start < 0 { start = 0 }
    if end < 0 { end = 0 }
    return start, end
}

func makeGrid(start, end, step float64) []float64 {
    n := int(math.Floor((end-start)/step)) + 1
    g := make([]float64, n)
    for i := 0; i < n; i++ {
        g[i] = start + float64(i)*step
    }
    return g
}

func resample(grid, t, v []float64) ([]float64, error) {
    if len(t) != len(v) || len(t) == 0 { return nil, errors.New("invalid series") }
    // Ensure sorted
    type tv struct{ t, v float64 }
    arr := make([]tv, len(t))
    for i := range t { arr[i] = tv{t: t[i], v: v[i]} }
    sort.Slice(arr, func(i, j int) bool { return arr[i].t < arr[j].t })
    for i := range t { t[i], v[i] = arr[i].t, arr[i].v }

    out := make([]float64, len(grid))
    j := 0
    for i, x := range grid {
        for j < len(t)-1 && t[j+1] < x { j++ }
        if j == len(t)-1 { out[i] = v[j]; continue }
        // Linear interpolation
        t0, t1 := t[j], t[j+1]
        v0, v1 := v[j], v[j+1]
        if t1 == t0 { out[i] = v0; continue }
        alpha := (x - t0) / (t1 - t0)
        out[i] = v0 + alpha*(v1 - v0)
    }
    return out, nil
}

func zscore(x []float64) []float64 {
    m := mean(x)
    s := stddev(x, m)
    if s == 0 { s = 1 }
    y := make([]float64, len(x))
    for i := range x { y[i] = (x[i]-m)/s }
    return y
}

func mean(x []float64) float64 {
    var s float64
    for _, v := range x { s += v }
    return s / float64(len(x))
}

func stddev(x []float64, m float64) float64 {
    var s float64
    for _, v := range x { d := v - m; s += d*d }
    return math.Sqrt(s / float64(len(x)))
}

func pearson(a, b []float64) float64 {
    if len(a) != len(b) || len(a) == 0 { return 0 }
    ma := mean(a)
    mb := mean(b)
    var num, da, db float64
    for i := range a {
        x := a[i] - ma
        y := b[i] - mb
        num += x * y
        da += x * x
        db += y * y
    }
    if da == 0 || db == 0 { return 0 }
    return num / math.Sqrt(da*db)
}

func main() {
    svc := NewService()

    mux := http.NewServeMux()
    mux.HandleFunc("/health", svc.handleHealth)
    mux.HandleFunc("/v1/synchrony/session/start", svc.handleStartSession)
    mux.HandleFunc("/v1/synchrony/session/", func(w http.ResponseWriter, r *http.Request) {
        // Routes: /v1/synchrony/session/{id}/ingest or /metrics
        if strings.HasSuffix(r.URL.Path, "/ingest") && r.Method == http.MethodPost {
            svc.handleIngest(w, r)
            return
        }
        if strings.HasSuffix(r.URL.Path, "/metrics") && r.Method == http.MethodGet {
            svc.handleMetrics(w, r)
            return
        }
        http.NotFound(w, r)
    })

    addr := ":8090"
    log.Printf("Starting Synchrony Analytics (offline) on %s", addr)
    log.Fatal(http.ListenAndServe(addr, mux))
}
