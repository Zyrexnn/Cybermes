---
name: binary-protection-bypass
description: >-
  Binary protection bypass playbook. Use when identifying and bypassing ASLR, PIE, NX/DEP, stack canary, RELRO, FORTIFY_SOURCE, CET, and MTE protections in ELF binaries to enable exploitation.
---

# SKILL: Binary Protection Bypass — Expert Attack Playbook

> **AI LOAD INSTRUCTION**: Expert binary protection identification and bypass techniques. Covers ASLR, PIE, NX, RELRO, canary, FORTIFY_SOURCE, stack clash, CET shadow stack, and ARM MTE. Each protection is paired with its bypass methods and required primitives. Distilled from ctf-wiki mitigation sections and real-world exploitation. Base models often confuse which protections block which attacks and miss the combinatorial effect of multiple protections.

## 0. RELATED ROUTING

- [stack-overflow-and-rop](../stack-overflow-and-rop/SKILL.md) — ROP chains to bypass NX, ret2libc for ASLR bypass
- [format-string-exploitation](../format-string-exploitation/SKILL.md) — primary method for leaking canary, PIE, libc addresses
- [heap-exploitation](../heap-exploitation/SKILL.md) — heap attacks for RELRO bypass (when GOT is read-only)
- [arbitrary-write-to-rce](../arbitrary-write-to-rce/SKILL.md) — what to overwrite when GOT is protected by RELRO

### Advanced Reference

Load [PROTECTION_BYPASS_MATRIX.md](./PROTECTION_BYPASS_MATRIX.md) for comprehensive protection × bypass × primitive matrix.

---

## 1. PROTECTION IDENTIFICATION

```bash
$ checksec ./binary
[*] '/path/to/binary'
    Arch:     amd64-64-little
    RELRO:    Full RELRO          ← GOT read-only
    Stack:    Canary found        ← stack canary enabled
    NX:       NX enabled          ← stack not executable
    PIE:      PIE enabled         ← position-independent code
    FORTIFY:  Enabled             ← fortified libc functions
```

### Quick Identification Table

| Protection | Check Command | Binary Indicator |
|---|---|---|
| ASLR | `cat /proc/sys/kernel/randomize_va_space` | OS-level (0=off, 1=partial, 2=full) |
| PIE | `checksec` or `readelf -h` (Type: DYN) | Binary compiled with `-pie` |
| NX | `checksec` or `readelf -l` (no RWE segment) | `gcc -z noexecstack` (default on) |
| Canary | `checksec` or look for `__stack_chk_fail@plt` | `gcc -fstack-protector-all` |
| Partial RELRO | `readelf -l` (GNU_RELRO segment, `.got.plt` writable) | `gcc -Wl,-z,relro` |
| Full RELRO | `readelf -l` + `.got` section read-only | `gcc -Wl,-z,relro,-z,now` |
| FORTIFY | Presence of `__printf_chk`, `__memcpy_chk` etc. | `gcc -D_FORTIFY_SOURCE=2` |

---

## 2. ASLR BYPASS

ASLR randomizes base addresses of stack, heap, libc, and mmap regions at each execution.

| Bypass Method | Required Primitive | Notes |
|---|---|---|
| Information leak | Any read primitive (format string, OOB read, UAF) | Leak libc/stack/heap address → calculate base |
| Partial overwrite | Write primitive (limited length) | Overwrite last 1-2 bytes (page offset fixed) |
| Brute force (32-bit) | Ability to reconnect/retry | ~256–4096 attempts (8-12 bits entropy) |
| Return-to-PLT | Stack overflow | PLT addresses are at fixed offset from binary base (if no PIE) |
| ret2dlresolve | Stack overflow + write primitive | Resolve arbitrary function without knowing libc base |
| Format string leak | Format string vulnerability | `%N$p` for stack/libc/heap addresses |
| Stack reading | Byte-by-byte (fork server) | Read stack byte-by-byte via crash oracle |

### ASLR Entropy (x86-64 Linux)

| Region | Entropy (bits) | Positions |
|---|---|---|
| Stack | 22 | ~4M |
| mmap / libc | 28 | ~256M |
| Heap (brk) | 13 | ~8K |
| PIE binary | 28 | ~256M |

---

## 3. PIE BYPASS

PIE (Position Independent Executable) randomizes the binary's own code/data base address.

