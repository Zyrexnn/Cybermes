# 🚀 Next Steps: Release v3.5.1 & Partnership Fulfillment

Panduan 5 langkah tersusun secara runtut untuk menyelesaikan proses rilis dan partnership:

---

### 1. Merge Pull Request #78 (`dev` ➔ `main`)
Review dan merge PR yang baru saja dibuat:
- **Via Browser:** Buka [PR #78](https://github.com/Zyrexnn/Cybermes/pull/78) lalu klik **Merge pull request**.
- **Atau via Terminal:**
  ```bash
  gh pr merge 78 --merge --delete-branch=false
  ```

---

### 2. Buat & Push Git Tag Release (`v3.5.1`)
Setelah PR dimerge ke `main`, buat tag resmi agar pipeline CI/CD otomatis berjalan:
```bash
git checkout main
git pull origin main
git tag v3.5.1
git push origin v3.5.1
```
> ⚡ **Yang otomatis terjadi di GitHub Actions:**
> 1. Mengompilasi binary `cybermes-mcp` untuk 5 arsitektur (Windows, Linux, macOS).
> 2. Membuat halaman rilis resmi di GitHub Releases.
> 3. Mempublikasikan package **`cybermes-mcp@3.5.1`** langsung ke registry NPM menggunakan token terdaftar.

---

### 3. Verifikasi Publikasi NPM & GitHub Release
Setelah proses CI/CD selesai (~3-5 menit):
```bash
# Verifikasi versi aktif di registry NPM
npm view cybermes-mcp version
# Output diharapkan: 3.5.1
```
Cek juga halaman rilis di: [GitHub Releases v3.5.1](https://github.com/Zyrexnn/Cybermes/releases/tag/v3.5.1)

---

### 4. Broadcast Pengumuman Sponsorship (Placement 3)
Gunakan draft siap pakai dari file [docs/announcements/2026-11-13-infistar-sponsorship.md](file:///home/ikhsan/Documents/Cybermes/docs/announcements/2026-11-13-infistar-sponsorship.md):
1. **GitHub Discussions:** Buat postingan kategori *Announcements* di repository.
2. **Telegram / Discord Komunitas:** Kirim pesan broadcast ringkas + lampirkan file banner `assets/sponsors/infistar_banner.png` atau QR `assets/sponsors/infistar_qr.png`.
3. **Twitter (X):** Posting tweet pengumuman dengan tag `@infistar_ai`.

---

### 5. Kirim Bukti Pemenuhan ke Pihak Infistar
Kirimkan email balasan ke pihak Infistar untuk melaporkan bahwa ke-3 placement wajib sudah live:
1. **README Placement:** `https://github.com/Zyrexnn/Cybermes#readme` *(banner + placement copy)*
2. **Documentation Placement:** `https://github.com/Zyrexnn/Cybermes/blob/main/docs/INFISTAR_SETUP.md` *(promosi gateway & referral URL)*
3. **Announcement Placement:** Tautan ke postingan pengumuman (GitHub Discussions / Telegram / Twitter)
