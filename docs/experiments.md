# PipeForge — Empirical Engineering Experiments Log

This document records the exact results of tests and failure drills executed during the development and verification of PipeForge.

---

## Experiment 1: Host Port Collision Diagnosis
- **Trigger**: Starting `pipeforge-service` on default port `8080`.
- **Observed Result**: `curl http://127.0.0.1:8080/healthz` returned Apache HTTP Server `404 Not Found`.
- **Evidence**: `netstat -ano | findstr 8080` revealed PID `5348 (httpd)` listening on `8080`.
- **Recovery / Solution**: Overrode port via 12-factor configuration: `PIPEFORGE_PORT=8085`.
- **Proof**: `curl http://127.0.0.1:8085/healthz` returned `{"status":"healthy"}` and `200 OK`.

---

## Experiment 2: Cross-Environment Toolchain Compatibility
- **Trigger**: Multi-stage Docker build invoking `go mod download`.
- **Observed Result**: Build failed with `go.mod requires go >= 1.26.1 (running go 1.22.12)`.
- **Evidence**: Windows Go version was `1.26.1`, but the standard container base was `golang:1.22-alpine`.
- **Recovery / Solution**: Adjusted `go.mod` directive to `go 1.22.0`.
- **Proof**: `docker build` succeeded and produced a minimal `~30MB` container image.

---

## Experiment 3: Automated Candidate Rejection & Release Rollback
- **Precondition**: `pipeforge-service:1.0.0` was deployed and verified active (`deploy/current-release.env`).
- **Trigger**: Deploying candidate image `pipeforge-service:1.0.1-broken` via `./deploy/deploy.sh pipeforge-service:1.0.1-broken 1.0.1`.
- **Observed Result**:
  1. `smoke-test.sh` polled `/readyz` 10 times and timed out (`HTTP 000`).
  2. `deploy.sh` detected failure and triggered `deploy/rollback.sh`.
  3. `rollback.sh` saved broken container logs to `deploy/diagnostics/failed_release_20261016_051344.log`.
  4. Previous release `pipeforge-service:1.0.0` was redeployed automatically.
  5. Restored container passed smoke tests (`/readyz`, `/healthz`, `/version` verified `1.0.0`).
  6. Script exited with code 1, rejecting the broken release.
- **Proof**: `deploy/current-release.env` confirmed `ACTIVE_VERSION="1.0.0"`. Service remained 100% available.
