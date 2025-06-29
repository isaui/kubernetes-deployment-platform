package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pendeploy-simple/dto"
	"github.com/pendeploy-simple/models"
	"github.com/pendeploy-simple/services"
)

// ServiceController handles service-related API endpoints
type ServiceController struct {
	serviceService *services.ServiceService
}

// NewServiceController creates a new service controller
func NewServiceController() *ServiceController {
	return &ServiceController{
		serviceService: services.NewServiceService(),
	}
}


// RegisterRoutes registers service routes
func (c *ServiceController) RegisterRoutes(router *gin.RouterGroup) {
	servicesGroup := router.Group("/services")
	{
		servicesGroup.GET("", c.ListServices)
		servicesGroup.GET("/:id", c.GetService)
		servicesGroup.POST("", c.CreateService)
		servicesGroup.PUT("/:id", c.UpdateService)
		servicesGroup.DELETE("/:id", c.DeleteService)
		servicesGroup.GET("/:id/deployments", c.GetDeploymentList)
		servicesGroup.GET("/:id/latest-deployment", c.GetLatestDeployment)
	}

	// Also add project-specific service routes
	projects := router.Group("/projects")
	{
		projects.GET("/:id/services", c.ListProjectServices)
	}
}

func (c *ServiceController) GetLatestDeployment(ctx *gin.Context) {
	// Get service ID from URL
	serviceID := ctx.Param("id")
	// Get userId and role from context
	roleValue, _ := ctx.Get("role")
	role, _ := roleValue.(string)
	isAdmin := role == "admin"
	userIDValue, _ := ctx.Get("userId")
	userID := userIDValue.(string)

	deployment, err := c.serviceService.GetLatestDeployment(serviceID, userID, isAdmin)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"deployment": deployment,
		},
	})
}

func (c *ServiceController) GetDeploymentList(ctx *gin.Context) {
	// Get service ID from URL
	serviceID := ctx.Param("id")
	
	// Get userId and role from context
	roleValue, _ := ctx.Get("role")
	role, _ := roleValue.(string)
	isAdmin := role == "admin"
	userIDValue, _ := ctx.Get("userId")
	userID := userIDValue.(string)

	deployments, err := c.serviceService.GetDeploymentList(serviceID, userID, isAdmin)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"deployments": deployments,
		},
	})
}

// ListServices retrieves all services (admin only)
func (c *ServiceController) ListServices(ctx *gin.Context) {
	// Get userId and role from context
	roleValue, _ := ctx.Get("role")
	role, _ := roleValue.(string)
	isAdmin := role == "admin"

	if !isAdmin {
		ctx.JSON(http.StatusForbidden, gin.H{
			"error": "Only administrators can view all services",
		})
		return
	}

	services, err := c.serviceService.ListAllServices()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve services",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"services": services,
		},
	})
}

// ListProjectServices retrieves all services for a specific project
func (c *ServiceController) ListProjectServices(ctx *gin.Context) {
	// Get project ID from URL
	projectID := ctx.Param("id")
	
	// Get userId and role from context
	userIDValue, _ := ctx.Get("userId")
	userID := userIDValue.(string)
	roleValue, _ := ctx.Get("role")
	role, _ := roleValue.(string)
	isAdmin := role == "admin"

	services, err := c.serviceService.ListProjectServices(projectID, userID, isAdmin)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"services": services,
		},
	})
}

// GetService retrieves a specific service
func (c *ServiceController) GetService(ctx *gin.Context) {
	// Get service ID from URL
	serviceID := ctx.Param("id")
	
	// Get userId and role from context
	userIDValue, _ := ctx.Get("userId")
	userID := userIDValue.(string)
	roleValue, _ := ctx.Get("role")
	role, _ := roleValue.(string)
	isAdmin := role == "admin"

	service, err := c.serviceService.GetServiceDetail(serviceID, userID, isAdmin)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data": service,
	})
}

// ServiceRequest represents a service creation/update request
type ServiceRequest struct {
	// Common fields for all service types
	Name          string             `json:"name" binding:"required"`
	Type          models.ServiceType `json:"type" binding:"required"` // "git" or "managed"
	ProjectID     string             `json:"projectId" binding:"required"`
	EnvironmentID string             `json:"environmentId" binding:"required"`
	
	// Git-specific fields (required only when Type is "git")
	RepoURL       string             `json:"repoUrl"`
	Branch        string             `json:"branch"`
	Port          int                `json:"port"`
	BuildCommand  string             `json:"buildCommand"`
	StartCommand  string             `json:"startCommand"`
	
	// Managed service specific fields (required only when Type is "managed")
	ManagedType   string             `json:"managedType"` // postgresql, redis, minio, etc.
	Version       string             `json:"version"`     // 14, 6.0, latest, etc.
	StorageSize   string             `json:"storageSize"` // 1Gi, 10Gi, etc.
	
	// Common configuration fields
	EnvVars       models.EnvVars     `json:"envVars"`
	CPULimit      string             `json:"cpuLimit"`
	MemoryLimit   string             `json:"memoryLimit"`
	IsStaticReplica bool             `json:"isStaticReplica"`
	Replicas      int                `json:"replicas"`
	MinReplicas   int                `json:"minReplicas"`
	MaxReplicas   int                `json:"maxReplicas"`
	CustomDomain  string             `json:"customDomain"`
}

