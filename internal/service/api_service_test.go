package service

import (
	"apicategorywithfallback/internal/domain"
	"apicategorywithfallback/pkg/config"
	"apicategorywithfallback/pkg/database"
	"apicategorywithfallback/pkg/logger"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func TestNewAPIService(t *testing.T) {
	// Initialize logger
	logger.Init()

	// Create temporary database
	dbPath := "/tmp/test_api_service.db"
	defer os.Remove(dbPath)

	db, err := database.Init(dbPath)
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	cfg := &config.Config{
		APITimeout:     10 * time.Second,
		MaxConcurrency: 5,
		RateLimit:      100, // Allow 100 requests per second
		CacheTTL: map[string]time.Duration{
			"/api/v1/home": 15 * time.Minute,
		},
	}

	service := NewAPIService(db, cfg)
	if service == nil {
		t.Errorf("NewAPIService returned nil")
	}

	if service.db != db {
		t.Errorf("Database not set correctly")
	}

	if service.cache == nil {
		t.Errorf("Cache not initialized")
	}

	if service.config != cfg {
		t.Errorf("Config not set correctly")
	}
}

func TestProcessRequestWithMockAPI(t *testing.T) {
	// Initialize logger
	logger.Init()

	// Create mock API server
	mockResponse := map[string]interface{}{
		"confidence_score": 0.8,
		"message":          "success",
		"source":           "mock_api",
		"top10":            []interface{}{},
		"new_eps":          []interface{}{},
		"movies":           []interface{}{},
		"jadwal_rilis":     map[string]interface{}{},
	}

	mockData, _ := json.Marshal(mockResponse)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(mockData)
	}))
	defer server.Close()

	// Create temporary database
	dbPath := "/tmp/test_process_request.db"
	defer os.Remove(dbPath)

	db, err := database.Init(dbPath)
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Add category and endpoint first
	_, err = db.Exec(`INSERT OR IGNORE INTO categories (id, name, is_active) VALUES (1, 'anime', 1)`)
	if err != nil {
		t.Fatalf("Failed to insert category: %v", err)
	}

	_, err = db.Exec(`INSERT OR IGNORE INTO endpoints (id, category_id, path) VALUES (1, 1, '/api/v1/home')`)
	if err != nil {
		t.Fatalf("Failed to insert endpoint: %v", err)
	}

	// Add mock API source to database
	_, err = db.Exec(`
		INSERT INTO api_sources (endpoint_id, source_name, base_url, priority, is_primary, is_active) 
		VALUES (1, 'mock_api', ?, 1, 1, 1)
	`, server.URL)
	if err != nil {
		t.Fatalf("Failed to insert mock API source: %v", err)
	}

	cfg := &config.Config{
		APITimeout:     10 * time.Second,
		MaxConcurrency: 5,
		RateLimit:      100, // Allow 100 requests per second
		CacheTTL: map[string]time.Duration{
			"/api/v1/home": 15 * time.Minute,
		},
	}

	service := NewAPIService(db, cfg)

	// Clear cache to ensure we hit the API, not cache
	cacheKey := service.cache.GenerateKey("anime", "/api/v1/home", map[string]string{})
	service.cache.Delete(cacheKey)

	ctx := &domain.RequestContext{
		Endpoint:   "/api/v1/home",
		Category:   "anime",
		Parameters: map[string]string{},
		ClientIP:   "127.0.0.1",
		UserAgent:  "test-agent",
		StartTime:  time.Now(),
	}

	response, err := service.ProcessRequest(ctx)
	if err != nil {
		t.Errorf("ProcessRequest failed: %v", err)
	}

	if response == nil {
		t.Errorf("Response is nil")
		return
	}

	if response.SourceName != "mock_api" {
		t.Errorf("Expected source name 'mock_api', got '%s'", response.SourceName)
	}

	if len(response.Data) == 0 {
		t.Errorf("Response data is empty")
	}
}

