---
name: bug-bounty-target-prioritization
description: Autonomously prioritize and select high-value targets from reconnaissance results for deep vulnerability analysis in bug bounty engagements.
---

# Bug Bounty Target Prioritization

## Purpose
After initial recon (subdomain enumeration → DNS resolution → HTTP probing), autonomously identify and prioritize high-value targets for deep vulnerability testing without asking the user to choose.

## User Preference
**When user says "fokus yang menurtumu potensial ada celah nya" (focus on what you think has potential vulnerabilities):**
- They want **autonomous decision-making and execution**
- Don't present options: "Which do you want: (1) WordPress (2) Django (3) APIs?"
- **Do this instead:** Analyze findings, pick top 3-5 priority targets, explain briefly why they're valuable, and immediately start deep testing

## High-Value Target Indicators (Priority Order)

### 🔴 Critical Priority (Immediate Deep Dive)

**1. Legacy Software with Version Disclosure**
```
Indicators:
- WordPress < 5.5 (anything older than 2 years)
- Django with exposed /admin/ + no IP restrictions
- Apache < 2.4.48, Nginx < 1.20
- PHP 7.2 or below (all EOL since 2020)
- MySQL 5.x exposed via tech detection
- OpenSSL 1.1.1k or older

Why: Known CVEs, public exploits, low-hanging fruit

Detection pattern:
grep -E 'WordPress.*5\.[0-4]|Apache.*2\.4\.[0-3]|PHP.*7\.[0-2]' httpx.txt
```

**Example from real session:**
- `bus-web02-v01.ocio.monash.edu` → WordPress 5.0.1 (2019, 7 years outdated)
- Stack: Apache 2.4.37, PHP 7.2.24, OpenSSL 1.1.1k (all EOL)
- **Immediate priority** → Known CVEs present

**2. Admin Panels Exposed on Dev/Staging Environments**
```
Indicators:
- Django /admin/login/ without IP whitelisting
- WordPress /wp-admin/ on -dev/-uat/-staging subdomains
- PHPMyAdmin, Adminer, pgAdmin accessible
- Custom admin interfaces with "Administration" in title

Why: Weak/test credentials common, verbose errors, production-like data

Detection:
curl -s https://target-dev/admin/login/ | grep -i "admin\|login\|csrf"
```

**Example from session:**
- `crams-cloud-api-dev.erc.monash.edu/admin/login/`
- Title: "Crams-DB Administration"
- CSRF token present → live Django admin panel
- **Medium-high priority** → Brute force target, credential stuffing vector

### 🟡 High Priority (Quick Wins)

**3. Directory Listing Enabled**
```
Indicators:
- Nginx/Apache autoindex enabled on /static/, /assets/, /backup/
- Look for: .env files, config files, source maps, API documentation

Detection:
curl -s https://target/static/ | grep -i "index of"

What to look for in listings:
- admin/ directories (framework admin assets)
- .env.example (configuration patterns)
- Source maps (*.js.map → original source code)
- Backup files (.bak, .old, .backup)
- API documentation (swagger.json, openapi.yaml)
```

**Example from session:**
- `crams-cloud-api-dev.erc.monash.edu/static/` exposed
- Contents: `admin/`, `django_extensions/`, `rest_framework/`
- File timestamps reveal deployment dates
- **Medium priority** → Information disclosure, technology fingerprinting

**4. Dev/UAT/Staging Environments Publicly Accessible**
```
Subdomain patterns:
- *-dev.*, *-uat.*, *-staging.*, *.test.*
- dev.*, staging.*, qa.*, test.*

Port indicators:
- :8080, :3000, :8000, :8888

Why: Test credentials, debug mode enabled, verbose errors, production-like data exposure

Quick enumeration:
grep -oP 'https?://[^\s\[]+' httpx.txt | \
  grep -E '(admin|dev|test|staging|uat|api|portal)' > targets-priority.txt
```

**Example from session (30+ found):**
- `admin-forms-dev.monash.edu` → API endpoint `/api/v1/submissions?is_processed=false`
- `crams-cloud-api-dev.erc.monash.edu` → Full CRAMS system with admin panel
- `moodle-staging.vle.monash.edu` → Staging LMS environment

### 🟢 Medium Priority (Systematic Testing)

**5. REST API Endpoints Without Auth**
```
Paths to check:
- /api/, /api/v1/, /api/v2/
- /graphql (schema introspection)
- /wp-json/ (WordPress REST API)

Quick tests:
curl -s https://target/api/v1/ | python3 -m json.tool
curl -s https://target/wp-json/wp/v2/users  # User enumeration
curl -s https://target/graphql -d '{"query":"{__schema{types{name}}}"}'
```

**6. Technology Stack Fingerprinting**
```
Headers to check:
- X-Powered-By (framework/language)
- Server (web server version)
- X-AspNet-Version, X-Django-Version

Framework paths:
- /.well-known/ (configs, security.txt)
- /debug/ (Django debug mode)
- /static/admin/ (Django admin assets)
- /wp-includes/ (WordPress)

JavaScript source maps:
katana -u https://target -d 2 -jc -silent | grep '\.js$'
```

