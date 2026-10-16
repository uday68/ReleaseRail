#!/usr/bin/env bash
# deploy/rollback.sh - Automated rollback to previous known-good release
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
STATE_FILE="${SCRIPT_DIR}/current-release.env"
DIAGNOSTIC_DIR="${SCRIPT_DIR}/diagnostics"

echo "=========================================="
echo " INITIATING AUTOMATED RELEASE ROLLBACK"
echo "=========================================="

mkdir -p "${DIAGNOSTIC_DIR}"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)

# Collect failure diagnostics from current container before removing it
if docker ps -a --format '{{.Names}}' | grep -q "^pipeforge-app$"; then
  echo "==> Capturing failed container logs for root-cause analysis..."
  docker logs pipeforge-app > "${DIAGNOSTIC_DIR}/failed_release_${TIMESTAMP}.log" 2>&1 || true
  echo "Saved failure logs to ${DIAGNOSTIC_DIR}/failed_release_${TIMESTAMP}.log"
fi

if [ ! -f "${STATE_FILE}" ]; then
  echo "FATAL: State file ${STATE_FILE} not found! Cannot determine previous release."
  exit 1
fi

# Load previous known-good state
# shellcheck source=/dev/null
source "${STATE_FILE}"

if [ -z "${PREVIOUS_IMAGE:-}" ]; then
  echo "FATAL: No PREVIOUS_IMAGE recorded in state file! Rollback aborted."
  exit 1
fi

echo "==> Restoring previous known-good image: ${PREVIOUS_IMAGE} (version: ${PREVIOUS_VERSION:-unknown})"

# Stop failed container
docker compose -f "${PROJECT_ROOT}/compose.yaml" down --timeout 5 || true

# Start previous release
PIPEFORGE_IMAGE="${PREVIOUS_IMAGE}" \
RELEASE_VERSION="${PREVIOUS_VERSION:-restored}" \
HOST_PORT="${HOST_PORT:-8085}" \
docker compose -f "${PROJECT_ROOT}/compose.yaml" up -d

echo "==> Verifying restored release health..."
if "${SCRIPT_DIR}/smoke-test.sh" "127.0.0.1" "${HOST_PORT:-8085}" "${PREVIOUS_VERSION:-}"; then
  echo "SUCCESS: Rollback complete. Restored version ${PREVIOUS_VERSION:-} is healthy and serving traffic."
  # Update state file to reflect restored image as active
  cat <<EOF > "${STATE_FILE}"
ACTIVE_IMAGE="${PREVIOUS_IMAGE}"
ACTIVE_VERSION="${PREVIOUS_VERSION:-unknown}"
ACTIVE_DEPLOYED_AT="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
PREVIOUS_IMAGE=""
PREVIOUS_VERSION=""
EOF
  exit 0
else
  echo "CRITICAL: Restored release failed smoke test! Manual intervention required."
  exit 2
fi
