# 🔌 Cybermes MCP Server — Universal AI Client Setup & Integration Guide

The **Cybermes MCP Server** implements the official **Model Context Protocol (MCP) JSON-RPC 2.0** standard over `stdio` transport. It is **100% Provider-Agnostic and Model-Agnostic** — compatible with any AI foundation model (**Claude 3.7/3.5, GPT-4o/o3, DeepSeek R1/V3, Gemini 2.0/1.5 Pro, Llama 3.3, Qwen 2.5 Coder**) and any MCP-enabled client (**OpenCode, Cursor, Claude Desktop, Windsurf, Cline, Roo Code, Claude Code, Continue.dev, Zed, Kilo, Hermes, Codex**, etc.).

---

## 🚀 Method 0: 1-Click Universal Auto-Installer (Recommended)

Cybermes includes an intelligent, non-destructive **Universal Auto-Injector** that automatically detects your installed AI clients across **Windows, macOS, and Linux**, creates timestamped backups of your configurations (`.bak`), and cleanly wires Cybermes into all of them.

### Instant Global Installation (Zero-Go NPX)
```bash
# Auto-detect and configure all installed AI clients
npx -y cybermes-mcp install

# Install ONLY to specific AI providers:
npx -y cybermes-mcp install --kilo
npx -y cybermes-mcp install --gemini --cursor
```

### Global Installation (No NPX Latency)
```bash
# Install globally via npm:
npm install -g cybermes-mcp

# Wire clients to use global 'cybermes-mcp' command:
cybermes-mcp install --global
cybermes-mcp install --gemini --global
```

### Auto-Installer Options & Provider Flags:
| Flag / Option | Description | Example |
| :--- | :--- | :--- |
| `--gemini`, `--antigravity` | Install **only** into **Google Antigravity / Gemini** (`~/.gemini/config/mcp_config.json`) | `npx cybermes-mcp install --gemini` |
| `--kilo`, `--kilo-code` | Install **only** into **Kilo Code IDE** (`~/.kilo/mcp.json`) | `npx cybermes-mcp install --kilo` |
| `--cursor` | Install **only** into **Cursor IDE** (`~/.cursor/mcp.json`) | `npx cybermes-mcp install --cursor` |
| `--claude` | Install **only** into **Claude Desktop** | `npx cybermes-mcp install --claude` |
| `--claude-code` | Install **only** into **Claude Code CLI** (`~/.claude.json`) | `npx cybermes-mcp install --claude-code` |
| `--windsurf` | Install **only** into **Windsurf IDE** (`~/.codeium/...`) | `npx cybermes-mcp install --windsurf` |
| `--cline`, `--roo` | Install **only** into **Cline / Roo Code** | `npx cybermes-mcp install --cline` |
| `--opencode` | Install **only** into **OpenCode Interpreter** | `npx cybermes-mcp install --opencode` |
| `--zed`, `--continue` | Install **only** into **Zed Editor** / **Continue.dev** | `npx cybermes-mcp install --zed` |
| `--hermes`, `--codex` | Install **only** into **Hermes Agent** / **Codex CLI** | `npx cybermes-mcp install --hermes` |
| `--global`, `-g` | Wire clients to invoke global binary directly without `npx` | `cybermes-mcp install --global` |
| `--dry-run` | Preview configuration changes without modifying files | `npx cybermes-mcp install --dry-run` |
| `--status` | Display discovery matrix & health status for all AI clients | `npx cybermes-mcp status` |
| `--uninstall` | Cleanly remove Cybermes from AI client configs | `npx cybermes-mcp uninstall` |
| `--force` | Generate config files even if AI client is not yet detected | `npx cybermes-mcp install --force` |

---

## 🛠️ Method 1: Local Cross-Platform MCP Manager & Optimizer

Cybermes includes a zero-dependency Python management CLI and interactive TUI for real-time latency diagnostics, server toggling, and startup optimization:

```bash
# Windows (CMD / PowerShell)
.\mcp.bat            # or .\mcp.ps1

# Linux / macOS (Bash)
./mcp.sh             # or python3 scripts/mcp.py

# Direct CLI Commands:
python scripts/mcp.py status          # Live stdio JSON-RPC latency probing
python scripts/mcp.py toggle cybermes # Enable / disable server (RAM & token optimizer)
python scripts/mcp.py optimize        # Pre-cache remote npx servers locally
python scripts/mcp.py backup          # Create or restore timestamped snapshots
```

---

## ⚡ Method 2: Manual Zero-Go NPX Setup (Instant)

