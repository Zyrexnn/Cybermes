---
name: subdomain-takeover
description: >-
  Subdomain takeover detection and exploitation playbook. Use when targets have
  dangling CNAME/NS/MX records pointing to deprovisioned cloud resources, expired
  third-party services, or unclaimed SaaS tenants that an attacker can register
  to serve content under the victim's domain.
---

# SKILL: Subdomain Takeover — Detection & Exploitation Playbook

> **AI LOAD INSTRUCTION**: Covers CNAME/NS/MX takeover, per-provider fingerprint matching, claim procedures, and defensive monitoring. Base models often confuse "CNAME exists" with "takeover possible" — the key is whether the *resource behind the CNAME is unclaimed and claimable*.

## 0. RELATED ROUTING

- [ssrf-server-side-request-forgery](../ssrf-server-side-request-forgery/SKILL.md) when a subdomain takeover is used to bypass SSRF allowlists trusting `*.target.com`
- [cors-cross-origin-misconfiguration](../cors-cross-origin-misconfiguration/SKILL.md) when CORS trusts `*.target.com` — takeover → full cross-origin read
- [xss-cross-site-scripting](../xss-cross-site-scripting/SKILL.md) takeover gives you script execution under target origin (cookie theft, OAuth redirect abuse)
- [http-host-header-attacks](../http-host-header-attacks/SKILL.md) when Host routing leads to subdomain-scoped cache or auth issues
- [web-cache-deception](../web-cache-deception/SKILL.md) when a taken-over subdomain shares cache with the main domain

---

## 1. CORE CONCEPT

Subdomain takeover occurs when:

1. `sub.target.com` has a DNS record (CNAME, NS, A) pointing to an external service
2. The external resource is **no longer provisioned** (deleted S3 bucket, removed Heroku app, etc.)
3. The attacker can **register/claim** that exact resource name on the provider
4. The attacker now controls content served under `sub.target.com`

**Impact**: cookie theft (parent domain cookies), OAuth token interception, phishing under trusted domain, CORS bypass, CSP bypass via whitelisted subdomain.

---

## 2. DETECTION METHODOLOGY

### 2.1 CNAME Enumeration

```
1. Collect subdomains (amass, subfinder, assetfinder, crt.sh, SecurityTrails)
2. Resolve DNS for each:
   dig CNAME sub.target.com +short
3. For each CNAME → check if the CNAME target returns NXDOMAIN or a provider error
4. Match error response against fingerprint table (Section 3)
```

### 2.2 Key Signals

| Signal | Meaning |
|---|---|
| CNAME → `xxx.s3.amazonaws.com` + HTTP 404 "NoSuchBucket" | S3 bucket deleted, claimable |
| CNAME → `xxx.herokuapp.com` + "No such app" | Heroku app deleted |
| CNAME → `xxx.github.io` + 404 "There isn't a GitHub Pages site here" | GitHub Pages unclaimed |
| NXDOMAIN on the CNAME target domain itself | Target domain expired or never existed |
| CNAME → provider but HTTP 200 with default parking page | May or may not be claimable — verify |

### 2.3 Automated Tools

| Tool | Purpose |
|---|---|
| `subjack` | Automated CNAME takeover checking |
| `nuclei -t takeovers/` | Nuclei takeover detection templates |
| `can-i-take-over-xyz` (GitHub) | Reference for which services are vulnerable |
| `dnsreaper` | Multi-provider takeover scanner |
| `subzy` | Fast subdomain takeover verification |

---

## 3. SERVICE PROVIDER FINGERPRINT TABLE

| Provider | CNAME Pattern | Fingerprint (HTTP Response) | Claimable? |
|---|---|---|---|
| **AWS S3** | `*.s3.amazonaws.com` / `*.s3-website-*.amazonaws.com` | `NoSuchBucket` (404) | Yes — create bucket with matching name |
| **GitHub Pages** | `*.github.io` | `There isn't a GitHub Pages site here` (404) | Yes — create repo + enable Pages |
| **Heroku** | `*.herokuapp.com` / `*.herokudns.com` | `No such app` | Yes — create app with matching name |
| **Azure** | `*.azurewebsites.net` / `*.cloudapp.azure.com` / `*.trafficmanager.net` | Various default pages, NXDOMAIN | Yes — register matching resource |
| **Shopify** | `*.myshopify.com` | `Sorry, this shop is currently unavailable` | Yes — create shop, add custom domain |
| **Fastly** | CNAME to Fastly edge | `Fastly error: unknown domain` | Yes — add domain to Fastly service |
| **Pantheon** | `*.pantheonsite.io` | `404 Site Not Found` with Pantheon branding | Yes |
| **Tumblr** | `*.tumblr.com` (custom domain CNAME) | `There's nothing here` / `Whatever you were looking for doesn't exist` | Yes |
| **WordPress.com** | CNAME to `*.wordpress.com` | `Do you want to register` | Yes — claim domain in WP.com |
| **Zendesk** | `*.zendesk.com` | `Help Center Closed` / Zendesk branding on error | Yes — create matching subdomain |
| **Unbounce** | `*.unbouncepages.com` | `The requested URL was not found` | Yes |
| **Ghost** | `*.ghost.io` | `404 Not Found` Ghost error | Yes |
| **Surge.sh** | `*.surge.sh` | `project not found` | Yes |
| **Fly.io** | CNAME to `*.fly.dev` | Fly.io default 404 | Yes |

---

## 4. TAKEOVER PROCEDURE — COMMON PROVIDERS

### 4.1 AWS S3

```
1. Confirm: curl -s http://sub.target.com → "NoSuchBucket"
2. Extract bucket name from CNAME (e.g., sub.target.com.s3.amazonaws.com → bucket = "sub.target.com")
3. aws s3 mb s3://sub.target.com --region <region>
4. Upload index.html proving control
5. Enable static website hosting
```

