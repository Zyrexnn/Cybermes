---
name: bug-bounty-reporting-workflow
description: Class-level workflow for authorized bug bounty / web pentest engagements — covers report language, mandatory validation gate, sensitive-data emphasis, and CORS/API discovery techniques. Use whenever the user asks to "check", "recon", "pentest", or write a bug bounty report for a target.
---

# Bug Bounty Reporting Workflow

## Purpose
Encodes durable preferences and validated techniques for bug bounty / web pentest engagements discovered during the monash.edu assessment (Session 2026-08-16). Governs HOW reports are produced and HOW findings are validated, complementing the `reporting` and `web-pentest` skills (which are manually authored and should be consulted for template/structure).

## User Preferences (Hard Requirements)

**1. Report Language = English**
Even if the user communicates in another language (Indonesian, etc.), the bug bounty deliverable MUST be written in English. Explicit user instruction: "berbahasa inggris".

**2. Validation Gate (Mandatory)**
Never add a finding to the report until it is validated with concrete evidence. User instruction: "jika menemukan sesuatu validasi lagi dan masukin ke laporan" (if you find something, validate it again then include it).
- Automated detections (Nuclei) still require manual confirmation before report inclusion.
- Include a **Validation Status Summary** table at report end: `| Finding | Status | Validation Method | Confirmed |`.

**3. Sensitive Data Emphasis**
When the user requests "full detail api atau data sensitif lainnya" (full detail on APIs or other sensitive data), the report must:
- Enumerate ALL discovered API endpoints: path, HTTP method, auth requirement (e.g. 401 without token), data type exposed.
- Document exposed secrets, tokens, keys, or PII-handling endpoints.
- Include evidence of data sensitivity via field names found in JS bundles or API responses (e.g. `studentID`, `staffID`, `email`, `sapEmployeeID`).

**4. Stealth / WAF-Avoidance (Hard Requirement)**
User instruction (verbatim, Session 2026-08-16): "jangan sampe ke block cloudflare dan sejenis nya" — never trip WAF/Cloudflare/rate-limit blocks during scanning. This is a standing constraint for this user, not a one-off.
- **Passive-first:** Start with sources that do NOT send requests to the target — subfinder, crt.sh (cert transparency), gau/wayback. Resolve DNS with dnsx (no HTTP to target) before any probing.
- **Low-rate active probing:** Always cap request rate. Validated-safe values: `httpx -rate-limit 15`, `katana -rate-limit 8 -delay 1500`. Treat these as ceilings, not targets.
- **Skip Cloudflare-Bot-Management hosts:** If tech-detect shows `Cloudflare` + `Cloudflare Bot Management`, do NOT fuzz/brute those hosts — passive enumeration only. Active fuzzing there is what triggers blocks.
- **Background long crawls:** `katana` at depth ≥2 exceeds the 180s foreground tool cap. Run it with `terminal(background=true, notify_on_complete=true)` and poll, rather than foreground.
- **Detect & back off:** If you see `429`/`403`/`000`/empty responses mid-scan, stop that host and note it as a control — do not retry aggressively.
- See `references/stealthy-recon-recipe.md` for the exact command sequence that completed cleanly against nurulfikri.ac.id (80 subdomains, 76 live, 0 blocks).

## Engagement Workflow (Quick Check → Deep Dive)

**Phase 0 — Recon (quick, stealthy):**
```bash
# Passive, no direct hits to target
subfinder -d <target> -silent -recursive -o subdomains.txt
dnsx -l subdomains.txt -silent -r 8.8.8.8,1.1.1.1 -o resolved.txt
# Active probe — RATE LIMITED to avoid WAF/Cloudflare blocks
httpx -l resolved.txt -silent -t 30 -rate-limit 15 -timeout 8 \
  -ports 80,443 -title -status-code -tech-detect -server -follow-redirects -o httpx.txt
# Optional passive historical URLs (wayback, no direct hits)
gau --subs <target> 2>/dev/null | sort -u > gau.txt
# NOTE: if httpx shows Cloudflare Bot Management on a host, do NOT fuzz it.
```

**Phase 1 — Prioritize:** Flag legacy software (WordPress < 5.5, EOL PHP/Apache), dev/uat subdomains (`*-dev`, `*-uat`, `*-staging`), exposed admin panels, directory listing.

