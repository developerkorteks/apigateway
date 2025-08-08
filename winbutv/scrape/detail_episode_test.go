package scrape

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"testing"

	"github.com/gocolly/colly/v2"
)

// ===================================================================================
// 1. DEFINISI STRUCT (Tidak ada perubahan)
// ===================================================================================

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
}

func cleanEpisodeText(text string) string {
	re := regexp.MustCompile(`\s+`)
	return strings.TrimSpace(re.ReplaceAllString(text, " "))
}

func getStreamURL(postID, nume, dataType string) (string, error) {
	ajaxURL := "https://winbu.tv/wp-admin/admin-ajax.php"
	formData := url.Values{
		"action": {"player_ajax"},
		"post":   {postID},
		"nume":   {nume},
		"type":   {dataType},
	}

	req, err := http.NewRequest("POST", ajaxURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return "", err
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/115.0.0.0 Safari/537.36")
	req.Header.Add("X-Requested-With", "XMLHttpRequest")
	req.Header.Add("Referer", "https://winbu.tv/")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	htmlResponse := string(body)
	re := regexp.MustCompile(`src='([^']*)'|src="([^"]*)"`)
	matches := re.FindStringSubmatch(htmlResponse)

	if len(matches) < 2 {
		return "", fmt.Errorf("tidak dapat menemukan URL src iframe di dalam respons: %s", htmlResponse)
	}

	streamURL := matches[1]
	if streamURL == "" {
		streamURL = matches[2]
	}

	return streamURL, nil
}

// ===================================================================================
// 2. FUNGSI TEST LENGKAP DENGAN SEMUA SELECTOR YANG DIPERBAIKI
// ===================================================================================
func TestScrapeEpisodePageComplete(t *testing.T) {
	startURL := "https://winbu.tv/mikadono-sanshimai-wa-angai-choroi-episode-6/"
	detail := &CompleteEpisodePage{
		URL: startURL,
	}

	c := colly.NewCollector(
		colly.AllowedDomains("winbu.tv"),
		colly.Async(true),
	)
	c.Limit(&colly.LimitRule{DomainGlob: "*", Parallelism: 8})
	c.UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/115.0.0.0 Safari/537.36"

	// PERBAIKAN: Selector untuk Judul Episode
	c.OnHTML("div.list-title h2", func(e *colly.HTMLElement) {
		detail.EpisodeTitle = cleanEpisodeText(e.Text)
	})

	// PERBAIKAN: Mengambil info dari box info series di bawah
	c.OnHTML("div.m-info div.movies-list-full div.t-item", func(e *colly.HTMLElement) {
		detail.SeriesTitle = cleanEpisodeText(e.ChildText(".mli-info .judul"))
		detail.SeriesInfo.PosterImageURL = e.ChildAttr(".mli-thumb-box img", "src")

		// Iterasi melalui setiap baris info
		e.ForEach(".mli-mvi", func(_ int, el *colly.HTMLElement) {
			text := cleanEpisodeText(el.Text)
			if strings.HasPrefix(text, "Rating") {
				re := regexp.MustCompile(`(\d+\.\d+)`)
				matches := re.FindStringSubmatch(text)
				if len(matches) > 1 {
					detail.SeriesInfo.Rating = matches[1]
				}
			} else if strings.HasPrefix(text, "Genre") {
				el.ForEach("a", func(_ int, genreEl *colly.HTMLElement) {
					detail.SeriesInfo.Genres = append(detail.SeriesInfo.Genres, cleanEpisodeText(genreEl.Text))
				})
			}
		})
		detail.SeriesInfo.Synopsis = cleanEpisodeText(e.ChildText(".mli-desc"))
	})

	// Navigasi Episode (sudah benar)
	c.OnHTML("div.naveps", func(e *colly.HTMLElement) {
		detail.EpisodeNav.PreviousEpisodeURL = e.ChildAttr("div.nvs a", "href")
		detail.EpisodeNav.AllEpisodesURL = e.ChildAttr("div.nvs.nvsc a", "href")
		detail.EpisodeNav.NextEpisodeURL = e.ChildAttr("div.nvs.rght a", "href")
	})

	// Pilihan Server Streaming (sudah benar)
	c.OnHTML("div.player-modes div.dropdown", func(e *colly.HTMLElement) {
		quality := cleanEpisodeText(e.ChildText("button.dropdown-toggle"))
		var servers []StreamServer
		e.ForEach(".dropdown-item .east_player_option", func(_ int, el *colly.HTMLElement) {
			servers = append(servers, StreamServer{
				Name: cleanEpisodeText(el.ChildText("span")), PostID: el.Attr("data-post"), Nume: el.Attr("data-nume"), DataType: el.Attr("data-type"),
			})
		})
		detail.StreamGroups = append(detail.StreamGroups, StreamQualityGroup{Quality: quality, Servers: servers})
	})

	// Link Download (sudah benar)
	c.OnHTML("div.download-eps ul li", func(e *colly.HTMLElement) {
		qualityGroup := DownloadQualityGroup{Quality: cleanEpisodeText(e.ChildText("strong"))}
		e.ForEach("span a", func(_ int, el *colly.HTMLElement) {
			qualityGroup.DownloadLinks = append(qualityGroup.DownloadLinks, DownloadLink{Provider: cleanEpisodeText(el.Text), URL: el.Attr("href")})
		})
		if qualityGroup.Quality != "" {
			detail.DownloadGroups = append(detail.DownloadGroups, qualityGroup)
		}
	})

	// PERBAIKAN: Selector untuk rekomendasi
	c.OnHTML("div.rekom .ml-item-rekom", func(e *colly.HTMLElement) {
		rec := RecommendationItem{
			Title:    cleanEpisodeText(e.ChildText(".judul")),
			URL:      e.ChildAttr("a.ml-mask", "href"),
			ImageURL: e.ChildAttr("img.mli-thumb", "src"),
			Rating:   cleanEpisodeText(e.ChildText(".mli-mvi")),
		}
		detail.Recommendations = append(detail.Recommendations, rec)
	})

	c.OnRequest(func(r *colly.Request) { fmt.Println("Mengunjungi:", r.URL.String()) })

	err := c.Visit(startURL)
	if err != nil {
		t.Fatalf("Gagal memulai scraping: %v", err)
	}
	c.Wait()

	var wg sync.WaitGroup
	fmt.Println("\n--- Mengambil URL Stream dari semua server ---")

	for i := range detail.StreamGroups {
		for j := range detail.StreamGroups[i].Servers {
			wg.Add(1)
			go func(i, j int) {
				defer wg.Done()
				server := &detail.StreamGroups[i].Servers[j]
				streamURL, err := getStreamURL(server.PostID, server.Nume, server.DataType)
				if err != nil {
					log.Printf("Gagal mengambil stream untuk %s (%s): %v", server.Name, detail.StreamGroups[i].Quality, err)
					server.StreamURL = "Gagal Mengambil URL"
				} else {
					log.Printf("Berhasil -> %s (%s)", server.Name, detail.StreamGroups[i].Quality)
					server.StreamURL = streamURL
				}
			}(i, j)
		}
	}
	wg.Wait()

	jsonData, err := json.MarshalIndent(detail, "", "  ")
	if err != nil {
		t.Fatalf("JSON marshaling failed: %v", err)
	}

	fmt.Println("\n--- SCRAPING HALAMAN EPISODE SELESAI (DATA LENGKAP) ---")
	fmt.Println(string(jsonData))
}
