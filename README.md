# Cybermes 🛡️🤖

> **Autonomous Offensive Security, Bug Bounty & Red Teaming Agent Framework** powered by Hermes Agent, specialized security skills, toolchain orchestration, and multi-model LLM reasoning.

---

## 📖 Overview

**Cybermes** is an end-to-end autonomous security research workspace that integrates:
- **Hermes Agent Core**: Autonomous agent reasoning and tool execution loop.
- **Offensive Security Skills**: 50+ modular security reasoning units (IDOR, SSRF, XSS, SQLi, Auth Bypass, OAuth, Cloud, SAML, etc.).
- **Security Knowledge Base**: Curated references from PayloadsAllTheThings, HackTricks, Claude-BugHunter, and Strix skills.
- **Recon & Web Tooling**: Unified toolchain integration (`nuclei`, `httpx`, `subfinder`, `katana`, `ffuf`, `dalfox`, `sqlmap`, `amass`, `gau`, `waybackurls`, `nmap`).
- **Containerized Architecture**: Production-ready Docker & Docker Compose setup with pre-configured Playwright / Puppeteer browser MCP support.
- **Reporting Engine**: Automated CVSS v3.1 report generation with step-by-step reproduction and minimal-impact PoC scripts.

---

## 📂 Project Structure

```text
Cybermes/
├── .dockerignore                     # Docker build exclusions
├── .env.example                      # Environment configuration template
├── .gitignore                        # Git ignore rules for safety & clean repo
├── Dockerfile                        # Multi-stage security container build
├── docker-compose.yml                # Agent orchestration and volume mappings
├── entrypoint.sh                     # Container entrypoint script
├── env.sh                            # Host environment activation script
├── hermes                            # Host execution wrapper
├── bin/
│   └── hermes                        # Binary execution shortcut
├── knowledge/                        # Curated offensive security knowledge base
│   ├── Claude-BugHunter/             # Claude BugHunter methodologies & evaluation
│   ├── PayloadsAllTheThings/         # Attack payloads & bypass cheat sheets
│   ├── hack-skills/                  # Specialized offensive skills
│   ├── hacktricks/                   # Extensive penetration testing wiki
│   └── strix-skills/                 # Multi-agent coordination and tooling skills
├── mock_vulnerable_app.py            # Local vulnerable test harness for validation
├── recon/                            # Reconnaissance output directory (.gitkeep)
├── reports/                          # Generated vulnerability reports and PoCs
│   ├── idor_finding.md               # Example validated finding
│   └── poc_idor.py                   # Example PoC verification script
├── scope.yaml                        # Assessment scope and rule-of-engagement definition
├── skills/                           # 50+ Hermes bug bounty skill modules
├── targets/                          # Target queue and scope assets (.gitkeep)
├── templates/
│   └── report_template.md            # Standardized CVSS v3.1 vulnerability report template
└── tools/                            # Security toolchain
    ├── bin/                          # Binary tool wrappers and execution scripts
    ├── sqlmap/                       # SQL injection testing engine
    ├── strix/                        # Penetration testing framework
    └── wordlists/                    # Curated parameter and directory wordlists
```

---

## 🚀 Quick Start

### 1. Environment Setup

Copy `.env.example` to `.env` and fill in your LLM provider and integration credentials:

```bash
cp .env.example .env
```

### 2. Docker Setup (Recommended)

Build and launch the Cybermes container with all dependencies:

```bash
# Build the Docker image
docker compose build

# Run interactive agent session
docker compose run --rm hermes-cybermes
```

### 3. Local Host Setup

Activate the Cybermes environment on your host:

```bash
source env.sh
./hermes --help
```

---

## 🎯 Scope & Rules of Engagement

Cybermes strictly enforces scope validation defined in `scope.yaml`:

```yaml
program: "Authorized Security Assessment"
authorization: "AUTHORIZED"

targets:
  - "http://127.0.0.1:8888"
  - "http://localhost:8888"

allowed:
  reconnaissance: true
  endpoint_discovery: true
  parameter_fuzzing: true
  vulnerability_testing: true
  authenticated_testing: true
  exploit_validation: true
  poc_generation: true
  browser_automation: true

restricted:
  destructive_actions: true
  denial_of_service: true
  data_destruction: true
```

---

## 🧪 Testing with Mock Target

Start the mock target application:

```bash
python3 mock_vulnerable_app.py
```

Then run Cybermes to execute automated discovery and validation against `http://127.0.0.1:8888`.

---

## 🔒 Security & Safe Usage

- Cybermes is designed exclusively for **authorized security testing**, **bug bounty engagements**, and **educational research**.
- Never perform unauthorized testing against systems without explicit, documented permission.
- Sensitive environment variables, auth tokens, session state, and credentials are protected and excluded via `.gitignore`.

---

## 👥 Contributors

- **[Zyrexnn](https://github.com/Zyrexnn)** — Lead Author & Architect
- **[Claude Opus 5 (1M context)](https://anthropic.com)** — AI Co-Author & Security Architecture Research

---

## 📄 License

This repository is distributed under the terms defined within individual subcomponents and tools. See respective component licenses for details.

