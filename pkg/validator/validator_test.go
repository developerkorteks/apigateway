package validator

import (
	"encoding/json"
	"testing"
)

func TestValidateHomeResponse(t *testing.T) {
	// Valid response
	validResponse := map[string]interface{}{
		"confidence_score": 0.8,
		"message":          "success",
		"source":           "test_source",
		"top10": []interface{}{
			map[string]interface{}{
				"judul":      "Test Anime",
				"url":        "https://example.com/anime/1",
				"anime_slug": "test-anime",
				"cover":      "https://example.com/cover.jpg",
			},
		},
		"new_eps": []interface{}{
			map[string]interface{}{
				"judul":      "Test Episode",
				"url":        "https://example.com/episode/1",
				"anime_slug": "test-anime",
				"episode":    "1",
				"cover":      "https://example.com/cover.jpg",
			},
		},
		"movies": []interface{}{},
		"jadwal_rilis": map[string]interface{}{
			"Monday": []interface{}{},
		},
	}

	data, _ := json.Marshal(validResponse)
	err := ValidateResponse("/api/v1/home", data)
	if err != nil {
		t.Errorf("Valid response should not return error: %v", err)
	}

	// Invalid confidence score
	invalidResponse := validResponse
	invalidResponse["confidence_score"] = 0.3

	data, _ = json.Marshal(invalidResponse)
	err = ValidateResponse("/api/v1/home", data)
	if err == nil {
		t.Errorf("Invalid confidence score should return error")
	}
}

func TestValidateSearchResponse(t *testing.T) {
	// Valid search response
	validResponse := map[string]interface{}{
		"confidence_score": 0.9,
		"message":          "success",
		"source":           "test_source",
		"data": []interface{}{
			map[string]interface{}{
				"judul":      "Search Result",
				"url":        "https://example.com/anime/search",
				"anime_slug": "search-result",
				"cover":      "https://example.com/cover.jpg",
				"status":     "ongoing",
				"tipe":       "TV",
				"skor":       "8.5",
				"genre":      []interface{}{"Action", "Adventure"},
			},
		},
	}

	data, _ := json.Marshal(validResponse)
	err := ValidateResponse("/api/v1/search", data)
	if err != nil {
		t.Errorf("Valid search response should not return error: %v", err)
	}
}

func TestValidateAnimeDetailResponse(t *testing.T) {
	// Valid anime detail response
	validResponse := map[string]interface{}{
		"confidence_score": 0.85,
		"message":          "success",
		"source":           "test_source",
		"judul":            "Test Anime Detail",
		"url":              "https://example.com/anime/detail",
		"anime_slug":       "test-anime-detail",
		"cover":            "https://example.com/cover.jpg",
		"episode_list": []interface{}{
			map[string]interface{}{
				"episode":      "1",
				"title":        "Episode 1",
				"url":          "https://example.com/episode/1",
				"episode_slug": "episode-1",
				"release_date": "2024-01-01",
			},
		},
		"recommendations": []interface{}{},
		"status":          "completed",
		"tipe":            "TV",
		"skor":            "9.0",
		"genre":           []interface{}{"Action"},
		"details": map[string]interface{}{
			"Japanese": "テストアニメ",
			"Status":   "Completed",
		},
		"rating": map[string]interface{}{
			"score": "9.0",
			"users": "1000",
		},
	}

	data, _ := json.Marshal(validResponse)
	err := ValidateResponse("/api/v1/anime-detail", data)
	if err != nil {
		t.Errorf("Valid anime detail response should not return error: %v", err)
	}
}

func TestValidateEpisodeDetailResponse(t *testing.T) {
	// Valid episode detail response
	validResponse := map[string]interface{}{
		"confidence_score": 0.9,
		"message":          "success",
		"source":           "test_source",
		"title":            "Test Episode",
		"thumbnail_url":    "https://example.com/thumb.jpg",
		"streaming_servers": []interface{}{
			map[string]interface{}{
				"server_name":   "Server 1",
				"streaming_url": "https://example.com/stream/1",
			},
		},
		"release_info": "Released on 2024-01-01",
		"download_links": map[string]interface{}{
			"MP4": map[string]interface{}{
				"720p": []interface{}{
					map[string]interface{}{
						"provider": "Provider 1",
						"url":      "https://example.com/download/1",
					},
				},
			},
		},
		"navigation": map[string]interface{}{
			"previous_episode_url": nil,
			"all_episodes_url":     "https://example.com/episodes",
			"next_episode_url":     "https://example.com/episode/2",
		},
		"anime_info": map[string]interface{}{
			"title":         "Test Anime",
			"thumbnail_url": "https://example.com/thumb.jpg",
			"synopsis":      "Test synopsis",
			"genres":        []interface{}{"Action"},
		},
		"other_episodes": []interface{}{},
	}

	data, _ := json.Marshal(validResponse)
	err := ValidateResponse("/api/v1/episode-detail", data)
	if err != nil {
		t.Errorf("Valid episode detail response should not return error: %v", err)
	}
}

func TestValidateInvalidEndpoint(t *testing.T) {
	data := []byte(`{"test": "data"}`)
	err := ValidateResponse("/invalid/endpoint", data)
	if err == nil {
		t.Errorf("Invalid endpoint should return error")
	}
}

func TestValidateInvalidJSON(t *testing.T) {
	data := []byte(`invalid json`)
	err := ValidateResponse("/api/v1/home", data)
	if err == nil {
		t.Errorf("Invalid JSON should return error")
	}
}

func TestValidateMissingRequiredFields(t *testing.T) {
	// Missing required fields
	invalidResponse := map[string]interface{}{
		"confidence_score": 0.8,
		"message":          "success",
		"source":           "test_source",
		// Missing top10, new_eps, movies, jadwal_rilis
	}

	data, _ := json.Marshal(invalidResponse)
	err := ValidateResponse("/api/v1/home", data)
	if err == nil {
		t.Errorf("Missing required fields should return error")
	}
}

func TestValidateURLFields(t *testing.T) {
	// Invalid URL
	invalidResponse := map[string]interface{}{
		"confidence_score": 0.8,
		"message":          "success",
		"source":           "test_source",
		"top10": []interface{}{
			map[string]interface{}{
				"judul":      "Test Anime",
				"url":        "invalid-url", // Invalid URL
				"anime_slug": "test-anime",
				"cover":      "https://example.com/cover.jpg",
			},
		},
		"new_eps":      []interface{}{},
		"movies":       []interface{}{},
		"jadwal_rilis": map[string]interface{}{},
	}

	data, _ := json.Marshal(invalidResponse)
	err := ValidateResponse("/api/v1/home", data)
	if err == nil {
		t.Errorf("Invalid URL should return error")
	}
}
