package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/pendeploy-simple/dto"
	"github.com/pendeploy-simple/services"
)

// RegistryController handles HTTP requests for registries
type RegistryController struct {
	registryService *services.RegistryService
}

// NewRegistryController creates a new registry controller instance
func NewRegistryController() *RegistryController {
	return &RegistryController{
		registryService: services.NewRegistryService(),
	}
}

// GetRegistries handles GET /api/registries
func (c *RegistryController) GetRegistries(ctx *gin.Context) {
	// Parse query parameters
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("pageSize", "10"))
	search := ctx.Query("search")
	sortBy := ctx.DefaultQuery("sortBy", "created_at")
	sortOrder := ctx.DefaultQuery("sortOrder", "desc")
	onlyActive := ctx.DefaultQuery("onlyActive", "false") == "true"

	// Create filter
	filter := dto.RegistryFilter{
		Page:       page,
		PageSize:   pageSize,
		Search:     search,
		SortBy:     sortBy,
		SortOrder:  sortOrder,
		OnlyActive: onlyActive,
	}

	// Get registries
	result, err := c.registryService.ListRegistries(filter)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, result)
}

// GetRegistry handles GET /api/registries/:id
func (c *RegistryController) GetRegistry(ctx *gin.Context) {
	id := ctx.Param("id")

	registry, err := c.registryService.GetRegistryByID(id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Registry not found"})
		return
	}

	ctx.JSON(http.StatusOK, registry)
}

// GetRegistryDetails handles GET /api/registries/:id/details
func (c *RegistryController) GetRegistryDetails(ctx *gin.Context) {
	id := ctx.Param("id")

	details, err := c.registryService.GetRegistryDetails(id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Registry not found"})
		return
	}

	ctx.JSON(http.StatusOK, details)
}

// CreateRegistry handles POST /api/registries
func (c *RegistryController) CreateRegistry(ctx *gin.Context) {
	var request dto.CreateRegistryRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	registry, err := c.registryService.CreateRegistry(request)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, registry)
}

// UpdateRegistry handles PUT /api/registries/:id
func (c *RegistryController) UpdateRegistry(ctx *gin.Context) {
	id := ctx.Param("id")

	var request dto.UpdateRegistryRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	registry, err := c.registryService.UpdateRegistry(id, request)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, registry)
}

// DeleteRegistry handles DELETE /api/registries/:id
func (c *RegistryController) DeleteRegistry(ctx *gin.Context) {
	id := ctx.Param("id")

	err := c.registryService.DeleteRegistry(id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Registry deleted successfully"})
}

// StreamBuildLogs handles GET /api/registries/:id/logs/stream
// This endpoint streams build logs from Kubernetes
func (c *RegistryController) StreamBuildLogs(ctx *gin.Context) {
	id := ctx.Param("id")

	// Set headers for streaming
	ctx.Writer.Header().Set("Content-Type", "text/event-stream")
	ctx.Writer.Header().Set("Cache-Control", "no-cache")
	ctx.Writer.Header().Set("Connection", "keep-alive")
	ctx.Writer.Header().Set("Transfer-Encoding", "chunked")

	// Stream logs
	err := c.registryService.StreamRegistryBuildLogs(ctx.Request.Context(), id, ctx.Writer)
	if err != nil {
		// Don't send error as JSON as we've already started streaming
		ctx.Writer.Write([]byte("Stream ended: " + err.Error()))
	}
}
