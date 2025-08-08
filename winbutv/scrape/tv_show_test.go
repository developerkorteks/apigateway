package scrape

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"testing"

	"github.com/gocolly/colly/v2"
)

// --- Structs untuk TV Show ---
// Mirip dengan SeriesItem
type TVShowItem struct {
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

type TVShowScrapedResult struct {
	SourceURL      string         `json:"sourceUrl"`
	PagesScraped   int            `json:"pagesScraped"`
	PaginationInfo PaginationInfo `json:"paginationInfo"`
	TotalItems     int            `json:"totalItems"`
	Items          []TVShowItem   `json:"items"`
}

// --- Fungsi Helper (Bisa digunakan kembali) ---
// func cleanText(text string) string {
// 	re := regexp.MustCompile(`\s+`)
// 	return strings.TrimSpace(re.ReplaceAllString(text, " "))
// }

// --- FUNGSI TEST BARU UNTUK SCRAPE TV SHOW ---
func TestScrapeTVShowLimited(t *testing.T) {
	startURL := "https://winbu.tv/tvshow/" // URL diganti ke /tvshow/
	result := &TVShowScrapedResult{
		SourceURL: startURL,
		Items:     []TVShowItem{},
	}

	const maxPagesToScrape = 3
	pagesScraped := 0

	c := colly.NewCollector(
		colly.AllowedDomains("winbu.tv"),
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36"),
	)

	// Callback untuk setiap item TV Show
	c.OnHTML("div.ml-item.ml-item-anime.ml-item-latest.ml-potrait", func(e *colly.HTMLElement) {
		item := TVShowItem{
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

	// Callback untuk pagination (Logika sama persis, tidak ada perubahan)
	// Namun, selector next page sedikit berbeda di file HTML ini.
	c.OnHTML("ul.pagination", func(e *colly.HTMLElement) {
		currentPage := cleanText(e.ChildText("li.active a"))

		// Untuk halaman ini, file HTML tidak memiliki tombol next ">", jadi kita ambil link halaman terbesar yang terlihat.
		// Ini adalah adaptasi kecil berdasarkan file HTML yang diberikan.
		var nextPageURL string
		var lastVisiblePage string
		e.ForEach("a.page", func(_ int, el *colly.HTMLElement) {
			nextPageURL = el.Attr("href") // Ambil href terakhir
			lastVisiblePage = cleanText(el.Text)
		})

		result.PaginationInfo.CurrentPage = currentPage
		result.PaginationInfo.LastVisiblePage = lastVisiblePage

		if pagesScraped >= maxPagesToScrape {
			fmt.Printf("Batas %d halaman tercapai. Berhenti.\n", maxPagesToScrape)
			return
		}

		// Hanya kunjungi halaman berikutnya jika URL ditemukan dan berbeda dari halaman saat ini
		if nextPageURL != "" && !strings.Contains(e.Request.URL.String(), nextPageURL) {
			result.PaginationInfo.NextPageURL = nextPageURL
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

	fmt.Println("\n--- SCRAPING TV SHOW SELESAI (TERBATAS) ---")
	fmt.Println(string(jsonData))
}
