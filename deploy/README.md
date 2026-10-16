# PipeForge Deployment and Rollback Automation

This directory contains production deployment scripts for orchestrating releases on a Linux host using Docker and Docker Compose.

## Scripts Overview

- **`deploy.sh <image> [version]`**:
  - Pulls and validates candidate image.
  - Snapshots currently active release in `current-release.env`.
  - Starts candidate release.
  - Runs `smoke-test.sh` bounded polling.
  - **Auto-rollback**: If smoke tests fail or time out, `rollback.sh` is invoked automatically to restore the previous release.
- **`smoke-test.sh [host] [port] [expected_version]`**:
  - Probes `/readyz` up to `MAX_RETRIES` (default: 10 attempts).
  - Probes `/healthz` for process liveness.
  - Probes `/version` and asserts that the running release matches the candidate version.
- **`rollback.sh`**:
  - Dumps failed candidate container logs to `deploy/diagnostics/failed_release_<timestamp>.log`.
  - Re-deploys the last known-good image from `current-release.env`.
  - Re-runs smoke tests against the restored release to verify recovery.

## State Management

`deploy/current-release.env` records:
- `ACTIVE_IMAGE`
- `ACTIVE_VERSION`
- `ACTIVE_DEPLOYED_AT`
- `PREVIOUS_IMAGE`
- `PREVIOUS_VERSION`
