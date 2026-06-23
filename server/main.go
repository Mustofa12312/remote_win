package main

import (
	"log"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/mustofa/commender/server/bot"
	"github.com/mustofa/commender/server/config"
	"github.com/mustofa/commender/server/database"
	"github.com/mustofa/commender/server/handlers"
	"github.com/mustofa/commender/server/middleware"
)

func main() {
	// Load .env if exists
	_ = godotenv.Load()
	config.Load()

	// Init database
	database.Init(config.Cfg.DatabasePath)

	// Start Telegram bot
	teleBot, err := bot.Setup(config.Cfg.TelegramToken, config.Cfg.TelegramOwnerID)
	if err != nil {
		log.Printf("⚠️  Bot error: %v", err)
	}

	// Start notification engine
	handlers.StartNotificationEngine(teleBot, config.Cfg.TelegramOwnerID)

	// Gin router
	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.DebugMode)
	}
	r := gin.Default()

	// CORS — allow dashboard and dev tools
	r.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-API-Key"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: false,
	}))

	// ── Public Routes ──────────────────────────────────────
	r.POST("/api/devices/register", handlers.RegisterDevice)

	// ── Agent Routes (Device Auth) ─────────────────────────
	agent := r.Group("/api/agent")
	agent.Use(middleware.DeviceAuth())
	{
		agent.POST("/heartbeat", handlers.Heartbeat)
		agent.GET("/commands/poll", handlers.PollCommands)
		agent.POST("/commands/:id/result", handlers.ReportCommandResult)
	}

	// ── Dashboard Routes ───────────────────────────────────
	// In production, protect with ServerAuth. Dev mode: open.
	api := r.Group("/api")
	{
		api.GET("/devices", handlers.ListDevices)
		api.GET("/devices/:device_id", handlers.GetDevice)
		api.GET("/devices/:device_id/metrics", handlers.GetMetrics)
		api.GET("/commands/history", handlers.ListCommandHistory)
		api.POST("/commands", handlers.EnqueueCommand)
		api.GET("/notifications", handlers.ListNotifications)
	}

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "workspace-commander"})
	})

	log.Printf("🚀 Server running on :%s", config.Cfg.Port)
	if err := r.Run(":" + config.Cfg.Port); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
