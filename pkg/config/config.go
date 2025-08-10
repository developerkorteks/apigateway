package config

import (
	"encoding/json"
	"os"
	"strconv"
	"strings"
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

	// Dynamic API Sources Configuration
	// This allows unlimited API sources to be configured via environment variables
	// Format: API_SOURCES_JSON or individual API_SOURCE_<NAME>_URL variables
	APISources map[string]string
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

		// Load dynamic API sources
		APISources: loadAPISources(),
	}

	// Set configurable cache TTL for different endpoints
	cfg.CacheTTL = map[string]time.Duration{
		"/api/v1/home":           getEnvDuration("CACHE_TTL_HOME", 15*time.Minute),
		"/api/v1/jadwal-rilis":   getEnvDuration("CACHE_TTL_JADWAL_RILIS", 30*time.Minute),
		"/api/v1/anime-terbaru":  getEnvDuration("CACHE_TTL_ANIME_TERBARU", 15*time.Minute),
		"/api/v1/movie":          getEnvDuration("CACHE_TTL_MOVIE", 1*time.Hour),
		"/api/v1/anime-detail":   getEnvDuration("CACHE_TTL_ANIME_DETAIL", 1*time.Hour),
		"/api/v1/episode-detail": getEnvDuration("CACHE_TTL_EPISODE_DETAIL", 30*time.Minute),
		"/api/v1/search":         getEnvDuration("CACHE_TTL_SEARCH", 10*time.Minute),
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

// loadAPISources loads API sources dynamically from environment variables
// Supports multiple formats:
// 1. API_SOURCES_JSON: JSON string with all sources
// 2. Individual API_SOURCE_<NAME>_URL variables
// 3. Legacy individual variables for backward compatibility
func loadAPISources() map[string]string {
	sources := make(map[string]string)

	// Method 1: Load from JSON configuration
	if jsonSources := os.Getenv("API_SOURCES_JSON"); jsonSources != "" {
		var jsonMap map[string]string
		if err := json.Unmarshal([]byte(jsonSources), &jsonMap); err == nil {
			for name, url := range jsonMap {
				sources[strings.ToLower(name)] = url
			}
		}
	}

	// Method 2: Load from individual API_SOURCE_<NAME>_URL variables
	for _, env := range os.Environ() {
		if strings.HasPrefix(env, "API_SOURCE_") && strings.HasSuffix(env, "_URL=") {
			parts := strings.SplitN(env, "=", 2)
			if len(parts) == 2 {
				// Extract source name from API_SOURCE_<NAME>_URL
				key := parts[0]
				sourceName := strings.TrimSuffix(strings.TrimPrefix(key, "API_SOURCE_"), "_URL")
				sourceName = strings.ToLower(sourceName)
				sources[sourceName] = parts[1]
			}
		}
	}

	// Method 3: Legacy support for backward compatibility
	legacySources := map[string]string{
		"gomunime":       getEnv("GOMUNIME_URL", ""),
		"multiplescrape": getEnv("MULTIPLESCRAPE_URL", ""),
		"winbutv":        getEnv("WINBUTV_URL", ""),
		"samehadaku":     getEnv("SAMEHADAKU_URL", ""),
		"otakudesu":      getEnv("OTAKUDESU_URL", ""),
		"kusonime":       getEnv("KUSONIME_URL", ""),
	}

	// Add legacy sources if not already defined and not empty
	for name, url := range legacySources {
		if url != "" && sources[name] == "" {
			sources[name] = url
		}
	}

	// Set default sources if none configured
	if len(sources) == 0 {
		sources = map[string]string{
			"gomunime":       "http://localhost:8001",
			"winbutv":        "http://localhost:8002",
			"samehadaku":     "http://128.199.109.211:8182",
			"multiplescrape": "http://multiplescrape:8081",
			"otakudesu":      "https://otakudesu.quest",
			"kusonime":       "https://kusonime.com",
		}
	}

	return sources
}
