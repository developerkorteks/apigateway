package handlers

import (
	"apicategorywithfallback/internal/domain"
	"apicategorywithfallback/internal/service"
	"apicategorywithfallback/pkg/config"
	"apicategorywithfallback/pkg/database"
	"apicategorywithfallback/pkg/logger"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func setupTestHandler() (*APIHandler, func()) {
	// Initialize logger
	logger.Init()

	// Create temporary database
	dbPath := "/tmp/test_handler.db"

	db, err := database.Init(dbPath)
	if err != nil {
		panic(err)
	}

	cfg := &config.Config{
		APITimeout:     10 * time.Second,
		MaxConcurrency: 5,
		RateLimit:      100, // Allow 100 requests per second
		CacheTTL: map[string]time.Duration{
			"/api/v1/home": 15 * time.Minute,
		},
	}

	apiService := service.NewAPIService(db, cfg)
	handler := NewAPIHandler(apiService)

	cleanup := func() {
		db.Close()
		os.Remove(dbPath)
	}

	return handler, cleanup
}

func TestNewAPIHandler(t *testing.T) {
	handler, cleanup := setupTestHandler()
	defer cleanup()

	if handler == nil {
		t.Errorf("NewAPIHandler returned nil")
	}

	if handler.apiService == nil {
		t.Errorf("APIService not set in handler")
	}
}

func TestBuildRequestContext(t *testing.T) {
	handler, cleanup := setupTestHandler()
	defer cleanup()

	// Setup Gin context
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Create test request
	req, _ := http.NewRequest("GET", "/api/v1/home?category=anime&page=1", nil)
	req.Header.Set("User-Agent", "test-agent")
	req.RemoteAddr = "127.0.0.1:12345"
	c.Request = req

	ctx := handler.buildRequestContext(c, "/api/v1/home")

	if ctx.Endpoint != "/api/v1/home" {
		t.Errorf("Expected endpoint '/api/v1/home', got '%s'", ctx.Endpoint)
	}

	if ctx.Category != "anime" {
		t.Errorf("Expected category 'anime', got '%s'", ctx.Category)
	}

	if ctx.Parameters["page"] != "1" {
		t.Errorf("Expected page parameter '1', got '%s'", ctx.Parameters["page"])
	}

	if ctx.UserAgent != "test-agent" {
		t.Errorf("Expected user agent 'test-agent', got '%s'", ctx.UserAgent)
	}
}

func TestBuildRequestContextDefaultCategory(t *testing.T) {
	handler, cleanup := setupTestHandler()
	defer cleanup()

	// Setup Gin context without category parameter
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req, _ := http.NewRequest("GET", "/api/v1/home", nil)
	c.Request = req

	ctx := handler.buildRequestContext(c, "/api/v1/home")

	if ctx.Category != "anime" {
		t.Errorf("Expected default category 'anime', got '%s'", ctx.Category)
	}
}

func TestHandleHomeEndpoint(t *testing.T) {
	handler, cleanup := setupTestHandler()
	defer cleanup()

	// Setup Gin
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/api/v1/home", handler.HandleHome)

	// Create test request (use unique parameter to avoid cache collision)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/home?category=anime&test=handler", nil)
	router.ServeHTTP(w, req)

	// Since we don't have actual API sources configured, expect service unavailable
	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected status %d, got %d", http.StatusServiceUnavailable, w.Code)
	}

	// Check response format
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("Failed to parse response JSON: %v", err)
	}

	if response["error"] != true {
		t.Errorf("Expected error field to be true")
	}

	if response["source"] != "apicategorywithfallback" {
		t.Errorf("Expected source 'apicategorywithfallback', got '%v'", response["source"])
	}
}

func TestHandleSearchEndpoint(t *testing.T) {
	handler, cleanup := setupTestHandler()
	defer cleanup()

	// Setup Gin
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/api/v1/search", handler.HandleSearch)

	// Create test request with search query
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/search?q=naruto&category=anime", nil)
	router.ServeHTTP(w, req)

	// Since we don't have actual API sources configured, expect service unavailable
	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected status %d, got %d", http.StatusServiceUnavailable, w.Code)
	}
}

func TestHandleJadwalRilisDayEndpoint(t *testing.T) {
	handler, cleanup := setupTestHandler()
	defer cleanup()

	// Setup Gin
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/api/v1/jadwal-rilis/:day", handler.HandleJadwalRilisDay)

	// Create test request with day parameter
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/jadwal-rilis/monday?category=anime", nil)
	router.ServeHTTP(w, req)

	// Since we don't have actual API sources configured, expect service unavailable
	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected status %d, got %d", http.StatusServiceUnavailable, w.Code)
	}
}

func TestProcessRequestWithMockService(t *testing.T) {
	// Use real service for now - we'll skip this test
	t.Skip("Skipping mock service test - needs interface refactoring")
}

func TestProcessRequestWithCachedResponse(t *testing.T) {
	// Use real service for now - we'll skip this test
	t.Skip("Skipping mock service test - needs interface refactoring")
}

// Mock API Service for testing
type MockAPIService struct {
	shouldSucceed bool
	mockResponse  *domain.APIResponse
}

func (m *MockAPIService) ProcessRequest(ctx *domain.RequestContext) (*domain.APIResponse, error) {
	if m.shouldSucceed {
		return m.mockResponse, nil
	}
	return nil, domain.ErrAllAPIsFailed
}

func (m *MockAPIService) GetHealthStatus() ([]database.HealthCheck, error) {
	return []database.HealthCheck{}, nil
}

func (m *MockAPIService) GetRequestLogs(limit int) ([]database.RequestLog, error) {
	return []database.RequestLog{}, nil
}

func (m *MockAPIService) GetCategories() ([]database.Category, error) {
	return []database.Category{
		{ID: 1, Name: "anime", IsActive: true},
	}, nil
}

func (m *MockAPIService) StartHealthChecker() {
	// Mock implementation
}
