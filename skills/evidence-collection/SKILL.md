---
name: evidence-collection
description: Captures, timestamps, formats, and sanitizes verifiable evidence (HTTP transactions, deterministic PoC scripts, screenshots, and terminal logs) following the "No PoC, No Finding" standard.
---

# Evidence Collection Skill — Deterministic PoC & Artifacts

## Purpose
Preserve high-fidelity, auditable, and sanitized evidence for every security observation. Ensure that every reported finding has a deterministic, self-contained Proof-of-Concept (PoC) script that allows program triagers to reproduce the exact behavior in seconds.

## Core Principle: "No PoC, No Finding"
A vulnerability is never finalized into a formal report unless it is backed by an unambiguous, working, and non-destructive Proof-of-Concept.

## Workflow & Artifact Generation

### 1. Deterministic Python / Bash PoC Script
For every valid finding, create a self-contained reproduction script under `/workspace/reports/evidence/<target>/<finding_id>/poc.py` (or `poc.sh`):

```python
#!/usr/bin/env python3
"""
Reproducible PoC for [Vulnerability Title]
Target: https://target.example.com
Author: Hermes Security Agent
"""
import requests

TARGET_URL = "https://target.example.com/api/v1/resource"
HEADERS = {
    "User-Agent": "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36",
    "X-Test-Header": "PoC-Verification"
}

def verify_vulnerability():
    print(f"[*] Sending verification probe to {TARGET_URL}...")
    response = requests.get(TARGET_URL, headers=HEADERS, timeout=10)
    
    # Assert deterministic evidence condition
    if response.status_code == 200 and "root:" in response.text:
        print("[+] VULNERABILITY CONFIRMED: Deterministic indicator matched.")
        print(f"[+] Status: {response.status_code}, Response snippet: {response.text[:200]}")
        return True
    else:
        print("[-] Verification failed or indicator not observed.")
        return False

if __name__ == "__main__":
    verify_vulnerability()
```

### 2. Preserve Raw HTTP Request & Response
Store raw HTTP transactions in standard format:
```http
POST /api/v1/user/profile HTTP/1.1
Host: target.example.com
Authorization: Bearer [REDACTED_TEST_TOKEN]
Content-Type: application/json

{"name": "test_user"}

HTTP/1.1 200 OK
Content-Type: application/json

{"status": "success", "id": 1234}
```

### 3. Visual Browser Evidence (Browser MCP)
When verifying client-side or UI vulnerabilities (DOM XSS, CSRF, sensitive page exposure), capture a screenshot via Browser MCP:
```bash
# Saved to: /workspace/reports/evidence/<target>/<finding_id>/screenshot.png
```

### 4. Sanitization & Zero False Positives
- Redact PII, customer secrets, and production credentials with `[REDACTED]`.
- Confirm that the evidence does not rely on transient server blips or generic 404/500 error pages.

## Evidence Directory Layout
```text
/workspace/reports/evidence/<target>/<finding_id>/
├── poc.py                 # Self-contained executable reproduction script
├── poc_request.txt        # Full raw HTTP request
├── poc_response.txt       # Full raw HTTP response
├── screenshot.png         # Optional visual proof (for client-side/DOM flaws)
└── execution_log.txt      # Execution trace log
```
