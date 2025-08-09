package validator

import (
	"encoding/json"
	"fmt"
	"net/url"
	"reflect"
	"strings"
)

// BaseResponse represents the common structure of all API responses
type BaseResponse struct {
	ConfidenceScore float64 `json:"confidence_score"`
	Message         string  `json:"message"`
	Source          string  `json:"source"`
}

// HomeResponse represents the structure for /api/v1/home endpoint
type HomeResponse struct {
	BaseResponse
	Top10       []AnimeItem               `json:"top10"`
	NewEps      []EpisodeItem             `json:"new_eps"`
	Movies      []MovieItem               `json:"movies"`
	JadwalRilis map[string][]ScheduleItem `json:"jadwal_rilis"`
}

// JadwalRilisResponse represents the structure for /api/v1/jadwal-rilis endpoint
type JadwalRilisResponse struct {
	BaseResponse
	Data map[string][]ScheduleItem `json:"data"`
}

// JadwalRilisDayResponse represents the structure for /api/v1/jadwal-rilis/{day} endpoint
type JadwalRilisDayResponse struct {
	BaseResponse
	Data []ScheduleItem `json:"data"`
}

// AnimeTerbaruResponse represents the structure for /api/v1/anime-terbaru endpoint
type AnimeTerbaruResponse struct {
	BaseResponse
	Data []EpisodeItem `json:"data"`
}

// MovieResponse represents the structure for /api/v1/movie endpoint
type MovieResponse struct {
	BaseResponse
	Data []MovieItem `json:"data"`
}

// AnimeDetailResponse represents the structure for /api/v1/anime-detail endpoint
type AnimeDetailResponse struct {
	BaseResponse
	Judul           string               `json:"judul"`
	URL             string               `json:"url"`
	AnimeSlug       string               `json:"anime_slug"`
	Cover           string               `json:"cover"`
	EpisodeList     []EpisodeListItem    `json:"episode_list"`
	Recommendations []RecommendationItem `json:"recommendations"`
	Status          string               `json:"status"`
	Tipe            string               `json:"tipe"`
	Skor            string               `json:"skor"`
	Penonton        string               `json:"penonton"`
	Sinopsis        string               `json:"sinopsis"`
	Genre           []string             `json:"genre"`
	Details         AnimeDetails         `json:"details"`
	Rating          RatingInfo           `json:"rating"`
}

// EpisodeDetailResponse represents the structure for /api/v1/episode-detail endpoint
type EpisodeDetailResponse struct {
	BaseResponse
	Title            string                               `json:"title"`
	ThumbnailURL     string                               `json:"thumbnail_url"`
	StreamingServers []StreamingServer                    `json:"streaming_servers"`
	ReleaseInfo      string                               `json:"release_info"`
	DownloadLinks    map[string]map[string][]DownloadLink `json:"download_links"`
	Navigation       NavigationInfo                       `json:"navigation"`
	AnimeInfo        AnimeInfo                            `json:"anime_info"`
	OtherEpisodes    []EpisodeListItem                    `json:"other_episodes"`
}

// SearchResponse represents the structure for /api/v1/search endpoint
type SearchResponse struct {
	BaseResponse
	Data []SearchResultItem `json:"data"`
}

// Supporting structs
type AnimeItem struct {
	Judul     string   `json:"judul"`
	URL       string   `json:"url"`
	AnimeSlug string   `json:"anime_slug"`
	Rating    string   `json:"rating"`
	Cover     string   `json:"cover"`
	Genres    []string `json:"genres"`
}

type EpisodeItem struct {
	Judul     string `json:"judul"`
	URL       string `json:"url"`
	AnimeSlug string `json:"anime_slug"`
	Episode   string `json:"episode"`
	Uploader  string `json:"uploader,omitempty"`
	Rilis     string `json:"rilis"`
	Cover     string `json:"cover"`
}

type MovieItem struct {
	Judul     string   `json:"judul"`
	URL       string   `json:"url"`
	AnimeSlug string   `json:"anime_slug"`
	Status    string   `json:"status,omitempty"`
	Skor      string   `json:"skor,omitempty"`
	Sinopsis  string   `json:"sinopsis,omitempty"`
	Views     string   `json:"views,omitempty"`
	Cover     string   `json:"cover"`
	Genres    []string `json:"genres"`
	Tanggal   string   `json:"tanggal"`
}

