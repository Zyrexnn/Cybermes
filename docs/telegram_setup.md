# 🤖 Panduan Integrasi & Penggunaan Telegram Bot Cybermes

Cybermes dilengkapi dengan gateway perpesanan otonom berbasis **Hermes Agent Gateway**, yang memungkinkan Anda mengontrol proses asesmen keamanan, menjalankan recon, hingga menerima laporan temuan secara real-time langsung melalui Telegram.

---

## 📋 1. Persiapan Kredensial

### A. Buat Bot Telegram via @BotFather
1. Buka aplikasi Telegram dan cari **[@BotFather](https://t.me/BotFather)**.
2. Kirim perintah `/newbot`.
3. Masukkan nama tampilan bot (misal: `Cybermes Security Agent`).
4. Masukkan username bot (harus berakhiran `bot`, misal: `my_cybermes_sec_bot`).
5. Simpan **HTTP API Token** yang diberikan (format: `1234567890:ABCdefGhIJKlmNoPQRsTUVwxyZ`).

### B. Dapatkan Telegram User ID Anda (Keamanan & Whitelist)
Agar bot hanya merespons perintah dari Anda dan tidak bisa diakses orang lain:
1. Buka Telegram dan cari **[@userinfobot](https://t.me/userinfobot)**.
2. Kirim `/start`.
3. Catat angka `Id` Anda (misal: `123456789`).

---

## ⚙️ 2. Konfigurasi Environment (`.env` & `.hermes/.env`)

Edit file `.env` atau `.hermes/.env`:

```bash
# ─────────────────────────────────────────────────────────────────────────────
# TELEGRAM BOT INTEGRATION (Gateway)
# ─────────────────────────────────────────────────────────────────────────────
TELEGRAM_BOT_TOKEN=1234567890:ABCdefGhIJKlmNoPQRsTUVwxyZ
TELEGRAM_ALLOWED_USERS=123456789 # Masukkan User ID Anda (pisahkan dengan koma jika multi-user)
GATEWAY_ALLOW_ALL_USERS=false    # Set false agar aman dari akses tidak dikenal
HERMES_YOLO_MODE=1              # Mengizinkan eksekusi tool otomatis tanpa konfirmasi manual
```

---

## 🚀 3. Menjalankan Bot di Docker

### Start Bot (Background Gateway)
```bash
docker compose up -d
```

### Memeriksa Status Log Bot
```bash
docker compose logs -f
```
Jika berhasil terhubung, Anda akan melihat output:
```text
✓ telegram connected
[Telegram] Connected to Telegram (polling mode)
[Telegram] 60 commands registered
```

### Restart Bot
```bash
docker compose restart
```

---

## 📱 4. Perintah & Command Penting di Telegram

| Command | Fungsi | Keterangan |
| :--- | :--- | :--- |
| `/new` atau `/reset` | **Mereset sesi / Memory baru** | Bersihkan riwayat chat dan mulai konteks baru |
| `/status` | **Cek Status Agent** | Melihat model aktif, penggunaan konteks, dan uptime |
| `/skills` | **Daftar Skills Aktif** | Menampilkan skill keamanan yang tersedia (200+ skills) |
| `/help` | **Panduan Perintah** | Menampilkan semua command yang didukung |
| `/model` | **Ganti Model LLM** | Menampilkan / mengganti model inference aktif |
| `/stop` | **Hentikan Eksekusi** | Membatalkan eksekusi task/tool yang sedang berjalan |

---

## 💡 5. Tips & Best Practice Berinteraksi

1. **Gunakan Perintah Target yang Jelas**:
   ```text
   Lakukan pasif recon pada http://localhost:8888 sesuai cakupan scope.yaml dan identifikasi endpoint yang aktif.
   ```
2. **Jika Bot Merespons dengan Refusal (Menolak)**:
   * Ketik `/reset` untuk membersihkan context percakapan sebelumnya.
   * Tekankan bahwa target berada di dalam ruang lingkup audit resmi (`scope.yaml`) atau lingkungan lab terisolasi.
