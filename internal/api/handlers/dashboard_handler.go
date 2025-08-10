package handlers

import (
	"apicategorywithfallback/internal/service"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type DashboardHandler struct {
	apiService *service.APIService
}

func NewDashboardHandler(apiService *service.APIService) *DashboardHandler {
	return &DashboardHandler{
		apiService: apiService,
	}
}

// ShowDashboard renders the main dashboard page
func (h *DashboardHandler) ShowDashboard(c *gin.Context) {
	c.HTML(http.StatusOK, "dashboard.html", gin.H{
		"title": "API Fallback Dashboard",
	})
}

// ShowEnhancedDashboard renders the enhanced dashboard page
func (h *DashboardHandler) ShowEnhancedDashboard(c *gin.Context) {
	c.HTML(http.StatusOK, "dashboard_improved.html", gin.H{
		"title": "API Fallback System",
	})
}

// ShowManagement renders the management dashboard page
func (h *DashboardHandler) ShowManagement(c *gin.Context) {
	c.HTML(http.StatusOK, "dashboard_management.html", gin.H{
		"title": "API Fallback Management",
	})
}

// GetHealthStatus returns the health status of all API sources
func (h *DashboardHandler) GetHealthStatus(c *gin.Context) {
	healthStatus, err := h.apiService.GetHealthStatus()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get health status",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   healthStatus,
	})
}

// GetRequestLogs returns recent request logs
func (h *DashboardHandler) GetRequestLogs(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "100")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 100
	}

	logs, err := h.apiService.GetRequestLogs(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get request logs",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   logs,
	})
}

// GetStatistics returns API usage statistics
func (h *DashboardHandler) GetStatistics(c *gin.Context) {
	stats, err := h.apiService.GetStatistics()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get statistics",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   stats,
	})
}

// GetCategories returns all categories
func (h *DashboardHandler) GetCategories(c *gin.Context) {
	categories, err := h.apiService.GetCategories()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get categories",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   categories,
	})
}

// GetCategoryNames returns category names for dynamic Swagger documentation
// @Summary Get available category names
// @Description Returns list of all active category names for dynamic API documentation
// @Tags System
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "List of available category names"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/categories/names [get]
func (h *DashboardHandler) GetCategoryNames(c *gin.Context) {
	categoryNames, err := h.apiService.GetCategoryNames()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get category names",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   categoryNames,
	})
}

// CreateCategory creates a new category
func (h *DashboardHandler) CreateCategory(c *gin.Context) {
	var req struct {
		Name     string `json:"name" binding:"required"`
		IsActive bool   `json:"is_active"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	err := h.apiService.CreateCategory(req.Name, req.IsActive)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create category",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status":  "success",
		"message": "Category created successfully",
		"data": gin.H{
			"name":      req.Name,
			"is_active": req.IsActive,
		},
	})
}

// UpdateCategory updates an existing category
func (h *DashboardHandler) UpdateCategory(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid category ID",
		})
		return
	}

	var req struct {
		Name     string `json:"name" binding:"required"`
		IsActive bool   `json:"is_active"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	err = h.apiService.UpdateCategory(id, req.Name, req.IsActive)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update category",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Category updated successfully",
		"data": gin.H{
			"id":        id,
			"name":      req.Name,
			"is_active": req.IsActive,
		},
	})
}

// DeleteCategory deletes a category
func (h *DashboardHandler) DeleteCategory(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid category ID",
		})
		return
	}

	err = h.apiService.DeleteCategory(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to delete category",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Category deleted successfully",
		"data": gin.H{
			"id": id,
		},
	})
}

// GetAPISources returns all API sources
func (h *DashboardHandler) GetAPISources(c *gin.Context) {
	sources, err := h.apiService.GetAllAPISources()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get API sources",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   sources,
	})
}

// CreateAPISource creates a new API source
func (h *DashboardHandler) CreateAPISource(c *gin.Context) {
	var req struct {
		EndpointID int    `json:"endpoint_id" binding:"required"`
		SourceName string `json:"source_name" binding:"required"`
		BaseURL    string `json:"base_url" binding:"required"`
		Priority   int    `json:"priority" binding:"required"`
		IsPrimary  bool   `json:"is_primary"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	err := h.apiService.CreateAPISource(req.EndpointID, req.SourceName, req.BaseURL, req.Priority, req.IsPrimary)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create API source",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status":  "success",
		"message": "API source created successfully",
		"data":    req,
	})
}

// CreateAPISourceForAllEndpoints creates a new API source for all endpoints in a category
func (h *DashboardHandler) CreateAPISourceForAllEndpoints(c *gin.Context) {
	var req struct {
		CategoryName string `json:"category_name" binding:"required"`
		SourceName   string `json:"source_name" binding:"required"`
		BaseURL      string `json:"base_url" binding:"required"`
		Priority     int    `json:"priority" binding:"required"`
		IsPrimary    bool   `json:"is_primary"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	err := h.apiService.CreateAPISourceForAllEndpoints(req.CategoryName, req.SourceName, req.BaseURL, req.Priority, req.IsPrimary)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create API source for all endpoints",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status":  "success",
		"message": fmt.Sprintf("API source '%s' created successfully for all endpoints in category '%s'", req.SourceName, req.CategoryName),
		"data":    req,
	})
}

