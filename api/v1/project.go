package v1

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pendeploy-simple/dto"
	"github.com/pendeploy-simple/models"
	"github.com/pendeploy-simple/services"
)

var projectService = services.NewProjectService()

// ListProjects godoc
// @Summary List projects with pagination and filtering
// @Description Get all projects for admin, or only user's projects for regular users
// @Tags projects
// @Accept json
// @Produce json
// @Param page query int false "Page number"
// @Param pageSize query int false "Page size"
// @Param search query string false "Search term for project name/description"
// @Param sortBy query string false "Field to sort by (created_at, updated_at, name)"
// @Param sortOrder query string false "Sort order (asc or desc)"
// @Success 200 {object} dto.ProjectListResponse
// @Router /projects [get]
func ListProjects(c *gin.Context) {
	// Get user info from context (set by AuthMiddleware)
	userID, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"status": "error", "message": "User not authenticated"})
		return
	}

	// Check if user is admin
	role, _ := c.Get("role")
	isAdmin := role == "admin"

	// Parse query parameters
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	if err != nil || pageSize < 1 {
		pageSize = 10
	}

	// Build filter
	filter := dto.ProjectFilter{
		UserID:    userID.(string),
		Search:    c.Query("search"),
		SortBy:    c.DefaultQuery("sortBy", "created_at"),
		SortOrder: c.DefaultQuery("sortOrder", "desc"),
		Page:      page,
		PageSize:  pageSize,
		IsAdmin:   isAdmin,
	}

	// Call service
	response, err := projectService.ListProjects(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to retrieve projects: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   response,
	})
}

// GetProject godoc
// @Summary Get a project by ID
// @Description Get details of a project by ID
// @Tags projects
// @Accept json
// @Produce json
// @Param id path string true "Project ID"
// @Success 200 {object} models.Project
// @Router /projects/{id} [get]
func GetProject(c *gin.Context) {
	// Get user info from context
	userID, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"status": "error", "message": "User not authenticated"})
		return
	}

	// Check if user is admin
	role, _ := c.Get("role")
	isAdmin := role == "admin"

	// Get project ID from URL
	projectID := c.Param("id")
	if projectID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Project ID is required"})
		return
	}

	// Get project with services
	project, err := projectService.GetProjectDetail(projectID, userID.(string), isAdmin)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "error",
			"message": "Project not found or access denied: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   project,
	})
}

// GetProjectStats godoc
// @Summary Get project statistics
// @Description Get statistics and dashboard data for a project
// @Tags projects
// @Accept json
// @Produce json
// @Param id path string true "Project ID"
// @Success 200 {object} dto.ProjectStatsResponse
// @Router /projects/{id}/stats [get]
func GetProjectStats(c *gin.Context) {
	// Get user info from context
	userID, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"status": "error", "message": "User not authenticated"})
		return
	}

	// Check if user is admin
	role, _ := c.Get("role")
	isAdmin := role == "admin"

	// Get project ID from URL
	projectID := c.Param("id")
	if projectID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Project ID is required"})
		return
	}

	// Get project stats
	stats, err := projectService.GetProjectStats(projectID, userID.(string), isAdmin)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "error",
			"message": "Failed to get project statistics: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   stats,
	})
}

// CreateProject godoc
// @Summary Create a new project
// @Description Create a new project for the authenticated user
// @Tags projects
// @Accept json
// @Produce json
// @Param project body dto.CreateProjectRequest true "Project Data"
// @Success 201 {object} dto.ProjectResponse
// @Router /projects [post]
func CreateProject(c *gin.Context) {
	// Get user info from context
	userID, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"status": "error", "message": "User not authenticated"})
		return
	}

	// Parse request body to DTO first
	var projectDTO dto.CreateProjectRequest
	if err := c.ShouldBindJSON(&projectDTO); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid request data: " + err.Error(),
		})
		return
	}
	
	// Map DTO to model
	now := time.Now()
	project := models.Project{
		Name:        projectDTO.Name,
		Description: projectDTO.Description,
		UserID:      userID.(string),
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// Create project
	newProject, err := projectService.CreateProject(project)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to create project: " + err.Error(),
		})
		return
	}

	// Map model to response DTO
	response := dto.ProjectResponse{
		ID:          newProject.ID,
		Name:        newProject.Name,
		Description: newProject.Description,
		UserID:      newProject.UserID,
		CreatedAt:   newProject.CreatedAt,
		UpdatedAt:   newProject.UpdatedAt,
	}
	
	c.JSON(http.StatusCreated, gin.H{
		"status": "success",
		"data":   response,
	})
}

// UpdateProject godoc
// @Summary Update an existing project
// @Description Update project details
// @Tags projects
// @Accept json
// @Produce json
// @Param id path string true "Project ID"
// @Param project body dto.UpdateProjectRequest true "Project Data"
// @Success 200 {object} dto.ProjectResponse
// @Router /projects/{id} [put]
func UpdateProject(c *gin.Context) {
	// Get user info from context
	userID, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"status": "error", "message": "User not authenticated"})
		return
	}

	// Get project ID from URL
	projectID := c.Param("id")
	if projectID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Project ID is required"})
		return
	}

	// Parse request body to DTO
	var projectDTO dto.UpdateProjectRequest
	if err := c.ShouldBindJSON(&projectDTO); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid request data: " + err.Error(),
		})
		return
	}

	// Check if user is admin
	role, _ := c.Get("role")
	isAdmin := role == "admin"

	// Map DTO to model changes - only updating specific fields
	projectChanges := models.Project{
		ID:          projectID, // Set ID yang akan diupdate
		Name:        projectDTO.Name,
		Description: projectDTO.Description,
	}

	// Find and update project dengan parameter yang benar
	updatedProject, err := projectService.UpdateProject(projectChanges, userID.(string), isAdmin)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to update project: " + err.Error(),
		})
		return
	}

	// Map model to response DTO
	response := dto.ProjectResponse{
		ID:          updatedProject.ID,
		Name:        updatedProject.Name,
		Description: updatedProject.Description,
		UserID:      updatedProject.UserID,
		CreatedAt:   updatedProject.CreatedAt,
		UpdatedAt:   updatedProject.UpdatedAt,
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   response,
	})
}

// DeleteProject godoc
// @Summary Delete a project
// @Description Delete an existing project
// @Tags projects
// @Accept json
// @Produce json
// @Param id path string true "Project ID"
// @Success 200 {object} map[string]interface{}
// @Router /projects/{id} [delete]
func DeleteProject(c *gin.Context) {
	// Get user info from context
	userID, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"status": "error", "message": "User not authenticated"})
		return
	}

	// Check if user is admin
	role, _ := c.Get("role")
	isAdmin := role == "admin"

	// Get project ID from URL
	projectID := c.Param("id")
	if projectID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Project ID is required"})
		return
	}

	// Delete project
	err := projectService.DeleteProject(projectID, userID.(string), isAdmin)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to delete project: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Project deleted successfully",
	})
}
