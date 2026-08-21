---
name: custom-ai-router-assessment
description: Assess self-hosted AI API routers/gateways built on CUSTOM web stacks (Next.js App Router class — Custom AI Router, 9router-style), as opposed to New API/One API forks. Use when target shows x-powered-by Next.js + /_next/static/chunks bundles, an AI-gateway product description, /subscription customer portal, /maintainer admin panel, or OpenAI-compatible /v1/* endpoints. Covers chunk route-mining, panel discovery via redirects, auth-gate probing with fake IDs, embedded-key verification (Firebase/OAuth), login rate-limit testing, common finding patterns, and non-destructive PoC discipline.
---

# Custom AI Router Assessment (Next.js-class self-hosted gateways)

Companion to `security/ai-api-gateway-security` (which covers New API/One API forks and is manually authored — do not expect to edit it; extend THIS skill instead when learning applies to custom-stack routers).

## 1. Fingerprint
| Signal | Meaning |
|---|---|
| `x-powered-by: Next.js`, `x-nextjs-cache: HIT`, `x-nextjs-prerender` | Next.js App Router behind CDN |
| Bundles at `/_next/static/chunks/<hash>.js` (Turbopack names, no `index.<hash>.js`) | route mining target |
| `/subscription` portal + Google Identity Services | customer commerce built-in (no separate topup.* storefront) |
| `/maintainer` → 307 `/maintainer/login` | separate admin panel, username/password auth |

**Panels hide from HTML links — discover them by following redirect targets.** Probe likely paths (`/admin`, `/dashboard`, `/login`, `/maintainer`) and read each 307's `Location`; that is how `/maintainer/login` surfaces even when no page links to it.

## 2. Route mining (Next.js variant)
```bash
curl -s https://TARGET/ | grep -oE '/_next/static/chunks/[a-zA-Z0-9_.-]+\.js'   # collect chunks
# download all, then per chunk:
grep -ohE '"(/api/[a-zA-Z0-9_/$ {}.-]+)"' chunk.js | sort -u                     # routes (~70 typical)
grep -n 'fetch("/api/' chunk.js                                                  # call sites → read ±300 chars for payload shapes
```
- Re-mine after fetching OTHER pages: the maintainer login page pulls extra chunks (the auth chunk holding `POST /api/auth/login {"username","password"}`).
- Template-literal routes appear as `/api/oauth/${a}/authorize` — normalize `${...}` to `{param}` when listing.

## 3. Auth-gate probing discipline (non-destructive)
- Write-shaped endpoints: send syntactically valid but **nonexistent IDs** (`{"orderId":"zz-nonexistent-probe-9x7q"}`) — proves the gate (401 vs processed) without creating artifacts.
- Destructive endpoints (`/api/version/shutdown`, reset/purge): **GET method-probe only**, never POST in production. 401 vs 405 answers the gate question.
- Expect and record uniform JSON errors: `{"error":"Unauthorized"}` (maintainer APIs) vs `{"error":"Subscription login required"}` (customer APIs) — two distinct auth realms worth mapping separately.

## 4. Embedded credentials — verification pitfalls
- Firebase web keys in bundles: validate via `POST https://identitytoolkit.googleapis.com/v1/accounts:createAuthUri?key=<KEY>` with body `{"continueUri":"https://TARGET","email":"probe@example.invalid"}`.
  - **MUST pass `?key=` as query param.** Omitting it returns a misleading 403 "unregistered callers" that looks like a restriction win.
  - `400 "API key not valid"` = revoked/dead → informational, NOT a finding.
- Google OAuth client IDs served publicly (e.g. via `/api/subscription/google-client-id`) are **by design public**; judge by redirect-URI allowlist, never report secrecy alone.

## 5. Login rate-limit & enumeration test
≤10 sequential fake-credential POSTs, ~0.4s apart. Good control looks like: 2×401 then `429` with `{"quarantined":true,"retryAfter":900}`. Quarantine usually also swallows follow-up username probes (uniform responses ⇒ no user enum). Record that your IP is now quarantined ~15 min — schedule other tests on different endpoints meanwhile.

## 6. Recurring finding patterns (checklist)
1. **Catalog dump unauth** — e.g. `GET /api/voucher-groups` returning hidden items (`isVisible:false`), prices, stock, sales stats, internal UUIDs. Severity Low; PoC = plain GET asserting hidden items present.
2. **Version disclosure** — unauth `/api/version`; also sanity-check feed logic (`latest < current` = bonus bug).
3. **Header gaps** — nosniff/XFO/referrer-policy often present; CSP and HSTS typically missing on Cloudflare-fronted Next.js.
4. **Sandbox endpoints shipped to prod** — e.g. `simulate-payment` (UI warns "sandbox only"), `shutdown`. Verify gate unauth; flag residual risk: if a low-role CUSTOMER session can call payment-simulation, escalate toward Critical (needs authenticated test account).
5. **No Cache-Control on API responses** — defense-in-depth note when responses lack any caching headers.

## 7. SPA-fallback false-positive elimination
Unknown paths may 307→portal or return the SPA HTML shell with 200/404. Before claiming "exposed file", verify content-type + size differ from baseline HTML (real config = application/json/text; fallback = text/html fixed size).

## 8. Report format (user preference)
Consolidated markdown per target: Executive Summary table (# / Finding / Severity / Status) up top; each finding = CVSS v3.1 full vector, raw request/response evidence, impact, remediation, paste-ready reproduction block. Include a "defenses confirmed working" section — operators value positives. Deliver as zip: SUMMARY.md + findings/ + pocs/ + evidence/.

## Pitfalls
- crt.sh JSON sometimes returns non-JSON (HTML challenge) — handle parse failure gracefully, don't conclude "no certs".
- `dig`/`whois`/`jq`/`zip` may be absent: DNS via DoH `curl 'https://dns.google/resolve?name=X&type=Y'`; archives via Python `zipfile`; JSON via `python3 -m json.tool`.
- Cloudflare-fronted targets: keep ≤10 req/s sequential during probing to avoid WAF 429 noise contaminating results.

## References
- `references/custom-router-case.md` — full case file: endpoint catalog, verified defenses, payload shapes, PoC recipes from a completed engagement.
