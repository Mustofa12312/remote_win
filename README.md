# Workspace Commander

> Remote management ringan berbasis Telegram + Web Dashboard untuk mengontrol komputer Windows/Ubuntu dari HP atau browser.

---

## Arsitektur

```
[Telegram] ──► [Workspace Server (Go)] ──► SQLite DB
                      │
          ┌───────────┴────────────┐
          ▼                        ▼
 [Ubuntu Agent]           [Windows Agent]
 (workspace-agent)        (workspace-agent.exe)
```

**Agent komunikasi via polling HTTP** — agent membuka koneksi ke server, tidak perlu port forwarding di sisi device.

---

## Komponen

| Komponen | Direktori | Teknologi |
|---|---|---|
| Server | `server/` | Go + Gin + GORM + SQLite |
| Agent | `agent/` | Go (single binary) |
| Dashboard | `dashboard/` | Flutter Web |

---

## Quick Start

### 1. Setup Server

```bash
cd server
cp .env.example .env
# Edit .env: isi TELEGRAM_TOKEN dan TELEGRAM_OWNER_ID
nano .env
```

**Cara dapat Telegram Bot Token:**
1. Chat ke [@BotFather](https://t.me/botfather) di Telegram
2. Kirim `/newbot`, ikuti instruksi
3. Copy token ke `TELEGRAM_TOKEN`

**Cara dapat Telegram Owner ID:**
1. Chat ke [@userinfobot](https://t.me/userinfobot)
2. Copy angka ID ke `TELEGRAM_OWNER_ID`

### 2. Jalankan Server

```bash
make run-server
# atau
./bin/workspace-server
```

Server berjalan di `http://localhost:8080`

### 3. Setup Agent di komputer target

```bash
cd agent
cp .env.example .env
# Edit .env: isi SERVER_URL dengan IP/domain server
nano .env
```

### 4. Jalankan Agent

```bash
make run-agent
# atau
./bin/workspace-agent
```

Agent otomatis mendaftar ke server dan mulai mengirim heartbeat.

### 5. Gunakan via Telegram

Buka bot Telegram Anda dan ketik `/start` untuk melihat menu.

---

## Build

```bash
# Build semua
make all

# Build server saja
make server

# Build agent untuk Linux
make agent

# Cross-compile agent untuk Windows
make agent-windows
```

---

## API Endpoints

### Public
| Method | Endpoint | Deskripsi |
|---|---|---|
| POST | `/api/devices/register` | Registrasi device baru |
| GET | `/health` | Health check |

### Agent (auth: Bearer device_id:secret)
| Method | Endpoint | Deskripsi |
|---|---|---|
| POST | `/api/agent/heartbeat` | Kirim status + metrics |
| GET | `/api/agent/commands/poll` | Ambil command pending |
| POST | `/api/agent/commands/:id/result` | Lapor hasil command |

### Dashboard
| Method | Endpoint | Deskripsi |
|---|---|---|
| GET | `/api/devices` | List semua device |
| GET | `/api/devices/:id` | Detail device + metrics |
| GET | `/api/devices/:id/metrics` | History metrics |
| POST | `/api/commands` | Kirim command |
| GET | `/api/commands/history` | Riwayat command |
| GET | `/api/notifications` | Riwayat notifikasi |

---

## Command Types

| Type | Payload | Deskripsi |
|---|---|---|
| `shutdown` | - | Matikan perangkat |
| `restart` | - | Restart perangkat |
| `sleep` | - | Sleep/suspend |
| `lock` | - | Lock screen |
| `logout` | - | Logout user |
| `screenshot` | `{"monitor":""}` | Ambil screenshot |
| `media` | `{"action":"play\|pause\|next\|prev\|vol_up\|vol_down\|mute"}` | Media control |
| `list_dir` | `{"path":"/home"}` | List direktori |
| `search` | `{"root":"/","pattern":"laporan"}` | Cari file |
| `status` | - | Ambil metrics realtime |

---

## Notification Alerts

Server otomatis mengirim alert ke Telegram jika:
- Device offline > 2 menit
- RAM > 90%
- Baterai < 15% (dan tidak charging)
- Disk < 5 GB (coming soon)

---

## Struktur File

```
commender/
├── server/              # Go server
│   ├── main.go
│   ├── config/
│   ├── database/
│   ├── models/
│   ├── handlers/
│   ├── middleware/
│   ├── bot/             # Telegram bot
│   └── .env.example
│
├── agent/               # Go agent
│   ├── main.go
│   ├── config/
│   ├── client/          # HTTP client
│   ├── system/          # System modules
│   └── .env.example
│
├── dashboard/           # Flutter Web (coming soon)
├── bin/                 # Compiled binaries
├── Makefile
└── README.md
```

---

## Screenshot Tools (Linux)

Agent mencoba screenshot tool berikut secara berurutan:
1. `scrot` (recommended: `apt install scrot`)
2. `import` (ImageMagick: `apt install imagemagick`)
3. `gnome-screenshot`
4. `spectacle` (KDE)

```bash
# Install scrot (recommended)
sudo apt install scrot
```

---

## Media Control (Linux)

Agent menggunakan:
1. `xdotool` untuk media keys
2. `playerctl` sebagai fallback

```bash
sudo apt install xdotool playerctl
```

---

## Lisensi

MIT
