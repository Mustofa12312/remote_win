package handlers

import (
	"log"
	"strconv"
	"time"

	"github.com/mustofa/commender/server/bot"
	"github.com/mustofa/commender/server/database"
	"github.com/mustofa/commender/server/models"
	tele "gopkg.in/telebot.v3"
)

const (
	DiskAlertThresholdGB  = 5.0
	RAMAlertThresholdPct  = 90.0
	BatteryAlertThreshold = 15
)

// StartNotificationEngine runs background goroutines for alerting
func StartNotificationEngine(b *tele.Bot, ownerID int64) {
	go checkDeviceOffline(b, ownerID)
	go checkMetricAlerts(b, ownerID)
	go processCommandResults(b)
	log.Println("🔔 Notification engine started")
}

// checkDeviceOffline marks devices offline and sends alert
func checkDeviceOffline(b *tele.Bot, ownerID int64) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		var devices []models.Device
		database.DB.Where("status = ? AND last_seen < ?", "online", time.Now().Add(-2*time.Minute)).Find(&devices)
		for _, d := range devices {
			database.DB.Model(&d).Update("status", "offline")
			msg := "Device *" + d.Name + "* offline!"
			// Check if we already sent this alert recently
			var existing models.Notification
			err := database.DB.Where("device_id = ? AND type = ? AND created_at > ?",
				d.DeviceID, models.NotifTypeOffline, time.Now().Add(-10*time.Minute)).
				First(&existing).Error
			if err != nil { // not found = no recent alert
				notif := models.Notification{
					DeviceID: d.DeviceID,
					Type:     models.NotifTypeOffline,
					Message:  msg,
					Sent:     true,
				}
				database.DB.Create(&notif)
				bot.SendAlert(b, ownerID, msg)
			}
		}
	}
}

// checkMetricAlerts monitors metrics and sends threshold alerts
func checkMetricAlerts(b *tele.Bot, ownerID int64) {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		var devices []models.Device
		database.DB.Where("status = ?", "online").Find(&devices)

		for _, d := range devices {
			var metric models.Metric
			if err := database.DB.Where("device_id = ?", d.DeviceID).Order("recorded_at desc").First(&metric).Error; err != nil {
				continue
			}
			checkRAMAlert(b, ownerID, d, metric)
			checkBatteryAlert(b, ownerID, d, metric)
		}
	}
}

func checkRAMAlert(b *tele.Bot, ownerID int64, d models.Device, m models.Metric) {
	if m.RAMTotalMB == 0 {
		return
	}
	pct := float64(m.RAMUsedMB) / float64(m.RAMTotalMB) * 100
	if pct < RAMAlertThresholdPct {
		return
	}
	var existing models.Notification
	err := database.DB.Where("device_id = ? AND type = ? AND created_at > ?",
		d.DeviceID, models.NotifTypeRAM, time.Now().Add(-30*time.Minute)).First(&existing).Error
	if err != nil {
		msg := "⚠ RAM di atas 90% pada *" + d.Name + "* (" + d.Hostname + ")"
		database.DB.Create(&models.Notification{DeviceID: d.DeviceID, Type: models.NotifTypeRAM, Message: msg, Sent: true})
		bot.SendAlert(b, ownerID, msg)
	}
}

func checkBatteryAlert(b *tele.Bot, ownerID int64, d models.Device, m models.Metric) {
	if m.BatteryLevel < 0 || m.BatteryCharging {
		return
	}
	if m.BatteryLevel > BatteryAlertThreshold {
		return
	}
	var existing models.Notification
	err := database.DB.Where("device_id = ? AND type = ? AND created_at > ?",
		d.DeviceID, models.NotifTypeBattery, time.Now().Add(-15*time.Minute)).First(&existing).Error
	if err != nil {
		msg := "⚠ Baterai tinggal " + strconv.Itoa(m.BatteryLevel) + "% pada *" + d.Name + "*"
		database.DB.Create(&models.Notification{DeviceID: d.DeviceID, Type: models.NotifTypeBattery, Message: msg, Sent: true})
		bot.SendAlert(b, ownerID, msg)
	}
}

// processCommandResults polls for done commands and sends results to Telegram
func processCommandResults(b *tele.Bot) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		var commands []models.Command
		database.DB.Where("status IN (?) AND telegram_chat_id != 0",
			[]string{models.CommandStatusDone, models.CommandStatusFailed}).
			Where("result != ''").
			Find(&commands)

		for _, cmd := range commands {
			bot.HandleCommandResult(b, &cmd)
			// Mark as processed by clearing chat_id
			database.DB.Model(&cmd).Update("telegram_chat_id", 0)
		}
	}
}
