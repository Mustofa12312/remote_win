//go:build linux || darwin

package system

import (
	"bufio"
	"math"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
)

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

// Collect gathers all metrics
func Collect() *Metrics {
	m := &Metrics{
		BatteryLevel: -1,
	}

	m.CPUUsage = getCPUUsage()
	m.CPUTemp = getCPUTemp()
	m.CPUFreqMHz = getCPUFreq()
	collectRAM(m)
	m.DiskInfo = getDiskInfo()
	collectNetwork(m)
	collectBattery(m)
	m.LocalIP = getLocalIP()
	m.InternetOK = checkInternet()

	return m
}

func getCPUUsage() float64 {
	// Read /proc/stat twice with 500ms interval
	read := func() (idle, total uint64) {
		f, err := os.Open("/proc/stat")
		if err != nil {
			return
		}
		defer f.Close()
		scanner := bufio.NewScanner(f)
		if scanner.Scan() {
			fields := strings.Fields(scanner.Text())
			if len(fields) < 5 {
				return
			}
			var vals []uint64
			for _, s := range fields[1:] {
				n, _ := strconv.ParseUint(s, 10, 64)
				vals = append(vals, n)
			}
			for _, v := range vals {
				total += v
			}
			if len(vals) >= 4 {
				idle = vals[3]
			}
		}
		return
	}

	idle1, total1 := read()
	time.Sleep(500 * time.Millisecond)
	idle2, total2 := read()

	deltaTotal := total2 - total1
	deltaIdle := idle2 - idle1
	if deltaTotal == 0 {
		return 0
	}
	return math.Round((1.0-float64(deltaIdle)/float64(deltaTotal))*10000) / 100
}

func getCPUTemp() float64 {
	// Try hwmon
	dirs, _ := os.ReadDir("/sys/class/hwmon")
	for _, d := range dirs {
		path := "/sys/class/hwmon/" + d.Name()
		entries, _ := os.ReadDir(path)
		for _, e := range entries {
			if strings.HasPrefix(e.Name(), "temp") && strings.HasSuffix(e.Name(), "_input") {
				data, err := os.ReadFile(path + "/" + e.Name())
				if err == nil {
					n, _ := strconv.ParseFloat(strings.TrimSpace(string(data)), 64)
					return math.Round(n/1000*10) / 10
				}
			}
		}
	}
	return 0
}

func getCPUFreq() float64 {
	data, err := os.ReadFile("/sys/devices/system/cpu/cpu0/cpufreq/scaling_cur_freq")
	if err != nil {
		return 0
	}
	n, _ := strconv.ParseFloat(strings.TrimSpace(string(data)), 64)
	return math.Round(n/1000*10) / 10
}

func collectRAM(m *Metrics) {
	f, err := os.Open("/proc/meminfo")
	if err != nil {
		return
	}
	defer f.Close()
	info := map[string]uint64{}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		parts := strings.Fields(scanner.Text())
		if len(parts) >= 2 {
			key := strings.TrimSuffix(parts[0], ":")
			val, _ := strconv.ParseUint(parts[1], 10, 64)
			info[key] = val
		}
	}
	m.RAMTotalMB = info["MemTotal"] / 1024
	m.RAMFreeMB = (info["MemFree"] + info["Buffers"] + info["Cached"]) / 1024
	m.RAMUsedMB = m.RAMTotalMB - m.RAMFreeMB
}

func getDiskInfo() []DiskEntry {
	out, err := exec.Command("df", "-BG", "--output=target,size,used,avail,pcent").Output()
	if err != nil {
		return nil
	}
	var entries []DiskEntry
	lines := strings.Split(string(out), "\n")
	for _, line := range lines[1:] {
		fields := strings.Fields(line)
		if len(fields) < 5 {
			continue
		}
		mount := fields[0]
		if !strings.HasPrefix(mount, "/") || strings.HasPrefix(mount, "/sys") || strings.HasPrefix(mount, "/proc") || strings.HasPrefix(mount, "/dev") || strings.HasPrefix(mount, "/run") {
			continue
		}
		total, _ := strconv.ParseFloat(strings.TrimSuffix(fields[1], "G"), 64)
		used, _ := strconv.ParseFloat(strings.TrimSuffix(fields[2], "G"), 64)
		free, _ := strconv.ParseFloat(strings.TrimSuffix(fields[3], "G"), 64)
		pct, _ := strconv.ParseFloat(strings.TrimSuffix(fields[4], "%"), 64)
		entries = append(entries, DiskEntry{
			Mount:    mount,
			TotalGB:  total,
			UsedGB:   used,
			FreeGB:   free,
			UsagePct: pct,
		})
	}
	return entries
}

var prevRxBytes, prevTxBytes uint64
var prevNetTime time.Time

func collectNetwork(m *Metrics) {
	iface := getPrimaryInterface()
	if iface == "" {
		return
	}

	rx, tx := getNetBytes(iface)
	now := time.Now()

	if !prevNetTime.IsZero() {
		dt := now.Sub(prevNetTime).Seconds()
		if dt > 0 {
			m.UploadKBps = math.Round(float64(tx-prevTxBytes)/dt/1024*100) / 100
			m.DownloadKBps = math.Round(float64(rx-prevRxBytes)/dt/1024*100) / 100
		}
	}

	prevRxBytes = rx
	prevTxBytes = tx
	prevNetTime = now
}

func getPrimaryInterface() string {
	// Read default route from /proc/net/route
	f, err := os.Open("/proc/net/route")
	if err != nil {
		return ""
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	scanner.Scan() // skip header
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) >= 2 && fields[1] == "00000000" {
			return fields[0]
		}
	}
	return ""
}

func getNetBytes(iface string) (rx, tx uint64) {
	data, err := os.ReadFile("/proc/net/dev")
	if err != nil {
		return
	}
	for _, line := range strings.Split(string(data), "\n") {
		if strings.Contains(line, iface+":") {
			fields := strings.Fields(strings.TrimSpace(line))
			if len(fields) >= 10 {
				rx, _ = strconv.ParseUint(fields[1], 10, 64)
				tx, _ = strconv.ParseUint(fields[9], 10, 64)
			}
			return
		}
	}
	return
}

func getLocalIP() string {
	out, err := exec.Command("hostname", "-I").Output()
	if err != nil {
		return ""
	}
	parts := strings.Fields(string(out))
	if len(parts) > 0 {
		return parts[0]
	}
	return ""
}

func checkInternet() bool {
	out, err := exec.Command("ping", "-c", "1", "-W", "2", "8.8.8.8").Output()
	return err == nil && strings.Contains(string(out), "1 received")
}

func collectBattery(m *Metrics) {
	// Try /sys/class/power_supply
	entries, err := os.ReadDir("/sys/class/power_supply")
	if err != nil {
		return
	}
	for _, e := range entries {
		name := e.Name()
		if !strings.HasPrefix(name, "BAT") {
			continue
		}
		base := "/sys/class/power_supply/" + name
		capData, err := os.ReadFile(base + "/capacity")
		if err != nil {
			continue
		}
		level, _ := strconv.Atoi(strings.TrimSpace(string(capData)))
		m.BatteryLevel = level

		statusData, _ := os.ReadFile(base + "/status")
		status := strings.TrimSpace(string(statusData))
		m.BatteryCharging = status == "Charging" || status == "Full"
		return
	}
}

// GetOSName returns the current OS name
func GetOSName() string {
	return runtime.GOOS
}

// GetHostname returns the system hostname
func GetHostname() string {
	h, _ := os.Hostname()
	return h
}
