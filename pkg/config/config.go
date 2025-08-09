package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port         string
	DatabasePath string
	RedisAddr    string
	RedisDB      int

	// API Configuration
	APITimeout     time.Duration
	MaxConcurrency int
	CacheTTL       map[string]time.Duration

	// Rate Limiting
	RateLimit       int
	RateLimitWindow time.Duration

	// Health Check
	HealthCheckInterval time.Duration
}

func Load() *Config {
	cfg := &Config{
		Port:         getEnv("PORT", "8080"),
		DatabasePath: getEnv("DATABASE_PATH", "./data.db"),
		RedisAddr:    getEnv("REDIS_ADDR", "localhost:6379"),
		RedisDB:      getEnvInt("REDIS_DB", 0),

		APITimeout:     getEnvDuration("API_TIMEOUT", 20*time.Second),
		MaxConcurrency: getEnvInt("MAX_CONCURRENCY", 10),

		RateLimit:       getEnvInt("RATE_LIMIT", 100),
		RateLimitWindow: getEnvDuration("RATE_LIMIT_WINDOW", time.Minute),

		HealthCheckInterval: getEnvDuration("HEALTH_CHECK_INTERVAL", 10*time.Minute),
	}

	// Set default cache TTL for different endpoints
	cfg.CacheTTL = map[string]time.Duration{
		"/api/v1/home":           15 * time.Minute,
		"/api/v1/jadwal-rilis":   30 * time.Minute,
		"/api/v1/anime-terbaru":  15 * time.Minute,
		"/api/v1/movie":          1 * time.Hour,
		"/api/v1/anime-detail":   1 * time.Hour,
		"/api/v1/episode-detail": 30 * time.Minute,
		"/api/v1/search":         10 * time.Minute,
	}

	return cfg
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
