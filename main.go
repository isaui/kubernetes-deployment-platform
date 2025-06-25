package main

import (
	"log"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/pendeploy-simple/api/v1"
	"github.com/pendeploy-simple/database"
	"github.com/pendeploy-simple/middleware"
)

func main() {
	// Load .env file if exists
	_ = godotenv.Load()

	// Set Gin mode
	gin.SetMode(gin.ReleaseMode)
	
	// Initialize router
	router := gin.Default()

	// Initialize database connection
	database.Initialize()

	// CORS configuration
	corsAllowed := os.Getenv("CORS_ALLOWED")
	if corsAllowed == "" {
		corsAllowed = "http://localhost:5173" // Default value if not set
	}

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{corsAllowed},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowCredentials: true,
	}))

	// Root health check endpoint
	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "pendeploy-api",
			"version": "1.0.0",
		})
	})
	
	// Setup API v1 routes
	apiV1 := router.Group("/api/v1")
	// Apply middleware to the group - it has built-in exceptions for auth routes
	apiV1.Use(middleware.AuthMiddleware())
	// Register all routes
	v1.RegisterRoutes(apiV1)

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Start server
	log.Printf("🚀 PenDeploy API v1 starting on port %s", port)
	log.Printf("📚 API docs available at: http://localhost:%s/api/v1/health", port)
	
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}