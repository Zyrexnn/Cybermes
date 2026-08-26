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

### 1. Build Local Binary
```bash
go build -o tools/bin/cybermes-mcp.exe ./cmd/cybermes-mcp
```

### 2. Configure Client with Absolute Path (Windows Example)
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

### 3. Configure Client (Linux/macOS Example)
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

## 🧰 Available Capabilities Summary (10 Native Tools + 2 Resources + 2 Prompts)

| Type | Name | Purpose |
| :--- | :--- | :--- |
| **Tool** | `cybermes_validate_scope` | Scope Guard: Target evaluation against `scope.yaml` (Wildcard, CIDR, Exclude rules). |
| **Tool** | `cybermes_http_probe` | HTTP inspection, TLS certificate extraction, and framework fingerprinting. |
| **Tool** | `cybermes_recon_crawl` | Endpoint discovery & JS bundle mining with Smart Pipe token budgeting. |
| **Tool** | `cybermes_search_knowledge`| Sub-50ms query against 50,000+ curated payloads (PayloadsAllTheThings, HackTricks, Strix). |
| **Tool** | `cybermes_list_skills` | Catalog and filter 200+ offensive security playbooks. |
| **Tool** | `cybermes_get_skill` | Read complete offensive playbook SOPs or specific section headings. |
| **Tool** | `cybermes_scan_secrets` | 48-pattern credential leak detector with automated masking. |
| **Tool** | `cybermes_nuclei_scan` | Targeted vulnerability verification using community and custom templates. |
| **Tool** | `cybermes_check_environment`| Diagnostics for local toolchain and security binary availability. |
| **Tool** | `cybermes_record_evidence` | Record raw recon evidence, HTTP logs, and negative test results. |
| **Resource** | `skills://{skill_name}` | Direct read-only URI access to offensive playbook SOPs. |
| **Resource** | `reports://{target_slug}/summary` | Direct read-only URI access to executive engagement summaries. |
| **Prompt** | `security_recon_audit` | Standardized reconnaissance & attack-surface mapping prompt workflow. |
| **Prompt** | `api_security_review` | Deterministic REST, GraphQL, and gRPC vulnerability review workflow. |

---

## 🧪 Verification & Health Check

After saving the configuration and restarting your AI client:
1. In your AI client chat, ask:
   > *"List all available Cybermes MCP tools and search the knowledge base for 'JWT algorithm confusion'."*
2. The AI model will execute `cybermes_search_knowledge` and return validated payload snippets with zero latency.
