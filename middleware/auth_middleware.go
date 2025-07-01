package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/pendeploy-simple/services"
)

// AuthMiddleware creates a middleware to authenticate API requests
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip auth for public endpoints
		if c.Request.URL.Path == "/" || 
		   c.Request.URL.Path == "/api/v1/health" || 
		   c.Request.URL.Path == "/api/v1/auth/login" ||
		   c.Request.URL.Path == "/api/v1/auth/register" ||
		   c.Request.URL.Path == "/api/v1/auth/logout" ||
		   c.Request.URL.Path == "/api/v1/auth/refresh" ||
		   strings.HasPrefix(c.Request.URL.Path, "/api/v1/deployments") {
			c.Next()
			return
		}

		// Try to get token from different sources
		var tokenString string
		
		// 1. First check Authorization header (Bearer token)
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			// Check if the auth header has the Bearer format
			tokenParts := strings.Split(authHeader, " ")
			if len(tokenParts) == 2 && strings.ToLower(tokenParts[0]) == "bearer" {
				tokenString = tokenParts[1]
			}
		}
		
		// 2. If no valid Authorization header, try cookie
		if tokenString == "" {
			cookieValue, err := c.Cookie("access_token")
			if err == nil && cookieValue != "" {
				tokenString = cookieValue
			}
		}
		
		// 3. If no token found in either place, return unauthorized
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"status":  "error",
				"message": "Authentication required. Please login.",
			})
			c.Abort()
			return
		}
		
		// Validate token
		claims, err := services.ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"status":  "error",
				"message": "Invalid or expired token",
			})
			c.Abort()
			return
		}

		// Set user info in context
		c.Set("userId", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("role", claims.Role)

		// Continue to the next handler
		c.Next()
	}
}
