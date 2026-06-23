//go:build linux || darwin

package system

import (
	"os/exec"
	"runtime"
	"strings"
)

// Execute runs a system control command
func Execute(cmdType string) error {
	switch cmdType {
	case "shutdown":
		return exec.Command("shutdown", "-h", "now").Run()
	case "restart":
		return exec.Command("shutdown", "-r", "now").Run()
	case "sleep":
		if runtime.GOOS == "darwin" {
			return exec.Command("pmset", "sleepnow").Run()
		}
		// Try systemctl then pm-suspend
		if err := exec.Command("systemctl", "suspend").Run(); err != nil {
			return exec.Command("pm-suspend").Run()
		}
		return nil
	case "lock":
		// Try common Linux lock commands
		candidates := [][]string{
			{"gnome-screensaver-command", "--lock"},
			{"loginctl", "lock-session"},
			{"xdg-screensaver", "lock"},
			{"xlock"},
		}
		for _, cmd := range candidates {
			if err := exec.Command(cmd[0], cmd[1:]...).Run(); err == nil {
				return nil
			}
		}
		return exec.Command("loginctl", "lock-session").Run()
	case "logout":
		if runtime.GOOS == "darwin" {
			return exec.Command("osascript", "-e", `tell application "System Events" to log out`).Run()
		}
		// Try gnome then kde
		if err := exec.Command("gnome-session-quit", "--no-prompt").Run(); err != nil {
			return exec.Command("qdbus", "org.kde.ksmserver", "/KSMServer", "logout", "0", "0", "0").Run()
		}
		return nil
	}
	return nil
}

// ControlMedia sends media key events
func ControlMedia(action string) error {
	key := map[string]string{
		"play":    "XF86AudioPlay",
		"pause":   "XF86AudioPause",
		"next":    "XF86AudioNext",
		"prev":    "XF86AudioPrev",
		"vol_up":  "XF86AudioRaiseVolume",
		"vol_down": "XF86AudioLowerVolume",
		"mute":    "XF86AudioMute",
	}[action]

	if key == "" {
		// Also support play/pause toggle
		if action == "play" || action == "pause" {
			key = "XF86AudioPlay"
		}
	}

	if key == "" {
		return nil
	}

	// Try xdotool, then playerctl
	if err := exec.Command("xdotool", "key", key).Run(); err != nil {
		// Try playerctl
		playerAction := strings.ReplaceAll(action, "vol_up", "volume 0.1+")
		playerAction = strings.ReplaceAll(playerAction, "vol_down", "volume 0.1-")
		if action == "play" || action == "pause" {
			playerAction = "play-pause"
		} else if action == "next" {
			playerAction = "next"
		} else if action == "prev" {
			playerAction = "previous"
		} else if action == "mute" {
			playerAction = "volume 0"
		}
		_ = exec.Command("playerctl", playerAction).Run()
	}
	return nil
}
