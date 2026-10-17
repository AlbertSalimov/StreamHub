package config

import (
	"testing"
	"time"
)

func TestLoadRequiresDatabaseURL(t *testing.T) {
	t.Setenv("DATABASE_URL", "")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error when DATABASE_URL is missing")
	}
}

func TestLoadUsesDefaultTimeouts(t *testing.T) {
	t.Setenv("DATABASE_URL",
		"postgres://user:password@localhost:5432/streamhub?sslmode=disable")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.HTTPAddr != ":8080" {
		t.Errorf("HTTPAddr = %q, want %q", cfg.HTTPAddr, ":8080")
	}

	if cfg.ReadTimeout != 5*time.Second {
		t.Errorf("ReadTimeout = %v, want 5s", cfg.ReadTimeout)
	}
}

func TestLoadRejectsInvalidTimeout(t *testing.T) {
	t.Setenv("DATABASE_URL",
		"postgres://user:password@localhost:5432/streamhub?sslmode=disable")
	t.Setenv("HTTP_READ_TIMEOUT", "invalid")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for invalid timeout")
	}
}
