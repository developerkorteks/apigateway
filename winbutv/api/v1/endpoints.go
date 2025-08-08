package v1

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/nabilulilalbab/winbu.tv/config"
	"github.com/nabilulilalbab/winbu.tv/models"
	"github.com/nabilulilalbab/winbu.tv/scrapers"
)

type APIHandler struct {
	config          *config.Config
	homeScraper     *scrapers.HomeScraper
	animeScraper    *scrapers.AnimeScraper
	movieScraper    *scrapers.MovieScraper
	scheduleScraper *scrapers.ScheduleScraper
	searchScraper   *scrapers.SearchScraper
	detailScraper   *scrapers.DetailScraper
}

func NewAPIHandler(cfg *config.Config) *APIHandler {
	return &APIHandler{
		config:          cfg,
		homeScraper:     scrapers.NewHomeScraper(cfg),
		animeScraper:    scrapers.NewAnimeScraper(cfg),
		movieScraper:    scrapers.NewMovieScraper(cfg),
		scheduleScraper: scrapers.NewScheduleScraper(cfg),
		searchScraper:   scrapers.NewSearchScraper(cfg),
		detailScraper:   scrapers.NewDetailScraper(cfg),
	}
}

func SetupRoutes(r *gin.RouterGroup) {
	cfg := config.Load()
	handler := NewAPIHandler(cfg)

	r.GET("/home", handler.GetHome)
	r.GET("/anime-terbaru", handler.GetAnimeTerbaru)
	r.GET("/movie", handler.GetMovies)
	r.GET("/jadwal-rilis", handler.GetSchedule)
	r.GET("/jadwal-rilis/:day", handler.GetScheduleByDay)
	r.GET("/search", handler.GetSearch)
	r.GET("/anime-detail", handler.GetAnimeDetail)
	r.GET("/episode-detail", handler.GetEpisodeDetail)
}

// GetHome handles GET /api/v1/home
// @Summary Get homepage data
// @Description Mengambil data homepage termasuk top 10 anime, episode terbaru, film terbaru, dan jadwal rilis
// @Tags Homepage
// @Accept json
// @Produce json
// @Success 200 {object} models.HomeResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/v1/home [get]
func (h *APIHandler) GetHome(c *gin.Context) {
	data, err := h.homeScraper.ScrapeHome()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:           true,
			Message:         "Failed to scrape home data: " + err.Error(),
			ConfidenceScore: 0.0,
		})
		return
	}

	c.JSON(http.StatusOK, data)
}

// GetAnimeTerbaru handles GET /api/v1/anime-terbaru?page=<int>
// @Summary Get anime terbaru
// @Description Mengambil daftar anime terbaru dengan pagination
// @Tags Anime
// @Accept json
// @Produce json
// @Param page query int false "Nomor halaman" default(1)
// @Success 200 {object} models.AnimeTerbaruResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/v1/anime-terbaru [get]
func (h *APIHandler) GetAnimeTerbaru(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	data, err := h.animeScraper.ScrapeAnimeTerbaru(page)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:           true,
			Message:         "Failed to scrape anime terbaru data: " + err.Error(),
			ConfidenceScore: 0.0,
		})
		return
	}

	c.JSON(http.StatusOK, data)
}

// GetMovies handles GET /api/v1/movie?page=<int>
// @Summary Get movies
// @Description Mengambil daftar film dengan pagination
// @Tags Movies
// @Accept json
// @Produce json
// @Param page query int false "Nomor halaman" default(1)
// @Success 200 {object} models.MovieResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/v1/movie [get]
func (h *APIHandler) GetMovies(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	data, err := h.movieScraper.ScrapeMovies(page)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:           true,
			Message:         "Failed to scrape movie data: " + err.Error(),
			ConfidenceScore: 0.0,
		})
		return
	}

	c.JSON(http.StatusOK, data)
}

// GetSchedule handles GET /api/v1/jadwal-rilis
// @Summary Get jadwal rilis
// @Description Mengambil jadwal rilis anime per hari
// @Tags Schedule
// @Accept json
// @Produce json
// @Success 200 {object} models.ScheduleResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/v1/jadwal-rilis [get]
func (h *APIHandler) GetSchedule(c *gin.Context) {
	data, err := h.scheduleScraper.ScrapeSchedule()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:           true,
			Message:         "Failed to scrape schedule data: " + err.Error(),
			ConfidenceScore: 0.0,
		})
		return
	}

	c.JSON(http.StatusOK, data)
}

