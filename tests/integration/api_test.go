package integration_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/uday/pipeforge/internal/config"
	"github.com/uday/pipeforge/internal/httpapi"
)

// setupTestServer starts a live in-process server on a random free port.
func setupTestServer(t *testing.T) (string, *httpapi.ServerState, func()) {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen on random port: %v", err)
	}

	port := listener.Addr().(*net.TCPAddr).Port
	cfg := &config.Config{
		Host:            "127.0.0.1",
		Port:            port,
		ReadTimeout:     5 * time.Second,
		WriteTimeout:    5 * time.Second,
		IdleTimeout:     30 * time.Second,
		ShutdownTimeout: 5 * time.Second,
		Version:         "integration-test-v1",
		Commit:          "git-integration-sha",
		BuildTime:       time.Now().UTC().Format(time.RFC3339),
	}

	state := httpapi.NewServerState()
	server := httpapi.NewServer(cfg, state)

	go func() {
		_ = server.Serve(listener)
	}()

	baseURL := fmt.Sprintf("http://127.0.0.1:%d", port)

	teardown := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = server.Shutdown(ctx)
	}

	return baseURL, state, teardown
}

func TestIntegration_EndToEndEndpoints(t *testing.T) {
	baseURL, state, teardown := setupTestServer(t)
	defer teardown()

	client := &http.Client{Timeout: 3 * time.Second}

	// 1. Verify /healthz
	resp, err := client.Get(baseURL + "/healthz")
	if err != nil {
		t.Fatalf("failed GET /healthz: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 from /healthz, got %d", resp.StatusCode)
	}

	// 2. Verify /version payload
	vResp, err := client.Get(baseURL + "/version")
	if err != nil {
		t.Fatalf("failed GET /version: %v", err)
	}
	defer vResp.Body.Close()

	var versionPayload struct {
		Version string `json:"version"`
		Commit  string `json:"commit"`
	}
	if err := json.NewDecoder(vResp.Body).Decode(&versionPayload); err != nil {
		t.Fatalf("failed to parse /version JSON: %v", err)
	}

	if versionPayload.Version != "integration-test-v1" {
		t.Errorf("expected version 'integration-test-v1', got %q", versionPayload.Version)
	}

	// 3. Verify /readyz transitions (Ready -> Unready -> Ready)
	rResp, err := client.Get(baseURL + "/readyz")
	if err != nil {
		t.Fatalf("failed GET /readyz: %v", err)
	}
	rResp.Body.Close()
	if rResp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 from /readyz initially, got %d", rResp.StatusCode)
	}

	// Flip to unready
	state.SetReady(false)
	uResp, err := client.Get(baseURL + "/readyz")
	if err != nil {
		t.Fatalf("failed GET /readyz when unready: %v", err)
	}
	uResp.Body.Close()
	if uResp.StatusCode != http.StatusServiceUnavailable {
		t.Errorf("expected 503 from /readyz when unready, got %d", uResp.StatusCode)
	}
}

func TestIntegration_ConcurrentRequestsUnderLoad(t *testing.T) {
	baseURL, _, teardown := setupTestServer(t)
	defer teardown()

	client := &http.Client{Timeout: 3 * time.Second}
	concurrency := 50
	var wg sync.WaitGroup
	errCh := make(chan error, concurrency)

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			endpoint := baseURL + "/healthz"
			if id%2 == 0 {
				endpoint = baseURL + "/version"
			}
			resp, err := client.Get(endpoint)
			if err != nil {
				errCh <- fmt.Errorf("worker %d failed request: %w", id, err)
				return
			}
			resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				errCh <- fmt.Errorf("worker %d got status %d", id, resp.StatusCode)
			}
		}(i)
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		t.Errorf("concurrency error: %v", err)
	}
}
