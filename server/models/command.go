package models

import "gorm.io/gorm"

// Command statuses
const (
	CommandStatusPending  = "pending"
	CommandStatusRunning  = "running"
	CommandStatusDone     = "done"
	CommandStatusFailed   = "failed"
)

type Command struct {
	gorm.Model
	DeviceID string `json:"device_id" gorm:"index;not null"`
	Type     string `json:"type" gorm:"not null"` // "shutdown","restart","sleep","lock","logout","screenshot","media","file","status"
	Payload  string `json:"payload" gorm:"type:text"` // JSON payload
	Status   string `json:"status" gorm:"default:pending"`
	Result   string `json:"result" gorm:"type:text"`  // JSON result from agent
	TelegramChatID int64 `json:"telegram_chat_id"`        // reply to this chat
	TelegramMsgID  int   `json:"telegram_msg_id"`
}
