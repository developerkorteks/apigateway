package domain

import (
	"errors"
	"time"
)

// Common errors
var (
	ErrAllAPIsFailed   = errors.New("all APIs failed to respond")
	ErrInvalidEndpoint = errors.New("invalid endpoint")
	ErrCacheNotFound   = errors.New("cache entry not found")
)

// APIRequest represents a request to an external API
type APIRequest struct {
	URL        string
	Method     string
	Headers    map[string]string
	Timeout    time.Duration
	SourceName string
	Priority   int
	IsFallback bool
}

// APIResponse represents a response from an external API
type APIResponse struct {
	Data         []byte
	StatusCode   int
	ResponseTime time.Duration
	SourceName   string
	Error        error
	IsFallback   bool

	// Enhanced metadata for response tracking
	AllSourcesAttempted []string // All API sources that were attempted
	TotalAttempts       int      // Total number of attempts made
	ActualSourceURL     string   // The actual URL that was called successfully
}

// EnhancedResponse represents an enhanced response with source metadata
type EnhancedResponse struct {
	// Original data from downstream API
	Data interface{} `json:"data"`

	// Metadata about request/response
	Metadata ResponseMetadata `json:"_metadata"`

	// Success indicator
	Success bool `json:"success"`

	// Error information (if any)
	Error   string `json:"error,omitempty"`
	Message string `json:"message,omitempty"`
}

// ResponseMetadata contains information about the API call
type ResponseMetadata struct {
	// Source information
	Source     string   `json:"source"`            // Which API provided the data
	SourceURL  string   `json:"source_url"`        // Full URL that was called
	AllSources []string `json:"available_sources"` // All sources that were tried

	// Filter and category information
	Category      string `json:"category"`       // Category used for filtering (anime, korean-drama, etc)
	Endpoint      string `json:"endpoint"`       // The endpoint that was called
	FilterApplied string `json:"filter_applied"` // Description of filter applied

	// Performance information
	ResponseTime string `json:"response_time"` // Time taken for the successful request
	TotalTime    string `json:"total_time"`    // Total time including fallbacks
	Attempts     int    `json:"attempts"`      // Number of API calls made

	// Cache information
	CacheStatus string `json:"cache_status"`        // HIT, MISS, BYPASS
	CacheKey    string `json:"cache_key,omitempty"` // Cache key used (optional)

	// Request timestamp
	Timestamp string `json:"timestamp"` // When the request was made
}

// EndpointConfig represents configuration for an endpoint
type EndpointConfig struct {
	Path         string
	Category     string
	PrimaryAPIs  []APISource
	FallbackAPIs map[string][]string
	CacheTTL     time.Duration
}

// APISource represents an API source configuration
type APISource struct {
	ID         int
	SourceName string
	BaseURL    string
	Priority   int
	IsPrimary  bool
	IsActive   bool
}

// RequestContext represents the context of an incoming request
type RequestContext struct {
	Endpoint   string
	Category   string
	Parameters map[string]string
	ClientIP   string
	UserAgent  string
	StartTime  time.Time
}

// FallbackResult represents the result of a fallback operation
type FallbackResult struct {
	Success      bool
	Response     *APIResponse
	SourceUsed   string
	FallbackUsed bool
	TotalTime    time.Duration
}

// HealthStatus represents the health status of an API source
type HealthStatus struct {
	APISourceID  int
	SourceName   string
	Status       string // OK, TIMEOUT, ERROR
	ResponseTime int
	ErrorMessage string
	LastChecked  time.Time
}

// Statistics represents API usage statistics
type Statistics struct {
	TotalRequests       int64
	SuccessfulRequests  int64
	FailedRequests      int64
	FallbackUsage       int64
	AverageResponseTime float64
	SourceStats         map[string]SourceStats
}

// SourceStats represents statistics for a specific API source
type SourceStats struct {
	SourceName          string
	TotalRequests       int64
	SuccessRequests     int64
	FailedRequests      int64
	AverageResponseTime float64
	LastUsed            time.Time
}
