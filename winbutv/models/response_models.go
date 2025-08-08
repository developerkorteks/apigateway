package models

// Base response structure with confidence score
type BaseResponse struct {
	ConfidenceScore float64 `json:"confidence_score"`
	Message         string  `json:"message,omitempty"`
	Source          string  `json:"source,omitempty"`
}

type ErrorResponse struct {
	Error           bool    `json:"error"`
	Message         string  `json:"message"`
	ConfidenceScore float64 `json:"confidence_score"`
}

// Home page response models
type Top10Item struct {
	Judul     string   `json:"judul"`
	URL       string   `json:"url"`
	AnimeSlug string   `json:"anime_slug"`
	Rating    string   `json:"rating"`
	Cover     string   `json:"cover"`
	Genres    []string `json:"genres"`
}

type NewEpisodeItem struct {
	Judul     string `json:"judul"`
	URL       string `json:"url"`
	AnimeSlug string `json:"anime_slug"`
	Episode   string `json:"episode"`
	Rilis     string `json:"rilis"`
	Cover     string `json:"cover"`
}

type MovieItem struct {
	Judul     string   `json:"judul"`
	URL       string   `json:"url"`
	AnimeSlug string   `json:"anime_slug"`
	Tanggal   string   `json:"tanggal"`
	Cover     string   `json:"cover"`
	Genres    []string `json:"genres"`
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

type ScheduleData struct {
	Monday    []ScheduleItem `json:"Monday"`
	Tuesday   []ScheduleItem `json:"Tuesday"`
	Wednesday []ScheduleItem `json:"Wednesday"`
	Thursday  []ScheduleItem `json:"Thursday"`
	Friday    []ScheduleItem `json:"Friday"`
	Saturday  []ScheduleItem `json:"Saturday"`
	Sunday    []ScheduleItem `json:"Sunday"`
}

type HomeResponse struct {
	BaseResponse
	Top10       []Top10Item      `json:"top10"`
	NewEps      []NewEpisodeItem `json:"new_eps"`
	Movies      []MovieItem      `json:"movies"`
	JadwalRilis ScheduleData     `json:"jadwal_rilis"`
}

// Schedule response for /api/v1/jadwal-rilis endpoint
type ScheduleResponse struct {
	BaseResponse
	Data ScheduleData `json:"data"`
}

// Anime terbaru response models
type AnimeTerbaruItem struct {
	Judul     string `json:"judul"`
	URL       string `json:"url"`
	AnimeSlug string `json:"anime_slug"`
	Episode   string `json:"episode"`
	Uploader  string `json:"uploader"`
	Rilis     string `json:"rilis"`
	Cover     string `json:"cover"`
}

type AnimeTerbaruResponse struct {
	BaseResponse
	Data []AnimeTerbaruItem `json:"data"`
}

// Movie response models
type MovieDetailItem struct {
	Judul     string   `json:"judul"`
	URL       string   `json:"url"`
	AnimeSlug string   `json:"anime_slug"`
	Status    string   `json:"status"`
	Skor      string   `json:"skor"`
	Sinopsis  string   `json:"sinopsis"`
	Views     string   `json:"views"`
	Cover     string   `json:"cover"`
	Genres    []string `json:"genres"`
	Tanggal   string   `json:"tanggal"`
}

type MovieResponse struct {
	BaseResponse
	Data []MovieDetailItem `json:"data"`
}

// Single day schedule response
type DayScheduleResponse struct {
	BaseResponse
	Data []ScheduleItem `json:"data"`
}

// Search response models
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

type SearchResponse struct {
	BaseResponse
	Data []SearchResultItem `json:"data"`
}

// AnimeDetailResponse represents the response for anime detail endpoint
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
	Rating          AnimeRating          `json:"rating"`
}

// EpisodeListItem represents an episode in the anime detail
type EpisodeListItem struct {
	Episode     string `json:"episode"`
	Title       string `json:"title"`
	URL         string `json:"url"`
	EpisodeSlug string `json:"episode_slug"`
	ReleaseDate string `json:"release_date"`
}

