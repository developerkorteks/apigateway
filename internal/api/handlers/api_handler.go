package handlers

import (
	"apicategorywithfallback/internal/domain"
	"apicategorywithfallback/internal/service"
	"apicategorywithfallback/pkg/logger"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type APIHandler struct {
	apiService *service.APIService
}

func NewAPIHandler(apiService *service.APIService) *APIHandler {
	return &APIHandler{
		apiService: apiService,
	}
}

// HandleHome handles /api/v1/home endpoint
// @Summary Get home page content from multiple anime sources
// @Description Retrieve aggregated home page content from multiple anime APIs with fallback mechanism. The system automatically calls all available primary APIs and aggregates the results.
// @Tags Anime
// @Accept json
// @Produce json
// @Param category query string false "Content category for API routing (anime, korean-drama, all). Used internally for API source selection, not forwarded to external APIs" default(anime) Enums(anime, korean-drama, all)
// @Success 200 {object} map[string]interface{} "Aggregated home page content with sources info"
// @Failure 400 {object} map[string]interface{} "Bad request - invalid category"
// @Failure 500 {object} map[string]interface{} "Internal server error - all API sources failed"
// @Router /api/v1/home [get]
func (h *APIHandler) HandleHome(c *gin.Context) {
	ctx := h.buildRequestContext(c, "/api/v1/home")
	h.processRequest(c, ctx)
}

// HandleJadwalRilis handles /api/v1/jadwal-rilis endpoint
// @Summary Get anime release schedule from multiple sources
// @Description Retrieve aggregated anime release schedule for all days from multiple APIs with automatic fallback
// @Tags Anime
// @Accept json
// @Produce json
// @Param category query string false "Content category for API routing (anime, korean-drama, all). Internal parameter, not sent to external APIs" default(anime) Enums(anime, korean-drama, all)
// @Success 200 {object} map[string]interface{} "Aggregated anime release schedule with sources info"
// @Failure 400 {object} map[string]interface{} "Bad request - invalid category"
// @Failure 500 {object} map[string]interface{} "Internal server error - all API sources failed"
// @Router /api/v1/jadwal-rilis [get]
func (h *APIHandler) HandleJadwalRilis(c *gin.Context) {
	ctx := h.buildRequestContext(c, "/api/v1/jadwal-rilis")
	h.processRequest(c, ctx)
}

// HandleJadwalRilisDay handles /api/v1/jadwal-rilis/{day} endpoint
// @Summary Get anime release schedule by day from multiple sources
// @Description Retrieve aggregated anime release schedule for a specific day from multiple APIs
// @Tags Anime
// @Accept json
// @Produce json
// @Param day path string true "Day of the week (senin, selasa, rabu, kamis, jumat, sabtu, minggu)"
// @Param category query string false "Content category for API routing (anime, korean-drama, all). Internal parameter for source selection" default(anime) Enums(anime, korean-drama, all)
// @Success 200 {object} map[string]interface{} "Aggregated anime release schedule for the day with sources info"
// @Failure 400 {object} map[string]interface{} "Bad request - invalid day or category"
// @Failure 500 {object} map[string]interface{} "Internal server error - all API sources failed"
// @Router /api/v1/jadwal-rilis/{day} [get]
func (h *APIHandler) HandleJadwalRilisDay(c *gin.Context) {
	day := c.Param("day")
	ctx := h.buildRequestContext(c, "/api/v1/jadwal-rilis/"+day)
	h.processRequest(c, ctx)
}

// HandleAnimeTerbaru handles /api/v1/anime-terbaru endpoint
// @Summary Get latest anime from multiple sources
// @Description Retrieve aggregated latest anime releases from all available APIs with automatic fallback mechanism
// @Tags Anime
// @Accept json
// @Produce json
// @Param page query int false "Page number for pagination" default(1)
// @Param category query string false "Content category for API routing (anime, korean-drama, all). Used for internal source selection only" default(anime) Enums(anime, korean-drama, all)
// @Success 200 {object} map[string]interface{} "Aggregated latest anime list from multiple sources with metadata"
// @Failure 400 {object} map[string]interface{} "Bad request - invalid parameters"
// @Failure 500 {object} map[string]interface{} "Internal server error - all API sources failed"
// @Router /api/v1/anime-terbaru [get]
func (h *APIHandler) HandleAnimeTerbaru(c *gin.Context) {
	ctx := h.buildRequestContext(c, "/api/v1/anime-terbaru")
	h.processRequest(c, ctx)
}

// HandleMovie handles /api/v1/movie endpoint
// @Summary Get anime movies from multiple sources
// @Description Retrieve aggregated anime movie listings from all available APIs with fallback mechanism
// @Tags Movie
// @Accept json
// @Produce json
// @Param page query int false "Page number for pagination" default(1)
// @Param category query string false "Content category for API routing - dynamically loaded from database" default(anime)
// @Success 200 {object} map[string]interface{} "Aggregated anime movie list from multiple sources"
// @Failure 400 {object} map[string]interface{} "Bad request - invalid parameters"
// @Failure 500 {object} map[string]interface{} "Internal server error - all API sources failed"
// @Router /api/v1/movie [get]
func (h *APIHandler) HandleMovie(c *gin.Context) {
	ctx := h.buildRequestContext(c, "/api/v1/movie")
	h.processRequest(c, ctx)
}

// HandleAnimeDetail handles /api/v1/anime-detail endpoint
// @Summary Get anime details from multiple sources with fallback
// @Description Retrieve detailed information about a specific anime from all available APIs. Requires anime identifier (id, slug, or anime_slug). Uses fallback mechanism when primary APIs fail.
// @Tags Detail
// @Accept json
// @Produce json
// @Param id query string false "Anime ID (alternative to slug/anime_slug)"
// @Param slug query string false "Anime slug (alternative to id/anime_slug)"
// @Param anime_slug query string false "Anime slug (alternative to id/slug)"
// @Param category query string false "Content category for API routing (anime, korean-drama, all). Internal parameter for source selection" default(anime) Enums(anime, korean-drama, all)
// @Success 200 {object} map[string]interface{} "Aggregated anime details from available sources"
// @Failure 400 {object} map[string]interface{} "Bad request - missing required anime identifier"
// @Failure 404 {object} map[string]interface{} "Not found - anime not found in any source"
// @Failure 500 {object} map[string]interface{} "Internal server error - all API sources failed"
// @Router /api/v1/anime-detail [get]
func (h *APIHandler) HandleAnimeDetail(c *gin.Context) {
	// Validate required anime identifier parameters
	id := c.Query("id")
	slug := c.Query("slug")
	animeSlug := c.Query("anime_slug")

	if id == "" && slug == "" && animeSlug == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "missing required parameter: one of 'id', 'slug', or 'anime_slug' is required",
			"source":  "apicategorywithfallback",
		})
		return
	}

	// Use the actual request path to maintain consistency
	requestPath := c.Request.URL.Path
	ctx := h.buildDetailRequestContext(c, requestPath, "anime")
	h.processDetailRequest(c, ctx)
}

