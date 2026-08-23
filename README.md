<div align="center">

<img src="assets/banner.jpg" alt="Cybermes Autonomous Security Agent Banner" width="100%" style="border-radius: 10px; margin-bottom: 20px;">

# 🛡️ Cybermes

### **Autonomous Offensive Security, Bug Bounty & Red Teaming Agent Framework**

[![Release: v1.5.0](https://img.shields.io/badge/Release-v1.5.0-orange.svg)](https://github.com/Zyrexnn/Cybermes/releases/tag/v1.5.0)
[![License: PolyForm Noncommercial](https://img.shields.io/badge/License-PolyForm_Noncommercial-blue.svg)](LICENSE)
[![Python: 3.11+](https://img.shields.io/badge/Python-3.11+-blue.svg)](https://www.python.org/)
[![Docker: Ready](https://img.shields.io/badge/Docker-Supported-2496ED.svg)](docker-compose.yml)
[![Windows: Supported](https://img.shields.io/badge/Windows-Supported-brightgreen.svg)](docs/INSTALL_WINDOWS.md)
[![Hermes: Powered](https://img.shields.io/badge/Hermes%20Agent-Core-purple.svg)](https://github.com/NousResearch/Hermes-Agent)
[![PDF Reporting: Automated](https://img.shields.io/badge/Reports-PDF%20%26%20HTML%20Automated-brightgreen.svg)](#-automated-executive-reporting-pdf--html)
[![Token Economy: 85% Saved](https://img.shields.io/badge/Token%20Economy-Smart%20Filter-blueviolet.svg)](#-token-economy--smart-output-filtering)
[![AI Standards: AGENTS.md](https://img.shields.io/badge/AI%20Standards-AGENTS.md%20%2B%20.cursorrules-success.svg)](AGENTS.md)

<p align="center">
  <b>Cybermes</b> is an enterprise-grade, autonomous security research agent designed for high-signal reconnaissance, attack surface discovery, authenticated vulnerability research, zero-false-positive exploit validation, token-efficient context management, and automated executive PDF/HTML report generation.<br><br>
  <i>✅ Works seamlessly on <strong>Linux</strong>, <strong>macOS</strong>, and <strong>Windows</strong></i>
</p>

[Quick Start](#-installation--quick-start) • [Architecture](#-architecture--core-engine) • [Automated PDF Reports](#-automated-executive-reporting-pdf--html) • [Skills Layer](#-offensive-skills-layer-50-modules) • [Documentation](docs/) • [Release Notes](#-release--version-history)

</div>

---

## ⚡ What Makes Cybermes Different?

| Traditional Security Scanners ❌ | Cybermes Autonomous Agent 🛡️ |
| :--- | :--- |
| **Noisy & Speculative**: Dumps hundreds of unverified alerts based on simple regex. | **Zero-False-Positive Gate**: Requires deterministic HTTP proof, status codes, and standalone Python PoC scripts before reporting. |
| **Context Window Bloat**: Dumps 5,000+ raw output lines into LLM context, causing hallucinations. | **Smart Output Filter & Token Economy**: Compresses verbose logs by 70–85% with `smart_pipe.py` and native Markdown MCP converters. |
| **Markdown-Only Deliverables**: Leaves users with raw markdown files scattered across directories. | **End-to-End Automated PDF/HTML Engine**: Generates pixel-perfect executive PDF reports (`REPORT.pdf`) and interactive HTML dashboards. |
| **Fragile File Permissions**: Docker and root processes create locked files (`NoPermissions`). | **Live Background Permission Daemon**: Integrated POSIX ACLs and live permission keeper guaranteeing `-rw-rw-rw-` open access. |
| **Single-Phase Execution**: Scans without understanding application logic or multi-step auth. | **Autonomous Reasoning Loop**: Mines JS bundles, tests multi-account auth matrices, and validates complex business logic. |

---

## 📑 Table of Contents

- [🏛️ Architecture & Core Engine](#-architecture--core-engine)
- [📑 Automated Executive Reporting (PDF & HTML)](#-automated-executive-reporting-pdf--html)
- [🧠 Token Economy & Smart Output Filtering](#-token-economy--smart-output-filtering)
- [🛡️ Universal AI Agent Standards (`AGENTS.md`)](#-universal-ai-agent-standards-agentsmd--cursorrules)
- [🔄 Operational Methodology (Phases 1–6)](#-operational-methodology-phases-16)
- [🧰 Available Toolchain & MCP Bridge](#-available-toolchain--mcp-bridge)
- [📁 Target-Scoped Directory Structure](#-target-scoped-directory-structure)
- [🚀 Installation & Quick Start](#-installation--quick-start)
  - [💻 Installing on Windows](#-installing-on-windows)
    - [Method 1A: PowerShell Native Setup (RECOMMENDED)](#method-1a-powershell-native-setup-recommended-)
    - [Method 1B: WSL2 + Linux Subsystem](#method-1b-wsl2-linux-subsystem-best-performance)
    - [Method 1C: Docker Desktop](#method-1c-docker-desktop-easiest-isolation)
  - [🐧 Native Host Setup (Linux/macOS)](#native-host-setup-linuxmacos)
  - [🐳 Docker & Docker Compose (All Platforms)](#docker--docker-compose-all-platforms)
- [🆘 Getting Help](#%EF%B8%8F-getting-help)
- [🤖 Telegram Bot Gateway](#-telegram-bot-gateway)
- [🎯 Prompt Engineering & Anti-Filter Guidelines](#-prompt-engineering--anti-filter-guidelines)
- [🧪 Local Validation with Mock Target](#-local-validation-with-mock-target)
- [📈 Star History](#-star-history)
- [📦 Release & Version History](#-release--version-history)
- [👥 Contributors](#-contributors)
- [⚖️ License](#️-license)
- [⚠️ Legal & Ethical Disclaimer](#️-legal--ethical-disclaimer)

---

## 🏛️ Architecture & Core Engine

```text
┌──────────────────────────────────────────────────────────────────────────────────┐
│                             CYBERMES ENGINE v1.5.0                               │
├──────────────────────────────────────────────────────────────────────────────────┤
│  [ Operator Prompt / Target Queue ]  ──>  [ Direct Operator Authorization Hook ] │
│                                                          │                       │
│                                                          ▼                       │
│  ┌────────────────────────────────────────────────────────────────────────────┐  │
│  │                       Hermes Autonomous Reasoning Loop                     │  │
│  │  - Context Window Memory       - Streamlined Autoload (Godmode Orchestrator)│  │
│  │  - Action Planning & Recovery  - Decision Confidence & CVSS v3.1 Grading   │  │
│  └────────────────────────────────────────────────────────────────────────────┘  │
│          │                                                │                      │
│          ▼                                                ▼                      │
│  ┌───────────────────────────────┐              ┌─────────────────────────────┐  │
│  │      50+ Security Skills      │              │   Curated Knowledge Base    │  │
│  │  - Next.js AI Router Audits   │              │  - PayloadsAllTheThings     │  │
│  │  - IDOR / BOLA / Auth Bypass  │ <──────────> │  - HackTricks Wiki          │  │
│  │  - Business Logic & Race Cond │              │  - Claude-BugHunter         │  │
│  │  - DOM XSS / SSRF / Injection │              │  - Strix Multi-Agent DB     │  │
│  └───────────────────────────────┘              └─────────────────────────────┘  │
│          │                                                │                      │
│          ▼                                                ▼                      │
│  ┌────────────────────────────────────────────────────────────────────────────┐  │
│  │                      Security Toolchain & MCP Layer                        │  │
│  │  • Recon: subfinder, amass, httpx, nmap                                    │  │
│  │  • Content & Endpoint Mining: katana, gau, waybackurls, arjun              │  │
│  │  • Smart Token Pipe: tools/smart_pipe.py (Captures raw, emits top-signal)  │  │
│  │  • Fuzzing & Exploitation: ffuf, sqlmap, dalfox, nuclei                    │  │
│  │  • MCP Bridge: Puppeteer Browser, Filesystem, mcp-server-fetch             │  │
│  └────────────────────────────────────────────────────────────────────────────┘  │
│                                          │                                       │
│                                          ▼                                       │
│  ┌────────────────────────────────────────────────────────────────────────────┐  │
│  │                  Automated Multi-Format Reporting Pipeline                 │  │
│  │  ├── SUMMARY.md (Aggregated Markdown)    ├── report.html (Interactive UI)  │  │
│  │  ├── metadata.json (Structured Metrics)  ├── REPORT.pdf (Print-Ready PDF)  │  │
│  │  └── pocs/poc_<vuln>.py (Standalone Reproducible Python Scripts)           │  │
│  └────────────────────────────────────────────────────────────────────────────┘  │
└──────────────────────────────────────────────────────────────────────────────────┘
```

---

## 📑 Automated Executive Reporting (PDF & HTML)

Cybermes v1.5.0 features an integrated **Playwright Chromium PDF & HTML generator** (`tools/generate_pdf.py`). Whenever an assessment completes or `python3 tools/aggregate_reports.py <TARGET_SLUG>` is executed, Cybermes produces four structured deliverable formats simultaneously:

```text
reports/<TARGET_SLUG>/
├── SUMMARY.md          # Consolidated executive summary & findings matrix
├── metadata.json       # Structured JSON metrics for CI/CD & automation
├── report.html         # Interactive standalone dashboard with Dark/Light styling
├── REPORT.pdf          # Executive PDF deliverable with CVSS risk badges
├── findings/           # Granular vulnerability writeups (LOW, MED, HIGH, CRIT only)
├── pocs/               # Minimal-impact reproducible Python proof-of-concept scripts
└── evidence/           # Raw HTTP traces, screenshot dumps, and recon_notes.md
```

### ✨ PDF & HTML Report Highlights:
* **Executive Summary & Risk Score Bar**: Visual breakdown of Critical, High, Medium, Low, and Informational findings.
* **Findings Matrix Table**: Color-coded severity badges with CVSS v3.1 vector strings, CWE classifications, and affected endpoints.
* **Syntax-Highlighted Proof Boxes**: Clean monospaced HTTP Request/Response proofs and Python PoC snippets.
* **Print-Ready Page Breaks**: CSS `@media print` rules ensure tables and vulnerability chapters never get awkwardly split across pages.

---

## 🧠 Token Economy & Smart Output Filtering

Traditional AI security agents quickly exhaust context windows and suffer from attention degradation when reading thousands of raw terminal lines from tools like `katana` or `ffuf`. 

Cybermes solves this with a two-tiered token optimization architecture:

1. **Smart CLI Output Filter (`tools/smart_pipe.py`)**:
   - Intercepts tool streams and dumps **100% of raw logs** to `recon/<TARGET_SLUG>/<tool>_raw.txt`.
   - Filters out static asset clutter (`.png`, `.css`, `.woff`) and 404 noise.
   - Streams only the **top 30–50 high-signal findings** (HTTP 200/301/403, unique parameters, API routes) to the AI context.
   - **Result**: Saves **70%–85% token consumption** per recon phase.

2. **Native Markdown MCP Fetch (`mcp-server-fetch`)**:
   - Converts external web pages and documentation directly into clean markdown, stripping massive raw HTML boilerplate before LLM evaluation.

---

## 🛡️ Universal AI Agent Standards (`AGENTS.md` & `.cursorrules`)

Cybermes is built for seamless collaboration across all modern AI developer ecosystems:

* **[`AGENTS.md`](AGENTS.md)**: Universal master operational directives defining core persona, zero-false-positive boundaries, toolchain syntax, anti-hallucination gates, and self-healing error recovery.
* **[`.cursorrules`](.cursorrules)**: Coding and workspace standards governing Python PoC construction (`requests`, explicit timeouts, error handling), snake_case file naming without shell brackets `[...]`, and secret hygiene.

---

## 🧰 Available Toolchain & MCP Bridge

All tools are pre-configured and accessible across host and Docker environments:

| Tool | Primary Purpose | Standard Syntax |
| :--- | :--- | :--- |
| **subfinder** | Passive Subdomain Discovery | `subfinder -d <target> -silent` |
| **httpx** | Probing & Tech Detection | `httpx -silent -status-code -title -tech-detect` |
| **katana** | Crawler & SPA Endpoint Miner | `katana -u <url> -silent -depth 3` |
| **smart_pipe.py** | Smart Filter & Token Saver | `<tool_cmd> \| python3 tools/smart_pipe.py --target <SLUG> --tool <NAME>` |
| **ffuf** | Directory & Parameter Fuzzing | `ffuf -u <url>/FUZZ -w tools/wordlists/common.txt -mc 200,301,302,403` |
| **nuclei** | Vulnerability Verification | `nuclei -u <url> -tags cve,auth-bypass -silent` |
| **sqlmap** | SQL Injection Auditor | `sqlmap -u "<url>?id=1" --batch --banner` |
| **dalfox** | XSS Scanner & Parameter Analyzer | `dalfox url <url> --silence` |
| **generate_pdf.py**| Automated PDF/HTML Generator | `python3 tools/generate_pdf.py <TARGET_SLUG>` |
| **update_tools.sh** | Toolchain & Template Auto-Updater | `./tools/update_tools.sh` |
| **search_knowledge.py** | Offline Payload & CheatSheet Search | `python3 tools/search_knowledge.py "<query>"` |
| **windows_compat_check.py** | Windows System Diagnostics | `python tools\windows_compat_check.py` |
| **Puppeteer MCP** | Browser DOM Automation | Native MCP tools for dynamic SPA testing & screenshot capture |
| **Fetch MCP** | Clean Web-to-Markdown Reader | Native MCP tool for token-efficient API inspection |

🪟 **Windows Users**: Run `.\\bin\\hermes.bat` instead of `./hermes` and use `setup_windows.ps1` for installation.

---

## 🚀 Installation & Quick Start

Choose your installation method based on your operating system and preferences:

| Method | OS | Difficulty | Performance | Best For |
|--------|----|------------|-------------|----------|
| **Native Setup** | Linux/macOS | ⭐ Easy | ⭐⭐⭐⭐⭐ | Linux users |
| **PowerShell Setup** | Windows 10/11 | ⭐ Easy | ⭐⭐⭐⭐ | Beginners, easy setup |
| **WSL2 Setup** | Windows → Linux | ⭐⭐ Medium | ⭐⭐⭐⭐⭐ | Power users |
| **Docker Desktop** | Windows/Linux/Mac | ⭐ Easy | ⭐⭐⭐⭐ | Isolated environments |

---

### 💻 Installing on Windows

#### Method 1A: PowerShell Native Setup (RECOMMENDED) ⭐

Perfect for Windows beginners who want easy setup without virtualization:

```powershell
# 1. Clone the repository
git clone https://github.com/Zyrexnn/Cybermes.git
cd Cybermes

# 2. Run automated installer (one command!)
.\setup_windows.ps1

# 3. Configure API keys
notepad .env

# 4. Launch an assessment
.\bin\hermes.bat "Assess http://127.0.0.1:8888"
```

👉 **See [Windows Installation Guide](docs/INSTALL_WINDOWS.md) for detailed instructions**

#### Method 1B: WSL2 + Linux Subsystem (Best Performance)

For Windows power users who prefer Linux environment:

```powershell
# Enable WSL2 (run once as Administrator)
wsl --install

# Then follow Linux instructions below
```

#### Method 1C: Docker Desktop (Easiest Isolation)

Fully isolated environment perfect for beginners:

```powershell
# 1. Clone repository
git clone https://github.com/Zyrexnn/Cybermes.git
cd Cybermes

# 2. Configure environment
copy .env.example .env
notepad .env

# 3. Start containers
.\bin\docker_windows.bat up

# 4. Execute commands
docker compose exec hermes-cybermes hermes "Assess http://127.0.0.1:8888"
```

---

### 🐧 Native Host Setup (Linux/macOS)

Cybermes provides a single-command automated installer for Linux and macOS:

```bash
# 1. Clone the repository
git clone https://github.com/Zyrexnn/Cybermes.git
cd Cybermes

# 2. Run the automated installer (sets up venv, Playwright, MCPs, ACLs, and tools)
./setup.sh

# 3. Configure your API keys in .env
nano .env

# 4. Activate the environment
source env.sh

# 5. Launch an assessment
./hermes "Assess http://127.0.0.1:8888 and generate full report."
```

---

### 🐳 Docker & Docker Compose (All Platforms)

Run Cybermes inside a fully isolated, rootless-friendly container:

```bash
# 1. Clone and navigate to repository
git clone https://github.com/Zyrexnn/Cybermes.git
cd Cybermes

# 2. Copy and edit environment variables
cp .env.example .env
nano .env

# 3. Start Cybermes container
docker compose up -d

# 4. Execute commands inside container
docker compose exec hermes-cybermes hermes "Assess http://127.0.0.1:8888"
```

---

## 🆘 Getting Help

### Quick Links

- **[Windows Installation Guide](docs/INSTALL_WINDOWS.md)** - Complete Windows setup tutorial
- **[Troubleshooting Common Issues](docs/troubleshooting.md)** - Solutions for common problems
- **[Prompt Engineering Guide](docs/prompt_guide.md)** - Learn effective prompting techniques
- **[GitHub Discussions](https://github.com/Zyrexnn/Cybermes/discussions)** - Ask questions, share experiences
- **[Release Notes](https://github.com/Zyrexnn/Cybermes/releases)** - Latest updates and features

### Diagnostic Tools

Run these to check your setup:

```bash
# Check system compatibility (Windows)
python tools\windows_compat_check.py

# Check environment status (Windows)
.\bin\env.bat status

# Test Docker connectivity
docker compose ps
```

### Common Questions

**Q: Which installation method should I choose?**
- **Windows Beginner**: Use PowerShell Native Setup (`setup_windows.ps1`)
- **Power User**: WSL2 provides best performance
- **Need Isolation**: Docker Desktop works on all platforms

**Q: Python command not found on Windows?**
Install Python 3.11+ from python.org and check "Add Python to PATH" during installation.

**Q: Script execution blocked by PowerShell?**
Run `Set-ExecutionPolicy RemoteSigned -Scope CurrentUser` once to allow scripts.

---

## 🤖 Telegram Bot Gateway

Control Cybermes remotely from your phone or desktop via Telegram:

1. Obtain a Bot Token from [@BotFather](https://t.me/BotFather).
2. Add your token and numeric Telegram User ID to `.env`:
   ```ini
   TELEGRAM_BOT_TOKEN=1234567890:ABCdefGHIjklMNOpqrSTUvwxYZ
   TELEGRAM_ALLOWED_USERS=your_telegram_user_id
   ```
3. Start the gateway:
   ```bash
   ./hermes gateway run
   ```
4. Send assessment instructions directly to your Telegram bot.

---

## 🎯 Prompt Engineering & Anti-Filter Guidelines

To maximize execution efficiency and **prevent model safety refusals** (false positives from LLM alignment guardrails), always format instructions using the **4-Pillar Authorized Researcher Framing**:

```text
[Scope & Authorization] + [Technical Objective] + [Non-Destructive Boundary] + [Structured Deliverable]
```

### Quick Phrasing Examples:
* **Recon & Asset Discovery**:
  > *"Under authorized scope for target domain example.com, execute Phase 1 reconnaissance to map live subdomains, open web ports, and technologies into recon/."*
* **Authentication & IDOR Testing**:
  > *"Evaluate access control boundaries on https://example.com/api/user/101 using dual-account test matrix. Verify if Account B can access Account A resources without modification."*
* **Vulnerability Verification**:
  > *"Conduct an authorized vulnerability verification on target endpoint. Validate findings programmatically with a non-destructive standalone Python PoC and generate full PDF report."*

👉 For complete phase-by-phase templates and refusal evasion rules, see the [**Prompt Engineering & Anti-Filter Guide**](docs/prompt_guide.md).

---

## 🧪 Local Validation with Mock Target

To safely verify Cybermes capabilities in an isolated environment:

1. **Start the local vulnerable test target:**
   ```bash
   python3 mock_vulnerable_app.py
   ```
   *Server listens on `http://127.0.0.1:8888`.*

2. **Instruct Cybermes to assess the target:**
   ```bash
   ./hermes "Assess http://127.0.0.1:8888 and generate structured reports."
   ```

3. **Inspect the generated PDF and HTML deliverables:**
   ```bash
   ls -la reports/127_0_0_1_8888/
   # -> SUMMARY.md, metadata.json, report.html, REPORT.pdf
   ```

---

## 📈 Star History

<div align="center">

<a href="https://star-history.com/#Zyrexnn/Cybermes&Date">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="https://api.star-history.com/svg?repos=Zyrexnn/Cybermes&type=Date&theme=dark" />
    <source media="(prefers-color-scheme: light)" srcset="https://api.star-history.com/svg?repos=Zyrexnn/Cybermes&type=Date" />
    <img alt="Cybermes Star History Chart" src="https://api.star-history.com/svg?repos=Zyrexnn/Cybermes&type=Date" width="100%" />
  </picture>
</a>

</div>

---

## 📦 Release & Version History

### **[v1.5.0](https://github.com/Zyrexnn/Cybermes/releases/tag/v1.5.0)** — *Windows Native Ecosystem, Automated PowerShell Setup & Multi-Platform Diagnostics*
* **1-Click PowerShell Installer (`setup_windows.ps1`)**: Automated single-command installer for Windows 10/11 creating Python `venv`, workspace structure, installing requirements, and configuring `.env` / `.hermes/config.yaml`.
* **Windows Compatibility Diagnostic Tool (`tools/windows_compat_check.py`)**: Automated verification of Python runtime, active virtualenv, workspace directory permissions, security binaries in PATH, and console UTF-8 safety.
* **Native Execution & Docker Wrappers**: Added `bin/hermes.bat`, `bin/env.bat`, `bin/docker_windows.bat`, and `env.ps1` for frictionless CLI operation across PowerShell and Command Prompt.
* **Dedicated Windows Guide (`docs/INSTALL_WINDOWS.md`)**: Comprehensive documentation covering PowerShell native setup, WSL2 integration, Docker Desktop, and Windows-specific troubleshooting (ExecutionPolicy, path limits, antivirus exclusions).
* **Branching Strategy & Contribution Guidelines**: Formalized `dev` branch workflow in [`CONTRIBUTING.md`](CONTRIBUTING.md) and refreshed [`CONTRIBUTORS.md`](CONTRIBUTORS.md).

### **[v1.4.0](https://github.com/Zyrexnn/Cybermes/releases/tag/v1.4.0)** — *Toolchain Auto-Updater, Knowledge Search & Crawler Protection*
* **Offline Knowledge Search Engine**: Added `tools/search_knowledge.py` for rapid query lookups across local *PayloadsAllTheThings*, *HackTricks*, and *Claude-BugHunter* data.
* **Toolchain & Template Auto-Updater**: Added `tools/update_tools.sh` for continuous synchronization of security tools and wordlists.
* **Crawler Privacy & Exclusions**: Added `robots.txt` exclusion rules to protect payload databases from scraper indexing.
* **Licensing Framework**: Transitioned to PolyForm Noncommercial License 1.0.0.

> 💡 *For earlier release notes (v1.0.0 – v1.3.0), view the complete [GitHub Releases Archive](https://github.com/Zyrexnn/Cybermes/releases).*


---

## ⚖️ License

This project is licensed under the **[PolyForm Noncommercial License 1.0.0](LICENSE)**. Free for personal, research, education, and noncommercial use. Commercial use is strictly prohibited without explicit permission.

Third-party research materials, datasets, and upstream tools incorporated or referenced within this repository retain their respective original licenses (see [ATTRIBUTION.md](ATTRIBUTION.md) for full details).

---

## ⚠️ Legal & Ethical Disclaimer

> **IMPORTANT**: Cybermes is developed exclusively for **authorized security testing**, **legitimate bug bounty research**, and **academic security education**.
> 
> Testing against targets without explicit, prior written permission is illegal and strictly prohibited. The authors and contributors assume no liability and are not responsible for any misuse, damage, or legal consequences caused by the use of this software.

---

## 🙏 Acknowledgments & Upstream Credits

Cybermes stands on the shoulders of giants in the open-source and offensive security research communities:

| Project / Tool | Author / Maintainer | Contribution to Cybermes |
| :--- | :--- | :--- |
| **[HackTricks](https://github.com/carlospolop/hacktricks)** | [@carlospolop](https://github.com/carlospolop) (Carlos Polop) | Privilege escalation, service exploitation & pentesting knowledge |
| **[PayloadsAllTheThings](https://github.com/swisskyrepo/PayloadsAllTheThings)** | [@swisskyrepo](https://github.com/swisskyrepo) (Swissky) | Web application payloads and bypass vectors |
| **[SQLMap](https://github.com/sqlmapproject/sqlmap)** | Bernardo Damele & Miroslav Stampar | Automated SQL injection detection engine |
| **[Claude-BugHunter](https://github.com/sachinsharma-96/Claude-BugHunter)** | [@sachinsharma-96](https://github.com/sachinsharma-96) (Sachin Sharma) | Bug bounty engagement patterns & validation playbooks |
| **[Strix Framework](https://github.com/strix-security/strix)** | Strix Security Team | Autonomous multi-agent coordination architecture |
| **[ProjectDiscovery Suite](https://projectdiscovery.io/)** | ProjectDiscovery Team | Foundation tools (`nuclei`, `httpx`, `subfinder`, `katana`) |
| **[FFuF](https://github.com/ffuf/ffuf)** | [@joohoi](https://github.com/joohoi) | High-speed web fuzzer |

---

## 👥 Contributors

A huge thanks to everyone who has contributed to making **Cybermes** better!

<p align="center">
  <a href="https://github.com/Zyrexnn/Cybermes/graphs/contributors">
    <img src="https://contrib.rocks/image?repo=Zyrexnn/Cybermes" alt="Cybermes Contributors" />
  </a>
</p>

| Contributor | Profile | Role / Contribution | GitHub |
| :---: | :--- | :--- | :--- |
| <a href="https://github.com/Zyrexnn"><img src="https://github.com/Zyrexnn.png" width="55" style="border-radius:50%;" alt="Zyrexnn"/></a> | **[Zyrexnn](https://github.com/Zyrexnn)** | Project Lead, Core Architecture & Offensive Security Research | [@Zyrexnn](https://github.com/Zyrexnn) |
| <a href="https://github.com/claude"><img src="https://github.com/claude.png" width="55" style="border-radius:50%;" alt="Claude"/></a> | **[Claude](https://github.com/claude)** | AI Co-Author & Security Architecture Research | [@claude](https://github.com/claude) |
| <a href="https://github.com/msarg44"><img src="https://github.com/msarg44.png" width="55" style="border-radius:50%;" alt="msarg44"/></a> | **[msarg44](https://github.com/msarg44)** | Playwright PDF Rendering Engine Fix ([#1](https://github.com/Zyrexnn/Cybermes/pull/1)) | [@msarg44](https://github.com/msarg44) |

<br>

<p align="center">
  <i>Want to contribute? Check out our <a href="CONTRIBUTING.md">Contributing Guide</a> and submit a pull request to the <code>dev</code> branch!</i>
</p>
