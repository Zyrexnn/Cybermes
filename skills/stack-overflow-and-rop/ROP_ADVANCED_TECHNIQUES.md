# Advanced ROP Techniques — BROP, Partial Overwrite, JOP, COP

> **AI LOAD INSTRUCTION**: Load this when you need Blind ROP methodology, ret2vdso for ASLR bypass, partial overwrite for PIE bypass, or alternative code-reuse paradigms (JOP/COP). Assumes the main [SKILL.md](./SKILL.md) is already loaded for fundamental ROP, ret2csu, ret2dlresolve, and SROP.

---

## 1. BLIND ROP (BROP)

Exploit a remote stack overflow **without access to the binary**. Requires a service that forks (canary and ASLR layout persist across crashes).

### BROP Attack Phases

| Phase | Goal | Method |
|---|---|---|
| 1. Stack reading | Determine buffer offset + canary | Byte-by-byte brute-force (child process crash = wrong byte) |
| 2. Find stop gadget | Address that doesn't crash | Scan code section for `ret` into valid code (e.g., infinite loop, `sleep`) |
| 3. Find BROP gadget | `__libc_csu_init` gadget | Scan for 6-pop pattern: probe address, if crash after 6 pops+ret → BROP gadget |
| 4. Find puts/write PLT | Function to leak memory | Probe PLT entries: set RDI to known readable, call candidate, check for output |
| 5. Dump binary | Leak .text, .got, .dynamic | Use puts(addr) to read binary from memory page by page |
| 6. Standard ROP | Build exploit with leaked binary | ROPgadget on dumped binary, ret2libc |

### BROP Gadget Identification

The `__libc_csu_init` tail pops 6 registers then returns. Probe:

```
[overflow][canary][saved_rbp][candidate_addr][A][A][A][A][A][A][stop_gadget]
                                              rbx rbp r12 r13 r14 r15
```

If the process survives (reaches stop gadget) → candidate is a 6-pop gadget (high probability = BROP gadget).

### Trap Gadget vs Stop Gadget

- **Stop gadget**: address that causes the process to hang or respond predictably (not crash)
- **Trap gadget**: address that crashes (0x0, unmapped page) — used as a probe terminator

### PLT Identification

PLT entries are at fixed 16-byte intervals. Probe: set RDI to a known readable address, iterate PLT base + N*16, check if output appears on socket → found puts/write.

---

## 2. ret2vdso (32-bit ASLR Bypass)

The vDSO (virtual Dynamic Shared Object) is a kernel-mapped page containing optimized syscall stubs. On **32-bit Linux kernels < 3.18**, vDSO was mapped at a **fixed address** or with low entropy.

### Attack Method

1. Locate `sigreturn` gadget in vDSO (fixed or brute-forceable address)
2. Use SROP via vDSO's sigreturn to set all registers
3. Execute `execve("/bin/sh", 0, 0)` via syscall

| Kernel Version | vDSO ASLR | Entropy |
|---|---|---|
| < 2.6.18 (32-bit) | Fixed at 0xffffe000 | None |
| 2.6.18–3.17 (32-bit) | 1 page randomization | ~8 bits (256 positions) |
| ≥ 3.18 (32-bit) | Full ASLR | Same as mmap |
| 64-bit | Always randomized | Full ASLR |

**Modern relevance**: Limited to legacy 32-bit systems. On 64-bit, vDSO is fully randomized.

---

## 3. PARTIAL OVERWRITE (PIE Bypass)

When PIE is enabled, the code base is randomized but the **last 12 bits** (page offset) are always fixed. Overwriting only the lowest 1–2 bytes of a return address can redirect execution within the same page or to a nearby page.

### Technique

```
Original return address: 0x5555555551?? (last 12 bits = 0x1??, fixed)
Overwrite last 2 bytes:  0x555555551234 → redirect to offset 0x1234 in binary

If only last byte overwritten (no null terminator issue):
  Only 4 bits unknown (nibble brute-force = 16 attempts)
If last 2 bytes overwritten:
  Only 4 bits unknown (page alignment) = 16 attempts
```

### When to Use

| Scenario | Technique |
|---|---|
| PIE + no info leak | Partial overwrite low bytes of return address |
| PIE + one-byte overflow | Overwrite saved RBP low byte → misaligned frame → secondary leak |
| PIE + format string | Leak full PIE base first (preferred over partial overwrite) |

### Practical Notes

- Null bytes in addresses: 64-bit addresses typically contain `\x00` in upper bytes, making overflow-based overwrites write the null terminator naturally
- Probability: partial overwrite success depends on unknown nibble (1/16 per attempt)
- Combine with fork-based brute-force if process restarts with same layout

---

## 4. JOP — Jump-Oriented Programming

Alternative to ROP using indirect `jmp` instructions instead of `ret`.

### Dispatcher Gadget Pattern

```nasm
; Dispatcher: advances a "virtual PC" table and jumps to next gadget
add rax, 8       ; advance table pointer
jmp [rax]        ; jump to next functional gadget
```

Each functional gadget ends with `jmp [rax]` (or equivalent) to return to the dispatcher.

### JOP Gadget Types

| Type | Example | Purpose |
|---|---|---|
| Dispatcher | `add rax, 8; jmp [rax]` | Sequence control |
| Functional | `pop rdi; jmp [rax]` | Register setup |
| Initializer | Sets RAX to dispatch table address | Bootstrap |