// HandleEpisodeDetail handles /api/v1/episode-detail endpoint
// @Summary Get episode details from multiple sources with fallback
// @Description Retrieve detailed information about a specific episode from all available APIs. Requires episode identifier (id, episode_url, or episode_slug). Uses automatic fallback mechanism.
// @Tags Detail
// @Accept json
// @Produce json
// @Param id query string false "Episode ID (alternative to episode_url/episode_slug)"
// @Param episode_url query string false "Episode URL (alternative to id/episode_slug)"
// @Param episode_slug query string false "Episode slug (alternative to id/episode_url)"
// @Param category query string false "Content category for API routing - dynamically loaded from database" default(anime)
// @Success 200 {object} map[string]interface{} "Aggregated episode details from available sources"
// @Failure 400 {object} map[string]interface{} "Bad request - missing required episode identifier"
// @Failure 404 {object} map[string]interface{} "Not found - episode not found in any source"
// @Failure 500 {object} map[string]interface{} "Internal server error - all API sources failed"
// @Router /api/v1/episode-detail [get]
func (h *APIHandler) HandleEpisodeDetail(c *gin.Context) {
	// Validate required episode identifier parameters
	id := c.Query("id")
	episodeURL := c.Query("episode_url")
	episodeSlug := c.Query("episode_slug")

	if id == "" && episodeURL == "" && episodeSlug == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "missing required parameter: one of 'id', 'episode_url', or 'episode_slug' is required",
			"source":  "apicategorywithfallback",
		})
		return
	}

	// Use the actual request path to maintain consistency
	requestPath := c.Request.URL.Path
	ctx := h.buildDetailRequestContext(c, requestPath, "episode")
	h.processDetailRequest(c, ctx)
}

