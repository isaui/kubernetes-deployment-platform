package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pendeploy-simple/dto"
	"github.com/pendeploy-simple/services"
)

// Register handles user registration
func Register(c *gin.Context) {
	var req dto.RegisterRequest

	// Parse request body
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid request body",
			"error":   err.Error(),
		})
		return
	}

	// Register user
	user, err := services.Register(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Registration failed",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status":  "success",
		"message": "User registered successfully",
		"user":    user,
	})
}

// Login handles user authentication
func Login(c *gin.Context) {
	var req dto.LoginRequest

	// Parse request body
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid request body",
			"error":   err.Error(),
		})
		return
	}

	// Authenticate user
	authResponse, err := services.Login(req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "error",
			"message": "Authentication failed",
			"error":   err.Error(),
		})
		return
	}

	// Set token as HttpOnly cookie (expires in 24 hours)
	c.SetCookie(
		"access_token",    // name
		authResponse.Token,  // value
		86400,              // max age (24 hours in seconds)
		"/",              // path
		"",               // domain
		true,               // secure (HTTPS only)
		true,               // httpOnly (not accessible via JS)
	)

	// Also return token in response body for clients that prefer Bearer auth
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   authResponse,
	})
}

// GetCurrentUser returns the currently authenticated user's profile
func GetCurrentUser(c *gin.Context) {
	// Get user ID from the context (set by the AuthMiddleware)
	userID, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "error",
			"message": "User not authenticated",
		})
		return
	}
	
	// Get user details from database using userID
	user, err := services.GetUser(userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to retrieve user profile",
			"error":   err.Error(),
		})
		return
	}
	
	// Return user profile
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"user":   user,
	})
}