If you prefer manual configuration, Node.js (v18+) automatically downloads the verified binary from GitHub Releases:

```bash
npx -y cybermes-mcp
```

---

### 1. OpenCode Interpreter / OpenCode CLI
Add to your OpenCode configuration (`opencode.json` or `~/.config/opencode/config.json`):

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

### 2. Kilo Code / Kilo AI (VS Code Extension / Agent)
In **Kilo MCP Settings** or in `.kilo/mcp.json`:

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

### 3. Cursor IDE
1. Open **Cursor Settings** (`Ctrl + Shift + J` / `Cmd + Shift + J`) -> **Features** -> **MCP Servers**.
2. Click **+ Add New MCP Server**.
3. Configure:
   - **Name**: `cybermes`
   - **Type**: `command`
   - **Command**: `npx -y cybermes-mcp`

Or add `.cursor/mcp.json` to your project workspace root:
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
Add to your `claude_desktop_config.json`:
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

### 5. Windsurf IDE (Codeium)
Add to `~/.codeium/windsurf/mcp_config.json` or workspace `mcp_config.json`:

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

### 6. Cline & Roo Code (VS Code Extensions)
Open the MCP Servers extension tab, select **Edit Settings**, and add:

```json
{
  "mcpServers": {
    "cybermes": {
      "command": "npx",
      "args": ["-y", "cybermes-mcp"],
      "disabled": false,
      "autoApprove": [
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

### 7. Continue.dev (VS Code / JetBrains)
Add to `~/.continue/config.json`:

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

### 8. Zed Editor
Add to `~/.config/zed/settings.json`:

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

### 9. Claude Code CLI (`claude mcp`)
Run via CLI or add to `~/.claude.json`:

```bash
claude mcp add cybermes npx -- -y cybermes-mcp
```
Or in `~/.claude.json`:
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

### 10. Hermes Agent
Add to `~/.hermes/config.yaml` or `.hermes/config.yaml`:

```yaml
mcp_servers:
  cybermes:
    command: "npx"
    args: ["-y", "cybermes-mcp"]
```

---

### 11. Codex CLI
Add to `~/.codex/config.toml`:

```toml
[mcp_servers.cybermes]
command = "npx"
args = ["-y", "cybermes-mcp"]
```

---

### 12. Google Antigravity & Gemini Assistants
Add to `~/.gemini/config/mcp_config.json`:

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

Or install with direct flag:
```bash
npx -y cybermes-mcp install --gemini
```

---

## 🛠️ Method 3: Direct Native Go Binary (Offline / Air-Gapped / Enterprise)

If you have cloned the Cybermes repository or want to run the native pre-compiled Go binary directly without Node.js/npx:

### 1. Build All Local Binaries (1-Click)
```bash
# Cross-platform build script (compiles all 5 Go tools to tools/bin/):
go run scripts/build_tools.go

# Or compile only the MCP server:
go build -o tools/bin/cybermes-mcp.exe ./cmd/cybermes-mcp
```

### 2. Auto-Wire Local Binary to All Clients
```bash
# Using Python auto-installer:
python scripts/setup_mcp.py --local

