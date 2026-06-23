//go:build windows

package system

import (
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// Collect gathers all metrics for Windows
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
	out, err := exec.Command("powershell", "-Command",
		"(Get-CimInstance Win32_Processor | Measure-Object -Property LoadPercentage -Average).Average").Output()
	if err != nil {
		return 0
	}
	n, _ := strconv.ParseFloat(strings.TrimSpace(string(out)), 64)
	return n
}

func getCPUTemp() float64 {
	out, err := exec.Command("powershell", "-Command",
		"(Get-CimInstance -Namespace root/WMI -ClassName MSAcpi_ThermalZoneTemperature).CurrentTemperature").Output()
	if err != nil {
		return 0
	}
	n, _ := strconv.ParseFloat(strings.TrimSpace(string(out)), 64)
	return (n/10.0 - 273.15)
}

func getCPUFreq() float64 {
	out, err := exec.Command("powershell", "-Command",
		"(Get-CimInstance Win32_Processor).CurrentClockSpeed").Output()
	if err != nil {
		return 0
	}
	n, _ := strconv.ParseFloat(strings.TrimSpace(string(out)), 64)
	return n
}

func collectRAM(m *Metrics) {
	out, err := exec.Command("powershell", "-Command",
		"$os=Get-CimInstance Win32_OperatingSystem; Write-Output ($os.TotalVisibleMemorySize/1024); Write-Output ($os.FreePhysicalMemory/1024)").Output()
	if err != nil {
		return
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) >= 2 {
		total, _ := strconv.ParseUint(strings.TrimSpace(lines[0]), 10, 64)
		free, _ := strconv.ParseUint(strings.TrimSpace(lines[1]), 10, 64)
		m.RAMTotalMB = total
		m.RAMFreeMB = free
		m.RAMUsedMB = total - free
	}
}

func getDiskInfo() []DiskEntry {
	out, err := exec.Command("powershell", "-Command",
		"Get-PSDrive -PSProvider FileSystem | Select-Object Name,Used,Free | Format-Table -HideTableHeaders").Output()
	if err != nil {
		return nil
	}
	var entries []DiskEntry
	for _, line := range strings.Split(string(out), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}
		used, _ := strconv.ParseFloat(fields[1], 64)
		free, _ := strconv.ParseFloat(fields[2], 64)
		total := used + free
		if total == 0 {
			continue
		}
		entries = append(entries, DiskEntry{
			Mount:    fields[0] + ":",
			TotalGB:  total / 1e9,
			UsedGB:   used / 1e9,
			FreeGB:   free / 1e9,
			UsagePct: used / total * 100,
		})
	}
	return entries
}

var prevRxBytes, prevTxBytes uint64
var prevNetTime time.Time

func collectNetwork(m *Metrics) {
	out, err := exec.Command("powershell", "-Command",
		"$a=Get-NetAdapterStatistics; $a | Select-Object ReceivedBytes,SentBytes | Format-Table -HideTableHeaders").Output()
	if err != nil {
		return
	}
	var rx, tx uint64
	for _, line := range strings.Split(string(out), "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			r, _ := strconv.ParseUint(fields[0], 10, 64)
			t, _ := strconv.ParseUint(fields[1], 10, 64)
			rx += r
			tx += t
		}
	}
	now := time.Now()
	if !prevNetTime.IsZero() {
		dt := now.Sub(prevNetTime).Seconds()
		if dt > 0 {
			m.UploadKBps = float64(tx-prevTxBytes) / dt / 1024
			m.DownloadKBps = float64(rx-prevRxBytes) / dt / 1024
		}
	}
	prevRxBytes = rx
	prevTxBytes = tx
	prevNetTime = now
}

func getLocalIP() string {
	out, err := exec.Command("powershell", "-Command",
		"(Get-NetIPAddress -AddressFamily IPv4 | Where-Object {$_.InterfaceAlias -notmatch 'Loopback'} | Select-Object -First 1).IPAddress").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func checkInternet() bool {
	_, err := exec.Command("ping", "-n", "1", "-w", "2000", "8.8.8.8").Output()
	return err == nil
}

func collectBattery(m *Metrics) {
	out, err := exec.Command("powershell", "-Command",
		"$b=Get-CimInstance Win32_Battery; if($b){Write-Output $b.EstimatedChargeRemaining; Write-Output $b.BatteryStatus}").Output()
	if err != nil || strings.TrimSpace(string(out)) == "" {
		m.BatteryLevel = -1
		return
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) >= 1 {
		level, _ := strconv.Atoi(strings.TrimSpace(lines[0]))
		m.BatteryLevel = level
	}
	if len(lines) >= 2 {
		status, _ := strconv.Atoi(strings.TrimSpace(lines[1]))
		m.BatteryCharging = status == 2
	}
}

func GetOSName() string {
	return runtime.GOOS
}

func GetHostname() string {
	h, _ := os.Hostname()
	return h
}