**Phase 2 — Validate & Deep Dive (per priority target):**
- WordPress: check `/wp-json/wp/v2/users`, `/readme.html`, `/xmlrpc.php`, plugin enumeration.
- Django: check `/admin/login/`, `/static/` directory listing, CSRF token presence.
- JS bundles: **This is the highest-yield passive step — do it on every SPA target.** `curl` the `.js`, then `grep` for `api/v1/`, `sso(`, secrets, endpoints. See `references/js-bundle-api-endpoint-discovery.md` for the full procedure (quoted vs backtick endpoints, secrets hunt, PII field mapping, never `cat` a minified bundle). A 1.08 MB Monash React bundle yielded 20+ endpoints (incl. `/api/v1/user-overrides`, `/api/v1/submissions`) this way — no requests hit the API, so it is WAF-safe and never triggers blocks.
- API: test with/without auth, test CORS (see Technique below).

**Phase 3 — Report:** Use `reporting` skill structure + this skill's preferences. Single consolidated file for 3+ findings.

## Validated Technique: CORS Misconfiguration Discovery

When a target's API returns `access-control-allow-origin`, test for a hardcoded/misconfigured policy:
```bash
# Send arbitrary origins — if response header is IDENTICAL regardless of input, it's vulnerable
curl -s -I "https://target.com/api/me" -H "Origin: https://evil-attacker.com" | grep -i "access-control-allow-origin"
curl -s -I "https://target.com/api/me" -H "Origin: null" | grep -i "access-control-allow-origin"
curl -s -I "https://target.com/api/me" -H "Origin: https://attacker.example.com" | grep -i "access-control-allow-origin"
```
- **Vulnerable pattern:** Same value returned for all origins (e.g. always `https://uniweb-staging.apps.monash.edu`).
- **CWE-942** (Permissive Cross-domain Policy with Untrusted Domains).
- **Impact:** If the allowed origin is compromisable (XSS / subdomain takeover), cross-origin authenticated API access is possible.
- **Note:** A 401 on the endpoint does NOT negate the CORS finding — the misconfigured header is still exposed.

**Severity differentiation (do NOT lump all CORS findings together) — CORRECTED after round-3 re-validation proved the old `*`=Critical rule was inflation:**
- `access-control-allow-origin: *` + `Allow-Credentials: true` → **CRITICAL** (CVSS ~9.1). Credentialed cross-origin read; cookie/auth theft possible.
- `access-control-allow-origin: *` WITHOUT `Allow-Credentials` → **Medium** (CVSS ~5.3). Only UNAUTHENTICATED responses readable cross-origin; no cookie/auth theft. This was wrongly rated Critical in the monash.edu round-2 report.
- Static specific origin ignoring input (no `Allow-Credentials`) → **Medium** (CVSS ~5.3). Exploitable only via XSS chain on the trusted origin.
- Reflects attacker origin + `Allow-Credentials: true` → **Critical** (credentialed cross-origin read).
After ANY CORS header, send 3+ origins + `null` AND always grep for `access-control-allow-credentials`. That flag's presence/absence is the single biggest severity driver — never call `*` Critical without confirming it.

**Multi-host sweep:** the same misconfig is often copy-pasted across dev/uat hosts. After finding it on one, sweep all sibling environments with the same Origin test (see reference for the loop).

**Discovery pipeline:** CORS bugs are usually only reachable once you know the API paths exist. Download the SPA's main JS bundle (`<script src="/assets/index-*.js">`), grep `/api/vN/...` paths + `method:"..."` to build the endpoint map, then CORS-test each base path. This produced a 28-endpoint map + 4-host CORS finding in one pass during the monash.edu round-2 assessment.

- **Reference:** See `references/cors-misconfiguration-testing.md` for the full test recipe, wildcard variant, severity tiers, multi-host sweep, and evidence from the monash.edu assessment.

## Target-Class Testing Playbooks (updated 2026-08-18)
For specific app classes seen in the wild, see:
- `references/ai-gateway-newapi-midtrans-librechat-testing.md` — playbooks covering AI Gateway, New API, One API, Midtrans-backed storefronts, and LibreChat.
- `references/ai-router-assessment-xyrus.md` — case study for xyrusrouter.xyz (Vercel + Railway split architecture).
- `references/ai-router-assessment-heraxles.md` — case study for heraxles.my.id (Cloudflare-protected AI Router).
- `references/ai-router-testing-playbook.md` — standardized testing methodology for AI Router services.

Playbooks cover:
- **AI Gateway / Router (Xyrus Router, New API, One API, Heraxles)**:
  - **Information Disclosure:** Check `GET /api/v1/models` (or `/v1/models`). It frequently lacks auth and leaks the model catalog + pricing multipliers.
  - **Admin Endpoint Discovery:** Test `/admin/`, `/admin`, `/account/admin/`, `/dashboard/` for exposed admin panels.
  - **Infra Split:** Distinguish between the Control Plane (Dashboard) and Data Plane (Gateway). Keys may be scope-limited. Check backend domains like `*.railway.app` or `*.vercel.app` revealed in JS or docs.
  - **CORS:** Permissive `*` on login/landing pages is common.
  - **Registration Flow Testing:** For gateways with user registration (e.g., Xyrus Router), see `references/ai-router-registration-testing.md` for a detailed playbook on testing CAPTCHA, email verification, password policies, and rate limiting.
