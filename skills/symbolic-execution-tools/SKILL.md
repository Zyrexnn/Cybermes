---
name: symbolic-execution-tools
description: >-
  Symbolic execution and constraint solving playbook. Use when solving CTF
  reversing challenges, recovering keys, bypassing checks, or automating
  binary analysis with angr, Z3, or Unicorn Engine.
---

# SKILL: Symbolic Execution Tools — Expert Analysis Playbook

> **AI LOAD INSTRUCTION**: Expert symbolic execution techniques using angr, Z3, and Unicorn Engine. Covers CTF challenge automation, constraint solving patterns, function hooking, SimProcedure replacement, and emulation-based unpacking. Base models often produce broken angr scripts due to incorrect state initialization or missing hooks for libc functions.

## 0. RELATED ROUTING

- [anti-debugging-techniques](../anti-debugging-techniques/SKILL.md) when anti-debug checks need to be symbolically bypassed
- [code-obfuscation-deobfuscation](../code-obfuscation-deobfuscation/SKILL.md) when using symbolic execution for deobfuscation
- [vm-and-bytecode-reverse](../vm-and-bytecode-reverse/SKILL.md) when applying angr to custom VM challenges

### Advanced Reference

Also load [ANGR_COOKBOOK.md](./ANGR_COOKBOOK.md) when you need:
- 15+ ready-to-use angr script patterns for common CTF challenges
- Hook templates for scanf, printf, malloc, strcmp
- Symbolic file input, stdin, argv patterns
- Optimization tricks for path explosion management

### When to use which tool

| Scenario | Best Tool | Why |
|---|---|---|
| Pure math / equation system | Z3 | Direct constraint solving, no binary needed |
| Binary with control flow | angr | Explores paths, manages constraints automatically |
| Emulate specific code region | Unicorn | Fast, no symbolic overhead, good for unpacking |
| Complex binary + custom VM | angr + Unicorn (combo) | angr for control flow, Unicorn for VM handlers |
| Kernel / firmware code | Qiling | Full system emulation with OS awareness |

---

## 1. ANGR — CORE CONCEPTS

### 1.1 Pipeline

```
Project(binary)
  → Factory.entry_state() / blank_state(addr=)
    → SimulationManager(state)
      → explore(find=target, avoid=bad)
        → found[0].solver.eval(symbolic_var)
```

### 1.2 Essential Setup

```python
import angr
import claripy

proj = angr.Project('./challenge', auto_load_libs=False)

# Entry state: start from program entry point
state = proj.factory.entry_state()

# Blank state: start from arbitrary address
state = proj.factory.blank_state(addr=0x401000)

# Full init state: with command-line args
state = proj.factory.full_init_state(args=['./challenge', arg1_sym])

simgr = proj.factory.simulation_manager(state)
simgr.explore(find=0x401234, avoid=[0x401300])

if simgr.found:
    found = simgr.found[0]
    solution = found.solver.eval(symbolic_input, cast_to=bytes)
    print(f"Solution: {solution}")
```

### 1.3 Symbolic Variables (claripy)

```python
# Bitvector (fixed-size integer)
sym_input = claripy.BVS("input", 64)        # 64-bit symbolic
sym_byte = claripy.BVS("byte", 8)           # 8-bit symbolic
sym_buf = claripy.BVS("buffer", 8 * 32)     # 32-byte buffer

# Concrete bitvector
concrete = claripy.BVV(0x41, 8)             # concrete value 0x41

# Constraints
state.solver.add(sym_input > 0)
state.solver.add(sym_input < 100)
state.solver.add(sym_byte >= 0x20)           # printable ASCII
state.solver.add(sym_byte <= 0x7e)

# Evaluate
value = state.solver.eval(sym_input)
all_values = state.solver.eval_upto(sym_input, 10)  # up to 10 solutions
```

### 1.4 Symbolic stdin

```python
flag_len = 32
sym_stdin = claripy.BVS("stdin", 8 * flag_len)

state = proj.factory.entry_state(stdin=sym_stdin)

# Constrain to printable ASCII
for i in range(flag_len):
    byte = sym_stdin.get_byte(i)
    state.solver.add(byte >= 0x20)
    state.solver.add(byte <= 0x7e)
```

### 1.5 Hooking Functions

```python
# Hook by address (skip N bytes of original code)
@proj.hook(0x401100, length=5)
def skip_check(state):
    state.regs.eax = 1  # force success

# SimProcedure: replace library function
class MyStrcmp(angr.SimProcedure):
    def run(self, s1, s2):
        return claripy.If(
            self.state.memory.load(s1, 32) == self.state.memory.load(s2, 32),
            claripy.BVV(0, 32),
            claripy.BVV(1, 32)
        )

proj.hook_symbol('strcmp', MyStrcmp())

# Hook common problematic functions
proj.hook_symbol('printf', angr.SIM_PROCEDURES['libc']['printf']())
proj.hook_symbol('scanf', angr.SIM_PROCEDURES['libc']['scanf']())
proj.hook_symbol('puts', angr.SIM_PROCEDURES['libc']['puts']())
```

