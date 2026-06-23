package system

// Metrics holds all system monitoring data
type Metrics struct {
	// CPU
	CPUUsage   float64 `json:"cpu_usage"`
	CPUTemp    float64 `json:"cpu_temp"`
	CPUFreqMHz float64 `json:"cpu_freq_mhz"`

	// RAM
	RAMTotalMB uint64 `json:"ram_total_mb"`
	RAMUsedMB  uint64 `json:"ram_used_mb"`
	RAMFreeMB  uint64 `json:"ram_free_mb"`

	// Disk
	DiskInfo []DiskEntry `json:"disk_info"`

	// Network
	UploadKBps   float64 `json:"upload_kbps"`
	DownloadKBps float64 `json:"download_kbps"`
	LocalIP      string  `json:"local_ip"`
	InternetOK   bool    `json:"internet_ok"`

	// Battery
	BatteryLevel    int  `json:"battery_level"`    // -1 = no battery
	BatteryCharging bool `json:"battery_charging"`
	BatteryMinutes  int  `json:"battery_minutes"`
}

type DiskEntry struct {
	Mount     string  `json:"mount"`
	TotalGB   float64 `json:"total_gb"`
	UsedGB    float64 `json:"used_gb"`
	FreeGB    float64 `json:"free_gb"`
	UsagePct  float64 `json:"usage_pct"`
}
