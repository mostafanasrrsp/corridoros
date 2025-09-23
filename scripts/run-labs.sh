#!/usr/bin/env bash
set -euo pipefail

# Start/stop/status for CorridorLabs services.
# Services:
#  - physics-decoder (port 8085)
#  - helio-sim (port 8086)
#  - synchrony-analytics (port 8090)

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
LABS=("physics-decoder" "helio-sim" "synchrony-analytics")
PORTS=(8085 8086 8090)

log() { echo "[labs] $*"; }

start_one() {
  local name=$1
  local port=$2
  log "Starting $name on :$port"
  ("$ROOT_DIR/labs/$name/$name" & echo $! > "/tmp/corridoros-$name.pid")
  # Quick health check (best-effort)
  for i in $(seq 1 20); do
    if curl -sf "http://127.0.0.1:$port/health" >/dev/null 2>&1; then
      log "$name is up (port $port)"
      return 0
    fi
    sleep 0.3
  done
  log "WARN: $name did not respond to health check yet"
}

stop_one() {
  local name=$1
  local pid_file="/tmp/corridoros-$name.pid"
  if [[ -f "$pid_file" ]]; then
    local pid
    pid=$(cat "$pid_file")
    if kill -0 "$pid" 2>/dev/null; then
      log "Stopping $name (PID $pid)"
      kill "$pid" || true
    fi
    rm -f "$pid_file"
  fi
}

status() {
  for idx in "${!LABS[@]}"; do
    local name=${LABS[$idx]}
    local port=${PORTS[$idx]}
    if curl -sf "http://127.0.0.1:$port/health" >/dev/null 2>&1; then
      echo "  $name: RUNNING (:$port)"
    else
      echo "  $name: STOPPED  (:$port)"
    fi
  done
}

case "${1:-start}" in
  start)
    for idx in "${!LABS[@]}"; do
      start_one "${LABS[$idx]}" "${PORTS[$idx]}"
    done
    ;;
  stop)
    for name in "${LABS[@]}"; do
      stop_one "$name"
    done
    ;;
  status)
    status
    ;;
  *)
    echo "Usage: $0 {start|stop|status}"
    exit 1
    ;;
esac

