package controllers

import (
	"bytes"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/pendeploy-simple/dto"
	"github.com/pendeploy-simple/services"
	"github.com/pendeploy-simple/utils"
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
	
	// Read raw request body to handle potential newlines in commit messages
	requestBody, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body: " + err.Error()})
		return
	}
	
	// Smart escape: only escape literal newlines that aren't already escaped
	// This handles merge commit messages that contain literal \n characters
	bodyStr := string(requestBody)
	cleanedBody := smartEscapeJSON(bodyStr)
	
	// Reset the request body for JSON binding
	ctx.Request.Body = io.NopCloser(bytes.NewBufferString(cleanedBody))
	
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		// If there's a callbackUrl, notify of binding error
		if request.CallbackUrl != "" {
			go utils.SendErrorWebhook(request.CallbackUrl, "Invalid request format: " + err.Error())
		}
		return
	}

	response, err := c.deploymentService.CreateGitDeployment(request)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		// If there's a callbackUrl, notify of deployment creation error
		if request.CallbackUrl != "" {
			go utils.SendErrorWebhook(request.CallbackUrl, "Deployment error: " + err.Error())
		}
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
		ctx.Writer.Write([]byte("data: {\"error\": \"" + err.Error() + "}\n\n"))
	}
}

// smartEscapeJSON handles literal newlines in JSON strings more intelligently
// This prevents double-escaping while handling merge commit messages with newlines
func smartEscapeJSON(jsonStr string) string {
	// First, temporarily replace already escaped sequences to protect them
	placeholder1 := "__ESCAPED_NEWLINE__"
	placeholder2 := "__ESCAPED_CARRIAGE__"
	placeholder3 := "__ESCAPED_TAB__"
	
	// Protect already escaped sequences
	jsonStr = strings.ReplaceAll(jsonStr, "\\n", placeholder1)
	jsonStr = strings.ReplaceAll(jsonStr, "\\r", placeholder2)
	jsonStr = strings.ReplaceAll(jsonStr, "\\t", placeholder3)
	
	// Now escape literal newlines, carriage returns, and tabs
	jsonStr = strings.ReplaceAll(jsonStr, "\n", "\\n")
	jsonStr = strings.ReplaceAll(jsonStr, "\r", "\\r")
	jsonStr = strings.ReplaceAll(jsonStr, "\t", "\\t")
	
	// Restore the already escaped sequences
	jsonStr = strings.ReplaceAll(jsonStr, placeholder1, "\\n")
	jsonStr = strings.ReplaceAll(jsonStr, placeholder2, "\\r")
	jsonStr = strings.ReplaceAll(jsonStr, placeholder3, "\\t")
	
	return jsonStr
}
