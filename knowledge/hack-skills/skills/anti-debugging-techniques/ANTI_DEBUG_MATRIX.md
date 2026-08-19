# Anti-Debug Technique × OS × Detection × Bypass — Comprehensive Matrix

> **AI LOAD INSTRUCTION**: Load this when you need the full cross-reference of anti-debugging techniques, their OS applicability, detection methods, bypass tools, reliability ratings, and false-positive notes. Assumes the main [SKILL.md](./SKILL.md) is already loaded for conceptual understanding.

---

## 1. LINUX ANTI-DEBUG MATRIX

| # | Technique | Detection Method | Reliability | Bypass Method | Bypass Tool | False Positives |
|---|---|---|---|---|---|---|
| L1 | `ptrace(PTRACE_TRACEME)` | Self-attach; fails if already traced | High | `LD_PRELOAD` shim, NOP patch, GDB `catch syscall` | GDB, gcc | None — definitive |
| L2 | `/proc/self/status` TracerPid | Read TracerPid field; non-zero = traced | High | Hook `fopen`/`fread`, FUSE mount, patch string | Frida, LD_PRELOAD | Container environments may show artifacts |
| L3 | `/proc/self/maps` scanning | Search for debugger/instrumentation libraries | Medium | Filter maps output via hook, rename agent libs | Frida (rename gadget.so) | Security tools may trigger |
| L4 | `rdtsc` timing | Measure cycle count delta between two points | Medium | Fix registers at BP, hook timing source | GDB scripts, Frida | High CPU load can cause false positives |
| L5 | `clock_gettime` timing | Similar to rdtsc but via syscall | Medium | Hook `clock_gettime`, return controlled values | Frida, LD_PRELOAD | System load variation |
| L6 | `SIGTRAP` handler | Install handler, raise SIGTRAP; debugger swallows it | High | GDB: `handle SIGTRAP nostop pass` | GDB | None |
| L7 | `SIGSTOP`/`SIGCONT` self-send | Send SIGSTOP to self, measure if debugger intervenes | Low | Forward signals properly | GDB signal handling | Rare |
| L8 | Fork + ptrace watchdog | Child attaches to parent; fails if debugger present | High | Kill child, patch fork, dual-attach | GDB (follow-fork-mode) | None |
| L9 | `LD_PRELOAD` env check | `getenv("LD_PRELOAD")` | Low | Unset env var, hook `getenv` | Shell, Frida | Legitimate LD_PRELOAD usage |
| L10 | Parent PID check | `getppid()` — expect init/shell, not debugger | Low | Run from shell normally, hook `getppid` | Frida | Terminal multiplexers |
| L11 | `/proc/self/exe` readlink | Check if binary path matches expected | Low | Symlink or hook `readlink` | Shell | Custom install paths |
| L12 | Breakpoint scanning (0xCC) | Scan `.text` for `INT3` bytes | Medium | Use hardware breakpoints only | x86 HW BP (DR0-DR3) | Legitimate 0xCC in data |
| L13 | `prctl(PR_SET_DUMPABLE, 0)` | Prevent ptrace attach after start | Medium | Hook `prctl`, keep dumpable | LD_PRELOAD, Frida | None |
| L14 | `personality(ADDR_NO_RANDOMIZE)` | Detect if ASLR disabled (common debugger setting) | Low | Keep ASLR enabled while debugging | GDB (don't disable ASLR) | Manual ASLR disable |

---

## 2. WINDOWS ANTI-DEBUG MATRIX

| # | Technique | Detection Method | Reliability | Bypass Method | Bypass Tool | False Positives |
|---|---|---|---|---|---|---|
| W1 | `IsDebuggerPresent` | Reads `PEB.BeingDebugged` | High | Patch PEB byte, hook API | ScyllaHide, x64dbg | None |
| W2 | `CheckRemoteDebuggerPresent` | Calls `NtQueryInformationProcess(DebugPort)` | High | Hook underlying NtQIP | ScyllaHide | None |
| W3 | PEB.BeingDebugged | Direct PEB read (no API call) | High | Zero the byte at PEB+0x02 | ScyllaHide, manual patch | None |
| W4 | PEB.NtGlobalFlag (0x70) | Check for `FLG_HEAP_ENABLE_*` flags | High | Zero PEB+0xBC | ScyllaHide | None |
| W5 | Heap flags | `ProcessHeap.Flags` / `ForceFlags` | High | Patch heap header | ScyllaHide | None |
| W6 | `NtQueryInformationProcess` DebugPort | InfoClass 0x07 → non-zero if debugged | High | Hook NtQIP, return 0 | ScyllaHide, Frida | None |
| W7 | `NtQueryInformationProcess` DebugObjectHandle | InfoClass 0x1E → valid handle if debugged | High | Hook NtQIP, return error | ScyllaHide | None |
| W8 | `NtQueryInformationProcess` DebugFlags | InfoClass 0x1F → 0 if debugged (inverted!) | High | Hook NtQIP, return 1 | ScyllaHide | None |
| W9 | `OutputDebugString` timing | Measure time for ODS call (faster with debugger) | Medium | Hook ODS or fix timing | Frida | System load |
| W10 | INT 2D | Kernel debug interrupt; byte-skip behavior differs | High | Handle in VEH, NOP patch | ScyllaHide, manual | None |
| W11 | INT 3 (0xCC) | Breakpoint instruction behavior | Medium | Single-step past, VEH | Debugger built-in | None |
| W12 | UD2 (#UD exception) | Invalid opcode exception handling differs | Medium | Handle in VEH | Manual | None |
| W13 | TLS callbacks | Code runs before entry point | High | Break on TLS callback | x64dbg option, WinDbg | None |
| W14 | `NtSetInformationThread` HideFromDebugger | Thread becomes invisible to debugger | High | Hook NtSIT, NOP the call | ScyllaHide | None |
| W15 | DR register check | `GetThreadContext` reads DR0-DR3 | High | Hook GTC, zero DRx | ScyllaHide | None |
| W16 | `NtQuerySystemInformation` SystemKernelDebuggerInformation | Detects kernel debugger | High (kernel) | TitanHide (kernel driver) | TitanHide | None |
| W17 | VEH chain inspection | Walk VEH list for debugger-installed handlers | Low | Don't install VEH from debugger | Manual | Security software VEHs |
| W18 | `CloseHandle(invalid)` | With debugger: raises exception; without: returns error | Medium | Handle exception in VEH | ScyllaHide | None |
| W19 | `NtClose(invalid)` | Same as CloseHandle trick at NT level | Medium | Hook NtClose | ScyllaHide | None |
| W20 | SEH-based detection | Install SEH, trigger exception, check handler invocation | Medium | Ensure correct SEH dispatch | Debugger settings | None |
| W21 | `QueryPerformanceCounter` timing | Measure ticks between two points | Medium | Hook QPC, spoof delta | ScyllaHide, Frida | System load |
| W22 | `GetTickCount` / `GetTickCount64` timing | Millisecond-level timing check | Medium | Hook and spoof | ScyllaHide | System load |
| W23 | `RDTSC` instruction | Direct CPU timestamp counter | Medium | Patch comparison or hook via VEH on #UD | Frida (replace block) | CPU frequency changes |
| W24 | Parent process check | `NtQueryInformationProcess` → check parent is explorer.exe | Low | Spoof parent PID or hook | Frida | Non-standard launchers |
| W25 | Window class enumeration | `FindWindow("OLLYDBG")`, `FindWindow("x64dbg")` | Low | Rename debugger window class | Debugger plugin | None |
| W26 | Process enumeration | Enumerate processes for known debugger names | Low | Rename debugger executable | Shell | None |
| W27 | CRC / integrity check | Hash `.text` section, compare against stored value | Medium | Patch stored hash or hook CRC function | Manual, Frida | Legitimate code updates |
| W28 | `BlockInput(TRUE)` | Lock keyboard/mouse during sensitive operations | Low | Hook `BlockInput` | ScyllaHide | None |

---

## 3. TOOL COMPATIBILITY MATRIX

| Technique | GDB | x64dbg + ScyllaHide | WinDbg | IDA Remote | Frida | Qiling |
|---|---|---|---|---|---|---|
| ptrace self-attach | `catch syscall` | N/A | N/A | N/A | Hook | Emulate |
| /proc/self/status | Manual hook | N/A | N/A | N/A | Hook fopen | Emulate |
| PEB.BeingDebugged | N/A | Auto-patch | Manual | Plugin | Hook | Emulate |
| NtGlobalFlag | N/A | Auto-patch | Manual | Plugin | Hook | Emulate |
| NtQueryInformationProcess | N/A | Auto-hook | Manual | Plugin | Hook | Emulate |
| IsDebuggerPresent | N/A | Auto-hook | `eb kernel32!IsDebuggerPresent` | Plugin | Hook | Emulate |
| rdtsc timing | Register fixup | Spoof QPC | Manual | N/A | Replace block | Emulate |
| INT 2D / INT 3 | Handle signal | VEH/auto | Handle exception | N/A | Replace insn | Emulate |
| TLS callback | `starti` | Break on TLS | `sxe ld` | Break on entry | Early inject | Emulate |
| ThreadHideFromDebugger | N/A | Auto-NOP | Manual | Plugin | Hook NtSIT | Emulate |
| DR register check | N/A | Auto-zero DRx | Manual | N/A | Hook GTC | Emulate |
| fork+ptrace watchdog | follow-fork-mode | N/A | N/A | N/A | Hook fork | Emulate both |
| SIGTRAP handler | `handle pass` | N/A | N/A | N/A | Hook signal | Emulate |
| Breakpoint scan (0xCC) | Use HW BP | Use HW BP | Use HW BP | N/A | No BP needed | Emulate |

---

## 4. BYPASS PRIORITY CHECKLIST

Apply bypasses in this order for maximum coverage with minimum effort:

```
Phase 1 — Automated tools (covers ~80%)
├─ Windows: Load ScyllaHide with all options checked
├─ Linux: Set LD_PRELOAD with ptrace + timing shims
└─ Verify: program runs past initial checks

Phase 2 — TLS / early execution (covers +10%)
├─ Break on TLS callbacks or _init functions
└─ Patch any pre-main checks found

Phase 3 — Custom checks (covers +8%)
├─ Trace exit/abort calls → backtrack to find check
├─ Frida hook remaining detection functions
└─ Patch binary for persistent bypass

Phase 4 — Multi-process / kernel (covers +2%)
├─ Handle fork+ptrace with dual-process debugging
├─ TitanHide for kernel-level debugger detection
└─ Qiling full emulation for heavily protected targets
```

---

## 5. DETECTION RELIABILITY RATING GUIDE

| Rating | Meaning | Example |
|---|---|---|
| **High** | Definitive detection, no false positives | PEB.BeingDebugged, ptrace self-attach |
| **Medium** | Reliable but environment-sensitive | Timing checks (CPU load affects), breakpoint scanning |
| **Low** | Easily spoofed or many false positives | Process name enumeration, window class search, env var check |
