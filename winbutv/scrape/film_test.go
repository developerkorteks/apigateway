package scrape

import (
	"encoding/json"
	"fmt"
	"log"
	"testing"

	"github.com/gocolly/colly/v2"
)

// --- Structs Baru untuk Film ---
// Sedikit berbeda dari AnimeItem, ada Rating dan Quality.
type FilmItem struct {
	Title    string `json:"title"`
	URL      string `json:"url"`
	ImageURL string `json:"imageUrl"`
	Rating   string `json:"rating,omitempty"`
	Quality  string `json:"quality,omitempty"`
	Time     string `json:"time,omitempty"`
	Views    string `json:"views,omitempty"`
}

// Struct PaginationInfo dan cleanText bisa digunakan kembali (tidak perlu ditulis ulang jika dalam file yang sama)
//
//	type PaginationInfo struct {
//		CurrentPage     string `json:"currentPage"`
//		LastVisiblePage string `json:"lastVisiblePage"`
//		NextPageURL     string `json:"nextPageUrl,omitempty"`
//	}
//
// Struct Result untuk Film
type FilmScrapedResult struct {
	SourceURL      string         `json:"sourceUrl"`
	PagesScraped   int            `json:"pagesScraped"`
	PaginationInfo PaginationInfo `json:"paginationInfo"`
	TotalItems     int            `json:"totalItems"`
	Items          []FilmItem     `json:"items"`
}

// Fungsi cleanText dapat digunakan kembali
// func cleanText(text string) string {
// 	re := regexp.MustCompile(`\s+`)
// 	return strings.TrimSpace(re.ReplaceAllString(text, " "))
// }

// --- FUNGSI TEST BARU UNTUK SCRAPE FILM ---
func TestScrapeFilmLimited(t *testing.T) {
	startURL := "https://winbu.tv/film/"
	result := &FilmScrapedResult{
		SourceURL: startURL,
		Items:     []FilmItem{},
	}

	const maxPagesToScrape = 3 // Ambil 3 halaman saja
	pagesScraped := 0

	c := colly.NewCollector(
		colly.AllowedDomains("winbu.tv"),
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36"),
	)

	// Callback untuk setiap item film
	// Selector disesuaikan agar lebih spesifik untuk item di halaman ini
	c.OnHTML("div.ml-item.ml-item-anime.ml-item-latest.ml-potrait", func(e *colly.HTMLElement) {
		// Ekstrak rating yang berada di luar blok info utama
		rating := cleanText(e.ChildText("span.mli-mvi[style*='text-align:right']"))

		item := FilmItem{
			Title:    cleanText(e.ChildText(".judul")),
			URL:      e.ChildAttr("a.ml-mask", "href"),
			ImageURL: e.ChildAttr("img.mli-thumb", "src"),
			Rating:   rating,
			Quality:  cleanText(e.ChildText(".mli-quality")),
			Time:     cleanText(e.ChildText(".mli-waktu")),
			// Pastikan selector views tidak mengambil rating
			Views: cleanText(e.ChildText(".mli-info .mli-mvi")),
		}
		result.Items = append(result.Items, item)
	})

	// Callback untuk pagination (Logika sama persis seperti sebelumnya)
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

	fmt.Println("\n--- SCRAPING FILM SELESAI (TERBATAS) ---")
	fmt.Println(string(jsonData))
}
