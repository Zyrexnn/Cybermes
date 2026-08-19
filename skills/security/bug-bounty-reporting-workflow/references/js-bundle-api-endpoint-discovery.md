# JS Bundle API Endpoint Discovery

## When to Use
After identifying a SPA / React / Vue / Angular app during recon, the client-side JS bundle is the richest passive source of:
- Internal API endpoint paths (`/api/v1/...`, `/api/v2/...`)
- Expected parameter names and data shapes
- SSO / auth integration points (e.g. `rn.sso()`, Okta, Azure AD)
- Hidden administrative or state-changing functions (POST / PATCH / DELETE)
- Environment switchers (dev / uat / prod host switching)

This is **passive recon** — no requests hit the API, so it never trips WAF/Cloudflare and is safe on bot-managed hosts.

## Procedure (stealth, low-rate)

### 1. Locate the bundle from the homepage HTML
```bash
curl -s "https://TARGET/" | grep -oP 'src="[^"]+\.js[^"]*"' | sed 's/src="//;s/"$//'
# Module / type=module builds:
curl -s "https://TARGET/" | grep -oE '(https?://[^"'\''<>]+\.js)'
# Or look for the asset manifest tag
curl -s "https://TARGET/" | grep -iE 'script.*\.js'
```

### 2. Download (single polite request — no crawling the JS host)
```bash
curl -s "https://TARGET/admin/assets/index-BrgMPTOh.js" -o bundle.js
wc -c bundle.js   # bundles are often 0.5–2 MB; never cat them
```

### 3. Extract quoted API paths
```bash
grep -oE '"/api/v[0-9]/[a-zA-Z0-9/_-]+"' bundle.js | sort -u
```

### 4. Extract template-literal / dynamic API paths (easy to miss)
```bash
grep -oE 'fetch\(`/api/v[0-9]/[^`]+`' bundle.js | sort -u
grep -oE '"/api/v[0-9]/[a-zA-Z0-9/_-]+\?[^"]*"' bundle.js | sort -u
```

### 5. Hunt secrets / tokens / config leaks
```bash
grep -iE '(api_key|apikey|secret|token|password|access_key|private|credential|bearer)' bundle.js
# Also look for hardcoded hosts / internal URLs
grep -oE 'https?://[a-z0-9.-]+\.(monash|internal|local)[a-z0-9./-]*' bundle.js | sort -u
```

### 6. Map each endpoint to behaviour (manual, high-value)
Search the bundle for each endpoint string to find the surrounding `fetch()` call:
- HTTP method (GET vs POST/PATCH/DELETE → state-changing?)
- 401 handling (does it call `sso()` / redirect to login? → properly protected)
- Returned field names (studentID, staffID, email, sapEmployeeID → PII sensitivity)

## Real Example — Monash `admin-forms-dev.monash.edu`
From a 1.08 MB minified React bundle, extracted 20+ endpoints:
- `/api/v1/submissions?is_processed=false` — form submissions (PII)
- `/api/v1/user-lookup` — student/staff ID resolution
- `/api/v1/user-overrides` — privilege management (POST/PATCH/DELETE)
- `/api/v1/system-details` — system configuration
- `/api/v1/assessment-attributes`, `/api/v1/student-attributes` — academic data

All returned `401` without auth (properly protected), but a static CORS header (`access-control-allow-origin: https://uniweb-staging.apps.monash.edu` for every Origin) made them cross-origin readable if the staging host is compromised. The endpoints alone were Medium; combined with CORS they escalated to High.

## Pitfalls
- **Never `cat` a minified bundle** — terminal truncates at ~50 KB. Write to file, then `grep`.
- **Grep both quoted AND backtick forms** — dynamic endpoints only appear in template literals.
- **Don't assume unauthenticated access** — verify each endpoint returns 401 before reporting "exposed data". Endpoint *discovery* is informational; *exploitable exposure* requires the 401 to be bypassed or CORS/IDOR to apply.
- **This is recon, not exploitation** — never POST/PATCH/DELETE to state-changing endpoints without explicit auth scope. Read-only validation only.
- **Obfuscated bundles** may mangle paths; if `grep` finds nothing, try grepping for a known substring like `api/` or `v1/`.
