---
name: anti-debugging-techniques
description: >-
  Anti-debugging detection and bypass playbook. Use when reversing protected
  binaries that detect debuggers via ptrace, PEB flags, timing checks, or
  signal/exception handlers on Linux and Windows.
---

# SKILL: Anti-Debugging Techniques — Detection & Bypass Playbook

> **AI LOAD INSTRUCTION**: Expert anti-debug techniques across Linux and Windows. Covers ptrace, PEB flags, NtQueryInformationProcess, timing attacks, signal-based detection, TLS callbacks, VEH tricks, and all corresponding bypass methods. Base models often miss the distinction between user-mode and kernel-mode detection and the correct patching strategy for each.

## 0. RELATED ROUTING

- [code-obfuscation-deobfuscation](../code-obfuscation-deobfuscation/SKILL.md) when the binary also uses control flow flattening, VM protection, or string encryption
- [vm-and-bytecode-reverse](../vm-and-bytecode-reverse/SKILL.md) when the anti-debug sits inside a custom VM dispatcher
- [symbolic-execution-tools](../symbolic-execution-tools/SKILL.md) when you want to symbolically skip anti-debug checks entirely

### Advanced Reference

Also load [ANTI_DEBUG_MATRIX.md](./ANTI_DEBUG_MATRIX.md) when you need:
- Complete cross-reference matrix of technique × OS × detection method × bypass method
- Per-technique reliability ratings and false-positive notes
- Tool compatibility chart (GDB, x64dbg, WinDbg, Frida, ScyllaHide)

### Quick bypass picks

| Detection Class | First Bypass | Backup |
|---|---|---|
| ptrace-based (Linux) | `LD_PRELOAD` hook `ptrace()` → return 0 | Kernel module to hide tracer |
| PEB.BeingDebugged (Windows) | Patch PEB byte at `fs:[0x30]+0x2` | ScyllaHide auto-patch |
| Timing check (rdtsc) | Conditional BP after rdtsc, fix registers | Frida hook `rdtsc` return |
| IsDebuggerPresent | NOP the call / hook return 0 | x64dbg built-in hide |
| INT 2D / UD2 exception | Set VEH to handle gracefully | TitanHide driver |

---

## 1. LINUX ANTI-DEBUG TECHNIQUES

### 1.1 ptrace(PTRACE_TRACEME)

The classic self-attach: a process calls `ptrace(PTRACE_TRACEME, 0, 0, 0)`. If a debugger is already attached, the call fails (returns -1).

```c
if (ptrace(PTRACE_TRACEME, 0, 0, 0) == -1) {
    exit(1); // debugger detected
}
```

**Bypass methods**:

| Method | How |
|---|---|
| `LD_PRELOAD` shim | Compile shared lib: `long ptrace(int r, ...) { return 0; }` and set `LD_PRELOAD` |
| Binary patch | NOP the `ptrace` call or patch return value check |
| GDB catch | `catch syscall ptrace` → modify `$rax` to 0 on return |
| Kernel module | Hook `sys_ptrace` to allow multiple tracers |

### 1.2 /proc/self/status — TracerPid

```c
FILE *f = fopen("/proc/self/status", "r");
// parse TracerPid: if non-zero → debugger attached
```

**Bypass**: Mount a FUSE filesystem over `/proc/self`, or `LD_PRELOAD` hook `fopen`/`fread` to filter `TracerPid` to 0.

### 1.3 Timing Checks (rdtsc / clock_gettime)

Measures elapsed time between two points; debugger single-stepping causes noticeable delay.

```asm
rdtsc
mov ebx, eax       ; save low 32 bits
; ... protected code ...
rdtsc
sub eax, ebx
cmp eax, 0x1000    ; threshold
ja  debugger_detected
```

**Bypass**: Set hardware breakpoint after second `rdtsc`, modify `eax` to pass the comparison. Or use Frida to replace the timing function.

### 1.4 Signal-Based Detection (SIGTRAP)

```c
volatile int caught = 0;
void handler(int sig) { caught = 1; }
signal(SIGTRAP, handler);
raise(SIGTRAP);
if (!caught) exit(1); // debugger swallowed the signal
```

