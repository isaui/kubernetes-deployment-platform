package controllers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// HealthCheck returns the API status
func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "success",
		"service":   "pendeploy-handal",
		"version":   "1.0.0",
		"timestamp": time.Now().Format(time.RFC3339),
	})
}
