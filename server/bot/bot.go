package bot

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/mustofa/commender/server/database"
	"github.com/mustofa/commender/server/models"
	tele "gopkg.in/telebot.v3"
)

// Session tracks per-user state
type Session struct {
	SelectedDevice *models.Device
}

var sessions = map[int64]*Session{}

func getSession(userID int64) *Session {
	if s, ok := sessions[userID]; ok {
		return s
	}
	sessions[userID] = &Session{}
	return sessions[userID]
}

var globalBot *tele.Bot
var globalOwnerID int64

// Setup initializes and starts the Telegram bot
func Setup(token string, ownerID int64) (*tele.Bot, error) {
	globalOwnerID = ownerID

	if token == "" {
		log.Println("⚠️  TELEGRAM_TOKEN not set, bot disabled")
		return nil, nil
	}

	pref := tele.Settings{
		Token:     token,
		ParseMode: tele.ModeMarkdown,
		Poller:    &tele.LongPoller{Timeout: 10 * time.Second},
	}

	b, err := tele.NewBot(pref)
	if err != nil {
		return nil, fmt.Errorf("failed to create bot: %w", err)
	}
	globalBot = b

	// Auth middleware
	authMiddleware := func(next tele.HandlerFunc) tele.HandlerFunc {
		return func(c tele.Context) error {
			if ownerID != 0 && c.Sender().ID != ownerID {
				return c.Send("⛔ Unauthorized")
			}
			return next(c)
		}
	}

	b.Use(authMiddleware)

	// Commands
	b.Handle("/start", handleStart)
	b.Handle("/help", handleStart)
	b.Handle("/devices", handleDeviceList)
	b.Handle("/status", handleStatus)
	b.Handle("/screenshot", handleScreenshotCmd)
	b.Handle("/shutdown", handleShutdown)
	b.Handle("/restart", handleRestart)
	b.Handle("/sleep", handleSleep)
	b.Handle("/lock", handleLock)
	b.Handle("/logout", handleLogoutCmd)
	b.Handle("/files", handleFiles)
	b.Handle("/media", handleMedia)

	// Callback queries (inline buttons)
	b.Handle(tele.OnCallback, handleCallback)

	log.Println("🤖 Telegram bot started")
	go b.Start()
	return b, nil
}

// ── Main Menu ──────────────────────────────────────────

func handleStart(c tele.Context) error {
	menu := buildMainMenu()
	return c.Send("👋 *Workspace Commander*\n\nPilih menu:", menu, tele.ModeMarkdown)
}

func buildMainMenu() *tele.ReplyMarkup {
	menu := &tele.ReplyMarkup{}
	btnStatus := menu.Data("📊 Status", "menu_status")
	btnDevices := menu.Data("💻 Devices", "menu_devices")
	btnFiles := menu.Data("📁 Files", "menu_files")
	btnScreenshot := menu.Data("📷 Screenshot", "menu_screenshot")
	btnMedia := menu.Data("🎵 Media", "menu_media")
	btnSystem := menu.Data("⚙️ System", "menu_system")
	menu.Inline(
		menu.Row(btnStatus, btnDevices),
		menu.Row(btnFiles, btnScreenshot),
		menu.Row(btnMedia, btnSystem),
	)
	return menu
}

// ── Device List ────────────────────────────────────────

func handleDeviceList(c tele.Context) error {
	return sendDeviceList(c, false)
}

func sendDeviceList(c tele.Context, isEdit bool) error {
	var devices []models.Device
	database.DB.Model(&models.Device{}).
		Where("status = ? AND last_seen < ?", "online", time.Now().Add(-2*time.Minute)).
		Update("status", "offline")
	database.DB.Find(&devices)

	if len(devices) == 0 {
		msg := "📭 Belum ada device terdaftar.\n\nJalankan agent di komputer target untuk mendaftarkan device."
		if isEdit {
			return c.Edit(msg)
		}
		return c.Send(msg)
	}

	menu := &tele.ReplyMarkup{}
	var rows []tele.Row
	for _, d := range devices {
		icon := "🔴"
		if d.Status == "online" {
			icon = "🟢"
		}
		label := fmt.Sprintf("%s %s (%s)", icon, d.Name, d.OS)
		btn := menu.Data(label, "select_device|"+d.DeviceID)
		rows = append(rows, menu.Row(btn))
	}
	btnBack := menu.Data("◀️ Back", "menu_main")
	rows = append(rows, menu.Row(btnBack))
	menu.Inline(rows...)

	text := "💻 *Daftar Device*\n\nPilih device yang ingin dikontrol:"
	if isEdit {
		return c.Edit(text, menu, tele.ModeMarkdown)
	}
	return c.Send(text, menu, tele.ModeMarkdown)
}

// ── Status ─────────────────────────────────────────────

