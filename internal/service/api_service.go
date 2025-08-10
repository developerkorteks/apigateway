package service

import (
	"apicategorywithfallback/internal/domain"
	"apicategorywithfallback/pkg/cache"
	"apicategorywithfallback/pkg/config"
	"apicategorywithfallback/pkg/database"
	"apicategorywithfallback/pkg/logger"
	"apicategorywithfallback/pkg/validator"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type APIService struct {
	db          *database.DB
	cache       cache.Cache
	config      *config.Config
	httpClient  *http.Client
	rateLimiter *rate.Limiter
}

func NewAPIService(db *database.DB, cfg *config.Config) *APIService {
	// Initialize cache
	cacheInstance := cache.NewCache(cfg.RedisAddr, cfg.RedisDB)

	// Initialize HTTP client with timeout and redirect handling
	httpClient := &http.Client{
		Timeout: cfg.APITimeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Allow up to 10 redirects
			if len(via) >= 10 {
				return fmt.Errorf("stopped after 10 redirects")
			}
			// Preserve headers on redirect
			if len(via) > 0 {
				req.Header.Set("User-Agent", "APIFallback/1.0")
				req.Header.Set("Accept", "application/json")
			}
			return nil
		},
	}

	// Initialize rate limiter
	rateLimiter := rate.NewLimiter(rate.Limit(cfg.RateLimit), cfg.RateLimit)

	return &APIService{
		db:          db,
		cache:       cacheInstance,
		config:      cfg,
		httpClient:  httpClient,
		rateLimiter: rateLimiter,
	}
}

// ProcessRequest handles incoming API requests with fallback mechanism
func (s *APIService) ProcessRequest(ctx *domain.RequestContext) (*domain.APIResponse, error) {
	startTime := time.Now()

	// Check rate limit
	if !s.rateLimiter.Allow() {
		return nil, fmt.Errorf("rate limit exceeded")
	}

	// Generate cache key
	cacheKey := s.cache.GenerateKey(ctx.Category, ctx.Endpoint, ctx.Parameters)

	// Try to get from cache first
	if cachedData, err := s.cache.Get(cacheKey); err == nil && cachedData != nil {
		logger.Infof("Cache hit for key: %s", cacheKey)
		return &domain.APIResponse{
			Data:                cachedData,
			StatusCode:          200,
			ResponseTime:        time.Since(startTime),
			SourceName:          "cache",
			AllSourcesAttempted: []string{"cache"},
			TotalAttempts:       1,
			ActualSourceURL:     "cache",
		}, nil
	}

	// Handle "all" category to aggregate from all active categories
	if ctx.Category == "all" {
		return s.processAllCategories(ctx, startTime)
	}

	// Get API sources for this endpoint and category
	apiSources, err := s.db.GetAPISourcesByEndpoint(ctx.Endpoint, ctx.Category)
	if err != nil {
		return nil, fmt.Errorf("failed to get API sources: %v", err)
	}

	if len(apiSources) == 0 {
		return nil, fmt.Errorf("no API sources configured for endpoint %s in category %s", ctx.Endpoint, ctx.Category)
	}

	// Log all sources retrieved from database
	logger.Infof("Retrieved %d sources from database for %s in category %s:", len(apiSources), ctx.Endpoint, ctx.Category)
	for _, source := range apiSources {
		logger.Infof("  - %s (ID: %d, BaseURL: %s, IsPrimary: %t, IsActive: %t)",
			source.SourceName, source.ID, source.BaseURL, source.IsPrimary, source.IsActive)
	}

	// Collect all source names for metadata
	var allSourceNames []string
	for _, source := range apiSources {
		allSourceNames = append(allSourceNames, source.SourceName)
	}

	// Always try to get data from ALL primary sources (this is the main fix)
	logger.Infof("Attempting to fetch from all %d primary sources for %s", len(apiSources), ctx.Endpoint)
	result := s.tryAllPrimaryAPIsWithFallback(apiSources, ctx)

	// Log the request
	s.logRequest(ctx, result, time.Since(startTime))

	if !result.Success {
		return nil, fmt.Errorf("all API sources failed for endpoint %s", ctx.Endpoint)
	}

	// Enhance response with metadata
	if result.Response != nil {
		result.Response.AllSourcesAttempted = allSourceNames
		result.Response.TotalAttempts = len(apiSources)
		if result.Response.SourceName != "cache" {
			result.Response.ActualSourceURL = fmt.Sprintf("%s%s", result.Response.SourceName, ctx.Endpoint)
		}
	}

	// Cache successful response
	if result.Response != nil && result.Response.Data != nil {
		ttl := s.config.CacheTTL[ctx.Endpoint]
		if ttl == 0 {
			ttl = 15 * time.Minute // default TTL
		}

		if err := s.cache.Set(cacheKey, result.Response.Data, ttl); err != nil {
			logger.Errorf("Failed to cache response: %v", err)
		}
	}

	return result.Response, nil
}

// tryAPISources attempts to get data from primary API sources
func (s *APIService) tryAPISources(sources []database.APISource, ctx *domain.RequestContext) *domain.FallbackResult {
	// Create channels for concurrent requests
	resultChan := make(chan *domain.APIResponse, len(sources))
	var wg sync.WaitGroup

	// Start concurrent requests to all primary APIs
	for _, source := range sources {
		if !source.IsPrimary || !source.IsActive {
			continue
		}

		wg.Add(1)
		go func(src database.APISource) {
			defer wg.Done()

			url := s.buildURL(src.BaseURL, ctx.Endpoint, ctx.Parameters)
			resp := s.makeAPIRequest(url, src.SourceName, false)

			// Validate response
			if resp.Error == nil && resp.Data != nil {
				if err := validator.ValidateResponse(ctx.Endpoint, resp.Data); err != nil {
					logger.Warnf("Validation failed for %s: %v", src.SourceName, err)
					resp.Error = err
				}
			}

			resultChan <- resp
		}(source)
	}

	// Close channel when all goroutines complete
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Return the first successful response (based on priority)
	var bestResponse *domain.APIResponse
	var bestPriority int = 999999

	for resp := range resultChan {
		if resp.Error == nil {
			// Find the priority of this source
			for _, src := range sources {
				if src.SourceName == resp.SourceName && src.Priority < bestPriority {
					bestResponse = resp
					bestPriority = src.Priority
					break
				}
			}
		}
	}

	if bestResponse != nil {
		return &domain.FallbackResult{
			Success:      true,
			Response:     bestResponse,
			SourceUsed:   bestResponse.SourceName,
			FallbackUsed: false,
		}
	}

	return &domain.FallbackResult{Success: false}
}

