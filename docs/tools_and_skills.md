# 🧰 Tools, Skills & MCP Architecture Reference

Cybermes combines a curated offensive security toolchain, deep skill playbooks, and Model Context Protocol (MCP) servers to automate end-to-end vulnerability research and verification.

---

## 🛠️ 1. Integrated Security Toolchain

All native binary security tools reside in `/workspace/tools/bin` and are automatically registered on the container `$PATH`:

| Category | Binary Tool | Core Purpose |
| :--- | :--- | :--- |
| **Reconnaissance & DNS** | `subfinder` | Fast passive subdomain enumeration |
| | `amass` | In-depth network mapping & asset OSINT |
| | `assetfinder` | Rapid domain and subdomain discovery |
| | `httpx` | Fast multi-purpose HTTP probing & technology detection |
| | `nmap` | Network exploration, port probing, and service fingerprinting |
| **Crawling & Mining** | `katana` | Next-generation web crawler with SPA rendering capabilities |
| | `gau` | Fetch historical URLs from Wayback Machine, AlienVault, CommonCrawl |
| | `waybackurls` | Fetch endpoint archives from web historical indexes |
| **Fuzzing & Discovery** | `ffuf` | High-performance web fuzzer for virtual hosts and paths |
| | `feroxbuster` | Fast, recursive content and directory discovery |
| **Exploitation & Scanning** | `nuclei` | Fast, community-driven template vulnerability scanner |
| | `sqlmap` | Automated SQL injection detection and database takeover |
| | `dalfox` | High-performance parameter analysis and XSS scanner |
| **Utilities & Streaming** | `smart_pipe` | High-throughput recon stream filter & token economy engine |
| | `secret_scan` | 48-pattern credential & sensitive token scanner |
| | `search_knowledge` | Sub-millisecond offline knowledge & payload search engine |
| | `aggregate_reports` | Automated finding aggregator & executive summary indexer |
| | `rg` (ripgrep) | Ultra-fast regex searching through source dumps and JS bundles |

---

## 🔌 2. Model Context Protocol (MCP) Servers

Cybermes natively integrates MCP servers for client-side evaluation and structured system access:

1. **Browser MCP (`@modelcontextprotocol/server-puppeteer`)**:
   * Launches headless Chromium in container isolation.
   * Performs DOM tree inspection, automated button clicks, form submissions, and screenshot PoC generation.
2. **Filesystem MCP (`@modelcontextprotocol/server-filesystem`)**:
   * Provides structured read/write access to `/workspace` targets, logs, and reporting artifacts.

---

## 🎯 3. Offensive Skills (200+ Modules)

Skills reside in `skills/` and are automatically loaded into Hermes runtime context.

### Invoking Specific Playbooks:
Operators can prompt the agent to utilize specific methodologies:
* *"Apply 401-403 bypass techniques against /admin"*
* *"Audit parameters on /api/user following IDOR & BOLA methodology"*
* *"Execute race condition testing on the coupon redemption endpoint"*

---

## 📊 4. Automated Report Aggregator (`aggregate_reports`)

Cybermes includes a built-in report aggregator that automatically parses individual vulnerability reports and generates an executive summary:

```bash
# Aggregate findings for a specific target:
aggregate_reports <TARGET_SLUG>

# Aggregate findings across all tested targets:
aggregate_reports --all
```

Outputs generated in `reports/<TARGET_SLUG>/`:
- `SUMMARY.md`: Consolidated vulnerability matrix with severity badges, CVSS scores, and links.
- `metadata.json`: Machine-readable metadata and counts for programmatic integrations and dashboards.
