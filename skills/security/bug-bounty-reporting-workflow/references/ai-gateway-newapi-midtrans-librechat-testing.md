# AI Gateway (New API / One API fork) + Midtrans Storefront + LibreChat — Testing Playbook

Condensed, reusable technique bank from a full authorized non-destructive pentest of
`router.juan.web.id` (New API fork, Cloudflare-fronted) + `topup.juan.web.id` (Midtrans
storefront) + `chat.juan.web.id` (LibreChat). All probed at low sequential rate, zero WAF blocks.

## 1. New API / One API fork (unified AI API gateway)

**Fingerprint:** SPA React, `<title>New API</title>`, `server: cloudflare`, body "Unified AI
API gateway". JS bundles under `/static/js/index.*.js` (often 3+ MB). The admin UI and `/v1`
OpenAI-compatible proxy live on the SAME origin.

**Endpoint mining (do FIRST, passive, WAF-safe):**
```bash
curl -s https://TARGET/static/js/index.*.js -o /tmp/index.js
grep -oE '"/api/[a-zA-Z0-9_/{}.-]+"' /tmp/index.js | sort -u
# minified -> read_file with offset/limit, or awk a window:
awk '/api\/setup/{flag=1} flag{print} flag&&/}/{c++} c>2{print "----"; exit}' /tmp/index.js
```
Yielded 60+ routes: `/api/status`, `/api/setup`, `/api/user/self`, `/api/user/manage`,
`/api/user/token`, `/api/channel/*`, `/api/option/*`, `/api/models`, `/api/notice`, etc.

**Auth scheme split (common gotcha):** TWO credentials.
- `access_token` (JWT, ~392 chars) from `POST /api/user/login` → `Authorization: Bearer` for `/api/*`.
- `/v1` (OpenAI-compatible) needs an **API key**, NOT the access_token (returns `Invalid token`).
  Keys are generated via the user Token page; `/api/user/token` GET returns a regenerating
  token-like string per call that is NOT directly usable as a `/v1` Bearer.

**Vuln classes & tests (all confirmed safe on tested instance — reuse the patterns):**
| Vector | Test | Expected-secure result |
|--------|------|------------------------|
| Unauth admin setup | `GET /api/setup` (note `root_init:false`), then `POST /api/setup {"username","password"}` | GET may report `root_init:false` (cosmetic); POST returns "system already initialized", no account. **Classic One API admin-creation bug — always try the POST.** |
| IDOR read other users | `GET /api/user/<other_id>` with regular token | `403 AUTH_INSUFFICIENT_PRIVILEGE` |
| Privilege escalation | `PUT /api/user/self {"role":100,"group":"admin","quota":...}` | `success:true` but server ignores privileged fields; re-read shows `role:1` |
| Info disclosure | `GET /api/status` (no auth) | Returns config: oauth client_ids, telegram bot name, passkey RP, pricing, `register_enabled`. **Report Medium.** |
| Rate limiting | 6 rapid `POST /api/user/login` wrong creds | If all `200` + no `429`/`Retry-After`/lockout → **Medium** (brute-force/Credential Stuffing). Test early. |
| SPA fallback false-positive | `/.env`, `/config.json`, `/swagger`, `/debug`, `/.git/config`, `/metrics`, `/healthz`, `/version` | All return identical `index.html` (`content-type: text/html`, ~8431 bytes). Confirm via `content-type` + byte size. **Not a vuln.** |

## 2. Midtrans-backed custom storefront (Node/Express)

Usually a separate subdomain (here `topup.juan.web.id`). Single `app.js` bundle (~30KB).
Mine: `grep -oE "fetch\('/api/[a-zA-Z0-9_/-]+'" app.js`.
Endpoints: `GET /api/skus`, `POST /api/coupon/validate`, `GET /api/user-check?username=`,
`POST /api/order`, `GET /api/order/:token`, `POST /api/order/:token/bind`.

**Order/payment flow (read function context in app.js before testing):**
```
POST /api/order {sku,name,email,wa,username?,coupon_code?}
  -> {token, snap_token (Midtrans), amount_rp, fee_rp, ...}   # prices SERVER-computed from sku
User pays via Midtrans snap -> on success client calls:
POST /api/order/:token/bind {transaction_id, transaction_status, fraud_status, payment_type}
```

**Non-destructive test vectors:**
- **Payment bypass:** create order (unpaid, `status:pending`), then
  `POST /api/order/:token/bind {"transaction_status":"settlement",...}` FAKE.
  Secure: order stays `status:pending, bound:false` — server re-verifies with Midtrans. **No finding.**
- **Price/quota tampering:** `amount_rp`/`fee_rp`/`quota` come from server via `sku`; client only
  sends `sku`. Cannot tamper. **No finding.**
- **IDOR topup to victim:** `POST /api/order` with subs sku + fake `username` →
  rejected `"username tidak ditemukan"` (server-side). **No finding.**
- **Coupon/redeem:** `POST /api/coupon/validate {"code",...}` needs valid issuer code; common
  guesses all "not found". Brute-forcing requires rate-limit testing. **No finding** unless a real code leaks.
- **Catalog leak:** `GET /api/skus` (no auth) discloses products/prices/quotas → **Low** (public storefront).

## 3. LibreChat (chat GUI)

**Fingerprint:** `chat.*` subdomain, `/api/config` returns `{"appTitle":"LibreChat",...}`.
Auth paths: `/api/auth/register`, `/api/auth/login`, `/api/me`, `/api/user`.
Protected routes return `401 No auth token`.

**Tests:**
- `GET /api/config` (no auth) → discloses `registrationEnabled`, login methods, build `commit`. **Low** (LibreChat exposes by design).
- `/api/auth/register` exists; requires email verification + **rate-limited**
  ("Too many accounts created, please try again after 60 minutes"). Open-but-verified registration
  is NOT a直接 finding. Deep authed testing needs a verified account.
- Unauth file read: `/api/files`, `/api/messages`, `/api/conversations` → `401`. Static traversal
  (`/assets/../../etc/passwd`) → SPA fallback HTML, not real file. **No finding.**

## 4. Generic lessons reinforced
- **Cloudflare-fronted + low-rate sequential curl = safe.** No block observed on SIN edge. Stay
  sequential, ~1 req per endpoint.
- **"No PoC, no finding":** test every suspected critical (unauth admin, payment bypass, IDOR)
  with a real request and prove it safe — report the negative result too (shows coverage).
- **Eliminate SPA fallback false-positives** via `content-type` + byte-size, never by status code alone.
