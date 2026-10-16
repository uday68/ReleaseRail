#!/usr/bin/env bash
# scripts/check.sh - Local validation script matching GitHub Actions CI
set -euo pipefail

echo "======================================"
echo " Running PipeForge Local Verifications"
echo "======================================"

echo "==> 1. Checking Go formatting (gofmt)..."
UNFORMATTED=$(gofmt -l .)
if [ -n "$UNFORMATTED" ]; then
    echo "ERROR: Unformatted Go files found:"
    echo "$UNFORMATTED"
    echo "Run 'gofmt -w .' or 'make fmt' to fix."
    exit 1
fi
echo "Formatting OK."

echo "==> 2. Running static analysis (go vet)..."
go vet ./...
echo "Static analysis OK."

echo "==> 3. Running unit tests with race detector..."
go test -v -race ./...
echo "Unit tests OK."

echo "==> 4. Verifying binary compilation..."
mkdir -p bin
go build -ldflags "-X main.version=check -X main.commit=local -X main.buildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ)" -o bin/pipeforge-service ./cmd/pipeforge-service
echo "Compilation OK."

echo "======================================"
echo " ALL CHECKS PASSED SUCCESSFULLY!"
echo "======================================"