- **Midtrans-backed storefront**: the `POST /api/order` → `snap_token` → `POST /api/order/:token/bind` flow, and the payment-bypass test (fake `settlement` status is rejected because the server re-verifies with Midtrans).
- **LibreChat**: `/api/config` disclosure, verified+rate-limited registration, 401 on protected routes.
- **Vercel Targets:** Generally more permissive for active scanning than Cloudflare; less prone to aggressive bot-management blocks, but still respect rate limits.
- **AI Router (Heraxles)**: Admin endpoints under `/account/admin/*` with detailed error messages; missing rate limiting on all endpoints; see `references/ai-router-assessment-heraxles.md`.
- **AI Router Testing Insight**: Test both GET and POST methods on admin endpoints as they may return different error messages (GET: "Method Not Allowed", POST: detailed validation errors). Also test with and without trailing slashes as behavior may differ (`/admin` vs `/admin/`).
- **AI Router Testing Insight**: Test both GET and POST methods on admin endpoints as they may return different error messages (GET: "Method Not Allowed", POST: detailed validation errors). Also test with and without trailing slashes as behavior may differ (`/admin` vs `/admin/`).
- **AI Router Testing Playbook:** Standardized test cases for AI Router services; see `references/ai-router-testing-playbook.md`.

## Responsible Disclosure Contact Discovery (find WHERE to send it)
Before submitting, locate the disclosure channel. Passive recon only — no scanning:
1. **`security.txt` (RFC 9116):** `curl https://<target>/.well-known/security.txt` (+ http variant). A 200 with `Contact:` + `Expires:` is the ideal channel. A WordPress 404 page here means "none configured".
2. **Footer / contact-page email harvest:** `curl` the homepage + 1–2 hub subdomains (e.g. `lppm`, `pmb`, `cdc`) and `grep -oiE "[a-z0-9._%+-]+@[a-z0-9.-]+\.(ac\.id|go\.id|com|org|net|id)"`. Common hits: `info@`, `<unit>@` per subdomain. (nurulfikri yielded `info@`, `lppm@`, `pmb@`.)
3. **No disclosure page?** Send to the institution's general `info@` address, subject "Security Vulnerability Report — <target>", request acknowledgement within 72h.
4. **Escalation (Indonesia):** For `.ac.id` / `.go.id`, the registrar is **PANDI / ID Registry**; if no response, escalate via registrar abuse channel or the institution's official social media.
5. Record discovered contacts in a "Responsible Disclosure Contact" appendix in the report.

## Output Location & file resilience
- `/workspace` is often non-writable. Fall back to `~/reports/bug-bounty-<target>.md` (home).
- **Pick ONE canonical path** and write only there. In the nurulfikri session the report silently scattered to `~/Documents/`, `~/Downloads/`, `~/Desktop/` AND `~/reports/` — and `~/reports/` later showed 0 bytes while `~/Documents/` held the real file (a copy had been edited, the rest drifted). Avoid this: after the final edit, `cp` the single canonical file to any other location the user expects; never edit multiple copies separately (they diverge silently).

## Pitfalls
- Don't trust Nuclei/automated output as "confirmed" — always manually validate (curl/browser) before reporting.
- Don't write the report in the user's language if they asked for English.
- Don't omit the Validation Status Summary — the user explicitly wants validation tracked. Use `templates/validation-summary.md` as the starter table.
- Don't breach the stealth/WAF-avoidance rule — never rate-limit above the validated ceilings, and never fuzz Cloudflare-Bot-Management hosts.
- WordPress REST API user enumeration may be WAF-blocked (HTTP 000 / empty) — that's a control, not a dead end; note it.
- JS bundle recon is the highest-yield passive step for SPA targets — see `references/js-bundle-api-endpoint-discovery.md`. Never `cat` a minified bundle; write to file and `grep`. Endpoint *discovery* is informational; only report *exposure* if 401 is bypassed or CORS/IDOR applies.
- **Run heavy scanners in background.** Nuclei with broad `-tags` (e.g. `cors,misconfig,exposure`) and `arjun` parameter-mining reliably exceed the 180s foreground tool cap and time out. Launch them with `terminal(background=true, notify_on_complete=true)` and poll, exactly like katana. Use narrow templates (e.g. `-tags cors`) or specific template paths to keep foreground scans fast. `arjun` uses `-d <ms>` for delay (NOT `--delay`), and pair with `-t 2 -q` for low-rate stealth.
- **Do a Round-2 deep-dive after Round-1 validation.** The first pass confirms obvious findings but misses variants: CORS wildcard `*` vs static-origin, multi-host CORS sweeps (misconfig copy-pasted across dev/uat), Okta/OAuth config disclosure via unauthenticated `/sso/redirect`, and missing-authZ routes that return SPA HTML (200) instead of 401. Each was found only in round 2 of the monash.edu assessment.
- **Test both GET and POST methods on admin endpoints.** Different HTTP methods may return different error messages and validation details. Found during heraxles.my.id assessment where GET returned "Method Not Allowed" but POST returned detailed field validation errors (e.g., `{"type":"missing","loc":["body","days"],"msg":"Field required"}`).
- **Test with and without trailing slashes.** Endpoints with and without trailing slashes may behave differently (307 redirects vs 404). Found during heraxles.my.id assessment where `/admin` returned 405 but `/admin/` returned 404.
- **Test both GET and POST methods on admin endpoints.** Different HTTP methods may return different error messages and validation details. Found during heraxles.my.id assessment where GET returned "Method Not Allowed" but POST returned detailed field validation errors.