type ScheduleItem struct {
	Title       string   `json:"title"`
	URL         string   `json:"url"`
	AnimeSlug   string   `json:"anime_slug"`
	CoverURL    string   `json:"cover_url"`
	Type        string   `json:"type"`
	Score       string   `json:"score"`
	Genres      []string `json:"genres"`
	ReleaseTime string   `json:"release_time"`
}

type EpisodeListItem struct {
	Episode      string `json:"episode"`
	Title        string `json:"title"`
	URL          string `json:"url"`
	EpisodeSlug  string `json:"episode_slug,omitempty"`
	ReleaseDate  string `json:"release_date"`
	ThumbnailURL string `json:"thumbnail_url,omitempty"`
}

type RecommendationItem struct {
	Title     string `json:"title"`
	URL       string `json:"url"`
	AnimeSlug string `json:"anime_slug"`
	CoverURL  string `json:"cover_url"`
	Rating    string `json:"rating"`
	Episode   string `json:"episode"`
}

type AnimeDetails struct {
	Japanese     string `json:"Japanese"`
	Synonyms     string `json:"Synonyms"`
	English      string `json:"English"`
	Status       string `json:"Status"`
	Type         string `json:"Type"`
	Source       string `json:"Source"`
	Duration     string `json:"Duration"`
	TotalEpisode string `json:"Total Episode"`
	Studio       string `json:"Studio"`
	Producers    string `json:"Producers"`
	Released     string `json:"Released:"`
}

type RatingInfo struct {
	Score string `json:"score"`
	Users string `json:"users"`
}

type StreamingServer struct {
	ServerName   string `json:"server_name"`
	StreamingURL string `json:"streaming_url"`
}

type DownloadLink struct {
	Provider string `json:"provider"`
	URL      string `json:"url"`
}

type NavigationInfo struct {
	PreviousEpisodeURL string `json:"previous_episode_url"`
	AllEpisodesURL     string `json:"all_episodes_url"`
	NextEpisodeURL     string `json:"next_episode_url"`
}

type AnimeInfo struct {
	Title        string   `json:"title"`
	ThumbnailURL string   `json:"thumbnail_url"`
	Synopsis     string   `json:"synopsis"`
	Genres       []string `json:"genres"`
}

type SearchResultItem struct {
	Judul     string   `json:"judul"`
	URL       string   `json:"url"`
	AnimeSlug string   `json:"anime_slug"`
	Status    string   `json:"status"`
	Tipe      string   `json:"tipe"`
	Skor      string   `json:"skor"`
	Penonton  string   `json:"penonton"`
	Sinopsis  string   `json:"sinopsis"`
	Genre     []string `json:"genre"`
	Cover     string   `json:"cover"`
}

// ValidateResponse validates API response based on endpoint
func ValidateResponse(endpoint string, data []byte) error {
	// First, validate basic structure and confidence score
	var baseResp BaseResponse
	if err := json.Unmarshal(data, &baseResp); err != nil {
		return fmt.Errorf("invalid JSON structure: %v", err)
	}

	// Check confidence score
	if baseResp.ConfidenceScore < 0.5 {
		return fmt.Errorf("confidence score too low: %f", baseResp.ConfidenceScore)
	}

	// Validate specific endpoint structure
	switch endpoint {
	case "/api/v1/home":
		return validateHomeResponse(data)
	case "/api/v1/jadwal-rilis":
		return validateJadwalRilisResponse(data)
	case "/api/v1/anime-terbaru":
		return validateAnimeTerbaruResponse(data)
	case "/api/v1/movie":
		return validateMovieResponse(data)
	case "/api/v1/anime-detail":
		return validateAnimeDetailResponse(data)
	case "/api/v1/episode-detail":
		return validateEpisodeDetailResponse(data)
	case "/api/v1/search":
		return validateSearchResponse(data)
	default:
		if strings.HasPrefix(endpoint, "/api/v1/jadwal-rilis/") {
			return validateJadwalRilisDayResponse(data)
		}
		return fmt.Errorf("unknown endpoint: %s", endpoint)
	}
}

