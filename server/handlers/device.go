package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mustofa/commender/server/database"
	"github.com/mustofa/commender/server/models"
)

type RegisterDeviceRequest struct {
	Name         string `json:"name" binding:"required"`
	OS           string `json:"os" binding:"required"`
	Hostname     string `json:"hostname"`
	AgentVersion string `json:"agent_version"`
}

type HeartbeatRequest struct {
	LocalIP  string         `json:"local_ip"`
	Metrics  *models.Metric `json:"metrics"`
}

// POST /api/devices/register
func RegisterDevice(c *gin.Context) {
	var req RegisterDeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Generate device_id and secret
	deviceID, err := generateToken(16)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate device id"})
		return
	}
	secret, err := generateToken(32)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate secret"})
		return
	}

	device := models.Device{
		Name:         req.Name,
		OS:           req.OS,
		Hostname:     req.Hostname,
		DeviceID:     deviceID,
		Secret:       secret,
		AgentVersion: req.AgentVersion,
		Status:       "online",
		LastSeen:     time.Now(),
	}

	if err := database.DB.Create(&device).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to register device"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"device_id": deviceID,
		"secret":    secret,
		"message":   "device registered successfully",
	})
}

// POST /api/devices/heartbeat  (requires DeviceAuth)
func Heartbeat(c *gin.Context) {
	device := c.MustGet("device").(*models.Device)

	var req HeartbeatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	now := time.Now()

	// Update device status
	updates := map[string]interface{}{
		"status":    "online",
		"last_seen": now,
	}
	if req.LocalIP != "" {
		updates["local_ip"] = req.LocalIP
	}
	database.DB.Model(device).Updates(updates)

	// Save metrics if provided
	if req.Metrics != nil {
		req.Metrics.DeviceID = device.DeviceID
		req.Metrics.RecordedAt = now
		database.DB.Create(req.Metrics)
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok", "server_time": now.Unix()})
}

// GET /api/devices  (dashboard)
func ListDevices(c *gin.Context) {
	var devices []models.Device

	// Mark devices offline if not seen for > 2 minutes
	database.DB.Model(&models.Device{}).
		Where("status = ? AND last_seen < ?", "online", time.Now().Add(-2*time.Minute)).
		Update("status", "offline")

	database.DB.Find(&devices)
	c.JSON(http.StatusOK, gin.H{"devices": devices})
}

// GET /api/devices/:device_id  (dashboard)
func GetDevice(c *gin.Context) {
	deviceID := c.Param("device_id")
	var device models.Device
	if err := database.DB.Where("device_id = ?", deviceID).First(&device).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "device not found"})
		return
	}

	// Get latest metrics
	var metric models.Metric
	database.DB.Where("device_id = ?", deviceID).Order("recorded_at desc").First(&metric)

	c.JSON(http.StatusOK, gin.H{
		"device": device,
		"metric": metric,
	})
}

// GET /api/devices/:device_id/metrics
func GetMetrics(c *gin.Context) {
	deviceID := c.Param("device_id")
	var metrics []models.Metric
	database.DB.Where("device_id = ?", deviceID).
		Order("recorded_at desc").
		Limit(20).
		Find(&metrics)
	c.JSON(http.StatusOK, gin.H{"metrics": metrics})
}

func generateToken(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
