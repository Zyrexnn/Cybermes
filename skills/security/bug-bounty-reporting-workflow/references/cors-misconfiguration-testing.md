# CORS Misconfiguration Testing — Recipe & Evidence

**Discovered:** Session 2026-08-16, monash.edu assessment
**CWE:** CWE-942 (Permissive Cross-domain Policy with Untrusted Domains)
**CVSS example:** 7.1 (CVSS:3.1/AV:N/AC:L/PR:N/UI:R/S:C/C:H/I:L/A:N)

## What to look for
Any API endpoint returning `access-control-allow-origin` in responses. Common on SPAs
with separate API backends (`/api/v1/*`, `/api/*`).

## Test commands
```bash
BASE="https://admin-forms-dev.monash.edu/api/v1/me"

# 1. Attacker-controlled origin
curl -s -I "$BASE" -H "Origin: https://evil-attacker.com" | grep -i "access-control-allow-origin"

# 2. null origin (sandboxed iframes, data: URLs)
curl -s -I "$BASE" -H "Origin: null" | grep -i "access-control-allow-origin"

# 3. Another arbitrary origin
curl -s -I "$BASE" -H "Origin: https://attacker.example.com" | grep -i "access-control-allow-origin"
```

## Vulnerable pattern (confirmed at monash.edu)
All three requests returned the SAME header:
```
access-control-allow-origin: https://uniweb-staging.apps.monash.edu
```
The server ignored the incoming `Origin` and echoed a hardcoded staging domain.
This is a misconfigured (not reflective) CORS policy.

## Why it matters even with 401
The endpoint returned `HTTP 401` without credentials — auth was enforced. BUT the
misconfigured CORS header is still exposed. Impact: if `uniweb-staging.apps.monash.edu`
is compromised (XSS or subdomain takeover), an attacker can make cross-origin
authenticated requests to the admin API from that origin.

## Full header evidence (monash.edu)
```
HTTP/2 401
access-control-allow-origin: https://uniweb-staging.apps.monash.edu
cache-control: no-store, max-age=0, must-revalidate
content-security-policy: default-src 'self' 'nonce-...' blob: data: www.google.com www.gstatic.com
server: Google Frontend
```

## Remediation to include in report
- Remove hardcoded origin; only allow the specific frontend domain that consumes the API.
- Validate `Origin` against an allowlist; reflect dynamically only for approved domains.
- If API is not browser-consumed, disable CORS entirely.
- Monitor the allowed origin for subdomain takeover risk.

---

## ROUND 2 ADDITIONS (Session 2026-08-17)

### Severity tiers — do NOT treat all CORS issues as equal
| Pattern returned | Severity | Why |
|------------------|----------|-----|
| `access-control-allow-origin: *` | **CRITICAL** (CVSS ~9.1) | Any origin accepted. If any endpoint reads cookies/session, full cross-origin read from any site. Even without credentials, unauthenticated API data is readable cross-origin. |
| Static specific origin (ignores input) | **High** (CVSS ~7.4) | Only exploitable if that trusted origin is itself compromised (XSS / subdomain takeover). Chain dependency. |
| Reflects attacker origin + `Allow-Credentials: true` | **Critical** | Direct credentialed cross-origin read. |
| Reflects attacker origin, no credentials | Medium | Limited to unauthenticated data. |

**Key check:** after finding ANY CORS header, test with 3+ distinct origins AND `null`. If the value is literally `*` → Critical. If it's identical static value for all → High (static). If it echoes your sent origin → reflective (check credentials flag).

### Wildcard `*` variant (found on Django host)
```bash
curl -s -I "https://crams-cloud-api-dev.erc.monash.edu/" \
  -H "Origin: https://evil.com" | grep -i "access-control"
# vary: Origin
# access-control-allow-origin: *
```
Confirmed on `/`, `/api/`, `/static/` paths. This is worse than the static-origin
case because NO origin validation occurs at all.

### Multi-host CORS sweep (find every affected host at once)
After discovering the static origin on one host, sweep ALL dev/uat hosts with the same
test — often the same misconfig is copy-pasted across environments:
```bash
HOSTS="admin-forms-dev.monash.edu forms-dev.monash.edu admin-forms-uat.monash.edu graduationportal-uat.monash.edu"
for h in $HOSTS; do
  AC=$(curl -s -I "https://$h/" -H "Origin: https://evil-attacker.com" --max-time 6 2>/dev/null \
        | grep -i "access-control-allow-origin" | tr -d '\r')
  [ -n "$AC" ] && echo "[CORS] $h -> $AC"
done
# All 4 returned: access-control-allow-origin: https://uniweb-staging.apps.monash.edu
```

### Discovery pipeline: JS bundle → API map → CORS test
The CORS misconfig is usually only reachable AFTER you know the API paths exist.
1. Download the SPA's main JS bundle (find `<script src="/assets/index-*.js">` in page HTML).
2. Grep endpoints + methods:
   ```bash
   grep -oE '/api/v[0-9]/[a-zA-Z0-9/_-]+' bundle.js | sort -u
   grep -oE '/api/v1/[a-zA-Z0-9/_-]+.{0,40}method:"[A-Z]+"' bundle.js | grep -oE '/api/v1/[a-zA-Z0-9/_-]+|method:"[A-Z]+"' | paste - -
   ```
3. For each base path, send a CORS-preflight-style request with a malicious `Origin`
   and grep `access-control-allow-origin`. This is how the 28-endpoint map + 4-host
   CORS finding was produced in one pass.

### Trust-dependency analysis for static-origin findings
- Resolve the trusted origin's DNS (dig/nslookup). If it points to a SaaS/CMS
  (e.g. Squiz Cloud, `mon-web.matrix.squiz.cloud`) it is NOT a dangling record →
  no DNS subdomain takeover, but XSS-on-that-origin still enables the chain.
- Note the trusted origin in the report and recommend replacing the static header
  with dynamic allowlist validation.
