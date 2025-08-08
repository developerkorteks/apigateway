package scrape

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"regexp"
	"strings"
	"testing"

	"github.com/gocolly/colly/v2"
)

// ===================================================================================
// 1. DEFINISI STRUCT UNIK UNTUK MENAMPUNG OPSI FILTER
// ===================================================================================

// FilterOption merepresentasikan satu pilihan filter (misal: Nama "Popular", Value "popular").
type FilterOption struct {
	DisplayName string `json:"displayName"`
	QueryValue  string `json:"queryValue"`
}

// AvailableFilters adalah struct utama yang menampung semua kategori filter yang tersedia.
type AvailableFilters struct {
	StatusOptions []FilterOption `json:"statusOptions"`
	TypeOptions   []FilterOption `json:"typeOptions"`
	OrderOptions  []FilterOption `json:"orderOptions"`
	GenreOptions  []FilterOption `json:"genreOptions"`
}

// ===================================================================================
// 2. FUNGSI TEST SPESIFIK UNTUK MENGAMBIL SEMUA OPSI FILTER
// ===================================================================================

func TestScrapeAvailableFilters(t *testing.T) {
	startURL := "https://winbu.tv/daftar-anime-2/"
	availableFilters := &AvailableFilters{}

	// Fungsi helper untuk membersihkan teks dari label genre
	cleanGenreText := func(fullText string) string {
		re := regexp.MustCompile(`\s\s+`) // Menghapus spasi ganda
		return strings.TrimSpace(re.ReplaceAllString(fullText, " "))
	}

	c := colly.NewCollector(
		colly.AllowedDomains("winbu.tv"),
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/115.0.0.0 Safari/537.36"),
	)

	// Callback untuk setiap baris (<tr>) di dalam form filter
	c.OnHTML("div.filtersearch form tr", func(e *colly.HTMLElement) {
		filterTitle := cleanText(e.ChildText(".filter_title"))

		switch filterTitle {
		case "Status":
			e.ForEach("label.radio", func(_ int, el *colly.HTMLElement) {
				option := FilterOption{
					DisplayName: cleanText(el.Text),
					QueryValue:  el.ChildAttr("input[type=radio]", "value"),
				}
				availableFilters.StatusOptions = append(availableFilters.StatusOptions, option)
			})

		case "Type":
			e.ForEach("label.radio", func(_ int, el *colly.HTMLElement) {
				option := FilterOption{
					DisplayName: cleanText(el.Text),
					QueryValue:  el.ChildAttr("input[type=radio]", "value"),
				}
				availableFilters.TypeOptions = append(availableFilters.TypeOptions, option)
			})

		case "Urutkan Berdasarkan":
			e.ForEach("ul.filter-sort li", func(_ int, el *colly.HTMLElement) {
				option := FilterOption{
					DisplayName: cleanText(el.ChildText("label")),
					QueryValue:  el.ChildAttr("input[type=radio]", "value"),
				}
				availableFilters.OrderOptions = append(availableFilters.OrderOptions, option)
			})

		case "Genre":
			e.ForEach("label.tax_fil", func(_ int, el *colly.HTMLElement) {
				option := FilterOption{
					DisplayName: cleanGenreText(el.Text),
					QueryValue:  el.ChildAttr("input[type=checkbox]", "value"),
				}
				availableFilters.GenreOptions = append(availableFilters.GenreOptions, option)
			})
		}
	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Mengunjungi:", r.URL.String())
	})

	c.OnError(func(r *colly.Response, err error) {
		log.Printf("Error: %s on %s (Status Code: %d)\n", err, r.Request.URL, r.StatusCode)
	})

	err := c.Visit(startURL)
	if err != nil {
		t.Fatalf("Failed to start scraping: %v", err)
	}

	c.Wait()

	jsonData, err := json.MarshalIndent(availableFilters, "", "  ")
	if err != nil {
		t.Fatalf("JSON marshaling failed: %v", err)
	}

	fmt.Println("\n--- OPSI FILTER YANG TERSEDIA ---")
	fmt.Println(string(jsonData))
}

// ===================================================================================
// 1. DEFINISI STRUCT UNIK UNTUK HALAMAN "DAFTAR ANIME"
// ===================================================================================

