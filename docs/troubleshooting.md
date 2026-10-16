# PipeForge — Operational Troubleshooting & Runbook

This runbook guides on-call and systems engineers through diagnosing service and deployment failures using real system evidence.

---

## Scenario 1: Port Collision / Address Already in Use

### Symptom
- Running `pipeforge-service` returns `bind: address already in use` or incoming requests return unexpected responses (e.g. 404 from another web server).
### Diagnosis
```bash
# On Linux / WSL:
netstat -tulpn | grep :8080
# Or using ss:
ss -tulpn | grep :8080

# On Windows:
netstat -ano | findstr 8080
Get-Process -Id <PID>
```
### Resolution
- Update the listening port via environment variable without recompiling:
```bash
export PIPEFORGE_PORT=8085
```

---

## Scenario 2: Candidate Deployment Fails Smoke Test & Auto-Rollback Occurs

### Symptom
- `deploy.sh` outputs `Release verification FAILED` and executes rollback.
### Diagnosis
1. Inspect the preserved diagnostic logs:
```bash
cat deploy/diagnostics/failed_release_*.log
```
2. Check `deploy/current-release.env` to confirm the restored image version.
3. Test container manually in an isolated namespace:
```bash
docker run --rm -it -e PIPEFORGE_PORT=8080 <candidate-image>
```

---

## Scenario 3: Container Exits with Permission Denied

### Symptom
- Container fails immediately upon start with a permissions error.
### Diagnosis
- The image runs as non-root UID `10001:10001`. If application directories or files are owned by `root`, the service cannot access them.
- Check user identity:
```bash
docker run --rm --entrypoint id pipeforge-service:latest
```
### Resolution
- Ensure directories created in `Dockerfile` grant read/write access to `nonroot:nonroot` or run without local file writing dependencies.
