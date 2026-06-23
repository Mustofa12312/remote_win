package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/mustofa/commender/server/database"
	"github.com/mustofa/commender/server/models"
)

// DeviceAuth validates device_id + secret from Authorization header
// Header format: "Bearer <device_id>:<secret>"
func DeviceAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing authorization"})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization format"})
			return
		}

		credentials := strings.SplitN(parts[1], ":", 2)
		if len(credentials) != 2 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials format"})
			return
		}

		deviceID := credentials[0]
		secret := credentials[1]

		var device models.Device
		if err := database.DB.Where("device_id = ? AND secret = ?", deviceID, secret).First(&device).Error; err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid device credentials"})
			return
		}

		c.Set("device", &device)
		c.Next()
	}
}

// ServerAuth validates a static server API key for dashboard access
func ServerAuth(apiKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.GetHeader("X-API-Key")
		if key == "" {
			key = c.Query("api_key")
		}
		if key != apiKey || apiKey == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid api key"})
			return
		}
		c.Next()
	}
}