## Independent Re-Validation Standard (anti-inflation, mandatory on user request)

When the user asks to "validate ulang" / re-validate / "jangan percaya laporan sebelumnya" / re-check all findings independently, apply this standard. It is a DIFFERENT mode from the build-and-report flow: here the job is to *tear down* inflated claims, not accumulate them.

**Core rule: every finding must be reproduced from scratch. A prior report's severity, CVSS, or "CONFIRMED" claim is NOT evidence.** Treat the previous report as an untrusted hint list.

**Per-finding verdict vocabulary (use exactly these labels):**
- `CONFIRMED` — reproduced with fresh request + concrete response indicator.
- `PROBABLE` — reproduced but impact depends on an unproven condition.
- `POTENTIAL` — version/behavior observed but exploitability unproven (e.g. outdated software with no PoC).
- `INFORMATIONAL` — real behavior, zero security impact on its own (endpoint names, public client_id, public dev host).
- `FALSE POSITIVE` — artifact mistaken for vuln (SPA HTML fallback ≠ auth bypass; CSRF token ≠ vulnerability).
- `NOT REPRODUCIBLE` — could not reproduce now (WAF/000/rate-limit/patched). Do NOT submit; note the prior state if known.
- `STALE` — was true earlier, no longer.
- `DUPLICATE` — same root cause as another finding (e.g. CORS on 4 sibling hosts = one finding, don't double-count).

**Anti-inflation rules (these were the explicit corrections in the monash.edu round-3 review):**
1. **CORS `*` without `Allow-Credentials` = Medium, NOT Critical.** No cookie/auth theft possible. Exploitation limited to unauthenticated responses.
2. **CSRF token presence = GOOD security, never a vulnerability.** Django admin exposed = publicly reachable login (brute-force risk), the token is protection.
3. **Outdated software ≠ Critical.** "WordPress 5.0.1 / 30+ CVEs" must be reduced to the CVEs that are actually applicable AND exploitable. Check: do they require auth? Is there a public unauthenticated RCE? If not, Medium (outdated software), not High.
4. **API endpoint discovery with 401 everywhere = INFORMATIONAL.** Endpoint names are not a vulnerability. Only report exposure if 401 is bypassed, or CORS/IDOR applies.
5. **Public dev/uat host = INFORMATIONAL by itself.** Not a bug unless you demonstrate impact: weaker auth than prod, debug/stack traces, real PII in test data, or direct prod-DB access from dev.
6. **OAuth `client_id` is public by spec — not a secret.** Only a finding if `redirect_uri` is manipulable (open redirect) or `state`/`nonce` missing. PKCE present = good.
7. **Distinguish SPA 200 (frontend fallback HTML) from API 200 (JSON data).** A `/api/*` path returning `<!DOCTYPE html>` is a routing artifact, not an auth bypass. Check `content-type` + first bytes before claiming missing-authZ.

**Re-validation report format (when producing the deliverable):** File `reports/FINAL-REVALIDATION-REPORT.md` with a summary table columns: `| # | Original Finding | Re-Test Status | New Severity | Evidence | Submit? | Reason |`. Then a Severity-Inflation-Corrections block showing old→new per finding, and a final "Submit / Don't Submit" split. Include a CONFIRMED FACTS vs HYPOTHETICAL IMPACT split for each kept finding so impact isn't inflated.

- **Reference:** See `references/independent-revalidation-standard.md` for the full checklist, verdict vocabulary, anti-inflation rules, deliverable format, and the monash.edu round-3 verdicts as worked examples. (linked: js-bundle-api-endpoint-discovery.md, stealthy-recon-recipe.md, cors-misconfiguration-testing.md, independent-revalidation-standard.md)
