package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds process configuration loaded from the environment.
type Config struct {
	DatabaseURL   string
	ListenAddr    string
	DBPingRetries int
	DBPingWait    time.Duration
}

// Load reads configuration from environment variables with safe defaults
// matching the Compose setup.
func Load() Config {
	return Config{
		DatabaseURL:   getenv("DATABASE_URL", "postgres://capacity:capacity@localhost:5432/capacity?sslmode=disable"),
		ListenAddr:    getenv("LISTEN_ADDR", ":8080"),
		DBPingRetries: getenvInt("DB_PING_RETRIES", 30),
		DBPingWait:    time.Duration(getenvInt("DB_PING_WAIT_SEC", 1)) * time.Second,
	}
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getenvInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}