### When JOP Matters

- **CET/Shadow Stack**: Intel CET marks `ret` with shadow stack validation — ROP returns fail, but `jmp` gadgets are not checked by shadow stack (though IBT may restrict indirect jumps)
- Some binaries have abundant `jmp` gadgets but few `ret` gadgets

---

## 5. COP — Call-Oriented Programming

Uses indirect `call` instructions. Each gadget ends with `call [reg]` to chain to the next.

```nasm
; Example COP gadget
mov rdi, rbx
call [rax + 0x10]   ; chains to next gadget via function pointer table
```

### COP vs ROP vs JOP Comparison

| Aspect | ROP | JOP | COP |
|---|---|---|---|
| Chaining mechanism | `ret` | `jmp [reg]` | `call [reg]` |
| Stack consumption | Yes (RSP advances) | No (table-based) | Yes (pushes return addr) |
| CET Shadow Stack | Blocked | Not directly blocked | Partially blocked (IBT) |
| Gadget availability | Most common | Moderate | Least common |
| Complexity | Low | High (need dispatcher) | High |

---

## 6. STACK-BASED ORW (open-read-write)

When `execve` is blocked by seccomp but `open`/`read`/`write` are allowed, build a ROP chain to read the flag file.

### x86-64 Syscall Numbers for ORW

| Syscall | Number | Args |
|---|---|---|
| `open` (or `openat`) | 2 (257) | RDI=path, RSI=flags, RDX=mode |
| `read` | 0 | RDI=fd, RSI=buf, RDX=count |
| `write` | 1 | RDI=fd(1), RSI=buf, RDX=count |

```python
# ORW ROP chain skeleton
rop = b''
# open("flag", O_RDONLY)
rop += p64(pop_rdi) + p64(flag_str_addr)
rop += p64(pop_rsi_r15) + p64(0) + p64(0)
rop += p64(pop_rax) + p64(2) + p64(syscall_ret)
# read(fd=3, buf, 0x100)
rop += p64(pop_rdi) + p64(3)
rop += p64(pop_rsi_r15) + p64(buf_addr) + p64(0)
rop += p64(pop_rdx) + p64(0x100)
rop += p64(pop_rax) + p64(0) + p64(syscall_ret)
# write(1, buf, 0x100)
rop += p64(pop_rdi) + p64(1)
rop += p64(pop_rsi_r15) + p64(buf_addr) + p64(0)
rop += p64(pop_rdx) + p64(0x100)
rop += p64(pop_rax) + p64(1) + p64(syscall_ret)
```

---

## 7. ret2csu EXTENDED PATTERNS

### Variant: Using csu_init for Indirect Call to GOT

```python
csu_pop = elf_base + 0x40123a  # pop rbx..r15; ret
csu_call = elf_base + 0x401220 # mov rdx,r13; mov rsi,r14; mov edi,r15d; call [r12+rbx*8]

payload = b'A' * offset
payload += p64(csu_pop)
payload += p64(0)             # rbx = 0
payload += p64(1)             # rbp = 1 (skip loop)
payload += p64(elf.got['write'])  # r12 → call *GOT[write]
payload += p64(8)             # r13 → rdx = 8 (count)
payload += p64(elf.got['puts'])   # r14 → rsi = GOT[puts] (leak)
payload += p64(1)             # r15 → edi = 1 (stdout)
payload += p64(csu_call)
payload += b'A' * 56          # padding for 7 pops after call
payload += p64(main_addr)     # return to main
```

### When ret2csu Fails

- **PIE enabled**: csu gadget addresses unknown (need leak first)
- **Static binary**: `__libc_csu_init` may not exist → fall back to SROP
- **Only `edi` set**: r15d → edi (32-bit zero-extended), cannot set full 64-bit RDI → use supplementary `pop rdi` if available

---

## 8. ARCHITECTURE-SPECIFIC NOTES

### ARM ROP

| Aspect | ARM32 | AArch64 |
|---|---|---|
| Return register | LR (R14) | LR (X30) |
| Key gadget | `pop {r0-r3, pc}` | `ldp x29, x30, [sp]; ret` |
| Syscall | `svc #0` | `svc #0` |
| NOP sled | `mov r0, r0` (0xe1a00000) | `nop` (0xd503201f) |
| Thumb mode | Mixed ARM/Thumb gadgets | N/A (A64 only) |

### MIPS ROP

- No NX by default (stack executable on many MIPS devices) → shellcode often viable
- Branch delay slots: instruction after branch always executes
- Gadget: `jalr $t9` with `$a0`–`$a3` for args
- Cache coherency: may need `sleep(1)` between write and execute for I-cache flush

---

## 9. TOOLCHAIN CHEAT SHEET

```bash
# Find specific gadgets
ROPgadget --binary ./pwn --only "pop|ret" | grep rdi
ropper -f ./pwn --search "pop rdi; ret"

# Auto-generate chain
ROPgadget --binary ./pwn --ropchain

# Find one_gadget constraints
one_gadget ./libc.so.6 -l 2  # level 2 = more results, looser constraints

# Verify gadget in GDB
gdb ./pwn -ex "x/3i 0x401234"

# pwntools template
pwn template --host remote.ctf --port 1337 ./pwn > exploit.py
```
