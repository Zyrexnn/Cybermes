---
name: web-auth-defenses-testing
description: Methodology for testing authentication defenses on web apps — login lockout/rate-limit scope determination (global vs per-account vs per-IP), forgot-password error handling, auth-route method/logic anomalies, and hosted-app (Next.js/Vercel) auth probing. Produces correctly-scored findings (Critical vs Medium) and safe, non-destructive testing.
---

# Web Auth Defenses Testing

## Purpose
Test login/account-defense controls the way a bug-bounty reviewer scores them: determine *scope*
of any lockout, find logic anomalies, and confirm (or refute) hosted-app auth weaknesses —
without locking real users or sustaining an outage.

## When to use
- Target exposes `POST /api/auth/{login,register,forgot-password,mfa/verify}` or similar.
- You see a "too many attempts" / lockout / rate-limit response during auth testing.
- You're assessing a Next.js/Vercel or other hosted-SPA app with separate auth + backend services.
- Companion to `web-pentest` (this skill drills the auth-defense sub-area in depth).

## Phase 1 — Lockout scope determination (MOST IMPORTANT, drives severity)

A failed-login lockout is keyed one of three ways; impact differs by ~10x:
- **per-account**  → only the targeted victim denied (Medium/High)
- **per-IP**       → only attacker's source throttled (Low)
- **GLOBAL/shared**→ entire site cannot log in (Critical, CVSS ~9.1, full auth outage)

### Probe recipe (run in order, NEVER use real users' emails)
```bash
# (a) ~6 bad logins to a RANDOM fresh email:
for i in $(seq 1 6); do
  curl -s -X POST https://TARGET/api/auth/login -H "Content-Type: application/json" \
    -d "{\"email\":\"locktest_$i@randomexample.com\",\"password\":\"wrongpass\"}"
done

# (b) test a DIFFERENT fresh email once:
curl -s -X POST https://TARGET/api/auth/login -H "Content-Type: application/json" \
  -d '{"email":"other_fresh@randomexample.com","password":"wrongpass"}'
#  "N attempts remaining" -> per-EMAIL (per-account)
#  "locked" + SAME resetTime -> GLOBAL

# (c) DEFINITIVE: brand-new email never contacted before:
curl -s -X POST https://TARGET/api/auth/login -H "Content-Type: application/json" \
  -d "{\"email\":\"neverseen_$(date +%s)@fresh.io\",\"password\":\"x\"}"
#  locked w/ identical resetTime as (a) -> GLOBAL shared counter CONFIRMED.
```
Capture `resetTime` each time; identical timestamps across unrelated emails prove a shared
counter. Compare to `date -u` for lock-window duration. **Always run (c) before scoring.**

## Phase 2 — Operational safety (do not cause harm)
- Use only random never-seen emails as lock targets. Do NOT probe guessed real addresses
  (admin@, support@, info@) — you may lock real users and trigger a real global outage.
- **Stop sending login requests once the lockout is confirmed.** A global lock may not clear
  within a short window; continued probing *sustains* the outage. Confirm, then cease.
- If a real-looking address was accidentally locked, note it as impact evidence, then stop.
- Registering a test account to reach authenticated areas may fail while a global lock is active
  — wait it out or use a fresh source rather than hammering.

## Phase 3 — Error-handling & enumeration differentials
- Invalid email *format* vs valid *format* vs valid + wrong password often yield different
  messages (`400 "Invalid email"` vs `401 "Invalid credentials. N remaining"`). Low (enum aid).
- `/forgot-password` returning `500` for every valid-format input (while `400` for bad format)
  = broken reset flow / weak error handling (Medium). Record the exact body.

## Phase 4 — Auth-route method / logic anomalies
- After `OPTIONS` on an auth route, read `Allow:` (e.g. `DELETE, OPTIONS, POST`). Test surprises:
  ```bash
  curl -s -X DELETE https://TARGET/api/auth/login -H "Content-Type: application/json" -d '{}'
  # 200 {"success":true} without auth = logic anomaly (Medium)
  ```
- `PUT`/`TRACE` should be `405`. Method-override headers should be ignored.

## Phase 5 — Hosted-app (Next.js/Vercel) auth probing
- `httpx -tech-detect` → `Vercel`, `HSTS`, sometimes `C3.js`; apex IP is Vercel anycast.
- Protected routes: expect `307 → /login` (GET) and `/api/dashboard/*` → `401` without session.
  That is access control WORKING — record as confirmed non-issue, not a finding.
- Source maps `<chunk>.js.map` → usually `403` (safe). Hidden `.env`/`.git`/`.svn` → `403/404`.
- CORS `Access-Control-Allow-Origin: *` is LOW unless `Access-Control-Allow-Credentials: true`
  also present.
- Missing `X-Frame-Options` / `CSP frame-ancestors` / `X-Content-Type-Options` / `Referrer-Policy`
  → clickjacking Low finding.
- Mass-assignment: send `role":"admin"` / `isAdmin":true` in register; verify server echoes
  `role":"user"` (rejected) vs honors it.

## Phase 6 — Separate backend gateway discovery
- `/docs` often embeds a real backend URL in sample `curl` (e.g. `https://<svc>.up.railway.app`).
  Probe separately: `$B/health` (leaks service name/mode/uptime), `$B/v1/chat/completions` (401
  with distinct no-key vs bad-key message). Flag CORS `*` only if a valid key can reach a browser
  context; if frontend proxies server-side, it's low-risk. Info/Low.

## Reporting
- Score lockout by scope: GLOBAL → Critical; per-account → High/Medium; per-IP → Low.
- Always include the resetTime comparison as evidence for a global finding.
- Provide a self-contained `poc.py` that reproduces with a random email and stops after confirming.

## References
- `references/lockout_and_hosted_app_testing.md` — full probe recipes, curl snippets, and
  tooling notes for lockout-scope determination, hosted-app specifics, and backend discovery.
