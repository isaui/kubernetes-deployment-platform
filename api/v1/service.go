package v1

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pendeploy-simple/dto"
	"github.com/pendeploy-simple/models"
	"github.com/pendeploy-simple/services"
	"github.com/pendeploy-simple/utils"
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

// GetLatestDeployment returns latest deployment - UPDATED untuk handle managed services
func (c *ServiceController) GetLatestDeployment(ctx *gin.Context) {
	// Get service ID from URL
	serviceID := ctx.Param("id")
	
	// Get userId and role from context
	roleValue, _ := ctx.Get("role")
	role, _ := roleValue.(string)
	isAdmin := role == "admin"
	userIDValue, _ := ctx.Get("userId")
	userID := userIDValue.(string)

	// Check if this is a git service first
	service, err := c.serviceService.GetServiceDetail(serviceID, userID, isAdmin)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	if service.Type != models.ServiceTypeGit {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Deployments are only available for git services. Managed services don't have deployments.",
		})
		return
	}

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


// GetDeploymentList returns deployment list - UPDATED untuk handle managed services
func (c *ServiceController) GetDeploymentList(ctx *gin.Context) {
	// Get service ID from URL
	serviceID := ctx.Param("id")
	
	// Get userId and role from context
	roleValue, _ := ctx.Get("role")
	role, _ := roleValue.(string)
	isAdmin := role == "admin"
	userIDValue, _ := ctx.Get("userId")
	userID := userIDValue.(string)

	// Check if this is a git service first
	service, err := c.serviceService.GetServiceDetail(serviceID, userID, isAdmin)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	if service.Type != models.ServiceTypeGit {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Deployments are only available for git services. Managed services don't have deployments.",
		})
		return
	}

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

// CreateService creates a new service - UPDATED untuk managed services validation
func (c *ServiceController) CreateService(ctx *gin.Context) {
	// Get userId and role from context
	userIDValue, _ := ctx.Get("userId")
	userID := userIDValue.(string)
	roleValue, _ := ctx.Get("role") 
	role, _ := roleValue.(string)
	isAdmin := role == "admin"

	// Parse request body
	var req dto.ServiceRequest
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
		// Managed services require ManagedType and validation
		if req.ManagedType == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": "ManagedType is required for managed services",
			})
			return
		}
		
		// Validate managed service type
		if !utils.IsValidManagedServiceType(req.ManagedType) {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": "Unsupported managed service type: " + req.ManagedType,
			})
			return
		}
		
		// For managed services, EnvVars should not be provided by user (auto-generated)
		if req.EnvVars != nil && len(req.EnvVars) > 0 {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": "Environment variables are auto-generated for managed services and cannot be specified",
			})
			return
		}
		
		// Managed services don't need git-specific fields
		if req.RepoURL != "" || req.Branch != "" || req.BuildCommand != "" || req.StartCommand != "" {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": "Git-specific fields (repoUrl, branch, buildCommand, startCommand) are not allowed for managed services",
			})
			return
		}
		
		// Port is auto-determined for managed services
		if req.Port != 0 {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": "Port is auto-determined for managed services and cannot be specified",
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
		EnvVars:        req.EnvVars, // Will be empty for managed services
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

// UpdateService updates an existing service - UPDATED untuk use existing DTO
func (c *ServiceController) UpdateService(ctx *gin.Context) {
	// Get service ID from URL
	serviceID := ctx.Param("id")
	log.Println(serviceID)
	
	// Get userId and role from context
	userIDValue, _ := ctx.Get("userId")
	userID := userIDValue.(string)
	roleValue, _ := ctx.Get("role")
	role, _ := roleValue.(string)
	isAdmin := role == "admin"

	// First, get existing service to verify type and permissions
	existingService, err := c.serviceService.GetServiceDetail(serviceID, userID, isAdmin)
	if err != nil {
		log.Println("service not found")
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": "Service not found or access denied",
		})
		return
	}

	// Parse request body using existing DTO
	var updateReq dto.ServiceUpdateRequest
	if err := ctx.ShouldBindJSON(&updateReq); err != nil {
		log.Println("error binding json")
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Validate that the request type matches the existing service type
	serviceType := string(existingService.Type)
	if updateReq.Type != serviceType {
		log.Println("service type not match")
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Cannot change service type",
		})
		return
	}

	// Validate the update request
	if err := updateReq.ValidateServiceUpdateRequest(); err != nil {
		log.Println("error validating update request")
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Create a service object for update
	service := models.Service{
		ID: serviceID,
	}

	// Use the DTO to update service model
	updateReq.UpdateServiceModel(&service)
	log.Println("update service model")
    log.Println(service)
	// Call service layer to update
	updatedService, err := c.serviceService.UpdateService(service, userID, isAdmin)
	if err != nil {
		log.Println("error updating service")
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
