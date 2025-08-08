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

// --- Structs (tidak ada perubahan) ---
type AnimeItem struct {
	Title    string `json:"title"`
	URL      string `json:"url"`
	ImageURL string `json:"imageUrl"`
	Episode  string `json:"episode"`
	Time     string `json:"time"`
	Views    string `json:"views"`
}

type PaginationInfo struct {
	CurrentPage     string `json:"currentPage"`
	LastVisiblePage string `json:"lastVisiblePage"`
	NextPageURL     string `json:"nextPageUrl,omitempty"`
}

type ScrapedResult struct {
	SourceURL      string         `json:"sourceUrl"`
	PagesScraped   int            `json:"pagesScraped"` // BARU: Menambahkan info jumlah halaman yang di-scrape
	PaginationInfo PaginationInfo `json:"paginationInfo"`
	TotalItems     int            `json:"totalItems"`
	Items          []AnimeItem    `json:"items"`
}

func cleanText(text string) string {
	re := regexp.MustCompile(`\s+`)
	return strings.TrimSpace(re.ReplaceAllString(text, " "))
}

func TestScrapeAnimeDonghuaLimited(t *testing.T) {
	startURL := "https://winbu.tv/animedonghua/"
	result := &ScrapedResult{
		SourceURL: startURL,
		Items:     []AnimeItem{},
	}

	// --- BARU: Definisikan batasan dan penghitung halaman ---
	const maxPagesToScrape = 3 // Ambil 3 halaman saja
	pagesScraped := 0

	c := colly.NewCollector(
		colly.AllowedDomains("winbu.tv"),
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36"),
	)

	// Callback untuk item anime (tidak ada perubahan)
	c.OnHTML("div.movies-list div.ml-item.ml-item-anime", func(e *colly.HTMLElement) {
		item := AnimeItem{
			Title:    cleanText(e.ChildText(".judul")),
			URL:      e.ChildAttr("a.ml-mask", "href"),
			ImageURL: e.ChildAttr("img.mli-thumb", "src"),
			Episode:  cleanText(e.ChildText(".mli-episode")),
			Time:     cleanText(e.ChildText(".mli-waktu")),
			Views:    cleanText(e.ChildText(".mli-mvi")),
		}
		result.Items = append(result.Items, item)
	})

	// Callback untuk pagination (dengan logika berhenti)
	c.OnHTML("ul.pagination", func(e *colly.HTMLElement) {
		// Logika ekstraksi info pagination (tidak ada perubahan)
		currentPage := cleanText(e.ChildText("li.active a"))
		nextPageURL := e.ChildAttr("li:last-child a", "href")
		lastVisiblePage := e.ChildText("li:nth-last-child(2) a")
		result.PaginationInfo.CurrentPage = currentPage
		result.PaginationInfo.LastVisiblePage = lastVisiblePage
		result.PaginationInfo.NextPageURL = nextPageURL

		// --- BARU: Logika untuk berhenti setelah `maxPagesToScrape` tercapai ---
		if pagesScraped >= maxPagesToScrape {
			fmt.Printf("Batas %d halaman tercapai. Berhenti.\n", maxPagesToScrape)
			return // Hentikan proses pencarian halaman berikutnya
		}

		if nextPageURL != "" && currentPage != lastVisiblePage {
			c.Visit(nextPageURL)
		}
	})

	c.OnError(func(r *colly.Response, err error) {
		log.Printf("Error scraping %s: %s (Status: %d)\n", r.Request.URL, err, r.StatusCode)
	})

	c.OnRequest(func(r *colly.Request) {
		// --- BARU: Tambahkan 1 ke penghitung setiap kali mengunjungi halaman baru ---
		pagesScraped++
		fmt.Printf("Mengunjungi Halaman %d: %s\n", pagesScraped, r.URL.String())
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

	fmt.Println("\n--- SCRAPING SELESAI (TERBATAS) ---")
	fmt.Println(string(jsonData))
}