| Bypass Method | Required Primitive | Notes |
|---|---|---|
| Information leak | Read return address from stack | PIE base = leaked_addr - known_offset |
| Partial overwrite | One-byte or two-byte write | Last 12 bits of page offset are fixed |
| Format string leak | Format string vulnerability | `%N$p` where N points to .text return address |
| Relative addressing | Knowledge of binary layout | If you know relative offsets, only need one leak |

### Partial Overwrite Details

```
PIE binary loaded at: 0x555555554000 (example)
Function at offset 0x1234: 0x555555555234

Overwrite return address last 2 bytes: 0x?234 → 0x?XXX
Unknown: bits 12-15 (one nibble = 4 bits = 16 possibilities)
Success rate: 1/16 per attempt
```

---

## 4. NX / DEP BYPASS

NX (No-eXecute) / DEP (Data Execution Prevention) prevents execution of code on the stack/heap.

| Bypass Method | Detail |
|---|---|
| ROP (Return-Oriented Programming) | Chain existing code gadgets ending in `ret` |
| ret2libc | Call libc functions (system, execve) directly |
| ret2csu | Use `__libc_csu_init` gadgets for controlled function calls |
| ret2dlresolve | Forge dynamic linker structures to resolve arbitrary functions |
| SROP | Use sigreturn to set all registers from fake signal frame |
| mprotect ROP | Chain mprotect(addr, size, PROT_RWX) → make page executable → jump to shellcode |
| JIT spray | In JIT environments (V8, etc.), create executable code via JIT compiler |

### mprotect Chain

```python
# Make stack executable, then jump to shellcode
rop = b'A' * offset
rop += p64(pop_rdi) + p64(stack_page)     # page-aligned address
rop += p64(pop_rsi) + p64(0x1000)         # size
rop += p64(pop_rdx) + p64(7)              # PROT_READ|PROT_WRITE|PROT_EXEC
rop += p64(mprotect_addr)
rop += p64(shellcode_addr)                 # jump to shellcode on now-executable stack
```

---

## 5. RELRO BYPASS

| RELRO Level | GOT Status | Bypass |
|---|---|---|
| No RELRO | GOT fully writable | Direct GOT overwrite |
| Partial RELRO | `.got.plt` writable (lazy binding) | GOT overwrite still works |
| Full RELRO | All GOT entries resolved at load, GOT read-only | Cannot write GOT → target other structures |

### Full RELRO Alternative Targets

| Target | When | How |
|---|---|---|
| `__malloc_hook` | glibc < 2.34 | Overwrite with one_gadget |
| `__free_hook` | glibc < 2.34 | Overwrite with `system`, trigger `free("/bin/sh")` |
| `_IO_FILE vtable` | Any glibc | FSOP / vtable hijack |
| `__exit_funcs` | Any glibc | Overwrite exit handler list |
| `TLS_dtor_list` | glibc ≥ 2.34 | Thread-local destructor list (needs pointer guard) |
| `.fini_array` | If writable | Overwrite destructor function pointers |
| Stack return address | Direct stack write | Overwrite return address for ROP |

See [arbitrary-write-to-rce](../arbitrary-write-to-rce/SKILL.md) for comprehensive target list.

---

## 6. CANARY BYPASS

| Method | Condition | Detail |
|---|---|---|
| Format string leak | printf(user_input) | `%N$p` to read canary from stack |
| Brute-force | fork() server (canary persists in child) | Byte-by-byte: 256 × (canary_size-1) attempts |
| Stack reading | Partial overwrite / info leak | Overwrite canary's null byte, leak via output |
| Thread canary overwrite | Overflow reaches TLS | Canary at `fs:[0x28]`; overflow past buffer to TLS → overwrite canary with known value |
| Canary-relative overwrite | Overflow after canary but before return addr | Skip canary, only overwrite return address (rare layout) |
| Heap-based | Vulnerability is on heap, not stack | Canary only protects stack |
| __stack_chk_fail GOT overwrite | Partial RELRO | Overwrite `__stack_chk_fail@GOT` to point to harmless function → canary check passes |

### Canary Format

```
x86:    0x00XXXXXX (4 bytes, leading null byte)
x86-64: 0x00XXXXXXXXXXXXXX (8 bytes, leading null byte)
```

The leading `\x00` prevents string operations from accidentally reading the canary.

---

## 7. FORTIFY_SOURCE BYPASS

`_FORTIFY_SOURCE=2` adds buffer size checking and restricts format string operations.

