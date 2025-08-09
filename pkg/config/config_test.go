package config

import (
	"os"
	"testing"
	"time"
)

func TestLoadConfig(t *testing.T) {
	// Clear environment variables to test defaults
	originalPort := os.Getenv("PORT")
	os.Unsetenv("PORT")
	defer func() {
		if originalPort != "" {
			os.Setenv("PORT", originalPort)
		}
	}()

	// Test default values
	cfg := Load()

	if cfg.Port != "8080" {
		t.Errorf("Expected default port 8080, got %s", cfg.Port)
	}

	if cfg.DatabasePath != "./data.db" {
		t.Errorf("Expected default database path ./data.db, got %s", cfg.DatabasePath)
	}

	if cfg.APITimeout != 20*time.Second {
		t.Errorf("Expected default API timeout 20s, got %v", cfg.APITimeout)
	}

	if cfg.RateLimit != 100 {
		t.Errorf("Expected default rate limit 100, got %d", cfg.RateLimit)
	}
}

func TestLoadConfigWithEnvVars(t *testing.T) {
	// Set environment variables
	os.Setenv("PORT", "9090")
	os.Setenv("DATABASE_PATH", "/tmp/test.db")
	os.Setenv("API_TIMEOUT", "30s")
	os.Setenv("RATE_LIMIT", "200")

	defer func() {
		os.Unsetenv("PORT")
		os.Unsetenv("DATABASE_PATH")
		os.Unsetenv("API_TIMEOUT")
		os.Unsetenv("RATE_LIMIT")
	}()

	cfg := Load()

	if cfg.Port != "9090" {
		t.Errorf("Expected port 9090, got %s", cfg.Port)
	}

	if cfg.DatabasePath != "/tmp/test.db" {
		t.Errorf("Expected database path /tmp/test.db, got %s", cfg.DatabasePath)
	}

	if cfg.APITimeout != 30*time.Second {
		t.Errorf("Expected API timeout 30s, got %v", cfg.APITimeout)
	}

	if cfg.RateLimit != 200 {
		t.Errorf("Expected rate limit 200, got %d", cfg.RateLimit)
	}
}

func TestCacheTTLConfiguration(t *testing.T) {
	cfg := Load()

	expectedTTLs := map[string]time.Duration{
		"/api/v1/home":           15 * time.Minute,
		"/api/v1/jadwal-rilis":   30 * time.Minute,
		"/api/v1/anime-terbaru":  15 * time.Minute,
		"/api/v1/movie":          1 * time.Hour,
		"/api/v1/anime-detail":   1 * time.Hour,
		"/api/v1/episode-detail": 30 * time.Minute,
		"/api/v1/search":         10 * time.Minute,
	}

	for endpoint, expectedTTL := range expectedTTLs {
		if cfg.CacheTTL[endpoint] != expectedTTL {
			t.Errorf("Expected TTL for %s to be %v, got %v", endpoint, expectedTTL, cfg.CacheTTL[endpoint])
		}
	}
}
