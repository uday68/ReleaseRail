#!/usr/bin/env bash
# deploy/smoke-test.sh - Smoke test and release validation gate
set -euo pipefail

HOST="${1:-127.0.0.1}"
PORT="${2:-8085}"
EXPECTED_VERSION="${3:-}"
MAX_RETRIES="${MAX_RETRIES:-10}"
RETRY_INTERVAL="${RETRY_INTERVAL:-2}"

BASE_URL="http://${HOST}:${PORT}"
echo "=========================================="
echo " Running Smoke Tests against ${BASE_URL}"
if [ -n "${EXPECTED_VERSION}" ]; then
  echo " Expected Version: ${EXPECTED_VERSION}"
fi
echo " Max Retries: ${MAX_RETRIES} (${RETRY_INTERVAL}s interval)"
echo "=========================================="

# 1. Bounded polling for Readiness (/readyz)
echo "==> Checking Readiness (/readyz)..."
READY=0
for i in $(seq 1 "${MAX_RETRIES}"); do
  HTTP_STATUS=$(curl -s -o /dev/null -w "%{http_code}" "${BASE_URL}/readyz" || true)
  if [ "${HTTP_STATUS}" -eq 200 ]; then
    echo "Service is ready! (attempt ${i}/${MAX_RETRIES})"
    READY=1
    break
  fi
  echo "Readiness probe returned HTTP ${HTTP_STATUS}, retrying in ${RETRY_INTERVAL}s... (${i}/${MAX_RETRIES})"
  sleep "${RETRY_INTERVAL}"
done

if [ "${READY}" -ne 1 ]; then
  echo "ERROR: Service failed readiness check within timeout threshold!"
  exit 1
fi

# 2. Check Liveness (/healthz)
echo "==> Checking Liveness (/healthz)..."
HEALTH_RESP=$(curl -s -f "${BASE_URL}/healthz")
if [[ "${HEALTH_RESP}" != *"healthy"* ]]; then
  echo "ERROR: Unexpected /healthz response: ${HEALTH_RESP}"
  exit 1
fi
echo "Liveness probe: PASS (200 OK)"

# 3. Check Version metadata (/version)
echo "==> Checking Version (/version)..."
VERSION_RESP=$(curl -s -f "${BASE_URL}/version")
echo "Version response: ${VERSION_RESP}"

if [ -n "${EXPECTED_VERSION}" ]; then
  if [[ "${VERSION_RESP}" != *"${EXPECTED_VERSION}"* ]]; then
    echo "ERROR: Version mismatch! Expected '${EXPECTED_VERSION}', got '${VERSION_RESP}'"
    exit 1
  fi
  echo "Version check: PASS (matches ${EXPECTED_VERSION})"
fi

echo "=========================================="
echo " ALL SMOKE TESTS PASSED!"
echo "=========================================="
exit 0
