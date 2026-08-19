# AI Router Testing Playbook

## Purpose
Standardized testing methodology for AI Router/Gateway services (Xyrus Router, Heraxles, New API, One API, etc.).

## Target Classification
AI Router services typically have:
- **Control Plane**: User dashboard, admin panel, package management (often Vercel/Netlify)
- **Data Plane**: API Gateway that routes to various AI models (often Railway/Render/AWS)
- **Split Architecture**: Control and Data planes may have separate authentication

## Standard Test Cases

### 1. Information Disclosure
**Test:** `GET /api/v1/models` or `/v1/models`
```bash
curl -s "https://<target>/api/v1/models" -H "User-Agent: Mozilla/5.0" -w "\nHTTP Status: %{http_code}\n"
curl -s "https://<target>/v1/models" -H "User-Agent: Mozilla/5.0" -w "\nHTTP Status: %{http_code}\n"
```
**Expected:** Should return 401/403 without authentication
**Finding:** If returns 200 with model list → **Information Disclosure (HIGH)**
**CVSS:** 7.5 (CVSS:3.1/AV:N/AC:L/PR:N/UI:N/S:U/C:H/I:N/A:N)

### 2. Admin Endpoint Discovery
**Test common admin paths:**
```bash
for path in "admin" "admin/" "dashboard" "dashboard/" "account/admin" "account/admin/" "panel" "panel/"; do
  echo "=== Testing: /$path ==="
  curl -s -I "https://<target>/$path" -H "User-Agent: Mozilla/5.0" -w "HTTP Status: %{http_code}\n" -o /dev/null 2>&1
done
```
**Expected:** 404 or 401/403
**Finding:** If returns 200 or detailed error → **Admin Panel Exposure**

### 3. Admin API Endpoint Testing
**Test admin endpoints with different methods:**
```bash
# GET request
curl -s "https://<target>/account/admin/vouchers" -H "User-Agent: Mozilla/5.0" -w "\nHTTP Status: %{http_code}\n"

# POST request with empty body
curl -s -X POST "https://<target>/account/admin/vouchers" -H "Content-Type: application/json" -d '{}' -w "\nHTTP Status: %{http_code}\n"

# POST request with test data
curl -s -X POST "https://<target>/account/admin/vouchers" -H "Content-Type: application/json" -d '{"days": 1}' -w "\nHTTP Status: %{http_code}\n"
```
**Expected:** 401 Unauthorized (login required)
**Finding:** If returns validation errors or field requirements → **Information Disclosure via Error Messages (MEDIUM)**

### 4. Rate Limiting Testing
**Test for missing rate limiting:**
```bash
# Send 20 rapid requests
for i in {1..20}; do
  curl -s -o /dev/null -w "%{http_code} " "https://<target>/api/v1/models"
done
echo ""
```
**Expected:** 429 Too Many Requests after threshold
**Finding:** If all requests return 200/401 → **Missing Rate Limiting (MEDIUM)**

### 5. API Key Testing
**Test with invalid/missing API keys:**
```bash
# No API key
curl -s "https://<target>/v1/usage/detail" -H "User-Agent: Mozilla/5.0" -w "\nHTTP Status: %{http_code}\n"

# Invalid API key
curl -s "https://<target>/v1/usage/detail" -H "Authorization: Bearer invalid_key" -w "\nHTTP Status: %{http_code}\n"
```
**Expected:** 401 Unauthorized
**Finding:** If returns detailed error messages → **Information Disclosure (MEDIUM)**

### 6. CORS Testing
**Test CORS headers with arbitrary origins:**
```bash
for origin in "https://evil.com" "https://attacker.com" "null"; do
  echo "=== Testing Origin: $origin ==="
  curl -s -I "https://<target>/" -H "Origin: $origin" -H "User-Agent: Mozilla/5.0" | grep -i "access-control-allow"
done
```
**Expected:** No CORS headers or specific allowed origins
**Finding:** If `access-control-allow-origin: *` → **CORS Misconfiguration**
**Severity:**
- `*` + `Access-Control-Allow-Credentials: true` → **CRITICAL (9.1)**
- `*` without credentials → **MEDIUM (5.3)**
- Static origin ignoring input → **MEDIUM (5.3)**

