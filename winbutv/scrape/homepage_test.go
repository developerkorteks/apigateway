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

// Mendefinisikan struct yang sangat lengkap untuk semua data dengan tag JSON.

type PageInfo struct {
	Title         string `json:"title"`
	Description   string `json:"description"`
	CanonicalURL  string `json:"canonicalUrl"`
	OgTitle       string `json:"ogTitle"`
	OgDescription string `json:"ogDescription"`
	OgURL         string `json:"ogUrl"`
	OgImage       string `json:"ogImage"`
	TwitterCard   string `json:"twitterCard"`
	TwitterSite   string `json:"twitterSite"`
}

type NavItem struct {
	Text string `json:"text"`
	URL  string `json:"url"`
}

type ContentItem struct {
	Rank     string `json:"rank,omitempty"`
	Title    string `json:"title"`
	URL      string `json:"url"`
	ImageURL string `json:"imageUrl"`
	Rating   string `json:"rating,omitempty"`
	Episode  string `json:"episode,omitempty"`
	Time     string `json:"time,omitempty"`
	Views    string `json:"views,omitempty"`
	Quality  string `json:"quality,omitempty"`
}

type Genre struct {
	Name  string `json:"name"`
	URL   string `json:"url"`
	Count string `json:"count"`
}

type ContentSection struct {
	SectionTitle string        `json:"sectionTitle"`
	MoreLink     string        `json:"moreLink"`
	Items        []ContentItem `json:"items"`
}

// Struct utama untuk menampung semua data yang di-scrape.
type ScrapedData struct {
	PageInfo           PageInfo       `json:"pageInfo"`
	Notice             string         `json:"notice"`
	NavigationMenu     []NavItem      `json:"navigationMenu"`
	Top10Series        []ContentItem  `json:"top10Series"`
	LatestDonghuaAnime ContentSection `json:"latestDonghuaAnime"`
	Genres             []Genre        `json:"genres"`
	Top10Films         []ContentItem  `json:"top10Films"`
	LatestFilms        ContentSection `json:"latestFilms"`
	OtherSeries        ContentSection `json:"otherSeries"` // Jepang, Korea, China, Barat
	TVShows            ContentSection `json:"tvShows"`
}

