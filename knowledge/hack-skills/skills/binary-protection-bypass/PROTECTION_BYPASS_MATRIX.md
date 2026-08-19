# Protection Bypass Matrix — Comprehensive Cross-Reference

> **AI LOAD INSTRUCTION**: Load this for a systematic lookup of which bypass technique works against which protection, and what primitive is required. Assumes [SKILL.md](./SKILL.md) is loaded for individual protection details.

---

## 1. PROTECTION × BYPASS × PRIMITIVE MATRIX

### ASLR Bypass

| Bypass Technique | Required Primitive | Success Rate | Architecture | Notes |
|---|---|---|---|---|
| Format string leak | `printf(user_input)` | 100% (deterministic) | Any | `%p` leak stack/libc/heap addresses |
| OOB read | Array bounds violation | 100% | Any | Read adjacent pointers |
| UAF read | Use-after-free | 100% | Any | Read freed chunk fd/bk → libc/heap |
| Brute force | Ability to retry | ~1/4096 (32-bit) | x86 only | Infeasible on 64-bit (28-bit entropy) |
| Partial overwrite | 1-2 byte write | 1/16 per nibble | Any | Page offset (12 bits) is fixed |
| ret2dlresolve | Stack overflow + writable area | 100% | Any | No ASLR knowledge needed |
| SROP | Stack overflow + sigreturn gadget | 100% | Any | Set all registers without knowledge |
| Return-to-PLT | Stack overflow | 100% | No PIE | PLT addresses fixed without PIE |
| Stack reading | Fork server + crash oracle | 100% | Any | Byte-by-byte in child process |
| stdout FILE abuse | Write to stdout structure | 100% | Any | Partial overwrite `_IO_write_base` |

### PIE Bypass

| Bypass Technique | Required Primitive | Success Rate | Notes |
|---|---|---|---|
| Leak .text pointer | Read from stack (return addr) | 100% | PIE base = addr - offset |
| Partial overwrite | 1-2 byte overflow | 1/16 | Last 12 bits fixed |
| BROP | Fork server, crash probing | ~100% | Blind discovery without binary |
| Relative addressing | Known offset between objects | 100% | Intra-binary references |

### NX Bypass

| Bypass Technique | Required Primitive | Gadget Requirement | Notes |
|---|---|---|---|
| ROP chain | Stack overflow | `pop reg; ret` gadgets | Standard approach |
| ret2libc | Stack overflow | `pop rdi; ret` (64-bit) | Call system/execve |
| ret2csu | Stack overflow | `__libc_csu_init` | 3 args without `pop rdx` |
| SROP | Stack overflow | `syscall; ret` + sigreturn | Set all registers |
| mprotect chain | Stack overflow + known address | `pop rdi/rsi/rdx; ret` | Make page RWX |
| JIT spray | JIT engine present | None | Plant code in JIT pages |

### Canary Bypass

| Bypass Technique | Required Primitive | Condition | Notes |
|---|---|---|---|
| Format string leak | `printf(user_input)` | Canary on stack before return addr | `%N$p` reads canary |
| Brute force | Fork server | Canary same in child | 256 × 7 attempts (64-bit) |
| Stack reading | One-byte write/read | Overwrite null byte, read error | Output-based oracle |
| Thread TLS overwrite | Large overflow | Overflow reaches `fs:[0x28]` | Overwrite canary source |
| `__stack_chk_fail` GOT | Partial RELRO + write | GOT writable | Replace with no-op |
| Avoid stack entirely | Heap vulnerability | No canary on heap | Heap exploitation path |

### RELRO Bypass

| RELRO Level | Writable Targets | Bypass |
|---|---|---|
| None | `.got`, `.got.plt`, `.dynamic` | Direct GOT overwrite |
| Partial | `.got.plt` (lazy binding entries) | GOT overwrite on lazy-bound functions |
| Full | Nothing in GOT | `__malloc_hook` (pre-2.34), `_IO_FILE`, `exit_funcs`, stack, `.fini_array` |

### Full RELRO Alternative Target Matrix

| Target | glibc Version | Required Knowledge | Overwrite Size | Trigger |
|---|---|---|---|---|
| `__malloc_hook` | < 2.34 | libc base | 8 bytes | Any malloc (printf with large fmt) |
| `__free_hook` | < 2.34 | libc base | 8 bytes | Any free |
| `__realloc_hook` | < 2.34 | libc base | 8 bytes | Any realloc |
| `_IO_list_all` | Any | libc base | 8 bytes | exit / abort |
| `_IO_FILE vtable` | Any (bypass varies) | libc base + heap | 8 bytes + fake vtable | I/O operation or exit |
| `__exit_funcs` | Any | libc base + pointer guard | 8 bytes (mangled) | exit() |
| `TLS_dtor_list` | ≥ 2.34 | TLS addr + pointer guard | 8 bytes (mangled) | Thread exit / exit() |
| `.fini_array` | If writable | Binary base | 8 bytes | Normal program exit |
| `_dl_fini` link_map | Any | ld.so base | Multiple fields | exit() |
| Stack return address | Always | Stack address | 8 bytes | Function return |

