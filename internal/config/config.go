package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds runtime configuration for the PipeForge HTTP service.
type Config struct {
	Host            string
	Port            int
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
	Version         string
	Commit          string
	BuildTime       string
}

// LoadFromEnv populates Config using environment variables with safe defaults.
func LoadFromEnv(version, commit, buildTime string) (*Config, error) {
	host := getEnv("PIPEFORGE_HOST", "0.0.0.0")

	portStr := getEnv("PIPEFORGE_PORT", "8080")
	port, err := strconv.Atoi(portStr)
	if err != nil || port < 1 || port > 65535 {
		return nil, fmt.Errorf("invalid PIPEFORGE_PORT %q: must be between 1 and 65535", portStr)
	}

	readTimeout, err := parseDurationEnv("PIPEFORGE_READ_TIMEOUT", 5*time.Second)
	if err != nil {
		return nil, err
	}

	writeTimeout, err := parseDurationEnv("PIPEFORGE_WRITE_TIMEOUT", 10*time.Second)
	if err != nil {
		return nil, err
	}

	idleTimeout, err := parseDurationEnv("PIPEFORGE_IDLE_TIMEOUT", 60*time.Second)
	if err != nil {
		return nil, err
	}

	shutdownTimeout, err := parseDurationEnv("PIPEFORGE_SHUTDOWN_TIMEOUT", 10*time.Second)
	if err != nil {
		return nil, err
	}

	if version == "" {
		version = "dev"
	}
	if commit == "" {
		commit = "none"
	}
	if buildTime == "" {
		buildTime = "unknown"
	}

	return &Config{
		Host:            host,
		Port:            port,
		ReadTimeout:     readTimeout,
		WriteTimeout:    writeTimeout,
		IdleTimeout:     idleTimeout,
		ShutdownTimeout: shutdownTimeout,
		Version:         version,
		Commit:          commit,
		BuildTime:       buildTime,
	}, nil
}

// ListenAddr returns the formatted host:port string for http.Server.
func (c *Config) ListenAddr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func parseDurationEnv(key string, defaultVal time.Duration) (time.Duration, error) {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal, nil
	}
	d, err := time.ParseDuration(val)
	if err != nil {
		return 0, fmt.Errorf("invalid duration for %s: %w", key, err)
	}
	if d <= 0 {
		return 0, fmt.Errorf("%s must be positive, got %v", key, d)
	}
	return d, nil
}
