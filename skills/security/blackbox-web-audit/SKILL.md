---
name: blackbox-web-audit
description: Authorized black-box web/API audit WITHOUT test accounts — unauth probing discipline, unauth-IDOR rounds using leaked IDs, JS-bundle-first attack surface mapping, non-destructive gates, and local report delivery conventions (no zip, follow sibling example folder). Use when asked to audit/pentest a target where you hold no credentials, and when writing the local report deliverable afterwards.
---

# Black-Box Web Audit Without Credentials

Class workflow for authorized audits where you have NO test account (operator-requested or VDP scope). Complements the manually-authored `bb-methodology` (5-phase workflow) and `bug-bounty-reporting-workflow` (validation standard) — load those too; this skill adds the no-credentials specifics they lack.

## Hard Constraints (non-destructive discipline)

1. **GET-only on anything state-changing-adjacent.** Shutdown/reset/redeem endpoints: probe with GET even if the app uses POST; NEVER send the real state-changing verb on production (a `POST /api/version/shutdown` was deliberately never sent).
2. **Fake IDs for write-path probes.** Create-order/payment endpoints tested only with nonexistent UUIDs; expect-and-document 401.
3. **Credentials are always fake** for rate-limit/login tests. A quarantine of your own IP after 429 is *evidence the control works* — record it as such.
4. Embedded keys found in bundles get **tested live against their real API** before any claim (Firebase: `identitytoolkit.googleapis.com/v1/accounts:createAuthUri?key=` → 400 "API key not valid" = revoked, informational only).

## Attack Surface When You Have No Account

Subfinder/gau often return near-zero on these targets. The bundle IS the map:

1. Download all `/_next/static/chunks/*.js` (or equivalent SPA bundles) to disk; never `cat` minified files — grep them.
2. Extract endpoint inventory (regex `/api/[a-z0-9/_-]+`, quoted AND backtick strings) → typically 50–100 endpoints from a ~1MB bundle.
3. Probe each with GET, no auth. **Verify content-type on EVERY probe**: Next.js/SPA backends answer unknown `/api/*` paths with HTTP 200 `text/html` (index fallback). Only `application/json` counts; HTML fallback ≠ open endpoint. This one check kills most false-positive "exposure" claims.
4. Mine JS for: hardcoded config (`authMode`, OIDC labels), sandbox/test features shipped to prod (`simulate-payment`, debug flags), embedded keys/client IDs.
5. Rate-limit test: N sequential bad logins to `POST /api/auth/login` → document lockout window (`retryAfter`) as a verified control.

## Unauth-IDOR Round (no second account needed)

Real IDs from ANY public leak make a meaningful object-access pass possible:

1. **Harvest real object UUIDs** from public catalog/listing endpoints (e.g. `/api/voucher-groups` leaking full catalog incl. hidden items).
2. **Replay GET-only** against every ID-bearing variant: direct path `/api/x/{id}`, query forms `?xId=` / `?x_id=`, sibling objects (`orders`, `vouchers`, `user`), admin-ish resources.
3. **Controls are mandatory:** same request with fake/nonexistent ID (expect 404) proves the route resolves IDs; valid-ID-no-auth returning uniform 401 proves auth gating. 200-valid/404-fake differential on a gated endpoint = protection working, NOT IDOR.
4. **Write the negative matrix into the report** (probe → status → verdict table in SUMMARY.md). Confirmed-no-IDOR with evidence demonstrates coverage; silence demonstrates nothing.
5. If a per-ID detail endpoint IS open (by design), fold it into the parent disclosure finding rather than inventing an IDOR claim.

## Complementary Negatives (run all, ~5 min)

- Parameter reflection: unique 8+ char canary through query params on main pages; count occurrences (0 expected on well-built SPAs).
- Open redirect: `?next=https://evil.com` / `?redirect=` — check Location header; fixed internal redirect = safe.
- Injection probes on path params (`'`, ` OR 1=1`, traversal): expect uniform 404/400 (parameterized); varied error bodies = investigate.
- CORS: send foreign origins + `null`; no reflection = clean (severity tiers live in `bug-bounty-reporting-workflow`).
- dalfox single-pass on parameterized pages for a scanner second-opinion.

## Environment Pitfalls

- **Cloudflare blocks default `python-urllib` UA with 403** while curl/browser UA passes. Every urllib/requests-based PoC MUST set a browser User-Agent or it fails where your curl probes succeeded.
- No `dig/host/nslookup` on many boxes → DNS-over-HTTPS via curl (`https://dns.google/resolve?name=X&type=A`).
- crt.sh intermittently returns malformed short responses → fall back to JS mining for subdomain/endpoint hints.
- Missing binaries are setup issues, not findings — note once, work around, move on.

## Report Delivery (user preferences — hard rules)

1. **NEVER ship a zip archive.** User explicitly rejected it ("kenapa bentuknya zip"). Plain individual files, ONE canonical location only.
2. **Before creating the report folder, look for a sibling example target and copy its structure exactly** (e.g. `/workspace/reports/<previous-target>/`). Canonical observed shape:
   - `SUMMARY.md` — emoji severity matrix (Severity | Title | CVSS | CWE | Endpoint | Link), PoC list, evidence list, verified-working-controls table, priority recommendations.
   - `metadata.json` — `{target, scan_time, total_findings, severity_summary{CRITICAL..INFORMATIONAL}, findings[{file_name,title,severity,cvss,cwe,endpoint,last_modified}], pocs[], evidence_files[]}`.
   - `findings/[SEVERITY]_slug.md` — Target/Severity/CVSS/CWE/Status header, raw request/response blocks, impact, verified-negatives, remediation.
   - `pocs/*.py`, `evidence/*`.
3. **Report language = whatever the CURRENT engagement specifies** (this user alternates: English ("berbahasa inggris") and Bahasa Indonesia). Ask once if unspecified; never inherit the last session's choice.
4. **Execute every delivered PoC end-to-end before shipping.** Both bugs found in the initial PoC (truncated-body slice breaking json.loads; leftover experimental line) surfaced only on first run.

## References

- `references/nextjs-gateway-case-study.md` — worked example: hardened Next.js+Cloudflare AI gateway, 71-endpoint JS-mined surface, unauth-IDOR round with negative matrix, 1 Low + 5 Info outcome.
