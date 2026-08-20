---
name: autonomous-godmode-hunter
description: Autonomous end-to-end security research and full-surface bug hunting engine. Executes non-stop reconnaissance, attack-surface discovery, parameter mining, multi-stage vulnerability verification, minimal-impact PoC generation, and structured reporting on authorized targets.
---

# Autonomous God-Mode Bug Hunter Skill

## 1. Operational Directive & Scope Rule
1. **Explicit Operator Authorization**: When target domains, URLs, or CIDRs are specified in `scope.yaml` or conversation prompt, proceed with aggressive, exhaustive, and continuous technical validation.
2. **Non-Destructive & Rate Controlled**: Never perform DoS, volume saturation, or destructive data alterations. Always strive for **Maximum Technical Validation with Minimum Unnecessary Impact**.
3. **Zero False Positives**: Never declare a vulnerability confirmed without reproducible HTTP request/response proofs, browser console logs, or execution evidence.

## 2. Autonomous Execution Pipeline

```
                [ Target Defined in scope.yaml ]
                              │
               ┌──────────────┴──────────────┐
               ▼                             ▼
        [ Passive Recon ]             [ Active Discovery ]
     - Subfinder / Assetfinder      - HTTPx Probe & Tech Detect
     - Historical URL Mining (gau)  - Katana SPA / JS Crawling
               │                             │
               └──────────────┬──────────────┘
                              ▼
               [ Attack Surface Ranking & Fuzzing ]
     - Content / Directory Fuzzing (ffuf / feroxbuster)
     - Parameter Mining (burp-parameter-names)
     - Client-side JS Secret & Endpoint Extraction
                              │
                              ▼
               [ Deep Vulnerability Hypothesis ]
     - Auth & Session Handling (Dual-account matrix)
     - IDOR / BOLA / Mass-Assignment Testing
     - Injection Vectors (SQLi, SSTI, Command Injection)
     - Client-side Flaws (DOM XSS, CORS, CSRF via Puppeteer MCP)
     - SSRF & OOB Interaction Testing
                              │
                              ▼
               [ Controlled Exploit Validation ]
     - Replay & Mutate Requests
     - Isolate Root Cause Flaw
     - Generate Minimal-Impact PoC Script (Python)
                              │
                              ▼
               [ Evidence Capture & Reporting ]
     - Raw HTTP Request & Response Proofs
     - Write CVSS v3.1 Structured Markdown Report to reports/
```

## 3. Autonomous Tooling Integration Guide

- **Reconnaissance**: `subfinder -d <target> | httpx -silent -status-code -title`
- **Crawling**: `katana -u <url> -silent -depth 3`
- **Fuzzing**: `ffuf -u <url>/FUZZ -w /home/ikhsan/Documents/Cybermes/tools/wordlists/common.txt -mc 200,301,302,403`
- **XSS Analysis**: `dalfox url <url> --silence`
- **SQLi Verification**: `sqlmap -u "<url>?param=1" --batch --banner`
- **Browser Automation**: Use MCP tools (`puppeteer_navigate`, `puppeteer_screenshot`, `puppeteer_evaluate`) for dynamic JavaScript and DOM verification.

## 4. Verification Checkpoint Matrix

| Vulnerability Type | Validation Standard | Evidence Requirement |
|---|---|---|
| **IDOR / BOLA** | Cross-tenant access between Account A & B | Both Request/Response pairs with leaked object |
| **Auth Bypass** | Accessing protected endpoint without valid credentials | Response payload showing privileged view |
| **SQL Injection** | DBMS version or deterministic differential response | Database banner or timing difference proof |
| **XSS / DOM Flaw** | Script execution context or DOM sink manipulation | Browser MCP console logs / DOM snapshot |
| **SSRF** | Server-side request callback or internal metadata fetch | Raw response headers and body |

## 5. Target-Scoped Output Hierarchy & Deliverables
Always create target-scoped directories before saving outputs (`TARGET_SLUG` e.g. `example_com` or `127_0_0_1_8888`):
- Directory Setup: `mkdir -p reports/<TARGET_SLUG>/findings reports/<TARGET_SLUG>/pocs reports/<TARGET_SLUG>/evidence`
- Finding Report: `reports/<TARGET_SLUG>/findings/[<SEVERITY>]_<vuln_name>.md`
- Standalone PoC: `reports/<TARGET_SLUG>/pocs/poc_<vuln_name>.py`
- Visual / HTTP Evidence: `reports/<TARGET_SLUG>/evidence/<filename>`
- Consolidated Summary: Run `python3 tools/aggregate_reports.py <TARGET_SLUG>` to index findings into `reports/<TARGET_SLUG>/SUMMARY.md`.
