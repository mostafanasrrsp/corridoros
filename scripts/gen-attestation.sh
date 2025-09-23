#!/usr/bin/env bash
set -euo pipefail

# Generate an attestation ticket via attestd for a device ID
# Usage: scripts/gen-attestation.sh <device_id>

ATTESTD_URL="${ATTESTD_URL:-http://localhost:8084}"

if [[ $# -lt 1 ]]; then
  echo "Usage: $0 <device_id>" >&2
  exit 1
fi

DEVICE_ID="$1"
BODY=$(cat <<JSON
{ "device_id": "${DEVICE_ID}", "require_pqc": true }
JSON
)

resp=$(curl -s -f -X POST -H 'Content-Type: application/json' \
  -d "${BODY}" "${ATTESTD_URL}/v1/attest/device")

echo "$resp" | sed -e 's/^/attestd: /'
echo "Ticket ID: $(echo "$resp" | sed -n 's/.*\"attestation_id\":\"\([^"]*\)\".*/\1/p')" || true