type DaftarAnimeItem struct {
	Title    string `json:"title"`
	URL      string `json:"url"`
	ImageURL string `json:"imageUrl"`
	Time     string `json:"time,omitempty"`
	Views    string `json:"views,omitempty"`
}

type DaftarAnimePaginationInfo struct {
	CurrentPage     string `json:"currentPage"`
	LastVisiblePage string `json:"lastVisiblePage"`
	NextPageURL     string `json:"nextPageUrl,omitempty"`
}

type DaftarAnimeScrapedResult struct {
	SourceURL      string                    `json:"sourceUrl"`
	PagesScraped   int                       `json:"pagesScraped"`
	PaginationInfo DaftarAnimePaginationInfo `json:"paginationInfo"`
	TotalItems     int                       `json:"totalItems"`
	Items          []DaftarAnimeItem         `json:"items"`
}

// Fungsi helper
func cleanTextDaftar(text string) string {
	re := regexp.MustCompile(`\s+`)
	return strings.TrimSpace(re.ReplaceAllString(text, " "))
}

// ===================================================================================
// 2. FUNGSI TEST SPESIFIK UNTUK SCRAPE "DAFTAR ANIME" DENGAN FILTER
// ===================================================================================

func TestScrapeDaftarAnimeFiltered(t *testing.T) {
	// --- PENGATURAN FILTER ---
	// Anda bisa mengubah nilai-nilai ini untuk mendapatkan hasil yang berbeda
	baseURL := "https://winbu.tv/daftar-anime-2/"
	searchTitle := "naruto"
	searchStatus := "" // Kosongkan untuk "All"
	searchType := "TV"
	searchOrder := "popular"
	searchGenres := []string{"action", "adventure"} // Contoh genre

	// Membangun URL dengan parameter filter
	u, err := url.Parse(baseURL)
	if err != nil {
		t.Fatalf("Gagal mem-parse base URL: %v", err)
	}
	q := u.Query()
	q.Set("title", searchTitle)
	q.Set("status", searchStatus)
	q.Set("type", searchType)
	q.Set("order", searchOrder)
	for _, genre := range searchGenres {
		q.Add("genre[]", genre)
	}
	u.RawQuery = q.Encode()
	startURL := u.String()
	// Hasil URL: https://winbu.tv/daftar-anime-2/?genre%5B%5D=action&genre%5B%5D=adventure&order=popular&status=&title=naruto&type=TV

	// --- LOGIKA SCRAPING (Sama seperti sebelumnya) ---
	result := &DaftarAnimeScrapedResult{
		SourceURL: startURL,
		Items:     []DaftarAnimeItem{},
	}

	const maxPagesToScrape = 3
	pagesScraped := 0

	c := colly.NewCollector(
		colly.AllowedDomains("winbu.tv"),
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/115.0.0.0 Safari/537.36"),
	)

	c.OnHTML("div.ml-item.ml-item-anime.ml-item-latest", func(e *colly.HTMLElement) {
		item := DaftarAnimeItem{
			Title:    cleanTextDaftar(e.ChildText(".judul")),
			URL:      e.ChildAttr("a.ml-mask", "href"),
			ImageURL: e.ChildAttr("img.mli-thumb", "src"),
			Time:     cleanTextDaftar(e.ChildText(".mli-waktu")),
			Views:    cleanTextDaftar(e.ChildText(".mli-info .mli-mvi")),
		}
		result.Items = append(result.Items, item)
	})

	c.OnHTML("ul.pagination", func(e *colly.HTMLElement) {
		currentPage := cleanTextDaftar(e.ChildText("li.active a"))
		nextPageURL := e.ChildAttr("li:last-child a", "href")
		lastVisiblePage := cleanTextDaftar(e.ChildText("li:nth-last-child(2) a"))

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

	err = c.Visit(startURL)
	if err != nil {
		t.Fatalf("Failed to start scraping: %v", err)
	}

	c.Wait()

	result.PagesScraped = pagesScraped
	result.TotalItems = len(result.Items)
	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal JSON: %v", err)
	}

	fmt.Println("\n--- SCRAPING DAFTAR ANIME (FILTERED) SELESAI ---")
	fmt.Println(string(jsonData))
}
