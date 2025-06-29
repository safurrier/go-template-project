package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds application configuration.
type Config struct {
	Port         int           `json:"port"`
	Host         string        `json:"host"`
	Debug        bool          `json:"debug"`
	ReadTimeout  time.Duration `json:"read_timeout"`
	WriteTimeout time.Duration `json:"write_timeout"`
	DatabaseURL  string        `json:"database_url,omitempty"`
}

// Load creates a new configuration from environment variables.
func Load() (*Config, error) {
	cfg := &Config{
		Port:         8080,
		Host:         "0.0.0.0",
		Debug:        false,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	// Override with environment variables
	if port := os.Getenv("PORT"); port != "" {
		p, err := strconv.Atoi(port)
		if err != nil {
			return nil, fmt.Errorf("invalid PORT value: %w", err)
		}
		cfg.Port = p
	}

	if host := os.Getenv("HOST"); host != "" {
		cfg.Host = host
	}

	if debug := os.Getenv("DEBUG"); debug == "true" {
		cfg.Debug = true
	}

	if timeout := os.Getenv("READ_TIMEOUT"); timeout != "" {
		t, err := time.ParseDuration(timeout)
		if err != nil {
			return nil, fmt.Errorf("invalid READ timeout: %w", err)
		}
		cfg.ReadTimeout = t
	}

	if timeout := os.Getenv("WRITE_TIMEOUT"); timeout != "" {
		t, err := time.ParseDuration(timeout)
		if err != nil {
			return nil, fmt.Errorf("invalid write timeout: %w", err)
		}
		cfg.WriteTimeout = t
	}

	cfg.DatabaseURL = os.Getenv("DATABASE_URL")

	return cfg, nil
}

// Address returns the full address to bind to.
func (c *Config) Address() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}
