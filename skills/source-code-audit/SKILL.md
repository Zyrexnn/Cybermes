---
name: source-code-audit
description: Performs whitebox static application security testing (SAST), code review, API route extraction, credential hunting, and vulnerability pattern detection in local repositories or GitHub codebases.
---

# Source Code Audit Skill (SAST & Whitebox Security)

## Purpose
Simulate advanced whitebox penetration testing by inspecting source code repositories, extracting application routes, detecting insecure database queries, identifying weak authentication/authorization logic, and locating hardcoded secrets.

## Core Capabilities
- **Route & Controller Mapping**: Map all public and internal API endpoints (Express, Django, Flask, FastAPI, Spring Boot, Laravel, Next.js).
- **Injection Flaws Detection**: Scan for unsanitized SQL, command execution (`eval`, `exec`, `os.system`), template injection (Jinja2, Thymeleaf), and XPath/LDAP queries.
- **Authorization & Access Control**: Analyze middleware, RBAC checks, and missing object-level permission guards (IDOR).
- **Secret & Key Hunting**: Fast regex and entropy scanning for credentials, JWT secrets, and third-party API keys using `trufflehog` and `ripgrep`.

## Workflow

### 1. Codebase Profiling
Identify technology stack, framework, database connectors, and dependency manifests:
```bash
# Check dependencies and package configs
find . -maxdepth 3 -name "package.json" -o -name "requirements.txt" -o -name "pom.xml" -o -name "composer.json" -o -name "go.mod"
```

### 2. Secret & Sensitive Data Extraction
Run fast local entropy search for leaked credentials:
```bash
# Fast regex search with ripgrep
rg -i "(api_key|secret_key|private_key|password|jwt_secret|bearer)\s*[:=]\s*['\"][A-Za-z0-9_\-\.]{8,}['\"]" .

# Deep TruffleHog file scan
trufflehog filesystem . --only-verified --json 2>/dev/null
```

### 3. API Route & Attack Surface Mining
Extract declared HTTP routes:
- **Python (Flask/FastAPI/Django)**: `rg "@(app|router)\.(get|post|put|delete|patch)|path\(" .`
- **Node.js (Express/NestJS)**: `rg "(app|router)\.(get|post|put|delete|use)\(" .`
- **PHP/Laravel**: `rg "Route::(get|post|put|delete|any)\(" .`

### 4. High-Risk Vulnerability Pattern Scanning
- **Command Injection**: `rg -n "(subprocess\.Popen|os\.system|exec\(|child_process\.exec|shell_exec)" .`
- **SQL Injection**: `rg -n "(\.raw\(|\.execute\(|SELECT\s+.*\%s|f\"SELECT\s+|\.query\(\s*[\`\"']SELECT)" .`
- **Server-Side Request Forgery (SSRF)**: `rg -n "(requests\.get\(|axios\.get\(|fetch\(|curl_exec|urllib\.request)" .`
- **Insecure Deserialization**: `rg -n "(pickle\.loads|yaml\.load\(|unserialize\(|Marshal\.load)" .`

### 5. Formulation of Working Proof-of-Concept
Translate identified code flaws into concrete HTTP test cases (URLs, headers, JSON body payloads) to verify exploitability against the live deployment.

## Output Artifacts
- `/workspace/output/sast_findings.md` - Identified code-level flaws, affected file lines, and actionable exploitation paths.