When a debugger is attached, `SIGTRAP` is consumed by the debugger rather than delivered to the handler. **Bypass**: In GDB, use `handle SIGTRAP nostop pass` to forward the signal.

### 1.5 /proc/self/maps & LD_PRELOAD Detection

Checks for injected libraries or memory regions characteristic of debuggers/instrumentation.

```c
FILE *f = fopen("/proc/self/maps", "r");
while (fgets(buf, sizeof(buf), f)) {
    if (strstr(buf, "frida") || strstr(buf, "LD_PRELOAD"))
        exit(1);
}
```

**Bypass**: Hook `fopen("/proc/self/maps")` to return a filtered version, or rename Frida's agent library.

### 1.6 Environment Variable Checks

Some protections check for `LD_PRELOAD`, `LINES`, `COLUMNS` (set by GDB's terminal), or debugger-specific env vars.

**Bypass**: Unset suspicious env vars before launch, or hook `getenv()`.

---

## 2. WINDOWS ANTI-DEBUG TECHNIQUES

### 2.1 IsDebuggerPresent / CheckRemoteDebuggerPresent

```c
if (IsDebuggerPresent()) ExitProcess(1);

BOOL debugged = FALSE;
CheckRemoteDebuggerPresent(GetCurrentProcess(), &debugged);
if (debugged) ExitProcess(1);
```

**Bypass**: Hook `kernel32!IsDebuggerPresent` to return 0, or patch PEB directly.

### 2.2 PEB Flags

| Field | Offset (x64) | Debugged Value | Normal Value |
|---|---|---|---|
| `BeingDebugged` | `PEB+0x02` | 1 | 0 |
| `NtGlobalFlag` | `PEB+0xBC` | `0x70` (FLG_HEAP_*) | 0 |
| `ProcessHeap.Flags` | Heap+0x40 | `0x40000062` | `0x00000002` |
| `ProcessHeap.ForceFlags` | Heap+0x44 | `0x40000060` | 0 |

```asm
mov rax, gs:[0x60]    ; PEB
movzx eax, byte [rax+0x02]  ; BeingDebugged
test eax, eax
jnz debugger_detected
```

**Bypass**: Zero all four fields. ScyllaHide does this automatically.

### 2.3 NtQueryInformationProcess

| InfoClass | Value | Debugged Return |
|---|---|---|
| `ProcessDebugPort` | 0x07 | Non-zero port |
| `ProcessDebugObjectHandle` | 0x1E | Valid handle |
| `ProcessDebugFlags` | 0x1F | 0 (inverted!) |

**Bypass**: Hook `ntdll!NtQueryInformationProcess` to return clean values per info class.

### 2.4 Hardware Breakpoint Detection

```c
CONTEXT ctx;
ctx.ContextFlags = CONTEXT_DEBUG_REGISTERS;
GetThreadContext(GetCurrentThread(), &ctx);
if (ctx.Dr0 || ctx.Dr1 || ctx.Dr2 || ctx.Dr3)
    ExitProcess(1);
```

**Bypass**: Hook `GetThreadContext` to zero DR0–DR3, or use `NtSetInformationThread(ThreadHideFromDebugger)` preemptively (ironically, the anti-debug technique itself).

### 2.5 INT 2D / INT 3 / UD2 Exception Tricks

`INT 2D` is the kernel debug service interrupt. Without a debugger, it raises `STATUS_BREAKPOINT`; with a debugger, behavior differs (byte skipping).

```asm
xor eax, eax
int 2dh
nop          ; debugger may skip this byte
; ... divergent execution path ...
```

**Bypass**: Handle in VEH or patch the interrupt instruction.

### 2.6 TLS Callbacks

TLS callbacks execute before `main()` / `WinMain()`. Anti-debug checks placed here run before the debugger's initial break.

**Bypass**: In x64dbg, set "Break on TLS Callbacks" option. In WinDbg, use `sxe ld` to break on module load.

### 2.7 NtSetInformationThread(ThreadHideFromDebugger)

```c
NtSetInformationThread(GetCurrentThread(), ThreadHideFromDebugger, NULL, 0);
```

After this call, the thread becomes invisible to the debugger — breakpoints and single-stepping stop working silently.

**Bypass**: Hook `NtSetInformationThread` to NOP when `ThreadInfoClass == 0x11`.

### 2.8 VEH-Based Detection

Registers a Vectored Exception Handler that checks `EXCEPTION_RECORD` for debugger-specific behavior (single-step flag, guard page violations with debugger semantics).

**Bypass**: Understand the VEH logic and ensure the exception chain behaves identically to non-debugged execution.

---

## 3. ADVANCED MULTI-LAYER TECHNIQUES

### 3.1 Self-Debugging (fork + ptrace)

The process forks a child that attaches to the parent via ptrace. If an external debugger is already attached, the child's ptrace fails.

```c
pid_t child = fork();
if (child == 0) {
    if (ptrace(PTRACE_ATTACH, getppid(), 0, 0) == -1)
        kill(getppid(), SIGKILL);
    else
        ptrace(PTRACE_DETACH, getppid(), 0, 0);
    _exit(0);
}
wait(NULL);
```

**Bypass**: Patch the `fork()` return or kill/detach the watchdog child.

### 3.2 Multi-Process Debugging Detection

Parent and child cooperatively check each other's debug state, creating a mutual-watch pattern.

**Bypass**: Attach to both processes (GDB `follow-fork-mode`, or two debugger instances).

### 3.3 Timing-Based with Multiple Checkpoints

Distributes timing checks across multiple functions, comparing cumulative drift. Single patches fail because the total still exceeds threshold.

**Bypass**: Frida `Interceptor.replace` all timing sources (`rdtsc`, `clock_gettime`, `QueryPerformanceCounter`) to return controlled values.

### 3.4 Nanomite / INT3 Patching

Original conditional jumps are replaced with `INT3` (0xCC). A parent debugger process handles each `INT3`, evaluates the condition, and sets the child's EIP accordingly.

**Bypass**: Reconstruct the original jump table by tracing all `INT3` handlers, then patch the binary.

---

## 4. COUNTERMEASURE TOOLS

| Tool | Platform | Capability |
|---|---|---|
| **ScyllaHide** | Windows (x64dbg/IDA/OllyDbg) | Auto-patches PEB, hooks NtQuery*, hides threads, fixes timing |
| **TitanHide** | Windows (kernel driver) | Kernel-level hiding for all user-mode checks |
| **Frida** | Cross-platform | Script-based hooking of any function, timing spoofing |
| **LD_PRELOAD shims** | Linux | Replace ptrace, getenv, fopen at load time |
| **GDB scripts** | Linux | `catch syscall`, conditional BP, register fixup |
| **Qiling** | Cross-platform | Full-system emulation, bypass all hardware checks |

---

## 5. SYSTEMATIC BYPASS METHODOLOGY

```
Step 1: Static analysis — identify anti-debug calls
  └─ Search for: ptrace, IsDebuggerPresent, NtQuery, rdtsc,
     GetTickCount, SIGTRAP, INT 2D, TLS directory entries

Step 2: Classify each check
  ├─ API-based → hook or patch the call
  ├─ Flag-based → patch PEB/proc fields
  ├─ Timing-based → spoof time source
  ├─ Exception-based → forward/handle exception correctly
  └─ Multi-process → handle both processes

Step 3: Apply bypass (order matters)
  1. Load ScyllaHide / set LD_PRELOAD (covers 80% of checks)
  2. Handle TLS callbacks (break before main)
  3. Patch remaining custom checks (Frida or binary patch)
  4. Verify: run with breakpoints, confirm no premature exit

Step 4: Validate bypass completeness
  └─ Set BP on ExitProcess/exit/_exit — if hit unexpectedly,
     a check was missed → trace back from exit call
```

---

## 6. DECISION TREE

```
Binary exits/crashes under debugger?
│
├─ Crashes immediately before main?
│  └─ TLS callback anti-debug
│     └─ Enable TLS callback breaking in debugger
│
├─ Crashes at startup?
│  ├─ Linux: check for ptrace(TRACEME)
│  │  └─ LD_PRELOAD hook or NOP patch
│  └─ Windows: check IsDebuggerPresent / PEB
│     └─ ScyllaHide or manual PEB patch
│
├─ Crashes after some execution?
│  ├─ Consistent crash point → API-based check
│  │  ├─ NtQueryInformationProcess → hook return values
│  │  ├─ /proc/self/status → filter TracerPid
│  │  └─ Hardware BP detection → hook GetThreadContext
│  │
│  ├─ Variable crash point → timing-based check
│  │  └─ Hook rdtsc / QueryPerformanceCounter
│  │
│  └─ Crash on breakpoint hit → exception-based check
│     ├─ INT 2D / INT 3 trick → handle in VEH
│     └─ SIGTRAP handler → GDB: handle SIGTRAP pass
│
├─ Debugger loses control silently?
│  └─ ThreadHideFromDebugger
│     └─ Hook NtSetInformationThread
│
├─ Child process detects and kills parent?
│  └─ Self-debugging (fork+ptrace)
│     └─ Patch fork() or handle both processes
│
└─ All basic bypasses applied but still detected?
   └─ Multi-layer / custom checks
      ├─ Use Frida for comprehensive API hooking
      ├─ Full emulation with Qiling
      └─ Trace all calls to exit/abort to find remaining checks
```

---

## 7. CTF & REAL-WORLD PATTERNS

### Common CTF Anti-Debug Patterns

| Pattern | Frequency | Quick Bypass |
|---|---|---|
| Single `ptrace(TRACEME)` | Very common | `LD_PRELOAD` one-liner |
| `IsDebuggerPresent` + `NtGlobalFlag` | Common | ScyllaHide |
| rdtsc timing in loop | Moderate | Patch comparison threshold |
| signal(SIGTRAP) + raise | Moderate | GDB signal forwarding |
| fork + ptrace watchdog | Rare but tricky | Kill child or patch fork |
| Nanomite INT3 replacement | Rare (advanced) | Reconstruct jump table |

### Real-World Protections

| Protector | Primary Anti-Debug | Recommended Tool |
|---|---|---|
| VMProtect | PEB + timing + driver-level | TitanHide + ScyllaHide |
| Themida | Multi-layer PEB + SEH + timing | ScyllaHide + manual patches |
| Enigma Protector | IsDebuggerPresent + CRC checks | x64dbg + ScyllaHide |
| UPX (custom) | Usually none (just packing) | Standard unpack |
| Custom (malware) | Varies widely | Frida + Qiling for analysis |

---

## 8. QUICK REFERENCE — BYPASS CHEAT SHEET

### Linux One-Liners

```bash
# LD_PRELOAD anti-ptrace
echo 'long ptrace(int r, ...) { return 0; }' > /tmp/ap.c
gcc -shared -o /tmp/ap.so /tmp/ap.c
LD_PRELOAD=/tmp/ap.so ./target

# GDB: catch and bypass ptrace
(gdb) catch syscall ptrace
(gdb) commands
> set $rax = 0
> continue
> end
```

### Frida Anti-Debug Bypass (Cross-Platform)

```javascript
// Hook IsDebuggerPresent (Windows)
Interceptor.replace(
  Module.getExportByName('kernel32.dll', 'IsDebuggerPresent'),
  new NativeCallback(() => 0, 'int', [])
);

// Hook ptrace (Linux)
Interceptor.replace(
  Module.getExportByName(null, 'ptrace'),
  new NativeCallback(() => 0, 'long', ['int', 'int', 'pointer', 'pointer'])
);

// Timing spoof
Interceptor.attach(Module.getExportByName(null, 'clock_gettime'), {
  onLeave(retval) {
    // manipulate timespec to hide debugger delay
  }
});
```

### x64dbg ScyllaHide Quick Setup

1. Plugins → ScyllaHide → Options
2. Check: PEB BeingDebugged, NtGlobalFlag, HeapFlags
3. Check: NtQueryInformationProcess (all classes)
4. Check: NtSetInformationThread (HideFromDebugger)
5. Check: GetTickCount, QueryPerformanceCounter
6. Apply → restart debugging session