// tryFallbackAPIs attempts to get data from fallback APIs
func (s *APIService) tryFallbackAPIs(sources []database.APISource, ctx *domain.RequestContext) *domain.FallbackResult {
	for _, source := range sources {
		if !source.IsActive {
			continue
		}

		// Get fallback APIs for this source
		fallbacks, err := s.db.GetFallbackAPIs(source.ID)
		if err != nil {
			logger.Errorf("Failed to get fallback APIs for source %s: %v", source.SourceName, err)
			continue
		}

		// Try each fallback API
		for _, fallback := range fallbacks {
			url := s.buildURL(fallback.FallbackURL, ctx.Endpoint, ctx.Parameters)
			resp := s.makeAPIRequest(url, source.SourceName, true)

			// Validate response
			if resp.Error == nil && resp.Data != nil {
				if err := validator.ValidateResponse(ctx.Endpoint, resp.Data); err != nil {
					logger.Warnf("Validation failed for fallback %s: %v", fallback.FallbackURL, err)
					continue
				}

				// Success with fallback
				return &domain.FallbackResult{
					Success:      true,
					Response:     resp,
					SourceUsed:   source.SourceName,
					FallbackUsed: true,
				}
			}
		}
	}

	return &domain.FallbackResult{Success: false}
}

// tryAllAPISources attempts to get data from all primary API sources and aggregate results
func (s *APIService) tryAllAPISources(sources []database.APISource, ctx *domain.RequestContext) *domain.FallbackResult {
	// Create channels for concurrent requests
	resultChan := make(chan *domain.APIResponse, len(sources))
	var wg sync.WaitGroup

	// Start concurrent requests to all primary APIs
	for _, source := range sources {
		if !source.IsPrimary || !source.IsActive {
			continue
		}

		wg.Add(1)
		go func(src database.APISource) {
			defer wg.Done()

			url := s.buildURL(src.BaseURL, ctx.Endpoint, ctx.Parameters)
			resp := s.makeAPIRequest(url, src.SourceName, false)

			// Validate response
			if resp.Error == nil && resp.Data != nil {
				if err := validator.ValidateResponse(ctx.Endpoint, resp.Data); err != nil {
					logger.Warnf("Validation failed for %s: %v", src.SourceName, err)
					resp.Error = err
				}
			}

			resultChan <- resp
		}(source)
	}

	// Close channel when all goroutines complete
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect all successful responses
	var successfulResponses []*domain.APIResponse
	for resp := range resultChan {
		if resp.Error == nil {
			successfulResponses = append(successfulResponses, resp)
		}
	}

	if len(successfulResponses) == 0 {
		return &domain.FallbackResult{Success: false}
	}

	// If only one successful response, return it
	if len(successfulResponses) == 1 {
		return &domain.FallbackResult{
			Success:      true,
			Response:     successfulResponses[0],
			SourceUsed:   successfulResponses[0].SourceName,
			FallbackUsed: false,
		}
	}

	// Aggregate multiple successful responses
	aggregatedResponse := s.aggregateResponses(successfulResponses, ctx.Endpoint)

	return &domain.FallbackResult{
		Success:      true,
		Response:     aggregatedResponse,
		SourceUsed:   fmt.Sprintf("aggregated_%d_sources", len(successfulResponses)),
		FallbackUsed: false,
	}
}

// aggregateResponses combines data from multiple successful API responses
func (s *APIService) aggregateResponses(responses []*domain.APIResponse, endpoint string) *domain.APIResponse {
	if len(responses) == 0 {
		return nil
	}

	// Use the first response as base
	baseResponse := responses[0]

	// Create aggregated response
	aggregatedData := make(map[string]interface{})

	// Add metadata
	aggregatedData["confidence_score"] = 1.0
	aggregatedData["message"] = "Data berhasil diambil dari multiple sources"

	// Collect all source names
	var sources []string
	for _, resp := range responses {
		sources = append(sources, resp.SourceName)
	}
	aggregatedData["sources"] = sources

	// Aggregate data based on endpoint type
	switch endpoint {
	case "/api/v1/home":
		s.aggregateHomeData(aggregatedData, responses)
	case "/api/v1/anime-terbaru":
		s.aggregateListData(aggregatedData, responses, "data")
	case "/api/v1/movie":
		s.aggregateListData(aggregatedData, responses, "data")
	case "/api/v1/jadwal-rilis":
		s.aggregateScheduleData(aggregatedData, responses)
	case "/api/v1/search":
		s.aggregateListData(aggregatedData, responses, "data")
	case "/api/v1/anime-detail", "/api/v1/anime-detail/":
		// For detail endpoints, return the first successful response (no aggregation needed)
		return baseResponse
	case "/api/v1/episode-detail", "/api/v1/episode-detail/":
		// For detail endpoints, return the first successful response (no aggregation needed)
		return baseResponse
	default:
		// For unknown endpoints, try to aggregate as list data
		s.aggregateListData(aggregatedData, responses, "data")
	}

	// Convert aggregated data back to JSON bytes
	aggregatedJSON, err := json.Marshal(aggregatedData)
	if err != nil {
		logger.Errorf("Failed to marshal aggregated data: %v", err)
		return baseResponse
	}

	return &domain.APIResponse{
		Data:         aggregatedJSON,
		StatusCode:   200,
		ResponseTime: baseResponse.ResponseTime,
		SourceName:   "aggregated",
		IsFallback:   false,
	}
}