// TestHome akan melakukan scraping dan mencetak hasilnya dalam format JSON.
func TestHome(t *testing.T) {
	url := "https://winbu.tv/"

	c := colly.NewCollector()
	data := ScrapedData{}

	// Fungsi helper untuk membersihkan teks dari spasi dan karakter aneh.
	cleanText := func(text string) string {
		// Menghapus tab, newline, dan spasi berlebih
		re := regexp.MustCompile(`\s\s+`)
		return strings.TrimSpace(re.ReplaceAllString(text, " "))
	}

	// 1. Ekstrak Metadata Halaman (Sangat Lengkap)
	c.OnHTML("head", func(e *colly.HTMLElement) {
		data.PageInfo = PageInfo{
			Title:         cleanText(e.ChildText("title")),
			Description:   e.ChildAttr("meta[name=description]", "content"),
			CanonicalURL:  e.ChildAttr("link[rel=canonical]", "href"),
			OgTitle:       e.ChildAttr("meta[property='og:title']", "content"),
			OgDescription: e.ChildAttr("meta[property='og:description']", "content"),
			OgURL:         e.ChildAttr("meta[property='og:url']", "content"),
			OgImage:       e.ChildAttr("meta[property='og:image']", "content"),
			TwitterCard:   e.ChildAttr("meta[name='twitter:card']", "content"),
			TwitterSite:   e.ChildAttr("meta[name='twitter:site']", "content"),
		}
	})

	// Ekstrak Notice
	c.OnHTML("div.marquee", func(e *colly.HTMLElement) {
		if data.Notice == "" { // Hanya ambil yang pertama untuk menghindari duplikasi dari plugin marquee
			data.Notice = cleanText(e.Text)
		}
	})

	// Ekstrak Menu Navigasi
	c.OnHTML("ul#menu-menu > li.menu-item", func(e *colly.HTMLElement) {
		item := NavItem{
			Text: cleanText(e.ChildText("span[itemprop=name]")),
			URL:  e.ChildAttr("a", "href"),
		}
		data.NavigationMenu = append(data.NavigationMenu, item)
	})

	// 2. Ekstrak Semua Bagian Konten
	c.OnHTML("div.movies-list-wrap", func(e *colly.HTMLElement) {
		sectionTitle := cleanText(e.ChildText(".list-title h2"))

		switch {
		// Top 10 Series & Top 10 Film
		case strings.Contains(sectionTitle, "Top 10 Series"), strings.Contains(sectionTitle, "Top 10 Film"):
			e.ForEach(".ml-item-potrait .ml-item", func(_ int, el *colly.HTMLElement) {
				item := ContentItem{
					Title:    cleanText(el.ChildText(".judul")),
					URL:      el.ChildAttr("a.ml-mask", "href"),
					ImageURL: el.ChildAttr("img.mli-thumb", "src"),
					Rating:   cleanText(el.ChildText(".mli-mvi")),
					Rank:     cleanText(el.ChildText(".mli-topten b")),
				}
				if strings.Contains(sectionTitle, "Top 10 Series") {
					data.Top10Series = append(data.Top10Series, item)
				} else {
					data.Top10Films = append(data.Top10Films, item)
				}
			})

		// Section dengan List Item (Anime, Film, dll)
		case strings.Contains(sectionTitle, "Anime Donghua Terbaru"), strings.Contains(sectionTitle, "Film Terbaru"), strings.Contains(sectionTitle, "Jepang Korea China Barat"), strings.Contains(sectionTitle, "TV Show"):
			var items []ContentItem
			e.ForEach(".ml-item", func(_ int, el *colly.HTMLElement) {
				ratingText := ""
				// Untuk Film Terbaru, rating ada di tempat berbeda
				if strings.Contains(sectionTitle, "Film Terbaru") {
					ratingText = cleanText(el.ChildText("span.mli-mvi[style*='text-align:right']"))
				}

				item := ContentItem{
					Title:    cleanText(el.ChildText(".judul")),
					URL:      el.ChildAttr("a.ml-mask", "href"),
					ImageURL: el.ChildAttr("img.mli-thumb", "src"),
					Episode:  cleanText(el.ChildText(".mli-episode")),
					Time:     cleanText(el.ChildText(".mli-waktu")),
					Views:    cleanText(el.ChildText("span.mli-mvi:not([style*='text-align:right'])")),
					Quality:  cleanText(el.ChildText(".mli-quality")),
					Rating:   ratingText,
				}
				items = append(items, item)
			})

			section := ContentSection{
				SectionTitle: sectionTitle,
				MoreLink:     e.ChildAttr("a.pull-right", "href"),
				Items:        items,
			}

			if strings.Contains(sectionTitle, "Anime Donghua Terbaru") {
				data.LatestDonghuaAnime = section
			} else if strings.Contains(sectionTitle, "Film Terbaru") {
				data.LatestFilms = section
			} else if strings.Contains(sectionTitle, "Jepang Korea China Barat") {
				data.OtherSeries = section
			} else {
				data.TVShows = section
			}
		}
	})

	// 3. Ekstrak Genres
	c.OnHTML("aside#sidebar ul.genres li", func(e *colly.HTMLElement) {
		count := cleanText(e.ChildText("span"))
		fullText := cleanText(e.ChildText("a"))
		name := strings.TrimSpace(strings.Replace(fullText, count, "", 1))

		genre := Genre{
			Name:  name,
			URL:   e.ChildAttr("a", "href"),
			Count: strings.Trim(count, "()"),
		}
		data.Genres = append(data.Genres, genre)
	})

	c.OnError(func(r *colly.Response, err error) {
		log.Println("Request URL:", r.Request.URL, "failed with response:", r, "\nError:", err)
	})

	// Setelah selesai, marshal data ke JSON dan cetak.
	c.OnScraped(func(r *colly.Response) {
		jsonData, err := json.MarshalIndent(data, "", "  ") // Pakai indentasi untuk keterbacaan
		if err != nil {
			log.Fatal("Gagal melakukan marshal JSON:", err)
		}

		fmt.Println(string(jsonData))
	})

	// Mulai mengunjungi URL
	err := c.Visit(url)
	if err != nil {
		log.Fatal("Gagal mengunjungi URL:", err)
	}
}
