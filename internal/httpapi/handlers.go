package httpapi

import (
	"encoding/json"
	"net/http"
	"sync/atomic"
)

// ServerState allows toggling readiness for testing failure and recovery scenarios.
type ServerState struct {
	ready atomic.Bool
}

func NewServerState() *ServerState {
	s := &ServerState{}
	s.ready.Store(true) // Ready by default
	return s
}

func (s *ServerState) SetReady(ready bool) {
	s.ready.Store(ready)
}

func (s *ServerState) IsReady() bool {
	return s.ready.Load()
}

// HealthzResponse defines payload for /healthz.
type HealthzResponse struct {
	Status string `json:"status"`
}

// ReadyzResponse defines payload for /readyz.
type ReadyzResponse struct {
	Status string `json:"status"`
}

// VersionResponse defines payload for /version.
type VersionResponse struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	BuildTime string `json:"build_time"`
}

// HealthzHandler returns 200 OK indicating the process is alive.
func HealthzHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(HealthzResponse{Status: "healthy"})
	}
}

// ReadyzHandler returns 200 OK if the service is ready for traffic, 503 if not.
func ReadyzHandler(state *ServerState) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if state != nil && !state.IsReady() {
			w.WriteHeader(http.StatusServiceUnavailable)
			_ = json.NewEncoder(w).Encode(ReadyzResponse{Status: "unready"})
			return
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(ReadyzResponse{Status: "ready"})
	}
}

// VersionHandler returns version, commit hash, and build timestamp.
func VersionHandler(version, commit, buildTime string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(VersionResponse{
			Version:   version,
			Commit:    commit,
			BuildTime: buildTime,
		})
	}
}