// aggregateHomeData combines home page data from multiple sources
func (s *APIService) aggregateHomeData(result map[string]interface{}, responses []*domain.APIResponse) {
	var allTop10 []interface{}
	var allNewEps []interface{}
	var allMovies []interface{}
	var allSchedules []interface{}

	for _, resp := range responses {
		var data map[string]interface{}
		if err := json.Unmarshal(resp.Data, &data); err != nil {
			logger.Warnf("Failed to unmarshal response from %s: %v", resp.SourceName, err)
			continue
		}

		// Aggregate top10
		if top10, exists := data["top10"]; exists {
			if top10List, ok := top10.([]interface{}); ok {
				allTop10 = append(allTop10, top10List...)
			}
		}

		// Aggregate new episodes
		if newEps, exists := data["new_eps"]; exists {
			if newEpsList, ok := newEps.([]interface{}); ok {
				allNewEps = append(allNewEps, newEpsList...)
			}
		}

		// Aggregate movies
		if movies, exists := data["movies"]; exists {
			if moviesList, ok := movies.([]interface{}); ok {
				allMovies = append(allMovies, moviesList...)
			}
		}

		// Aggregate schedules
		if schedule, exists := data["jadwal_rilis"]; exists {
			allSchedules = append(allSchedules, schedule)
		}
	}

	result["top10"] = allTop10
	result["new_eps"] = allNewEps
	result["movies"] = allMovies
	result["jadwal_rilis"] = allSchedules
}

// aggregateListData combines list data from multiple sources with deduplication
func (s *APIService) aggregateListData(result map[string]interface{}, responses []*domain.APIResponse, dataKey string) {
	var allData []interface{}
	seenItems := make(map[string]bool) // For deduplication based on unique identifiers

	for _, resp := range responses {
		var data map[string]interface{}
		if err := json.Unmarshal(resp.Data, &data); err != nil {
			logger.Warnf("Failed to unmarshal response from %s: %v", resp.SourceName, err)
			continue
		}

		if listData, exists := data[dataKey]; exists {
			if list, ok := listData.([]interface{}); ok {
				for _, item := range list {
					if itemMap, ok := item.(map[string]interface{}); ok {
						// Create unique key based on available identifiers
						var uniqueKey string
						if slug, exists := itemMap["anime_slug"]; exists {
							uniqueKey = fmt.Sprintf("%v", slug)
						} else if judul, exists := itemMap["judul"]; exists {
							uniqueKey = fmt.Sprintf("%v", judul)
						} else if url, exists := itemMap["url"]; exists {
							uniqueKey = fmt.Sprintf("%v", url)
						} else {
							// Fallback: use JSON representation as key
							if jsonBytes, err := json.Marshal(item); err == nil {
								uniqueKey = string(jsonBytes)
							}
						}

						// Only add if not seen before
						if uniqueKey != "" && !seenItems[uniqueKey] {
							seenItems[uniqueKey] = true
							allData = append(allData, item)
							logger.Debugf("Added unique item from %s: %s", resp.SourceName, uniqueKey)
						} else if uniqueKey != "" {
							logger.Debugf("Skipped duplicate item from %s: %s", resp.SourceName, uniqueKey)
						}
					} else {
						// If item is not a map, add it directly (no deduplication possible)
						allData = append(allData, item)
					}
				}
			}
		}
	}

	logger.Infof("Aggregated %d unique items from %d sources", len(allData), len(responses))
	result[dataKey] = allData
}

// aggregateScheduleData combines schedule data from multiple sources
func (s *APIService) aggregateScheduleData(result map[string]interface{}, responses []*domain.APIResponse) {
	scheduleMap := make(map[string][]interface{})

	for _, resp := range responses {
		var data map[string]interface{}
		if err := json.Unmarshal(resp.Data, &data); err != nil {
			logger.Warnf("Failed to unmarshal response from %s: %v", resp.SourceName, err)
			continue
		}

		if scheduleData, exists := data["data"]; exists {
			if schedule, ok := scheduleData.(map[string]interface{}); ok {
				for day, dayData := range schedule {
					if dayList, ok := dayData.([]interface{}); ok {
						scheduleMap[day] = append(scheduleMap[day], dayList...)
					}
				}
			}
		}
	}

	result["data"] = scheduleMap
}

// makeAPIRequest makes an HTTP request to an API with robust error handling
func (s *APIService) makeAPIRequest(url, sourceName string, isFallback bool) *domain.APIResponse {
	startTime := time.Now()

	// Validate URL
	if url == "" {
		return &domain.APIResponse{
			Error:        fmt.Errorf("empty URL provided"),
			SourceName:   sourceName,
			IsFallback:   isFallback,
			ResponseTime: time.Since(startTime),
		}
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return &domain.APIResponse{
			Error:        fmt.Errorf("failed to create request: %w", err),
			SourceName:   sourceName,
			IsFallback:   isFallback,
			ResponseTime: time.Since(startTime),
		}
	}

	// Set headers
	req.Header.Set("User-Agent", "APIFallback/1.0")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Cache-Control", "no-cache")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return &domain.APIResponse{
			Error:        fmt.Errorf("request failed: %w", err),
			SourceName:   sourceName,
			IsFallback:   isFallback,
			ResponseTime: time.Since(startTime),
		}
	}
	defer func() {
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}
	}()

	// Check for HTTP errors
	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		return &domain.APIResponse{
			Error:        fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status),
			StatusCode:   resp.StatusCode,
			SourceName:   sourceName,
			IsFallback:   isFallback,
			ResponseTime: time.Since(startTime),
		}
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return &domain.APIResponse{
			Error:        fmt.Errorf("failed to read response body: %w", err),
			StatusCode:   resp.StatusCode,
			SourceName:   sourceName,
			IsFallback:   isFallback,
			ResponseTime: time.Since(startTime),
		}
	}

	// Validate response data
	if len(data) == 0 {
		return &domain.APIResponse{
			Error:        fmt.Errorf("empty response body"),
			StatusCode:   resp.StatusCode,
			SourceName:   sourceName,
			IsFallback:   isFallback,
			ResponseTime: time.Since(startTime),
		}
	}

	return &domain.APIResponse{
		Data:         data,
		StatusCode:   resp.StatusCode,
		SourceName:   sourceName,
		IsFallback:   isFallback,
		ResponseTime: time.Since(startTime),
	}
}

