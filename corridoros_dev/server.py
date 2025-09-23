from fastapi import FastAPI, Request
from fastapi.responses import HTMLResponse, FileResponse, JSONResponse
from fastapi.staticfiles import StaticFiles
from pydantic import BaseModel
from typing import List, Optional, Dict
import uvicorn
import time

app = FastAPI(title="CorridorOS DEV API")

# In-memory state
STATE = {
    "corridors": {},
    "ffm": {}
}

# ----- Models -----
class CorridorRequest(BaseModel):
    type: str = "SiCorridor"
    lanes: int
    lambda_nm: List[int]
    min_gbps: int
    latency_budget_ns: int
    reach_mm: int
    mode: str = "waveguide"
    qos: Dict[str, object] = {}
    attestation_required: bool = True

class RecalibrateRequest(BaseModel):
    target_ber: float = 1e-12
    ambient_profile: Optional[str] = "lab_default"

class FFMAllocRequest(BaseModel):
    bytes: int
    latency_class: str
    bandwidth_floor_GBs: int = 0
    persistence: str = "none"
    shareable: bool = True
    security_domain: Optional[str] = None

# ----- Health -----
@app.get("/health/corrd")
def health_corrd():
    return {"ok": True, "service": "corrd", "ts": time.time()}

@app.get("/health/memqosd")
def health_memqosd():
    return {"ok": True, "service": "memqosd", "ts": time.time()}

# ----- Photonic Corridors -----
@app.post("/v1/corridors")
def allocate_corridor(req: CorridorRequest):
    cid = f"cor-{len(STATE['corridors'])+1:04d}"
    reply = {
        "corridor_id": cid,
        "achievable_gbps": float(req.lanes) * 52.0,
        "ber": 1.0e-12,
        "eye_margin": "ok"
    }
    STATE["corridors"][cid] = {
        "req": req.dict(),
        "telemetry": {"ber": 1.0e-12, "tempC": 47.5, "power_pJ_per_bit": 0.9, "drift": "low"}
    }
    return reply

@app.get("/v1/corridors/{cid}/telemetry")
def corridor_telemetry(cid: str):
    telem = STATE["corridors"].get(cid, {}).get("telemetry")
    if not telem:
        return JSONResponse({"error": "not_found"}, status_code=404)
    return telem

@app.post("/v1/corridors/{cid}/recalibrate")
def corridor_recalibrate(cid: str, body: RecalibrateRequest):
    if cid not in STATE["corridors"]:
        return JSONResponse({"error": "not_found"}, status_code=404)
    # pretend to nudge bias / lambda
    lanes = STATE["corridors"][cid]["req"]["lanes"]
    return {
        "status": "converged",
        "new_bias_mV": [5.0]*lanes,
        "lambda_shifts_nm": [0.05]*lanes
    }

# ----- Free-Form Memory -----
@app.post("/v1/ffm/alloc")
def ffm_alloc(req: FFMAllocRequest):
    hid = f"ffm-{len(STATE['ffm'])+1:04d}"
    STATE["ffm"][hid] = {
        "req": req.dict(),
        "achieved_GBs": float(req.bandwidth_floor_GBs),
        "moved_pages": 0,
        "tail_p99_ms": 2.0
    }
    return {"ffm_handle": hid, "fds": ["/dev/null"], "policy_lease_ttl_s": 3600}

@app.patch("/v1/ffm/{hid}/bandwidth")
def ffm_bw(hid: str, body: dict):
    if hid not in STATE["ffm"]:
        return JSONResponse({"error":"not_found"}, status_code=404)
    floor = body.get("floor_GBs", 0)
    STATE["ffm"][hid]["achieved_GBs"] = max(STATE["ffm"][hid]["achieved_GBs"], float(floor))
    return {"ok": True, "achieved_GBs": STATE["ffm"][hid]["achieved_GBs"]}

@app.patch("/v1/ffm/{hid}/latency_class")
def ffm_lat(hid: str, body: dict):
    if hid not in STATE["ffm"]:
        return JSONResponse({"error":"not_found"}, status_code=404)
    STATE["ffm"][hid]["req"]["latency_class"] = body.get("target", STATE["ffm"][hid]["req"]["latency_class"])
    STATE["ffm"][hid]["moved_pages"] += 123456
    return {"ok": True, "target": STATE["ffm"][hid]["req"]["latency_class"]}

@app.get("/v1/ffm/{hid}/telemetry")
def ffm_telem(hid: str):
    if hid not in STATE["ffm"]:
        return JSONResponse({"error":"not_found"}, status_code=404)
    d = STATE["ffm"][hid]
    return {"achieved_GBs": d["achieved_GBs"], "moved_pages": d["moved_pages"], "tail_p99_ms": d["tail_p99_ms"]}

# ----- Static dashboard -----
app.mount("/static", StaticFiles(directory="static"), name="static")

@app.get("/", response_class=HTMLResponse)
@app.get("/corridoros_dashboard.html", response_class=HTMLResponse)
def dash():
    return FileResponse("static/corridoros_dashboard.html")

if __name__ == "__main__":
    uvicorn.run(app, host="0.0.0.0", port=8000)