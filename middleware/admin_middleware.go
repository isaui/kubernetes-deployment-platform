package middleware

import (
	"net/http"
	"github.com/gin-gonic/gin"
)

// AdminMiddleware creates a middleware that ensures the user has admin role
// This middleware should be used after AuthMiddleware
func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get role from context (set by AuthMiddleware)
		role, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"status":  "error",
				"message": "Authentication required",
			})
			c.Abort()
			return
		}

		// Check if role is admin
		if roleStr, ok := role.(string); !ok || roleStr != "admin" {
			c.JSON(http.StatusForbidden, gin.H{
				"status":  "error",
				"message": "Admin privileges required",
			})
			c.Abort()
			return
		}

		// User is admin, continue
		c.Next()
	}
}