// buildURL constructs the full URL with parameters (excluding internal parameters)
func (s *APIService) buildURL(baseURL, endpoint string, params map[string]string) string {
	url := baseURL + endpoint

	// Filter out internal parameters that shouldn't be sent to external APIs
	internalParams := map[string]bool{
		"category":  true, // Internal parameter for API fallback routing
		"aggregate": true, // Internal parameter for aggregation mode
	}

	// Parameter name mapping for different endpoints
	paramMapping := map[string]map[string]string{
		"/api/v1/search": {
			"q": "query", // Map 'q' parameter to 'query' for search endpoints
		},
	}

	// Build query string only with external parameters
	queryParams := make(map[string]string)
	for key, value := range params {
		if !internalParams[key] {
			// Check if parameter name needs mapping
			mappedKey := key
			if endpointMapping, exists := paramMapping[endpoint]; exists {
				if newKey, needsMapping := endpointMapping[key]; needsMapping {
					mappedKey = newKey
				}
			}
			queryParams[mappedKey] = value
		}
	}

	// Special handling for different APIs
	if endpoint == "/api/v1/search" {
		// Ensure trailing slash for samehadaku search endpoint
		if strings.Contains(baseURL, "samehadaku") {
			if !strings.HasSuffix(url, "/") {
				url += "/"
			}
			// Add force_refresh parameter if not present
			if _, exists := queryParams["force_refresh"]; !exists {
				queryParams["force_refresh"] = "false"
			}
		}
	}

	if len(queryParams) > 0 {
		url += "?"
		first := true
		for key, value := range queryParams {
			if !first {
				url += "&"
			}
			url += fmt.Sprintf("%s=%s", key, value)
			first = false
		}
	}

	return url
}

// logRequest logs the API request
func (s *APIService) logRequest(ctx *domain.RequestContext, result *domain.FallbackResult, responseTime time.Duration) {
	sourceUsed := ""
	fallbackUsed := false
	statusCode := 500

	if result.Success && result.Response != nil {
		sourceUsed = result.SourceUsed
		fallbackUsed = result.FallbackUsed
		statusCode = result.Response.StatusCode
	}

	logEntry := database.RequestLog{
		Endpoint:     ctx.Endpoint,
		Category:     ctx.Category,
		SourceUsed:   sourceUsed,
		FallbackUsed: fallbackUsed,
		ResponseTime: int(responseTime.Milliseconds()),
		StatusCode:   statusCode,
		ClientIP:     ctx.ClientIP,
		UserAgent:    ctx.UserAgent,
	}

	if err := s.db.LogRequest(logEntry); err != nil {
		logger.Errorf("Failed to log request: %v", err)
	}
}

// StartHealthChecker starts the background health checker
func (s *APIService) StartHealthChecker() {
	ticker := time.NewTicker(s.config.HealthCheckInterval)
	defer ticker.Stop()

	logger.Info("Starting health checker")

	for {
		select {
		case <-ticker.C:
			s.performHealthChecks()
		}
	}
}

// performHealthChecks performs health checks on all API sources
func (s *APIService) performHealthChecks() {
	logger.Info("Performing health checks")

	categories, err := s.db.GetCategories()
	if err != nil {
		logger.Errorf("Failed to get categories for health check: %v", err)
		return
	}

	for _, category := range categories {
		if !category.IsActive {
			continue
		}

		endpoints, err := s.db.GetEndpointsByCategory(category.Name)
		if err != nil {
			logger.Errorf("Failed to get endpoints for category %s: %v", category.Name, err)
			continue
		}

		for _, endpoint := range endpoints {
			sources, err := s.db.GetAPISourcesByEndpoint(endpoint.Path, category.Name)
			if err != nil {
				logger.Errorf("Failed to get API sources for endpoint %s: %v", endpoint.Path, err)
				continue
			}

			for _, source := range sources {
				s.checkAPIHealth(source, endpoint.Path)
			}
		}
	}
}

// checkAPIHealth checks the health of a single API source
func (s *APIService) checkAPIHealth(source database.APISource, endpoint string) {
	// Add test parameters for endpoints that require them
	testParams := map[string]string{
		"/api/v1/search":          "?query=a", // Use 'a' for better test results
		"/api/v1/anime-detail/":   "?id=1",
		"/api/v1/episode-detail/": "?id=1",
		"/api/v1/anime-detail":    "?id=1",
		"/api/v1/episode-detail":  "?id=1",
	}

	url := source.BaseURL + endpoint
	if params, exists := testParams[endpoint]; exists {
		url += params
	}

	// Special handling for samehadaku search endpoint
	if strings.Contains(source.BaseURL, "samehadaku") && endpoint == "/api/v1/search" {
		if !strings.HasSuffix(url, "/") {
			url += "/"
		}
		// Add force_refresh parameter for samehadaku
		if strings.Contains(url, "?") {
			url += "&force_refresh=false"
		} else {
			url += "?force_refresh=false"
		}
	}
	startTime := time.Now()

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		s.db.UpdateHealthCheck(source.ID, "ERROR", 0, err.Error())
		return
	}

	req.Header.Set("User-Agent", "APIFallback-HealthCheck/1.0")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	responseTime := int(time.Since(startTime).Milliseconds())

	if err != nil {
		status := "ERROR"
		if err.Error() == "timeout" {
			status = "TIMEOUT"
		}
		s.db.UpdateHealthCheck(source.ID, status, responseTime, err.Error())
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		s.db.UpdateHealthCheck(source.ID, "OK", responseTime, "")
	} else {
		s.db.UpdateHealthCheck(source.ID, "ERROR", responseTime, fmt.Sprintf("HTTP %d", resp.StatusCode))
	}
}

