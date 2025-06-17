package main

import (
	"log"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/pendeploy-simple/handlers"
	"github.com/pendeploy-simple/utils"
)

func main() {
	// Set Gin mode
	gin.SetMode(gin.ReleaseMode)
	
	// Initialize router
	router := gin.Default()

	// CORS configuration
	router.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-API-Key"},
		AllowCredentials: true,
	}))
	
	// Add API key authentication middleware
	router.Use(utils.SimpleAuthMiddleware())

	// Health check endpoint
	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "pendeploy-simple",
			"version": "1.0.0",
		})
	})

	// Main deployment endpoint
	router.POST("/create-deployment", handlers.CreateDeployment)

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Start server
	log.Printf("🚀 PenDeploy Simple starting on port %s", port)
	log.Printf("💡 API Authentication: %s", func() string {
		if os.Getenv("PENDEPLOY_API_KEY") != "" {
			return "Enabled"
		}
		return "Disabled (INSECURE)"
	}())
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}