package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	v1 "github.com/nabilulilalbab/winbu.tv/api/v1"
	"github.com/nabilulilalbab/winbu.tv/config"
	"github.com/nabilulilalbab/winbu.tv/docs"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title Winbu.TV Web Scraping API
// @version 1.0
// @description API web scraping untuk mengambil data dari situs Winbu.TV
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8002
// @BasePath /
// @schemes http https

func main() {
	// Load configuration
	cfg := config.Load()

	// Set Gin mode
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create Gin router
	r := gin.Default()

	// Add CORS middleware
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"message": "API is running",
		})
	})

	// Custom swagger.json endpoint with dynamic host
	r.GET("/api/swagger.json", func(c *gin.Context) {
		// Get current request info
		host := c.Request.Host
		scheme := "http"

		// Check for HTTPS
		if c.Request.TLS != nil ||
			c.GetHeader("X-Forwarded-Proto") == "https" ||
			c.GetHeader("X-Forwarded-Ssl") == "on" ||
			c.GetHeader("X-Url-Scheme") == "https" {
			scheme = "https"
		}

		// Get original swagger JSON
		doc := docs.SwaggerInfo.ReadDoc()

		// Replace host and schemes dynamically (using correct patterns that match the generated JSON)
		doc = strings.Replace(doc, `"host": "localhost:8002"`, `"host": "`+host+`"`, 1)
		doc = strings.Replace(doc, `"schemes": ["http","https"]`, `"schemes": ["`+scheme+`"]`, 1)

		c.Header("Content-Type", "application/json")
		c.String(200, doc)
	})

	// Dynamic Swagger documentation
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, ginSwagger.URL("/api/swagger.json")))

	// API v1 routes
	v1Group := r.Group("/api/v1")
	v1.SetupRoutes(v1Group)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = cfg.Port
	}

	log.Printf("Starting server on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
