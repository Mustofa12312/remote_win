//go:build windows

package system

import "os/exec"

// Execute runs a system control command on Windows
func Execute(cmdType string) error {
	switch cmdType {
	case "shutdown":
		return exec.Command("shutdown", "/s", "/t", "0").Run()
	case "restart":
		return exec.Command("shutdown", "/r", "/t", "0").Run()
	case "sleep":
		return exec.Command("rundll32.exe", "powrprof.dll,SetSuspendState", "0,1,0").Run()
	case "lock":
		return exec.Command("rundll32.exe", "user32.dll,LockWorkStation").Run()
	case "logout":
		return exec.Command("shutdown", "/l").Run()
	}
	return nil
}

// ControlMedia sends media key events on Windows via PowerShell
func ControlMedia(action string) error {
	psMap := map[string]string{
		"play":     `Add-Type -AssemblyName System.Windows.Forms; [System.Windows.Forms.SendKeys]::SendWait('%{PAUSE}')`,
		"pause":    `Add-Type -AssemblyName System.Windows.Forms; [System.Windows.Forms.SendKeys]::SendWait('%{PAUSE}')`,
		"next":     `$wsh = New-Object -com WScript.Shell; $wsh.SendKeys([char]0xB0)`,
		"prev":     `$wsh = New-Object -com WScript.Shell; $wsh.SendKeys([char]0xB1)`,
		"vol_up":   `$wsh = New-Object -com WScript.Shell; $wsh.SendKeys([char]0xAF)`,
		"vol_down": `$wsh = New-Object -com WScript.Shell; $wsh.SendKeys([char]0xAE)`,
		"mute":     `$wsh = New-Object -com WScript.Shell; $wsh.SendKeys([char]0xAD)`,
	}
	ps, ok := psMap[action]
	if !ok {
		return nil
	}
	return exec.Command("powershell", "-Command", ps).Run()
}