func TestProcessRequestWithCache(t *testing.T) {
	// Initialize logger
	logger.Init()

	// Create temporary database
	dbPath := "/tmp/test_cache_request.db"
	defer os.Remove(dbPath)

	db, err := database.Init(dbPath)
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	cfg := &config.Config{
		APITimeout:     10 * time.Second,
		MaxConcurrency: 5,
		RateLimit:      100, // Allow 100 requests per second
		CacheTTL: map[string]time.Duration{
			"/api/v1/home": 15 * time.Minute,
		},
	}

	service := NewAPIService(db, cfg)

	// Pre-populate cache
	cacheKey := "anime:/api/v1/home:d41d8cd98f00b204e9800998ecf8427e" // Empty params hash
	cacheData := []byte(`{"cached": true, "source": "cache"}`)
	service.cache.Set(cacheKey, cacheData, 15*time.Minute)

	ctx := &domain.RequestContext{
		Endpoint:   "/api/v1/home",
		Category:   "anime",
		Parameters: map[string]string{},
		ClientIP:   "127.0.0.1",
		UserAgent:  "test-agent",
		StartTime:  time.Now(),
	}

	response, err := service.ProcessRequest(ctx)
	if err != nil {
		t.Errorf("ProcessRequest with cache failed: %v", err)
	}

	if response == nil {
		t.Errorf("Response is nil")
		return
	}

	if response.SourceName != "cache" {
		t.Errorf("Expected source name 'cache', got '%s'", response.SourceName)
	}
}

func TestBuildURL(t *testing.T) {
	// Initialize logger
	logger.Init()

	service := &APIService{}

	// Test without parameters
	url := service.buildURL("http://example.com", "/api/v1/home", map[string]string{})
	expected := "http://example.com/api/v1/home"
	if url != expected {
		t.Errorf("Expected URL '%s', got '%s'", expected, url)
	}

	// Test with parameters
	params := map[string]string{
		"page": "1",
		"sort": "latest",
	}
	url = service.buildURL("http://example.com", "/api/v1/search", params)

	// URL should contain base URL and endpoint
	if !contains(url, "http://example.com/api/v1/search") {
		t.Errorf("URL should contain base URL and endpoint")
	}

	// URL should contain parameters
	if !contains(url, "page=1") || !contains(url, "sort=latest") {
		t.Errorf("URL should contain parameters")
	}
}

func TestGetHealthStatus(t *testing.T) {
	// Initialize logger
	logger.Init()

	// Create temporary database
	dbPath := "/tmp/test_health_status.db"
	defer os.Remove(dbPath)

	db, err := database.Init(dbPath)
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Add some health check data
	db.LogHealthCheck(1, "OK", 500, "")
	db.LogHealthCheck(2, "ERROR", 0, "Connection failed")

	cfg := &config.Config{}
	service := NewAPIService(db, cfg)

	healthStatus, err := service.GetHealthStatus()
	if err != nil {
		t.Errorf("GetHealthStatus failed: %v", err)
	}

	if len(healthStatus) == 0 {
		t.Errorf("Expected health status data, got empty slice")
	}
}

func TestGetRequestLogs(t *testing.T) {
	// Initialize logger
	logger.Init()

	// Create temporary database
	dbPath := "/tmp/test_request_logs.db"
	defer os.Remove(dbPath)

	db, err := database.Init(dbPath)
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Add some request logs
	log := database.RequestLog{
		Endpoint:     "/api/v1/home",
		Category:     "anime",
		SourceUsed:   "test_source",
		FallbackUsed: false,
		ResponseTime: 1000,
		StatusCode:   200,
		ClientIP:     "127.0.0.1",
		UserAgent:    "test-agent",
	}
	db.LogRequest(log)

	cfg := &config.Config{}
	service := NewAPIService(db, cfg)

	logs, err := service.GetRequestLogs(10)
	if err != nil {
		t.Errorf("GetRequestLogs failed: %v", err)
	}

	if len(logs) == 0 {
		t.Errorf("Expected request logs, got empty slice")
	}

	if logs[0].Endpoint != "/api/v1/home" {
		t.Errorf("Expected endpoint '/api/v1/home', got '%s'", logs[0].Endpoint)
	}
}

func TestGetCategories(t *testing.T) {
	// Initialize logger
	logger.Init()

	// Create temporary database
	dbPath := "/tmp/test_categories.db"
	defer os.Remove(dbPath)

	db, err := database.Init(dbPath)
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	cfg := &config.Config{}
	service := NewAPIService(db, cfg)

	categories, err := service.GetCategories()
	if err != nil {
		t.Errorf("GetCategories failed: %v", err)
	}

	if len(categories) == 0 {
		t.Errorf("Expected categories, got empty slice")
	}

	// Should have default "anime" category
	found := false
	for _, category := range categories {
		if category.Name == "anime" {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Expected 'anime' category not found")
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsSubstring(s, substr)))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
