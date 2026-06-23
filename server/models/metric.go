package models

import (
	"time"

	"gorm.io/gorm"
)

type Metric struct {
	gorm.Model
	DeviceID    string    `json:"device_id" gorm:"index;not null"`
	RecordedAt  time.Time `json:"recorded_at"`

	// CPU
	CPUUsage    float64 `json:"cpu_usage"`
	CPUTemp     float64 `json:"cpu_temp"`
	CPUFreqMHz  float64 `json:"cpu_freq_mhz"`

	// RAM
	RAMTotalMB  uint64  `json:"ram_total_mb"`
	RAMUsedMB   uint64  `json:"ram_used_mb"`
	RAMFreeMB   uint64  `json:"ram_free_mb"`

	// Disk (serialized JSON array for multiple disks)
	DiskInfo    string  `json:"disk_info" gorm:"type:text"`

	// Network
	UploadKBps   float64 `json:"upload_kbps"`
	DownloadKBps float64 `json:"download_kbps"`
	PublicIP     string  `json:"public_ip"`
	LocalIP      string  `json:"local_ip"`
	InternetOK   bool    `json:"internet_ok"`

	// Battery
	BatteryLevel   int    `json:"battery_level"`    // -1 = no battery
	BatteryCharging bool   `json:"battery_charging"`
	BatteryMinutes  int    `json:"battery_minutes"`
}
