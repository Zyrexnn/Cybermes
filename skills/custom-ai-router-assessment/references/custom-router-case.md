# Case file: Next.js-class Next.js AI router (ai-router.example.com, 2026-08)

Condensed session record backing SKILL.md. All data below came from live, non-destructive probing.

## Fingerprint (verified)
- Next.js App Router + Turbopack behind Cloudflare; `x-powered-by: Next.js`, `x-nextjs-cache: HIT`, double `x-nextjs-prerender`.
- Customer portal `/subscription` (Google Identity Services; client id served publicly by `GET /api/subscription/google-client-id` — by design).
- Admin panel `/maintainer/login` → `POST /api/auth/login {"username","password"}`. Discovered via 307 `Location: /maintainer/login` from `/maintainer`, NOT from HTML links.
- ~71 API routes recoverable from bundles alone.

## Route-mining recipe (Next.js variant)
```bash
curl -s https://TARGET/ | grep -oE '/_next/static/chunks/[a-zA-Z0-9_.-]+\.js'   # collect chunks
# download each, then:
grep -ohE '"(/api/[a-zA-Z0-9_/$ {}.-]+)"' chunk.js | sort -u                     # mine routes
grep -n 'fetch("/api/' chunk.js                                                  # call sites → read ±300 chars for payload shapes
```
Login-page HTML references extra chunks (e.g. the auth chunk with `/api/auth/login`) — re-mine after fetching `/maintainer/login`.

## Endpoint catalog (observed)
- Unauth-readable: `/api/auth/status` (`requireLogin`, `authMode`, OIDC flags), `/api/version`, `/api/voucher-groups`, `/api/subscription/google-client-id`.
- Auth-gated (verified 401): `/api/settings`, `/api/models*`, `/api/providers*`, `/api/provider-nodes*`, `/api/combos`, `/api/public/telegram`, `/api/cli-tools/*`, `/api/support-users/*`, `/api/usage/stream`, `/api/version/shutdown` (GET probe only), `/api/subscription/{status,models,redemptions,referral,support,simulate-payment,create-order,redeem}`.
- OpenAI-compat `/v1/*`: dummy Bearer → 401 `{"error":{"message":"API key required for remote API access","code":"api_key_required"}}`.
- Two auth realms with distinct error strings: maintainer APIs say `"Unauthorized"`; customer subscription APIs say `"Subscription login required"`.

## Payload shapes (from bundle call sites)
- Create order: `POST /api/subscription/create-order {"groupId","quantity","requestId":"<crypto.randomUUID()>","useWallet":false,"paymentMethod":"qris"}`.
- Simulate payment: `POST /api/subscription/simulate-payment {"orderId"}` (UI warns "sandbox only").
- Google login: `POST /api/subscription/google-login {"credential","referralCode"?}` (referral code from sessionStorage `ref_code`).
- Maintainer login: `POST /api/auth/login {"username","password"}`; response may carry `mustChangePassword` or `{"quarantined":true,"redirectTo":"/subscription"}`.

## Findings from this engagement
1. **LOW — `/api/voucher-groups` unauth catalog dump**: ALL groups incl. `isVisible:false` (hidden/draft products: "Qoder Pro…", "100 AKUN GROK", "BOT GROK UNLIMITED V2"), prices Rp2.5k–35k, `soldCount`, stock counts, internal UUIDs. Fix: server-side filter + strip internal stats. PoC = plain GET asserting hidden items present.
2. **INFO — `/api/version` unauth** leaks build version; feed anomaly (`latestVersion:"0.5.55" < currentVersion:"0.7.6"`).
3. **INFO — headers**: nosniff/XFO/referrer-policy present; NO CSP, NO HSTS, no Cache-Control on API JSON.
4. **INFO — sandbox endpoints in prod**: `simulate-payment` + `version/shutdown`; both 401 unauth. Residual risk to check with a real low-role account: customer calling simulate-payment ⇒ Critical payment bypass.
5. **INFO — embedded keys**: Firebase key dead (400 "API key not valid"); OAuth client IDs public by design.

## Defenses confirmed working
- Login brute-force control: 2 failed POSTs → `429 {"quarantined":true,"retryAfter":900,"redirectTo":"/subscription"}`; quarantine also swallowed username-enumeration probes (uniform responses).
- Consistent 401 JSON on every sensitive route; no CORS origin reflection; unknown paths 307→`/subscription` or 404 SPA shell (no file leaks); TLS 1.3 AES_256_GCM; HTTP→HTTPS 301; nuclei clean (only WAF detect).

## Embedded-key verification pitfalls
- Validate Firebase web keys with `POST https://identitytoolkit.googleapis.com/v1/accounts:createAuthUri?key=<KEY>` body `{"continueUri":"https://TARGET","email":"probe@example.invalid"}`.
  - **MUST pass `?key=` as query param** — omitting it yields misleading 403 "unregistered callers".
  - `400 "API key not valid"` = revoked/dead → informational only.

## Discipline notes
- Probe write-shaped endpoints with syntactically valid but nonexistent IDs to prove gates without artifacts.
- Never POST shutdown/reset endpoints in production; GET method-probe suffices (401 vs 405).
- Rate-limit tests: ≤10 sequential fake-cred logins, ~0.4s spacing; expect uniform 401s (bad) or 429/quarantine (good).
