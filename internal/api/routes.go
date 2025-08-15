package api

import (
	"apicategorywithfallback/internal/api/handlers"
	"apicategorywithfallback/internal/service"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(router *gin.Engine, apiService *service.APIService) {
	// Add CORS middleware
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Initialize handlers
	apiHandler := handlers.NewAPIHandler(apiService)
	dashboardHandler := handlers.NewDashboardHandler(apiService)
	swaggerHandler := handlers.NewSwaggerHandler(apiService)

	// API routes
	v1 := router.Group("/api/v1")
	{
		// Handle both with and without trailing slashes
		v1.GET("/home", apiHandler.HandleHome)
		v1.GET("/home/", apiHandler.HandleHome)
		v1.GET("/jadwal-rilis", apiHandler.HandleJadwalRilis)
		v1.GET("/jadwal-rilis/", apiHandler.HandleJadwalRilis)
		v1.GET("/jadwal-rilis/:day", apiHandler.HandleJadwalRilisDay)
		v1.GET("/anime-terbaru", apiHandler.HandleAnimeTerbaru)
		v1.GET("/anime-terbaru/", apiHandler.HandleAnimeTerbaru)
		v1.GET("/movie", apiHandler.HandleMovie)
		v1.GET("/movie/", apiHandler.HandleMovie)
		v1.GET("/anime-detail", apiHandler.HandleAnimeDetail)
		v1.GET("/anime-detail/", apiHandler.HandleAnimeDetail)
		v1.GET("/episode-detail", apiHandler.HandleEpisodeDetail)
		v1.GET("/episode-detail/", apiHandler.HandleEpisodeDetail)
		v1.GET("/search", apiHandler.HandleSearch)
		v1.GET("/search/", apiHandler.HandleSearch)
	}

	// Dashboard routes
	dashboard := router.Group("/dashboard")
	{
		dashboard.GET("/", dashboardHandler.ShowDashboard)
		dashboard.GET("/enhanced", dashboardHandler.ShowEnhancedDashboard)
		dashboard.GET("/management", dashboardHandler.ShowManagement)
		dashboard.GET("/health", dashboardHandler.GetHealthStatus)
		dashboard.POST("/health/check", dashboardHandler.RunManualHealthCheck)
		dashboard.GET("/logs", dashboardHandler.GetRequestLogs)
		dashboard.GET("/stats", dashboardHandler.GetStatistics)

		// API management routes
		dashboard.GET("/categories", dashboardHandler.GetCategories)
		dashboard.POST("/categories", dashboardHandler.CreateCategory)
		dashboard.PUT("/categories/:id", dashboardHandler.UpdateCategory)
		dashboard.DELETE("/categories/:id", dashboardHandler.DeleteCategory)

		// Endpoints management routes
		dashboard.GET("/endpoints", dashboardHandler.GetEndpoints)
		dashboard.POST("/endpoints", dashboardHandler.CreateEndpoint)
		dashboard.PUT("/endpoints/:id", dashboardHandler.UpdateEndpoint)
		dashboard.DELETE("/endpoints/:id", dashboardHandler.DeleteEndpoint)

		// API Sources management routes
		dashboard.GET("/api-sources", dashboardHandler.GetAPISources)
		dashboard.POST("/api-sources", dashboardHandler.CreateAPISource)
		dashboard.POST("/api-sources/bulk", dashboardHandler.CreateAPISourceForAllEndpoints)
		dashboard.PUT("/api-sources/:id", dashboardHandler.UpdateAPISource)
		dashboard.DELETE("/api-sources/:id", dashboardHandler.DeleteAPISource)

		// Bulk API Sources management
		dashboard.GET("/api-sources/by-name", dashboardHandler.GetAPISourcesByName)
		dashboard.DELETE("/api-sources/by-name", dashboardHandler.DeleteAPISourceByName)

		// Cache management routes
		dashboard.DELETE("/cache/clear", apiHandler.HandleClearCache)
	}

	// Public API routes for system information
	api := router.Group("/api")
	{
		api.GET("/categories/names", dashboardHandler.GetCategoryNames)
	}

	// Custom Swagger UI with dynamic categories
	router.GET("/swagger-ui", swaggerHandler.ServeSwaggerUI)
	router.GET("/swagger-ui/", swaggerHandler.ServeSwaggerUI)

	// Health check endpoint
	router.GET("/health", apiHandler.HandleHealthCheck)

	// Serve static files for dashboard
	router.Static("/static", "./web/static")
	router.LoadHTMLGlob("web/templates/*")
}
