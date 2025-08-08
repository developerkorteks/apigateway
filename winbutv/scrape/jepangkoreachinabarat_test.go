package scrape

import (
	"encoding/json"
	"fmt"
	"log"
	"testing"

	"github.com/gocolly/colly/v2"
)

// --- Structs untuk Series ---
type SeriesItem struct {
	Title    string `json:"title"`
	URL      string `json:"url"`
	ImageURL string `json:"imageUrl"`
	Rating   string `json:"rating,omitempty"`
	Episode  string `json:"episode,omitempty"`
	Time     string `json:"time,omitempty"`
	Views    string `json:"views,omitempty"`
}

// --- Structs Pagination & Result (Bisa digunakan kembali) ---
// type PaginationInfo struct {
// 	CurrentPage     string `json:"currentPage"`
// 	LastVisiblePage string `json:"lastVisiblePage"`
// 	NextPageURL     string `json:"nextPageUrl,omitempty"`
// }

type SeriesScrapedResult struct {
	SourceURL      string         `json:"sourceUrl"`
	PagesScraped   int            `json:"pagesScraped"`
	PaginationInfo PaginationInfo `json:"paginationInfo"`
	TotalItems     int            `json:"totalItems"`
	Items          []SeriesItem   `json:"items"`
}

// --- Fungsi Helper (Bisa digunakan kembali) ---
// func cleanText(text string) string {
// 	re := regexp.MustCompile(`\s+`)
// 	return strings.TrimSpace(re.ReplaceAllString(text, " "))
// }

// --- FUNGSI TEST BARU UNTUK SCRAPE OTHERS/SERIES ---
func TestScrapeOthersLimited(t *testing.T) {
	startURL := "https://winbu.tv/others/" // URL diganti ke /others/
	result := &SeriesScrapedResult{
		SourceURL: startURL,
		Items:     []SeriesItem{},
	}

	const maxPagesToScrape = 3
	pagesScraped := 0

	c := colly.NewCollector(
		colly.AllowedDomains("winbu.tv"),
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36"),
	)

	// Callback untuk setiap item series
	c.OnHTML("div.ml-item.ml-item-anime.ml-item-latest.ml-potrait", func(e *colly.HTMLElement) {
		item := SeriesItem{
			Title:    cleanText(e.ChildText(".judul")),
			URL:      e.ChildAttr("a.ml-mask", "href"),
			ImageURL: e.ChildAttr("img.mli-thumb", "src"),
			Rating:   cleanText(e.ChildText("span.mli-mvi[style*='text-align:right']")),
			Episode:  cleanText(e.ChildText(".mli-episode")),
			Time:     cleanText(e.ChildText(".mli-waktu")),
			Views:    cleanText(e.ChildText(".mli-info .mli-mvi")),
		}
		result.Items = append(result.Items, item)
	})

	// Callback untuk pagination (Logika sama persis, tidak perlu diubah)
	c.OnHTML("ul.pagination", func(e *colly.HTMLElement) {
		currentPage := cleanText(e.ChildText("li.active a"))
		nextPageURL := e.ChildAttr("li:last-child a", "href")
		lastVisiblePage := e.ChildText("li:nth-last-child(2) a")

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

	fmt.Println("\n--- SCRAPING OTHERS SELESAI (TERBATAS) ---")
	fmt.Println(string(jsonData))
}
