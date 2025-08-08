package claude

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/patrickmn/go-cache"
)

// ===================================================================================
// 1. STRUCTS & CONSTANTS
// ===================================================================================

const (
	// Cache settings
	DefaultCacheExpiration = 15 * time.Minute
	CleanupInterval        = 30 * time.Minute

	// Rate limiting - Optimized for faster scraping
	MaxConcurrentRequests = 100
	RequestTimeout        = 20 * time.Second

	// Connection pool settings - Increased for better performance
	MaxIdleConns        = 200
	MaxIdleConnsPerHost = 50
	IdleConnTimeout     = 120 * time.Second
)

type StreamServer struct {
	Name      string `json:"name"`
	StreamURL string `json:"streamUrl,omitempty"`
	PostID    string `json:"-"`
	Nume      string `json:"-"`
	DataType  string `json:"-"`
}

type StreamQualityGroup struct {
	Quality string         `json:"quality"`
	Servers []StreamServer `json:"servers"`
}

type DownloadLink struct {
	Provider string `json:"provider"`
	URL      string `json:"url"`
}

type DownloadQualityGroup struct {
	Quality       string         `json:"quality"`
	DownloadLinks []DownloadLink `json:"downloadLinks"`
}

type EpisodeNavigation struct {
	PreviousEpisodeURL string `json:"previousEpisodeUrl,omitempty"`
	AllEpisodesURL     string `json:"allEpisodesUrl,omitempty"`
	NextEpisodeURL     string `json:"nextEpisodeUrl,omitempty"`
}

type SeriesInfo struct {
	PosterImageURL string   `json:"posterImageUrl"`
	Rating         string   `json:"rating"`
	Genres         []string `json:"genres"`
	Synopsis       string   `json:"synopsis"`
}

type RecommendationItem struct {
	Title    string `json:"title"`
	URL      string `json:"url"`
	ImageURL string `json:"imageUrl"`
	Rating   string `json:"rating"`
}

type CompleteEpisodePage struct {
	URL             string                 `json:"url"`
	EpisodeTitle    string                 `json:"episodeTitle"`
	SeriesTitle     string                 `json:"seriesTitle"`
	EpisodeNav      EpisodeNavigation      `json:"episodeNav"`
	SeriesInfo      SeriesInfo             `json:"seriesInfo"`
	StreamGroups    []StreamQualityGroup   `json:"streamGroups"`
	DownloadGroups  []DownloadQualityGroup `json:"downloadGroups"`
	Recommendations []RecommendationItem   `json:"recommendations"`
	CachedAt        time.Time              `json:"cachedAt"`
}

// ===================================================================================
// 2. SCRAPER SERVICE WITH CACHING & RATE LIMITING
// ===================================================================================

type ScraperService struct {
	cache      *cache.Cache
	httpClient *http.Client
	semaphore  chan struct{}
	collectors sync.Pool
	regex      struct {
		whitespace *regexp.Regexp
		rating     *regexp.Regexp
		srcURL     *regexp.Regexp
	}
}

func NewScraperService() *ScraperService {
	// Optimized HTTP client with connection pooling
	transport := &http.Transport{
		MaxIdleConns:        MaxIdleConns,
		MaxIdleConnsPerHost: MaxIdleConnsPerHost,
		IdleConnTimeout:     IdleConnTimeout,
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 20 * time.Second,
	}

	s := &ScraperService{
		cache: cache.New(DefaultCacheExpiration, CleanupInterval),
		httpClient: &http.Client{
			Transport: transport,
			Timeout:   RequestTimeout,
		},
		semaphore: make(chan struct{}, MaxConcurrentRequests),
	}

	// Pre-compile regex patterns
	s.regex.whitespace = regexp.MustCompile(`\s+`)
	s.regex.rating = regexp.MustCompile(`(\d+\.\d+)`)
	s.regex.srcURL = regexp.MustCompile(`src='([^']*)'|src="([^"]*)"`)

	// Initialize collector pool
	s.collectors.New = func() interface{} {
		return s.createCollector()
	}

	return s
}

