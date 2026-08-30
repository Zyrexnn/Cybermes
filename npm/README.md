# cybermes-mcp

> **Zero-Go Modular Model Context Protocol (MCP) Server & Universal 1-Click Auto-Injector for the Cybermes Autonomous Security Framework**

Connect your favorite AI assistants and IDEs (**Google Antigravity / Gemini**, **Kilo Code**, **Cursor**, **Claude Desktop**, **OpenCode**, **Windsurf**, **Cline**, **Roo Code**, **Claude Code CLI**, **Continue.dev**, **Zed**, **Hermes**, **Codex**) directly to **Cybermes** native offensive security intelligence without needing to install Go or compile binaries.

Compatible with all major foundation LLMs: **Claude 3.7/3.5, GPT-4o/o3, DeepSeek R1/V3, Gemini 2.0/1.5 Pro, Llama 3.3, and Qwen 2.5 Coder**.

---

## ⚡ Quick Start & Universal CLI Commands

The `cybermes-mcp` package operates 100% standalone with **zero external dependencies** using pure native Node.js:

```bash
# 1. Start Interactive Setup Wizard (select AI clients via terminal checklist):
npx -y cybermes-mcp install

# 2. Display Help & All Commands:
npx -y cybermes-mcp -help

# 3. Perform Deep Diagnostic & Live JSON-RPC MCP Handshake:
npx -y cybermes-mcp doctor

# 4. View Active Client Discovery & Status Matrix:
npx -y cybermes-mcp status

# 5. List all 14+ Registered MCP Tools & Auto-Approve Permissions:
npx -y cybermes-mcp tools

# 6. Search 200+ Offline Security Playbooks & SOPs:
npx -y cybermes-mcp skills jwt

# 7. Manage Persistent Settings (~/.cybermes/config.json):
cybermes-mcp config set workspace "C:\MyRecon"
```

---

## 🧰 Command & Flag Reference

| Command / Flag | Syntax Example | Function & Output |
| :--- | :--- | :--- |
| **`-help` / `--help` / `help`** | `cybermes-mcp -help`<br>`cybermes-mcp help install` | Displays complete CLI documentation or contextual help for subcommands. |
| **`(no command)`** | `npx -y cybermes-mcp` | Spawns the native Cybermes MCP server over `stdio` (JSON-RPC 2.0). |
| **`install`** | `cybermes-mcp install` | **Interactive Checklist Wizard** (or target specific IDEs with flags). |
| **`doctor`** | `cybermes-mcp doctor` | **Deep Diagnostic**: Checks OS/Node runtime, verifies native binary, and tests live JSON-RPC handshake. |
| **`status`** | `cybermes-mcp status` | Visual matrix of detected and configured AI clients. |
| **`tools`** | `cybermes-mcp tools` | Full capability table of all 14+ active tools (including `cybermes_generate_pdf`). |
| **`skills [query]`** | `cybermes-mcp skills jwt` | Search and preview offensive methodology SOPs directly in terminal. |
| **`config`** | `cybermes-mcp config list`<br>`cybermes-mcp config set rateLimit 15` | View and modify global settings in `~/.cybermes/config.json`. |
| **`uninstall`** | `cybermes-mcp uninstall` | Cleanly removes Cybermes configurations from target clients. |

---

## 🎯 Target Provider Flags

You can install into specific AI IDEs directly using flags:

```bash
# Install ONLY into Kilo Code IDE:
npx -y cybermes-mcp install --kilo

# Install into Antigravity / Gemini and Cursor:
npx -y cybermes-mcp install --gemini --cursor

# Install globally without npx overhead:
npm install -g cybermes-mcp
cybermes-mcp install --global
```

---

## 🛡️ Available Security Capabilities (14 Tools, 2 Resources, 2 Prompts)

| Capability | Category | Auto-Approve | Purpose |
| :--- | :--- | :--- | :--- |
| `cybermes_generate_pdf` | Reporting | ✔ Enabled | Native PDF security report export via headless Chromium (`chromedp`). |
| `cybermes_aggregate_report` | Reporting | ○ Ask User | Compiles `SUMMARY.md`, `metadata.json`, and interactive `report.html`. |
| `cybermes_validate_scope` | Safety | ✔ Enabled | Validates target domains/IPs against wildcard & CIDR rules in `scope.yaml`. |
| `cybermes_http_probe` | Recon | ○ Ask User | Web tech detection, TLS certificate extraction & response header analysis. |
| `cybermes_recon_crawl` | Recon | ○ Ask User | Smart Pipe token-budgeted crawler & JS bundle miner. |
| `cybermes_subdomain_discovery` | Recon | ✔ Enabled | Passive subdomain discovery with stream deduplication. |
| `cybermes_search_knowledge` | Intelligence | ✔ Enabled | Sub-50ms query against 50,000+ curated security payloads. |
| `cybermes_list_skills` | Intelligence | ✔ Enabled | Catalog and filter 200+ offensive security playbooks. |
| `cybermes_get_skill` | Intelligence | ✔ Enabled | Load step-by-step offensive methodology SOP into LLM memory. |
| `cybermes_scan_secrets` | Secret Mining | ✔ Enabled | 48-pattern credential leak detector with automated masking. |
| `cybermes_fuzz_endpoints` | Active Audit | ✔ Enabled | Directory & parameter mutation fuzzer with Smart Pipe rate-limiting. |
| `cybermes_filter_stream` | Token Saver | ✔ Enabled | Stream output filter minimizing noisy context token usage. |
| `cybermes_record_finding` | Reporting | ○ Ask User | Records validated findings and generates reproducible PoC scripts. |
| `cybermes_list_findings` | Reporting | ○ Ask User | Lists confirmed findings and severity breakdown per target. |

---

## 📄 License

Apache-2.0 (c) Zyrexnn