// GetSearch handles GET /api/v1/search?q=<string>&page=<int>
// @Summary Search anime
// @Description Mencari anime berdasarkan judul
// @Tags Search
// @Accept json
// @Produce json
// @Param query query string true "Query pencarian"
// @Success 200 {object} models.SearchResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/v1/search [get]
func (h *APIHandler) GetSearch(c *gin.Context) {
	query := c.Query("query")
	if query == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:           true,
			Message:         "Query parameter 'query' is required",
			ConfidenceScore: 0.0,
		})
		return
	}

	data, err := h.searchScraper.SearchAnime(query, 1)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:           true,
			Message:         "Failed to search data: " + err.Error(),
			ConfidenceScore: 0.0,
		})
		return
	}

	c.JSON(http.StatusOK, data)
}

// GetAnimeDetail handles GET /api/v1/anime-detail?anime_slug=<string>
// @Summary Get anime/movie/series detail
// @Description Mengambil detail anime, film, atau series termasuk episode, sinopsis, dan rekomendasi. Slug dapat berupa 'nama-anime', 'film/nama-film', atau 'series/nama-series'
// @Tags Detail
// @Accept json
// @Produce json
// @Param anime_slug query string true "Anime/Movie/Series slug (contoh: 'kobane-2022', 'film/kobane-2022', 'series/legend-of-the-female-general')"
// @Success 200 {object} models.AnimeDetailResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/v1/anime-detail [get]
func (h *APIHandler) GetAnimeDetail(c *gin.Context) {
	animeSlug := c.Query("anime_slug")
	if animeSlug == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:           true,
			Message:         "Query parameter 'anime_slug' is required",
			ConfidenceScore: 0.0,
		})
		return
	}

	data, err := h.detailScraper.ScrapeAnimeDetail(animeSlug)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:           true,
			Message:         "Failed to scrape anime detail: " + err.Error(),
			ConfidenceScore: 0.0,
		})
		return
	}

	c.JSON(http.StatusOK, data)
}

// GetEpisodeDetail handles GET /api/v1/episode-detail?episode_url=<string>
// @Summary Get episode detail
// @Description Mengambil detail episode termasuk server streaming dan link download
// @Tags Detail
// @Accept json
// @Produce json
// @Param episode_url query string true "URL episode"
// @Success 200 {object} models.EpisodeDetailResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/v1/episode-detail [get]
func (h *APIHandler) GetEpisodeDetail(c *gin.Context) {
	episodeURL := c.Query("episode_url")
	if episodeURL == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:           true,
			Message:         "Query parameter 'episode_url' is required",
			ConfidenceScore: 0.0,
		})
		return
	}

	data, err := h.detailScraper.ScrapeEpisodeDetail(episodeURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:           true,
			Message:         "Failed to scrape episode detail: " + err.Error(),
			ConfidenceScore: 0.0,
		})
		return
	}

	c.JSON(http.StatusOK, data)
}

// GetScheduleByDay handles GET /api/v1/jadwal-rilis/:day
// @Summary Get jadwal rilis by day
// @Description Mengambil jadwal rilis anime untuk hari tertentu
// @Tags Schedule
// @Accept json
// @Produce json
// @Param day path string true "Nama hari (monday, tuesday, wednesday, thursday, friday, saturday, sunday)"
// @Success 200 {object} models.DayScheduleResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/v1/jadwal-rilis/{day} [get]
func (h *APIHandler) GetScheduleByDay(c *gin.Context) {
	day := c.Param("day")
	if day == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:           true,
			Message:         "Day parameter is required",
			ConfidenceScore: 0.0,
		})
		return
	}

	data, err := h.scheduleScraper.ScrapeScheduleByDay(day)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:           true,
			Message:         "Failed to get schedule for day: " + err.Error(),
			ConfidenceScore: 0.0,
		})
		return
	}

	c.JSON(http.StatusOK, data)
}
