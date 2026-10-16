#!/usr/bin/env bash
# deploy/deploy.sh - Automated deployment orchestrator with rollback gate
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
STATE_FILE="${SCRIPT_DIR}/current-release.env"

IMAGE="${1:-}"
VERSION="${2:-}"
HOST_PORT="${HOST_PORT:-8085}"

if [ -z "${IMAGE}" ]; then
  echo "Usage: $0 <image-reference> [version-tag]"
  echo "Example: $0 pipeforge-service:1.0.0 1.0.0"
  exit 1
fi

echo "=========================================="
echo " Starting PipeForge Deployment"
echo " Target Image:   ${IMAGE}"
echo " Target Version: ${VERSION:-auto}"
echo " Host Port:      ${HOST_PORT}"
echo "=========================================="

# 1. Verify image exists locally or attempt pull
echo "==> Verifying image availability..."
if ! docker image inspect "${IMAGE}" >/dev/null 2>&1; then
  echo "Image ${IMAGE} not present locally. Attempting pull..."
  if ! docker pull "${IMAGE}"; then
    echo "ERROR: Failed to pull image ${IMAGE}. Deployment aborted."
    exit 1
  fi
fi
echo "Image verified."

# 2. Record currently running release
CURRENT_IMAGE=""
CURRENT_VERSION=""
if [ -f "${STATE_FILE}" ]; then
  # shellcheck source=/dev/null
  source "${STATE_FILE}"
  CURRENT_IMAGE="${ACTIVE_IMAGE:-}"
  CURRENT_VERSION="${ACTIVE_VERSION:-}"
  echo "Currently active release: ${CURRENT_IMAGE:-none} (version: ${CURRENT_VERSION:-none})"
fi

# Write candidate transition state
cat <<EOF > "${STATE_FILE}"
ACTIVE_IMAGE="${CURRENT_IMAGE}"
ACTIVE_VERSION="${CURRENT_VERSION}"
PREVIOUS_IMAGE="${CURRENT_IMAGE}"
PREVIOUS_VERSION="${CURRENT_VERSION}"
CANDIDATE_IMAGE="${IMAGE}"
CANDIDATE_VERSION="${VERSION}"
DEPLOYMENT_STARTED_AT="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
EOF

# 3. Deploy Candidate Container
echo "==> Starting candidate container..."
PIPEFORGE_IMAGE="${IMAGE}" \
RELEASE_VERSION="${VERSION}" \
HOST_PORT="${HOST_PORT}" \
docker compose -f "${PROJECT_ROOT}/compose.yaml" up -d --remove-orphans

# 4. Run Smoke Tests & Verification Gate
echo "==> Running release verification gate..."
if "${SCRIPT_DIR}/smoke-test.sh" "127.0.0.1" "${HOST_PORT}" "${VERSION}"; then
  echo "==> Release verification PASSED! Promoting candidate release."
  cat <<EOF > "${STATE_FILE}"
ACTIVE_IMAGE="${IMAGE}"
ACTIVE_VERSION="${VERSION}"
ACTIVE_DEPLOYED_AT="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
PREVIOUS_IMAGE="${CURRENT_IMAGE}"
PREVIOUS_VERSION="${CURRENT_VERSION}"
EOF
  echo "=========================================="
  echo " DEPLOYMENT SUCCEEDED: ${IMAGE}"
  echo "=========================================="
  exit 0
else
  echo "==> ERROR: Release verification FAILED for candidate ${IMAGE}!"
  if [ -n "${CURRENT_IMAGE}" ]; then
    echo "==> Triggering automated rollback to previous known-good release (${CURRENT_IMAGE})..."
    "${SCRIPT_DIR}/rollback.sh"
    echo "Rollback successfully restored ${CURRENT_IMAGE}."
  else
    echo "WARNING: No previous release existed to roll back to."
  fi
  echo "=========================================="
  echo " DEPLOYMENT FAILED AND ROLLED BACK"
  echo "=========================================="
  exit 1
fi
