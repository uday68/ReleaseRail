#!/usr/bin/env bash
# deploy/deploy.sh - Zero-Downtime Blue/Green Deployment Orchestrator
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
STATE_FILE="${SCRIPT_DIR}/current-release.env"
UPSTREAM_FILE="${SCRIPT_DIR}/nginx/upstream.conf"

IMAGE="${1:-}"
VERSION="${2:-}"
HOST_PORT="${HOST_PORT:-8085}"

if [ -z "${IMAGE}" ]; then
  echo "Usage: $0 <image-reference> [version-tag]"
  echo "Example: $0 pipeforge-service:1.0.0 1.0.0"
  exit 1
fi

echo "=========================================="
echo " Starting Zero-Downtime Blue/Green Deploy"
echo " Candidate Image:   ${IMAGE}"
echo " Candidate Version: ${VERSION:-auto}"
echo " Host Ingress Port: ${HOST_PORT}"
echo "=========================================="

# 1. Verify image exists locally or attempt pull
if ! docker image inspect "${IMAGE}" >/dev/null 2>&1; then
  echo "==> Pulling image ${IMAGE}..."
  if ! docker pull "${IMAGE}"; then
    echo "ERROR: Failed to pull ${IMAGE}"
    exit 1
  fi
fi

# 2. Determine currently active and idle target color
ACTIVE_COLOR="none"
ACTIVE_IMAGE=""
ACTIVE_VERSION=""

if [ -f "${STATE_FILE}" ]; then
  # shellcheck source=/dev/null
  source "${STATE_FILE}"
  ACTIVE_COLOR="${ACTIVE_COLOR:-blue}"
  ACTIVE_IMAGE="${ACTIVE_IMAGE:-}"
  ACTIVE_VERSION="${ACTIVE_VERSION:-}"
fi

if [ "${ACTIVE_COLOR}" == "blue" ]; then
  TARGET_COLOR="green"
  TARGET_PORT=8082
  TARGET_SERVICE="pipeforge-green"
else
  TARGET_COLOR="blue"
  TARGET_PORT=8081
  TARGET_SERVICE="pipeforge-blue"
fi

echo "==> Currently Active: ${ACTIVE_COLOR} (${ACTIVE_IMAGE:-none})"
echo "==> Deploying Target: ${TARGET_COLOR} on internal port ${TARGET_PORT}"

# 3. Start candidate on target environment
if [ "${TARGET_COLOR}" == "blue" ]; then
  BLUE_IMAGE="${IMAGE}" docker compose -f "${PROJECT_ROOT}/compose.yaml" up -d --no-recreate pipeforge-blue
else
  GREEN_IMAGE="${IMAGE}" docker compose -f "${PROJECT_ROOT}/compose.yaml" up -d --no-recreate pipeforge-green
fi

# 4. Verify proxy container is running
if ! docker ps --format '{{.Names}}' | grep -q "^pipeforge-proxy$"; then
  echo "==> Starting ingress reverse proxy..."
  docker compose -f "${PROJECT_ROOT}/compose.yaml" up -d proxy
fi

# 5. Run verification gate against candidate instance
echo "==> Running candidate smoke tests on target ${TARGET_COLOR} (port ${TARGET_PORT})..."
if "${SCRIPT_DIR}/smoke-test.sh" "127.0.0.1" "${TARGET_PORT}" "${VERSION}"; then
  echo "==> Candidate passed verification! Switching traffic to ${TARGET_COLOR}..."

  # Update upstream configuration
  cat <<EOF > "${UPSTREAM_FILE}"
upstream pipeforge_backend {
    server ${TARGET_SERVICE}:8080;
}
EOF

  # Reload Nginx without dropping connections
  docker exec pipeforge-proxy nginx -s reload
  echo "==> Traffic switched to ${TARGET_COLOR} with ZERO DOWNTIME."

  # Stop previous idle environment
  if [ "${ACTIVE_COLOR}" != "none" ]; then
    echo "==> Stopping previous environment (${ACTIVE_COLOR})..."
    if [ "${ACTIVE_COLOR}" == "blue" ]; then
      docker compose -f "${PROJECT_ROOT}/compose.yaml" stop pipeforge-blue || true
    else
      docker compose -f "${PROJECT_ROOT}/compose.yaml" stop pipeforge-green || true
    fi
  fi

  # Update state file
  cat <<EOF > "${STATE_FILE}"
ACTIVE_COLOR="${TARGET_COLOR}"
ACTIVE_IMAGE="${IMAGE}"
ACTIVE_VERSION="${VERSION}"
ACTIVE_DEPLOYED_AT="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
PREVIOUS_COLOR="${ACTIVE_COLOR}"
PREVIOUS_IMAGE="${ACTIVE_IMAGE}"
PREVIOUS_VERSION="${ACTIVE_VERSION}"
EOF

  echo "=========================================="
  echo " BLUE/GREEN DEPLOYMENT SUCCEEDED!"
  echo " Active: ${TARGET_COLOR} (${IMAGE})"
  echo "=========================================="
  exit 0
else
  echo "==> ERROR: Candidate verification FAILED on ${TARGET_COLOR}!"
  echo "==> Preserving diagnostic logs..."
  mkdir -p "${SCRIPT_DIR}/diagnostics"
  TIMESTAMP=$(date +%Y%m%d_%H%M%S)
  docker logs "${TARGET_SERVICE}" > "${SCRIPT_DIR}/diagnostics/failed_candidate_${TARGET_COLOR}_${TIMESTAMP}.log" 2>&1 || true

  echo "==> Tearing down failed candidate (${TARGET_COLOR})..."
  docker compose -f "${PROJECT_ROOT}/compose.yaml" stop "${TARGET_SERVICE}" || true

  echo "==> Traffic remains safely on ${ACTIVE_COLOR} (${ACTIVE_IMAGE}). Zero downtime impact!"
  echo "=========================================="
  echo " DEPLOYMENT REJECTED — ACTIVE STATE PRESERVED"
  echo "=========================================="
  exit 1
fi