func handleStatus(c tele.Context) error {
	sess := getSession(c.Sender().ID)
	if sess.SelectedDevice == nil {
		return c.Send("⚠️ Pilih device dulu dengan /devices")
	}
	return sendStatusForDevice(c, sess.SelectedDevice, false)
}

func sendStatusForDevice(c tele.Context, device *models.Device, isEdit bool) error {
	var metric models.Metric
	database.DB.Where("device_id = ?", device.DeviceID).Order("recorded_at desc").First(&metric)

	icon := "🔴"
	if device.Status == "online" {
		icon = "🟢"
	}

	battery := "🔋 N/A"
	if metric.BatteryLevel >= 0 {
		chargingStr := ""
		if metric.BatteryCharging {
			chargingStr = " ⚡"
		}
		battery = fmt.Sprintf("🔋 %d%%%s", metric.BatteryLevel, chargingStr)
	}

	text := fmt.Sprintf(
		"📊 *Status: %s*\n\n"+
			"%s Status: *%s*\n"+
			"🕒 Last seen: %s\n"+
			"🌐 IP: %s\n\n"+
			"🖥️ CPU: *%.1f%%*\n"+
			"🌡️ CPU Temp: *%.1f°C*\n"+
			"💾 RAM: *%d MB / %d MB*\n"+
			"%s\n"+
			"⬆️ Upload: *%.1f KB/s*\n"+
			"⬇️ Download: *%.1f KB/s*",
		device.Name,
		icon, device.Status,
		device.LastSeen.Local().Format("02 Jan 15:04:05"),
		device.LocalIP,
		metric.CPUUsage,
		metric.CPUTemp,
		metric.RAMUsedMB, metric.RAMTotalMB,
		battery,
		metric.UploadKBps,
		metric.DownloadKBps,
	)

	menu := &tele.ReplyMarkup{}
	btnRefresh := menu.Data("🔄 Refresh", "status|"+device.DeviceID)
	btnBack := menu.Data("◀️ Back", "menu_main")
	menu.Inline(menu.Row(btnRefresh, btnBack))

	if isEdit {
		return c.Edit(text, menu, tele.ModeMarkdown)
	}
	return c.Send(text, menu, tele.ModeMarkdown)
}

// ── System Control ─────────────────────────────────────

func handleShutdown(c tele.Context) error  { return sendConfirmation(c, "shutdown", "🛑 Shutdown") }
func handleRestart(c tele.Context) error   { return sendConfirmation(c, "restart", "🔄 Restart") }
func handleSleep(c tele.Context) error     { return sendConfirmation(c, "sleep", "😴 Sleep") }
func handleLock(c tele.Context) error      { return sendConfirmation(c, "lock", "🔒 Lock Screen") }
func handleLogoutCmd(c tele.Context) error { return sendConfirmation(c, "logout", "🚪 Logout User") }

func sendConfirmation(c tele.Context, cmdType, label string) error {
	sess := getSession(c.Sender().ID)
	if sess.SelectedDevice == nil {
		return c.Send("⚠️ Pilih device dulu dengan /devices")
	}
	menu := &tele.ReplyMarkup{}
	btnYes := menu.Data("✅ Ya, "+label, "confirm|"+cmdType+"|"+sess.SelectedDevice.DeviceID)
	btnNo := menu.Data("❌ Batal", "menu_main")
	menu.Inline(menu.Row(btnYes, btnNo))
	return c.Send(
		fmt.Sprintf("⚠️ *Are you sure?*\n\nDevice: *%s*\nAksi: *%s*", sess.SelectedDevice.Name, label),
		menu, tele.ModeMarkdown,
	)
}

// ── Screenshot ─────────────────────────────────────────

func handleScreenshotCmd(c tele.Context) error {
	sess := getSession(c.Sender().ID)
	if sess.SelectedDevice == nil {
		return c.Send("⚠️ Pilih device dulu dengan /devices")
	}
	return enqueueAndNotify(c, sess.SelectedDevice.DeviceID, "screenshot", "")
}

// ── Media Control ──────────────────────────────────────

func handleMedia(c tele.Context) error {
	sess := getSession(c.Sender().ID)
	if sess.SelectedDevice == nil {
		return c.Send("⚠️ Pilih device dulu dengan /devices")
	}
	return sendMediaMenu(c, sess.SelectedDevice.DeviceID, false)
}

