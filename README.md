<div align="center">

<img src="assets/banner.jpg" alt="Cybermes Banner" width="100%" style="border-radius: 10px; margin-bottom: 20px;">

# 🛡️ Cybermes

### **Autonomous Offensive Security & Bug Bounty Automation Framework**

[![Release](https://img.shields.io/badge/Release-v3.0.0-orange.svg)](https://github.com/Zyrexnn/Cybermes/releases)
[![Go: 1.22+](https://img.shields.io/badge/Go-1.22+-00ADD8.svg?logo=go&logoColor=white)](https://go.dev/)
[![Python: 3.11+](https://img.shields.io/badge/Python-3.11+-3776AB.svg?logo=python&logoColor=white)](https://www.python.org/)
[![License: Apache 2.0](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](LICENSE)
[![Platforms](https://img.shields.io/badge/Platform-Linux%20%7C%20macOS%20%7C%20Windows%20%7C%20Docker-brightgreen.svg)](#-installation--setup)
[![MCP Server](https://img.shields.io/badge/MCP-Supported-purple.svg)](docs/MCP_SETUP.md)

<p align="center">
  <b>Cybermes</b> is an offensive security assistant and automation framework designed for authorized bug bounty hunting, reconnaissance, vulnerability research, and structured reporting.
  <br>
  It integrates native Go performance tools, 200+ modular offensive skills, automated recon pipelines, and Model Context Protocol (MCP) support for AI clients.
</p>

[Quickstart](#-quickstart) • [Why Cybermes?](#-why-cybermes) • [Features](#-core-features) • [Installation](#-installation--setup) • [MCP Integration](#-mcp-server-integration) • [Workflow](#-operational-workflow) • [Toolchain](#-toolchain) • [Documentation](docs/)

</div>

---

## 💡 Why Cybermes?

| Capability / Challenge | Traditional Workflow | Cybermes Solution |
| :--- | :--- | :--- |
| **AI Assistant Integration** | Manual copy-pasting of terminal outputs or fragile custom wrappers. | **Native Go MCP Server (`cybermes-mcp`)**: 1-click integration exposing 10+ security tools directly to Cursor, Claude Desktop, Windsurf, and Cline. |
| **LLM Token & Context Noise** | Fuzzer/crawler outputs flood context with thousands of lines of 404s and static assets. | **High-Speed Stream Filter (`smart_pipe`)**: Pure Go filter that archives raw logs to disk while streaming only high-signal endpoints, status codes, and secrets to the LLM. |
| **Exploit Verification** | Speculative or pattern-matched alerts without reproducible verification. | **Zero-False-Positive PoC Gate**: Requires standalone, non-destructive Python scripts (`pocs/poc_<name>.py`) and raw HTTP traces before logging findings. |
| **Payload & Methodology Retrieval** | Manually searching external wikis, repositories, and cheat sheets online. | **Sub-50ms Local Knowledge Base (`search_knowledge`)**: Instant offline query engine across 200+ SOP playbooks, PayloadsAllTheThings, and HackTricks datasets. |
| **Structured Deliverables** | Unorganized text dumps requiring tedious manual report compilation. | **Multi-Format Report Aggregator**: Automated generation of `SUMMARY.md`, `metadata.json`, standalone interactive HTML, and print-ready `REPORT.pdf`. |
| **OS & Environment Portability** | Most offensive tools assume Kali/Linux, complicating Windows setups. | **Native Multi-Platform Support**: Automated native installers for Windows (PowerShell), Linux, macOS, and Docker with a built-in health check and repair tool (`doctor.py`). |

---

## 🚀 Core Features

- **⚡ Native Go High-Speed Toolchain**: Zero-overhead compiled utilities for output streaming filter (`smart_pipe`), multi-pattern credential scanner (`secret_scan`), sub-50ms offline knowledge search (`search_knowledge`), and report aggregation (`aggregate_reports`).
- **🔌 Model Context Protocol (MCP)**: Native Go MCP server (`cybermes-mcp`) providing 10+ security tools and context providers to AI assistants like Claude Desktop, Cursor, Windsurf, and VS Code/Cline.
- **🎯 200+ Offensive Playbooks & Methodologies**: Comprehensive SOPs in `skills/` covering API security (IDOR/BOLA, JWT, BPLA), web vulnerabilities (SSRF, XSS, SQLi, Race Conditions), and cloud misconfigurations.
- **🔍 Reconnaissance & Probing Integration**: Out-of-the-box integration with industry-standard tools (`subfinder`, `httpx`, `katana`, `ffuf`, `nuclei`, `sqlmap`).
- **📑 Structured Multi-Format Reporting**: Automatic aggregation of findings into structured target workspaces with Markdown (`SUMMARY.md`), JSON (`metadata.json`), HTML dashboards, and print-ready PDF reports (`REPORT.pdf`).
- **💻 Cross-Platform Support**: First-class support for Windows (native PowerShell / `.bat`), Linux, macOS, and Docker.

---

## ⚡ Quickstart

### 1. Installation

Choose your platform:

#### 🪟 Windows (Native PowerShell)
```powershell
# Clone and run the automated installer
git clone https://github.com/Zyrexnn/Cybermes.git
cd Cybermes
.\setup_windows.ps1

# Configure your API keys in .env
notepad .env
```
*(See [Windows Guide](docs/INSTALL_WINDOWS.md) for WSL2 or manual instructions)*

#### 🐧 Linux / macOS
```bash
# Clone and run the automated installer
git clone https://github.com/Zyrexnn/Cybermes.git
cd Cybermes
./setup.sh

# Configure your API keys in .env
nano .env
```

#### 🐳 Docker
```bash
git clone https://github.com/Zyrexnn/Cybermes.git && cd Cybermes
cp .env.example .env && nano .env
docker compose up -d
```

---

### 2. Verify Installation

Run the built-in system doctor to verify dependencies and health:

```bash
# Check environment and toolchain
python tools/doctor.py

# Auto-repair missing tools or directories if needed
python tools/doctor.py --fix
```

---

### 3. Launching an Assessment

```bash
# Linux / macOS
./cybermes "Assess https://example.com"

# Windows
.\cybermes.bat "Assess https://example.com"

# Docker
docker compose exec cybermes cybermes "Assess https://example.com"
```

> **Testing locally?** Run the mock test app in `examples/`:
> ```bash
> python examples/mock_vulnerable_app.py   # Runs local server on http://127.0.0.1:8888
> ./cybermes "Assess http://127.0.0.1:8888"
> ```

---

## 🔌 MCP Server Integration

Cybermes includes a high-performance native Go MCP server (`cybermes-mcp`) to supercharge your AI coding assistant with security tools.

### 1-Click Auto-Installer
```bash
# Automatically detects and configures Cursor, Claude Desktop, Windsurf, Cline, etc.
npx -y cybermes-mcp install
```

Or run the offline Python injector:
```bash
python scripts/setup_mcp.py
```

See [docs/MCP_SETUP.md](docs/MCP_SETUP.md) for manual JSON configs, flags (`--status`, `--dry-run`), and full tool specifications.

---

## 🔄 Operational Workflow

Cybermes organizes testing systematically across structured phases:

```text
[1. Passive Recon] ──> [2. Active Probing] ──> [3. Skill Execution]
                                                       │
[6. Exec Reporting] <── [5. Secret Mining]  <── [4. PoC Validation]
```

1. **Passive Recon**: Subdomain discovery (`subfinder`) and archive mining (`gau`).
2. **Active Probing**: Service discovery (`httpx`), crawler & endpoint discovery (`katana`, `ffuf`) via `smart_pipe`.
3. **Skill Execution**: Application of relevant vulnerability playbooks from `skills/` and payload retrieval via `search_knowledge`.
4. **PoC Validation**: Deterministic exploit validation with non-destructive standalone Python scripts.
5. **Secret Mining**: Scanning responses and downloaded bundles for leaked credentials (`secret_scan`).
6. **Executive Reporting**: Aggregation into `reports/<TARGET_SLUG>/` (`SUMMARY.md`, `report.html`, `REPORT.pdf`).

---

## 📁 Target Workspace & Deliverables

Every target assessment creates a dedicated directory structure under `reports/` and `recon/`:

```text
Cybermes/
├── reports/<TARGET_SLUG>/        # Verified findings & deliverables
│   ├── SUMMARY.md                # Consolidated executive findings matrix
│   ├── metadata.json             # Structured JSON metrics & counters
│   ├── report.html               # Standalone interactive HTML report
│   ├── REPORT.pdf                # Print-ready PDF report
│   ├── findings/                 # Confirmed vulnerability writeups (LOW/MED/HIGH/CRIT)
│   │   └── high_idor_orders.md
│   ├── pocs/                     # Standalone Python proof-of-concept scripts
│   │   └── poc_idor_orders.py
│   └── evidence/                 # Raw logs, screenshots, recon notes
│       └── recon_notes.md
└── recon/<TARGET_SLUG>/          # Raw tool outputs & streaming dumps
    ├── subdomains.txt
    └── endpoints.txt
```

---

## 🧰 Toolchain

| Tool | Type | Purpose | Example Command |
| :--- | :--- | :--- | :--- |
| **`smart_pipe`** | Go Binary | Output filter & stream token optimizer | `katana -u <url> \| smart_pipe --target <slug> --tool katana` |
| **`secret_scan`** | Go Binary | Multi-pattern credential leak scanner | `secret_scan recon/<slug>/` |
| **`search_knowledge`**| Go Binary | Offline exploit & payload search | `search_knowledge "jwt algorithm confusion"` |
| **`aggregate_reports`**| Go Binary | Markdown findings indexer & compiler | `aggregate_reports <slug>` |
| **`cybermes-mcp`** | Go Binary / NPX | Native MCP Server for AI assistants | `npx -y cybermes-mcp` |
| **`subfinder`** | External | Passive subdomain discovery | `subfinder -d <target> -silent` |
| **`httpx`** | External | Web probing & technology fingerprinting | `httpx -u <url> -tech-detect` |
| **`katana`** | External | Web crawler & SPA endpoint miner | `katana -u <url> -depth 3` |
| **`ffuf`** | External | High-speed web fuzzer | `ffuf -u <url>/FUZZ -w <wordlist>` |
| **`nuclei`** | External | Template-based vulnerability scanner | `nuclei -u <url> -tags cve` |
| **`sqlmap`** | Python | Automated SQL injection auditor | `sqlmap -u "<url>?id=1" --batch` |
| **`doctor.py`** | Python | Environment health check & repair | `python tools/doctor.py --fix` |
| **`generate_pdf.py`** | Python | PDF/HTML executive report builder | `python tools/generate_pdf.py <slug>` |

---

## 📚 Documentation

Detailed documentation is available in the [`docs/`](docs/) directory:

- 📖 **[Windows Installation Guide](docs/INSTALL_WINDOWS.md)**: Setup via PowerShell, WSL2, and troubleshooting.
- 🔌 **[MCP Setup Guide](docs/MCP_SETUP.md)**: Configuration guide for all supported AI coding editors.
- 🎯 **[Prompt Engineering Guide](docs/prompt_guide.md)**: Structuring security research prompts effectively.
- 🤖 **[Telegram Gateway Setup](docs/telegram_setup.md)**: Setting up remote control via Telegram.
- 🛠️ **[Tools & Skills Reference](docs/tools_and_skills.md)**: In-depth breakdown of playbooks and tools.
- ❓ **[Troubleshooting & FAQ](docs/troubleshooting.md)**: Common errors and solutions.

---

## 🙏 Acknowledgments & Upstream Credits

Cybermes builds upon and integrates the valuable work of the open-source and offensive security research communities:

| Project / Tool | Maintainer / Author | Contribution / Reference |
| :--- | :--- | :--- |
| **[HackTricks](https://github.com/carlospolop/hacktricks)** | [@carlospolop](https://github.com/carlospolop) | Privilege escalation, service exploitation & offensive knowledge base |
| **[PayloadsAllTheThings](https://github.com/swisskyrepo/PayloadsAllTheThings)** | [@swisskyrepo](https://github.com/swisskyrepo) | Web application payloads and bypass vectors |
| **[ProjectDiscovery Suite](https://projectdiscovery.io/)** | ProjectDiscovery Team | Core toolchain (`nuclei`, `httpx`, `subfinder`, `katana`) |
| **[SQLMap](https://github.com/sqlmapproject/sqlmap)** | Bernardo Damele & Miroslav Stampar | Automated SQL injection auditor |
| **[Claude-BugHunter](https://github.com/sachinsharma-96/Claude-BugHunter)** | [@sachinsharma-96](https://github.com/sachinsharma-96) | Bug bounty engagement patterns & validation playbooks |
| **[Strix Framework](https://github.com/strix-security/strix)** | Strix Security Team | Autonomous multi-agent coordination architecture |
| **[FFuF](https://github.com/ffuf/ffuf)** | [@joohoi](https://github.com/joohoi) | High-speed web fuzzer |
| **[Hermes-Agent](https://github.com/NousResearch/Hermes-Agent)** | NousResearch | Autonomous agent core framework |

*(See [ATTRIBUTION.md](ATTRIBUTION.md) for full license notices and upstream details)*

---

## 👥 Contributors

Thank you to all contributors who help build, maintain, and research **Cybermes**:

- **[Zyrexnn](https://github.com/Zyrexnn)** — Project Lead & Architecture
- **[msarg44](https://github.com/msarg44)** — Playwright PDF rendering engine fix
- **[Mortify4315](https://github.com/Mortify4315)** — Windows Python launcher & long path docs
- **[xsoft](https://github.com/xsoft)** — Linux setup audit, Docker config mounts & workflow diagnostics
- **[Muzakie-ID](https://github.com/Muzakie-ID)** — Windows PowerShell setup & parser fixes

Want to contribute? Check out our [Contributing Guide](CONTRIBUTING.md) and [Contributors List](CONTRIBUTORS.md).

---

## ⚖️ License

This project is licensed under the **[Apache License 2.0](LICENSE)**.

Third-party research materials, datasets, and upstream tools referenced or incorporated within this repository retain their respective original licenses (see [ATTRIBUTION.md](ATTRIBUTION.md)).

---

## ⚠️ Legal & Ethical Disclaimer

> **IMPORTANT**: Cybermes is developed exclusively for **authorized security testing**, **legitimate bug bounty research**, and **academic security education**.
> 
> Testing against targets without explicit, prior written permission from the target owner is illegal and strictly prohibited. The maintainers assume no liability and are not responsible for any misuse, damage, or legal consequences resulting from this tool.
