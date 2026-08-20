# Lockout-Scope Determination & Hosted-App Probing Recipes

Reusable recipes distilled from a real bug-bounty engagement against a Next.js (App Router) +
Vercel app with a separate Express backend. Use with the `web-auth-defenses-testing` skill.

## 1. Lockout-scope determination (global vs per-account vs per-IP)

Keyed three ways; impact differs by an order of magnitude:
- per-account   -> only the targeted victim denied (Medium/High)
- per-IP        -> only attacker's source throttled (Low)
- GLOBAL/shared -> entire site cannot log in (Critical — full auth outage)

### Probe recipe (run in order, NEVER use real users' emails)
```
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
counter. Compare to `date -u` for lock-window duration. ALWAYS run (c) before scoring.

## 2. Operational safety
- Only random never-seen emails as lock targets. Do NOT probe guessed real addresses.
- Stop sending login requests once lockout is confirmed — a global lock may not clear quickly;
  continued probing sustains the outage. Confirm, then cease.
- If a real-looking address was accidentally locked, note it as impact evidence, then stop.

## 3. Next.js / Vercel auth probing
- `httpx -tech-detect` -> `Vercel`, `HSTS`, sometimes `C3.js`; apex IP is Vercel anycast.
- Protected routes: `307 -> /login` (GET), `/api/dashboard/*` -> `401` without session = access
  control WORKING (record as non-issue, not a finding).
- Source maps `<chunk>.js.map` -> usually `403` (safe). Hidden `.env`/`.git`/`.svn` -> `403/404`.
- CORS `*` is LOW unless `Access-Control-Allow-Credentials: true` also present.
- Missing `X-Frame-Options` / `CSP frame-ancestors` / `X-Content-Type-Options` / `Referrer-Policy`
  -> clickjacking Low finding.
- Mass-assignment: send `role":"admin"` / `isAdmin":true` in register; verify server echoes
  `role":"user"` (rejected) vs honors it.

## 4. Auth-route method / logic anomalies
```
curl -s -X DELETE https://TARGET/api/auth/login -H "Content-Type: application/json" -d '{}'
# 200 {"success":true} without auth = logic anomaly (Medium)
```
- `PUT`/`TRACE` should be `405`. Method-override headers should be ignored.
- `/forgot-password` `500` for every valid-format input (while `400` for bad format) = broken
  reset flow / weak error handling (Medium).

## 5. Separate backend gateway discovery
- `/docs` often embeds a real backend URL in sample `curl` (e.g. `https://<svc>.up.railway.app`).
  Probe: `$B/health` (leaks service name/mode/uptime), `$B/v1/chat/completions` (401 with
  distinct no-key vs bad-key message). Flag CORS `*` only if a valid key can reach a browser
  context; server-side proxy -> low-risk. Info/Low.

## 6. Tooling notes (verify in your env — versions differ)
- `katana`: stray boolean flags like `-kf` error unless given a value (`all|robotstxt|sitemapxml`).
- `dalfox url`: verify list-flag syntax with `dalfox url --help` before batching.
- If a report aggregator expects `/workspace/reports/<target>/` but you wrote to `/root/reports/`,
  copy findings into the expected path before running it.
