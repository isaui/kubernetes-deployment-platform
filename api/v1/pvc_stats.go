package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pendeploy-simple/services"
)

// GetPVCStats returns statistics about Persistent Volume Claims in the cluster
func GetPVCStats(c *gin.Context) {
	pvcStatsService := services.NewPVCStatsService()
	
	// Get namespace query param if any
	namespace := c.Query("namespace")
	
	data, err := pvcStatsService.GetPVCStats(namespace)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get PVC stats: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, data)
}
