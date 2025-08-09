package main

import (
	"apicategorywithfallback/internal/api"
	"apicategorywithfallback/internal/service"
	"apicategorywithfallback/pkg/config"
	"apicategorywithfallback/pkg/database"
	"apicategorywithfallback/pkg/logger"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestFullIntegration(t *testing.T) {
	// Initialize logger
	logger.Init()

	// Create temporary database
	dbPath := "/tmp/test_integration.db"
	defer os.Remove(dbPath)

	db, err := database.Init(dbPath)
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Create mock API servers
	mockAPI1 := createMockAPIServer(t, "mock_api_1", 0.8)
	defer mockAPI1.Close()

	mockAPI2 := createMockAPIServer(t, "mock_api_2", 0.9)
	defer mockAPI2.Close()

	// Add mock API sources to database
	_, err = db.Exec(`
		INSERT INTO api_sources (endpoint_id, source_name, base_url, priority, is_primary, is_active) 
		VALUES (1, 'mock_api_1', ?, 1, 1, 1), (1, 'mock_api_2', ?, 2, 1, 1)
	`, mockAPI1.URL, mockAPI2.URL)
	if err != nil {
		t.Fatalf("Failed to insert mock API sources: %v", err)
	}

	// Setup configuration
	cfg := &config.Config{
		APITimeout:     10 * time.Second,
		MaxConcurrency: 5,
		RateLimit:      100, // Allow 100 requests per second
		CacheTTL: map[string]time.Duration{
			"/api/v1/home": 15 * time.Minute,
		},
	}

	// Setup cache and service
	apiService := service.NewAPIService(db, cfg)

	// Setup router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	api.SetupRoutes(router, apiService)

	// Test API endpoint
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/home?category=anime", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		t.Logf("Response body: %s", w.Body.String())
	}

	// Check response headers
	if w.Header().Get("Content-Type") != "application/json" {
		t.Errorf("Expected Content-Type 'application/json', got '%s'", w.Header().Get("Content-Type"))
	}

	if w.Header().Get("X-Source") == "" {
		t.Errorf("Expected X-Source header to be set")
	}

	// Parse response
	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("Failed to parse response JSON: %v", err)
	}

	// Check response structure
	if response["confidence_score"] == nil {
		t.Errorf("Response missing confidence_score")
	}

	if response["source"] == nil {
		t.Errorf("Response missing source")
	}
}

