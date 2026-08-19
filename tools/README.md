# Cybermes Security Tools

This directory contains security tools, wordlists, and integration utilities used by the Cybermes Offensive Security & Bug Bounty Agent.

## Directory Structure

- `bin/`: Pre-compiled binary tools and wrappers (`nmap`, `subfinder`, `httpx`, `nuclei`, `katana`, `ffuf`, `dalfox`, `amass`, `gau`, etc.).
- `sqlmap/`: Integrated SQL injection testing engine.
- `strix/`: Autonomous penetration testing and multi-agent coordination framework.
- `wordlists/`: Curated directory, parameter, and endpoint discovery wordlists (`raft-medium-directories.txt`, `burp-parameter-names.txt`, `common.txt`).

## Pre-compiled Binaries in Docker

In containerized deployments (`Dockerfile`), security tool binaries are automatically packaged or pulled into `/usr/local/bin` during build. For host development, place your platform-specific binaries into `tools/bin/` and source `env.sh`.