func validateHomeResponse(data []byte) error {
	var resp HomeResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return fmt.Errorf("invalid home response structure: %v", err)
	}

	// Check if response has required structure by unmarshaling to map first
	var respMap map[string]interface{}
	if err := json.Unmarshal(data, &respMap); err != nil {
		return fmt.Errorf("invalid response structure: %v", err)
	}

	// Check required top-level fields
	requiredFields := []string{"top10", "new_eps", "movies", "jadwal_rilis"}
	for _, field := range requiredFields {
		if _, exists := respMap[field]; !exists {
			return fmt.Errorf("required field '%s' is missing", field)
		}
	}

	// Validate required fields in arrays
	for _, item := range resp.Top10 {
		if err := validateRequiredFields(item, []string{"judul", "url", "anime_slug", "cover"}); err != nil {
			return fmt.Errorf("invalid top10 item: %v", err)
		}
	}

	for _, item := range resp.NewEps {
		if err := validateRequiredFields(item, []string{"judul", "url", "anime_slug", "cover"}); err != nil {
			return fmt.Errorf("invalid new_eps item: %v", err)
		}
	}

	for _, item := range resp.Movies {
		if err := validateRequiredFields(item, []string{"judul", "url", "anime_slug", "cover"}); err != nil {
			return fmt.Errorf("invalid movies item: %v", err)
		}
	}

	return nil
}

func validateJadwalRilisResponse(data []byte) error {
	var resp JadwalRilisResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return fmt.Errorf("invalid jadwal-rilis response structure: %v", err)
	}

	for _, dayItems := range resp.Data {
		for _, item := range dayItems {
			if err := validateRequiredFields(item, []string{"title", "url", "anime_slug", "cover_url"}); err != nil {
				return fmt.Errorf("invalid jadwal-rilis item: %v", err)
			}
		}
	}

	return nil
}

func validateJadwalRilisDayResponse(data []byte) error {
	var resp JadwalRilisDayResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return fmt.Errorf("invalid jadwal-rilis day response structure: %v", err)
	}

	for _, item := range resp.Data {
		if err := validateRequiredFields(item, []string{"title", "url", "anime_slug", "cover_url"}); err != nil {
			return fmt.Errorf("invalid jadwal-rilis day item: %v", err)
		}
	}

	return nil
}

func validateAnimeTerbaruResponse(data []byte) error {
	var resp AnimeTerbaruResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return fmt.Errorf("invalid anime-terbaru response structure: %v", err)
	}

	for _, item := range resp.Data {
		if err := validateRequiredFields(item, []string{"judul", "url", "anime_slug", "cover"}); err != nil {
			return fmt.Errorf("invalid anime-terbaru item: %v", err)
		}
	}

	return nil
}

func validateMovieResponse(data []byte) error {
	var resp MovieResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return fmt.Errorf("invalid movie response structure: %v", err)
	}

	for _, item := range resp.Data {
		if err := validateRequiredFields(item, []string{"judul", "url", "anime_slug", "cover"}); err != nil {
			return fmt.Errorf("invalid movie item: %v", err)
		}
	}

	return nil
}

func validateAnimeDetailResponse(data []byte) error {
	var resp AnimeDetailResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return fmt.Errorf("invalid anime-detail response structure: %v", err)
	}

	if err := validateRequiredFields(resp, []string{"judul", "url", "anime_slug", "cover"}); err != nil {
		return fmt.Errorf("invalid anime-detail: %v", err)
	}

	return nil
}

func validateEpisodeDetailResponse(data []byte) error {
	var resp EpisodeDetailResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return fmt.Errorf("invalid episode-detail response structure: %v", err)
	}

	if err := validateRequiredFields(resp, []string{"title", "thumbnail_url"}); err != nil {
		return fmt.Errorf("invalid episode-detail: %v", err)
	}

	// Validate streaming servers
	for _, server := range resp.StreamingServers {
		if err := validateRequiredFields(server, []string{"server_name", "streaming_url"}); err != nil {
			return fmt.Errorf("invalid streaming server: %v", err)
		}
	}

	return nil
}

