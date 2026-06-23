package models

import "gorm.io/gorm"

const (
	NotifTypeDisk    = "disk_alert"
	NotifTypeRAM     = "ram_alert"
	NotifTypeBattery = "battery_alert"
	NotifTypeOffline = "device_offline"
)

type Notification struct {
	gorm.Model
	DeviceID string `json:"device_id" gorm:"index"`
	Type     string `json:"type"`
	Message  string `json:"message" gorm:"type:text"`
	Sent     bool   `json:"sent" gorm:"default:false"`
}
