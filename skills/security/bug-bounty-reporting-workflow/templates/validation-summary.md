# Validation Summary — Template

Drop this table + sections into the final report. Every finding must have a ✅ in
"Manual Verification" AND a reproducible PoC before it earns "CONFIRMED".

## Validation Status Summary

| # | Finding | Severity | Manual Verification | Automated Scan | PoC Generated | Status |
|---|---------|----------|---------------------|----------------|---------------|--------|
| 1 | <name> | High | ✅ <method> | ✅ <tool> | ✅ <cmd> | **CONFIRMED** |
| 2 | <name> | Medium | ✅ <method> | ❌ | ✅ <cmd> | **CONFIRMED** |
| 3 | <name> | Low | ✅ <method> | ✅ <tool> | ⚠️ partial | **CONFIRMED** |

## Validation Methods Used
- ✅ Direct HTTP probing (curl, httpx) — rate-limited
- ✅ Multiple-origin CORS testing (≥3 malicious origins)
- ✅ JS bundle reverse engineering (X MB analyzed)
- ✅ Version fingerprinting (from multiple sources)
- ✅ CSRF token / session cookie extraction
- ✅ Directory listing content extraction
- ✅ Nuclei automated scanning

## False Positive Elimination (must document negatives)
- ✅ All <N> API endpoints return 401 without auth (proper access control)
- ✅ Production `.env` NOT exposed (only `.env.example` found)
- ✅ `wp-config.php` properly processed by PHP (HTTP 200, 0 bytes — not leaked)
- ✅ <other negative checks relevant to this target>

## Per-Finding PoC Block (example)
```
# Finding #N — <title>
curl -sI "https://TARGET/path" -H "Origin: https://evil-attacker.com" | grep -i "access-control"
# → access-control-allow-origin: https://trusted-host.example.com   (static, vuln)
```