| Fortified Function | Restriction | Bypass |
|---|---|---|
| `__printf_chk` | `%n` with positional args (`%N$n`) forbidden | Use non-positional `%n` or `%hn` chain |
| `__memcpy_chk` | Destination buffer size checked | Use heap overflow instead of stack |
| `__strcpy_chk` | Same | |
| `__read_chk` | Read size checked against buffer | |

### Format String with FORTIFY_SOURCE

```python
# %1$n is blocked by __printf_chk
# But sequential (non-positional) %n may still work:
# Print exact byte count, then %hn — must be very precise
# Or: find unfortified printf in binary/libc via ROP
```

---

## 8. CET (Control-flow Enforcement Technology)

Intel CET adds two mechanisms:

### Shadow Stack

- Hardware-maintained copy of return addresses
- On `ret`, CPU checks shadow stack matches actual stack
- Mismatch → `#CP` fault (control protection exception)

| Impact | Detail |
|---|---|
| ROP blocked | Return address overwrite detected on `ret` |
| JOP possible | `jmp [reg]` not checked by shadow stack |
| COP possible | `call [reg]` pushes to shadow stack but target validated by IBT |

### Indirect Branch Tracking (IBT)

- Indirect `jmp`/`call` must land on `ENDBR64` instruction
- Non-ENDBR landing → `#CP` fault

**Bypass**: 
- Data-only attacks (don't change control flow)
- Find valid ENDBR gadgets that chain into useful operations
- JOP with ENDBR-prefixed gadgets
- Target structures outside CFI scope (modprobe_path, function pointer arrays)

---

## 9. MTE (Memory Tagging Extension, ARM)

ARM MTE assigns 4-bit tags to memory pointers and allocations. Tag mismatch = fault.

| Aspect | Detail |
|---|---|
| Tag bits | 4 bits in pointer (bits 56-59) = 16 possible tags |
| Granule | 16 bytes (each 16-byte granule has one tag) |
| Check | Load/store: pointer tag must match memory tag |
| Probabilistic | Random tag → 1/16 chance attacker guesses correctly |

### Bypass Approaches

| Method | Success Rate |
|---|---|
| Brute-force | 1/16 per attempt (6.25%) |
| Tag oracle | Side-channel to determine tag (timing, error messages) |
| In-bounds exploit | Stay within same tagged region (use relative offsets) |
| Tag bypass gadget | Use `LDGM`/`STGM` instructions if accessible |
| Speculative execution | Spectre-style bypass of tag check |

---

## 10. DECISION TREE

```
Binary analysis: checksec output
├── NX disabled?
│   └── Shellcode on stack/heap (simplest path)
│
├── NX enabled (standard modern binary)?
│   ├── Need code execution → ROP/ret2libc
│   │
│   ├── Canary enabled?
│   │   ├── fork server? → byte-by-byte brute-force
│   │   ├── Format string? → leak canary via %p
│   │   ├── Heap vuln? → canary doesn't protect heap
│   │   └── Partial RELRO? → overwrite __stack_chk_fail@GOT
│   │
│   ├── PIE enabled?
│   │   ├── Format string? → leak .text address → PIE base
│   │   ├── Partial overwrite → last 12 bits fixed (1/16 brute-force)
│   │   └── OOB read? → leak code pointer
│   │
│   ├── ASLR enabled?
│   │   ├── Info leak available → leak libc base
│   │   ├── No leak → ret2dlresolve or SROP
│   │   ├── 32-bit? → brute-force feasible (~4096 attempts)
│   │   └── Return-to-PLT (no libc base needed for PLT calls)
│   │
│   ├── RELRO level?
│   │   ├── None/Partial → GOT overwrite
│   │   └── Full → alternative targets:
│   │       ├── glibc < 2.34 → __malloc_hook / __free_hook
│   │       ├── glibc ≥ 2.34 → _IO_FILE / exit_funcs / TLS_dtor_list
│   │       ├── .fini_array (if writable)
│   │       └── Stack return address
│   │
│   └── FORTIFY_SOURCE?
│       ├── Blocks positional %n → use sequential %n or heap exploit
│       └── Blocks buffer overflows in fortified functions → use unfortified paths
│
├── CET (shadow stack)?
│   ├── ROP blocked → data-only attack or JOP
│   └── ENDBR-gadget chaining
│
└── MTE (ARM)?
    ├── 1/16 brute-force
    └── Stay in-bounds for relative corruption
```
