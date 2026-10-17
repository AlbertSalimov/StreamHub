package config

import (
	"fmt"
	"os"
	"time"
)

type Config struct {
	HTTPAddr        string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
	DatabaseURL     string
}

func Load() (Config, error) {
	cfg := Config{
		HTTPAddr:    env("HTTP_ADDR", ":8080"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
	}

	var err error

	if cfg.ReadTimeout, err = durationEnv(
		"HTTP_READ_TIMEOUT", 5*time.Second,
	); err != nil {
		return Config{}, err
	}

	if cfg.WriteTimeout, err = durationEnv(
		"HTTP_WRITE_TIMEOUT", 10*time.Second,
	); err != nil {
		return Config{}, err
	}

	if cfg.IdleTimeout, err = durationEnv(
		"HTTP_IDLE_TIMEOUT", 60*time.Second,
	); err != nil {
		return Config{}, err
	}

	if cfg.ShutdownTimeout, err = durationEnv(
		"HTTP_SHUTDOWN_TIMEOUT", 10*time.Second,
	); err != nil {
		return Config{}, err
	}

	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}

	if cfg.ReadTimeout <= 0 ||
		cfg.WriteTimeout <= 0 ||
		cfg.IdleTimeout <= 0 ||
		cfg.ShutdownTimeout <= 0 {
		return Config{}, fmt.Errorf("HTTP timeouts must be positive")
	}

	return cfg, nil
}

func env(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}

func durationEnv(key string, fallback time.Duration) (time.Duration, error) {
	value := os.Getenv(key)
	if value == "" {
		return fallback, nil
	}

	duration, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("invalid %s: %w", key, err)
	}

	return duration, nil
}
