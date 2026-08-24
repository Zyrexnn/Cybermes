# Cybermes Security Tools & Utilities

This directory contains security tools, wordlists, and integration utilities used by the Cybermes Offensive Security & Bug Bounty Agent.

## Directory Structure

- `bin/`: Downloaded & compiled binary tools (`smart_pipe`, `secret_scan`, `search_knowledge`, `aggregate_reports`, `nmap`, `subfinder`, `httpx`, `nuclei`, `katana`, etc.).
- `wordlists/`: Curated directory, parameter, and endpoint discovery wordlists (`raft-medium-directories.txt`, `burp-parameter-names.txt`, `common.txt`, `bypass-headers.txt`).
- `doctor.py`: Universal cross-platform health check and auto-repair utility (`python tools/doctor.py --fix`).
- `generate_pdf.py`: Executive PDF & HTML dashboard report generator (Playwright / Markdown).
- `update_tools.sh`: Automated toolchain updater for Linux/macOS (x86_64 & ARM64).
- `update_tools.ps1`: Automated toolchain updater for Windows PowerShell (x86_64 & ARM64).
- `validate_skills.py`: Skill pack integrity and `SKILL.md` completeness auditor.
- `windows_compat_check.py`: Backward-compatible wrapper calling `doctor.py`.

## High-Performance Go Core Tools (`cmd/` & `pkg/`)

The native Go pipeline tools are maintained in the root `pkg/` and `cmd/` directories:
- `cmd/smart_pipe`: High-throughput recon stream filter & token economy engine.
- `cmd/secret_scan`: 48-pattern credential and token miner.
- `cmd/search_knowledge`: Sub-50ms offline knowledge & payload search engine.
- `cmd/aggregate_reports`: Finding aggregator and `SUMMARY.md` / `metadata.json` indexer.

To compile all Go tools to `tools/bin/`:
```bash
go build -ldflags="-s -w" -o tools/bin/smart_pipe ./cmd/smart_pipe
go build -ldflags="-s -w" -o tools/bin/secret_scan ./cmd/secret_scan
go build -ldflags="-s -w" -o tools/bin/search_knowledge ./cmd/search_knowledge
go build -ldflags="-s -w" -o tools/bin/aggregate_reports ./cmd/aggregate_reports
```

## Toolchain Auto-Downloader

To automatically download the latest ProjectDiscovery toolchain (`subfinder`, `httpx`, `katana`, `nuclei`) and update Nuclei templates:

* **Linux / macOS**:
  ```bash
  ./tools/update_tools.sh
  ```
* **Windows**:
  ```powershell
  powershell -ExecutionPolicy Bypass -File tools\update_tools.ps1
  ```
