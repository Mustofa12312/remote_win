# PRD — Workspace Commander

## 1. Ringkasan Produk

**Workspace Commander** adalah aplikasi remote management ringan berbasis Telegram dan Web Dashboard yang memungkinkan pengguna mengontrol, memonitor, dan mengelola beberapa komputer Windows dan Ubuntu dari satu perangkat (HP atau browser).

Fokus utama:

- Ringan
- Aman
- Tidak membutuhkan Python di komputer target
- Tidak membutuhkan Node.js di komputer target
- Agent berupa satu file executable
- Mendukung Windows 11 dan Ubuntu
- Dapat berjalan di jaringan lokal maupun internet

---

# 2. Tujuan Produk

Masalah yang ingin diselesaikan:

- Sulit mengakses komputer saat jauh dari perangkat.
- Sulit mengambil file dari komputer kantor atau rumah.
- Sulit mengetahui status komputer tanpa membuka remote desktop.
- Sulit mengontrol media yang sedang berjalan.
- Sulit memonitor banyak komputer sekaligus.

Workspace Commander menjadi pusat kontrol seluruh perangkat.

---

# 3. Target Pengguna

### Developer

- Flutter Developer
- Web Developer
- Linux User
- Administrator Server

### Pengguna Umum

- Pekerja kantor
- Guru
- Mahasiswa
- Pemilik banyak perangkat

---

# 4. Platform

### Agent

- Windows 11
- Ubuntu

### Client

- Telegram
- Browser
- Android
- iPhone

---

# 5. Arsitektur Sistem

```text
                    Telegram
                        │
                        ▼
              Workspace Server
                        │
        ┌───────────────┼───────────────┐
        ▼                               ▼
 Windows Agent                  Ubuntu Agent
        │                               │
        ▼                               ▼
 Local System                  Local System
```

---

# 6. Komponen Sistem

## Workspace Server

Fungsi:

- Device Management
- Authentication
- Command Routing
- Monitoring Storage
- Notification Engine

---

## Agent

Fungsi:

- Menjalankan perintah
- Monitoring perangkat
- Mengambil file
- Mengirim screenshot
- Mengontrol media

---

# 7. Fitur Utama

---

## Modul Device Management

### Daftar Perangkat

Menampilkan:

```text
🟢 Windows Kantor
🟢 Ubuntu Rumah
🔴 Laptop Cadangan
```

Informasi:

- Nama perangkat
- Sistem operasi
- Status online
- IP lokal
- Versi agent
- Last seen

---

## Pemilihan Device Aktif

Semua perintah berikutnya dijalankan pada device yang dipilih.

Contoh:

```text
Current Device:
Windows Kantor
```

---

# 8. Modul System Control

## Shutdown

Telegram:

```text
/shutdown
```

Aksi:

- Mematikan perangkat

---

## Restart

```text
/restart
```

---

## Sleep

```text
/sleep
```

---

## Lock Screen

```text
/lock
```

---

## Logout User

```text
/logout
```

---

# 9. Modul Monitoring

## CPU Monitor

Menampilkan:

```text
CPU Usage
CPU Temperature
CPU Frequency
```

---

## RAM Monitor

Menampilkan:

```text
Total RAM
Used RAM
Free RAM
```

---

## Disk Monitor

Menampilkan:

```text
Disk C
Disk D
Disk Usage
Disk Free Space
```

---

## Network Monitor

Menampilkan:

```text
Upload Speed
Download Speed
Current IP
Internet Status
```

---

## Battery Monitor

Laptop:

```text
Battery Level
Charging Status
Remaining Time
```

---

# 10. Modul File Explorer

## Browse Drive

Windows:

```text
C:
D:
E:
```

Ubuntu:

```text
/
home
mnt
media
```

---

## Folder Navigation

Telegram menampilkan:

```text
📁 Documents
📁 Downloads
📁 Project
```

Menggunakan tombol inline.

---

## Download File

Telegram:

```text
📄 laporan.pdf
```

Ketika dipilih:

```text
Download
```

File dikirim ke Telegram.

---

## Upload File

Kirim file ke Telegram Bot.

Pilih lokasi:

```text
Desktop
Downloads
Documents
Custom Folder
```

File disimpan ke komputer.

---

## Search File

Perintah:

```text
/find laporan
```

Hasil:

```text
D:\Project\laporan.pdf
C:\Users\Admin\Downloads\laporan.pdf
```

---

## Open Folder

Membuka folder di komputer target.

---

# 11. Modul Screenshot

## Full Desktop Screenshot

Perintah:

```text
/screenshot
```

Agent:

- Mengambil screenshot
- Mengirim ke Telegram

---

## Multi Monitor Support

Jika monitor lebih dari satu:

```text
Monitor 1
Monitor 2
All Monitor
```

---

# 12. Modul Media Control

## Media Detection

Mendeteksi media aktif:

- Browser
- YouTube
- Spotify
- VLC
- Local Media

---

## Play

```text
▶ Play
```

---

## Pause

```text
⏸ Pause
```

---

## Next

```text
⏭ Next
```

---

## Previous

```text
⏮ Previous
```

---

## Volume Up

```text
🔊 Volume +
```

---

## Volume Down

```text
🔉 Volume -
```

---

## Mute

```text
🔇 Mute
```

---

# 13. Modul Notification Engine

## Disk Alert

Contoh:

```text
⚠ Disk C tersisa 5 GB
```

---

## RAM Alert

```text
⚠ RAM di atas 90%
```

---

## Battery Alert

```text
⚠ Baterai tinggal 15%
```

---

## Device Offline Alert

```text
⚠ Windows Kantor Offline
```

---

# 14. Modul Telegram Bot

## Menu Utama

```text
📊 Status
💻 Devices
📁 Files
📷 Screenshot
🎵 Media
⚙ System
```

---

## Device Menu

```text
🟢 Windows Kantor
🟢 Ubuntu Rumah
```

---

## File Menu

```text
Browse
Search
Upload
Download
```

---

# 15. Modul Web Dashboard

## Dashboard Overview

Menampilkan:

- Semua device
- CPU
- RAM
- Storage
- Status online

---

## Device Page

Menampilkan:

- Detail device
- Monitoring
- Screenshot
- File Browser

---

## Notification Page

Menampilkan:

- Semua alert
- Riwayat aktivitas

---

# 16. Keamanan

## Device Registration

Setiap device memiliki:

```text
Device ID
Device Secret
```

---

## Telegram Authorization

Hanya Telegram ID tertentu:

```text
Allowed User
```

dapat mengakses sistem.

---

## Command Confirmation

Untuk perintah berbahaya:

```text
Shutdown
Restart
Logout
```

muncul:

```text
Are you sure?
```

---

## Encrypted Communication

Semua komunikasi:

```text
HTTPS
TLS
```

---

# 17. Teknologi

## Agent

Bahasa:

Go

Output:

```text
workspace-agent.exe
workspace-agent
```

---

## Server

Bahasa:

Go

---

## Database

SQLite

---

## Dashboard

Flutter

---

# 18. Non-Goals (Versi 1)

Tidak termasuk:

- Remote Desktop penuh
- Kamera webcam
- Perekaman layar
- Keylogger
- Monitoring aktivitas pengguna
- Eksekusi command arbitrer tanpa whitelist
- Build APK Flutter
- Kontrol mouse dan keyboard jarak jauh

Fokus V1 adalah menjadi **pusat kontrol, monitoring, file explorer, screenshot, dan media controller** yang ringan, aman, dan dapat digunakan harian pada Windows 11 dan Ubuntu.
