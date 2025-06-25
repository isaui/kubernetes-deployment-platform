package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/pendeploy-simple/middleware"
)

// RegisterRoutes registers all v1 API routes
func RegisterRoutes(router *gin.RouterGroup) {
	// Health check endpoint
	router.GET("/health", HealthCheck)

	// Auth endpoints
	authGroup := router.Group("/auth")
	{
		authGroup.POST("/register", Register)
		authGroup.POST("/login", Login)
		authGroup.POST("/logout", Logout)
		// Use auth middleware here only for the /me endpoint
		authGroup.GET("/me", middleware.AuthMiddleware(), GetCurrentUser)
	}

	// Project endpoints - protected by AuthMiddleware
	projectGroup := router.Group("/projects")
	projectGroup.Use(middleware.AuthMiddleware())
	{
		projectGroup.GET("", ListProjects)
		projectGroup.POST("", CreateProject)
		projectGroup.GET("/:id", GetProject)
		projectGroup.PUT("/:id", UpdateProject)
		projectGroup.DELETE("/:id", DeleteProject)
		projectGroup.GET("/:id/stats", GetProjectStats)
	}

	// Environment endpoints - protected by AuthMiddleware
	environmentController := NewEnvironmentController()
	authRouter := router.Group("")
	authRouter.Use(middleware.AuthMiddleware())
	environmentController.RegisterRoutes(authRouter)
	
	// Service endpoints - protected by AuthMiddleware
	serviceController := NewServiceController()
	serviceController.RegisterRoutes(authRouter)

	// Admin endpoints - protected by AdminMiddleware
	statsGroup := router.Group("/admin")
	// Apply admin middleware to ensure only admins can access these routes
	statsGroup.Use(middleware.AdminMiddleware())
	{
		statsGroup.GET("/stats/pods", GetPodStats)
		statsGroup.GET("/stats/nodes", GetNodeStats)
		statsGroup.GET("/stats/deployments", GetDeploymentStats)
		statsGroup.GET("/stats/services", GetServiceStats)
		statsGroup.GET("/stats/ingress", GetIngressStats)
		statsGroup.GET("/stats/certificates", GetCertificateStats)
		statsGroup.GET("/cluster/info", GetClusterInfo)
	}
}
