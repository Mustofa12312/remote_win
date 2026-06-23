package system

import (
	"encoding/base64"
	"fmt"
	"os/exec"
	"strings"
)

// TakeScreenshot captures the screen and returns base64-encoded PNG
func TakeScreenshot(monitor string) (string, error) {
	// Try scrot (Linux), then import (ImageMagick), then gnome-screenshot
	tmpFile := "/tmp/ws_screenshot.png"

	var err error

	// Try scrot
	if monitor == "" || monitor == "all" {
		err = exec.Command("scrot", "--overwrite", tmpFile).Run()
	} else {
		// Parse monitor index
		idx := 0
		fmt.Sscanf(monitor, "%d", &idx)
		err = exec.Command("scrot", "--overwrite", "--monitor", fmt.Sprintf("%d", idx), tmpFile).Run()
	}

	if err != nil {
		// Try import (ImageMagick)
		err = exec.Command("import", "-window", "root", tmpFile).Run()
	}

	if err != nil {
		// Try gnome-screenshot
		err = exec.Command("gnome-screenshot", "-f", tmpFile).Run()
	}

	if err != nil {
		// Try spectacle (KDE)
		err = exec.Command("spectacle", "-b", "-n", "-o", tmpFile).Run()
	}

	if err != nil {
		return "", fmt.Errorf("no screenshot tool available: %v", err)
	}

	// Read the file and encode as base64
	data, err := readFile(tmpFile)
	if err != nil {
		return "", fmt.Errorf("failed to read screenshot: %v", err)
	}

	return base64.StdEncoding.EncodeToString(data), nil
}

// GetMonitorCount returns the number of connected monitors (Linux)
func GetMonitorCount() int {
	out, err := exec.Command("xrandr", "--query").Output()
	if err != nil {
		return 1
	}
	count := strings.Count(string(out), " connected")
	if count == 0 {
		return 1
	}
	return count
}