func (s *ScraperService) createCollector() *colly.Collector {
	c := colly.NewCollector(
		colly.AllowedDomains("winbu.tv"),
		colly.Async(true),
	)

	c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: 20,                    // Increased parallelism
		Delay:       50 * time.Millisecond, // Reduced delay
	})

	c.UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/115.0.0.0 Safari/537.36"

	// Set timeouts
	c.SetRequestTimeout(RequestTimeout)

	return c
}

func (s *ScraperService) cleanText(text string) string {
	return strings.TrimSpace(s.regex.whitespace.ReplaceAllString(text, " "))
}

func (s *ScraperService) generateCacheKey(url string) string {
	hash := md5.Sum([]byte(url))
	return hex.EncodeToString(hash[:])
}

// ===================================================================================
// 3. MAIN SCRAPING METHOD WITH CACHING
// ===================================================================================

func (s *ScraperService) ScrapeEpisodePage(ctx context.Context, episodeURL string) (*CompleteEpisodePage, error) {
	// Check cache first
	cacheKey := s.generateCacheKey(episodeURL)
	if cached, found := s.cache.Get(cacheKey); found {
		if result, ok := cached.(*CompleteEpisodePage); ok {
			log.Printf("Cache hit for URL: %s", episodeURL)
			return result, nil
		}
	}

	// Rate limiting
	select {
	case s.semaphore <- struct{}{}:
		defer func() { <-s.semaphore }()
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	log.Printf("Scraping fresh data for URL: %s", episodeURL)

	detail := &CompleteEpisodePage{
		URL:      episodeURL,
		CachedAt: time.Now(),
	}

	// Get collector from pool
	c := s.collectors.Get().(*colly.Collector)
	defer s.collectors.Put(c)

	// Error handling
	var scrapeError error
	c.OnError(func(r *colly.Response, err error) {
		scrapeError = fmt.Errorf("scraping error: %v", err)
	})

	// Setup all selectors
	s.setupSelectors(c, detail)

	// Execute scraping with timeout
	done := make(chan error, 1)
	go func() {
		done <- c.Visit(episodeURL)
	}()

	select {
	case err := <-done:
		if err != nil {
			return nil, fmt.Errorf("failed to visit URL: %v", err)
		}
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	c.Wait()

	if scrapeError != nil {
		return nil, scrapeError
	}

	// Fetch stream URLs concurrently
	if err := s.fetchStreamURLs(ctx, detail); err != nil {
		log.Printf("Warning: Failed to fetch some stream URLs: %v", err)
	}

	// Cache the result
	s.cache.Set(cacheKey, detail, DefaultCacheExpiration)

	return detail, nil
}

// ===================================================================================
// 4. SELECTOR SETUP
// ===================================================================================

func (s *ScraperService) setupSelectors(c *colly.Collector, detail *CompleteEpisodePage) {
	// Episode title - fix selector
	c.OnHTML("div.list-title h2", func(e *colly.HTMLElement) {
		text := s.cleanText(e.Text)
		// Skip if it's just "Genres" or other metadata
		if text != "" && text != "Genres" && text != "Rating" {
			detail.EpisodeTitle = text
		}
	})

	// Alternative episode title selectors
	c.OnHTML("h1.entry-title", func(e *colly.HTMLElement) {
		if detail.EpisodeTitle == "" {
			detail.EpisodeTitle = s.cleanText(e.Text)
		}
	})

	c.OnHTML(".post-title h1", func(e *colly.HTMLElement) {
		if detail.EpisodeTitle == "" {
			detail.EpisodeTitle = s.cleanText(e.Text)
		}
	})

	// Series info
	c.OnHTML("div.m-info div.movies-list-full div.t-item", func(e *colly.HTMLElement) {
		detail.SeriesTitle = s.cleanText(e.ChildText(".mli-info .judul"))
		detail.SeriesInfo.PosterImageURL = e.ChildAttr(".mli-thumb-box img", "src")

		e.ForEach(".mli-mvi", func(_ int, el *colly.HTMLElement) {
			text := s.cleanText(el.Text)
			if strings.HasPrefix(text, "Rating") {
				if matches := s.regex.rating.FindStringSubmatch(text); len(matches) > 1 {
					detail.SeriesInfo.Rating = matches[1]
				}
			} else if strings.HasPrefix(text, "Genre") {
				el.ForEach("a", func(_ int, genreEl *colly.HTMLElement) {
					detail.SeriesInfo.Genres = append(detail.SeriesInfo.Genres, s.cleanText(genreEl.Text))
				})
			}
		})
		detail.SeriesInfo.Synopsis = s.cleanText(e.ChildText(".mli-desc"))
	})

	// Episode navigation
	c.OnHTML("div.naveps", func(e *colly.HTMLElement) {
		detail.EpisodeNav.PreviousEpisodeURL = e.ChildAttr("div.nvs a", "href")
		detail.EpisodeNav.AllEpisodesURL = e.ChildAttr("div.nvs.nvsc a", "href")
		detail.EpisodeNav.NextEpisodeURL = e.ChildAttr("div.nvs.rght a", "href")
	})

	// Streaming servers
	c.OnHTML("div.player-modes div.dropdown", func(e *colly.HTMLElement) {
		quality := s.cleanText(e.ChildText("button.dropdown-toggle"))
		var servers []StreamServer
		e.ForEach(".dropdown-item .east_player_option", func(_ int, el *colly.HTMLElement) {
			servers = append(servers, StreamServer{
				Name:     s.cleanText(el.ChildText("span")),
				PostID:   el.Attr("data-post"),
				Nume:     el.Attr("data-nume"),
				DataType: el.Attr("data-type"),
			})
		})
		if quality != "" {
			detail.StreamGroups = append(detail.StreamGroups, StreamQualityGroup{
				Quality: quality,
				Servers: servers,
			})
		}
	})

	// Download links
	c.OnHTML("div.download-eps ul li", func(e *colly.HTMLElement) {
		qualityGroup := DownloadQualityGroup{
			Quality: s.cleanText(e.ChildText("strong")),
		}
		e.ForEach("span a", func(_ int, el *colly.HTMLElement) {
			qualityGroup.DownloadLinks = append(qualityGroup.DownloadLinks, DownloadLink{
				Provider: s.cleanText(el.Text),
				URL:      el.Attr("href"),
			})
		})
		if qualityGroup.Quality != "" {
			detail.DownloadGroups = append(detail.DownloadGroups, qualityGroup)
		}
	})

	// Recommendations
	c.OnHTML("div.rekom .ml-item-rekom", func(e *colly.HTMLElement) {
		rec := RecommendationItem{
			Title:    s.cleanText(e.ChildText(".judul")),
			URL:      e.ChildAttr("a.ml-mask", "href"),
			ImageURL: e.ChildAttr("img.mli-thumb", "src"),
			Rating:   s.cleanText(e.ChildText(".mli-mvi")),
		}
		if rec.Title != "" {
			detail.Recommendations = append(detail.Recommendations, rec)
		}
	})
}

// ===================================================================================
// 5. OPTIMIZED STREAM URL FETCHING
// ===================================================================================

func (s *ScraperService) fetchStreamURLs(ctx context.Context, detail *CompleteEpisodePage) error {
	var wg sync.WaitGroup
	errChan := make(chan error, 10) // Buffer for errors

	streamSemaphore := make(chan struct{}, 20) // Limit concurrent stream requests

	for i := range detail.StreamGroups {
		for j := range detail.StreamGroups[i].Servers {
			wg.Add(1)
			go func(i, j int) {
				defer wg.Done()

				select {
				case streamSemaphore <- struct{}{}:
					defer func() { <-streamSemaphore }()
				case <-ctx.Done():
					return
				}

				server := &detail.StreamGroups[i].Servers[j]
				streamURL, err := s.getStreamURL(ctx, server.PostID, server.Nume, server.DataType)
				if err != nil {
					select {
					case errChan <- fmt.Errorf("failed to get stream for %s: %v", server.Name, err):
					default:
					}
					server.StreamURL = "Unavailable"
				} else {
					server.StreamURL = streamURL
				}
			}(i, j)
		}
	}

	// Wait for completion
	go func() {
		wg.Wait()
		close(errChan)
	}()

	// Collect errors
	var errors []error
	for err := range errChan {
		errors = append(errors, err)
		if len(errors) >= 5 { // Stop collecting after 5 errors
			break
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("encountered %d errors during stream URL fetching", len(errors))
	}

	return nil
}

func (s *ScraperService) getStreamURL(ctx context.Context, postID, nume, dataType string) (string, error) {
	ajaxURL := "https://winbu.tv/wp-admin/admin-ajax.php"
	formData := url.Values{
		"action": {"player_ajax"},
		"post":   {postID},
		"nume":   {nume},
		"type":   {dataType},
	}

	req, err := http.NewRequestWithContext(ctx, "POST", ajaxURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/115.0.0.0 Safari/537.36")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("Referer", "https://winbu.tv/")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	htmlResponse := string(body)
	matches := s.regex.srcURL.FindStringSubmatch(htmlResponse)

	if len(matches) < 2 {
		return "", fmt.Errorf("no src URL found in response")
	}

	streamURL := matches[1]
	if streamURL == "" {
		streamURL = matches[2]
	}

	return streamURL, nil
}

// ===================================================================================
// 6. CACHE MANAGEMENT METHODS
// ===================================================================================

func (s *ScraperService) ClearCache() {
	s.cache.Flush()
}

func (s *ScraperService) GetCacheStats() (int, map[string]cache.Item) {
	return s.cache.ItemCount(), s.cache.Items()
}

func (s *ScraperService) WarmupCache(urls []string) error {
	ctx := context.Background()
	var wg sync.WaitGroup

	for _, url := range urls {
		wg.Add(1)
		go func(u string) {
			defer wg.Done()
			_, err := s.ScrapeEpisodePage(ctx, u)
			if err != nil {
				log.Printf("Failed to warmup cache for %s: %v", u, err)
			}
		}(url)
	}

	wg.Wait()
	return nil
}

// ===================================================================================
// 7. HTTP HANDLER FOR API
// ===================================================================================

func (s *ScraperService) HandleEpisodeAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	episodeURL := r.URL.Query().Get("url")
	if episodeURL == "" {
		http.Error(w, "Missing 'url' parameter", http.StatusBadRequest)
		return
	}

	// Validate URL format
	if !strings.HasPrefix(episodeURL, "https://winbu.tv/") {
		http.Error(w, "Invalid URL domain", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()

	result, err := s.ScrapeEpisodePage(ctx, episodeURL)
	if err != nil {
		log.Printf("Scraping failed for %s: %v", episodeURL, err)
		http.Error(w, "Failed to scrape episode data", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "public, max-age=900") // 15 minutes browser cache

	if err := json.NewEncoder(w).Encode(result); err != nil {
		log.Printf("JSON encoding failed: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// ===================================================================================
// 8. USAGE EXAMPLE
// ===================================================================================

func RunAPIServer() {
	scraper := NewScraperService()

	// Setup HTTP server
	http.HandleFunc("/api/episode", scraper.HandleEpisodeAPI)
	http.HandleFunc("/api/cache/clear", func(w http.ResponseWriter, r *http.Request) {
		scraper.ClearCache()
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Cache cleared"))
	})

	// Optional: Warmup cache with popular episodes
	popularURLs := []string{
		"https://winbu.tv/mikadono-sanshimai-wa-angai-choroi-episode-6/",
		// Add more popular URLs here
	}
	go scraper.WarmupCache(popularURLs)

	log.Println("API server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// Test function untuk scraping episode
func TestDetailEpisode(t *testing.T) {
	// Initialize scraper service
	scraper := NewScraperService()

	// Test URL
	testURL := "https://winbu.tv/mikadono-sanshimai-wa-angai-choroi-episode-6/"

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Test scraping
	t.Run("ScrapeEpisodePage", func(t *testing.T) {
		result, err := scraper.ScrapeEpisodePage(ctx, testURL)
		if err != nil {
			t.Fatalf("Failed to scrape episode: %v", err)
		}

		// Validate result
		if result.URL != testURL {
			t.Errorf("Expected URL %s, got %s", testURL, result.URL)
		}

		if result.EpisodeTitle == "" {
			t.Error("Episode title is empty")
		}

		if result.SeriesTitle == "" {
			t.Error("Series title is empty")
		}

		// Print result
		jsonData, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			t.Fatalf("Failed to marshal JSON: %v", err)
		}

		fmt.Println("\n=== SCRAPING RESULT ===")
		fmt.Println(string(jsonData))
		fmt.Printf("\n=== PERFORMANCE INFO ===\n")
		fmt.Printf("Episode Title: %s\n", result.EpisodeTitle)
		fmt.Printf("Series Title: %s\n", result.SeriesTitle)
		fmt.Printf("Stream Groups: %d\n", len(result.StreamGroups))
		fmt.Printf("Download Groups: %d\n", len(result.DownloadGroups))
		fmt.Printf("Recommendations: %d\n", len(result.Recommendations))
		fmt.Printf("Cached At: %s\n", result.CachedAt.Format(time.RFC3339))
	})

	// Test cache functionality
	t.Run("CacheTest", func(t *testing.T) {
		// First request - should scrape fresh
		start := time.Now()
		result1, err := scraper.ScrapeEpisodePage(ctx, testURL)
		firstRequestDuration := time.Since(start)

		if err != nil {
			t.Fatalf("First request failed: %v", err)
		}

		// Second request - should use cache
		start = time.Now()
		result2, err := scraper.ScrapeEpisodePage(ctx, testURL)
		secondRequestDuration := time.Since(start)

		if err != nil {
			t.Fatalf("Second request failed: %v", err)
		}

		// Cache should be much faster
		if secondRequestDuration >= firstRequestDuration {
			t.Logf("Warning: Cache may not be working optimally. First: %v, Second: %v",
				firstRequestDuration, secondRequestDuration)
		}

		// Results should be identical
		if result1.EpisodeTitle != result2.EpisodeTitle {
			t.Error("Cached result differs from fresh result")
		}

		fmt.Printf("\n=== CACHE PERFORMANCE ===\n")
		fmt.Printf("First request (fresh): %v\n", firstRequestDuration)
		fmt.Printf("Second request (cached): %v\n", secondRequestDuration)
		fmt.Printf("Cache speedup: %.2fx\n", float64(firstRequestDuration)/float64(secondRequestDuration))

		// Check cache stats
		itemCount, _ := scraper.GetCacheStats()
		fmt.Printf("Cache items: %d\n", itemCount)
	})
}

// Benchmark test untuk mengukur performa
func BenchmarkScrapeEpisode(b *testing.B) {
	scraper := NewScraperService()
	testURL := "https://winbu.tv/mikadono-sanshimai-wa-angai-choroi-episode-6/"
	ctx := context.Background()

	// Warm up cache first
	scraper.ScrapeEpisodePage(ctx, testURL)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := scraper.ScrapeEpisodePage(ctx, testURL)
			if err != nil {
				b.Errorf("Scraping failed: %v", err)
			}
		}
	})
}

// Test concurrent access
func TestConcurrentAccess(t *testing.T) {
	scraper := NewScraperService()
	testURL := "https://winbu.tv/mikadono-sanshimai-wa-angai-choroi-episode-6/"

	// Number of concurrent requests to test
	concurrency := 50
	results := make(chan error, concurrency)

	start := time.Now()

	// Launch concurrent requests
	for i := 0; i < concurrency; i++ {
		go func(id int) {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			_, err := scraper.ScrapeEpisodePage(ctx, testURL)
			results <- err
		}(i)
	}

	// Collect results
	var errors []error
	for i := 0; i < concurrency; i++ {
		if err := <-results; err != nil {
			errors = append(errors, err)
		}
	}

	duration := time.Since(start)

	fmt.Printf("\n=== CONCURRENT ACCESS TEST ===\n")
	fmt.Printf("Concurrent requests: %d\n", concurrency)
	fmt.Printf("Total duration: %v\n", duration)
	fmt.Printf("Average per request: %v\n", duration/time.Duration(concurrency))
	fmt.Printf("Successful requests: %d\n", concurrency-len(errors))
	fmt.Printf("Failed requests: %d\n", len(errors))

	if len(errors) > 0 {
		fmt.Println("Errors:")
		for i, err := range errors {
			fmt.Printf("  %d: %v\n", i+1, err)
		}
	}

	// Allow up to 10% failure rate for concurrent access
	failureRate := float64(len(errors)) / float64(concurrency)
	if failureRate > 0.1 {
		t.Errorf("Failure rate too high: %.2f%% (%d/%d)", failureRate*100, len(errors), concurrency)
	}
}

// Test untuk validasi struktur data
func TestDataStructureValidation(t *testing.T) {
	scraper := NewScraperService()
	testURL := "https://winbu.tv/mikadono-sanshimai-wa-angai-choroi-episode-6/"

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	result, err := scraper.ScrapeEpisodePage(ctx, testURL)
	if err != nil {
		t.Fatalf("Failed to scrape: %v", err)
	}

	// Validate required fields
	t.Run("RequiredFields", func(t *testing.T) {
		if result.URL == "" {
			t.Error("URL is empty")
		}
		if result.EpisodeTitle == "" {
			t.Error("Episode title is empty")
		}
		if result.SeriesTitle == "" {
			t.Error("Series title is empty")
		}
		if result.CachedAt.IsZero() {
			t.Error("CachedAt timestamp is zero")
		}
	})

	// Validate stream groups
	t.Run("StreamGroups", func(t *testing.T) {
		if len(result.StreamGroups) == 0 {
			t.Error("No stream groups found")
			return
		}

		for i, group := range result.StreamGroups {
			if group.Quality == "" {
				t.Errorf("Stream group %d has empty quality", i)
			}
			if len(group.Servers) == 0 {
				t.Errorf("Stream group %d has no servers", i)
			}

			for j, server := range group.Servers {
				if server.Name == "" {
					t.Errorf("Stream group %d, server %d has empty name", i, j)
				}
				// StreamURL might be empty if fetch failed, that's ok
			}
		}
	})

	// Validate download groups
	t.Run("DownloadGroups", func(t *testing.T) {
		for i, group := range result.DownloadGroups {
			if group.Quality == "" {
				t.Errorf("Download group %d has empty quality", i)
			}

			for j, link := range group.DownloadLinks {
				if link.Provider == "" {
					t.Errorf("Download group %d, link %d has empty provider", i, j)
				}
				if link.URL == "" {
					t.Errorf("Download group %d, link %d has empty URL", i, j)
				}
			}
		}
	})

	// Validate recommendations
	t.Run("Recommendations", func(t *testing.T) {
		for i, rec := range result.Recommendations {
			if rec.Title == "" {
				t.Errorf("Recommendation %d has empty title", i)
			}
			if rec.URL == "" {
				t.Errorf("Recommendation %d has empty URL", i)
			}
		}
	})
}