func validateSearchResponse(data []byte) error {
	var resp SearchResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return fmt.Errorf("invalid search response structure: %v", err)
	}

	for _, item := range resp.Data {
		if err := validateRequiredFields(item, []string{"judul", "url", "anime_slug", "cover"}); err != nil {
			return fmt.Errorf("invalid search item: %v", err)
		}
	}

	return nil
}

// validateRequiredFields checks if required fields are present and not empty
func validateRequiredFields(item interface{}, requiredFields []string) error {
	// Handle map[string]interface{} case
	if itemMap, ok := item.(map[string]interface{}); ok {
		for _, fieldName := range requiredFields {
			value, exists := itemMap[fieldName]
			if !exists {
				return fmt.Errorf("required field '%s' not found", fieldName)
			}

			// Check if field is empty
			switch v := value.(type) {
			case string:
				if v == "" || isPlaceholderValue(v) {
					return fmt.Errorf("required field '%s' is empty or placeholder", fieldName)
				}
				// Validate URL fields
				if strings.Contains(strings.ToLower(fieldName), "url") {
					if !isValidURL(v) {
						return fmt.Errorf("field '%s' contains invalid URL: %s", fieldName, v)
					}
				}
			case nil:
				return fmt.Errorf("required field '%s' is null", fieldName)
			}
		}
		return nil
	}

	// Handle struct case
	v := reflect.ValueOf(item)
	t := reflect.TypeOf(item)

	for _, fieldName := range requiredFields {
		var field reflect.StructField
		var found bool

		// First try to find by JSON tag
		for i := 0; i < t.NumField(); i++ {
			structField := t.Field(i)
			jsonTag := structField.Tag.Get("json")
			if jsonTag != "" {
				// Remove omitempty and other options
				jsonName := strings.Split(jsonTag, ",")[0]
				if jsonName == fieldName {
					field = structField
					found = true
					break
				}
			}
		}

		// If not found by JSON tag, try by field name
		if !found {
			field, found = t.FieldByName(strings.Title(fieldName))
			if !found {
				// Try with exact case
				for i := 0; i < t.NumField(); i++ {
					if strings.ToLower(t.Field(i).Name) == strings.ToLower(fieldName) {
						field = t.Field(i)
						found = true
						break
					}
				}
			}
		}

		if !found {
			return fmt.Errorf("required field '%s' not found", fieldName)
		}

		fieldValue := v.FieldByName(field.Name)
		if !fieldValue.IsValid() {
			return fmt.Errorf("required field '%s' is invalid", fieldName)
		}

		// Check if field is empty
		switch fieldValue.Kind() {
		case reflect.String:
			if fieldValue.String() == "" || isPlaceholderValue(fieldValue.String()) {
				return fmt.Errorf("required field '%s' is empty or placeholder", fieldName)
			}
			// Validate URL fields
			if strings.Contains(strings.ToLower(fieldName), "url") {
				if !isValidURL(fieldValue.String()) {
					return fmt.Errorf("field '%s' contains invalid URL: %s", fieldName, fieldValue.String())
				}
			}
		case reflect.Slice:
			if fieldValue.Len() == 0 {
				return fmt.Errorf("required field '%s' is empty array", fieldName)
			}
		}
	}

	return nil
}

// isPlaceholderValue checks if a string value is a common error placeholder
func isPlaceholderValue(value string) bool {
	placeholders := []string{
		"error", "null", "undefined", "n/a", "not found", "404", "500",
		"failed", "timeout", "unavailable", "maintenance", "coming soon",
	}

	lowerValue := strings.ToLower(strings.TrimSpace(value))
	for _, placeholder := range placeholders {
		if strings.Contains(lowerValue, placeholder) {
			return true
		}
	}

	return false
}

// isValidURL checks if a string is a valid URL
func isValidURL(str string) bool {
	if str == "" {
		return false
	}

	u, err := url.Parse(str)
	if err != nil {
		return false
	}

	// Must have scheme and host
	return u.Scheme != "" && u.Host != ""
}
