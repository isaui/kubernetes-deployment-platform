package utils

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

// SimpleAuthMiddleware creates a middleware to authenticate API requests using a static API key
func SimpleAuthMiddleware() gin.HandlerFunc {
	// Get API key from environment variable
	apiKey := os.Getenv("PENDEPLOY_API_KEY")
	
	// Ensure API key is set in environment
	if apiKey == "" {
		log.Fatalf("❌ ERROR: No API key set in environment. Set PENDEPLOY_API_KEY environment variable")
	}
	
	return func(c *gin.Context) {
		// Skip auth check for health check endpoint
		if c.Request.URL.Path == "/" {
			c.Next()
			return
		}
		// Only check for X-API-Key header
		requestKey := c.GetHeader("X-API-Key")

		// Validate API key
		if requestKey == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"status":  "error",
				"message": "API key is required",
			})
			c.Abort()
			return
		}

		// Simple string comparison - for a quick solution
		if requestKey != apiKey {
			c.JSON(http.StatusUnauthorized, gin.H{
				"status":  "error",
				"message": "Invalid API key",
			})
			c.Abort()
			return
		}

		// Continue to the next handler if authentication passed
		c.Next()
	}
}