# Using NPX / NPM CLI:
npx -y cybermes-mcp install --local
```

### 3. Manual Configuration with Absolute Path (Windows Example)
```json
{
  "mcpServers": {
    "cybermes": {
      "command": "C:\\path\\to\\Cybermes\\tools\\bin\\cybermes-mcp.exe",
      "args": ["-workspace", "C:\\path\\to\\Cybermes"]
    }
  }
}
```

### 4. Manual Configuration (Linux/macOS Example)
```json
{
  "mcpServers": {
    "cybermes": {
      "command": "/path/to/Cybermes/tools/bin/cybermes-mcp",
      "args": ["-workspace", "/path/to/Cybermes"]
    }
  }
}
```

---

## 🛡️ Windows Defender & Antivirus Exclusion Guide

Because Cybermes contains 50,000+ educational pentest payloads, webshell references (e.g. PHP/ASP one-liners in `knowledge/PayloadsAllTheThings`), and the EICAR test string in text markdown documentation, Windows Defender's real-time memory scanner may flag these text strings during deep indexing or unit tests.

### How to Add Exclusion (PowerShell Administrator):
```powershell
# Run PowerShell as Administrator:
Add-MpPreference -ExclusionPath "C:\path\to\Cybermes"
```

> **Note**: These files are inert Markdown documentation (`.md` / `.txt`) and not executable malware.

---

## 🧰 Available Capabilities Summary (16 Native Tools + 5 Resources + 6 Workflow Prompts)

### 🛠️ 16 MCP Tools
| Tool Name | Engine & Scope | Purpose |
| :--- | :--- | :--- |
| `cybermes_validate_scope` | Scope Guard | Pre-flight validation against `scope.yaml` (Wildcard, CIDR, and Exclude rules). |
| `cybermes_http_probe` | Native Go / httpx | HTTP inspection, TLS info, tech stack detection, custom headers & cookies. |
| `cybermes_recon_crawl` | Native Go / katana | Deep recursive endpoint mining & JS asset discovery with auth session support. |
| `cybermes_subdomain_discovery`| Subfinder / crt.sh | Dual-engine subdomain enumeration with certificate transparency fallback. |
| `cybermes_fuzz_endpoints` | ffuf / Worker Pool | Rate-limited directory and parameter discovery (20-25 req/s) with common.txt fallback. |
| `cybermes_filter_stream` | Smart Pipe Scoring | Entropy & high-signal stream filtering to conserve LLM context window tokens. |
| `cybermes_search_knowledge` | BM25 Engine | Sub-50ms query over 50,000+ offensive payloads (PayloadsAllTheThings, HackTricks, Strix). |
| `cybermes_list_skills` | Metadata Index | Catalog and filter 200+ offensive security playbooks and methodology SOPs. |
| `cybermes_get_skill` | Markdown Parser | Retrieve complete offensive playbooks (e.g. `hunt-idor`, `hunt-llm-ai`, `jwt-oauth`). |
| `cybermes_scan_secrets` | 48 Regex Patterns | High-precision credential & token leak scanner with automated value masking. |
| `cybermes_nuclei_scan` | Nuclei Engine | Deterministic vulnerability verification with community and custom templates. |
| `cybermes_check_environment`| Diagnostics | Real-time toolchain readiness check with on-demand installation guidance. |
| `cybermes_record_finding` | Workspace Manager | Create structured finding reports in `reports/<slug>/findings/<sev>_<vuln>.md`. |
| `cybermes_record_evidence` | Workspace Manager | Append raw observations, logs, and negative test tables to `recon_notes.md`. |
| `cybermes_list_findings` | Aggregator | View all confirmed findings across active target engagement workspaces. |
| `cybermes_aggregate_report` | Aggregator | Compile target findings into executive `SUMMARY.md`, `metadata.json`, `report.html`, and `REPORT.pdf`. |
| `cybermes_generate_pdf` | PDF Engine | Render pixel-perfect executive PDF and interactive HTML dashboard reports via native Chrome DevTools Protocol. |

### 📂 5 Static & Dynamic Resources
| Resource URI | Type | Purpose |
| :--- | :---: | :--- |
| `skills://index` | Static | Browseable catalog of all 200+ offensive playbooks. |
| `reports://index` | Static | Overview matrix of all active engagement workspaces and severity counts. |
| `knowledge://cheatsheets` | Static | Curated index of payload categories and methodologies. |
| `skills://{skill_name}` | Dynamic | Read full playbook SOP for any skill (e.g. `skills://hunt-llm-ai`). |
| `reports://{target_slug}/summary` | Dynamic | Real-time `SUMMARY.md` engagement report for a specific target. |

### 🎯 6 Workflow Prompts
| Prompt Name | Primary Use Case |
| :--- | :--- |
| `cybermes_hunt` | Full-scope autonomous hunting and attack-surface discovery workflow. |
| `cybermes_triage` | Zero-False-Positive verification gate for candidate vulnerabilities. |
| `cybermes_api_audit` | REST, GraphQL, OpenAPI, and token security assessment SOP. |
| `cybermes_idor_matrix` | 4-step dual-account differential testing matrix (User A vs User B). |
| `cybermes_403_bypass` | WAF & 403 Forbidden evasion checklist (headers, path normalization, verbs). |
| `cybermes_ai_prompt_injection_audit`| OWASP LLM01/ASI01-ASI03 prompt injection, RAG, and agentic tool audit SOP. |

For detailed prompting techniques and anti-filter guidance, see [`docs/prompt_guide.md`](prompt_guide.md).

---

## 🧪 Verification & Health Check

After saving the configuration and restarting your AI client:
1. In your AI client chat, ask:
   > *"List all available Cybermes MCP tools and search the knowledge base for 'JWT algorithm confusion'."*
2. The AI model will execute `cybermes_search_knowledge` and return validated payload snippets with zero latency.