---

## 2. COMBINED PROTECTION SCENARIOS

### Scenario: NX + ASLR + Canary + Partial RELRO (Typical CTF)

```
1. Leak canary via format string or info disclosure
2. Leak libc address via format string or PLT leak (puts@GOT)
3. Craft ROP chain: pop_rdi → "/bin/sh" → system (ret2libc)
4. Overflow: [padding][canary][saved_rbp][ROP chain]
```

### Scenario: NX + ASLR + Canary + Full RELRO + PIE

```
1. Format string (multi-shot): leak canary + PIE base + libc base
2. Cannot overwrite GOT (Full RELRO) → target __malloc_hook or _IO_FILE
3. Second format string or overflow: write target address
4. Trigger: call malloc (for __malloc_hook) or exit (for _IO_FILE)
```

### Scenario: NX + ASLR + No Canary + No PIE + Partial RELRO (Beginner CTF)

```
1. No canary → direct overflow to return address
2. No PIE → binary addresses known → PLT/GOT addresses fixed
3. Leak libc: overflow → puts@plt(puts@GOT) → leak → main
4. Second overflow: system("/bin/sh") or one_gadget
```

### Scenario: Static Binary + NX + Canary + ASLR

```
1. No libc (static) → cannot ret2libc
2. Need canary bypass → leak via available primitive
3. SROP: sigreturn to set all registers → execve("/bin/sh", NULL, NULL)
4. Or: ret2dlresolve not applicable (static) → rely on ROP gadgets from large binary
```

---

## 3. COMPILATION FLAGS REFERENCE

| Protection | Enable Flag | Disable Flag |
|---|---|---|
| Canary | `-fstack-protector-all` | `-fno-stack-protector` |
| NX | Default on (`-z noexecstack`) | `-z execstack` |
| PIE | `-pie` (default on modern gcc) | `-no-pie` |
| Partial RELRO | `-Wl,-z,relro` (default) | `-Wl,-z,norelro` |
| Full RELRO | `-Wl,-z,relro,-z,now` | Omit `-z,now` |
| FORTIFY | `-D_FORTIFY_SOURCE=2 -O2` | Omit flag |
| ASLR | Kernel: `echo 2 > /proc/sys/kernel/randomize_va_space` | `echo 0 > ...` |

### QEMU/GDB Testing Flags

```bash
# Disable all protections for testing
gcc -fno-stack-protector -z execstack -no-pie -Wl,-z,norelro -o vuln vuln.c
# Disable ASLR for debugging
echo 0 | sudo tee /proc/sys/kernel/randomize_va_space
# Or per-process:
setarch $(uname -m) -R ./vuln
```

---

## 4. PROTECTION INTERACTION EFFECTS

| Protection A | Protection B | Combined Effect |
|---|---|---|
| ASLR | PIE | Both code AND data randomized → need two leaks (or relative offset) |
| Full RELRO | glibc ≥ 2.34 | GOT + hooks both unavailable → _IO_FILE or exit_funcs only targets |
| Canary | ASLR | Must leak canary AND libc base before exploitation |
| NX | ASLR | ROP chain needed, but gadget addresses unknown → leak first |
| CET | Full RELRO | ROP blocked + GOT blocked → data-only attacks |
| FORTIFY | Canary | Format string restricted + stack overflow detected → heap path preferred |

---

## 5. QUICK LOOKUP: "I HAVE X, I NEED Y"

| I Have | I Need | Path |
|---|---|---|
| Format string only | RCE | Leak canary + libc → overwrite GOT(printf→system) or hook |
| Stack overflow only (no leak) | RCE | ret2dlresolve or SROP (no ASLR knowledge needed) |
| Stack overflow + format string | RCE | Leak everything → classic ROP chain |
| Heap overflow only | RCE | Heap exploitation → leak libc → overwrite hook/_IO_FILE → trigger |
| OOB read only | Leak | Read GOT entries → libc base; read stack → canary + PIE |
| Arbitrary write (no leak) | RCE | Partial overwrite: .got.plt low bytes to redirect to win function |
| Arbitrary write + libc leak | RCE | Write one_gadget to hook (pre-2.34) or FSOP (any version) |