### 7. Error Message Analysis
**Collect error messages from all endpoints:**
```bash
# Test various endpoints and collect error responses
endpoints=("/" "/api/v1/models" "/v1/usage/detail" "/account/admin/vouchers" "/token/")
for endpoint in "${endpoints[@]}"; do
  echo "=== $endpoint ==="
  curl -s "https://<target>$endpoint" -H "User-Agent: Mozilla/5.0" 2>&1 | head -5
  echo ""
done
```
**Finding:** If error messages reveal internal structure → **Information Disclosure (MEDIUM)**

### 8. Security Headers Check
**Check for missing security headers:**
```bash
curl -s -I "https://<target>/" -H "User-Agent: Mozilla/5.0" | grep -iE "(content-security-policy|x-frame-options|x-xss-protection|permissions-policy|cross-origin)"
```
**Expected:** All major security headers present
**Finding:** Missing headers → **Security Misconfiguration (LOW-MEDIUM)**

## Common Findings for AI Router Services

### High Severity
1. **Unauthenticated Model Registry Access** (`/api/v1/models`)
   - CVSS: 7.5
   - Impact: Commercial model configuration exposed
   - CWE: CWE-200 (Information Exposure)

2. **Admin API Without Authentication** (`/account/admin/*`)
   - CVSS: 7.5-9.0 (depending on functionality)
   - Impact: Full admin access possible
   - CWE: CWE-284 (Improper Access Control)

### Medium Severity
1. **Missing Rate Limiting**
   - CVSS: 6.5
   - Impact: Brute force attacks possible
   - CWE: CWE-770 (Allocation of Resources Without Limits)

2. **Information Disclosure via Error Messages**
   - CVSS: 5.3
   - Impact: API structure and field names exposed
   - CWE: CWE-209 (Information Exposure Through Error Message)

3. **CORS Misconfiguration** (without credentials)
   - CVSS: 5.3
   - Impact: Cross-origin data access (unauthenticated only)
   - CWE: CWE-942 (Permissive Cross-domain Policy)

### Low Severity
1. **Missing Security Headers**
   - CVSS: 4.3
   - Impact: Increased vulnerability to common attacks
   - OWASP: Security Headers Best Practices

## Reporting Template for AI Router Findings

```markdown
### [F-XX] [Severity] - [Title]
**CVSS:** [score] ([vector])
**CWE:** [CWE-ID] - [CWE Name]
**OWASP:** [OWASP Category]

#### Description
[Brief description of the finding]

#### Evidence
**Request:**
```http
[HTTP request]
```

**Response:**
```http
[HTTP response]
```

**Reproduction Command:**
```bash
[curl command for reproduction]
```

#### Impact
- **Confidentiality:** [High/Medium/Low/None]
- **Integrity:** [High/Medium/Low/None]
- **Availability:** [High/Medium/Low/None]

#### Remediation
1. [Step 1]
2. [Step 2]
3. [Step 3]

#### References
- [OWASP Link]
- [CWE Link]
```

## Tools Configuration

### Safe Rate Limits for Cloudflare Targets
```bash
# Katana
katana -rate-limit 5 -delay 1000  # 5 req/sec with 1s delay

# httpx
httpx -rate-limit 15 -t 30  # 15 req/sec max, 30 threads

# curl
curl --max-time 15 -A "Mozilla/5.0"  # Always use browser UA
```

### Recommended Tools
- **Reconnaissance:** `subfinder`, `dnsx`, `httpx`
- **Crawling:** `katana` (with rate limiting)
- **API Testing:** `curl`
- **Browser Testing:** Browser MCP for JavaScript-heavy sites

## Case Studies
- **xyrusrouter.xyz:** See `references/ai-router-assessment-xyrus.md`
- **heraxles.my.id:** See `references/ai-router-assessment-heraxles.md`