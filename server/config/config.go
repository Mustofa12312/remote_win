package config

import (
	"fmt"
	"os"
)

type Config struct {
	Port         string
	DatabasePath string
	TelegramToken string
	TelegramOwnerID int64
	JWTSecret    string
	ServerURL    string
}

var Cfg *Config

func Load() {
	Cfg = &Config{
		Port:         getEnv("PORT", "8080"),
		DatabasePath: getEnv("DATABASE_PATH", "./data/commender.db"),
		TelegramToken: getEnv("TELEGRAM_TOKEN", ""),
		TelegramOwnerID: getEnvInt64("TELEGRAM_OWNER_ID", 0),
		JWTSecret:    getEnv("JWT_SECRET", "change-me-in-production"),
		ServerURL:    getEnv("SERVER_URL", "http://localhost:8080"),
	}
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getEnvInt64(key string, def int64) int64 {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	var result int64
	_, err := fmt.Sscanf(v, "%d", &result)
	if err != nil {
		return def
	}
	return result
}
