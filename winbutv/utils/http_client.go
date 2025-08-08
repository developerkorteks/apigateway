package utils

import (
	"log"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/nabilulilalbab/winbu.tv/config"
)

// CreateCollector creates a new colly collector with standard settings
func CreateCollector(cfg *config.Config) *colly.Collector {
	c := colly.NewCollector(
		colly.AllowedDomains("winbu.tv"),
		colly.UserAgent(cfg.UserAgent),
	)

	// Set timeout
	c.SetRequestTimeout(cfg.Timeout)

	// Add rate limiting
	c.Limit(&colly.LimitRule{
		DomainGlob:  "*winbu.tv*",
		Parallelism: 1,
		Delay:       cfg.RateLimit,
	})

	// Add debug if in development
	if cfg.Environment == "development" {
		c.OnRequest(func(r *colly.Request) {
			log.Printf("Visiting: %s", r.URL.String())
		})
	}

	return c
}

// CreateCollectorWithRetry creates a collector with retry logic
func CreateCollectorWithRetry(cfg *config.Config) *colly.Collector {
	c := CreateCollector(cfg)

	// Add retry logic
	retryCount := 0
	c.OnError(func(r *colly.Response, err error) {
		if retryCount < cfg.MaxRetries {
			retryCount++
			time.Sleep(time.Duration(retryCount) * time.Second) // Exponential backoff
			r.Request.Retry()
		}
	})

	return c
}
