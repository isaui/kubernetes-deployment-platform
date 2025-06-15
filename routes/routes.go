package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/isabu/pendeploy-handal/controllers"
)

// SetupRoutes mengatur semua routes API
func SetupRoutes(router *gin.Engine) {
	// Public routes
	router.GET("/", controllers.HealthCheck)
	
	// Auth routes
	auth := router.Group("/api/auth")
	{
		auth.POST("/register", controllers.Register)
		auth.POST("/login", controllers.Login)
	}

	// API routes
	api := router.Group("/api")
	{
		// Deployments
		deployments := api.Group("/deployments")
		{
			deployments.GET("/", controllers.GetDeployments)
			deployments.POST("/", controllers.CreateDeployment)
			deployments.GET("/:id", controllers.GetDeployment)
			deployments.PUT("/:id", controllers.UpdateDeployment)
			deployments.DELETE("/:id", controllers.DeleteDeployment)
			deployments.PATCH("/:id/status", controllers.UpdateDeploymentStatus)
		}
		
		
		// Users
		users := api.Group("/users")
		{
			users.GET("/", controllers.GetUsers)
			users.GET("/:id", controllers.GetUser)
			users.PUT("/:id", controllers.UpdateUser)
			users.DELETE("/:id", controllers.DeleteUser)
		}
	}
}