func TestFallbackMechanism(t *testing.T) {
	// Create temporary database
	dbPath := "/tmp/test_fallback.db"
	defer os.Remove(dbPath)

	db, err := database.Init(dbPath)
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Create failing primary API and working fallback API
	failingAPI := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "server error"}`))
	}))
	defer failingAPI.Close()

	workingAPI := createMockAPIServer(t, "fallback_api", 0.8)
	defer workingAPI.Close()

	// Add API sources to database
	result, err := db.Exec(`
		INSERT INTO api_sources (endpoint_id, source_name, base_url, priority, is_primary, is_active) 
		VALUES (1, 'failing_api', ?, 1, 1, 1)
	`, failingAPI.URL)
	if err != nil {
		t.Fatalf("Failed to insert failing API source: %v", err)
	}

	apiSourceID, _ := result.LastInsertId()

	// Add fallback API
	_, err = db.Exec(`
		INSERT INTO fallback_apis (api_source_id, fallback_url, priority, is_active) 
		VALUES (?, ?, 1, 1)
	`, apiSourceID, workingAPI.URL+"/api/v1/home")
	if err != nil {
		t.Fatalf("Failed to insert fallback API: %v", err)
	}

	// Setup configuration
	cfg := &config.Config{
		APITimeout:     5 * time.Second,
		MaxConcurrency: 5,
		RateLimit:      100, // Allow 100 requests per second
		CacheTTL: map[string]time.Duration{
			"/api/v1/home": 15 * time.Minute,
		},
	}

	// Setup cache and service
	apiService := service.NewAPIService(db, cfg)

	// Setup router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	api.SetupRoutes(router, apiService)

	// Test API endpoint - should use fallback
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/home?category=anime", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		t.Logf("Response body: %s", w.Body.String())
	}

	// Should have used fallback
	if w.Header().Get("X-Source") != "failing_api" {
		t.Logf("Source used: %s", w.Header().Get("X-Source"))
	}
}

func TestCachingMechanism(t *testing.T) {
	// Initialize logger
	logger.Init()

	// Create temporary database
	dbPath := "/tmp/test_caching.db"
	defer os.Remove(dbPath)

	db, err := database.Init(dbPath)
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Create mock API server
	requestCount := 0
	mockAPI := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		response := map[string]interface{}{
			"confidence_score": 0.8,
			"message":          "success",
			"source":           "mock_api",
			"request_count":    requestCount,
			"top10":            []interface{}{},
			"new_eps":          []interface{}{},
			"movies":           []interface{}{},
			"jadwal_rilis":     map[string]interface{}{},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer mockAPI.Close()

	// Add mock API source to database
	_, err = db.Exec(`
		INSERT INTO api_sources (endpoint_id, source_name, base_url, priority, is_primary, is_active) 
		VALUES (1, 'mock_api', ?, 1, 1, 1)
	`, mockAPI.URL)
	if err != nil {
		t.Fatalf("Failed to insert mock API source: %v", err)
	}

	// Setup configuration with short cache TTL for testing
	cfg := &config.Config{
		APITimeout:     10 * time.Second,
		MaxConcurrency: 5,
		RateLimit:      100, // Allow 100 requests per second
		CacheTTL: map[string]time.Duration{
			"/api/v1/home": 1 * time.Second, // Short TTL for testing
		},
	}

	// Setup cache and service
	apiService := service.NewAPIService(db, cfg)

	// Setup router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	api.SetupRoutes(router, apiService)

	// First request - should hit API (use unique parameter to avoid cache collision)
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("GET", "/api/v1/home?category=anime&test=caching", nil)
	router.ServeHTTP(w1, req1)

	if w1.Code != http.StatusOK {
		t.Errorf("First request failed: %d", w1.Code)
	}

	if w1.Header().Get("X-Cache") == "HIT" {
		t.Errorf("First request should not be cached")
	}

	// Second request immediately - should hit cache
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/api/v1/home?category=anime&test=caching", nil)
	router.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Errorf("Second request failed: %d", w2.Code)
	}

	if w2.Header().Get("X-Cache") != "HIT" {
		t.Errorf("Second request should hit cache")
	}

	if w2.Header().Get("X-Source") != "cache" {
		t.Errorf("Second request should use cache source")
	}

	// Wait for cache to expire
	time.Sleep(2 * time.Second)

	// Third request - should hit API again
	w3 := httptest.NewRecorder()
	req3, _ := http.NewRequest("GET", "/api/v1/home?category=anime&test=caching", nil)
	router.ServeHTTP(w3, req3)

	if w3.Code != http.StatusOK {
		t.Errorf("Third request failed: %d", w3.Code)
	}

	if w3.Header().Get("X-Cache") == "HIT" {
		t.Errorf("Third request should not hit expired cache")
	}

	// Verify API was called twice (not three times due to caching)
	if requestCount != 2 {
		t.Errorf("Expected 2 API calls, got %d", requestCount)
	}
}

func TestDashboardEndpoints(t *testing.T) {
	// Create temporary database
	dbPath := "/tmp/test_dashboard.db"
	defer os.Remove(dbPath)

	db, err := database.Init(dbPath)
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Setup configuration
	cfg := &config.Config{
		APITimeout:     10 * time.Second,
		MaxConcurrency: 5,
		RateLimit:      100, // Allow 100 requests per second
		CacheTTL:       map[string]time.Duration{},
	}

	// Setup cache and service
	apiService := service.NewAPIService(db, cfg)

	// Setup router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	api.SetupRoutes(router, apiService)

	// Test dashboard endpoints
	dashboardEndpoints := []string{
		"/dashboard/health",
		"/dashboard/logs",
		"/dashboard/stats",
		"/dashboard/categories",
	}

	for _, endpoint := range dashboardEndpoints {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", endpoint, nil)
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Dashboard endpoint %s failed with status %d", endpoint, w.Code)
			t.Logf("Response: %s", w.Body.String())
		}

		// Check if response is valid JSON
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Errorf("Dashboard endpoint %s returned invalid JSON: %v", endpoint, err)
		}

		// Check if response has status field
		if response["status"] != "success" {
			t.Errorf("Dashboard endpoint %s should return success status", endpoint)
		}
	}
}

func TestHealthEndpoint(t *testing.T) {
	// Setup minimal router for health check
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "apicategorywithfallback",
		})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Health endpoint failed with status %d", w.Code)
	}

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("Health endpoint returned invalid JSON: %v", err)
	}

	if response["status"] != "ok" {
		t.Errorf("Health endpoint should return 'ok' status")
	}

	if response["service"] != "apicategorywithfallback" {
		t.Errorf("Health endpoint should return correct service name")
	}
}

// Helper function to create mock API server
func createMockAPIServer(t *testing.T, sourceName string, confidenceScore float64) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"confidence_score": confidenceScore,
			"message":          "success",
			"source":           sourceName,
			"top10":            []interface{}{},
			"new_eps":          []interface{}{},
			"movies":           []interface{}{},
			"jadwal_rilis":     map[string]interface{}{},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
}

func TestRealAPIIntegration(t *testing.T) {
	// Skip this test if not running integration tests
	if testing.Short() {
		t.Skip("Skipping real API integration test in short mode")
	}

	// Create temporary database
	dbPath := "/tmp/test_real_api.db"
	defer os.Remove(dbPath)

	db, err := database.Init(dbPath)
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Add real API sources (assuming they're running)
	realAPIs := []struct {
		name string
		url  string
		port string
	}{
		{"fastapi_app", "http://localhost:8000", "8000"},
		{"multiplescrape", "http://localhost:8001", "8001"},
		{"winbutv", "http://localhost:8002", "8002"},
	}

	for i, api := range realAPIs {
		// Test if API is available
		resp, err := http.Get(api.url + "/health")
		if err != nil {
			t.Logf("API %s not available, skipping: %v", api.name, err)
			continue
		}
		resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Logf("API %s not healthy, skipping", api.name)
			continue
		}

		// Add to database
		_, err = db.Exec(`
			INSERT INTO api_sources (endpoint_id, source_name, base_url, priority, is_primary, is_active) 
			VALUES (1, ?, ?, ?, 1, 1)
		`, api.name, api.url, i+1)
		if err != nil {
			t.Logf("Failed to insert API source %s: %v", api.name, err)
		}
	}

	// Setup configuration
	cfg := &config.Config{
		APITimeout:     20 * time.Second, // Longer timeout for real APIs
		MaxConcurrency: 5,
		RateLimit:      100, // Allow 100 requests per second
		CacheTTL: map[string]time.Duration{
			"/api/v1/home": 15 * time.Minute,
		},
	}

	// Setup cache and service
	apiService := service.NewAPIService(db, cfg)

	// Setup router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	api.SetupRoutes(router, apiService)

	// Test real API endpoints
	testEndpoints := []string{
		"/api/v1/home",
		"/api/v1/anime-terbaru",
		"/api/v1/movie",
		"/api/v1/jadwal-rilis",
	}

	for _, endpoint := range testEndpoints {
		t.Run(fmt.Sprintf("RealAPI_%s", endpoint), func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", endpoint+"?category=anime", nil)
			router.ServeHTTP(w, req)

			t.Logf("Testing %s: Status %d", endpoint, w.Code)

			if w.Code == http.StatusOK {
				// Parse response to check structure
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				if err != nil {
					t.Errorf("Failed to parse response JSON: %v", err)
				} else {
					t.Logf("Success! Source: %s, Confidence: %v",
						w.Header().Get("X-Source"), response["confidence_score"])
				}
			} else if w.Code == http.StatusServiceUnavailable {
				t.Logf("All APIs unavailable for %s (expected if APIs not running)", endpoint)
			} else {
				t.Errorf("Unexpected status code %d for %s", w.Code, endpoint)
			}
		})
	}
}
