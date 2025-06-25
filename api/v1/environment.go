package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pendeploy-simple/dto"
	"github.com/pendeploy-simple/models"
	"github.com/pendeploy-simple/services"
)

// EnvironmentController handles environment-related API endpoints
type EnvironmentController struct {
	environmentService *services.EnvironmentService
}

// NewEnvironmentController creates a new environment controller
func NewEnvironmentController() *EnvironmentController {
	return &EnvironmentController{
		environmentService: services.NewEnvironmentService(),
	}
}

// RegisterRoutes registers environment routes
func (c *EnvironmentController) RegisterRoutes(router *gin.RouterGroup) {
	environments := router.Group("/environments")
	{
		environments.GET("", c.ListEnvironments)
		environments.GET("/:id", c.GetEnvironment)
		environments.POST("", c.CreateEnvironment)
		environments.PUT("/:id", c.UpdateEnvironment)
		environments.DELETE("/:id", c.DeleteEnvironment)
	}

	// Also add project-specific environment routes
	projects := router.Group("/projects")
	{
		projects.GET("/:id/environments", c.ListProjectEnvironments)
	}
}

// ListEnvironments retrieves all environments (admin only)
func (c *EnvironmentController) ListEnvironments(ctx *gin.Context) {
	// Get userId and role from context
	userIDValue, _ := ctx.Get("userId")
	userID := userIDValue.(string)
	roleValue, _ := ctx.Get("role")
	role, _ := roleValue.(string)
	isAdmin := role == "admin"
	
	// Only admins can list all environments
	if !isAdmin {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
		return
	}
	
	// Parse project filter if provided
	projectID := ctx.Query("projectId")
	
	var environments []models.Environment
	var err error
	
	if projectID != "" {
		environments, err = c.environmentService.ListEnvironments(projectID, userID, isAdmin)
	} else {
		// Just return empty for now, could implement global listing later
		environments = []models.Environment{}
		err = nil
	}
	
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	// Convert to response DTOs
	var response dto.EnvironmentListResponse
	response.Environments = make([]dto.EnvironmentResponse, 0)
	
	for _, env := range environments {
		response.Environments = append(response.Environments, dto.EnvironmentResponse{
			ID:          env.ID,
			Name:        env.Name,
			Description: env.Description,
			ProjectID:   env.ProjectID,
			CreatedAt:   env.CreatedAt,
			UpdatedAt:   env.UpdatedAt,
		})
	}
	
	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   response,
	})
}

// ListProjectEnvironments retrieves all environments for a specific project
func (c *EnvironmentController) ListProjectEnvironments(ctx *gin.Context) {
	// Get userId and role from context
	userIDValue, _ := ctx.Get("userId")
	userID := userIDValue.(string)
	roleValue, _ := ctx.Get("role")
	role, _ := roleValue.(string)
	isAdmin := role == "admin"
	projectID := ctx.Param("id")
	
	environments, err := c.environmentService.ListEnvironments(projectID, userID, isAdmin)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	// Convert to response DTOs
	var response dto.EnvironmentListResponse
	response.Environments = make([]dto.EnvironmentResponse, 0)
	
	for _, env := range environments {
		response.Environments = append(response.Environments, dto.EnvironmentResponse{
			ID:          env.ID,
			Name:        env.Name,
			Description: env.Description,
			ProjectID:   env.ProjectID,
			CreatedAt:   env.CreatedAt,
			UpdatedAt:   env.UpdatedAt,
		})
	}
	
	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   response,
	})
}

// GetEnvironment retrieves a specific environment
func (c *EnvironmentController) GetEnvironment(ctx *gin.Context) {
	// Get userId and role from context
	userIDValue, _ := ctx.Get("userId")
	userID := userIDValue.(string)
	roleValue, _ := ctx.Get("role")
	role, _ := roleValue.(string)
	isAdmin := role == "admin"
	environmentID := ctx.Param("id")
	
	environment, err := c.environmentService.GetEnvironmentDetail(environmentID, userID, isAdmin)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	response := dto.EnvironmentResponse{
		ID:          environment.ID,
		Name:        environment.Name,
		Description: environment.Description,
		ProjectID:   environment.ProjectID,
		CreatedAt:   environment.CreatedAt,
		UpdatedAt:   environment.UpdatedAt,
	}
	
	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   response,
	})
}

// CreateEnvironment creates a new environment
func (c *EnvironmentController) CreateEnvironment(ctx *gin.Context) {
	// Get userId and role from context
	userIDValue, _ := ctx.Get("userId")
	userID := userIDValue.(string)
	roleValue, _ := ctx.Get("role")
	role, _ := roleValue.(string)
	isAdmin := role == "admin"
	
	var request dto.EnvironmentRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// Create environment model
	environment := models.Environment{
		Name:        request.Name,
		Description: request.Description,
		ProjectID:   request.ProjectID,
	}
	
	// Call service to create
	createdEnv, err := c.environmentService.CreateEnvironment(environment, userID, isAdmin)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	// Return created environment
	response := dto.EnvironmentResponse{
		ID:          createdEnv.ID,
		Name:        createdEnv.Name,
		Description: createdEnv.Description,
		ProjectID:   createdEnv.ProjectID,
		CreatedAt:   createdEnv.CreatedAt,
		UpdatedAt:   createdEnv.UpdatedAt,
	}
	
	ctx.JSON(http.StatusCreated, gin.H{
		"status": "success",
		"data":   response,
	})
}

// UpdateEnvironment updates an existing environment
func (c *EnvironmentController) UpdateEnvironment(ctx *gin.Context) {
	// Get userId and role from context
	userIDValue, _ := ctx.Get("userId")
	userID := userIDValue.(string)
	roleValue, _ := ctx.Get("role")
	role, _ := roleValue.(string)
	isAdmin := role == "admin"
	environmentID := ctx.Param("id")
	
	var request dto.EnvironmentRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// Create environment model
	environment := models.Environment{
		ID:          environmentID,
		Name:        request.Name,
		Description: request.Description,
		// No need to set ProjectID as it cannot be changed after creation
	}
	
	// Call service to update
	updatedEnv, err := c.environmentService.UpdateEnvironment(environment, userID, isAdmin)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	// Return updated environment
	response := dto.EnvironmentResponse{
		ID:          updatedEnv.ID,
		Name:        updatedEnv.Name,
		Description: updatedEnv.Description,
		ProjectID:   updatedEnv.ProjectID,
		CreatedAt:   updatedEnv.CreatedAt,
		UpdatedAt:   updatedEnv.UpdatedAt,
	}
	
	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   response,
	})
}

// DeleteEnvironment deletes an environment
func (c *EnvironmentController) DeleteEnvironment(ctx *gin.Context) {
	// Get userId and role from context
	userIDValue, _ := ctx.Get("userId")
	userID := userIDValue.(string)
	roleValue, _ := ctx.Get("role")
	role, _ := roleValue.(string)
	isAdmin := role == "admin"
	environmentID := ctx.Param("id")
	
	err := c.environmentService.DeleteEnvironment(environmentID, userID, isAdmin)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
		"message": "Environment deleted successfully",
	})
}
