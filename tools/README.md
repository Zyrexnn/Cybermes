# Cybermes Security Tools

This directory contains security tools, wordlists, and integration utilities used by the Cybermes Offensive Security & Bug Bounty Agent.

## Directory Structure

- `bin/`: Pre-compiled binary tools and wrappers (`smart_pipe`, `secret_scan`, `search_knowledge`, `aggregate_reports`, `nmap`, `subfinder`, `httpx`, `nuclei`, `katana`, `ffuf`, `dalfox`, `amass`, `gau`, etc.).
- `sqlmap/`: Integrated SQL injection testing engine.
- `strix/`: Autonomous penetration testing and multi-agent coordination framework.
- `wordlists/`: Curated directory, parameter, and endpoint discovery wordlists (`raft-medium-directories.txt`, `burp-parameter-names.txt`, `common.txt`).

## High-Performance Go Core Tools (`cmd/` & `pkg/`)

The native Go pipeline tools are maintained in the root `pkg/` and `cmd/` directories:
- `cmd/smart_pipe`: High-throughput recon stream filter & token economy engine.
- `cmd/secret_scan`: 48-pattern credential and token miner.
- `cmd/search_knowledge`: Sub-50ms offline knowledge & payload search engine.
- `cmd/aggregate_reports`: Finding aggregator and `SUMMARY.md` / `metadata.json` indexer.

To recompile all Go tools to `tools/bin/`:
```bash
go build -ldflags="-s -w" -o tools/bin/smart_pipe ./cmd/smart_pipe
go build -ldflags="-s -w" -o tools/bin/secret_scan ./cmd/secret_scan
go build -ldflags="-s -w" -o tools/bin/search_knowledge ./cmd/search_knowledge
go build -ldflags="-s -w" -o tools/bin/aggregate_reports ./cmd/aggregate_reports
```

## Utility & Maintenance Scripts

- `generate_pdf.py`: Executive PDF & HTML dashboard report generator (Playwright / Markdown).
- `update_tools.sh`: Automated toolchain updater for Linux/macOS.
- `update_tools.ps1`: Automated toolchain updater for Windows PowerShell.
- `validate_skills.py`: Skill pack integrity and `SKILL.md` completeness auditor.
- `windows_compat_check.py`: Windows environment and dependency diagnostic tool.

## Pre-compiled Binaries in Docker & Host

In containerized deployments (`Dockerfile`), Go tools and external security tool binaries are compiled and placed into `/usr/local/bin` during build. For host development, run `./setup.sh` or `setup_windows.ps1` to compile tools into `tools/bin/` and source `env.sh` (or `env.ps1`).