func sendMediaMenu(c tele.Context, deviceID string, isEdit bool) error {
	menu := &tele.ReplyMarkup{}
	btnPlay := menu.Data("▶️", "media|play|"+deviceID)
	btnPause := menu.Data("⏸", "media|pause|"+deviceID)
	btnNext := menu.Data("⏭", "media|next|"+deviceID)
	btnPrev := menu.Data("⏮", "media|prev|"+deviceID)
	btnVolUp := menu.Data("🔊 Vol+", "media|vol_up|"+deviceID)
	btnVolDown := menu.Data("🔉 Vol-", "media|vol_down|"+deviceID)
	btnMute := menu.Data("🔇 Mute", "media|mute|"+deviceID)
	btnBack := menu.Data("◀️ Back", "menu_main")
	menu.Inline(
		menu.Row(btnPrev, btnPlay, btnPause, btnNext),
		menu.Row(btnVolDown, btnMute, btnVolUp),
		menu.Row(btnBack),
	)
	text := "🎵 *Media Control*"
	if isEdit {
		return c.Edit(text, menu, tele.ModeMarkdown)
	}
	return c.Send(text, menu, tele.ModeMarkdown)
}

// ── Files ──────────────────────────────────────────────

func handleFiles(c tele.Context) error {
	sess := getSession(c.Sender().ID)
	if sess.SelectedDevice == nil {
		return c.Send("⚠️ Pilih device dulu dengan /devices")
	}
	return sendFileBrowserRequest(c, sess.SelectedDevice.DeviceID, "", false)
}

func sendFileBrowserRequest(c tele.Context, deviceID, path string, isEdit bool) error {
	payload, _ := json.Marshal(map[string]string{"path": path})
	enqueueCommand(deviceID, "list_dir", string(payload), c.Chat().ID, 0)
	displayPath := path
	if displayPath == "" {
		displayPath = "/"
	}
	text := fmt.Sprintf("📁 *File Browser*\n\nPath: `%s`\n\n⏳ Memuat direktori...", displayPath)
	if isEdit {
		return c.Edit(text, tele.ModeMarkdown)
	}
	return c.Send(text, tele.ModeMarkdown)
}

// ── Callback Router ────────────────────────────────────

func handleCallback(c tele.Context) error {
	data := c.Data()
	parts := strings.SplitN(data, "|", 4)
	action := parts[0]

	switch action {
	case "menu_main":
		menu := buildMainMenu()
		return c.Edit("👋 *Workspace Commander*\n\nPilih menu:", menu, tele.ModeMarkdown)

	case "menu_devices":
		return sendDeviceList(c, true)

	case "menu_status":
		sess := getSession(c.Sender().ID)
		if sess.SelectedDevice == nil {
			return sendDeviceList(c, true)
		}
		return sendStatusForDevice(c, sess.SelectedDevice, true)

	case "menu_screenshot":
		sess := getSession(c.Sender().ID)
		if sess.SelectedDevice == nil {
			return sendDeviceList(c, true)
		}
		c.Respond()
		return enqueueAndNotify(c, sess.SelectedDevice.DeviceID, "screenshot", "")

	case "menu_media":
		sess := getSession(c.Sender().ID)
		if sess.SelectedDevice == nil {
			return sendDeviceList(c, true)
		}
		return sendMediaMenu(c, sess.SelectedDevice.DeviceID, true)

	case "menu_system":
		return sendSystemMenu(c)

	case "menu_files":
		sess := getSession(c.Sender().ID)
		if sess.SelectedDevice == nil {
			return sendDeviceList(c, true)
		}
		return sendFileBrowserRequest(c, sess.SelectedDevice.DeviceID, "", true)

	case "select_device":
		if len(parts) < 2 {
			return c.Respond()
		}
		deviceID := parts[1]
		var device models.Device
		if err := database.DB.Where("device_id = ?", deviceID).First(&device).Error; err != nil {
			return c.Edit("❌ Device tidak ditemukan")
		}
		sess := getSession(c.Sender().ID)
		sess.SelectedDevice = &device
		return sendStatusForDevice(c, &device, true)

	case "status":
		if len(parts) < 2 {
			return c.Respond()
		}
		deviceID := parts[1]
		var device models.Device
		database.DB.Where("device_id = ?", deviceID).First(&device)
		return sendStatusForDevice(c, &device, true)

	case "confirm":
		if len(parts) < 3 {
			return c.Respond()
		}
		cmdType := parts[1]
		deviceID := parts[2]
		_ = c.Respond(&tele.CallbackResponse{Text: "⏳ Mengirim perintah..."})
		enqueueCommand(deviceID, cmdType, "", c.Chat().ID, 0)
		return c.Edit(fmt.Sprintf("✅ Perintah *%s* dikirim ke device.", cmdType), tele.ModeMarkdown)

	case "media":
		if len(parts) < 3 {
			return c.Respond()
		}
		mediaAction := parts[1]
		deviceID := parts[2]
		payload, _ := json.Marshal(map[string]string{"action": mediaAction})
		_ = c.Respond(&tele.CallbackResponse{Text: "✅ " + mediaAction})
		enqueueCommand(deviceID, "media", string(payload), c.Chat().ID, 0)
		return nil
	}

	return c.Respond()
}

// ── System Menu ────────────────────────────────────────

