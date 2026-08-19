# WAF Product Bypass Matrix

> **AI LOAD INSTRUCTION**: Load this when you've identified the specific WAF product and need product-targeted bypass techniques. Assumes the main [SKILL.md](./SKILL.md) is already loaded for generic bypass methodology.

---

## 1. Cloudflare WAF

### Detection

- `cf-ray` header, `Server: cloudflare`, block page references "Cloudflare"
- Cookie: `__cfduid`, `__cf_bm`

### Known Bypass Techniques

| Category | Technique |
|---|---|
| Unicode normalization | Cloudflare normalizes Unicode differently than backend — `＜script＞` (fullwidth) may pass WAF but render as `<script>` |
| Chunked body | Split payloads across HTTP chunks; Cloudflare may not reassemble before inspection |
| Payload mutation (SQLi) | `/*!50000UniOn*/SeLeCt` — MySQL version comments bypass keyword matching |
| Payload mutation (XSS) | `<svg/onload=alert&#40;1&#41;>`, `<details open ontoggle=alert(1)>` |
| Origin direct access | Find origin IP via DNS history, Shodan `ssl.cert.subject.cn:target.com`, email headers |
| JSON body | Switch from form-urlencoded to JSON — different parser, weaker rules |
| Super-long parameter names | Parameter name >128 chars may cause Cloudflare to skip inspection |

### Cloudflare-Specific Notes

- Cloudflare has multiple WAF modes: "Managed Rules" (Cloudflare-authored) and "OWASP ModSecurity Core Rule Set". Each has different bypass surfaces.
- Cloudflare's free-tier WAF has significantly fewer rules than Business/Enterprise.
- Browser Integrity Check and Bot Management are separate from WAF — don't confuse them.

---

## 2. AWS WAF

### Detection

- `x-amzn-requestid` header, runs in front of ALB/CloudFront/API Gateway
- Block response often returns 403 with JSON body or custom error page

### Known Bypass Techniques

| Category | Technique |
|---|---|
| Regex complexity | AWS WAF regex rules have execution time limits — complex input can cause regex to timeout → request passes |
| Size limits | AWS WAF inspects first 8KB of body (16KB for CloudFront). Payload after this boundary is uninspected |
| Custom rule gaps | Default AWS Managed Rules miss many edge cases; custom rules often have logic errors |
| JSON depth | Deeply nested JSON objects may exceed parser depth limits |
| Base64 in parameters | AWS WAF doesn't auto-decode Base64 in parameter values (unless custom transform configured) |
| URI vs body rules | Rules may cover URI but not body, or vice versa — test both |

### AWS WAF-Specific Notes

- AWS WAF v2 (WAFV2) has `SizeConstraintStatement` — bodies over the size limit are either blocked or allowed, depending on config. If "allow on oversize", pad payload beyond 8KB.
- AWS Managed Rule Groups update regularly but lag behind novel attack patterns.
- IP reputation lists may be stale — new IPs from cloud providers often aren't listed.

---

## 3. ModSecurity + OWASP CRS

### Detection

- `Server: Apache` or `nginx` with ModSecurity module
- Block page: "ModSecurity" reference, or generic 403
- Error contains rule ID (e.g., `id:942100`)

### Known Bypass Techniques

