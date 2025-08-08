package scrape

import (
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strings"
	"testing"

	"github.com/gocolly/colly/v2"
)

// ===================================================================================
// 1. DEFINISI STRUCT UNIK UNTUK HALAMAN "ANIME TERBARU ANIMASU"
// ===================================================================================

type AnimasuItem struct {
	Title    string `json:"title"`
	URL      string `json:"url"`
	ImageURL string `json:"imageUrl"`
	Episode  string `json:"episode,omitempty"`
	Time     string `json:"time,omitempty"`
	Views    string `json:"views,omitempty"`
}

type AnimasuPaginationInfo struct {
	CurrentPage     string `json:"currentPage"`
	LastVisiblePage string `json:"lastVisiblePage"`
	NextPageURL     string `json:"nextPageUrl,omitempty"`
}

type AnimasuScrapedResult struct {
	SourceURL      string                `json:"sourceUrl"`
	PagesScraped   int                   `json:"pagesScraped"`
	PaginationInfo AnimasuPaginationInfo `json:"paginationInfo"`
	TotalItems     int                   `json:"totalItems"`
	Items          []AnimasuItem         `json:"items"`
}

// Fungsi helper (dapat didefinisikan ulang jika perlu atau diletakkan di file terpisah)
func cleanTextAnimasu(text string) string {
	re := regexp.MustCompile(`\s+`)
	return strings.TrimSpace(re.ReplaceAllString(text, " "))
}

// ===================================================================================
// 2. FUNGSI TEST SPESIFIK UNTUK SCRAPE "ANIME TERBARU ANIMASU"
// ===================================================================================

func TestScrapeAnimeTerbaruAnimasuLimited(t *testing.T) {
	startURL := "https://winbu.tv/anime-terbaru-animasu/"
	result := &AnimasuScrapedResult{
		SourceURL: startURL,
		Items:     []AnimasuItem{},
	}

	const maxPagesToScrape = 3
	pagesScraped := 0

	c := colly.NewCollector(
		colly.AllowedDomains("winbu.tv"),
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/115.0.0.0 Safari/537.36"),
	)

	// Callback untuk setiap item di halaman
	c.OnHTML("div.ml-item.ml-item-anime.ml-item-latest", func(e *colly.HTMLElement) {
		item := AnimasuItem{
			Title:    cleanTextAnimasu(e.ChildText(".judul")),
			URL:      e.ChildAttr("a.ml-mask", "href"),
			ImageURL: e.ChildAttr("img.mli-thumb", "src"),
			Episode:  cleanTextAnimasu(e.ChildText(".mli-episode")),
			Time:     cleanTextAnimasu(e.ChildText(".mli-waktu")),
			Views:    cleanTextAnimasu(e.ChildText(".mli-info .mli-mvi")),
		}
		result.Items = append(result.Items, item)
	})

	// Callback untuk pagination (menggunakan selector dari file HTML yang sesuai)
	c.OnHTML("ul.pagination", func(e *colly.HTMLElement) {
		currentPage := cleanTextAnimasu(e.ChildText("li.active a"))
		nextPageURL := e.ChildAttr("li:last-child a", "href")
		lastVisiblePage := cleanTextAnimasu(e.ChildText("li:nth-last-child(2) a"))

		result.PaginationInfo.CurrentPage = currentPage
		result.PaginationInfo.LastVisiblePage = lastVisiblePage
		result.PaginationInfo.NextPageURL = nextPageURL

		if pagesScraped >= maxPagesToScrape {
			fmt.Printf("Batas %d halaman tercapai. Berhenti.\n", maxPagesToScrape)
			return
		}

		if nextPageURL != "" && currentPage != lastVisiblePage {
			c.Visit(nextPageURL)
		}
	})

	c.OnRequest(func(r *colly.Request) {
		pagesScraped++
		fmt.Printf("Mengunjungi Halaman %d: %s\n", pagesScraped, r.URL.String())
	})

	c.OnError(func(r *colly.Response, err error) {
		log.Printf("Error scraping %s: %s (Status: %d)\n", r.Request.URL, err, r.StatusCode)
	})

	err := c.Visit(startURL)
	if err != nil {
		t.Fatalf("Failed to start scraping: %v", err)
	}

	c.Wait()

	// Cetak hasil final
	result.PagesScraped = pagesScraped
	result.TotalItems = len(result.Items)
	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal JSON: %v", err)
	}

	fmt.Println("\n--- SCRAPING ANIME TERBARU (ANIMASU) SELESAI ---")
	fmt.Println(string(jsonData))
}
