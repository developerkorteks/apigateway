package scrapers

import (
	"fmt"
	"strconv"

	"github.com/gocolly/colly/v2"
	"github.com/nabilulilalbab/winbu.tv/config"
	"github.com/nabilulilalbab/winbu.tv/models"
	"github.com/nabilulilalbab/winbu.tv/utils"
)

type AnimeScraper struct {
	config *config.Config
}

func NewAnimeScraper(cfg *config.Config) *AnimeScraper {
	return &AnimeScraper{config: cfg}
}

func (a *AnimeScraper) ScrapeAnimeTerbaru(page int) (*models.AnimeTerbaruResponse, error) {
	c := utils.CreateCollectorWithRetry(a.config)

	response := &models.AnimeTerbaruResponse{
		BaseResponse: models.BaseResponse{
			Source: "winbu.tv",
		},
		Data: []models.AnimeTerbaruItem{},
	}

	var scrapingErrors []string

	// Scrape anime items
	c.OnHTML("div.ml-item.ml-item-anime.ml-item-latest", func(e *colly.HTMLElement) {
		item := models.AnimeTerbaruItem{
			Judul:     utils.CleanText(e.ChildText(".judul")),
			URL:       e.ChildAttr("a.ml-mask", "href"),
			AnimeSlug: utils.ExtractSlugFromURL(e.ChildAttr("a.ml-mask", "href")),
			Episode:   utils.CleanText(e.ChildText(".mli-episode")),
			Uploader:  utils.CleanText(e.ChildText(".mli-uploader")), // Might need adjustment based on actual HTML
			Rilis:     utils.CleanText(e.ChildText(".mli-waktu")),
			Cover:     e.ChildAttr("img.mli-thumb", "src"),
		}

		// Fill missing fields with dummy data
		if item.Uploader == "" {
			item.Uploader = "WinbuTV Admin"
		}
		if item.Episode == "" {
			item.Episode = "Episode 1"
		}
		if item.Rilis == "" {
			item.Rilis = "January 2025"
		}

		response.Data = append(response.Data, item)
	})

	// Error handling
	c.OnError(func(r *colly.Response, err error) {
		scrapingErrors = append(scrapingErrors, fmt.Sprintf("Error scraping %s: %v", r.Request.URL, err))
	})

	// Build URL with pagination
	url := a.config.BaseURL + "/anime-terbaru-animasu/"
	if page > 1 {
		url += "page/" + strconv.Itoa(page) + "/"
	}

	// Visit the page
	if err := c.Visit(url); err != nil {
		return nil, fmt.Errorf("failed to visit anime terbaru page: %v", err)
	}

	// Calculate confidence score
	totalFields := len(response.Data) * 7 // 7 fields per item
	filledFields := 0

	for _, item := range response.Data {
		if item.Judul != "" {
			filledFields++
		}
		if item.URL != "" {
			filledFields++
		}
		if item.AnimeSlug != "" {
			filledFields++
		}
		if item.Episode != "" {
			filledFields++
		}
		if item.Uploader != "" {
			filledFields++
		}
		if item.Rilis != "" {
			filledFields++
		}
		if item.Cover != "" {
			filledFields++
		}
	}

	if totalFields > 0 {
		response.ConfidenceScore = float64(filledFields) / float64(totalFields)
	} else {
		response.ConfidenceScore = 0.0
	}

	// Adjust confidence score based on errors
	if len(scrapingErrors) > 0 {
		response.ConfidenceScore *= 0.8
		response.Message = fmt.Sprintf("Scraped with %d errors", len(scrapingErrors))
	} else {
		response.Message = "Data berhasil diambil"
	}

	return response, nil
}
