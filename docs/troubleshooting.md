# 🛠️ Panduan Pemecahan Masalah (Troubleshooting Guide)

Dokumen ini berisi panduan penanganan masalah umum saat menjalankan Cybermes.

---

## 1. Bot Telegram Tidak Merespons Perintah

### Gejala:
Perintah seperti `/new`, `/status`, atau pesan biasa terkirim di Telegram (centang dua), tetapi bot tidak membalas.

### Penyebab & Solusi:
1. **Container tidak berjalan dalam mode gateway:**
   * Pastikan `docker-compose.yml` memiliki `command: hermes gateway run`.
   * Cek status: `docker compose logs --tail 30` dan pastikan muncul `✓ telegram connected`.
2. **User ID Anda belum diizinkan (Whitelist):**
   * Periksa `.hermes/.env`.
   * Jika `GATEWAY_ALLOW_ALL_USERS=false`, pastikan `TELEGRAM_ALLOWED_USERS` sudah berisi Telegram User ID Anda.
3. **Session tersangkut atau model hang:**
   * Kirim perintah `/reset` atau `/new` di chat Telegram.
   * Atau restart container: `docker compose restart`.

---

## 2. Model Menolak Perintah (Refusal / Safety Error)

### Gejala:
Bot menjawab: *"Saya menolak melakukan pengujian... Anda belum menunjukkan bukti otorisasi..."*.

### Solusi:
1. **Gunakan Framing Otorisasi Jelas**:
   Format instruksi dengan menyertakan referensi `scope.yaml`:
   > *"Sesuai dengan scope.yaml untuk lingkungan lab lokal http://127.0.0.1:8888, lakukan audit otorisasi pada endpoint registration."*
2. **Reset Konteks**:
   Ketik `/reset` di Telegram agar histori percakapan yang terpicu safety refusal dibersihkan.

---

## 3. OmniRoute / LLM Connection Error (`Connection Refused`)

### Gejala:
Log container menampilkan `ConnectionRefusedError: http://localhost:20128/v1`.

### Solusi:
* Karena `docker-compose.yml` menggunakan `network_mode: host`, endpoint `http://localhost:20128/v1` mengarah langsung ke port OmniRoute di host.
* Pastikan instance OmniRoute / LM Studio / Local LLM provider sudah aktif di port yang sesuai.
* Jika menggunakan remote OpenRouter langsung:
  Ubah di `.hermes/.env`:
  ```ini
  OPENROUTER_BASE_URL=https://openrouter.ai/api/v1
  OPENROUTER_API_KEY=sk-or-v1-...
  ```

---

## 4. Izin File & Permissions pada Output

### Gejala:
File di folder `recon/`, `output/`, atau `reports/` dibuat dengan user `root` sehingga sulit diedit di host.

### Solusi:
Jalankan di terminal host:
```bash
sudo chown -R $USER:$USER recon/ output/ reports/ logs/
```
