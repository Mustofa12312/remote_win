package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	ServerURL              string
	DeviceName             string
	DeviceOS               string
	DeviceID               string // loaded after registration
	Secret                 string // loaded after registration
	PollIntervalSeconds    int
	HeartbeatIntervalSeconds int
	CredentialsFile        string
}

var Cfg *Config

func Load() {
	Cfg = &Config{
		ServerURL:              getEnv("SERVER_URL", "http://localhost:8080"),
		DeviceName:             getEnv("DEVICE_NAME", mustHostname()),
		DeviceOS:               getEnv("DEVICE_OS", detectOS()),
		PollIntervalSeconds:    getEnvInt("POLL_INTERVAL_SECONDS", 5),
		HeartbeatIntervalSeconds: getEnvInt("HEARTBEAT_INTERVAL_SECONDS", 30),
		CredentialsFile:        getEnv("CREDENTIALS_FILE", "./credentials.json"),
	}
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getEnvInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		n, err := strconv.Atoi(v)
		if err == nil {
			return n
		}
	}
	return def
}

func mustHostname() string {
	h, err := os.Hostname()
	if err != nil {
		return "unknown-host"
	}
	return h
}

func detectOS() string {
	// Will be overridden at compile time via GOOS detection in main.go
	return fmt.Sprintf("%s", os.Getenv("GOOS"))
}
