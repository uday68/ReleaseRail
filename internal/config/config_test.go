package config

import (
	"os"
	"testing"
	"time"
)

func TestLoadFromEnv_Defaults(t *testing.T) {
	// Ensure relevant environment variables are clean
	os.Unsetenv("PIPEFORGE_HOST")
	os.Unsetenv("PIPEFORGE_PORT")
	os.Unsetenv("PIPEFORGE_READ_TIMEOUT")

	cfg, err := LoadFromEnv("1.0.0", "abcdef", "2026-10-16T00:00:00Z")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Host != "0.0.0.0" {
		t.Errorf("expected default Host '0.0.0.0', got %q", cfg.Host)
	}
	if cfg.Port != 8080 {
		t.Errorf("expected default Port 8080, got %d", cfg.Port)
	}
	if cfg.ReadTimeout != 5*time.Second {
		t.Errorf("expected default ReadTimeout 5s, got %v", cfg.ReadTimeout)
	}
	if cfg.Version != "1.0.0" {
		t.Errorf("expected Version '1.0.0', got %q", cfg.Version)
	}
	if cfg.ListenAddr() != "0.0.0.0:8080" {
		t.Errorf("expected ListenAddr '0.0.0.0:8080', got %q", cfg.ListenAddr())
	}
}

func TestLoadFromEnv_Custom(t *testing.T) {
	t.Setenv("PIPEFORGE_HOST", "127.0.0.1")
	t.Setenv("PIPEFORGE_PORT", "9090")
	t.Setenv("PIPEFORGE_READ_TIMEOUT", "2s")

	cfg, err := LoadFromEnv("", "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Host != "127.0.0.1" {
		t.Errorf("expected Host '127.0.0.1', got %q", cfg.Host)
	}
	if cfg.Port != 9090 {
		t.Errorf("expected Port 9090, got %d", cfg.Port)
	}
	if cfg.ReadTimeout != 2*time.Second {
		t.Errorf("expected ReadTimeout 2s, got %v", cfg.ReadTimeout)
	}
	// Fallback check
	if cfg.Version != "dev" {
		t.Errorf("expected default Version 'dev', got %q", cfg.Version)
	}
}

func TestLoadFromEnv_InvalidPort(t *testing.T) {
	t.Setenv("PIPEFORGE_PORT", "invalid")
	_, err := LoadFromEnv("", "", "")
	if err == nil {
		t.Fatal("expected error for non-numeric port, got nil")
	}

	t.Setenv("PIPEFORGE_PORT", "70000")
	_, err = LoadFromEnv("", "", "")
	if err == nil {
		t.Fatal("expected error for out-of-range port, got nil")
	}
}

func TestLoadFromEnv_InvalidDuration(t *testing.T) {
	t.Setenv("PIPEFORGE_READ_TIMEOUT", "-5s")
	_, err := LoadFromEnv("", "", "")
	if err == nil {
		t.Fatal("expected error for negative duration, got nil")
	}
}
