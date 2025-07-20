package config

import (
	"os"
	"testing"
	"time"
)

func TestLoad(t *testing.T) {
	// Test default values
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if cfg.Port != 8080 {
		t.Errorf("Expected default port 8080, got %d", cfg.Port)
	}

	if cfg.Host != "0.0.0.0" {
		t.Errorf("Expected default host '0.0.0.0', got '%s'", cfg.Host)
	}

	if cfg.Debug {
		t.Error("Expected debug to be false by default")
	}

	if cfg.ReadTimeout != 15*time.Second {
		t.Errorf("Expected default read timeout 15s, got %v", cfg.ReadTimeout)
	}
}

func TestLoadWithEnvironment(t *testing.T) {
	// Set environment variables
	os.Setenv("PORT", "9000")
	os.Setenv("HOST", "127.0.0.1")
	os.Setenv("DEBUG", "true")
	os.Setenv("READ_TIMEOUT", "30s")
	os.Setenv("DATABASE_URL", "postgres://localhost/test")

	defer func() {
		os.Unsetenv("PORT")
		os.Unsetenv("HOST")
		os.Unsetenv("DEBUG")
		os.Unsetenv("READ_TIMEOUT")
		os.Unsetenv("DATABASE_URL")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if cfg.Port != 9000 {
		t.Errorf("Expected port 9000, got %d", cfg.Port)
	}

	if cfg.Host != "127.0.0.1" {
		t.Errorf("Expected host '127.0.0.1', got '%s'", cfg.Host)
	}

	if !cfg.Debug {
		t.Error("Expected debug to be true")
	}

	if cfg.ReadTimeout != 30*time.Second {
		t.Errorf("Expected read timeout 30s, got %v", cfg.ReadTimeout)
	}

	if cfg.DatabaseURL != "postgres://localhost/test" {
		t.Errorf("Expected database URL, got '%s'", cfg.DatabaseURL)
	}
}

func TestLoadInvalidPort(t *testing.T) {
	os.Setenv("PORT", "invalid")
	defer os.Unsetenv("PORT")

	_, err := Load()
	if err == nil {
		t.Error("Expected error for invalid port")
	}
}

func TestAddress(t *testing.T) {
	cfg := &Config{
		Host: "localhost",
		Port: 8080,
	}

	expected := "localhost:8080"
	if addr := cfg.Address(); addr != expected {
		t.Errorf("Expected address '%s', got '%s'", expected, addr)
	}
}
