package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pendeploy-simple/dto"
	"github.com/pendeploy-simple/services"
)

// DeploymentController handles HTTP requests for deployments
type DeploymentController struct {
	deploymentService *services.DeploymentService
}

// NewDeploymentController creates a new DeploymentController
func NewDeploymentController() *DeploymentController {
	return &DeploymentController{
		deploymentService: services.NewDeploymentService(),
	}
}

// RegisterRoutes registers routes for the DeploymentController
func (c *DeploymentController) RegisterRoutes(router *gin.RouterGroup) {
	deployGroup := router.Group("/deployments")
	{
		deployGroup.POST("/git", c.CreateDeployment)
		deployGroup.GET("/:id", c.GetDeployment)
		deployGroup.GET("/:id/logs/build", c.StreamBuildLogs)
		deployGroup.GET("/:id/logs/runtime", c.StreamRuntimeLogs)
	}
}

// CreateDeployment handles POST /api/deployments/git
// Creates a new Kubernetes job for building and deploying a Git repository
func (c *DeploymentController) CreateDeployment(ctx *gin.Context) {
	var request dto.GitDeployRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := c.deploymentService.CreateGitDeployment(request)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, response)
}

// GetDeployment handles GET /api/deployments/:id
// Gets status of a deployment
func (c *DeploymentController) GetDeployment(ctx *gin.Context) {
	id := ctx.Param("id")
	
	deployment, err := c.deploymentService.GetDeploymentByID(id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Deployment not found"})
		return
	}
	
	// Get Kubernetes resource status if available
	resourceStatus, err := c.deploymentService.GetResourceStatus(deployment.ServiceID)
	
	// For response, include both deployment information and Kubernetes resource status if available
	response := gin.H{
		"deployment": deployment,
	}
	
	if resourceStatus != nil && err == nil {
		response["resources"] = resourceStatus
	}
	
	ctx.JSON(http.StatusOK, response)
}

// StreamBuildLogs handles GET /api/deployments/:id/logs/build
// Streams build logs from Kubernetes job in Server-Sent Events format
func (c *DeploymentController) StreamBuildLogs(ctx *gin.Context) {
	id := ctx.Param("id")

	// Set headers for SSE streaming
	ctx.Writer.Header().Set("Content-Type", "text/event-stream")
	ctx.Writer.Header().Set("Cache-Control", "no-cache")
	ctx.Writer.Header().Set("Connection", "keep-alive")
	ctx.Writer.Header().Set("Transfer-Encoding", "chunked")
	ctx.Writer.Header().Set("X-Accel-Buffering", "no") // Prevent Nginx from buffering the response

	// Get the deployment by ID
	deployment, err := c.deploymentService.GetDeploymentByID(id)
	if err != nil {
		ctx.Writer.Write([]byte("data: {\"error\": \"Deployment not found\"}\n\n"))
		return
	}

	// Stream build logs
	err = c.deploymentService.GetServiceBuildLogsRealtime(deployment.ID, ctx.Writer)
	if err != nil {
		// Don't send error as JSON as we've already started streaming
		ctx.Writer.Write([]byte("data: {\"error\": \"" + err.Error() + "\"}\n\n"))
	}
}

// StreamRuntimeLogs handles GET /api/deployments/:id/logs/runtime
// Streams deployment logs from Kubernetes pods in Server-Sent Events format
func (c *DeploymentController) StreamRuntimeLogs(ctx *gin.Context) {
	id := ctx.Param("id")

	// Set headers for SSE streaming
	ctx.Writer.Header().Set("Content-Type", "text/event-stream")
	ctx.Writer.Header().Set("Cache-Control", "no-cache")
	ctx.Writer.Header().Set("Connection", "keep-alive")
	ctx.Writer.Header().Set("Transfer-Encoding", "chunked")
	ctx.Writer.Header().Set("X-Accel-Buffering", "no") // Prevent Nginx from buffering the response

	// Get the deployment by ID
	deployment, err := c.deploymentService.GetDeploymentByID(id)
	if err != nil {
		ctx.Writer.Write([]byte("data: {\"error\": \"Deployment not found\"}\n\n"))
		return
	}

	// Stream runtime logs from the service's pods
	err = c.deploymentService.GetServiceRuntimeLogsRealtime(deployment.ServiceID, ctx.Writer)
	if err != nil {
		// Don't send error as JSON as we've already started streaming
		ctx.Writer.Write([]byte("data: {\"error\": \"" + err.Error() + "\"}\n\n"))
	}
}
