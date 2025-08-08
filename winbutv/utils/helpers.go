package utils

import (
	"regexp"
	"strings"
)

// CleanText removes extra whitespace and trims text
func CleanText(text string) string {
	re := regexp.MustCompile(`\s+`)
	return strings.TrimSpace(re.ReplaceAllString(text, " "))
}

// ExtractSlugFromURL extracts anime slug from URL
func ExtractSlugFromURL(url string) string {
	// Extract slug from URL like https://winbu.tv/anime/one-piece/
	parts := strings.Split(strings.TrimSuffix(url, "/"), "/")
	if len(parts) >= 2 {
		return parts[len(parts)-1]
	}
	return ""
}

// CalculateConfidenceScore calculates confidence score based on data completeness
func CalculateConfidenceScore(totalFields, filledFields int) float64 {
	if totalFields == 0 {
		return 0.0
	}
	score := float64(filledFields) / float64(totalFields)
	if score > 1.0 {
		return 1.0
	}
	return score
}

// IsValidURL checks if URL is valid
func IsValidURL(url string) bool {
	return strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://")
}
