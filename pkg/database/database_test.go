package database

import (
	"os"
	"testing"
	"time"
)

func TestInit(t *testing.T) {
	// Use temporary database file
	dbPath := "/tmp/test_api_fallback.db"
	defer os.Remove(dbPath)

	db, err := Init(dbPath)
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Test if tables exist
	tables := []string{"categories", "endpoints", "api_sources", "fallback_apis", "health_checks", "request_logs"}

	for _, table := range tables {
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&count)
		if err != nil {
			t.Errorf("Failed to check table %s: %v", table, err)
		}
		if count != 1 {
			t.Errorf("Table %s does not exist", table)
		}
	}

	// Test if default data exists
	var categoryCount int
	err = db.QueryRow("SELECT COUNT(*) FROM categories").Scan(&categoryCount)
	if err != nil {
		t.Errorf("Failed to count categories: %v", err)
	}
	if categoryCount == 0 {
		t.Errorf("No default categories found")
	}

	var endpointCount int
	err = db.QueryRow("SELECT COUNT(*) FROM endpoints").Scan(&endpointCount)
	if err != nil {
		t.Errorf("Failed to count endpoints: %v", err)
	}
	if endpointCount == 0 {
		t.Errorf("No default endpoints found")
	}
}

func TestGetEndpointsByCategory(t *testing.T) {
	dbPath := "/tmp/test_endpoints.db"
	defer os.Remove(dbPath)

	db, err := Init(dbPath)
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	endpoints, err := db.GetEndpointsByCategory("anime")
	if err != nil {
		t.Errorf("Failed to get endpoints: %v", err)
	}

	if len(endpoints) == 0 {
		t.Errorf("No endpoints found for anime category")
	}

	// Check if expected endpoints exist
	expectedEndpoints := []string{
		"/api/v1/home",
		"/api/v1/jadwal-rilis",
		"/api/v1/anime-terbaru",
		"/api/v1/movie",
		"/api/v1/anime-detail",
		"/api/v1/episode-detail",
		"/api/v1/search",
	}

	endpointPaths := make(map[string]bool)
	for _, endpoint := range endpoints {
		endpointPaths[endpoint.Path] = true
	}

	for _, expected := range expectedEndpoints {
		if !endpointPaths[expected] {
			t.Errorf("Expected endpoint %s not found", expected)
		}
	}
}

func TestGetAPISourcesByEndpoint(t *testing.T) {
	dbPath := "/tmp/test_api_sources.db"
	defer os.Remove(dbPath)

	db, err := Init(dbPath)
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	sources, err := db.GetAPISourcesByEndpoint("/api/v1/home", "anime")
	if err != nil {
		t.Errorf("Failed to get API sources: %v", err)
	}

	if len(sources) == 0 {
		t.Errorf("No API sources found for /api/v1/home endpoint")
	}

	// Check if sources are ordered by priority
	for i := 1; i < len(sources); i++ {
		if sources[i-1].IsPrimary && !sources[i].IsPrimary {
			continue // Primary sources should come first
		}
		if sources[i-1].IsPrimary == sources[i].IsPrimary && sources[i-1].Priority > sources[i].Priority {
			t.Errorf("API sources are not properly ordered by priority")
		}
	}
}

func TestLogRequest(t *testing.T) {
	dbPath := "/tmp/test_request_log.db"
	defer os.Remove(dbPath)

	db, err := Init(dbPath)
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	log := RequestLog{
		Endpoint:     "/api/v1/home",
		Category:     "anime",
		SourceUsed:   "test_source",
		FallbackUsed: false,
		ResponseTime: 1500,
		StatusCode:   200,
		ClientIP:     "127.0.0.1",
		UserAgent:    "test-agent",
	}

	err = db.LogRequest(log)
	if err != nil {
		t.Errorf("Failed to log request: %v", err)
	}

	// Verify log was inserted
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM request_logs WHERE endpoint = ?", log.Endpoint).Scan(&count)
	if err != nil {
		t.Errorf("Failed to count request logs: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 request log, got %d", count)
	}
}

func TestLogHealthCheck(t *testing.T) {
	dbPath := "/tmp/test_health_check.db"
	defer os.Remove(dbPath)

	db, err := Init(dbPath)
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	err = db.LogHealthCheck(1, "OK", 500, "")
	if err != nil {
		t.Errorf("Failed to log health check: %v", err)
	}

	// Verify health check was logged
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM health_checks WHERE api_source_id = 1").Scan(&count)
	if err != nil {
		t.Errorf("Failed to count health checks: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 health check log, got %d", count)
	}
}

func TestGetHealthChecks(t *testing.T) {
	dbPath := "/tmp/test_get_health_checks.db"
	defer os.Remove(dbPath)

	db, err := Init(dbPath)
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Log some health checks
	db.LogHealthCheck(1, "OK", 500, "")
	db.LogHealthCheck(2, "ERROR", 0, "Connection failed")

	healthChecks, err := db.GetHealthChecks(10)
	if err != nil {
		t.Errorf("Failed to get health checks: %v", err)
	}

	if len(healthChecks) != 2 {
		t.Errorf("Expected 2 health checks, got %d", len(healthChecks))
	}

	// Check if ordered by checked_at DESC
	if len(healthChecks) > 1 {
		first := healthChecks[0].CheckedAt
		second := healthChecks[1].CheckedAt
		if first < second {
			t.Errorf("Health checks are not ordered by checked_at DESC")
		}
	}
}

func TestGetRequestLogs(t *testing.T) {
	dbPath := "/tmp/test_get_request_logs.db"
	defer os.Remove(dbPath)

	db, err := Init(dbPath)
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Log some requests
	log1 := RequestLog{
		Endpoint:     "/api/v1/home",
		Category:     "anime",
		SourceUsed:   "source1",
		FallbackUsed: false,
		ResponseTime: 1000,
		StatusCode:   200,
		ClientIP:     "127.0.0.1",
		UserAgent:    "test-agent",
	}

	log2 := RequestLog{
		Endpoint:     "/api/v1/search",
		Category:     "anime",
		SourceUsed:   "source2",
		FallbackUsed: true,
		ResponseTime: 2000,
		StatusCode:   200,
		ClientIP:     "127.0.0.1",
		UserAgent:    "test-agent",
	}

	db.LogRequest(log1)
	time.Sleep(1 * time.Second) // Ensure different timestamps
	db.LogRequest(log2)

	logs, err := db.GetRequestLogs(10)
	if err != nil {
		t.Errorf("Failed to get request logs: %v", err)
	}

	if len(logs) != 2 {
		t.Errorf("Expected 2 request logs, got %d", len(logs))
	}

	// Check if ordered by created_at DESC (most recent first)
	if len(logs) > 1 {
		if logs[0].Endpoint != "/api/v1/search" {
			t.Errorf("Request logs are not ordered by created_at DESC")
		}
	}
}
