# cybermes-mcp

> **Universal Zero-Go Model Context Protocol (MCP) Server & 1-Click Auto-Installer for the Cybermes Autonomous Security & Diagnostic Framework**

Connect your favorite AI assistants and IDEs (**Google Antigravity / Gemini**, **Kilo Code**, **Cursor**, **Claude Desktop**, **OpenCode**, **Windsurf**, **Cline**, **Roo Code**, **Claude Code CLI**, **Continue.dev**, **Zed**, **Hermes**, **Codex**) directly to **Cybermes** native offensive security intelligence without needing to install Go or compile binaries.

Compatible with all foundation LLMs: **Claude 3.7/3.5, GPT-4o/o3, DeepSeek R1/V3, Gemini 2.0/1.5 Pro, Llama 3.3, and Qwen 2.5 Coder**.

---

## 1-Click Universal Auto-Installer (Recommended)

Install and automatically inject Cybermes MCP into **all detected AI clients** or **specific providers only** (Windows, macOS, Linux):

```bash
# Auto-detect and configure all installed AI clients:
npx -y cybermes-mcp install

# Install ONLY to specific providers:
npx -y cybermes-mcp install --kilo
npx -y cybermes-mcp install --gemini --cursor
npx -y cybermes-mcp install kilo-code
```

### Global Installation (No NPX Overhead)

For zero-latency startup and offline reliability, install the package globally via npm:

```bash
# Install globally:
npm install -g cybermes-mcp

# Wire into all AI clients using global executable:
cybermes-mcp install --global

# Wire into specific provider using global executable:
cybermes-mcp install --kilo --global
cybermes-mcp install --gemini --global
```

### CLI Command & Flag Options:
| Command / Flag | Purpose |
| :--- | :--- |
| `npx -y cybermes-mcp install` | Auto-detect all installed AI IDEs and inject Cybermes configuration with safe `.bak` backups. |
| `cybermes-mcp install --global` | Configure clients to spawn `cybermes-mcp` directly without `npx` overhead. |
| `npx -y cybermes-mcp install --kilo` | Install **only** into **Kilo Code IDE** (`~/.kilo/mcp.json`). |
| `npx -y cybermes-mcp install --gemini` | Install **only** into **Google Antigravity / Gemini** (`~/.gemini/config/mcp_config.json`). |
| `npx -y cybermes-mcp install --cursor` | Install **only** into **Cursor IDE** (`~/.cursor/mcp.json`). |
| `npx -y cybermes-mcp install --claude` | Install **only** into **Claude Desktop**. |
| `npx -y cybermes-mcp install --dry-run` | Preview configuration changes without writing any files to disk. |
| `npx -y cybermes-mcp status` | View discovery and connection status matrix across all supported AI clients. |
| `npx -y cybermes-mcp uninstall --kilo` | Cleanly remove Cybermes configuration from a specific client. |
| `npx -y cybermes-mcp uninstall` | Cleanly remove Cybermes configuration from all AI clients. |

---

## Direct Server Execution (stdio)

When invoked without subcommands, the package executes the Cybermes MCP server over `stdio` (JSON-RPC 2.0):

```bash
npx -y cybermes-mcp
```

---

## Manual AI Client Configuration

If you prefer to configure manually, simply add `cybermes-mcp` to your client's MCP configuration:

### 1. Claude Desktop
Add to `claude_desktop_config.json`:
- **Windows**: `%APPDATA%\Claude\claude_desktop_config.json`
- **macOS**: `~/Library/Application Support/Claude/claude_desktop_config.json`
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

### 2. Cursor IDE
Add to `.cursor/mcp.json` or Global MCP settings:
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

### 3. OpenCode Interpreter
Add to `~/.config/opencode/opencode.json`:
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

### 4. Windsurf IDE (Codeium)
Add to `~/.codeium/windsurf/mcp_config.json`:
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

### 5. Cline & Roo Code (VS Code Extensions)
Add to `cline_mcp_settings.json`:
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

## Available Security Capabilities (10 Tools, 2 Resources, 2 Prompts)

| Capability | Purpose |
| :--- | :--- |
| `cybermes_validate_scope` | **Scope Guard**: Validates targets against `scope.yaml` (Wildcard `*.target.com`, CIDR `192.168.1.0/24`, path exclusions). |
| `cybermes_http_probe` | Web technology detection, TLS certificate extraction, and response header analysis. |
| `cybermes_recon_crawl` | Endpoint discovery and JS bundle mining with Smart Pipe token budgeting (top 25 high-signal routes). |
| `cybermes_search_knowledge` | Sub-50ms query against 50,000+ curated payloads (PayloadsAllTheThings, HackTricks, Strix). |
| `cybermes_list_skills` | Catalog and filter 200+ offensive security playbooks. |
| `cybermes_get_skill` | Read complete offensive security playbook SOPs into model memory. |
| `cybermes_scan_secrets` | 48-pattern credential leak detector with automated masking. |
| `cybermes_record_finding` | Record verified findings to `reports/<target>/findings/` and generate reproducible PoCs. |
| `cybermes_aggregate_report` | Executive summary generator (`SUMMARY.md` & `metadata.json`). |
| `cybermes_list_findings` | List confirmed findings and severity breakdown per target. |
| `skills://{skill_name}` | Direct read-only URI access to offensive playbook SOPs. |
| `reports://{target_slug}/summary` | Direct read-only URI access to executive engagement summaries. |
| `cybermes_hunt` & `cybermes_triage` | Zero-false-positive reasoning workflow templates for AI agents. |

---

## Verification & Health Check

After running `npx -y cybermes-mcp install` and restarting your AI client:
1. In your AI client chat, ask:
   > *"List all available Cybermes MCP tools and search the knowledge base for 'JWT algorithm confusion'."*
2. The AI model will execute `cybermes_search_knowledge` and return validated payload snippets with zero latency.

---

## License

Apache-2.0 (c) Zyrexnn

