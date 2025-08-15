package handlers

import (
	"encoding/json"
	"fmt"
	"time"

	"apicategorywithfallback/internal/domain"
	"apicategorywithfallback/pkg/logger"

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

	// Unwrap nested API responses (like gomunime) to prevent double nesting
	originalData = unwrapNestedAPIResponse(originalData, response.SourceName)

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

// unwrapNestedAPIResponse dynamically unwraps nested API responses to prevent double nesting
// This automatically detects and handles cases where APIs return {"data": {actual_data}, "metadata": ...}
func unwrapNestedAPIResponse(responseData interface{}, sourceName string) interface{} {
	logger.Infof("🔧 Analyzing response structure from %s for potential unwrapping", sourceName)

	// Try to cast to map for inspection
	responseMap, ok := responseData.(map[string]interface{})
	if !ok {
		logger.Infof("❌ Response from %s is not a map, returning as-is", sourceName)
		return responseData
	}

	// First, check if this is an aggregated response that needs category-level unwrapping
	if hasAggregatedStructure := hasDataByCategory(responseMap); hasAggregatedStructure {
		logger.Infof("🔍 Detected aggregated response structure from %s, checking for nested categories", sourceName)
		unwrappedAggregated := unwrapAggregatedResponse(responseMap, sourceName)
		return unwrappedAggregated
	}

	// Detect if this response needs unwrapping using dynamic pattern detection
	if !shouldUnwrapResponse(responseMap, sourceName) {
		logger.Infof("✅ Response from %s doesn't need unwrapping, structure is already flat", sourceName)
		return responseData
	}

	// Extract data and metadata
	dataField := responseMap["data"]
	dataMap := dataField.(map[string]interface{})

	logger.Infof("✅ Detected nested API response pattern from %s - proceeding with unwrapping", sourceName)
	logger.Infof("🔄 Extracting inner data and preserving metadata from %s", sourceName)

	// Extract the inner data and add metadata from the outer level
	unwrappedData := make(map[string]interface{})

	// Copy all data from the inner "data" field
	for key, value := range dataMap {
		unwrappedData[key] = value
	}

	// Preserve ALL metadata from the outer level (dynamic preservation)
	metadataFields := extractMetadataFields(responseMap)
	for key, value := range metadataFields {
		// Avoid overwriting fields from inner data
		if _, exists := unwrappedData[key]; !exists {
			unwrappedData[key] = value
		} else {
			// If conflict, preserve inner data but add metadata with prefix
			unwrappedData["_"+key] = value
		}
	}

	logger.Infof("🎯 Successfully unwrapped response from %s: moved %d fields from data.*, preserved %d metadata fields",
		sourceName, len(dataMap), len(metadataFields))

	return unwrappedData
}

// shouldUnwrapResponse uses dynamic pattern detection to determine if a response needs unwrapping
func shouldUnwrapResponse(responseMap map[string]interface{}, sourceName string) bool {
	// Pattern 1: Must have a "data" field
	dataField, hasData := responseMap["data"]
	if !hasData {
		logger.Infof("🔍 Pattern check: No 'data' field found in %s response", sourceName)
		return false
	}

	// Pattern 2: The "data" field must be a map (not a simple value)
	dataMap, isDataMap := dataField.(map[string]interface{})
	if !isDataMap {
		logger.Infof("🔍 Pattern check: 'data' field from %s is not a map/object", sourceName)
		return false
	}

	// Pattern 3: The data map should contain substantial content (not just metadata)
	if len(dataMap) < 2 {
		logger.Infof("🔍 Pattern check: 'data' field from %s has too few fields (%d), likely not actual content", sourceName, len(dataMap))
		return false
	}

	// Debug: show what fields are in the inner data
	innerFields := make([]string, 0, len(dataMap))
	for key := range dataMap {
		innerFields = append(innerFields, key)
	}
	logger.Infof("🔍 Inner data fields from %s: %v", sourceName, innerFields)

	// Pattern 4: Response should have metadata fields at root level (indicators of wrapping)
	metadataIndicators := []string{
		"confidence_score", "message", "source", "status", "success", "error", "code",
		// Additional API metadata indicators
		"timestamp", "api_version", "response_time", "cache", "hash",
		"request_id", "server", "method", "endpoint", "user_agent",
	}
	metadataCount := 0
	for _, indicator := range metadataIndicators {
		if _, exists := responseMap[indicator]; exists {
			metadataCount++
		}
	}

	if metadataCount < 1 {
		logger.Infof("🔍 Pattern check: No metadata indicators found in %s response, likely already flat", sourceName)
		return false
	}

	// Debug: show what metadata fields were found
	foundMetadata := make([]string, 0)
	for _, indicator := range metadataIndicators {
		if _, exists := responseMap[indicator]; exists {
			foundMetadata = append(foundMetadata, indicator)
		}
	}
	logger.Infof("🔍 Found metadata indicators from %s: %v", sourceName, foundMetadata)

	// Pattern 5: The inner data should look like actual content (has typical content fields)
	// Enhanced indicators for both anime detail and episode detail
	contentIndicators := []string{
		// Common fields
		"title", "name", "id", "slug", "url",
		// Anime detail specific
		"anime_slug", "cover", "genre", "rating", "synopsis", "status", "type",
		// Episode detail specific
		"anime_info", "download_links", "streaming_servers", "navigation",
		"other_episodes", "release_info", "thumbnail_url", "episode_number",
		// Additional content fields
		"description", "image", "poster", "year", "studio",
	}
	contentCount := 0
	for _, indicator := range contentIndicators {
		if _, exists := dataMap[indicator]; exists {
			contentCount++
		}
	}

	if contentCount < 1 {
		logger.Infof("🔍 Pattern check: Inner 'data' from %s doesn't contain enough content indicators (%d), might not be actual content", sourceName, contentCount)
		return false
	}

	// Debug: show what content indicators were found
	foundContent := make([]string, 0)
	for _, indicator := range contentIndicators {
		if _, exists := dataMap[indicator]; exists {
			foundContent = append(foundContent, indicator)
		}
	}
	logger.Infof("🔍 Found content indicators from %s: %v", sourceName, foundContent)

	logger.Infof("🎯 Pattern detection SUCCESS for %s: data=%d fields, metadata=%d indicators, content=%d indicators - WILL UNWRAP",
		sourceName, len(dataMap), metadataCount, contentCount)
	return true
}

// extractMetadataFields extracts metadata fields from the response, excluding the "data" field
func extractMetadataFields(responseMap map[string]interface{}) map[string]interface{} {
	metadata := make(map[string]interface{})

	// Common metadata field names to preserve
	metadataKeys := []string{
		"confidence_score", "message", "source", "status", "success", "error", "code",
		"timestamp", "version", "api_version", "total", "count", "page", "limit",
		"response_time", "cache", "signature", "hash", "checksum",
	}

	// Extract known metadata fields
	for _, key := range metadataKeys {
		if value, exists := responseMap[key]; exists {
			metadata[key] = value
		}
	}

	// Also extract any field that looks like metadata (starts with underscore or ends with common suffixes)
	for key, value := range responseMap {
		if key == "data" { // Skip the data field
			continue
		}

		// Already processed above
		if _, alreadyAdded := metadata[key]; alreadyAdded {
			continue
		}

		// Fields that look like metadata
		if isMetadataField(key) {
			metadata[key] = value
		}
	}

	return metadata
}

// isMetadataField determines if a field name looks like metadata
func isMetadataField(fieldName string) bool {
	// Fields starting with underscore
	if len(fieldName) > 0 && fieldName[0] == '_' {
		return true
	}

	// Fields ending with common metadata suffixes
	metadataSuffixes := []string{"_time", "_count", "_total", "_status", "_code", "_version", "_id", "_key"}
	for _, suffix := range metadataSuffixes {
		if len(fieldName) > len(suffix) && fieldName[len(fieldName)-len(suffix):] == suffix {
			return true
		}
	}

	// Fields that are typically metadata
	metadataNames := []string{"meta", "info", "debug", "trace", "log"}
	for _, name := range metadataNames {
		if fieldName == name {
			return true
		}
	}

	return false
}

// unwrapAggregatedResponse handles nested structure within aggregated responses (category=all)
func unwrapAggregatedResponse(responseMap map[string]interface{}, sourceName string) map[string]interface{} {
	dataInterface, exists := responseMap["data"]
	if !exists {
		return responseMap // No data field, nothing to unwrap
	}

	dataMap, ok := dataInterface.(map[string]interface{})
	if !ok {
		return responseMap // Data is not a map, nothing to unwrap
	}

	// Check if this is an aggregated response with data_by_category
	dataByCategoryInterface, exists := dataMap["data_by_category"]
	if !exists {
		return responseMap // Not an aggregated response
	}

	dataByCategoryMap, ok := dataByCategoryInterface.(map[string]interface{})
	if !ok {
		return responseMap // data_by_category is not a map
	}

	logger.Infof("🔧 Checking aggregated response from %s for nested structures in data_by_category", sourceName)

	// Process each category within data_by_category
	hasUnwrapped := false
	for category, categoryDataInterface := range dataByCategoryMap {
		categoryDataMap, ok := categoryDataInterface.(map[string]interface{})
		if !ok {
			continue // Category data is not a map, skip
		}

		logger.Infof("🔍 Analyzing category '%s' from %s for nested structure", category, sourceName)

		// Check if this category data has nested structure
		if shouldUnwrapResponse(categoryDataMap, fmt.Sprintf("%s:%s", sourceName, category)) {
			logger.Infof("🎯 Unwrapping nested structure in category '%s' from %s", category, sourceName)

			unwrappedCategoryData := performCategoryUnwrapping(categoryDataMap, sourceName, category)
			dataByCategoryMap[category] = unwrappedCategoryData
			hasUnwrapped = true
		}
	}

	if hasUnwrapped {
		logger.Infof("✅ Successfully unwrapped aggregated response from %s", sourceName)

		// Update the response with unwrapped categories
		dataMap["data_by_category"] = dataByCategoryMap
		responseMap["data"] = dataMap
	}

	return responseMap
}

// hasDataByCategory checks if the response has aggregated structure with data_by_category
func hasDataByCategory(responseMap map[string]interface{}) bool {
	logger.Infof("🔍 Checking if response has aggregated structure...")

	dataInterface, exists := responseMap["data"]
	if !exists {
		logger.Infof("🔍 No 'data' field found, not aggregated")
		return false
	}

	dataMap, ok := dataInterface.(map[string]interface{})
	if !ok {
		logger.Infof("🔍 'data' field is not a map, not aggregated")
		return false
	}

	_, hasDataByCategory := dataMap["data_by_category"]
	logger.Infof("🔍 Has 'data_by_category' field: %v", hasDataByCategory)
	return hasDataByCategory
}

// performCategoryUnwrapping extracts nested data from a single category within aggregated response
func performCategoryUnwrapping(categoryDataMap map[string]interface{}, sourceName, category string) map[string]interface{} {
	// Extract data and metadata
	dataField := categoryDataMap["data"]
	innerDataMap, ok := dataField.(map[string]interface{})
	if !ok {
		logger.Infof("❌ Unable to extract nested data from category '%s' in %s, returning as-is", category, sourceName)
		return categoryDataMap
	}

	logger.Infof("✅ Detected nested structure in category '%s' from %s - proceeding with unwrapping", category, sourceName)
	logger.Infof("🔄 Extracting inner data and preserving metadata for category '%s' from %s", category, sourceName)

	// Extract the inner data and add metadata from the outer level
	unwrappedData := make(map[string]interface{})

	// Copy all data from the inner "data" field
	for key, value := range innerDataMap {
		unwrappedData[key] = value
	}

	// Preserve ALL metadata from the category level (dynamic preservation)
	metadataFields := extractMetadataFields(categoryDataMap)
	for key, value := range metadataFields {
		// Avoid overwriting fields from inner data
		if _, exists := unwrappedData[key]; !exists {
			unwrappedData[key] = value
		} else {
			// If conflict, preserve inner data but add metadata with prefix
			unwrappedData["_"+key] = value
		}
	}

	logger.Infof("🎯 Successfully unwrapped category '%s' from %s: moved %d fields from data.*, preserved %d metadata fields",
		category, sourceName, len(innerDataMap), len(metadataFields))

	return unwrappedData
}
