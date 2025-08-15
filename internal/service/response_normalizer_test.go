package service

import (
	"encoding/json"
	"testing"

	"apicategorywithfallback/pkg/config"
	"apicategorywithfallback/pkg/database"
)

func TestNormalizeResponseStructure(t *testing.T) {
	// Create a test service
	cfg := &config.Config{}
	db := &database.DB{} // This would normally be properly initialized
	service := &APIService{
		config: cfg,
	}

	tests := []struct {
		name           string
		input          string
		sourceName     string
		expectedOutput map[string]interface{}
		shouldChange   bool
	}{
		{
			name:       "Gomunime nested structure should be flattened",
			sourceName: "gomunime.co",
			input: `{
				"data": {
					"confidence_score": 1,
					"data": {
						"anime_slug": "example-anime",
						"cover": "https://example.com/cover.jpg",
						"details": {
							"Duration": "24 min",
							"English": "Example Anime",
							"Japanese": "エグザンプル"
						}
					},
					"message": "Data berhasil diambil",
					"source": "gomunime.co"
				}
			}`,
			shouldChange: true,
			expectedOutput: map[string]interface{}{
				"data": map[string]interface{}{
					"anime_slug": "example-anime",
					"cover":      "https://example.com/cover.jpg",
					"details": map[string]interface{}{
						"Duration": "24 min",
						"English":  "Example Anime",
						"Japanese": "エグザンプル",
					},
					"confidence_score": float64(1),
					"message":          "Data berhasil diambil",
					"source":           "gomunime.co",
				},
			},
		},
		{
			name:       "Samehadaku flat structure should remain unchanged",
			sourceName: "samehadaku.how",
			input: `{
				"data": {
					"anime_slug": "example-anime",
					"cover": "https://example.com/cover.jpg",
					"details": {
						"Duration": "24 min",
						"English": "Example Anime",
						"Japanese": "エグザンプル"
					},
					"confidence_score": 1,
					"message": "Data berhasil diambil",
					"source": "samehadaku.how"
				}
			}`,
			shouldChange: false,
			expectedOutput: map[string]interface{}{
				"data": map[string]interface{}{
					"anime_slug": "example-anime",
					"cover":      "https://example.com/cover.jpg",
					"details": map[string]interface{}{
						"Duration": "24 min",
						"English":  "Example Anime",
						"Japanese": "エグザンプル",
					},
					"confidence_score": float64(1),
					"message":          "Data berhasil diambil",
					"source":           "samehadaku.how",
				},
			},
		},
		{
			name:       "Complex nested structure with episode_list and recommendations",
			sourceName: "gomunime.co",
			input: `{
				"data": {
					"confidence_score": 1,
					"data": {
						"anime_slug": "example-anime",
						"cover": "https://example.com/cover.jpg",
						"episode_list": [
							{
								"episode": "1",
								"episode_slug": "episode-1",
								"title": "First Episode"
							}
						],
						"recommendations": [
							{
								"anime_slug": "rec-anime",
								"cover_url": "https://example.com/rec.jpg"
							}
						]
					},
					"message": "Data berhasil diambil",
					"source": "gomunime.co"
				}
			}`,
			shouldChange: true,
			expectedOutput: map[string]interface{}{
				"data": map[string]interface{}{
					"anime_slug": "example-anime",
					"cover":      "https://example.com/cover.jpg",
					"episode_list": []interface{}{
						map[string]interface{}{
							"episode":      "1",
							"episode_slug": "episode-1",
							"title":        "First Episode",
						},
					},
					"recommendations": []interface{}{
						map[string]interface{}{
							"anime_slug": "rec-anime",
							"cover_url":  "https://example.com/rec.jpg",
						},
					},
					"confidence_score": float64(1),
					"message":          "Data berhasil diambil",
					"source":           "gomunime.co",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.normalizeResponseStructure([]byte(tt.input), tt.sourceName)
			if err != nil {
				t.Fatalf("normalizeResponseStructure() failed: %v", err)
			}

			var resultMap map[string]interface{}
			err = json.Unmarshal(result, &resultMap)
			if err != nil {
				t.Fatalf("Failed to unmarshal result: %v", err)
			}

			// Compare the data structure
			expectedData := tt.expectedOutput["data"].(map[string]interface{})
			actualData := resultMap["data"].(map[string]interface{})

			// Check key fields
			for key, expectedValue := range expectedData {
				actualValue, exists := actualData[key]
				if !exists {
					t.Errorf("Expected key %s not found in result", key)
					continue
				}

				// For simple values, compare directly
				if key == "anime_slug" || key == "cover" || key == "message" || key == "source" {
					if actualValue != expectedValue {
						t.Errorf("Field %s: expected %v, got %v", key, expectedValue, actualValue)
					}
				}

				// For confidence_score, compare as float
				if key == "confidence_score" {
					if actualValue != expectedValue {
						t.Errorf("Field %s: expected %v, got %v", key, expectedValue, actualValue)
					}
				}
			}
		})
	}
}

func TestNormalizeDataFields(t *testing.T) {
	cfg := &config.Config{}
	service := &APIService{
		config: cfg,
	}

	tests := []struct {
		name       string
		input      map[string]interface{}
		sourceName string
		expected   map[string]interface{}
	}{
		{
			name: "Should add missing confidence_score and source",
			input: map[string]interface{}{
				"anime_slug": "test-anime",
				"cover":      "https://example.com/cover.jpg",
			},
			sourceName: "test-source",
			expected: map[string]interface{}{
				"anime_slug":       "test-anime",
				"cover":            "https://example.com/cover.jpg",
				"confidence_score": 1.0,
				"source":           "test-source",
			},
		},
		{
			name: "Should preserve existing confidence_score and source",
			input: map[string]interface{}{
				"anime_slug":       "test-anime",
				"confidence_score": 0.8,
				"source":           "original-source",
			},
			sourceName: "test-source",
			expected: map[string]interface{}{
				"anime_slug":       "test-anime",
				"confidence_score": 0.8,
				"source":           "original-source",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.normalizeDataFields(tt.input, tt.sourceName)

			for key, expectedValue := range tt.expected {
				actualValue, exists := result[key]
				if !exists {
					t.Errorf("Expected key %s not found in result", key)
					continue
				}

				if actualValue != expectedValue {
					t.Errorf("Field %s: expected %v, got %v", key, expectedValue, actualValue)
				}
			}
		})
	}
}
