# 🔌 Cybermes MCP Server — Universal AI Client Setup & Integration Guide

The **Cybermes MCP Server** implements the official **Model Context Protocol (MCP) JSON-RPC 2.0** standard over `stdio` transport. It is **100% Provider-Agnostic and Model-Agnostic** — compatible with any AI foundation model (**Claude 3.7/3.5, GPT-4o/o3, DeepSeek R1/V3, Gemini 2.0/1.5 Pro, Llama 3.3, Qwen 2.5 Coder**) and any MCP-enabled client (**OpenCode, Kilo, Cursor, Claude Desktop, Windsurf, Cline, Roo Code, Continue.dev, Zed**, etc.).

---

## ⚡ Method 1: Zero-Go NPX (Instant & Recommended)

No Go SDK installation or local compilation required. Node.js (v18+) automatically detects your operating system and CPU architecture, downloads the verified binary from GitHub Releases, caches it in `~/.cybermes/bin/`, and starts the server with zero configuration.

```bash
npx -y @zyrexnn/cybermes-mcp
```

---

### 1. OpenCode Interpreter / OpenCode CLI
Add to your OpenCode configuration (`opencode.json` or `~/.config/opencode/config.json`):

```json
{
  "mcp_servers": {
    "cybermes": {
      "command": "npx",
      "args": ["-y", "@zyrexnn/cybermes-mcp"]
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
      "args": ["-y", "@zyrexnn/cybermes-mcp"]
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
   - **Command**: `npx -y @zyrexnn/cybermes-mcp`

Or add `.cursor/mcp.json` to your project workspace root:
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
      "args": ["-y", "@zyrexnn/cybermes-mcp"]
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
      "args": ["-y", "@zyrexnn/cybermes-mcp"]
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
      "args": ["-y", "@zyrexnn/cybermes-mcp"],
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
          "args": ["-y", "@zyrexnn/cybermes-mcp"]
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
      "args": ["-y", "@zyrexnn/cybermes-mcp"]
    }
  }
}
```

---

## 🛠️ Method 2: Direct Native Go Binary (Offline / Air-Gapped / Enterprise)

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
| **Tool** | `cybermes_record_finding` | Record verified findings to `reports/<target>/findings/` and generate PoCs. |
| **Tool** | `cybermes_aggregate_report` | Executive summary generator (`SUMMARY.md` & `metadata.json`). |
| **Tool** | `cybermes_list_findings` | List confirmed findings and severity breakdown per target. |
| **Resource** | `skills://{skill_name}` | Direct read-only URI access to offensive playbook SOPs. |
| **Resource** | `reports://{target_slug}/summary` | Direct read-only URI access to executive engagement summaries. |
| **Prompt** | `cybermes_hunt` & `cybermes_triage` | Zero-false-positive reasoning workflow templates for AI agents. |

---

## 🧪 Verification & Health Check

After saving the configuration and restarting your AI client:
1. In your AI client chat, ask:
   > *"List all available Cybermes MCP tools and search the knowledge base for 'JWT algorithm confusion'."*
2. The AI model will execute `cybermes_search_knowledge` and return validated payload snippets with zero latency.
