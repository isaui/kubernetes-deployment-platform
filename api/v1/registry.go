package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/pendeploy-simple/controllers"
)

// RegistryController wraps the controllers.RegistryController to use in API v1
type RegistryController struct {
	controller *controllers.RegistryController
}

// NewRegistryController creates a new registry controller for API v1
func NewRegistryController() *RegistryController {
	return &RegistryController{
		controller: controllers.NewRegistryController(),
	}
}

// RegisterRoutes registers registry API routes
func (rc *RegistryController) RegisterRoutes(router *gin.RouterGroup) {
	registryGroup := router.Group("/registries")
	{
		// List and create registries
		registryGroup.GET("", rc.controller.GetRegistries)
		registryGroup.POST("", rc.controller.CreateRegistry)
		
		// Single registry operations
		registryGroup.GET("/:id", rc.controller.GetRegistry)
		registryGroup.PUT("/:id", rc.controller.UpdateRegistry)
		registryGroup.DELETE("/:id", rc.controller.DeleteRegistry)
		
		// Registry details with K8s information
		registryGroup.GET("/:id/details", rc.controller.GetRegistryDetails)
		
		// Stream build logs endpoint - this uses server-sent events for real-time updates
		registryGroup.GET("/:id/logs/stream", rc.controller.StreamBuildLogs)
	}
}