func sendSystemMenu(c tele.Context) error {
	sess := getSession(c.Sender().ID)
	deviceID := ""
	if sess.SelectedDevice != nil {
		deviceID = sess.SelectedDevice.DeviceID
	}

	menu := &tele.ReplyMarkup{}
	btnShutdown := menu.Data("🛑 Shutdown", "confirm|shutdown|"+deviceID)
	btnRestart := menu.Data("🔄 Restart", "confirm|restart|"+deviceID)
	btnSleep := menu.Data("😴 Sleep", "confirm|sleep|"+deviceID)
	btnLock := menu.Data("🔒 Lock", "confirm|lock|"+deviceID)
	btnLogout := menu.Data("🚪 Logout", "confirm|logout|"+deviceID)
	btnBack := menu.Data("◀️ Back", "menu_main")
	menu.Inline(
		menu.Row(btnShutdown, btnRestart),
		menu.Row(btnSleep, btnLock, btnLogout),
		menu.Row(btnBack),
	)
	return c.Edit("⚙️ *System Control*\n\nPerintah akan meminta konfirmasi:", menu, tele.ModeMarkdown)
}

// ── Helpers ────────────────────────────────────────────

func enqueueCommand(deviceID, cmdType, payload string, chatID int64, msgID int) uint {
	cmd := models.Command{
		DeviceID:       deviceID,
		Type:           cmdType,
		Payload:        payload,
		Status:         models.CommandStatusPending,
		TelegramChatID: chatID,
		TelegramMsgID:  msgID,
	}
	database.DB.Create(&cmd)
	return cmd.ID
}

func enqueueAndNotify(c tele.Context, deviceID, cmdType, payload string) error {
	enqueueCommand(deviceID, cmdType, payload, c.Chat().ID, 0)
	return c.Send(fmt.Sprintf("⏳ Perintah *%s* dikirim. Menunggu respons dari agent...", cmdType), tele.ModeMarkdown)
}

// ── Result Handler ────────────────────────────────────

// HandleCommandResult processes completed commands and sends results to Telegram
func HandleCommandResult(b *tele.Bot, cmd *models.Command) {
	if b == nil || cmd.TelegramChatID == 0 {
		return
	}

	chat := &tele.Chat{ID: cmd.TelegramChatID}

	switch cmd.Type {
	case "screenshot":
		var res struct {
			Image  string `json:"image"` // base64
			Format string `json:"format"`
			Error  string `json:"error"`
		}
		if err := json.Unmarshal([]byte(cmd.Result), &res); err != nil || res.Error != "" {
			errMsg := "❌ Screenshot gagal"
			if res.Error != "" {
				errMsg += ": " + res.Error
			}
			b.Send(chat, errMsg)
			return
		}
		imgBytes, err := base64.StdEncoding.DecodeString(res.Image)
		if err != nil {
			b.Send(chat, "❌ Gagal decode screenshot")
			return
		}
		photo := &tele.Photo{
			File:    tele.FromReader(bytes.NewReader(imgBytes)),
			Caption: "📷 Screenshot",
		}
		b.Send(chat, photo)

	case "list_dir":
		var res struct {
			Path    string   `json:"path"`
			Entries []struct {
				Name  string `json:"name"`
				IsDir bool   `json:"is_dir"`
				Size  int64  `json:"size"`
			} `json:"entries"`
			Error string `json:"error"`
		}
		if err := json.Unmarshal([]byte(cmd.Result), &res); err != nil {
			b.Send(chat, "❌ Gagal membaca direktori")
			return
		}
		if res.Error != "" {
			b.Send(chat, "❌ "+res.Error)
			return
		}
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("📁 *%s*\n\n", res.Path))
		for _, e := range res.Entries {
			if e.IsDir {
				sb.WriteString(fmt.Sprintf("📂 %s\n", e.Name))
			} else {
				sb.WriteString(fmt.Sprintf("📄 %s (%s)\n", e.Name, formatBytes(e.Size)))
			}
		}
		b.Send(chat, sb.String(), tele.ModeMarkdown)

	default:
		if cmd.Status == models.CommandStatusDone {
			b.Send(chat, fmt.Sprintf("✅ Perintah *%s* berhasil dijalankan.", cmd.Type), tele.ModeMarkdown)
		} else {
			b.Send(chat, fmt.Sprintf("❌ Perintah *%s* gagal: %s", cmd.Type, cmd.Result), tele.ModeMarkdown)
		}
	}
}

// SendAlert sends an alert notification to the owner
func SendAlert(b *tele.Bot, ownerID int64, message string) {
	if b == nil || ownerID == 0 {
		return
	}
	chat := &tele.Chat{ID: ownerID}
	b.Send(chat, "⚠️ *ALERT*\n\n"+message, tele.ModeMarkdown)
}

func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return strconv.FormatInt(b, 10) + " B"
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}
