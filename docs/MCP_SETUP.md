# 🔌 Cybermes MCP Server — Universal AI Client Setup & Integration Guide

The **Cybermes MCP Server** implements the official **Model Context Protocol (MCP) JSON-RPC 2.0** standard over `stdio` transport. It is **100% Provider-Agnostic and Model-Agnostic** — compatible with any AI foundation model (**Claude 3.7/3.5, GPT-4o/o3, DeepSeek R1/V3, Gemini 2.0/1.5 Pro, Llama 3.3, Qwen 2.5 Coder**) and any MCP-enabled client (**Google Antigravity / Gemini CLI & IDE, Cursor, Kilo Code, Claude Desktop, Windsurf, OpenCode, Cline, Roo Code, Claude Code CLI, Continue.dev, Zed, Hermes, Codex**, etc.).

---

## ⚡ Quick Decision Matrix: Which Installation Method to Use?

| Method | Best For | Prerequisites | Command |
| :--- | :--- | :--- | :--- |
| **Method 0 (NPX Auto-Wizard)** | Fastest setup, zero manual editing | Node.js (v18+) | `npx -y cybermes-mcp install` |
| **Method 1 (Global Binary)** | Zero-latency, instant startup, offline | Node.js (v18+) | `npm i -g cybermes-mcp`<br>`cybermes-mcp install --global` |
| **Method 2 (Targeted Flags)** | Configuring 1 or 2 specific IDEs only | Node.js (v18+) | `npx cybermes-mcp install --gemini --cursor` |
| **Method 3 (Native Go)** | Air-gapped / Enterprise / Core Devs | Go compiler (v1.23+) | `go build -o tools/bin/cybermes-mcp ./cmd/cybermes-mcp` |

---

## 🚀 Method 0: Universal 1-Click Auto-Installer (Recommended)

Cybermes includes an intelligent, non-destructive **Universal Auto-Injector** that automatically detects your installed AI clients across **Windows, macOS, and Linux**, creates timestamped backups of your configurations (`.bak`), and cleanly wires Cybermes into your selected clients.

### 1. Interactive 2-Step Terminal Setup Wizard
Run the interactive wizard in your terminal:
```bash
npx -y cybermes-mcp install
```
- **Step 1 (Execution Mode)**: Choose between **Global Executable** (`cybermes-mcp`) or **NPX On-Demand** (`npx -y cybermes-mcp`).
- **Step 2 (Provider Selection)**: Multi-select checklist with arrow keys (`↑/↓`), spacebar (`Space`), select all (`a`), and confirm (`Enter`).

---

### 2. Targeted Provider Installation (Selective Flags)

You can target specific AI IDEs directly without going through the wizard:

```bash
# Configure ONLY Google Antigravity / Gemini:
npx -y cybermes-mcp install --gemini

# Configure ONLY Kilo Code IDE:
npx -y cybermes-mcp install --kilo

# Configure Antigravity / Gemini and Cursor together:
npx -y cybermes-mcp install --gemini --cursor

# Configure Claude Desktop and OpenCode:
npx -y cybermes-mcp install --claude --opencode
```

---

### 3. Global Installation (Zero NPX Overhead & Offline Ready)

Installing globally gives you instant startup without waiting for remote package checks:

```bash
# 1. Install package globally:
npm install -g cybermes-mcp

# 2. Selectively configure clients with global binary:
cybermes-mcp install --gemini --kilo --global

# 3. Or configure ALL detected clients with global binary:
cybermes-mcp install --all --global
```

---

## 🧰 Universal CLI Command Reference

The `cybermes-mcp` CLI contains built-in diagnostic and management utilities:

```bash
# View full colored documentation & contextual manuals:
npx -y cybermes-mcp -help
npx -y cybermes-mcp help install

# Perform deep healthcheck, binary audit & live JSON-RPC handshake:
npx -y cybermes-mcp doctor

# View configuration and discovery status across all AI IDEs:
npx -y cybermes-mcp status

# Inspect full capability table of all 14 active MCP tools and permissions:
npx -y cybermes-mcp tools

# Search 200+ offline security playbooks directly in terminal:
npx -y cybermes-mcp skills jwt
npx -y cybermes-mcp skills idor

# Manage persistent settings in ~/.cybermes/config.json:
cybermes-mcp config get workspace
cybermes-mcp config set rateLimit 15
cybermes-mcp config list

# Cleanly uninstall Cybermes configuration from AI clients:
npx -y cybermes-mcp uninstall
```

---

## 💻 Manual Configuration by AI Client

If you prefer to configure your AI clients manually, add the JSON/YAML/TOML snippets below to your respective config file:

### 1. Google Antigravity / Gemini
Config Location: `~/.gemini/config/mcp_config.json`
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

---

### 2. Cursor IDE
Config Location: `~/.cursor/mcp.json` (or workspace `.cursor/mcp.json`)
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

---

### 3. Kilo Code IDE (VS Code Extension / Standalone)
Config Location: `~/.kilo/mcp.json`
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

---

### 4. Claude Desktop
Config Location:
- **macOS**: `~/Library/Application Support/Claude/claude_desktop_config.json`
- **Windows**: `%APPDATA%\Claude\claude_desktop_config.json`
- **Linux**: `~/.config/Claude/claude_desktop_config.json`

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

---

### 5. Claude Code CLI (`claude mcp`)
Run directly in terminal:
```bash
claude mcp add cybermes npx -- -y cybermes-mcp
```
Or add to `~/.claude.json`:
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

---

### 6. Windsurf IDE (Codeium)
Config Location: `~/.codeium/windsurf/mcp_config.json`
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

---

