# 🧰 Tools, Skills & MCP Architecture Reference

Cybermes combines a curated offensive security toolchain, deep skill playbooks, and Model Context Protocol (MCP) servers to automate end-to-end vulnerability research and verification.

---

## 🛠️ 1. Integrated Security Toolchain

All native binary security tools reside in `/workspace/tools/bin` and are automatically registered on the container `$PATH`:

| Category | Binary Tool | Core Purpose |
| :--- | :--- | :--- |
| **Reconnaissance & DNS** | `subfinder` | Fast passive subdomain enumeration |
| | `amass` | In-depth network mapping & asset OSINT |
| | `assetfinder` | Rapid domain and subdomain discovery |
| | `httpx` | Fast multi-purpose HTTP probing & technology detection |
| | `nmap` | Network exploration, port probing, and service fingerprinting |
| **Crawling & Mining** | `katana` | Next-generation web crawler with SPA rendering capabilities |
| | `gau` | Fetch historical URLs from Wayback Machine, AlienVault, CommonCrawl |
| | `waybackurls` | Fetch endpoint archives from web historical indexes |
| **Fuzzing & Discovery** | `ffuf` | High-performance web fuzzer for virtual hosts and paths |
| | `feroxbuster` | Fast, recursive content and directory discovery |
| **Exploitation & Scanning** | `nuclei` | Fast, community-driven template vulnerability scanner |
| | `sqlmap` | Automated SQL injection detection and database takeover |
| | `dalfox` | High-performance parameter analysis and XSS scanner |
| **Utilities & Streaming** | `smart_pipe` | High-throughput recon stream filter & token economy engine |
| | `secret_scan` | 48-pattern credential & sensitive token scanner |
| | `search_knowledge` | Sub-millisecond offline knowledge & payload search engine |
| | `aggregate_reports` | Automated finding aggregator & executive summary indexer |
| | `rg` (ripgrep) | Ultra-fast regex searching through source dumps and JS bundles |

---

## 🔌 2. Model Context Protocol (MCP) Servers

Cybermes natively integrates MCP servers for client-side evaluation, structured system access, and external AI assistant integration:

1. **Cybermes Native MCP Server (`@zyrexnn/cybermes-mcp`)**:
   * High-speed, Go-native JSON-RPC 2.0 MCP server exposing **10 specialized security tools**, 2 resources, and 2 prompts.
   * Compatible with **all AI providers** (Claude, GPT-4o, DeepSeek, Gemini, Llama) and clients (OpenCode, Kilo, Cursor, Claude Desktop, Windsurf, Cline, Continue.dev, Zed).
   * Instant Zero-Go execution: `npx -y @zyrexnn/cybermes-mcp`.
   * See [docs/MCP_SETUP.md](MCP_SETUP.md) for 1-click configuration templates.
2. **Browser MCP (`@modelcontextprotocol/server-puppeteer`)**:
   * Launches headless Chromium in container isolation.
   * Performs DOM tree inspection, automated button clicks, form submissions, and screenshot PoC generation.
3. **Filesystem MCP (`@modelcontextprotocol/server-filesystem`)**:
   * Provides structured read/write access to `/workspace` targets, logs, and reporting artifacts.

---

## 🎯 3. Offensive Skills (200+ Modules)

Skills reside in `skills/` and are automatically loaded into Hermes runtime context.

### Invoking Specific Playbooks:
Operators can prompt the agent to utilize specific methodologies:
* *"Apply 401-403 bypass techniques against /admin"*
* *"Audit parameters on /api/user following IDOR & BOLA methodology"*
* *"Execute race condition testing on the coupon redemption endpoint"*

---

## 📊 4. Automated Report Aggregator (`aggregate_reports`)

Cybermes includes a built-in report aggregator that automatically parses individual vulnerability reports and generates an executive summary:

```bash
# Aggregate findings for a specific target:
aggregate_reports <TARGET_SLUG>

# Aggregate findings across all tested targets:
aggregate_reports --all
```

Outputs generated in `reports/<TARGET_SLUG>/`:
- `SUMMARY.md`: Consolidated vulnerability matrix with severity badges, CVSS scores, and links.
- `metadata.json`: Machine-readable metadata and counts for programmatic integrations and dashboards.

---

## 📑 5. Automated PDF & HTML Deliverable Generator (`generate_pdf.py`)

Transforms structured markdown reports into an executive PDF report and interactive HTML dashboard:

```bash
# Generate PDF & HTML dashboard for a specific target:
python tools/generate_pdf.py <TARGET_SLUG>

# Generate for all targets:
python tools/generate_pdf.py --all

# Generate HTML only without Chromium PDF export:
python tools/generate_pdf.py --no-pdf
```

Outputs generated in `reports/<TARGET_SLUG>/`:
- `report.html`: Standalone interactive HTML dashboard with Dark/Light theme toggle.
- `REPORT.pdf`: Executive printable PDF with CVSS badges, risk charts, and PoC breakdown.

---

## 🔄 6. Toolchain & Template Auto-Updaters

Automates downloading and updating ProjectDiscovery binaries (`subfinder`, `httpx`, `katana`, `nuclei`) and Nuclei templates:

* **Linux & macOS / Container**:
  ```bash
  ./tools/update_tools.sh
  ```
* **Windows (PowerShell)**:
  ```powershell
  powershell -ExecutionPolicy Bypass -File tools\update_tools.ps1
  ```

---

## 🔍 7. Skill Pack Integrity Auditor (`validate_skills.py`)

Verifies the integrity of all 200+ skill folders in `skills/`:

```bash
python tools/validate_skills.py
```
Reports total loaded skills, active skills, and highlights any missing or corrupted `SKILL.md` definitions.

---

## 🩺 8. Universal Health Check & Auto-Repair (`doctor.py`)

Inspects environment health across Windows, Linux, and macOS, and auto-repairs missing directories, toolchains, and configurations:

```bash
# Check environment state
python tools/doctor.py

# Automatically fix missing directories, tools, and configs
python tools/doctor.py --fix
```

