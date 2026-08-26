<div align="center">

# Cybermes

**Autonomous Offensive Security & Bug Bounty Automation Framework**

[![Release](https://img.shields.io/badge/Release-v3.1.1-orange.svg)](https://github.com/Zyrexnn/Cybermes/releases)
[![Go: 1.22+](https://img.shields.io/badge/Go-1.22+-00ADD8.svg?logo=go&logoColor=white)](https://go.dev/)
[![Python: 3.11+](https://img.shields.io/badge/Python-3.11+-3776AB.svg?logo=python&logoColor=white)](https://www.python.org/)
[![MCP Server](https://img.shields.io/badge/MCP-Supported-purple.svg)](docs/MCP_SETUP.md)
[![License: Apache 2.0](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](LICENSE)
[![Platforms](https://img.shields.io/badge/Platform-Linux%20%7C%20macOS%20%7C%20Windows%20%7C%20Docker-brightgreen.svg)](#installation--setup)

<img src="assets/bannernew.png" alt="Cybermes Architecture & Operational Pipeline" width="100%" style="border-radius: 8px; margin: 18px 0;">

<p align="center">
  <b>Cybermes</b> is an offensive security assistant and automation framework designed for authorized bug bounty hunting, reconnaissance, vulnerability research, and structured reporting.
  <br>
  It combines native Go performance utilities, 200+ modular offensive security playbooks, token-optimized streaming pipelines, and full Model Context Protocol (MCP) support for AI coding environments.
</p>

[Installation](#installation--setup) • [Architecture](#architecture--operational-pipeline) • [Why Cybermes?](#why-cybermes) • [Features](#core-features) • [Workspace Structure](#target-workspace--deliverables) • [Toolchain](#toolchain) • [Documentation](docs/)

</div>

---

## Installation & Setup

Cybermes can be used directly through your AI assistant via **Model Context Protocol (MCP)** or executed as a **standalone CLI / pipeline**.

### 1. MCP Server Installation (Recommended for AI Workflows)

Cybermes includes a high-performance native Go MCP server (`cybermes-mcp`) that exposes 10+ security tools and context providers directly to AI coding environments (**Google Antigravity / Gemini**, **Kilo Code**, **Cursor**, **Claude Desktop**, **Windsurf**, **Cline**, **Roo Code**, **OpenCode**, **Claude Code CLI**, **Continue.dev**, **Zed**, **Hermes**, and **Codex**).

#### Universal Auto-Installer (1-Click)
```bash
# Auto-detect and configure all installed AI clients
npx -y cybermes-mcp install

# Install ONLY to specific AI providers:
npx -y cybermes-mcp install --kilo
npx -y cybermes-mcp install --gemini --cursor
```

#### Global Installation (No NPX Startup Latency)
```bash
npm install -g cybermes-mcp
cybermes-mcp install --global
```

#### Local Interactive MCP Manager & Optimizer
```bash
# Windows
.\mcp.bat            # or .\mcp.ps1

# Linux / macOS
./mcp.sh             # or python3 scripts/mcp.py
```

#### Manual Client Configuration
To manually register the MCP server in your client configuration (`mcpServers` section):
```json
{
  "mcpServers": {
    "cybermes": {
      "command": "npx",
      "args": ["-y", "cybermes-mcp"]
    }
  }
}
```

> For client-specific paths, direct flags (`--kilo`, `--gemini`, `--global`, `--dry-run`), and native binary setup, see the **[MCP Integration Guide](docs/MCP_SETUP.md)**.

---

### 2. Standalone CLI Installation

If you plan to run Cybermes directly from the terminal or in headless CI/CD pipelines:

#### Windows (Native PowerShell)
```powershell
# Clone repository and execute installer
git clone https://github.com/Zyrexnn/Cybermes.git
cd Cybermes
.\setup_windows.ps1

# Configure environment variables and API keys
notepad .env
```
*(For WSL2 or manual setups, refer to the [Windows Installation Guide](docs/INSTALL_WINDOWS.md))*

#### Linux / macOS
```bash
# Clone repository and execute installer
git clone https://github.com/Zyrexnn/Cybermes.git
cd Cybermes
./setup.sh

# Configure environment variables and API keys
nano .env
```

#### Docker Container
```bash
git clone https://github.com/Zyrexnn/Cybermes.git && cd Cybermes
cp .env.example .env && nano .env
docker compose up -d
```

---

### 3. Environment Health Check

Verify system dependencies, path bindings, and tool integrity:

```bash
# Run environment diagnostics
python tools/doctor.py

# Automatically resolve and repair missing components
python tools/doctor.py --fix
```

---

### 4. Running an Assessment

```bash
# Linux / macOS
./cybermes "Assess https://example.com"

# Windows
.\cybermes.bat "Assess https://example.com"

# Docker
docker compose exec cybermes cybermes "Assess https://example.com"
```

> **Local Testing**: Run the included mock vulnerable application in `examples/`:
> ```bash
> python examples/mock_vulnerable_app.py   # Starts mock app on http://127.0.0.1:8888
> ./cybermes "Assess http://127.0.0.1:8888"
> ```

---

## Architecture & Operational Pipeline

Cybermes operates through a closed-loop, deterministic offensive research pipeline:

```text
[Hermes Reasoning] ──> [Reconnaissance] ──> [Analysis & Skills] ──> [Validation Gate] ──> [Structured Reporting]
```

1. **Reasoning & Orchestration (`Hermes`)**: Autonomous decision loop evaluating scope, parameter maps, and attack paths via MCP tool invocations.
2. **Reconnaissance & Probing**: Passive asset discovery (`subfinder`, `gau`) and active crawling/probing (`httpx`, `katana`, `ffuf`).
3. **Analysis & Skill Execution**: High-speed output filtering (`smart_pipe`) streaming high-signal endpoints into 200+ specialized vulnerability SOPs and offline knowledge lookups (`search_knowledge`).
4. **Validation Gate (Zero False-Positive)**: Mandatory creation and execution of standalone, non-destructive validation scripts (`pocs/poc_<name>.py`) backed by raw HTTP proof traces.
5. **Structured Reporting**: Automated compilation into `reports/<TARGET_SLUG>/` (`SUMMARY.md`, `metadata.json`, interactive HTML, and `REPORT.pdf`).

---

## Why Cybermes?

| Challenge | Traditional Approach | Cybermes Solution |
| :--- | :--- | :--- |
| **AI Assistant Integration** | Manual copy-pasting of terminal outputs or fragile custom wrappers. | **Native Go MCP Server (`cybermes-mcp`)**: Standardized JSON-RPC 2.0 integration exposing 10+ security tools directly to AI clients. |
| **LLM Token & Context Noise** | Fuzzer/crawler outputs flood context with thousands of lines of 404s and static assets. | **High-Speed Stream Filter (`smart_pipe`)**: Pure Go filter that archives raw logs to disk while streaming only high-signal endpoints, status codes, and secrets. |
| **Exploit Verification** | Speculative or pattern-matched alerts without reproducible verification. | **Zero-False-Positive PoC Gate**: Requires standalone, non-destructive Python scripts (`pocs/poc_<name>.py`) and raw HTTP traces before logging findings. |
| **Payload & Methodology Retrieval** | Manually searching external wikis, repositories, and cheat sheets online. | **Sub-50ms Local Knowledge Base (`search_knowledge`)**: Instant offline query engine across 200+ SOP playbooks, PayloadsAllTheThings, and HackTricks datasets. |
| **Structured Deliverables** | Unorganized text dumps requiring tedious manual report compilation. | **Multi-Format Report Aggregator**: Automated generation of `SUMMARY.md`, `metadata.json`, standalone interactive HTML, and print-ready `REPORT.pdf`. |
| **OS & Environment Portability** | Most offensive tools assume Kali/Linux, complicating Windows setups. | **Native Multi-Platform Support**: Automated native installers for Windows (PowerShell), Linux, macOS, and Docker with a built-in health check and repair tool (`doctor.py`). |

---

## Core Features

- **Native Go Toolchain**: Zero-overhead compiled utilities for output stream filtering (`smart_pipe`), credential scanning (`secret_scan`), offline knowledge retrieval (`search_knowledge`), and report aggregation (`aggregate_reports`).
- **Model Context Protocol (MCP)**: Native Go MCP server (`cybermes-mcp`) exposing security tools and context providers to AI assistants like Claude Desktop, Cursor, Windsurf, and VS Code / Cline.
- **200+ Offensive Playbooks**: Standard operating procedures in `skills/` covering API security (IDOR/BOLA, JWT, BPLA), web vulnerabilities (SSRF, XSS, SQLi, Race Conditions), and cloud misconfigurations.
- **Integrated Reconnaissance**: Pre-configured pipelines for `subfinder`, `httpx`, `katana`, `ffuf`, `nuclei`, and `sqlmap`.
- **Target Workspace Isolation**: Strict target-scoped evidence tracking, preventing context contamination across engagements.
- **Multi-Format Reporting**: Automated output generation in Markdown (`SUMMARY.md`), JSON metrics (`metadata.json`), standalone HTML dashboards, and PDF reports (`REPORT.pdf`).
- **Cross-Platform Compatibility**: Fully supported on Windows (native PowerShell), Linux, macOS, and Docker.

---

## Target Workspace & Deliverables

Every target assessment creates a dedicated directory structure under `reports/` and `recon/`:

```text
Cybermes/
├── reports/<TARGET_SLUG>/        # Verified findings & deliverables
│   ├── SUMMARY.md                # Consolidated executive findings matrix
│   ├── metadata.json             # Structured JSON metrics & counters
│   ├── report.html               # Standalone interactive HTML report
│   ├── REPORT.pdf                # Print-ready PDF report
│   ├── findings/                 # Confirmed vulnerability writeups (low, medium, high, critical)
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

## Toolchain

| Tool | Type | Purpose | Standard Syntax |
| :--- | :--- | :--- | :--- |
| **`smart_pipe`** | Go Binary | Stream output filter & context token optimizer | `katana -u <url> \| smart_pipe --target <slug> --tool katana` |
| **`secret_scan`** | Go Binary | Multi-pattern credential leak scanner | `secret_scan recon/<slug>/` |
| **`search_knowledge`**| Go Binary | Sub-50ms offline exploit & payload search | `search_knowledge "jwt algorithm confusion"` |
| **`aggregate_reports`**| Go Binary | Markdown findings indexer & report compiler | `aggregate_reports <slug>` |
| **`cybermes-mcp`** | Go Binary / NPX | Native MCP Server for AI assistants | `npx -y cybermes-mcp` |
| **`subfinder`** | External | Passive subdomain discovery | `subfinder -d <target> -silent` |
| **`httpx`** | External | Web probing & technology fingerprinting | `httpx -u <url> -tech-detect` |
| **`katana`** | External | Web crawler & SPA endpoint miner | `katana -u <url> -depth 3` |
| **`ffuf`** | External | High-speed web fuzzer | `ffuf -u <url>/FUZZ -w <wordlist>` |
| **`nuclei`** | External | Template-based vulnerability scanner | `nuclei -u <url> -tags cve` |
| **`sqlmap`** | Python | Automated SQL injection auditor | `sqlmap -u "<url>?id=1" --batch` |
| **`doctor.py`** | Python | Environment health check & auto-repair | `python tools/doctor.py --fix` |
| **`generate_pdf.py`** | Python | PDF and HTML executive report generator | `python tools/generate_pdf.py <slug>` |

---

## Documentation

Detailed guides and references are available in the [`docs/`](docs/) directory:

- [Windows Installation Guide](docs/INSTALL_WINDOWS.md) — Setup via PowerShell, WSL2, and troubleshooting.
- [MCP Setup Guide](docs/MCP_SETUP.md) — Integration guide for all supported AI coding editors and clients.
- [Prompt Engineering Guide](docs/prompt_guide.md) — Structuring security research prompts effectively.
- [Telegram Gateway Setup](docs/telegram_setup.md) — Configuring remote execution via Telegram.
- [Tools & Skills Reference](docs/tools_and_skills.md) — Breakdown of offensive playbooks and built-in tools.
- [Troubleshooting & FAQ](docs/troubleshooting.md) — Common error resolutions and diagnostics.

---

## Acknowledgments & Upstream Credits

Cybermes integrates and builds upon foundational work from the open-source and offensive security research communities:

| Project / Tool | Maintainer / Author | Contribution / Reference |
| :--- | :--- | :--- |
| [HackTricks](https://github.com/carlospolop/hacktricks) | [@carlospolop](https://github.com/carlospolop) | Privilege escalation, service exploitation & offensive knowledge base |
| [PayloadsAllTheThings](https://github.com/swisskyrepo/PayloadsAllTheThings) | [@swisskyrepo](https://github.com/swisskyrepo) | Web application payloads and bypass vectors |
| [ProjectDiscovery Suite](https://projectdiscovery.io/) | ProjectDiscovery Team | Core toolchain (`nuclei`, `httpx`, `subfinder`, `katana`) |
| [SQLMap](https://github.com/sqlmapproject/sqlmap) | Bernardo Damele & Miroslav Stampar | Automated SQL injection auditor |
| [Claude-BugHunter](https://github.com/sachinsharma-96/Claude-BugHunter) | [@sachinsharma-96](https://github.com/sachinsharma-96) | Bug bounty engagement patterns & validation playbooks |
| [Strix Framework](https://github.com/strix-security/strix) | Strix Security Team | Autonomous multi-agent coordination architecture |
| [FFuF](https://github.com/ffuf/ffuf) | [@joohoi](https://github.com/joohoi) | High-speed web fuzzer |
| [Hermes-Agent](https://github.com/NousResearch/Hermes-Agent) | NousResearch | Autonomous agent core framework |

*(See [ATTRIBUTION.md](ATTRIBUTION.md) for full license notices and upstream details)*

---

## Contributors

Thank you to everyone who helps build, maintain, and research Cybermes:

<div align="center">

<table>
  <tr>
    <td align="center" width="160px">
      <a href="https://github.com/Zyrexnn">
        <img src="https://github.com/Zyrexnn.png?size=120" width="80px" height="80px" alt="Zyrexnn" style="border-radius: 50%; border: 3px solid #00d2ff; box-shadow: 0 0 14px rgba(0, 210, 255, 0.7); padding: 2px;" /><br />
        <sub><b>Zyrexnn</b></sub>
      </a><br />
      <a href="https://github.com/Zyrexnn"><img src="https://img.shields.io/badge/★-Project_Lead-00d2ff?style=flat-square" alt="Project Lead" /></a>
    </td>
    <td align="center" width="130px">
      <a href="https://github.com/msarg44">
        <img src="https://github.com/msarg44.png?size=100" width="65px" height="65px" alt="msarg44" style="border-radius: 50%; border: 2px solid #2ea44f; padding: 2px;" /><br />
        <sub><b>msarg44</b></sub>
      </a><br />
      <a href="https://github.com/msarg44"><img src="https://img.shields.io/badge/Fork_%2F_PR-2ea44f?style=flat-square" alt="Fork / PR" /></a>
    </td>
    <td align="center" width="130px">
      <a href="https://github.com/Mortify4315">
        <img src="https://github.com/Mortify4315.png?size=100" width="65px" height="65px" alt="Mortify4315" style="border-radius: 50%; border: 2px solid #2ea44f; padding: 2px;" /><br />
        <sub><b>Mortify4315</b></sub>
      </a><br />
      <a href="https://github.com/Mortify4315"><img src="https://img.shields.io/badge/Fork_%2F_PR-2ea44f?style=flat-square" alt="Fork / PR" /></a>
    </td>
    <td align="center" width="130px">
      <a href="https://github.com/xsoft">
        <img src="https://github.com/xsoft.png?size=100" width="65px" height="65px" alt="xsoft" style="border-radius: 50%; border: 2px solid #8957e5; padding: 2px;" /><br />
        <sub><b>xsoft</b></sub>
      </a><br />
      <a href="https://github.com/xsoft"><img src="https://img.shields.io/badge/Accepted_Issue-8957e5?style=flat-square" alt="Accepted Issue" /></a>
    </td>
    <td align="center" width="130px">
      <a href="https://github.com/Muzakie-ID">
        <img src="https://github.com/Muzakie-ID.png?size=100" width="65px" height="65px" alt="Muzakie-ID" style="border-radius: 50%; border: 2px solid #8957e5; padding: 2px;" /><br />
        <sub><b>Muzakie-ID</b></sub>
      </a><br />
      <a href="https://github.com/Muzakie-ID"><img src="https://img.shields.io/badge/Accepted_Issue-8957e5?style=flat-square" alt="Accepted Issue" /></a>
    </td>
  </tr>
</table>

</div>

| Contributor | Type | Contribution | Reference |
| :--- | :--- | :--- | :--- |
| **[@Zyrexnn](https://github.com/Zyrexnn)** | `Project Lead` | Creator, Core Architecture & Offensive Framework | Main |
| **[@msarg44](https://github.com/msarg44)** | `Fork / PR` | Playwright PDF rendering engine fix | [#1](https://github.com/Zyrexnn/Cybermes/issues/1) |
| **[@Mortify4315](https://github.com/Mortify4315)** | `Fork / PR` | Windows Python launcher fallback & Long Path documentation | [#4](https://github.com/Zyrexnn/Cybermes/issues/4), [#5](https://github.com/Zyrexnn/Cybermes/issues/5) |
| **[@xsoft](https://github.com/xsoft)** | `Accepted Issue` | Linux setup audit, Docker config mounts & workflow diagnostic report | [#7](https://github.com/Zyrexnn/Cybermes/issues/7) |
| **[@Muzakie-ID](https://github.com/Muzakie-ID)** | `Accepted Issue` | Windows PowerShell setup & script parser bug report | [#10](https://github.com/Zyrexnn/Cybermes/issues/10) |

Want to contribute? Check out our [Contributing Guide](CONTRIBUTING.md) and [Contributors List](CONTRIBUTORS.md).

---

## License

This project is licensed under the [Apache License 2.0](LICENSE).

Third-party research materials, datasets, and upstream tools referenced or incorporated within this repository retain their respective original licenses (see [ATTRIBUTION.md](ATTRIBUTION.md)).

---

## Legal & Ethical Disclaimer

> **IMPORTANT**: Cybermes is developed exclusively for **authorized security testing**, **legitimate bug bounty research**, and **academic security education**.
> 
> Testing against targets without explicit, prior written permission from the target owner is illegal and strictly prohibited. The maintainers assume no liability and are not responsible for any misuse, damage, or legal consequences resulting from this tool.
