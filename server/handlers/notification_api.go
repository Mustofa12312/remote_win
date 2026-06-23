package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mustofa/commender/server/database"
	"github.com/mustofa/commender/server/models"
)

// GET /api/notifications
func ListNotifications(c *gin.Context) {
	deviceID := c.Query("device_id")
	query := database.DB.Order("created_at desc").Limit(100)
	if deviceID != "" {
		query = query.Where("device_id = ?", deviceID)
	}
	var notifs []models.Notification
	query.Find(&notifs)
	c.JSON(http.StatusOK, gin.H{"notifications": notifs})
}