| Category | Technique |
|---|---|
| Paranoia Level (PL) gaps | PL1 (default) has minimal rules; PL2-4 progressively stricter. Most deployments run PL1-2, missing many attack patterns |
| Rule ID specific bypass | Each rule targets specific patterns — identify blocking rule ID from error, craft bypass for that specific regex |
| SQL comment injection | `/*! ... */` MySQL conditional comments bypass many CRS SQLi rules |
| Unicode in PL1 | PL1 doesn't check Unicode-encoded payloads: `%u0027` for `'` |
| Transformation order | CRS applies `t:urlDecodeUni,t:htmlEntityDecode` but not all transformations on all rules |
| Multipart parser | CRS multipart parsing can be confused by malformed boundaries |
| Request body limit | `SecRequestBodyLimit` default is 13MB — but `SecRequestBodyNoFilesLimit` is only 128KB (changeable). Payloads in file upload fields bypass body rules if only `NoFiles` limit is enforced |

### CRS-Specific Notes

- CRS v4 (2023+) significantly improved coverage vs v3. Check target's CRS version.
- Anomaly scoring mode: individual rule violations add to score, blocked only if total exceeds threshold. Keep individual violations below detection but accumulate effect.
- `SecRuleRemoveById` directives in config may disable specific rules — test for holes.

---

## 4. Akamai (Kona Site Defender / App & API Protector)

### Detection

- `Server: AkamaiGHost`, `x-akamai-*` headers
- Error reference number in block page

### Known Bypass Techniques

| Category | Technique |
|---|---|
| Header injection | Akamai processes certain headers differently; `X-Forwarded-Host` injection can confuse routing |
| Encoding chains | Triple encoding or mixed encoding (URL + Unicode + HTML) |
| JSON body bypass | Akamai's JSON parser may not inspect deeply nested objects |
| Slow POST | Akamai has timeout-based protections; slow delivery may cause incomplete inspection |
| HTTP/2 push | H2 server push responses may bypass WAF inspection |
| IP rotation | Akamai rate limits per IP; rotating source IPs avoids behavioral blocks |

### Akamai-Specific Notes

- Akamai has "Adaptive Security Engine" — it learns application behavior. New attack patterns that don't match learned behavior may bypass initially.
- Penalty box: after triggering Akamai WAF, your IP may be rate-limited for minutes. Use fresh IP for each test.
- Akamai Pragma headers (`Pragma: akamai-x-check-cacheable`) can leak internal routing information useful for understanding the setup.

---

## 5. Imperva / Incapsula

### Detection

- `X-CDN: Imperva`, `Set-Cookie: incap_ses_*`, `visid_incap_*`
- Block page: "Powered by Incapsula" or Imperva branding

### Known Bypass Techniques

| Category | Technique |
|---|---|
| Parameter pollution | Duplicate parameters: Imperva inspects one occurrence, app processes another |
| JSON deep nesting | `{"a":{"b":{"c":{"d":"payload"}}}}` — deeply nested JSON exceeds parser depth |
| Multipart abuse | Malformed multipart boundaries confuse Imperva's parser |
| UTF-8 BOM injection | `\xEF\xBB\xBF` at start of body may shift parser alignment |
| Large Cookie header | Extremely long Cookie headers may cause truncated inspection |
| WebSocket upgrade | After WebSocket upgrade, subsequent traffic may bypass WAF inspection |

### Imperva-Specific Notes

- Imperva has "Client Classification" — browser fingerprinting. Headless browsers may be blocked before WAF rules even apply. Use real browser fingerprints.
- Imperva's API security module is separate from web WAF — API endpoints may have weaker protection.
- Custom rules in Imperva use "IncapRule" syntax — misconfigurations are common.

---

## 6. F5 BIG-IP ASM / Advanced WAF

### Detection

- `Server: BigIP`, `BIGipServer` cookie, `TS` cookie prefix
- Block page: "The requested URL was rejected" with support ID

### Known Bypass Techniques

| Category | Technique |
|---|---|
| Serialized format bypass | ASM has weak inspection of serialized data (Java, PHP, .NET serialization) |
| JSON/XML content switching | Switch between JSON and XML — ASM may have different rule sets per content type |
| Parameter meta-characters | ASM's "meta-character enforcement" can be bypassed with double encoding |
| Cookie manipulation | ASM sets tracking cookies; modifying them can cause session tracking issues that affect rule application |
| Evasion techniques | ASM has explicit "evasion detection" for directory traversal, multiple encoding, etc. But combinations of techniques may still bypass |
| Learning mode exploitation | If ASM is in "transparent" (learning) mode, no blocking occurs — test with obviously malicious payload first |

### F5-Specific Notes

- BIG-IP ASM distinguishes between "attack signatures" and "violations". Signatures are pattern-based; violations are structural (parameter length, data type). Both must be bypassed.
- ASM's "Bot Defense" module is separate and can be detected via JavaScript challenge injection.
- The `TS` cookie contains session data — tampering with it causes ASM to treat the request as a new session.

---

## 7. Sucuri WAF

### Detection

- `Server: Sucuri/Cloudproxy`, `X-Sucuri-ID` header
- Block page: "Access Denied - Sucuri Website Firewall"

### Known Bypass Techniques

| Category | Technique |
|---|---|
| Tag/event combos | Sucuri blocks common XSS tags but may miss: `<svg/onload>`, `<details/ontoggle>`, `<marquee onstart>` |
| SQL function alternatives | `MID()` instead of `SUBSTRING()`, `CONV()` for hex conversion |
| Path traversal encoding | `..%252f..%252f` (double URL encode) for directory traversal |
| Origin direct access | Sucuri is a reverse proxy; origin IP discovery bypasses it entirely |
| HTTP method switch | Sucuri may have different rules for GET vs POST vs PUT |
| Null byte injection | `%00` in parameter values may truncate Sucuri's inspection |

### Sucuri-Specific Notes

- Sucuri is common on WordPress sites — combine with WordPress-specific attack vectors.
- Sucuri's "Hardening" features (block PHP in uploads, etc.) are separate from WAF rules.
- Free Sucuri tier has significantly weaker WAF rules than paid tiers.

---

## 8. QUICK REFERENCE — BYPASS-BY-WAF CHEAT SHEET

| WAF | Top Bypass Vector | Size Limit | Key Weakness |
|---|---|---|---|
| Cloudflare | Unicode normalization + origin IP | 128KB | Fullwidth chars, free tier gaps |
| AWS WAF | Body size > 8KB | 8KB (body) | Size limit bypass, regex timeout |
| ModSecurity CRS | PL1 gaps + MySQL comments | Configurable | Low paranoia defaults |
| Akamai | Encoding chains + slow POST | Varies | Adaptive engine learning delay |
| Imperva | HPP + JSON nesting | Unknown | Parameter pollution |
| F5 BIG-IP | Serialized data + learning mode | Configurable | Weak serialization inspection |
| Sucuri | Origin IP + alt tags | Unknown | WordPress-centric rules |
