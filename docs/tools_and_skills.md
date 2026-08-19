# 🧰 Panduan Tools, Skills & MCP Server Cybermes

Cybermes menggabungkan ratusan modul offensive skills, toolchain eksternal, dan Model Context Protocol (MCP) untuk otomasi pengujian keamanan.

---

## 🛠️ 1. Daftar Security Toolchain

Semua binary security tools diletakkan di `/workspace/tools/bin` dan otomatis masuk ke `$PATH` container:

| Kategori | Nama Tool | Kegunaan Utama |
| :--- | :--- | :--- |
| **Reconnaissance** | `subfinder` | Passive subdomain enumeration kecepatan tinggi |
| | `amass` | In-depth network mapping & asset OSINT |
| | `assetfinder` | Pencarian domain & subdomain cepat |
| | `httpx` | HTTP probing, technology detection, & status check |
| | `nmap` | Port scanner & service version detector |
| **Crawling & Mining** | `katana` | Next-gen web crawler & SPA parser |
| | `gau` | Mengambil historical URLs dari Wayback, AlienVault, CommonCrawl |
| | `waybackurls` | Fetch endpoint historis dari web archive |
| **Fuzzing & Content Discovery** | `ffuf` | Web fuzzer berkecepatan tinggi untuk endpoint/direktori |
| | `feroxbuster` | Recursive content discovery engine |
| **Exploitation & Scanning** | `nuclei` | Template-based vulnerability scanner |
| | `sqlmap` | Database takeover & SQL injection testing |
| | `dalfox` | XSS scanner & parameter analysis engine |
| **Utilities** | `rg` (ripgrep) | Pencarian pola regex cepat pada source/JS dump |

---

## 🔌 2. Model Context Protocol (MCP) Servers

Cybermes mengintegrasikan MCP Servers untuk interaksi runtime yang lebih fleksibel:

1. **Browser MCP (`@modelcontextprotocol/server-puppeteer`)**:
   * Menjalankan browser Chromium headless di dalam container.
   * Mendukung DOM evaluation, automated button clicks, form filling, dan screenshot PoC.
2. **Filesystem MCP (`@modelcontextprotocol/server-filesystem`)**:
   * Akses terisolasi dan terstruktur ke direktori kerja `/workspace`.

---

## 🎯 3. Offensive Skills (200+ Modul)

Skills terletak di direktori `skills/` dan dimuat otomatis oleh runtime Hermes.

### Cara Memanggil Skill Tertentu:
Di terminal atau chat Telegram, Anda dapat langsung meminta agent menggunakan teknik tertentu:
* *"Gunakan teknik 401-403 bypass untuk mengakses /admin"*
* *"Audit parameter pada /api/user dengan metodologi IDOR & BOLA"*
* *"Lakukan race condition testing pada endpoint transfer saldo"*