### 1.6 Memory Operations

```python
# Read memory (symbolic-aware)
data = state.memory.load(addr, size)          # returns BV
data_concrete = state.solver.eval(data, cast_to=bytes)

# Write memory
state.memory.store(addr, claripy.BVV(0x41, 8))
state.memory.store(addr, sym_buf)

# Read/write registers
rax = state.regs.rax
state.regs.rdi = claripy.BVV(0x1000, 64)
```

---

## 2. Z3 CONSTRAINT SOLVING

### 2.1 Core API

```python
from z3 import *

# Sorts
x = BitVec('x', 32)    # 32-bit bitvector
y = Int('y')             # arbitrary precision integer
b = Bool('b')            # boolean

# Solver
s = Solver()
s.add(x + y == 42)
s.add(x > 0)
s.add(y > 0)

if s.check() == sat:
    m = s.model()
    print(f"x = {m[x]}, y = {m[y]}")
```

### 2.2 Common CTF Patterns

```python
# Serial key validation: each char satisfies constraints
key = [BitVec(f'k{i}', 8) for i in range(16)]
s = Solver()
for k in key:
    s.add(k >= 0x30, k <= 0x7a)  # alphanumeric-ish

# XOR key recovery
plaintext = b"known_plaintext"
ciphertext = b"\x12\x34..."
key_byte = BitVec('key', 8)
s = Solver()
for p, c in zip(plaintext, ciphertext):
    s.add(p ^ key_byte == c)

# System of linear equations (modular)
a, b, c = BitVecs('a b c', 32)
s = Solver()
s.add(3*a + 5*b + 7*c == 0x12345678)
s.add(2*a + 4*b + 6*c == 0xDEADBEEF)
s.add(a ^ b ^ c == 0xCAFEBABE)
```

### 2.3 Optimization

```python
from z3 import Optimize

opt = Optimize()
x = BitVec('x', 32)
opt.add(x > 0)
opt.add(x < 1000)
opt.minimize(x)  # find smallest satisfying value
opt.check()
print(opt.model())
```

---

## 3. UNICORN ENGINE — CODE EMULATION

### 3.1 Basic Setup

```python
from unicorn import *
from unicorn.x86_const import *
from capstone import Cs, CS_ARCH_X86, CS_MODE_64

mu = Uc(UC_ARCH_X86, UC_MODE_64)

CODE_ADDR = 0x400000
STACK_ADDR = 0x7fff0000
STACK_SIZE = 0x10000

mu.mem_map(CODE_ADDR, 0x10000)
mu.mem_map(STACK_ADDR, STACK_SIZE)

mu.mem_write(CODE_ADDR, code_bytes)
mu.reg_write(UC_X86_REG_RSP, STACK_ADDR + STACK_SIZE - 0x1000)
mu.reg_write(UC_X86_REG_RBP, STACK_ADDR + STACK_SIZE - 0x1000)

mu.emu_start(CODE_ADDR, CODE_ADDR + len(code_bytes))

result = mu.reg_read(UC_X86_REG_RAX)
```

### 3.2 Hooking Memory & Instructions

```python
# Hook memory access
def hook_mem(uc, access, address, size, value, user_data):
    if access == UC_MEM_WRITE:
        print(f"Write {value:#x} to {address:#x}")
    elif access == UC_MEM_READ:
        print(f"Read from {address:#x}")

mu.hook_add(UC_HOOK_MEM_READ | UC_HOOK_MEM_WRITE, hook_mem)

# Hook specific instruction (for tracing)
def hook_code(uc, address, size, user_data):
    code = uc.mem_read(address, size)
    md = Cs(CS_ARCH_X86, CS_MODE_64)
    for insn in md.disasm(bytes(code), address):
        print(f"  {insn.address:#x}: {insn.mnemonic} {insn.op_str}")

mu.hook_add(UC_HOOK_CODE, hook_code)
```

### 3.3 Use Cases

| Use Case | Approach |
|---|---|
| Unpack shellcode | Map shellcode, emulate, dump decoded payload |
| Decrypt strings | Emulate decryption function with controlled inputs |
| Brute-force short keys | Loop emulation with different key inputs |
| Analyze obfuscated function | Emulate function, observe register/memory state |
| Firmware code emulation | Map firmware memory layout, emulate routines |

---

## 4. ANGR EXPLORATION STRATEGIES

### 4.1 find/avoid

