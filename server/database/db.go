package database

import (
	"log"
	"os"
	"path/filepath"

	"github.com/mustofa/commender/server/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Init(dbPath string) {
	// Ensure directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Fatalf("failed to create database directory: %v", err)
	}

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	// Auto migrate all models
	if err := db.AutoMigrate(
		&models.Device{},
		&models.Command{},
		&models.Metric{},
		&models.Notification{},
	); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	DB = db
	log.Println("✅ Database connected and migrated")
}
