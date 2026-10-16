package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthzHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rr := httptest.NewRecorder()

	handler := HealthzHandler()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var resp HealthzResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode JSON: %v", err)
	}

	if resp.Status != "healthy" {
		t.Errorf("expected status 'healthy', got %q", resp.Status)
	}
}

func TestReadyzHandler(t *testing.T) {
	state := NewServerState()

	// 1. Ready state (default)
	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	rr := httptest.NewRecorder()
	handler := ReadyzHandler(state)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200 when ready, got %d", rr.Code)
	}

	var resp ReadyzResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode JSON: %v", err)
	}
	if resp.Status != "ready" {
		t.Errorf("expected status 'ready', got %q", resp.Status)
	}

	// 2. Unready state (simulating failure / warming up)
	state.SetReady(false)
	rrUnready := httptest.NewRecorder()
	handler.ServeHTTP(rrUnready, req)

	if rrUnready.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status 503 when unready, got %d", rrUnready.Code)
	}
}

func TestVersionHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/version", nil)
	rr := httptest.NewRecorder()

	handler := VersionHandler("v1.2.3", "commit-abc", "2026-10-16T12:00:00Z")
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var resp VersionResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode JSON: %v", err)
	}

	if resp.Version != "v1.2.3" || resp.Commit != "commit-abc" || resp.BuildTime != "2026-10-16T12:00:00Z" {
		t.Errorf("unexpected version response payload: %+v", resp)
	}
}
