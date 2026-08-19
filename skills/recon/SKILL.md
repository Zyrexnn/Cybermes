---
name: recon
description: Orchestrates scoped subdomain enumeration, DNS resolution, and active web probing for bug bounty targets.
---

# Recon Skill — Subdomain & Asset Discovery

## Purpose
Perform initial reconnaissance and asset mapping strictly within the authorized target scope.

## Workflow
1. **Target Setup & Subdomain Discovery**:
   Use `subfinder` to discover subdomains for the requested target:
   ```bash
   mkdir -p /workspace/recon/<target>
   subfinder -d <target> -silent -o /workspace/recon/<target>/subfinder.txt
   ```
2. **Asset Consolidation**:
   Include the main target and discovered subdomains:
   ```bash
   echo "<target>" >> /workspace/recon/<target>/subfinder.txt
   sort -u /workspace/recon/<target>/subfinder.txt -o /workspace/recon/<target>/in_scope.txt
   ```
4. **DNS Resolution**:
   Use `dnsx` to resolve A, CNAME, and AAAA records:
   ```bash
   dnsx -l /workspace/recon/<target>/in_scope.txt -silent -o /workspace/recon/<target>/resolved.txt
   ```
5. **HTTP Probing**:
   Use `httpx` to probe live web servers, extract status codes, page titles, and web tech stack:
   ```bash
   httpx -l /workspace/recon/<target>/resolved.txt -silent -title -status-code -tech-detect -json -o /workspace/recon/<target>/httpx.json
   ```

## Output Artifacts
- `/workspace/recon/<target>/subfinder.txt` - Raw discovered subdomains
- `/workspace/recon/<target>/in_scope.txt` - Validated in-scope hosts
- `/workspace/recon/<target>/resolved.txt` - Live DNS resolving hosts
- `/workspace/recon/<target>/httpx.json` - HTTP probe results with web technologies

## Safety Guidelines
- Never scan wildcards outside the allowed domain list.
- Keep rate limits conservative (default: 10 req/s).
