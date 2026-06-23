package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"runtime"
	"time"

	"github.com/joho/godotenv"
	"github.com/mustofa/commender/agent/client"
	"github.com/mustofa/commender/agent/config"
	"github.com/mustofa/commender/agent/system"
)

const AgentVersion = "1.0.0"

// Credentials are stored locally after registration
type Credentials struct {
	DeviceID string `json:"device_id"`
	Secret   string `json:"secret"`
}

func loadOrRegister(cfg *config.Config) (deviceID, secret string) {
	// Try loading from credentials file
	if data, err := os.ReadFile(cfg.CredentialsFile); err == nil {
		var creds Credentials
		if err := json.Unmarshal(data, &creds); err == nil && creds.DeviceID != "" {
			log.Printf("✅ Loaded credentials: device_id=%s", creds.DeviceID)
			return creds.DeviceID, creds.Secret
		}
	}

	// Register new device
	log.Println("📡 Registering device with server...")
	osName := runtime.GOOS
	hostname := system.GetHostname()

	devID, sec, err := client.Register(
		cfg.ServerURL,
		cfg.DeviceName,
		osName,
		hostname,
		AgentVersion,
	)
	if err != nil {
		log.Fatalf("❌ Registration failed: %v", err)
	}

	// Save credentials
	creds := Credentials{DeviceID: devID, Secret: sec}
	data, _ := json.MarshalIndent(creds, "", "  ")
	if err := os.WriteFile(cfg.CredentialsFile, data, 0600); err != nil {
		log.Printf("⚠️  Failed to save credentials: %v", err)
	}

	log.Printf("✅ Registered! device_id=%s", devID)
	return devID, sec
}

func main() {
	_ = godotenv.Load()
	config.Load()
	cfg := config.Cfg

	log.Printf("🤖 Workspace Agent v%s starting...", AgentVersion)
	log.Printf("   Server: %s", cfg.ServerURL)
	log.Printf("   Device: %s", cfg.DeviceName)

	deviceID, secret := loadOrRegister(cfg)
	cfg.DeviceID = deviceID
	cfg.Secret = secret

	c := client.New(cfg.ServerURL, deviceID, secret)

	// Start heartbeat goroutine
	go heartbeatLoop(c, cfg)

	// Start command polling loop (main loop)
	log.Printf("🔄 Polling for commands every %ds...", cfg.PollIntervalSeconds)
	for {
		commands, err := c.PollCommands()
		if err != nil {
			log.Printf("⚠️  Poll error: %v", err)
		} else {
			for _, cmd := range commands {
				go handleCommand(c, cmd)
			}
		}
		time.Sleep(time.Duration(cfg.PollIntervalSeconds) * time.Second)
	}
}

func heartbeatLoop(c *client.Client, cfg *config.Config) {
	for {
		metrics := system.Collect()
		localIP := metrics.LocalIP

		if err := c.Heartbeat(localIP, metrics); err != nil {
			log.Printf("⚠️  Heartbeat error: %v", err)
		}
		time.Sleep(time.Duration(cfg.HeartbeatIntervalSeconds) * time.Second)
	}
}

func handleCommand(c *client.Client, cmd client.Command) {
	log.Printf("📨 Executing command: %s (id=%d)", cmd.Type, cmd.ID)

	var result interface{}
	var status = "done"
	var execErr error

	switch cmd.Type {
	case "shutdown", "restart", "sleep", "lock", "logout":
		execErr = system.Execute(cmd.Type)
		if execErr != nil {
			result = map[string]string{"error": execErr.Error()}
			status = "failed"
		} else {
			result = map[string]string{"message": cmd.Type + " executed"}
		}

	case "screenshot":
		var payload struct {
			Monitor string `json:"monitor"`
		}
		_ = json.Unmarshal([]byte(cmd.Payload), &payload)
		imgBase64, err := system.TakeScreenshot(payload.Monitor)
		if err != nil {
			result = map[string]string{"error": err.Error()}
			status = "failed"
		} else {
			result = map[string]interface{}{
				"image":  imgBase64,
				"format": "png",
			}
		}

	case "media":
		var payload struct {
			Action string `json:"action"`
		}
		_ = json.Unmarshal([]byte(cmd.Payload), &payload)
		execErr = system.ControlMedia(payload.Action)
		if execErr != nil {
			result = map[string]string{"error": execErr.Error()}
			status = "failed"
		} else {
			result = map[string]string{"message": "media " + payload.Action}
		}

	case "list_dir":
		var payload struct {
			Path string `json:"path"`
		}
		_ = json.Unmarshal([]byte(cmd.Payload), &payload)
		listing := system.ListDir(payload.Path)
		result = listing

	case "search":
		var payload struct {
			Root    string `json:"root"`
			Pattern string `json:"pattern"`
		}
		_ = json.Unmarshal([]byte(cmd.Payload), &payload)
		matches, err := system.SearchFiles(payload.Root, payload.Pattern)
		if err != nil {
			result = map[string]string{"error": err.Error()}
			status = "failed"
		} else {
			result = map[string]interface{}{"matches": matches, "count": len(matches)}
		}

	case "status":
		metrics := system.Collect()
		result = metrics

	default:
		result = map[string]string{"error": fmt.Sprintf("unknown command: %s", cmd.Type)}
		status = "failed"
	}

	if err := c.ReportResult(cmd.ID, status, result); err != nil {
		log.Printf("⚠️  Failed to report result for cmd %d: %v", cmd.ID, err)
	}
	log.Printf("✅ Command %d (%s) → %s", cmd.ID, cmd.Type, status)
}
