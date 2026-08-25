# 🗺️ Cybermes MCP Server — Roadmap & Architecture Status

This document records the completed milestones, active architectural capabilities, and future engineering roadmap for the **Cybermes Model Context Protocol (MCP) Server**.

---

## 📌 Roadmap Status Overview

```text
[✅ PHASE 1] Core MCP Server & Native Tools (COMPLETED)
     └── [✅ PHASE 2] Active Recon & Scope Safety Engine (COMPLETED)
          └── [✅ PHASE 3] NPM Wrapper (`npx`) & Multi-Platform CI/CD (COMPLETED)
               └── [🔮 PHASE 4] Optional GPU/CUDA Acceleration & Local AI Triage (FUTURE BACKLOG)
```

---

## 📋 Completed Milestones (Phases 1, 2, & 3)

### 1. Foundation & Communication Protocol (Phase 1)
- **Core Engine**: Built using `github.com/mark3labs/mcp-go` v0.58.0 over `stdio` transport (JSON-RPC 2.0).
- **Workspace Auto-Discovery**: Automatic detection of project root directory by traversing upwards for `AGENTS.md`.
- **Logging Hygiene**: 100% of diagnostic logs are strictly routed to `stderr`, keeping `stdout` completely clean for JSON-RPC frame serialization.

### 2. Active Native Tools & Resources (Phases 1 & 2)

| Category | Name | Status & Capabilities |
| :--- | :--- | :--- |
| **Tool** | `cybermes_validate_scope` | Validates target domain/URL/IP against `scope.yaml` (Wildcard `*.target.com`, CIDR `192.168.1.0/24`, path exclusions). |
| **Tool** | `cybermes_http_probe` | HTTP inspection, TLS cert extraction, and framework fingerprinting (Next.js, Laravel, Spring Boot, WordPress, etc.). Dual engine (`httpx` + Go `net/http`). |
| **Tool** | `cybermes_recon_crawl` | Endpoint discovery & JS bundle mining with Smart Pipe token budgeting (top 25 high-signal routes in context) + raw dump preservation. |
| **Tool** | `cybermes_search_knowledge` | Sub-50ms search across 50,000+ curated payloads (PayloadsAllTheThings, HackTricks, Strix, Claude-BugHunter). |
| **Tool** | `cybermes_list_skills` | Catalog and filter 200+ offensive security playbooks. |
| **Tool** | `cybermes_get_skill` | Retrieve complete Markdown SOP playbooks or specific section headings. |
| **Tool** | `cybermes_scan_secrets` | 48-pattern credential leak detector (AWS, GCP, GitHub, Slack, Private Keys, JWTs) with automated masking. |
| **Tool** | `cybermes_record_finding` | Record verified findings to `reports/<target>/findings/` matching strict `AGENTS.md` format with standalone PoC scripts. |
| **Tool** | `cybermes_aggregate_report` | Executive summary generator aggregating metrics into `SUMMARY.md` and `metadata.json`. |
| **Tool** | `cybermes_list_findings` | List confirmed findings and severity breakdown per target. |
| **Resource** | `skills://{skill_name}` | Direct read-only URI access to playbook SOPs. |
| **Resource** | `reports://{target_slug}/summary` | Direct read-only URI access to executive engagement summaries. |
| **Prompt** | `cybermes_hunt` & `cybermes_triage` | Zero-false-positive reasoning workflow templates for AI agents. |

### 3. Zero-Go NPX Distribution & Multi-Platform CI/CD (Phase 3)
- **NPM Package**: `@zyrexnn/cybermes-mcp` (`npm/package.json` & `npm/bin/cybermes-mcp.js`).
- **1-Click Universal Auto-Injector Engine**: Instant multi-client discovery, non-destructive config injection, and backup rotation via:
  ```bash
  npx -y @zyrexnn/cybermes-mcp install
  ```
- **Local Python Setup Tool**: `scripts/setup_mcp.py` for local and air-gapped developer environments.
- **Automated Multi-Platform CI/CD**: GitHub Actions workflow (`.github/workflows/mcp-release.yml`) cross-compiling Go binaries for Windows AMD64, Linux AMD64, macOS ARM64 (Apple Silicon), and macOS AMD64 (Intel), complete with SHA256 checksums and automated npm publishing.
- **Universal AI Client Integration**: Comprehensive setup guides and auto-injection support for 11+ AI clients: OpenCode, Cursor IDE, Claude Desktop, Windsurf, Cline, Roo Code, Claude Code CLI, Continue.dev, Zed, Kilo Code, Hermes, Codex, and Google Antigravity in [`docs/MCP_SETUP.md`](MCP_SETUP.md).

---

## 🗺️ Future Engineering Roadmap (Phase 4 - Backlog)

### 🔵 Phase 4: GPU Acceleration & Local AI Triage (Optional)
*Focus: Heavy local compute acceleration for operators with dedicated NVIDIA RTX GPUs.*

1. **Local Vector Search / Semantic Indexing**:
   - Small quantized local embedding models (via ONNX/CUDA) to enable semantic vector retrieval over 50,000+ payload KB articles.
2. **Local LLM Pre-Triage**:
   - Local micro-model filtering via Ollama/llama.cpp to pre-score and eliminate benign false positives before forwarding to the primary LLM agent.