// CreateService creates a new service
func (c *ServiceController) CreateService(ctx *gin.Context) {
	// Get userId and role from context
	userIDValue, _ := ctx.Get("userId")
	userID := userIDValue.(string)
	roleValue, _ := ctx.Get("role") 
	role, _ := roleValue.(string)
	isAdmin := role == "admin"

	// Parse request body
	var req ServiceRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Validate fields based on service type
	if req.Type == models.ServiceTypeGit {
		// Git services require RepoURL
		if req.RepoURL == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": "Repository URL is required for git services",
			})
			return
		}
	} else if req.Type == models.ServiceTypeManaged {
		// Managed services require ManagedType
		if req.ManagedType == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": "ManagedType is required for managed services",
			})
			return
		}
	} else {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid service type",
		})
		return
	}

	// Create service object
	service := models.Service{
		Name:           req.Name,
		Type:           req.Type,
		ProjectID:      req.ProjectID,
		EnvironmentID:  req.EnvironmentID,
		
		// Git-specific fields
		RepoURL:        req.RepoURL,
		Branch:         req.Branch,
		Port:           req.Port,
		BuildCommand:   req.BuildCommand,
		StartCommand:   req.StartCommand,
		
		// Managed service fields
		ManagedType:    req.ManagedType,
		Version:        req.Version,
		StorageSize:    req.StorageSize,
		
		// Common configuration fields
		EnvVars:        req.EnvVars,
		CPULimit:       req.CPULimit,
		MemoryLimit:    req.MemoryLimit,
		IsStaticReplica: req.IsStaticReplica,
		Replicas:       req.Replicas,
		MinReplicas:    req.MinReplicas,
		MaxReplicas:    req.MaxReplicas,
		CustomDomain:   req.CustomDomain,
	}

	// Call service to create
	createdService, err := c.serviceService.CreateService(service, userID, isAdmin)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"data": createdService,
	})
}

// UpdateService updates an existing service
func (c *ServiceController) UpdateService(ctx *gin.Context) {
	// Get service ID from URL
	serviceID := ctx.Param("id")
	
	// Get userId and role from context
	userIDValue, _ := ctx.Get("userId")
	userID := userIDValue.(string)
	roleValue, _ := ctx.Get("role")
	role, _ := roleValue.(string)
	isAdmin := role == "admin"

	// First, get existing service to verify type and permissions
	existingService, err := c.serviceService.GetServiceDetail(serviceID, userID, isAdmin)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": "Service not found or access denied",
		})
		return
	}

	// Parse request body using the new DTO
	var updateReq dto.ServiceUpdateRequest
	if err := ctx.ShouldBindJSON(&updateReq); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Validate that the request type matches the existing service type
	serviceType := string(existingService.Type)
	if updateReq.Type != serviceType {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Cannot change service type",
		})
		return
	}

	// Create a service object for update
	service := models.Service{
		ID: serviceID,
	}

	// Apply the updates based on request type
	if updateReq.Type == "git" {
		if updateReq.Git == nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": "Missing git configuration data",
			})
			return
		}
		
		// Use the DTO to update service model
		updateReq.UpdateServiceModel(&service)
		
	} else if updateReq.Type == "managed" {
		if updateReq.Managed == nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": "Missing managed service configuration data",
			})
			return
		}
		
		// Use the DTO to update service model
		updateReq.UpdateServiceModel(&service)
		
	} else {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid service type",
		})
		return
	}

	// Call service layer to update
	updatedService, err := c.serviceService.UpdateService(service, userID, isAdmin)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data": updatedService,
	})
}

// DeleteService deletes a service
func (c *ServiceController) DeleteService(ctx *gin.Context) {
	// Get service ID from URL
	serviceID := ctx.Param("id")
	
	// Get userId and role from context
	userIDValue, _ := ctx.Get("userId")
	userID := userIDValue.(string)
	roleValue, _ := ctx.Get("role")
	role, _ := roleValue.(string)
	isAdmin := role == "admin"

	// Call service to delete
	err := c.serviceService.DeleteService(serviceID, userID, isAdmin)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"message": "Service deleted successfully",
		},
	})
}