### 7. Cline & Roo Code (VS Code Extensions)
Open extension MCP settings and add:
```json
{
  "mcpServers": {
    "cybermes": {
      "command": "npx",
      "args": ["-y", "cybermes-mcp"],
      "disabled": false,
      "autoApprove": [
        "cybermes_generate_pdf",
        "cybermes_search_knowledge",
        "cybermes_list_skills",
        "cybermes_get_skill",
        "cybermes_scan_secrets",
        "cybermes_validate_scope"
      ]
    }
  }
}
```

---

### 8. OpenCode Interpreter
Config Location: `~/.config/opencode/opencode.json` (or project `opencode.json`)
```json
{
  "mcp_servers": {
    "cybermes": {
      "command": "npx",
      "args": ["-y", "cybermes-mcp"]
    }
  }
}
```

---

### 9. Continue.dev
Config Location: `~/.continue/config.json`
```json
{
  "experimental": {
    "modelContextProtocolServers": [
      {
        "name": "cybermes",
        "transport": {
          "type": "stdio",
          "command": "npx",
          "args": ["-y", "cybermes-mcp"]
        }
      }
    ]
  }
}
```

---

### 10. Zed Editor
Config Location: `~/.config/zed/settings.json`
```json
{
  "context_servers": {
    "cybermes": {
      "command": "npx",
      "args": ["-y", "cybermes-mcp"]
    }
  }
}
```

---

### 11. Hermes Agent
Config Location: `~/.hermes/config.yaml`
```yaml
mcp_servers:
  cybermes:
    command: "npx"
    args: ["-y", "cybermes-mcp"]
```

---

### 12. Codex CLI
Config Location: `~/.codex/config.toml`
```toml
[mcp_servers.cybermes]
command = "npx"
args = ["-y", "cybermes-mcp"]
```

---

## 🛠️ Method 3: Local Go Binary (Enterprise & Air-Gapped)

For offline environments or custom security engine builds:

```bash
# 1. Build native MCP binary:
go build -o tools/bin/cybermes-mcp.exe ./cmd/cybermes-mcp

# 2. Auto-wire clients to local binary:
npx -y cybermes-mcp install --local

# 3. Or specify absolute path in config:
# Windows:
# "command": "C:\\path\\to\\Cybermes\\tools\\bin\\cybermes-mcp.exe"
# Linux/macOS:
# "command": "/path/to/Cybermes/tools/bin/cybermes-mcp"
```

---

## 🛡️ Complete Capabilities Catalog (14 Tools, Resources & Prompts)

### 🛠️ 14 Native MCP Tools
| Tool Name | Engine & Scope | Auto-Approve | Purpose |
| :--- | :--- | :---: | :--- |
| `cybermes_generate_pdf` | PDF Engine / `chromedp` | ✔ Yes | Export pixel-perfect executive PDF security reports and interactive HTML dashboards. |
| `cybermes_aggregate_report` | Aggregator Engine | ○ Ask | Compiles findings into `SUMMARY.md`, `metadata.json`, and `report.html`. |
| `cybermes_validate_scope` | Scope Guard | ✔ Yes | Pre-flight validation against `scope.yaml` (Wildcard, CIDR, and Exclude rules). |
| `cybermes_http_probe` | Native Go / `httpx` | ○ Ask | Web probing, TLS certificate inspection, response header analysis, and tech stack detection. |
| `cybermes_recon_crawl` | Native Go / `katana` | ○ Ask | Smart Pipe token-budgeted crawler & JS bundle endpoint miner. |
| `cybermes_subdomain_discovery`| Subfinder / crt.sh | ✔ Yes | Dual-engine subdomain discovery with certificate transparency stream deduplication. |
| `cybermes_fuzz_endpoints` | ffuf / Worker Pool | ✔ Yes | Rate-limited directory and parameter discovery (20-25 req/s) with common.txt fallback. |
| `cybermes_search_knowledge` | BM25 Engine | ✔ Yes | Sub-50ms query against 50,000+ curated offensive security payloads. |
| `cybermes_list_skills` | Metadata Index | ✔ Yes | Catalog and filter 200+ offensive security playbooks and methodology SOPs. |
| `cybermes_get_skill` | Markdown Parser | ✔ Yes | Retrieve complete offensive playbooks (e.g. `hunt-idor`, `hunt-llm-ai`, `jwt-oauth`). |
| `cybermes_scan_secrets` | 48 Regex Patterns | ✔ Yes | High-precision credential & token leak scanner with automated value masking. |
| `cybermes_filter_stream` | Smart Pipe Scoring | ✔ Yes | Entropy & high-signal stream filtering to conserve LLM context window tokens. |
| `cybermes_record_finding` | Workspace Manager | ○ Ask | Create structured finding reports in `reports/<slug>/findings/<sev>_<vuln>.md`. |
| `cybermes_list_findings` | Aggregator Engine | ○ Ask | View all confirmed findings and severity matrix across active target workspaces. |

---

### 📂 MCP Resources & URI Schemes
| Resource URI | Type | Description |
| :--- | :---: | :--- |
| `skills://index` | Static | Catalog index of all 200+ security playbooks. |
| `skills://{skill_name}` | Dynamic | Full markdown SOP for any skill (e.g. `skills://hunt-idor`). |
| `reports://index` | Static | Overview matrix of all active engagement workspaces and severity statistics. |
| `reports://{target_slug}/summary` | Dynamic | Real-time `SUMMARY.md` engagement report for a specific target. |
| `knowledge://cheatsheets` | Static | Curated index of payload categories and methodologies. |

---

## 🧪 Post-Installation Verification

1. Restart your AI IDE / Client.
2. In your chat window, enter this test prompt:
   > *"Check Cybermes system health and search knowledge base for JWT authentication bypasses."*
3. Your AI model will execute `cybermes_search_knowledge` and return structured security intelligence instantly.