// UpdateAPISource updates an existing API source
func (h *DashboardHandler) UpdateAPISource(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid API source ID",
		})
		return
	}

	var req struct {
		SourceName string `json:"source_name" binding:"required"`
		BaseURL    string `json:"base_url" binding:"required"`
		Priority   int    `json:"priority" binding:"required"`
		IsPrimary  bool   `json:"is_primary"`
		IsActive   bool   `json:"is_active"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	err = h.apiService.UpdateAPISource(id, req.SourceName, req.BaseURL, req.Priority, req.IsPrimary, req.IsActive)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update API source",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "API source updated successfully",
		"data": gin.H{
			"id":          id,
			"source_name": req.SourceName,
			"base_url":    req.BaseURL,
			"priority":    req.Priority,
			"is_primary":  req.IsPrimary,
			"is_active":   req.IsActive,
		},
	})
}

// DeleteAPISource deletes an API source
func (h *DashboardHandler) DeleteAPISource(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid API source ID",
		})
		return
	}

	err = h.apiService.DeleteAPISource(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to delete API source",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "API source deleted successfully",
		"data": gin.H{
			"id": id,
		},
	})
}

// RunManualHealthCheck performs manual health check on all API sources
func (h *DashboardHandler) RunManualHealthCheck(c *gin.Context) {
	results, err := h.apiService.RunManualHealthCheck()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to run manual health check",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   results,
	})
}

// GetEndpoints returns all endpoints
func (h *DashboardHandler) GetEndpoints(c *gin.Context) {
	endpoints, err := h.apiService.GetAllEndpoints()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get endpoints",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   endpoints,
	})
}

// CreateEndpoint creates a new endpoint
func (h *DashboardHandler) CreateEndpoint(c *gin.Context) {
	var req struct {
		CategoryID int    `json:"category_id" binding:"required"`
		Path       string `json:"path" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	endpoint, err := h.apiService.CreateEndpoint(req.CategoryID, req.Path)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create endpoint",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status":  "success",
		"message": "Endpoint created successfully",
		"data":    endpoint,
	})
}

// UpdateEndpoint updates an existing endpoint
func (h *DashboardHandler) UpdateEndpoint(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid endpoint ID",
		})
		return
	}

	var req struct {
		CategoryID int    `json:"category_id"`
		Path       string `json:"path"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	endpoint, err := h.apiService.UpdateEndpoint(id, req.CategoryID, req.Path)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update endpoint",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Endpoint updated successfully",
		"data":    endpoint,
	})
}

// DeleteEndpoint deletes an endpoint
func (h *DashboardHandler) DeleteEndpoint(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid endpoint ID",
		})
		return
	}

	err = h.apiService.DeleteEndpoint(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to delete endpoint",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Endpoint deleted successfully",
		"data": gin.H{
			"id": id,
		},
	})
}

// DeleteAPISourceByName deletes all API sources with the given source name
func (h *DashboardHandler) DeleteAPISourceByName(c *gin.Context) {
	var req struct {
		SourceName string `json:"source_name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// First, get all API sources with this name to show what will be deleted
	sources, err := h.apiService.GetAPISourcesByName(req.SourceName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get API sources",
			"details": err.Error(),
		})
		return
	}

	if len(sources) == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "No API sources found with the given name",
			"details": fmt.Sprintf("Source name: %s", req.SourceName),
		})
		return
	}

	// Delete all API sources with this name
	err = h.apiService.DeleteAPISourceByName(req.SourceName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to delete API sources",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": fmt.Sprintf("Successfully deleted %d API sources with name '%s'", len(sources), req.SourceName),
		"data": gin.H{
			"source_name":     req.SourceName,
			"deleted_count":   len(sources),
			"deleted_sources": sources,
		},
	})
}

// GetAPISourcesByName returns all API sources with the given source name
func (h *DashboardHandler) GetAPISourcesByName(c *gin.Context) {
	sourceName := c.Query("source_name")
	if sourceName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "source_name query parameter is required",
		})
		return
	}

	sources, err := h.apiService.GetAPISourcesByName(sourceName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get API sources",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   sources,
		"count":  len(sources),
	})
}
