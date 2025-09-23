#!/usr/bin/env bash
set -euo pipefail

# Offline-friendly build script.
# - Avoids adding new dependencies; builds existing components.
# - Uses cargo/go offline modes when caches are present.

root_dir="$(cd "$(dirname "$0")/.." && pwd)"

# Redirect Go caches to workspace to avoid seatbelt denials
export GOCACHE="$root_dir/.gocache"
export GOMODCACHE="$root_dir/.gomodcache"
mkdir -p "$GOCACHE" "$GOMODCACHE"

echo "[offline-build] Building corrd (Rust)"
(
  cd "$root_dir/daemon/corrd"
  if cargo --version | grep -q cargo; then
    if cargo build --help | grep -q -- --offline; then
      cargo build --release --offline || cargo build --release
    else
      cargo build --release
    fi
  fi
)

echo "[offline-build] Building Go daemons (best-effort offline)"
build_go_best_effort() {
  local dir="$1"; shift
  local bin="$1"; shift
  (
    cd "$dir"
    if GOFLAGS=${GOFLAGS:-} go build -o "$bin" .; then
      echo "built $bin"
    else
      if [[ -x "$bin" || -f "$bin" ]]; then
        echo "WARN: offline build failed in $dir, using existing binary $bin"
      else
        echo "ERROR: offline build failed in $dir and no existing binary found" >&2
        exit 1
      fi
    fi
  )
}

build_go_best_effort "$root_dir/daemon/memqosd" memqosd
build_go_best_effort "$root_dir/daemon/fabmand" fabmand
build_go_best_effort "$root_dir/daemon/heliopassd" heliopassd
build_go_best_effort "$root_dir/daemon/attestd" attestd

echo "[offline-build] Building CLIs (best-effort offline)"
(
  cd "$root_dir/cli"
  if go build -o corridor corridor/main.go; then echo built corridor; else echo "WARN: using existing CLI binary (cli/corridor/main)"; fi
  if go build -o ffm ffm/main.go; then echo built ffm; else echo "WARN: using existing CLI binary (cli/ffm/main)"; fi
)

echo "[offline-build] Building Labs (best-effort offline)"
build_go_best_effort "$root_dir/labs/physics-decoder" physics-decoder || true
build_go_best_effort "$root_dir/labs/helio-sim" helio-sim || true
(
  cd "$root_dir/labs/synchrony-analytics"
  go build -o synchrony-analytics .
)

echo "[offline-build] Done"
