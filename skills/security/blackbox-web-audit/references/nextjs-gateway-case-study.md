# Case Study: ai-gateway.example.com (Next.js AI Gateway) — 2026-08-21

Authorized audit by operator request. Black-box unauth, fully non-destructive. Report (Bahasa Indonesia): `/workspace/reports/example_gateway_com/`.

## Target Profile
- Next.js App Router (Turbopack) behind Cloudflare (WAF active, TLS 1.3).
- Dual panels: `/subscription` (customer, Google sign-in) + `/maintainer/login` (username/password → `POST /api/auth/login` JSON `{username,password}`).
- Attack surface almost entirely = JS bundles + API: subfinder 0 subs, gau 1 URL. Bundle mining yielded **71 API endpoints** (highest-yield step again).
- Notable API surface: `/api/voucher-groups`, `/api/vouchers*`, `/api/orders`, `/api/subscription/simulate-payment`, `/api/version/shutdown`, `/api/settings`, `/api/provider-nodes`, `/v1/*` (OpenAI-compatible), `/api/auth/status`, `/api/health`.

## Confirmed Findings (1 Low + 5 Info)
1. **LOW — catalog disclosure:** `GET /api/voucher-groups` AND per-ID variant `GET /api/voucher-groups/{uuid}` open without auth. All groups were `isVisible:false` (hidden from UI but shipped by API) + internal fields: `soldCount`, `quantity`, `wholesaleRules`, `groupingIds`. `?includeHidden=true` accepted too — the visibility filter is UI-side only, absent server-side.
2. INFO — `/api/auth/status` leaks auth config pre-login (`authMode:"password"`, `oidcConfigured:true`).
3. INFO — `/api/version` discloses app version (+ anomalous feed where latest < current).
4. INFO — no CSP, no HSTS (has X-Frame-Options/nosniff/referrer-policy).
5. INFO — sandbox endpoints shipped to prod (`simulate-payment`, `version/shutdown`) — verified auth-gated; GET-only probe on shutdown, POST deliberately never sent on prod.
6. INFO — embedded Firebase/OAuth creds in bundle → tested live against `identitytoolkit.googleapis.com/v1/accounts:createAuthUri?key=` → **400 "API key not valid"** (revoked). Verify before reporting any embedded key.

## IDOR Round Without Test Accounts
- Harvested real UUIDs from the public catalog; replayed GET-only against every ID-bearing variant: `/api/vouchers/{id}`, `/api/vouchers?groupId=`, `/api/orders?groupId=`, plus `/api/v4/user`, `/api/provider-nodes`, `/api/settings`.
- Controls: fake UUID → clean 404 `{"error":"Not found"}`; valid-ID-no-auth on gated endpoints → uniform 401. Verdict: object-level authorization solid, NO IDOR; negative matrix written into SUMMARY.md as coverage evidence.
- Complementary negatives: 0 canary reflection across pages; `?next=` fixed to `/subscription` (no open redirect); SQLi/traversal probes on UUID path → uniform 404 (parameterized); dalfox clean.

## Defenses Verified Working
- Login rate-limit: 2 fails → `429 {"error":"Too many failed attempts","retryAfter":900}` quarantine (~15 min). Probe IP got quarantined — proof of control.
- ~30 sensitive endpoints uniform 401; `/v1/*` rejects dummy Bearer (`api_key_required`); no CORS reflection; no `.env`/`.git`/`package.json` exposure; HTTP→HTTPS 301.

## Environment Pitfalls Hit
- **Cloudflare blocked python-urllib default UA with 403** while curl passed → all Python probes/PoCs need a browser User-Agent header.
- SPA fallback: unknown `/api/*` paths return HTTP 200 `text/html` (Next.js index) — always verify content-type or every endpoint check false-positives.
- crt.sh returned malformed 150-byte response → worked around via JS mining; no `dig/host/nslookup` on box → DNS-over-HTTPS via curl (`dns.google/resolve`).

## Non-Destructive Decisions That Held
- POST create-order probed ONLY with fake UUID; shutdown probed GET-only.
- Rate-limit test used fake credentials throughout; resulting self-quarantine documented as control evidence.

## Takeaway
Hardened target: 0 Critical/High/Medium. Single actionable fix = server-side visibility filter in voucher-group serialization. Best follow-up WITH test accounts: whether customer-role sessions can call `simulate-payment` (if yes → critical payment-bypass chain).
