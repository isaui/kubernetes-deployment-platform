package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pendeploy-simple/services"
)

// GetPodStats returns stats about pods in the cluster
func GetPodStats(c *gin.Context) {
	statsService, err := services.NewStatsService()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to initialize stats service: " + err.Error(),
		})
		return
	}

	data, err := statsService.GetPodStats(c.Query("namespace"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get pod stats: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, data)
}

// GetNodeStats returns stats about nodes in the cluster
func GetNodeStats(c *gin.Context) {
	statsService, err := services.NewStatsService()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to initialize stats service: " + err.Error(),
		})
		return
	}

	data, err := statsService.GetNodeStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get node stats: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, data)
}

// GetDeploymentStats returns stats about deployments in the cluster
func GetDeploymentStats(c *gin.Context) {
	statsService, err := services.NewStatsService()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to initialize stats service: " + err.Error(),
		})
		return
	}

	// Get namespace query param if any
	namespace := c.Query("namespace")

	data, err := statsService.GetDeploymentStats(namespace)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get deployment stats: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, data)
}

// GetClusterInfo returns general information about the cluster
func GetClusterInfo(c *gin.Context) {
	statsService, err := services.NewStatsService()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to initialize stats service: " + err.Error(),
		})
		return
	}

	data, err := statsService.GetClusterInfo()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get cluster info: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, data)
}

// GetServiceStats returns statistics about Kubernetes services
func GetServiceStats(c *gin.Context) {
	statsService, err := services.NewStatsService()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to initialize stats service: " + err.Error(),
		})
		return
	}

	// Get namespace query param if any
	namespace := c.Query("namespace")

	data, err := statsService.GetServiceStats(namespace)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get service stats: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, data)
}

// GetIngressStats returns statistics about Kubernetes ingress resources
func GetIngressStats(c *gin.Context) {
	statsService, err := services.NewStatsService()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to initialize stats service: " + err.Error(),
		})
		return
	}

	// Get namespace query param if any
	namespace := c.Query("namespace")

	data, err := statsService.GetIngressStats(namespace)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get ingress stats: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, data)
}

// GetCertificateStats returns statistics about cert-manager certificates
func GetCertificateStats(c *gin.Context) {
	statsService, err := services.NewStatsService()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to initialize stats service: " + err.Error(),
		})
		return
	}

	// Get namespace query param if any
	namespace := c.Query("namespace")

	data, err := statsService.GetCertificateStats(namespace)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get certificate stats: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, data)
}
