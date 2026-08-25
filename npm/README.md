# 🛡️ @zyrexnn/cybermes-mcp

> **Zero-Go Model Context Protocol (MCP) Server for the Cybermes Autonomous Security & Diagnostic Framework**

Connect your AI assistants (**OpenCode**, **Kilo Code**, **Claude Desktop**, **Cursor IDE**, **Windsurf**, **Cline**, **Continue.dev**, **Zed**) directly to **Cybermes** native security intelligence without needing to install or compile Go. Compatible with all LLM backends (Claude, GPT-4o, DeepSeek, Gemini, Llama).

---

## ⚡ Instant Quick Start (Zero Installation)

Run directly via `npx`:

```bash
npx -y @zyrexnn/cybermes-mcp
```

The wrapper automatically detects your OS/CPU architecture (Windows, macOS ARM/Intel, Linux), downloads the matching high-speed Go binary from GitHub Releases, and caches it locally in `~/.cybermes/bin/`.

---

## 🔌 1-Click Client Integration

### 1. Claude Desktop
Add this to your `claude_desktop_config.json`:
- **macOS**: `~/Library/Application Support/Claude/claude_desktop_config.json`
- **Windows**: `%APPDATA%\Claude\claude_desktop_config.json`

```json
{
  "mcpServers": {
    "cybermes": {
      "command": "npx",
      "args": ["-y", "@zyrexnn/cybermes-mcp"]
    }
  }
}
```

### 2. Cursor IDE
Add to your project's `.cursor/mcp.json` or Global MCP settings:

```json
{
  "mcpServers": {
    "cybermes": {
      "command": "npx",
      "args": ["-y", "@zyrexnn/cybermes-mcp"]
    }
  }
}
```

### 3. Windsurf IDE
Add to `mcp_config.json`:

```json
{
  "mcpServers": {
    "cybermes": {
      "command": "npx",
      "args": ["-y", "@zyrexnn/cybermes-mcp"]
    }
  }
}
```

---

## 🧰 Available Security Capabilities (10 Tools + Resources + Prompts)

| Tool Name | Description |
| :--- | :--- |
| `cybermes_validate_scope` | Scope Guard Engine validating targets against `scope.yaml` (wildcards, CIDRs, excludes). |
| `cybermes_http_probe` | HTTP inspection, TLS analysis, and technology detection (Next.js, Laravel, Spring Boot, etc.). |
| `cybermes_recon_crawl` | Endpoint & API crawler with Smart Pipe token budgeting (top 25 high-signal routes). |
| `cybermes_search_knowledge` | Sub-50ms search across PayloadsAllTheThings, HackTricks, & Strix. |
| `cybermes_list_skills` | Catalog and filter 200+ offensive security playbooks. |
| `cybermes_get_skill` | Read complete offensive security playbook SOPs. |
| `cybermes_scan_secrets` | 48-pattern credential leak detector with automated masking. |
| `cybermes_record_finding` | Record verified, reproducible findings to `reports/<target>/findings/`. |
| `cybermes_aggregate_report` | Executive summary generator (`SUMMARY.md` & `metadata.json`). |
| `cybermes_list_findings` | List confirmed findings per target. |

---

## ⚖️ License

Apache-2.0 © [Zyrexnn](https://github.com/Zyrexnn/Cybermes)
