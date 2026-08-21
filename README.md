<div align="center">

<img src="assets/banner.jpg" alt="Cybermes Autonomous Security Agent Banner" width="100%" style="border-radius: 10px; margin-bottom: 20px;">

# 🛡️ Cybermes

### **Autonomous Offensive Security, Bug Bounty & Red Teaming Agent Framework**

[![Release: v1.3.0](https://img.shields.io/badge/Release-v1.3.0-orange.svg)](https://github.com/Zyrexnn/Cybermes/releases/tag/v1.3.0)
[![License: PolyForm Noncommercial](https://img.shields.io/badge/License-PolyForm_Noncommercial-blue.svg)](LICENSE)
[![Python: 3.11+](https://img.shields.io/badge/Python-3.11+-blue.svg)](https://www.python.org/)
[![Docker: Ready](https://img.shields.io/badge/Docker-Supported-2496ED.svg)](docker-compose.yml)
[![Hermes: Powered](https://img.shields.io/badge/Hermes%20Agent-Core-purple.svg)](https://github.com/NousResearch/Hermes-Agent)
[![PDF Reporting: Automated](https://img.shields.io/badge/Reports-PDF%20%26%20HTML%20Automated-brightgreen.svg)](#-automated-executive-reporting-pdf--html)
[![Token Economy: 85% Saved](https://img.shields.io/badge/Token%20Economy-Smart%20Filter-blueviolet.svg)](#-token-economy--smart-output-filtering)
[![AI Standards: AGENTS.md](https://img.shields.io/badge/AI%20Standards-AGENTS.md%20%2B%20.cursorrules-success.svg)](AGENTS.md)

<p align="center">
  <b>Cybermes</b> is an enterprise-grade, autonomous security research agent designed for high-signal reconnaissance, attack surface discovery, authenticated vulnerability research, zero-false-positive exploit validation, token-efficient context management, and automated executive PDF/HTML report generation.
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
  - [Method 1: Native Host Setup (Recommended)](#method-1-native-host-setup-recommended)
  - [Method 2: Docker & Docker Compose](#method-2-docker--docker-compose)
- [🤖 Telegram Bot Gateway](#-telegram-bot-gateway)
- [🎯 Prompt Engineering & Anti-Filter Guidelines](#-prompt-engineering--anti-filter-guidelines)
- [🧪 Local Validation with Mock Target](#-local-validation-with-mock-target)
- [📦 Release & Version History](#-release--version-history)
- [⚖️ License](#️-license)
- [⚠️ Legal & Ethical Disclaimer](#️-legal--ethical-disclaimer)

---

## 🏛️ Architecture & Core Engine

```text
┌──────────────────────────────────────────────────────────────────────────────────┐
│                             CYBERMES ENGINE v1.3.0                               │
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

Cybermes v1.3.0 features an integrated **Playwright Chromium PDF & HTML generator** (`tools/generate_pdf.py`). Whenever an assessment completes or `python3 tools/aggregate_reports.py <TARGET_SLUG>` is executed, Cybermes produces four structured deliverable formats simultaneously:

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
| **Puppeteer MCP** | Browser DOM Automation | Native MCP tools for dynamic SPA testing & screenshot capture |
| **Fetch MCP** | Clean Web-to-Markdown Reader | Native MCP tool for token-efficient API inspection |

---

## 🚀 Installation & Quick Start

### Method 1: Native Host Setup (Recommended)

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

### Method 2: Docker & Docker Compose

Run Cybermes inside a fully isolated, rootless-friendly container with live permission auto-healing:

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

## 📦 Release & Version History

### **[v1.3.0](https://github.com/Zyrexnn/Cybermes/releases/tag/v1.3.0)** — *Automated PDF/HTML Reporting, Token Economy & Universal AI Standards*
* **Automated PDF & HTML Generator**: Built-in `tools/generate_pdf.py` using Playwright Headless Chromium to output print-ready `REPORT.pdf` and standalone `report.html` dashboards automatically.
* **Smart CLI Output Filter (`tools/smart_pipe.py`)**: Streams top-signal findings to AI context while saving full logs to disk, saving 70–85% token consumption.
* **Live Docker Permission Daemon**: Real-time permission keeper in `entrypoint.sh` and POSIX default ACLs in `setup.sh` eliminating `NoPermissions` errors permanently.
* **Universal AI Standards**: Added `AGENTS.md` master directives and `.cursorrules` coding standards for cross-platform AI pair programming.
* **Zero-Leak Credential Architecture**: Sanitized `.hermes/config.yaml.example` and dynamic environment variable injection via `setup.sh`.
* **New Offensive Skills**: Added `custom-ai-router-assessment`, `blackbox-web-audit`, and `engagement-deliverables-and-validation`.

### **[v1.2.0](https://github.com/Zyrexnn/Cybermes/releases/tag/v1.2.0)** — *Target-Scoped Reporting & Native Host Setup*
* **Target-Scoped Directory Hierarchy**: Organized findings, PoCs, and evidence per target slug (`reports/<target>/findings/`).
* **Automated Report Aggregator**: Built-in `tools/aggregate_reports.py` compiling `SUMMARY.md` matrices and `metadata.json`.
* **Native Host Installer**: Automated 1-click installer `setup.sh` with dynamic path resolution.

### **[v1.1.0](https://github.com/Zyrexnn/Cybermes/releases/tag/v1.1.0)** — *Cybermes Identity & Prompt Architecture Update*
* **Unified Cybermes Persona**: Streamlined system prompt and SOUL persona with automatic target authorization handling.
* **Anti-Filter Prompt Architecture**: Comprehensive [Prompt Engineering & Anti-Filter Guide](docs/prompt_guide.md).
* **Telegram Messaging Gateway**: Remote bot integration with session management.

### **[v1.0.0](https://github.com/Zyrexnn/Cybermes/releases/tag/v1.0.0)** — *Initial Production Release*
* **Autonomous Core Architecture**: Multi-step offensive security reasoning loop built upon the Hermes runtime.
* **50+ Offensive Security Skills**: Bundled playbooks for IDOR/BOLA, SQLi, SSRF, SSTI, SAML, OAuth, and Cloud/K8s.
* **Offline Knowledge Aggregation**: Embedded knowledge bases from *HackTricks*, *PayloadsAllTheThings*, *Claude-BugHunter*, and *Strix*.

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

- **[Zyrexnn](https://github.com/Zyrexnn)** — Lead Author & Architect
- **[@claude](https://github.com/claude)** — AI Co-Author & Security Architecture Research