## Autonomous Decision-Making Workflow

```bash
# After initial recon complete (httpx.txt populated):

# 1. Prioritize targets automatically
cd ~/recon/<target>
grep -oP 'https?://[^\s\[]+' web/httpx.txt | \
  grep -E '(admin|dev|test|staging|uat)' > targets-priority.txt

# Add legacy software targets
grep -E 'WordPress.*5\.[0-4]|Apache.*2\.4\.[0-3]' web/httpx.txt >> targets-priority.txt
sort -u targets-priority.txt -o targets-priority.txt

# 2. Launch Nuclei scan on priority targets (background)
nuclei -l targets-priority.txt \
  -severity critical,high,medium \
  -silent -nc -o nuclei/priority-scan.txt \
  -timeout 30 &

# 3. Immediately start manual deep dive on #1 priority
# Don't wait for user — they asked you to focus on what's vulnerable

# Example: Legacy WordPress
curl -s https://wp-target/wp-json/wp/v2/users
wpscan --url https://wp-target --enumerate vp,vt,u

# Example: Django admin
curl -s https://django-dev/admin/login/ | grep csrf
curl -s https://django-dev/static/ | grep "Index of"

# Example: Directory listing
curl -s https://target/static/ | \
  grep -oP 'href="[^"]+' | sed 's/href="//g'
```

## Report Findings While Working

After discovering each significant finding:
1. **Document evidence** (HTTP requests/responses, version numbers)
2. **Assess severity** (CVSS score, impact)
3. **Keep scanning** — don't stop for user approval unless they interrupt

**Example autonomous flow:**
```
✅ Found: WordPress 5.0.1 (7 years outdated, multiple CVEs)
→ Document as High severity finding
→ Continue: Check for exposed admin panel, enumerate users via REST API

✅ Found: Django admin publicly accessible on dev subdomain  
→ Document as Medium-High severity
→ Continue: Check directory listing, enumerate static files

✅ Nuclei scan completes → merge results into report
→ Continue: Deep dive on remaining priority targets
```

## Decision Framework Summary

**When presented with 50-100 live HTTP services, select 3-5 priority targets based on:**

1. ✅ **Confirmed outdated software** (version numbers in response headers/content)
2. ✅ **Admin/auth panels on non-production** (`-dev`, `-uat`, `-staging` subdomains)
3. ✅ **Directory listing + sensitive paths** (`/static/admin/`, `/backup/`, `/config/`)
4. ✅ **REST APIs with enumerable resources** (`/api/v1/users`, `/wp-json/wp/v2/users`)
5. ❌ **Skip for now:** Cloudflare-protected production (save for WAF bypass phase)

## Pitfalls

### ❌ Don't Present Options When Priorities Are Clear
**Wrong:**
> "I found 97 live services. Which should I focus on: (1) WordPress, (2) Django, (3) API endpoints?"

**Right:**
> "Found 97 live services. Top priority: WordPress 5.0.1 (7 years outdated, known CVEs) and Django admin exposed on dev environment. Starting with WordPress deep dive."

### ❌ Don't Wait for Approval Between Findings
**Wrong:**
> "Found directory listing on /static/. Should I continue scanning?"

**Right:**
> "Found directory listing on /static/ exposing Django admin assets. Documented as medium severity. Continuing with admin panel verification..."

### ✅ Use Home Directory, Not /workspace
**Pitfall:** `/workspace` may have restricted permissions

**Solution:**
```bash
mkdir -p ~/recon/<target>/{subdomains,ports,web,nuclei,reports}
cd ~/recon/<target>

# NOT: /workspace/recon/<target> (may fail with Permission denied)
```

## Evidence Collection Pattern

Compile findings into structured markdown report as you work:

```markdown
# Critical Findings - <target>

## 🎯 FINDING #1: Legacy WordPress 5.0.1
**Severity:** HIGH | **CVSS:** 7.5
**Asset:** https://bus-web02-v01.ocio.monash.edu/

**Evidence:**
- Version: WordPress 5.0.1 (January 2019, 7 years outdated)
- Stack: Apache 2.4.37, PHP 7.2.24, OpenSSL 1.1.1k (all EOL)
- Known CVEs: CVE-2019-8942 (Path Traversal, CVSS 7.5)

**Impact:** Known exploitable vulnerabilities present, REST API user enumeration

**Recommendation:** Update WordPress to 6.x, PHP to 8.x immediately
```

Save to: `~/recon/<target>/reports/critical-findings.md`

## Session Statistics Template

Always provide quantified summary:
```
Subdomains Found:    2,103
DNS Resolved:        1,142 (54%)
HTTP Live:           97 (8.5%)
Priority Targets:    5 selected autonomously
Critical Findings:   2
High Findings:       3
```

## Related Skills
- `recon` - Initial subdomain enumeration and HTTP probing
- `vulnerability-analysis` - Process Nuclei scan results
- `web-enumeration` - Deep endpoint discovery with katana
- `reporting` - Format final bug bounty reports with CVSS scoring
