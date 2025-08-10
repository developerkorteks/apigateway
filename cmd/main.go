package main

import (
	"apicategorywithfallback/internal/api"
	"apicategorywithfallback/internal/service"
	"apicategorywithfallback/pkg/config"
	"apicategorywithfallback/pkg/database"
	"apicategorywithfallback/pkg/logger"
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "apicategorywithfallback/docs" // Import generated docs
)

// @title API Fallback Service
// @version 1.0
// @description A robust API fallback service for anime content with automatic failover capabilities
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @BasePath /
// @schemes http https

// @tag.name Anime
// @tag.description Anime related endpoints
// @tag.name Movie
// @tag.description Movie related endpoints
// @tag.name Search
// @tag.description Search functionality
// @tag.name Health
// @tag.description Health check endpoints

func main() {
	// Initialize logger
	logger.Init()

	// Load configuration
	cfg := config.Load()

	// Initialize database
	db, err := database.Init(cfg.DatabasePath)
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer db.Close()

	// Initialize services
	apiService := service.NewAPIService(db, cfg)

	// Start background health checker
	go apiService.StartHealthChecker()

	// Initialize router
	router := gin.Default()

	// Load HTML templates
	router.LoadHTMLGlob("web/templates/*")

	// Setup API routes
	api.SetupRoutes(router, apiService)

	// Setup Swagger
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Create HTTP server
	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		logger.Info("Starting server on port " + cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutting down server...")

	// Give outstanding requests a deadline for completion
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	logger.Info("Server exited")
}
