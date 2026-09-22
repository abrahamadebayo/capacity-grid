package config_test

import (
	"strings"
	"testing"

	"capacity/api/config"
)

func TestLoadDefaults(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	t.Setenv("LISTEN_ADDR", "")
	t.Setenv("DB_PING_RETRIES", "")
	t.Setenv("DB_PING_WAIT_SEC", "")

	cfg := config.Load()
	if cfg.ListenAddr != ":8080" {
		t.Errorf("ListenAddr = %q", cfg.ListenAddr)
	}
	if cfg.DBPingRetries != 30 {
		t.Errorf("DBPingRetries = %d", cfg.DBPingRetries)
	}
	if !strings.Contains(cfg.DatabaseURL, "capacity") {
		t.Errorf("unexpected DSN: %s", cfg.DatabaseURL)
	}
}

func TestLoadOverrides(t *testing.T) {
	t.Setenv("LISTEN_ADDR", ":9090")
	t.Setenv("DB_PING_RETRIES", "5")
	cfg := config.Load()
	if cfg.ListenAddr != ":9090" {
		t.Errorf("ListenAddr = %q", cfg.ListenAddr)
	}
	if cfg.DBPingRetries != 5 {
		t.Errorf("DBPingRetries = %d", cfg.DBPingRetries)
	}
}