// GetHealthStatus returns the current health status of all API sources
func (s *APIService) GetHealthStatus() ([]map[string]interface{}, error) {
	return s.db.GetHealthStatusWithDetails()
}

// RunAllHealthChecks runs health checks on all active API sources
func (s *APIService) RunAllHealthChecks() error {
	sources, err := s.db.GetAllAPISourcesForHealthCheck()
	if err != nil {
		return fmt.Errorf("failed to get API sources: %v", err)
	}

	// Run health checks concurrently
	var wg sync.WaitGroup
	for _, source := range sources {
		wg.Add(1)
		go func(src database.APISource) {
			defer wg.Done()
			s.runHealthCheckForSource(src)
		}(source)
	}

	wg.Wait()
	return nil
}

// runHealthCheckForSource performs health check on a specific API source
func (s *APIService) runHealthCheckForSource(source database.APISource) {
	start := time.Now()

	// Build health check URL - try multiple endpoints
	healthURLs := []string{
		strings.TrimSuffix(source.BaseURL, "/") + "/health",
		strings.TrimSuffix(source.BaseURL, "/") + "/",
		strings.TrimSuffix(source.BaseURL, "/"),
	}

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	var lastErr error
	var resp *http.Response

	// Try each URL until one works
	for _, url := range healthURLs {
		resp, lastErr = client.Get(url)
		if lastErr == nil && resp != nil {
			break
		}
		if resp != nil {
			resp.Body.Close()
		}
	}

	responseTime := int(time.Since(start).Milliseconds())

	if lastErr != nil {
		s.db.UpdateHealthCheck(source.ID, "ERROR", responseTime, lastErr.Error())
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		s.db.UpdateHealthCheck(source.ID, "OK", responseTime, "")
	} else {
		s.db.UpdateHealthCheck(source.ID, "ERROR", responseTime, fmt.Sprintf("HTTP %d", resp.StatusCode))
	}
}

// GetRequestLogs returns recent request logs
func (s *APIService) GetRequestLogs(limit int) ([]database.RequestLog, error) {
	return s.db.GetRequestLogs(limit)
}

// GetStatistics returns real statistics from database
func (s *APIService) GetStatistics() (map[string]interface{}, error) {
	return s.db.GetStatistics()
}

// CreateCategory creates a new category
func (s *APIService) CreateCategory(name string, isActive bool) error {
	return s.db.CreateCategory(name, isActive)
}

// UpdateCategory updates an existing category
func (s *APIService) UpdateCategory(id int, name string, isActive bool) error {
	return s.db.UpdateCategory(id, name, isActive)
}

// DeleteCategory deletes a category
func (s *APIService) DeleteCategory(id int) error {
	return s.db.DeleteCategory(id)
}

// GetAllAPISources returns all API sources with details
func (s *APIService) GetAllAPISources() ([]database.APISourceWithDetails, error) {
	return s.db.GetAllAPISources()
}

// CreateAPISource creates a new API source
func (s *APIService) CreateAPISource(endpointID int, sourceName, baseURL string, priority int, isPrimary bool) error {
	return s.db.CreateAPISource(endpointID, sourceName, baseURL, priority, isPrimary)
}

