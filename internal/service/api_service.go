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

	// Initialize HTTP client with timeout
	httpClient := &http.Client{
		Timeout: cfg.APITimeout,
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
	default:
		// For other endpoints, just return the first successful response
		return baseResponse
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

// aggregateListData combines list data from multiple sources
func (s *APIService) aggregateListData(result map[string]interface{}, responses []*domain.APIResponse, dataKey string) {
	var allData []interface{}

	for _, resp := range responses {
		var data map[string]interface{}
		if err := json.Unmarshal(resp.Data, &data); err != nil {
			logger.Warnf("Failed to unmarshal response from %s: %v", resp.SourceName, err)
			continue
		}

		if listData, exists := data[dataKey]; exists {
			if list, ok := listData.([]interface{}); ok {
				allData = append(allData, list...)
			}
		}
	}

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

// makeAPIRequest makes an HTTP request to an API
func (s *APIService) makeAPIRequest(url, sourceName string, isFallback bool) *domain.APIResponse {
	startTime := time.Now()

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return &domain.APIResponse{
			Error:        err,
			SourceName:   sourceName,
			IsFallback:   isFallback,
			ResponseTime: time.Since(startTime),
		}
	}

	// Set headers
	req.Header.Set("User-Agent", "APIFallback/1.0")
	req.Header.Set("Accept", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return &domain.APIResponse{
			Error:        err,
			SourceName:   sourceName,
			IsFallback:   isFallback,
			ResponseTime: time.Since(startTime),
		}
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return &domain.APIResponse{
			Error:        err,
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

	// Build query string only with external parameters
	queryParams := make(map[string]string)
	for key, value := range params {
		if !internalParams[key] {
			queryParams[key] = value
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
	// Skip health check for endpoints that require mandatory parameters
	// These endpoints will return 400/422 without proper parameters
	skipEndpoints := map[string]bool{
		"/api/v1/search":          true, // requires 'q' parameter
		"/api/v1/anime-detail/":   true, // requires 'id' or 'slug' parameter
		"/api/v1/episode-detail/": true, // requires episode parameters
	}

	if skipEndpoints[endpoint] {
		logger.Debugf("Skipping health check for %s %s (requires mandatory parameters)", source.SourceName, endpoint)
		return
	}

	url := source.BaseURL + endpoint
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
	sources, err := s.db.GetAllAPISourcesForHealthCheck()
	if err != nil {
		return nil, err
	}

	results := make(map[string]interface{})
	var totalChecked, totalHealthy, totalUnhealthy int

	for _, source := range sources {
		// Construct full URL for health check
		fullURL := source.BaseURL + source.EndpointPath

		// Perform HTTP request with timeout
		client := &http.Client{
			Timeout: 10 * time.Second,
		}

		start := time.Now()
		resp, err := client.Get(fullURL)
		responseTime := int(time.Since(start).Milliseconds())

		var status string
		var errorMessage string

		if err != nil {
			status = "unhealthy"
			errorMessage = err.Error()
			totalUnhealthy++
		} else {
			defer resp.Body.Close()
			if resp.StatusCode >= 200 && resp.StatusCode < 300 {
				status = "healthy"
				totalHealthy++
			} else {
				status = "unhealthy"
				errorMessage = fmt.Sprintf("HTTP %d", resp.StatusCode)
				totalUnhealthy++
			}
		}

		totalChecked++

		// Log the health check result
		s.db.LogHealthCheck(source.ID, status, responseTime, errorMessage)
	}

	results["total_checked"] = totalChecked
	results["healthy"] = totalHealthy
	results["unhealthy"] = totalUnhealthy
	results["timestamp"] = time.Now().Format("2006-01-02 15:04:05")

	return results, nil
}

// GetCategories returns all categories
func (s *APIService) GetCategories() ([]database.Category, error) {
	return s.db.GetCategories()
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
		}
	}

	if len(primarySources) == 0 {
		logger.Warnf("No primary sources available for %s in category %s", ctx.Endpoint, ctx.Category)
		return &domain.FallbackResult{Success: false}
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
	logger.Infof("Trying primary source: %s", source.SourceName)

	// Try primary source first
	url := s.buildURL(source.BaseURL, ctx.Endpoint, ctx.Parameters)
	resp := s.makeAPIRequest(url, source.SourceName, false)

	// Validate response
	if resp.Error == nil && resp.Data != nil {
		if err := validator.ValidateResponse(ctx.Endpoint, resp.Data); err != nil {
			logger.Warnf("Validation failed for %s: %v", source.SourceName, err)
			resp.Error = err
		} else {
			// Primary source successful
			logger.Infof("Primary source %s successful", source.SourceName)
			resultChan <- resp
			return
		}
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
		return responses[0] // fallback to first response
	}

	return &domain.APIResponse{
		Data:         aggregatedJSON,
		StatusCode:   200,
		ResponseTime: responses[0].ResponseTime,
		SourceName:   "aggregated_all_categories",
		IsFallback:   false,
	}
}
