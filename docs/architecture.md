# PipeForge — Architecture and System Design

## 1. Overview

**PipeForge** is an automated, production-inspired deployment reliability pipeline for a containerized Go HTTP service. It is designed to demonstrate safe deployment mechanisms, automated validation, observability endpoints, and reliable rollback capabilities.

## 2. End-to-End Release & Deployment Flow

```mermaid
flowchart TD
    subgraph Developer_Inner_Loop["Developer Workflow"]
        A["Feature / Bugfix Branch"] --> B["Local Verification (check.sh, go test)"]
        B --> C["Pull Request opened against 'main'"]
    end

    subgraph CI_Pipeline["GitHub Actions CI (.github/workflows/ci.yml)"]
        C --> D["Trigger: PR & Push to main"]
        D --> E["1. Code Formatting (gofmt)"]
        E --> F["2. Static Analysis (go vet)"]
        F --> G["3. Unit & Integration Tests"]
        G --> H["4. Binary Compilation Check"]
    end

    subgraph Release_Pipeline["Artifact & Delivery (.github/workflows/release.yml)"]
        H --> I["Release Tag (e.g. v1.0.0, sha-xxxx)"]
        I --> J["Multi-stage Docker Build"]
        J --> K["Security Scanning (Trivy)"]
        K --> L["Publish Immutable Digest & Tag to GHCR"]
    end

    subgraph CD_Deployment["Linux Host Deployment (deploy/deploy.sh)"]
        L --> M["deploy.sh receives candidate image tag"]
        M --> N["Verify Image Pull & Record Active Version"]
        N --> O["Start Candidate Container"]
        O --> P{"Smoke & Health Checks (/healthz, /readyz, /version)?"}
        P -- "PASS" --> Q["Promote Candidate & Archive Previous"]
        P -- "FAIL / Timeout" --> R["deploy/rollback.sh Auto-Invoked"]
        R --> S["Restore Previous Known-Good Release"]
        S --> T["Verify Restored Release & Emit Diagnostics"]
    end
```

## 3. Core Architecture Components

### 3.1 HTTP Service (`cmd/pipeforge-service`)
- **Language**: Go standard library (minimal third-party dependencies).
- **Core Endpoints**:
  - `GET /healthz`: Liveness probe — returns `200 OK` if the process is alive.
  - `GET /readyz`: Readiness probe — returns `200 OK` when the service is ready to accept user traffic.
  - `GET /version`: Service metadata — outputs commit hash, build version, and runtime build date.
- **Reliability Features**:
  - Configurable server timeouts (Read, Write, Idle).
  - Clean signal handling (`SIGINT`, `SIGTERM`) with graceful connection draining.
  - Structured JSON logging with request tracing.

### 3.2 Container Architecture
- **Multi-Stage Build**: Separates heavy Go compilation tooling from the minimal runtime image.
- **Security Posture**: Runs as an unprivileged, non-root user (`nonroot:nonroot`).
- **Container Health Check**: Evaluates container health without installing heavyweight debugging tools inside the production image.

### 3.3 Deployment & Rollback Strategy
- **Orchestration**: Docker Compose on a Linux host.
- **Verification Gates**:
  1. Image validation before container startup.
  2. Startup grace period followed by bounded polling of `/readyz`.
  3. Release smoke test verifying HTTP status and release version.
- **Rollback Guarantee**: In-flight deployments preserve the last-known-good release identifier. If health checks fail or exceed timeout thresholds, the previous version is restored immediately, verified, and diagnostic logs are persisted.

