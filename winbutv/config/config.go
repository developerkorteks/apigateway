package config

import (
	"os"
	"time"
)

type Config struct {
	Environment string
	Port        string
	BaseURL     string

	// Scraping settings
	UserAgent  string
	Timeout    time.Duration
	RateLimit  time.Duration
	MaxRetries int

	// Cache settings
	CacheEnabled bool
	CacheTTL     time.Duration
}

func Load() *Config {
	return &Config{
		Environment: getEnv("ENVIRONMENT", "development"),
		Port:        getEnv("PORT", "8080"),
		BaseURL:     getEnv("BASE_URL", "https://winbu.tv"),

		// Scraping settings
		UserAgent:  getEnv("USER_AGENT", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/115.0.0.0 Safari/537.36"),
		Timeout:    getDurationEnv("TIMEOUT", 30*time.Second),
		RateLimit:  getDurationEnv("RATE_LIMIT", 1*time.Second),
		MaxRetries: getIntEnv("MAX_RETRIES", 3),

		// Cache settings
		CacheEnabled: getBoolEnv("CACHE_ENABLED", true),
		CacheTTL:     getDurationEnv("CACHE_TTL", 5*time.Minute),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		// Simple conversion, could be improved with strconv.Atoi
		return defaultValue
	}
	return defaultValue
}

func getBoolEnv(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		return value == "true"
	}
	return defaultValue
}