```python
simgr.explore(
    find=lambda s: b"Correct" in s.posix.dumps(1),   # stdout contains "Correct"
    avoid=lambda s: b"Wrong" in s.posix.dumps(1)      # avoid "Wrong" output
)
```

### 4.2 Managing Path Explosion

| Strategy | Implementation |
|---|---|
| Constrain input space | Add constraints (printable, length limits) |
| Avoid dead-end paths | Use `avoid=` for known failure addresses |
| Hook complex functions | Replace with simplified SimProcedure |
| Limit loop iterations | `state.options.add(angr.options.LAZY_SOLVES)` |
| Use veritesting | `simgr.explore(..., technique=angr.exploration_techniques.Veritesting())` |
| DFS instead of BFS | `simgr.use_technique(angr.exploration_techniques.DFS())` |
| Timeout per path | `simgr.explore(..., num_find=1)` + timeout wrapper |

### 4.3 Concrete + Symbolic Hybrid

```python
state = proj.factory.entry_state(
    add_options={angr.options.UNICORN}  # use Unicorn for concrete regions
)
```

This dramatically speeds up execution: concrete code runs natively via Unicorn, switching to symbolic only when symbolic variables are involved.

---

## 5. PRACTICAL WORKFLOW

### 5.1 CTF Binary Solving Workflow

```
1. Static analysis: identify input method, success/fail conditions
   └─ Find "Correct" / "Wrong" strings → get their xref addresses

2. Choose tool:
   ├─ Pure math (no binary needed) → Z3
   ├─ Small binary, clear success/fail → angr explore
   └─ Specific function to emulate → Unicorn

3. Set up symbolic input:
   ├─ stdin → claripy.BVS + entry_state(stdin=)
   ├─ argv → full_init_state(args=[...])
   ├─ file input → SimFile
   └─ specific memory → state.memory.store(addr, sym)

4. Hook problematic functions:
   ├─ printf/puts → SimProcedure or no-op
   ├─ scanf → custom handler
   ├─ time/random → return concrete value
   └─ anti-debug → skip entirely

5. Explore and extract:
   └─ simgr.explore(find=, avoid=) → solver.eval()
```

---

## 6. DECISION TREE

```
Need to solve a reversing challenge?
│
├─ Is the challenge pure math / equations?
│  └─ Yes → Z3
│     ├─ Linear equations → BitVec + Solver
│     ├─ Modular arithmetic → BitVec (natural mod 2^n)
│     ├─ Boolean logic → Bool + Solver
│     └─ Optimization → Optimize + minimize/maximize
│
├─ Is it a compiled binary with clear success/fail?
│  └─ Yes → angr
│     ├─ Input via stdin → symbolic stdin
│     ├─ Input via argv → full_init_state with symbolic args
│     ├─ Input via file → SimFile
│     ├─ Path explosion → add constraints, avoid paths, hook loops
│     └─ Complex library calls → hook with SimProcedure
│
├─ Need to emulate a specific function/region?
│  └─ Yes → Unicorn Engine
│     ├─ Decryption routine → map code + data, emulate, read result
│     ├─ Shellcode analysis → map shellcode, hook syscalls
│     └─ Key schedule → emulate with different inputs
│
├─ Need to analyze firmware / exotic arch?
│  └─ Yes → Qiling (full system emulation with OS support)
│
├─ Binary has VM protection?
│  └─ angr for handler analysis + Z3 for bytecode constraints
│
└─ None of the above working?
   ├─ Combine: Unicorn for concrete regions + Z3 for constraints
   ├─ Manual reverse engineering with debugger
   └─ Side-channel approach (timing, power analysis for hardware)
```

---

## 7. COMMON PITFALLS & FIXES

| Problem | Cause | Fix |
|---|---|---|
| angr hangs forever | Path explosion in loops | Add `avoid=` for loop-back edges, or hook the loop |
| Z3 returns `unknown` | Non-linear constraints too complex | Simplify, split into sub-problems, use `set_param("timeout", 5000)` |
| Unicorn crashes on syscall | Syscall not handled | Hook syscall interrupt, handle or skip |
| angr wrong result | Incorrect state initialization | Verify initial memory layout matches actual binary |
| Symbolic memory too large | Unbounded symbolic reads | Concretize array indices where possible |
| SimProcedure wrong types | Argument type mismatch | Check calling convention (cdecl vs fastcall) |
| angr can't load binary | Missing libraries | Use `auto_load_libs=False` + hook needed symbols |

---

## 8. TOOL VERSIONS & INSTALLATION

```bash
# angr (Python 3.8+)
pip install angr

# Z3
pip install z3-solver

# Unicorn Engine
pip install unicorn

# Capstone (disassembly, pairs with Unicorn)
pip install capstone

# Keystone (assembly)
pip install keystone-engine
```