// HandleSearch handles /api/v1/search endpoint
// @Summary Search anime from multiple sources with fallback
// @Description Search for anime by query string across all available APIs. Aggregates search results from multiple sources with automatic fallback mechanism.
// @Tags Search
// @Accept json
// @Produce json
// @Param q query string true "Search query string - title, genre, or keyword to search for"
// @Param page query int false "Page number for pagination" default(1)
// @Param category query string false "Content category for API routing - dynamically loaded from database" default(anime)
// @Success 200 {object} map[string]interface{} "Aggregated search results from multiple sources with metadata"
// @Failure 400 {object} map[string]interface{} "Bad request - missing or empty search query"
// @Failure 404 {object} map[string]interface{} "Not found - no results found in any source"
// @Failure 500 {object} map[string]interface{} "Internal server error - all API sources failed"
// @Router /api/v1/search [get]
func (h *APIHandler) HandleSearch(c *gin.Context) {
	ctx := h.buildRequestContext(c, "/api/v1/search")
	h.processRequest(c, ctx)
}

// buildRequestContext creates a request context from Gin context
func (h *APIHandler) buildRequestContext(c *gin.Context, endpoint string) *domain.RequestContext {
	// Get category from query parameter, default to "anime"
	category := c.DefaultQuery("category", "anime")

	// Extract all query parameters
	parameters := make(map[string]string)
	for key, values := range c.Request.URL.Query() {
		if len(values) > 0 {
			parameters[key] = values[0]
		}
	}

	return &domain.RequestContext{
		Endpoint:   endpoint,
		Category:   category,
		Parameters: parameters,
		ClientIP:   c.ClientIP(),
		UserAgent:  c.GetHeader("User-Agent"),
		StartTime:  time.Now(),
	}
}

// processRequest processes the API request and returns response
func (h *APIHandler) processRequest(c *gin.Context, ctx *domain.RequestContext) {
	logger.Infof("Processing request: %s for category: %s", ctx.Endpoint, ctx.Category)

	response, err := h.apiService.ProcessRequest(ctx)
	if err != nil {
		logger.Errorf("Request failed: %v", err)

		// Return appropriate error response
		statusCode := http.StatusServiceUnavailable
		if err.Error() == "rate limit exceeded" {
			statusCode = http.StatusTooManyRequests
		}

		c.JSON(statusCode, gin.H{
			"error":   true,
			"message": err.Error(),
			"source":  "apicategorywithfallback",
		})
		return
	}

	// Return successful response
	c.Header("Content-Type", "application/json")
	c.Header("X-Source", response.SourceName)
	c.Header("X-Response-Time", response.ResponseTime.String())

	// If response is from cache, add cache header
	if response.SourceName == "cache" {
		c.Header("X-Cache", "HIT")
	} else {
		c.Header("X-Cache", "MISS")
	}

	c.Data(http.StatusOK, "application/json", response.Data)
}

// createEnhancedResponse creates an enhanced response with metadata
func (h *APIHandler) createEnhancedResponse(ctx *domain.RequestContext, response *domain.APIResponse, startTime time.Time, allSources []string, attempts int) *domain.EnhancedResponse {
	// Parse original response data
	var originalData interface{}
	if err := json.Unmarshal(response.Data, &originalData); err != nil {
		// If we can't parse JSON, return raw data as string
		originalData = string(response.Data)
	}

	// Calculate total time
	totalTime := time.Since(startTime)

	// Determine cache status
	cacheStatus := "MISS"
	if response.SourceName == "cache" {
		cacheStatus = "HIT"
	}

	// Create filter description
	filterApplied := fmt.Sprintf("Category filter: '%s' applied to select appropriate API sources", ctx.Category)
	if len(ctx.Parameters) > 0 {
		filterApplied += fmt.Sprintf(", Parameters filtered: %d parameters processed", len(ctx.Parameters))
	}

	return &domain.EnhancedResponse{
		Data:    originalData,
		Success: true,
		Metadata: domain.ResponseMetadata{
			Source:        response.SourceName,
			SourceURL:     fmt.Sprintf("%s%s", response.SourceName, ctx.Endpoint), // This will be updated in service layer
			AllSources:    allSources,
			Category:      ctx.Category,
			Endpoint:      ctx.Endpoint,
			FilterApplied: filterApplied,
			ResponseTime:  response.ResponseTime.String(),
			TotalTime:     totalTime.String(),
			Attempts:      attempts,
			CacheStatus:   cacheStatus,
			Timestamp:     time.Now().Format(time.RFC3339),
		},
	}
}

