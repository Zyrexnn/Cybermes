---
name: fuzzing-and-content-discovery
description: Discovers hidden directories, backup files, configuration endpoints, and virtual hosts using high-speed ffuf with intelligent filtering.
---

# Fuzzing & Content Discovery Skill

## Purpose
Perform high-speed, rate-controlled fuzzing to uncover unindexed web routes, exposed backup archives, git repositories, configuration files (`.env`), and admin portals using `ffuf`.

## Inputs
- Live targets: `/workspace/recon/<target>/httpx_live_*.txt` or resolved subdomains.
- Wordlists: `/workspace/wordlists/sensitive-files.txt`, `/workspace/wordlists/raft-medium-directories.txt`, `/workspace/wordlists/common.txt`.

## Workflow

### 1. High-Impact Sensitive Files Discovery (Fast Triage)
Test for critical misconfigurations and exposed environments:
```bash
ffuf -u "https://FUZZ.target.com" \
     -w /workspace/wordlists/sensitive-files.txt \
     -mc 200,204,301,302,307,401,403,500 \
     -rate 30 \
     -t 10 \
     -timeout 5 \
     -s \
     -o /workspace/output/fuzzing/<target>_sensitive.json -of json
```

### 2. Recursive Directory Fuzzing
Fuzz primary application directories with dynamic size/word filtering:
```bash
ffuf -u "https://target.com/FUZZ" \
     -w /workspace/wordlists/raft-medium-directories.txt \
     -mc 200,204,301,302,307,403 \
     -ac \
     -rate 50 \
     -t 20 \
     -o /workspace/output/fuzzing/<target>_dirs.json -of json
```
*(Note: `-ac` enables smart auto-calibration to discard custom 404 pages).*

### 3. Virtual Host (VHost) Fuzzing
Discover internal admin portals mapped to the same IP:
```bash
ffuf -u "https://target.com" \
     -H "Host: FUZZ.target.com" \
     -w /workspace/wordlists/common.txt \
     -ac \
     -rate 30 \
     -o /workspace/output/fuzzing/<target>_vhosts.json -of json
```

## Output Artifacts
- `/workspace/output/fuzzing/<target>_sensitive.json` - High-priority exposed configuration/backup findings.
- `/workspace/output/fuzzing/<target>_dirs.json` - Discovered directories and admin panels.
- `/workspace/output/fuzzing/<target>_vhosts.json` - Valid virtual hosts.

## Safety & Scope Guidelines
- Always stay within rate limits (default `-rate 30` to `-rate 50`) to avoid DoS on target servers.
- Filter out out-of-scope targets and avoid fuzzing destructive endpoints (e.g. `/delete`, `/logout`, `/reset`).
