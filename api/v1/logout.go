package v1

import (
	"net/http"
	"github.com/gin-gonic/gin"
)

// Logout handles user logout
func Logout(c *gin.Context) {
	// Clear the cookie by setting max-age to -1 (expired)
	c.SetCookie(
		"access_token", // name
		"",             // value (empty)
		-1,             // max age (expired)
		"/",            // path
		"",             // domain
		true,           // secure (HTTPS only)
		true,           // httpOnly (not accessible via JS)
	)

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Logged out successfully",
	})
}
