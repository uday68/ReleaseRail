package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/uday/pipeforge/internal/config"
	"github.com/uday/pipeforge/internal/httpapi"
)

// Build-time variables injected via -ldflags during CI/release builds
var (
	version   = "dev"
	commit    = "none"
	buildTime = "unknown"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	logger.Info("starting pipeforge-service",
		"version", version,
		"commit", commit,
		"build_time", buildTime,
	)

	cfg, err := config.LoadFromEnv(version, commit, buildTime)
	if err != nil {
		logger.Error("failed to load configuration", "error", err)
		os.Exit(1)
	}

	state := httpapi.NewServerState()
	server := httpapi.NewServer(cfg, state)

	// Channel to listen for OS interrupt / termination signals
	shutdownSignal := make(chan os.Signal, 1)
	signal.Notify(shutdownSignal, os.Interrupt, syscall.SIGTERM)

	// Start server in background goroutine
	serverErr := make(chan error, 1)
	go func() {
		logger.Info("http server listening", "addr", cfg.ListenAddr())
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	// Wait for OS shutdown signal or unexpected server failure
	select {
	case err := <-serverErr:
		logger.Error("server startup/runtime error", "error", err)
		os.Exit(1)
	case sig := <-shutdownSignal:
		logger.Info("shutdown signal received", "signal", sig.String())
	}

	// Flip readiness to false immediately so load balancers stop sending traffic
	state.SetReady(false)
	logger.Info("marked service unready, draining active connections")

	// Create context with bounded timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("graceful shutdown failed", "error", err)
		_ = server.Close()
		os.Exit(1)
	}

	logger.Info("pipeforge-service stopped cleanly")
}
