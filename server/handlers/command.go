package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mustofa/commender/server/database"
	"github.com/mustofa/commender/server/models"
)

type EnqueueCommandRequest struct {
	DeviceID      string `json:"device_id" binding:"required"`
	Type          string `json:"type" binding:"required"`
	Payload       string `json:"payload"`
	TelegramChatID int64 `json:"telegram_chat_id"`
	TelegramMsgID  int   `json:"telegram_msg_id"`
}

// POST /api/commands  (internal — called by bot)
func EnqueueCommand(c *gin.Context) {
	var req EnqueueCommandRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cmd := models.Command{
		DeviceID:       req.DeviceID,
		Type:           req.Type,
		Payload:        req.Payload,
		Status:         models.CommandStatusPending,
		TelegramChatID: req.TelegramChatID,
		TelegramMsgID:  req.TelegramMsgID,
	}

	if err := database.DB.Create(&cmd).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to enqueue command"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"command_id": cmd.ID})
}

// GET /api/commands/poll  (agent — requires DeviceAuth)
func PollCommands(c *gin.Context) {
	device := c.MustGet("device").(*models.Device)

	var commands []models.Command
	database.DB.Where("device_id = ? AND status = ?", device.DeviceID, models.CommandStatusPending).
		Order("created_at asc").
		Limit(10).
		Find(&commands)

	// Mark as running
	for _, cmd := range commands {
		database.DB.Model(&cmd).Update("status", models.CommandStatusRunning)
	}

	c.JSON(http.StatusOK, gin.H{"commands": commands})
}

// POST /api/commands/:id/result  (agent — requires DeviceAuth)
func ReportCommandResult(c *gin.Context) {
	cmdID := c.Param("id")

	var body struct {
		Status  string          `json:"status"` // "done" | "failed"
		Result  json.RawMessage `json:"result"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var cmd models.Command
	if err := database.DB.First(&cmd, cmdID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "command not found"})
		return
	}

	resultStr := string(body.Result)
	database.DB.Model(&cmd).Updates(map[string]interface{}{
		"status": body.Status,
		"result": resultStr,
	})

	c.JSON(http.StatusOK, gin.H{"ok": true, "command_id": cmd.ID})
}

// GET /api/commands/history  (dashboard)
func ListCommandHistory(c *gin.Context) {
	deviceID := c.Query("device_id")
	query := database.DB.Order("created_at desc").Limit(50)
	if deviceID != "" {
		query = query.Where("device_id = ?", deviceID)
	}
	var commands []models.Command
	query.Find(&commands)
	c.JSON(http.StatusOK, gin.H{"commands": commands})
}