// CreateAPISourceForAllEndpoints creates a new API source for all endpoints in a category
func (s *APIService) CreateAPISourceForAllEndpoints(categoryName, sourceName, baseURL string, priority int, isPrimary bool) error {
	// Get all endpoints for the specified category
	endpoints, err := s.db.GetEndpointsByCategory(categoryName)
	if err != nil {
		return fmt.Errorf("failed to get endpoints for category %s: %v", categoryName, err)
	}

	if len(endpoints) == 0 {
		return fmt.Errorf("no endpoints found for category: %s", categoryName)
	}

	// Create API source for each endpoint
	var errors []string
	successCount := 0

	for _, endpoint := range endpoints {
		err := s.db.CreateAPISource(endpoint.ID, sourceName, baseURL, priority, isPrimary)
		if err != nil {
			errors = append(errors, fmt.Sprintf("endpoint %s (ID: %d): %v", endpoint.Path, endpoint.ID, err))
		} else {
			successCount++
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("partially failed - %d success, %d errors: %s",
			successCount, len(errors), strings.Join(errors, "; "))
	}

	return nil
}

// UpdateAPISource updates an existing API source
func (s *APIService) UpdateAPISource(id int, sourceName, baseURL string, priority int, isPrimary, isActive bool) error {
	return s.db.UpdateAPISource(id, sourceName, baseURL, priority, isPrimary, isActive)
}

// DeleteAPISource deletes an API source
func (s *APIService) DeleteAPISource(id int) error {
	return s.db.DeleteAPISource(id)
}

// RunManualHealthCheck performs manual health check on all active API sources
func (s *APIService) RunManualHealthCheck() (map[string]interface{}, error) {
	// Run all health checks
	err := s.RunAllHealthChecks()
	if err != nil {
		return nil, fmt.Errorf("failed to run health checks: %v", err)
	}

	// Get updated health status
	healthStatus, err := s.GetHealthStatus()
	if err != nil {
		return nil, fmt.Errorf("failed to get health status: %v", err)
	}

	// Calculate summary
	var totalChecked, totalHealthy, totalUnhealthy int
	totalChecked = len(healthStatus)

	for _, status := range healthStatus {
		if statusStr, ok := status["status"].(string); ok && statusStr == "healthy" {
			totalHealthy++
		} else {
			totalUnhealthy++
		}
	}

	results := map[string]interface{}{
		"total_checked":   totalChecked,
		"total_healthy":   totalHealthy,
		"total_unhealthy": totalUnhealthy,
		"health_percentage": func() int {
			if totalChecked > 0 {
				return int((float64(totalHealthy) / float64(totalChecked)) * 100)
			}
			return 0
		}(),
		"checked_at": time.Now().Format("2006-01-02 15:04:05"),
		"details":    healthStatus,
	}

	return results, nil
}

// DeleteAPISourceByName deletes all API sources with the given source name
func (s *APIService) DeleteAPISourceByName(sourceName string) error {
	return s.db.DeleteAPISourceByName(sourceName)
}

// GetAPISourcesByName returns all API sources with the given source name
func (s *APIService) GetAPISourcesByName(sourceName string) ([]database.APISourceWithDetails, error) {
	return s.db.GetAPISourcesByName(sourceName)
}

// GetCategories returns all categories
func (s *APIService) GetCategories() ([]database.Category, error) {
	return s.db.GetCategories()
}

// GetCategoryNames returns all category names for dynamic Swagger documentation
func (s *APIService) GetCategoryNames() ([]string, error) {
	return s.db.GetCategoryNames()
}

// GetAllEndpoints returns all endpoints with details
func (s *APIService) GetAllEndpoints() ([]database.EndpointWithDetails, error) {
	return s.db.GetAllEndpoints()
}

// CreateEndpoint creates a new endpoint
func (s *APIService) CreateEndpoint(categoryID int, path string) (*database.Endpoint, error) {
	return s.db.CreateEndpoint(categoryID, path)
}

// UpdateEndpoint updates an existing endpoint
func (s *APIService) UpdateEndpoint(id int, categoryID int, path string) (*database.Endpoint, error) {
	return s.db.UpdateEndpoint(id, categoryID, path)
}

// DeleteEndpoint deletes an endpoint
func (s *APIService) DeleteEndpoint(id int) error {
	return s.db.DeleteEndpoint(id)
}

// processAllCategories handles requests for category "all" by fetching from all active categories
func (s *APIService) processAllCategories(ctx *domain.RequestContext, startTime time.Time) (*domain.APIResponse, error) {
	logger.Infof("Processing request for all categories: %s", ctx.Endpoint)

	// Get all active categories
	categories, err := s.db.GetCategories()
	if err != nil {
		return nil, fmt.Errorf("failed to get categories: %v", err)
	}

	var allResponses []*domain.APIResponse
	var wg sync.WaitGroup
	responseChan := make(chan *domain.APIResponse, len(categories))

	// Process each active category concurrently
	for _, category := range categories {
		if !category.IsActive {
			continue
		}

		wg.Add(1)
		go func(cat database.Category) {
			defer wg.Done()

			// Create new context for this category
			categoryCtx := &domain.RequestContext{
				Endpoint:   ctx.Endpoint,
				Category:   cat.Name,
				Parameters: ctx.Parameters,
				ClientIP:   ctx.ClientIP,
				UserAgent:  ctx.UserAgent,
				StartTime:  startTime,
			}

			// Get API sources for this category
			apiSources, err := s.db.GetAPISourcesByEndpoint(ctx.Endpoint, cat.Name)
			if err != nil {
				logger.Warnf("Failed to get API sources for category %s: %v", cat.Name, err)
				return
			}

			if len(apiSources) == 0 {
				logger.Warnf("No API sources for category %s, endpoint %s", cat.Name, ctx.Endpoint)
				return
			}

			// Try all primary APIs for this category
			result := s.tryAllPrimaryAPIsWithFallback(apiSources, categoryCtx)
			if result.Success && result.Response != nil {
				// Add category metadata to response
				var responseData map[string]interface{}
				if err := json.Unmarshal(result.Response.Data, &responseData); err == nil {
					responseData["category"] = cat.Name
					if modifiedData, err := json.Marshal(responseData); err == nil {
						result.Response.Data = modifiedData
					}
				}
				responseChan <- result.Response
			}
		}(category)
	}

	// Wait for all goroutines to complete
	go func() {
		wg.Wait()
		close(responseChan)
	}()

	// Collect all successful responses
	for resp := range responseChan {
		allResponses = append(allResponses, resp)
	}

	if len(allResponses) == 0 {
		return nil, fmt.Errorf("no successful responses from any category")
	}

	// Aggregate responses from all categories
	aggregatedResponse := s.aggregateResponsesFromAllCategories(allResponses, ctx.Endpoint)
	return aggregatedResponse, nil
}

// tryAllPrimaryAPIsWithFallback tries all primary APIs and uses fallback for each that fails
func (s *APIService) tryAllPrimaryAPIsWithFallback(sources []database.APISource, ctx *domain.RequestContext) *domain.FallbackResult {
	logger.Infof("Trying all %d primary APIs for %s in category %s", len(sources), ctx.Endpoint, ctx.Category)

	// Filter only primary sources
	var primarySources []database.APISource
	for _, source := range sources {
		if source.IsPrimary && source.IsActive {
			primarySources = append(primarySources, source)
			logger.Infof("Found primary source: %s (ID: %d)", source.SourceName, source.ID)
		}
	}

	if len(primarySources) == 0 {
		logger.Warnf("No primary sources available for %s in category %s", ctx.Endpoint, ctx.Category)
		return &domain.FallbackResult{Success: false}
	}

	logger.Infof("Total primary sources found: %d", len(primarySources))

	// Special handling for detail endpoints - bruteforce all sources and return first valid
	// Support both with and without trailing slash
	if ctx.Endpoint == "/api/v1/anime-detail/" || ctx.Endpoint == "/api/v1/anime-detail" ||
		ctx.Endpoint == "/api/v1/episode-detail/" || ctx.Endpoint == "/api/v1/episode-detail" {
		return s.bruteforceDetailSources(primarySources, ctx)
	}

	// Create channels for concurrent requests
	resultChan := make(chan *domain.APIResponse, len(primarySources))
	var wg sync.WaitGroup

	// Try each primary source concurrently
	for _, source := range primarySources {
		wg.Add(1)
		go func(src database.APISource) {
			defer wg.Done()
			s.trySourceWithFallback(src, ctx, resultChan)
		}(source)
	}

	// Close channel when all goroutines complete
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect all successful responses
	var successfulResponses []*domain.APIResponse
	for resp := range resultChan {
		if resp.Error == nil {
			successfulResponses = append(successfulResponses, resp)
			logger.Infof("Successfully got data from source: %s", resp.SourceName)
		}
	}

	if len(successfulResponses) == 0 {
		logger.Warnf("All primary and fallback APIs failed for %s", ctx.Endpoint)
		return &domain.FallbackResult{Success: false}
	}

	// If only one successful response, return it
	if len(successfulResponses) == 1 {
		return &domain.FallbackResult{
			Success:      true,
			Response:     successfulResponses[0],
			SourceUsed:   successfulResponses[0].SourceName,
			FallbackUsed: successfulResponses[0].IsFallback,
		}
	}

	// Aggregate multiple successful responses
	logger.Infof("Aggregating %d successful responses from different sources", len(successfulResponses))
	aggregatedResponse := s.aggregateResponses(successfulResponses, ctx.Endpoint)

	return &domain.FallbackResult{
		Success:      true,
		Response:     aggregatedResponse,
		SourceUsed:   fmt.Sprintf("aggregated_%d_sources", len(successfulResponses)),
		FallbackUsed: false,
	}
}

// trySourceWithFallback tries a primary source and its fallbacks
func (s *APIService) trySourceWithFallback(source database.APISource, ctx *domain.RequestContext, resultChan chan<- *domain.APIResponse) {
	logger.Infof("Trying primary source: %s (ID: %d, BaseURL: %s)", source.SourceName, source.ID, source.BaseURL)

	// Try primary source first
	url := s.buildURL(source.BaseURL, ctx.Endpoint, ctx.Parameters)
	logger.Infof("Built URL for %s: %s", source.SourceName, url)
	resp := s.makeAPIRequest(url, source.SourceName, false)

	// Special debug for winbutv
	if source.SourceName == "winbutv" {
		logger.Infof("WINBUTV DEBUG: Error=%v, DataLen=%d, StatusCode=%d", resp.Error, len(resp.Data), resp.StatusCode)
		if resp.Data != nil && len(resp.Data) > 0 {
			maxLen := 200
			if len(resp.Data) < maxLen {
				maxLen = len(resp.Data)
			}
			// Safe slice operation
			if maxLen > 0 {
				logger.Infof("WINBUTV DEBUG: First %d chars of response: %s", maxLen, string(resp.Data[:maxLen]))
			}
		}
	}

	// Validate response
	if resp.Error == nil && resp.Data != nil {
		if err := validator.ValidateResponse(ctx.Endpoint, resp.Data); err != nil {
			logger.Warnf("Validation failed for %s: %v", source.SourceName, err)
			if source.SourceName == "winbutv" {
				logger.Errorf("WINBUTV VALIDATION FAILED: %v", err)
			}
			resp.Error = err
		} else {
			// Primary source successful
			logger.Infof("Primary source %s successful with %d bytes of data", source.SourceName, len(resp.Data))
			if source.SourceName == "winbutv" {
				logger.Infof("WINBUTV SUCCESS: Sending to result channel")
			}
			resultChan <- resp
			return
		}
	} else {
		logger.Warnf("Primary source %s failed: Error=%v, DataLen=%d", source.SourceName, resp.Error, len(resp.Data))
	}

	// Primary failed, try fallbacks
	logger.Warnf("Primary source %s failed, trying fallbacks", source.SourceName)
	fallbacks, err := s.db.GetFallbackAPIs(source.ID)
	if err != nil {
		logger.Errorf("Failed to get fallback APIs for source %s: %v", source.SourceName, err)
		return
	}

	// Try each fallback
	for _, fallback := range fallbacks {
		logger.Infof("Trying fallback: %s", fallback.FallbackURL)
		fallbackURL := s.buildURL(fallback.FallbackURL, ctx.Endpoint, ctx.Parameters)
		fallbackResp := s.makeAPIRequest(fallbackURL, source.SourceName+"_fallback", true)

		// Validate fallback response
		if fallbackResp.Error == nil && fallbackResp.Data != nil {
			if err := validator.ValidateResponse(ctx.Endpoint, fallbackResp.Data); err != nil {
				logger.Warnf("Validation failed for fallback %s: %v", fallback.FallbackURL, err)
				continue
			}

			// Fallback successful
			logger.Infof("Fallback successful for %s", source.SourceName)
			resultChan <- fallbackResp
			return
		}
	}

	logger.Warnf("All attempts failed for source %s", source.SourceName)
}

// bruteforceDetailSources implements parallel bruteforce approach for detail endpoints
// This method hits ALL available sources concurrently and returns the first valid response
func (s *APIService) bruteforceDetailSources(primarySources []database.APISource, ctx *domain.RequestContext) *domain.FallbackResult {
	logger.Infof("Starting bruteforce approach for %s - hitting all %d sources concurrently", ctx.Endpoint, len(primarySources))

	// Collect all available URLs (primary + fallbacks)
	var allSources []bruteforceSource
	for _, source := range primarySources {
		// Add primary source
		primaryURL := s.buildURL(source.BaseURL, ctx.Endpoint, ctx.Parameters)
		allSources = append(allSources, bruteforceSource{
			URL:        primaryURL,
			SourceName: source.SourceName,
			Priority:   source.Priority,
			IsFallback: false,
		})

		// Add fallback sources
		fallbacks, err := s.db.GetFallbackAPIs(source.ID)
		if err != nil {
			logger.Warnf("Failed to get fallback APIs for source %s: %v", source.SourceName, err)
			continue
		}

		for i, fallback := range fallbacks {
			fallbackURL := s.buildURL(fallback.FallbackURL, ctx.Endpoint, ctx.Parameters)
			allSources = append(allSources, bruteforceSource{
				URL:        fallbackURL,
				SourceName: fmt.Sprintf("%s_fallback_%d", source.SourceName, i+1),
				Priority:   source.Priority + 1000 + i, // Lower priority than primary
				IsFallback: true,
			})
		}
	}

	if len(allSources) == 0 {
		logger.Warnf("No sources available for bruteforce")
		return &domain.FallbackResult{Success: false}
	}

	logger.Infof("Bruteforcing %d total sources (primary + fallback)", len(allSources))

	// Channel to receive results
	resultChan := make(chan *domain.APIResponse, len(allSources))
	firstValidChan := make(chan *domain.APIResponse, 1)
	var wg sync.WaitGroup
	var once sync.Once

	// Start all requests concurrently
	for _, source := range allSources {
		wg.Add(1)
		go func(src bruteforceSource) {
			defer wg.Done()

			logger.Debugf("Trying source: %s at %s", src.SourceName, src.URL)
			resp := s.makeAPIRequest(src.URL, src.SourceName, src.IsFallback)

			// Check if response is valid
			if resp.Error == nil && resp.Data != nil {
				if err := validator.ValidateResponse(ctx.Endpoint, resp.Data); err != nil {
					logger.Warnf("Validation failed for %s: %v", src.SourceName, err)
					resp.Error = err
					resultChan <- resp
					return
				}

				logger.Infof("✓ Valid data found from source: %s", src.SourceName)
				resp.Priority = src.Priority // Store priority for sorting

				// Send to result channel for collection
				resultChan <- resp

				// Try to send to first valid channel (non-blocking)
				// This allows us to return immediately when we get the first valid result
				once.Do(func() {
					select {
					case firstValidChan <- resp:
						logger.Infof("First valid response selected from: %s", src.SourceName)
					default:
						// Channel already has a response
					}
				})
			} else {
				logger.Debugf("Failed to get valid data from %s: %v", src.SourceName, resp.Error)
				resultChan <- resp
			}
		}(source)
	}

	// Close channels when all goroutines complete
	go func() {
		wg.Wait()
		close(resultChan)
		close(firstValidChan)
	}()

	// Wait for first valid response or all to complete
	select {
	case validResp, ok := <-firstValidChan:
		// Check if channel is still open and we got a valid response
		if ok && validResp != nil {
			logger.Infof("Bruteforce SUCCESS: Got valid data from %s", validResp.SourceName)

			// Still wait for other goroutines to complete to avoid resource leaks
			go func() {
				wg.Wait()
				logger.Debugf("All bruteforce goroutines completed")
			}()

			return &domain.FallbackResult{
				Success:      true,
				Response:     validResp,
				SourceUsed:   validResp.SourceName,
				FallbackUsed: validResp.IsFallback,
			}
		} else {
			logger.Warnf("Received nil or closed channel in firstValidChan")
		}
	case <-time.After(time.Duration(len(allSources)) * time.Second * 2): // Dynamic timeout based on source count
		// Timeout - collect any results we got
		logger.Warnf("Bruteforce timeout reached, collecting partial results")

		var allResponses []*domain.APIResponse
		// Drain the result channel
		func() {
			for {
				select {
				case resp := <-resultChan:
					allResponses = append(allResponses, resp)
				default:
					return
				}
			}
		}()

		// Find best valid response if any
		var bestValid *domain.APIResponse
		bestPriority := 999999
		for _, resp := range allResponses {
			if resp.Error == nil && resp.Priority < bestPriority {
				bestValid = resp
				bestPriority = resp.Priority
			}
		}

		if bestValid != nil {
			logger.Infof("Found valid response after timeout from: %s", bestValid.SourceName)
			return &domain.FallbackResult{
				Success:      true,
				Response:     bestValid,
				SourceUsed:   bestValid.SourceName,
				FallbackUsed: bestValid.IsFallback,
			}
		}
	}

	logger.Errorf("Bruteforce FAILED: No valid data found from any of %d sources", len(allSources))
	return &domain.FallbackResult{Success: false}
}

// bruteforceSource represents a source for bruteforce attempt
type bruteforceSource struct {
	URL        string
	SourceName string
	Priority   int
	IsFallback bool
}

// aggregateResponsesFromAllCategories combines responses from different categories
func (s *APIService) aggregateResponsesFromAllCategories(responses []*domain.APIResponse, endpoint string) *domain.APIResponse {
	if len(responses) == 0 {
		return nil
	}

	// Create aggregated response
	aggregatedData := make(map[string]interface{})

	// Add metadata
	aggregatedData["confidence_score"] = 1.0
	aggregatedData["message"] = fmt.Sprintf("Data berhasil diambil dari %d categories", len(responses))

	// Collect all source names and categories
	var sources []string
	var categories []string
	categoryData := make(map[string]interface{})

	for _, resp := range responses {
		sources = append(sources, resp.SourceName)

		var data map[string]interface{}
		if err := json.Unmarshal(resp.Data, &data); err != nil {
			logger.Warnf("Failed to unmarshal response from %s: %v", resp.SourceName, err)
			continue
		}

		// Extract category from response
		if category, exists := data["category"]; exists {
			categoryName := fmt.Sprintf("%v", category)
			categories = append(categories, categoryName)

			// Remove category field from individual data before storing
			delete(data, "category")
			categoryData[categoryName] = data
		}
	}

	aggregatedData["sources"] = sources
	aggregatedData["categories"] = categories
	aggregatedData["data_by_category"] = categoryData

	// Convert aggregated data back to JSON bytes
	aggregatedJSON, err := json.Marshal(aggregatedData)
	if err != nil {
		logger.Errorf("Failed to marshal aggregated data: %v", err)
		// Safe fallback - check if responses slice is not empty
		if len(responses) > 0 {
			return responses[0]
		}
		// Return error response if no responses available
		return &domain.APIResponse{
			Error:        fmt.Errorf("failed to marshal aggregated data: %w", err),
			StatusCode:   500,
			ResponseTime: 0,
			SourceName:   "aggregated_all_categories",
			IsFallback:   false,
		}
	}

	// Safe access to response time
	var responseTime time.Duration
	if len(responses) > 0 {
		responseTime = responses[0].ResponseTime
	}

	return &domain.APIResponse{
		Data:         aggregatedJSON,
		StatusCode:   200,
		ResponseTime: responseTime,
		SourceName:   "aggregated_all_categories",
		IsFallback:   false,
	}
}
