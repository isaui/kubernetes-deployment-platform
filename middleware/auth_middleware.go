package middleware

import (
	"github.com/gin-gonic/gin"
)

// AuthMiddleware untuk autentikasi
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Implementasi autentikasi akan ditambahkan nanti
		c.Next()
	}
}

// AdminMiddleware untuk otorisasi admin
func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Implementasi otorisasi admin akan ditambahkan nanti
		c.Next()
	}
}
