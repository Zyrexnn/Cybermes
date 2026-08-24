<div align="center">

<img src="assets/banner.jpg" alt="Cybermes Autonomous Security Agent Banner" width="100%" style="border-radius: 10px; margin-bottom: 20px;">

# 🛡️ Cybermes

### **Autonomous Offensive Security, Bug Bounty & Red Teaming Agent Framework**

[![Release: v2.1.0](https://img.shields.io/badge/Release-v2.1.0-orange.svg)](https://github.com/Zyrexnn/Cybermes/releases/tag/v2.1.0)
[![The New 100](https://www.theagenticleaderboard.com/badges/new/cybermes.svg)](https://www.theagenticleaderboard.com)
[![License: Apache 2.0](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](LICENSE)
[![Discussions: Active](https://img.shields.io/badge/Discussions-Join%20Community-blue.svg)](https://github.com/Zyrexnn/Cybermes/discussions)
[![Go: 1.22+](https://img.shields.io/badge/Go-1.22+-00ADD8.svg?logo=go&logoColor=white)](https://go.dev/)
[![Python: 3.11+](https://img.shields.io/badge/Python-3.11+-3776AB.svg?logo=python&logoColor=white)](https://www.python.org/)
[![Docker: Ready](https://img.shields.io/badge/Docker-Supported-2496ED.svg?logo=docker&logoColor=white)](docker-compose.yml)
[![Windows: Supported](https://img.shields.io/badge/Windows-Supported-brightgreen.svg?logo=windows&logoColor=white)](docs/INSTALL_WINDOWS.md)
[![Hermes: Powered](https://img.shields.io/badge/Hermes%20Agent-Core-purple.svg)](https://github.com/NousResearch/Hermes-Agent)
[![PDF Reporting: Automated](https://img.shields.io/badge/Reports-PDF%20%26%20HTML%20Automated-brightgreen.svg)](#-automated-executive-reporting-pdf--html)
[![Token Economy: 85% Saved](https://img.shields.io/badge/Token%20Economy-Smart%20Filter-blueviolet.svg)](#-token-economy--smart-output-filtering)
[![AI Standards: AGENTS.md](https://img.shields.io/badge/AI%20Standards-AGENTS.md%20%2B%20.cursorrules-success.svg)](AGENTS.md)

<p align="center">
  <b>Cybermes</b> is an enterprise-grade, autonomous security research agent designed for high-signal reconnaissance, attack surface discovery, authenticated vulnerability research, zero-false-positive exploit validation, token-efficient context management, and automated executive PDF/HTML report generation.<br><br>
  <i>⚡ Works seamlessly on <strong>Linux</strong>, <strong>macOS</strong>, and <strong>Windows</strong></i>
</p>

[Quick Start](#-installation--quick-start) • [Architecture](#-architecture--core-engine) • [Automated PDF Reports](#-automated-executive-reporting-pdf--html) • [Skills Layer](#-offensive-skills-layer-50-modules) • [Discussions & Rules](#-community--discussions) • [Documentation](docs/) • [Release Notes](#-release--version-history)

</div>

---

## ⚡ What Makes Cybermes Different?

| Traditional Security Scanners ❌ | Cybermes Autonomous Agent 🛡️ |
| :--- | :--- |
| **Noisy & Speculative**: Dumps hundreds of unverified alerts based on simple regex. | **Zero-False-Positive Gate**: Requires deterministic HTTP proof, status codes, and standalone Python PoC scripts before reporting. |
| **Context Window Bloat**: Dumps 5,000+ raw output lines into LLM context, causing hallucinations. | **Smart Output Filter & Token Economy**: Compresses verbose logs by 70–85% with high-throughput native Go `smart_pipe` and Markdown MCP converters. |
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
- [💬 Community & Discussions](#-community--discussions)
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
┌──────────────────────────────────────────────────────────────────────────────────────────────────┐
│                                      CYBERMES ENGINE v2.0.0                                      │
├──────────────────────────────────────────────────────────────────────────────────────────────────┤
│  [ Operator Prompt / Target Queue ]  ────────>  [ Direct Operator Authorization Hook ]           │
│                                                                │                                 │
│                                                                ▼                                 │
│  ┌────────────────────────────────────────────────────────────────────────────────────────────┐  │
│  │                              Hermes Autonomous Reasoning Loop                              │  │
│  │  • Context Window Memory Management         • Streamlined Autoload (Godmode Orchestrator)  │  │
│  │  • Action Planning & Self-Healing Recovery  • Decision Confidence & CVSS v3.1 Scoring      │  │
│  └────────────────────────────────────────────────────────────────────────────────────────────┘  │
│                 │                                                             │                  │
│                 ▼                                                             ▼                  │
│  ┌─────────────────────────────────────────┐          ┌───────────────────────────────────────┐  │
│  │           50+ Security Skills           │          │         Curated Knowledge Base        │  │
│  │  • Next.js / AI Router Audits           │          │  • PayloadsAllTheThings Database      │  │
│  │  • IDOR / BOLA / Auth Bypass Playbooks  │ <──────> │  • HackTricks Privilege Wiki          │  │
│  │  • Race Condition & Logic Flaw Tests    │          │  • Claude-BugHunter Methodologies     │  │
│  │  • DOM XSS, SSRF & Injection Matrices   │          │  • Sub-50ms Search (search_knowledge) │  │
│  └─────────────────────────────────────────┘          └───────────────────────────────────────┘  │
│                 │                                                             │                  │
│                 ▼                                                             ▼                  │
│  ┌────────────────────────────────────────────────────────────────────────────────────────────┐  │
│  │                             Security Toolchain & Native Go Core                            │  │
│  │  • High-Throughput Stream Filter : smart_pipe (Zero-allocation Shannon entropy scoring)    │  │
│  │  • Concurrent Secret Scanner     : secret_scan (48-pattern multithreaded credential miner) │  │
│  │  • Knowledge Base Engine         : search_knowledge (Instant offline payload retrieval)    │  │
│  │  • Recon & Discovery Toolchain   : subfinder, httpx, nmap, katana, gau, ffuf, nuclei       │  │
│  │  • MCP Client-Side Bridge        : Puppeteer (Headless Browser DOM) & Fetch (Clean MD)     │  │
│  └────────────────────────────────────────────────────────────────────────────────────────────┘  │
│                                                │                                                 │
│                                                ▼                                                 │
│  ┌────────────────────────────────────────────────────────────────────────────────────────────┐  │
│  │                             Target Deliverables & Multi-Format Pipeline                    │  │
│  │  • Executive PDF Deliverable : REPORT.pdf (Print-ready CVSS scorecards & risk badges)      │  │
│  │  • Interactive Dashboard     : report.html (Standalone UI with Dark/Light styling)         │  │
│  │  • Index & Metadata          : SUMMARY.md & metadata.json (Auto-indexed via aggregator)    │  │
│  │  • Reproducible Exploits     : pocs/poc_<vuln_name>.py (Deterministic Python verification) │  │
│  └────────────────────────────────────────────────────────────────────────────────────────────┘  │
└──────────────────────────────────────────────────────────────────────────────────────────────────┘
```

---

## 📑 Automated Executive Reporting (PDF & HTML)

Cybermes features an integrated **Playwright Chromium PDF & HTML generator** (`tools/generate_pdf.py`) and native **Go Report Aggregator** (`aggregate_reports`). Whenever an assessment completes or `aggregate_reports <TARGET_SLUG>` is executed, Cybermes produces four structured deliverable formats simultaneously:

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
* **Vulnerability Breakdown**: Detailed markdown findings rendered into clean typography with reproducible request/response proofs.
* **Print-Ready CSS Paging**: Optimized margins, headers, page numbers, and breaks for direct PDF export.

---

## 🧠 Token Economy & Smart Output Filtering

Traditional AI security agents quickly exhaust context windows and suffer from attention degradation when reading thousands of raw terminal lines from tools like `katana` or `ffuf`. 

Cybermes solves this with a two-tiered token optimization architecture:

1. **High-Throughput Stream Filter (`smart_pipe`)**:
   - Built in pure Go for zero-latency stdout streaming.
   - Intercepts tool streams and dumps **100% of raw logs** to `recon/<TARGET_SLUG>/<tool>_raw.txt`.
   - Filters out static asset clutter (`.png`, `.css`, `.woff`) and 404 noise.
   - Streams only the **top 30–50 high-signal findings** (HTTP 200/301/403, unique parameters, API routes, high Shannon entropy secrets) to the AI context.
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

All tools are pre-compiled in `tools/bin/` and accessible on system `$PATH`:

| Tool | Primary Purpose | Standard Syntax |
| :--- | :--- | :--- |
| **subfinder** | Passive Subdomain Discovery | `subfinder -d <target> -silent` |
| **httpx** | Probing & Tech Detection | `httpx -silent -status-code -title -tech-detect` |
| **katana** | Crawler & SPA Endpoint Miner | `katana -u <url> -silent -depth 3` |
| **smart_pipe** | Stream Output Filter & Token Saver | `<tool_cmd> \| smart_pipe --target <SLUG> --tool <NAME>` |
| **ffuf** | Directory & Parameter Fuzzing | `ffuf -u <url>/FUZZ -w tools/wordlists/common.txt -mc 200,301,302,403` |
| **nuclei** | Vulnerability Verification | `nuclei -u <url> -tags cve,auth-bypass -silent` |
| **sqlmap** | SQL Injection Auditor | `sqlmap -u "<url>?id=1" --batch --banner` |
| **dalfox** | XSS Scanner & Parameter Analyzer | `dalfox url <url> --silence` |
| **secret_scan** | 48-Pattern Secret & Credential Miner | `secret_scan <target_files_or_dirs>` |
| **search_knowledge** | Fast Offline Payload & CheatSheet Search | `search_knowledge "<query>"` |
| **aggregate_reports** | Automated Report Aggregator & Indexer | `aggregate_reports <TARGET_SLUG>` |
| **generate_pdf.py**| Automated PDF/HTML Generator | `python3 tools/generate_pdf.py <TARGET_SLUG>` |
| **update_tools.sh** | Toolchain & Template Auto-Updater (Linux/macOS) | `./tools/update_tools.sh` |
| **update_tools.ps1**| Toolchain & Template Auto-Updater (Windows) | `powershell -File tools/update_tools.ps1` |
| **validate_skills.py**| Skill Pack Integrity & Health Auditor | `python tools/validate_skills.py` |
| **windows_compat_check.py** | Windows System Diagnostics | `python tools\windows_compat_check.py` |
| **Puppeteer MCP** | Browser DOM Automation | Native MCP tools for dynamic SPA testing & screenshot capture |
| **Fetch MCP** | Clean Web-to-Markdown Reader | Native MCP tool for token-efficient API inspection |

🪟 **Windows Users**: Run `.\cybermes.bat` (or `.\hermes.bat`) and use `.\setup_windows.ps1` for installation.

---

## 🚀 60-Second Fast Track Quickstart

Get Cybermes up and running in under a minute:

### Option A: Docker (Zero Local Setup) 🐳
```bash
git clone https://github.com/Zyrexnn/Cybermes.git && cd Cybermes
cp .env.example .env && nano .env
docker compose up -d
docker compose exec cybermes cybermes "Assess https://example.com"
```

### Option B: Linux / macOS Host (One Command) 🐧
```bash
git clone https://github.com/Zyrexnn/Cybermes.git && cd Cybermes
./setup.sh
nano .env
./cybermes "Assess https://example.com"
```

### Option C: Windows Host (Native PowerShell) 💻
```powershell
git clone https://github.com/Zyrexnn/Cybermes.git; cd Cybermes
.\setup_windows.ps1
notepad .env
.\cybermes.bat "Assess https://example.com"
```

> [!TIP]
> **Testing Locally?** You can test against the included mock vulnerable application:
> ```bash
> python3 examples/mock_vulnerable_app.py   # Starts local test server on :8888
> ./cybermes "Assess http://127.0.0.1:8888" # In a new terminal tab
> ```

---

## 📑 Installation Guide & Methods

Choose your preferred deployment method:

| Method | OS | Setup Time | Performance | Best For |
|--------|----|------------|-------------|----------|
| **PowerShell Native** | Windows 10/11 | ~1 min | ⭐⭐⭐⭐⭐ | Windows host, direct hardware access |
| **Native Host** | Linux/macOS | ~1 min | ⭐⭐⭐⭐⭐ | Linux & macOS native environments |
| **Docker (GHCR)** | All Platforms | <30 sec | ⭐⭐⭐⭐ | Zero-pollution isolated environment |
| **WSL2** | Windows → Linux | ~2 min | ⭐⭐⭐⭐⭐ | Windows power users |

---

### 💻 Installing on Windows (Native PowerShell)

```powershell
# 1. Clone repository
git clone https://github.com/Zyrexnn/Cybermes.git
cd Cybermes

# 2. Run automated installer (installs venv, Go tools & ProjectDiscovery binaries)
.\setup_windows.ps1

# 3. Configure your API keys
notepad .env

# 4. Verify system health
python tools\doctor.py

# 5. Launch an assessment
.\cybermes.bat "Assess https://example.com"
```

👉 **See [Windows Installation Guide](docs/INSTALL_WINDOWS.md) for step-by-step walkthrough**

---

### 🐧 Native Host Setup (Linux/macOS)

```bash
# 1. Clone repository
git clone https://github.com/Zyrexnn/Cybermes.git
cd Cybermes

# 2. Run automated installer
./setup.sh

# 3. Configure API keys in .env
nano .env

# 4. Verify system health
python3 tools/doctor.py

# 5. Launch an assessment
./cybermes "Assess https://example.com"
```

---

### 🐳 Docker & Docker Compose (Pre-built GHCR)

Run Cybermes inside a container with pre-packaged tools:

```bash
# 1. Clone and navigate to repository
git clone https://github.com/Zyrexnn/Cybermes.git
cd Cybermes

# 2. Configure environment variables
cp .env.example .env
nano .env

# 3. Start container
docker compose up -d

# 4. Execute assessment
docker compose exec cybermes cybermes "Assess https://example.com"
```

---

## 🆘 Getting Help & System Doctor

### Health Check & Auto-Repair Tool

Cybermes includes a universal diagnostic tool (`tools/doctor.py`) to verify and auto-repair your environment:

```bash
# Run diagnostics check
python tools/doctor.py

# Automatically repair missing folders, download missing tools & templates
python tools/doctor.py --fix
```

### Quick Links

- **[Windows Installation Guide](docs/INSTALL_WINDOWS.md)** - Complete Windows setup tutorial
- **[Troubleshooting Common Issues](docs/troubleshooting.md)** - Solutions for common problems
- **[Prompt Engineering Guide](docs/prompt_guide.md)** - Learn effective prompting techniques
- **[GitHub Discussions](https://github.com/Zyrexnn/Cybermes/discussions)** - Ask questions, share experiences
- **[Release Notes](https://github.com/Zyrexnn/Cybermes/releases)** - Latest updates and features

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

## 💬 Community & Discussions

We invite security researchers, bug bounty hunters, and developers to collaborate in our **[GitHub Discussions](https://github.com/Zyrexnn/Cybermes/discussions)**!

### 🧭 Discussion Categories

| Category | Description | Direct Link |
| :--- | :--- | :--- |
| **💬 Q&A** | Get help with installation, Windows/Docker troubleshooting, and LLM setup | [Go to Q&A](https://github.com/Zyrexnn/Cybermes/discussions/categories/q-a) |
| **💡 Ideas & RFCs** | Propose new security tools, autonomous agent skills, and architecture features | [Go to Ideas](https://github.com/Zyrexnn/Cybermes/discussions/categories/ideas) |
| **🚀 Show and Tell** | Share your custom setups, offensive workflows, and sanitized PoC demos | [Go to Show & Tell](https://github.com/Zyrexnn/Cybermes/discussions/categories/show-and-tell) |
| **📢 General** | Community conversations, security research news, and general feedback | [Go to General](https://github.com/Zyrexnn/Cybermes/discussions/categories/general) |

### 📜 Visitor Rules & Code of Conduct

Before posting, please review the complete **[Community Guidelines (.github/DISCUSSIONS.md)](.github/DISCUSSIONS.md)**:

1. **🛡️ Authorized Research Only**: Discussion must strictly concern authorized security testing, legitimate bug bounty programs, educational concepts, or defensive diagnostics. Never request or share attack vectors targeting unauthorized systems.
2. **🔒 Zero Sensitive Data**: Sanitize all logs, tokens, API keys, and target URLs. Use RFC 2606 standard domains (`example.com`) and RFC 5737 IPs (`198.51.100.x`).
3. **🤝 Respectful & Professional**: Maintain a welcoming and constructive atmosphere. Harassment, spam, and malicious activity are strictly prohibited.
4. **🚦 Where to Post**:
   - Software bugs in Cybermes → [Open an Issue](https://github.com/Zyrexnn/Cybermes/issues)
   - Security vulnerabilities within Cybermes itself → [Security Advisory (.github/SECURITY.md)](.github/SECURITY.md)
   - Questions, ideas & showcases → [GitHub Discussions](https://github.com/Zyrexnn/Cybermes/discussions)

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
   python3 examples/mock_vulnerable_app.py
   ```
   *Server listens on `http://127.0.0.1:8888`.*

2. **Instruct Cybermes to assess the target:**
   ```bash
   ./cybermes "Assess http://127.0.0.1:8888 and generate structured reports."
   ```


3. **Inspect the generated PDF and HTML deliverables:**
   ```bash
   ls -la reports/127_0_0_1_8888/
   # -> SUMMARY.md, metadata.json, report.html, REPORT.pdf
   ```

---

## 📈 Star History

<div align="center">

<img src="assets/star_history.svg" alt="Cybermes Stargazers Growth" width="100%" />

</div>

---

## 📦 Release & Version History
 
### **[v2.1.0](https://github.com/Zyrexnn/Cybermes/releases/tag/v2.1.0)** — *Frictionless Setup, Universal Toolchain & Health-Check Engine*
* **60-Second Fast Track Installation**: Single-command automated installers (`setup.sh` and `setup_windows.ps1`) for Linux, macOS (Intel & Apple Silicon M-series), and Windows PowerShell.
* **Universal System Doctor (`tools/doctor.py`)**: Cross-platform environment diagnostic utility featuring `--fix` flag to automatically inspect system health and auto-repair missing tools, folders, and configs.
* **Automated ProjectDiscovery Toolchain**: Integrated auto-downloading for `subfinder`, `httpx`, `katana`, `nuclei`, and automatic Nuclei template synchronization across `amd64` and `arm64` architectures.
* **Pre-built GHCR Container Pipeline**: Automated multi-arch Docker image publishing (`ghcr.io/zyrexnn/cybermes:latest` and `ghcr.io/zyrexnn/cybermes:v2.1.0`) with resilient directory bind mounts.
* **Standardized Python Package (PEP 621)**: Added `pyproject.toml` support for standard editable installs (`pip install -e .`).
* **Unified CLI Entrypoints & Clean Layout**: Standardized primary launchers (`cybermes`, `cybermes.bat`, `cybermes.ps1`), relocated mock targets to `examples/`, and eliminated redundant wrapper scripts.

### **[v2.0.0](https://github.com/Zyrexnn/Cybermes/releases/tag/v2.0.0)** — *The High-Performance Native Go Core Architecture*

* **Native Go Core Toolchain (`pkg/*` & `cmd/*`)**: Complete refactoring of core performance bottlenecks into zero-external-dependency, compiled Go binaries (`tools/bin/*`):
  * `smart_pipe`: High-throughput, zero-allocation stdout streaming and Shannon entropy filtering.
  * `secret_scan`: 48-pattern credential scanner with concurrent worker pools.
  * `search_knowledge`: Sub-50ms offline knowledge & payload search across local security wikis.
  * `aggregate_reports`: Resilient markdown matrix generator and structured metadata indexer.
* **Elimination of Python Script Latency**: Deleted deprecated utility wrappers (`smart_pipe.py`, `search_knowledge.py`, `aggregate_reports.py`) in favor of unified single-binary tools exposed directly on `$PATH`.
* **Multi-Platform Native Builder**: Automated Go tool compilation integrated into `setup.sh` (Linux/macOS), `setup_windows.ps1` (Windows native `.exe`), and `Dockerfile` (Multi-arch AMD64/ARM64 container builds).
* **Modernized Directives**: Standardized `AGENTS.md`, `.cursorrules`, and 50+ offensive skill playbooks for binary-native execution.

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
 
This project is licensed under the **[Apache License 2.0](LICENSE)**. Free for personal, research, academic, and commercial use under the terms of the license.
 
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

<div align="center">

<p>Thank you to all the amazing people who make <b>Cybermes</b> possible!</p>

<a href="https://github.com/Zyrexnn/Cybermes/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=Zyrexnn/Cybermes&max=50&columns=12" alt="Cybermes Contributors" />
</a>

<br><br>

<table style="border: none; border-collapse: collapse;">
  <tr>
    <td align="center" width="140px" style="padding: 10px; border: none;">
      <a href="https://github.com/Zyrexnn">
        <img src="https://github.com/Zyrexnn.png" width="80px" height="80px" style="border-radius: 50%; border: 3px solid #58a6ff; box-shadow: 0 0 15px rgba(88,166,255,0.6);" alt="Zyrexnn"/><br><br>
        <img src="https://img.shields.io/badge/Lead-Zyrexnn-58a6ff?style=flat-square" alt="Zyrexnn" />
      </a>
    </td>
    <td align="center" width="140px" style="padding: 10px; border: none;">
      <a href="https://github.com/claude">
        <img src="https://github.com/claude.png" width="80px" height="80px" style="border-radius: 50%; border: 3px solid #bc8cff; box-shadow: 0 0 15px rgba(188,140,255,0.6);" alt="Claude"/><br><br>
        <img src="https://img.shields.io/badge/AI_Research-Claude-bc8cff?style=flat-square" alt="Claude" />
      </a>
    </td>
    <td align="center" width="140px" style="padding: 10px; border: none;">
      <a href="https://github.com/msarg44">
        <img src="https://github.com/msarg44.png" width="80px" height="80px" style="border-radius: 50%; border: 3px solid #3fb950; box-shadow: 0 0 15px rgba(63,185,80,0.6);" alt="msarg44"/><br><br>
        <img src="https://img.shields.io/badge/Contributor-msarg44-3fb950?style=flat-square" alt="msarg44" />
      </a>
    </td>
    <td align="center" width="140px" style="padding: 10px; border: none;">
      <a href="https://github.com/Mortify4315">
        <img src="https://github.com/Mortify4315.png" width="80px" height="80px" style="border-radius: 50%; border: 3px solid #f0883e; box-shadow: 0 0 15px rgba(240,136,62,0.6);" alt="Mortify4315"/><br><br>
        <img src="https://img.shields.io/badge/Contributor-Mortify4315-f0883e?style=flat-square" alt="Mortify4315" />
      </a>
    </td>
  </tr>
</table>

<br>

<p>
  <i>Want to contribute? Check out our <a href="CONTRIBUTING.md">Contributing Guide</a>, review our <a href="CODE_OF_CONDUCT.md">Code of Conduct</a>, and submit a pull request to the <code>dev</code> branch!</i>
</p>

</div>
