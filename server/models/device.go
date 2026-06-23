package models

import (
	"time"

	"gorm.io/gorm"
)

type Device struct {
	gorm.Model
	Name       string    `json:"name" gorm:"not null"`
	OS         string    `json:"os" gorm:"not null"`         // "windows" | "ubuntu" | "linux"
	Hostname   string    `json:"hostname"`
	LocalIP    string    `json:"local_ip"`
	DeviceID   string    `json:"device_id" gorm:"uniqueIndex;not null"`
	Secret     string    `json:"-" gorm:"not null"`           // hashed
	AgentVersion string  `json:"agent_version"`
	Status     string    `json:"status" gorm:"default:offline"` // "online" | "offline"
	LastSeen   time.Time `json:"last_seen"`
}
