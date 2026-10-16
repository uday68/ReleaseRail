package httpapi

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/uday/pipeforge/internal/config"
)

// statusRecorder wraps http.ResponseWriter to capture HTTP status code for structured logging.
type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.statusCode = code
	r.ResponseWriter.WriteHeader(code)
}

// LoggingMiddleware logs every request as structured JSON using log/slog.
func LoggingMiddleware(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(rec, r)

		logger.Info("http request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", rec.statusCode,
			"duration_ms", time.Since(start).Milliseconds(),
			"remote_addr", r.RemoteAddr,
			"user_agent", r.UserAgent(),
		)
	})
}

// NewServer configures and returns an *http.Server instance and its router.
func NewServer(cfg *config.Config, state *ServerState) *http.Server {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", HealthzHandler())
	mux.HandleFunc("GET /readyz", ReadyzHandler(state))
	mux.HandleFunc("GET /version", VersionHandler(cfg.Version, cfg.Commit, cfg.BuildTime))

	// Catch-all for API root or documentation
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"service": "pipeforge",
			"docs":    "Endpoints: /healthz, /readyz, /version",
		})
	})

	loggedMux := LoggingMiddleware(logger, mux)

	return &http.Server{
		Addr:         cfg.ListenAddr(),
		Handler:      loggedMux,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}
}
