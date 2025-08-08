package scrape

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly/v2"
)

type Episode struct {
	Title string `json:"title"`
	URL   string `json:"url"`
}

type Recommendation struct {
	Title    string `json:"title"`
	URL      string `json:"url"`
	ImageURL string `json:"imageUrl"`
	Rating   string `json:"rating"`
}

type AnimeDetail struct {
	URL             string           `json:"url"`
	Title           string           `json:"title"`
	PosterImageURL  string           `json:"posterImageUrl"`
	TrailerURL      string           `json:"trailerUrl"`
	Rating          string           `json:"rating"`
	ReleaseDate     string           `json:"releaseDate"`
	Genres          []string         `json:"genres"`
	Synopsis        string           `json:"synopsis"`
	Episodes        []Episode        `json:"episodes"`
	Recommendations []Recommendation `json:"recommendations"`
}

func cleanDetailText(text string) string {
	re := regexp.MustCompile(`\s+`)
	return strings.TrimSpace(re.ReplaceAllString(text, " "))
}

func TestScrapeAnimeDetail(t *testing.T) {
	startURL := "https://winbu.tv/series/legend-of-the-female-general/"
	detail := &AnimeDetail{
		URL: startURL,
	}

	c := colly.NewCollector(
		colly.AllowedDomains("winbu.tv"),
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/115.0.0.0 Safari/537.36"),
	)

	// Info utama
	c.OnHTML("div.m-info", func(e *colly.HTMLElement) {
		detail.Title = cleanDetailText(e.ChildText(".mli-info .judul"))
		detail.PosterImageURL = e.ChildAttr(".mli-thumb-box img", "src")

		ratingText := e.DOM.Find(".mli-mvi").FilterFunction(func(i int, s *goquery.Selection) bool {
			return strings.Contains(s.Text(), "Rating")
		}).Text()
		re := regexp.MustCompile(`(\d+\.\d+)\s+/\s+\d+`)
		matches := re.FindStringSubmatch(ratingText)
		if len(matches) > 1 {
			detail.Rating = matches[1]
		}

		detail.ReleaseDate = cleanDetailText(e.DOM.Find(".mli-mvi").FilterFunction(func(i int, s *goquery.Selection) bool {
			// Bisa diubah sesuai format tanggal yang muncul
			return strings.Contains(s.Text(), "202") // cari tahun
		}).Text())

		e.ForEach(".mli-mvi a[rel=tag]", func(_ int, el *colly.HTMLElement) {
			detail.Genres = append(detail.Genres, cleanDetailText(el.Text))
		})

		detail.Synopsis = cleanDetailText(e.ChildText(".mli-desc p"))
	})

	// Trailer
	c.OnHTML("#pop-trailer", func(e *colly.HTMLElement) {
		detail.TrailerURL = e.ChildAttr("iframe", "src")
	})

	// Episode list
	c.OnHTML("div.tvseason div.les-content", func(e *colly.HTMLElement) {
		var episodes []Episode
		e.ForEach("a", func(_ int, el *colly.HTMLElement) {
			ep := Episode{
				Title: cleanDetailText(el.Text),
				URL:   el.Attr("href"),
			}
			episodes = append(episodes, ep)
		})
		// Urutan dibalik
		for i, j := 0, len(episodes)-1; i < j; i, j = i+1, j-1 {
			episodes[i], episodes[j] = episodes[j], episodes[i]
		}
		detail.Episodes = episodes
	})

	// Rekomendasi
	c.OnHTML("div.rekom .ml-item-rekom", func(e *colly.HTMLElement) {
		rec := Recommendation{
			Title:    cleanDetailText(e.ChildText(".judul")),
			URL:      e.ChildAttr("a.ml-mask", "href"),
			ImageURL: e.ChildAttr("img.mli-thumb", "src"),
			Rating:   cleanDetailText(e.ChildText(".mli-mvi")),
		}
		detail.Recommendations = append(detail.Recommendations, rec)
	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Mengunjungi:", r.URL.String())
	})

	// Jalankan scraping
	if err := c.Visit(startURL); err != nil {
		t.Fatalf("Failed to start scraping: %v", err)
	}
	c.Wait()

	// Fallback episodes jika kosong
	if len(detail.Episodes) == 0 {
		if strings.Contains(startURL, "/film/") {
			detail.Episodes = []Episode{{Title: "film", URL: startURL}}
		} else {
			detail.Episodes = []Episode{{Title: "Unknown", URL: startURL}}
		}
	}

	// Cetak hasil
	jsonData, err := json.MarshalIndent(detail, "", "  ")
	if err != nil {
		t.Fatalf("JSON marshaling failed: %v", err)
	}
	fmt.Println("\n--- SCRAPING DETAIL ANIME SELESAI ---")
	fmt.Println(string(jsonData))
}