// sendEnhancedResponse sends an enhanced JSON response with metadata
func (h *APIHandler) sendEnhancedResponse(c *gin.Context, enhancedResponse *domain.EnhancedResponse) {
	// Set standard headers
	c.Header("Content-Type", "application/json")
	c.Header("X-Source", enhancedResponse.Metadata.Source)
	c.Header("X-Response-Time", enhancedResponse.Metadata.ResponseTime)
	c.Header("X-Endpoint", enhancedResponse.Metadata.Endpoint)
	c.Header("X-Cache", enhancedResponse.Metadata.CacheStatus)
	c.Header("X-Category", enhancedResponse.Metadata.Category)
	c.Header("X-Total-Time", enhancedResponse.Metadata.TotalTime)
	c.Header("X-Attempts", fmt.Sprintf("%d", enhancedResponse.Metadata.Attempts))

	c.JSON(http.StatusOK, enhancedResponse)
}

// buildDetailRequestContext creates a specialized request context for detail endpoints
func (h *APIHandler) buildDetailRequestContext(c *gin.Context, endpoint string, detailType string) *domain.RequestContext {
	// Get category from query parameter, default to "anime"
	category := c.DefaultQuery("category", "anime")

	// Extract all query parameters
	parameters := make(map[string]string)
	for key, values := range c.Request.URL.Query() {
		if len(values) > 0 {
			parameters[key] = values[0]
		}
	}

	// Map common parameter variations for different APIs
	h.normalizeDetailParameters(parameters, detailType)

	return &domain.RequestContext{
		Endpoint:   endpoint,
		Category:   category,
		Parameters: parameters,
		ClientIP:   c.ClientIP(),
		UserAgent:  c.GetHeader("User-Agent"),
		StartTime:  time.Now(),
	}
}

// normalizeDetailParameters maps different parameter names to common ones for API compatibility
func (h *APIHandler) normalizeDetailParameters(params map[string]string, detailType string) {
	if detailType == "anime" {
		// For anime detail, ensure we have the right parameter names for different APIs
		if id := params["id"]; id != "" {
			params["anime_slug"] = id // Map id to anime_slug for APIs that expect this
			params["slug"] = id       // Map id to slug for APIs that expect this
		}
		if slug := params["slug"]; slug != "" {
			params["anime_slug"] = slug // Map slug to anime_slug
			params["id"] = slug         // Map slug to id for APIs that expect this
		}
		if animeSlug := params["anime_slug"]; animeSlug != "" {
			params["slug"] = animeSlug // Map anime_slug to slug
			params["id"] = animeSlug   // Map anime_slug to id
		}
	} else if detailType == "episode" {
		// For episode detail, ensure we have the right parameter names
		if id := params["id"]; id != "" {
			params["episode_url"] = id  // Map id to episode_url for APIs that expect this
			params["episode_slug"] = id // Map id to episode_slug for APIs that expect this
		}
		if episodeURL := params["episode_url"]; episodeURL != "" {
			params["id"] = episodeURL           // Map episode_url to id
			params["episode_slug"] = episodeURL // Map episode_url to episode_slug
		}
		if episodeSlug := params["episode_slug"]; episodeSlug != "" {
			params["id"] = episodeSlug          // Map episode_slug to id
			params["episode_url"] = episodeSlug // Map episode_slug to episode_url
		}
	}
}

// processDetailRequest processes detail endpoint requests with enhanced aggregation
func (h *APIHandler) processDetailRequest(c *gin.Context, ctx *domain.RequestContext) {
	startTime := time.Now()
	logger.Infof("Processing detail request: %s for category: %s with params: %+v", ctx.Endpoint, ctx.Category, ctx.Parameters)

	response, err := h.apiService.ProcessRequest(ctx)
	if err != nil {
		logger.Errorf("Detail request failed: %v", err)

		// Create enhanced error response
		allSources := []string{"no sources available"}
		attempts := 0

		// Try to extract source info from error if available
		if response != nil {
			allSources = response.AllSourcesAttempted
			attempts = response.TotalAttempts
		}

		enhancedError := createEnhancedErrorResponse(ctx, err, startTime, allSources, attempts)

		// Return appropriate status code
		statusCode := http.StatusServiceUnavailable
		if err.Error() == "rate limit exceeded" {
			statusCode = http.StatusTooManyRequests
		} else if err.Error() == "no API sources configured for endpoint" {
			statusCode = http.StatusNotFound
		} else if err.Error() == "missing required parameters" {
			statusCode = http.StatusBadRequest
		}

		c.JSON(statusCode, enhancedError)
		return
	}

	// Create enhanced success response
	enhancedResponse := createEnhancedResponse(ctx, response, startTime, response.AllSourcesAttempted, response.TotalAttempts)

	// Send enhanced response with all metadata
	sendEnhancedResponse(c, enhancedResponse)
}
