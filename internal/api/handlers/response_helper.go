package handlers

import (
	"encoding/json"
	"fmt"
	"time"

	"apicategorywithfallback/internal/domain"

	"github.com/gin-gonic/gin"
)

// createEnhancedResponse creates an enhanced response with metadata
func createEnhancedResponse(ctx *domain.RequestContext, response *domain.APIResponse, startTime time.Time, allSources []string, attempts int) *domain.EnhancedResponse {
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

	// Create source URL from response if available
	sourceURL := "N/A"
	if response.SourceName != "cache" && response.SourceName != "" {
		sourceURL = fmt.Sprintf("Internal API routing to %s%s", response.SourceName, ctx.Endpoint)
	}

	return &domain.EnhancedResponse{
		Data:    originalData,
		Success: true,
		Metadata: domain.ResponseMetadata{
			Source:        response.SourceName,
			SourceURL:     sourceURL,
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
func sendEnhancedResponse(c *gin.Context, enhancedResponse *domain.EnhancedResponse) {
	// Set standard headers
	c.Header("Content-Type", "application/json")
	c.Header("X-Source", enhancedResponse.Metadata.Source)
	c.Header("X-Response-Time", enhancedResponse.Metadata.ResponseTime)
	c.Header("X-Endpoint", enhancedResponse.Metadata.Endpoint)
	c.Header("X-Cache", enhancedResponse.Metadata.CacheStatus)
	c.Header("X-Category", enhancedResponse.Metadata.Category)
	c.Header("X-Total-Time", enhancedResponse.Metadata.TotalTime)
	c.Header("X-Attempts", fmt.Sprintf("%d", enhancedResponse.Metadata.Attempts))
	c.Header("X-All-Sources", fmt.Sprintf("%v", enhancedResponse.Metadata.AllSources))

	c.JSON(200, enhancedResponse)
}

// createEnhancedErrorResponse creates an enhanced error response with metadata
func createEnhancedErrorResponse(ctx *domain.RequestContext, err error, startTime time.Time, allSources []string, attempts int) *domain.EnhancedResponse {
	totalTime := time.Since(startTime)

	filterApplied := fmt.Sprintf("Category filter: '%s' applied to select appropriate API sources", ctx.Category)
	if len(ctx.Parameters) > 0 {
		filterApplied += fmt.Sprintf(", Parameters filtered: %d parameters processed", len(ctx.Parameters))
	}

	return &domain.EnhancedResponse{
		Data:    nil,
		Success: false,
		Error:   "API request failed",
		Message: err.Error(),
		Metadata: domain.ResponseMetadata{
			Source:        "apicategorywithfallback",
			SourceURL:     "N/A",
			AllSources:    allSources,
			Category:      ctx.Category,
			Endpoint:      ctx.Endpoint,
			FilterApplied: filterApplied,
			ResponseTime:  "0s",
			TotalTime:     totalTime.String(),
			Attempts:      attempts,
			CacheStatus:   "BYPASS",
			Timestamp:     time.Now().Format(time.RFC3339),
		},
	}
}
