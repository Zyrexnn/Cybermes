---
name: web-enumeration
description: Discovers web application endpoints, Javascript files, hidden paths, and technology fingerprints using httpx and katana.
---

# Web Enumeration Skill

## Purpose
Catalog endpoints, static assets, APIs, and client-side routes across live HTTP services identified during recon.

## Workflow
1. **Input**: Take live hosts from `/workspace/recon/<target>/httpx.json` or `/workspace/recon/<target>/resolved.txt`.
2. **Endpoint Crawling with Katana**:
   ```bash
   katana -list /workspace/recon/<target>/resolved.txt \
          -depth 2 \
          -concurrency 5 \
          -rate-limit 10 \
          -silent \
          -o /workspace/recon/<target>/katana_endpoints.txt
   ```
3. **Javascript & Asset Extraction**:
   Filter discovered URLs for `.js`, `.json`, `.xml`, `.env`, API endpoints:
   ```bash
   grep -Ei '\.js(\?.*)?$' /workspace/recon/<target>/katana_endpoints.txt > /workspace/recon/<target>/js_files.txt || true
   grep -Ei '/(api|v1|v2|v3|graphql|swagger|openapi)/' /workspace/recon/<target>/katana_endpoints.txt > /workspace/recon/<target>/api_candidates.txt || true
   ```
4. **Header & Security Fingerprinting**:
   Extract missing security headers (CORS, CSP, HSTS, X-Frame-Options):
   ```bash
   httpx -l /workspace/recon/<target>/resolved.txt \
         -silent \
         -response-headers-to-store \
         -include-response \
         -o /workspace/recon/<target>/headers.json
   ```

## Output Artifacts
- `/workspace/recon/<target>/katana_endpoints.txt` - All crawled paths & URLs
- `/workspace/recon/<target>/js_files.txt` - Javascript source files for client-side review
- `/workspace/recon/<target>/api_candidates.txt` - Potential API routes

## Safety Guidelines
- Limit crawl depth to avoid infinite crawler loops.
- Exclude destructive actions (logout, delete, checkout) based on scope regex.
