package main

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

func main() {
	// Test endpoints untuk bruteforce anime detail
	testEndpoints := []struct {
		name string
		url  string
	}{
		{
			name: "Anime Detail - Naruto",
			url:  "http://localhost:8080/api/v1/anime-detail?anime_slug=naruto-shippuden&category=anime",
		},
		{
			name: "Anime Detail - One Piece",
			url:  "http://localhost:8080/api/v1/anime-detail?anime_slug=one-piece&category=anime",
		},
		{
			name: "Episode Detail",
			url:  "http://localhost:8080/api/v1/episode-detail?episode_url=https://winbu.tv/anime/naruto-episode-1&category=anime",
		},
	}

	fmt.Println("🚀 Starting Bruteforce API Testing...")
	fmt.Println(strings.Repeat("=", 50))

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	for _, test := range testEndpoints {
		fmt.Printf("\n🧪 Testing: %s\n", test.name)
		fmt.Printf("📡 URL: %s\n", test.url)

		start := time.Now()

		resp, err := client.Get(test.url)
		if err != nil {
			fmt.Printf("❌ Error: %v\n", err)
			continue
		}
		defer resp.Body.Close()

		duration := time.Since(start)

		// Check response headers for source info
		source := resp.Header.Get("X-Source")
		responseTime := resp.Header.Get("X-Response-Time")
		cacheStatus := resp.Header.Get("X-Cache")

		fmt.Printf("✅ Status: %s (%d)\n", resp.Status, resp.StatusCode)
		fmt.Printf("⏱️  Response Time: %v\n", duration)
		if source != "" {
			fmt.Printf("🔗 Source: %s\n", source)
		}
		if responseTime != "" {
			fmt.Printf("⚡ API Response Time: %s\n", responseTime)
		}
		if cacheStatus != "" {
			fmt.Printf("💾 Cache: %s\n", cacheStatus)
		}

		// Read response body (first 500 chars)
		if resp.StatusCode == 200 {
			buf := make([]byte, 500)
			n, _ := resp.Body.Read(buf)
			fmt.Printf("📄 Response Preview: %s...\n", string(buf[:n]))
		}

		fmt.Println(strings.Repeat("─", 40))
	}

	fmt.Println("\n✨ Testing completed!")
	fmt.Println("\n💡 Key Features of Bruteforce Implementation:")
	fmt.Println("   • Parallel requests to ALL configured API sources")
	fmt.Println("   • Returns first valid response (based on priority)")
	fmt.Println("   • Automatic fallback to alternative URLs")
	fmt.Println("   • Data validation before response")
	fmt.Println("   • Comprehensive logging and monitoring")
}