// RecommendationItem represents a recommended anime
type RecommendationItem struct {
	Title     string `json:"title"`
	URL       string `json:"url"`
	AnimeSlug string `json:"anime_slug"`
	CoverURL  string `json:"cover_url"`
	Rating    string `json:"rating"`
	Episode   string `json:"episode"`
}

// AnimeDetails represents detailed information about an anime
type AnimeDetails struct {
	Japanese     string `json:"Japanese"`
	English      string `json:"English"`
	Status       string `json:"Status"`
	Type         string `json:"Type"`
	Source       string `json:"Source"`
	Duration     string `json:"Duration"`
	TotalEpisode string `json:"Total Episode"`
	Season       string `json:"Season"`
	Studio       string `json:"Studio"`
	Producers    string `json:"Producers"`
	Released     string `json:"Released:"`
}

// AnimeRating represents rating information
type AnimeRating struct {
	Score string `json:"score"`
	Users string `json:"users"`
}

// EpisodeDetailResponse represents the response for episode detail endpoint
type EpisodeDetailResponse struct {
	BaseResponse
	Title            string             `json:"title"`
	ThumbnailURL     string             `json:"thumbnail_url"`
	StreamingServers []StreamingServer  `json:"streaming_servers"`
	ReleaseInfo      string             `json:"release_info"`
	DownloadLinks    DownloadLinksGroup `json:"download_links"`
	Navigation       EpisodeNavigation  `json:"navigation"`
	AnimeInfo        AnimeInfo          `json:"anime_info"`
	OtherEpisodes    []OtherEpisode     `json:"other_episodes"`
}

// EpisodeNavigation represents navigation between episodes
type EpisodeNavigation struct {
	PreviousEpisodeURL string `json:"previous_episode_url,omitempty"`
	AllEpisodesURL     string `json:"all_episodes_url,omitempty"`
	NextEpisodeURL     string `json:"next_episode_url,omitempty"`
}

// SeriesInfo represents information about the series
type SeriesInfo struct {
	PosterImageURL string   `json:"poster_image_url"`
	Rating         string   `json:"rating"`
	Genres         []string `json:"genres"`
	Synopsis       string   `json:"synopsis"`
}

// StreamQualityGroup represents streaming options grouped by quality
type StreamQualityGroup struct {
	Quality string         `json:"quality"`
	Servers []StreamServer `json:"servers"`
}

// StreamServer represents a streaming server
type StreamServer struct {
	Name      string `json:"name"`
	StreamURL string `json:"stream_url,omitempty"`
}

// DownloadQualityGroup represents download options grouped by quality
type DownloadQualityGroup struct {
	Quality       string         `json:"quality"`
	DownloadLinks []DownloadLink `json:"download_links"`
}

// DownloadLink represents a download link
type DownloadLink struct {
	Provider string `json:"provider"`
	URL      string `json:"url"`
}

// StreamingServer represents a streaming server for episode detail
type StreamingServer struct {
	ServerName   string `json:"server_name"`
	StreamingURL string `json:"streaming_url"`
}

// DownloadLinksGroup represents all download links grouped by format and quality
type DownloadLinksGroup struct {
	MKV  map[string][]DownloadLink `json:"MKV"`
	MP4  map[string][]DownloadLink `json:"MP4"`
	X265 map[string][]DownloadLink `json:"x265 [Mode Irit Kuota tapi Kualitas Sama Beningnya]"`
}

// AnimeInfo represents information about the anime series
type AnimeInfo struct {
	Title        string   `json:"title"`
	ThumbnailURL string   `json:"thumbnail_url"`
	Synopsis     string   `json:"synopsis"`
	Genres       []string `json:"genres"`
}

// OtherEpisode represents other episodes from the same series
type OtherEpisode struct {
	Title        string `json:"title"`
	URL          string `json:"url"`
	ThumbnailURL string `json:"thumbnail_url"`
	ReleaseDate  string `json:"release_date"`
}