### 4.2 GitHub Pages

```
1. Confirm: curl -s https://sub.target.com → "There isn't a GitHub Pages site here"
2. Create GitHub repo (any name)
3. Add CNAME file containing "sub.target.com"
4. Enable GitHub Pages in repo settings
5. Wait for DNS propagation (GitHub verifies CNAME match)
```

### 4.3 Heroku

```
1. Confirm: curl -s http://sub.target.com → "No such app"
2. heroku create <app-name-from-cname>
3. heroku domains:add sub.target.com
4. Deploy proof-of-concept page
```

---

## 5. NS TAKEOVER — HIGH SEVERITY

NS takeover is **far more dangerous** than CNAME takeover: you control **all DNS resolution** for the zone.

### How It Happens

```
target.com NS → ns1.expireddomain.com
                 ↓
attacker registers expireddomain.com
                 ↓
attacker now controls ALL DNS for target.com
(A records, MX records, TXT records — everything)
```

### Detection

```
1. Enumerate NS records: dig NS target.com +short
2. Check each NS domain: whois ns1.example.com → is the domain expired or available?
3. Also check: dig A ns1.example.com → NXDOMAIN/SERVFAIL?
4. Subdelegated zones: check NS for sub.target.com specifically
```

### Impact

- Full domain takeover (serve any content, intercept email, issue TLS certs via DNS-01)
- Issue DV certificates from any CA using DNS challenge
- Modify SPF/DKIM/DMARC → send authenticated email as target

---

## 6. MX TAKEOVER — EMAIL INTERCEPTION

When MX records point to deprovisioned mail services:

```
target.com MX → mail.deadservice.com (service discontinued)
```

If attacker can claim `mail.deadservice.com` or the mail tenant:
- Receive password reset emails
- Intercept sensitive communications
- Potentially reset accounts that use email-based auth

### Common Scenario

Expired Google Workspace / Microsoft 365 tenant → MX still points to Google/Microsoft → attacker creates new tenant and claims the domain.

---

## 7. WILDCARD DNS RISKS

If `*.target.com` has a wildcard CNAME to a claimable service:
- **Every** undefined subdomain is vulnerable
- `anything.target.com` can be taken over
- Massively increases attack surface

Detection: `dig A random1234567.target.com` — if it resolves, wildcard exists.

---

## 8. DETECTION & EXPLOITATION DECISION TREE

```
Subdomain discovered (sub.target.com)?
├── Resolve DNS records
│   ├── Has CNAME → external service?
│   │   ├── HTTP response matches known fingerprint? (Section 3)
│   │   │   ├── YES → Attempt claim on provider (Section 4)
│   │   │   │   ├── Claim successful → TAKEOVER CONFIRMED
│   │   │   │   └── Claim blocked (name reserved, region locked) → document, try variations
│   │   │   └── NO → Service active, no takeover
│   │   └── CNAME target NXDOMAIN?
│   │       ├── Target is a registrable domain? → Register it → full control
│   │       └── Target is a subdomain of active provider → check provider claim process
│   │
│   ├── Has NS records → external nameserver?
│   │   ├── NS domain expired/available? → Register → FULL ZONE TAKEOVER
│   │   └── NS domain active → no takeover
│   │
│   ├── Has MX → external mail service?
│   │   ├── Mail service deprovisioned/claimable? → Claim tenant → EMAIL INTERCEPTION
│   │   └── Active mail service → no takeover
│   │
│   └── Has A record → IP address?
│       ├── IP belongs to elastic cloud (AWS EIP, Azure, GCP)?
│       │   ├── IP unassigned? → Claim IP → serve content
│       │   └── IP assigned to another customer → no takeover
│       └── IP belongs to dedicated server → no takeover
│
└── Post-takeover impact assessment
    ├── Shared cookies with parent domain? → Session hijacking
    ├── CORS trusts *.target.com? → Cross-origin data theft
    ├── CSP whitelists *.target.com? → XSS via taken-over subdomain
    ├── OAuth redirect_uri allows sub.target.com? → Token theft
    └── Can issue TLS cert for sub.target.com? → Full MITM
```

---

## 9. DEFENSE & REMEDIATION

| Action | Priority |
|---|---|
| Remove DNS records when deprovisioning cloud resources | Critical |
| Monitor CNAME targets for NXDOMAIN responses | High |
| Use DNS monitoring tools (SecurityTrails, DNSHistory) | High |
| Claim/reserve resource names before deleting DNS records | High |
| Audit NS delegations — ensure NS domains are owned and renewed | Critical |
| Avoid wildcard CNAMEs to third-party services | Medium |
| Implement Certificate Transparency monitoring | Medium |

---

## 10. TRICK NOTES — WHAT AI MODELS MISS

1. **CNAME ≠ takeover**: A CNAME to S3 that returns 403 (bucket exists, private) is NOT vulnerable. Only `NoSuchBucket` (404) is.
2. **Region matters for S3**: Bucket names are global, but website endpoints are regional. Try matching the region from the CNAME.
3. **GitHub Pages verification**: GitHub added domain verification — org-verified domains cannot be claimed by others. Check if target uses this.
4. **Edge cases**: Some providers (e.g., Cloudfront) require specific distribution configuration, not just domain claiming.
5. **Second-order takeover**: `sub.target.com CNAME → other.target.com CNAME → dead-service.com` — the chain must be followed fully.
6. **SPF subdomain takeover**: If SPF includes `include:sub.target.com` and you take over `sub.target.com`, you can modify its SPF TXT record to authorize your mail server → send spoofed email as `target.com`.
