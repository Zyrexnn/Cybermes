# angr Cookbook — Ready-to-Use Script Patterns for CTF Challenges

> **AI LOAD INSTRUCTION**: Load this when you need drop-in angr script templates for common CTF challenge patterns. Each recipe includes the scenario description, complete script, and adaptation notes. Assumes the main [SKILL.md](./SKILL.md) is already loaded for angr fundamentals.

---

## Recipe 1: Basic find/avoid by Address

**Scenario**: Binary prints "Correct" at 0x401234, "Wrong" at 0x401300. Input via stdin.

```python
import angr

proj = angr.Project('./challenge', auto_load_libs=False)
state = proj.factory.entry_state()
simgr = proj.factory.simulation_manager(state)

simgr.explore(find=0x401234, avoid=[0x401300])

if simgr.found:
    print(simgr.found[0].posix.dumps(0))  # fd 0 = stdin
```

---

## Recipe 2: Find by stdout Content

**Scenario**: Success/failure messages not at fixed addresses (PIE binary or indirect calls).

```python
import angr

proj = angr.Project('./challenge', auto_load_libs=False)
state = proj.factory.entry_state()
simgr = proj.factory.simulation_manager(state)

simgr.explore(
    find=lambda s: b"Correct" in s.posix.dumps(1),
    avoid=lambda s: b"Wrong" in s.posix.dumps(1)
)

if simgr.found:
    print(simgr.found[0].posix.dumps(0))
```

---

## Recipe 3: Symbolic stdin with Printable Constraint

**Scenario**: Flag is 32 printable characters read via scanf/fgets.

```python
import angr
import claripy

proj = angr.Project('./challenge', auto_load_libs=False)

FLAG_LEN = 32
flag = claripy.BVS("flag", 8 * FLAG_LEN)

state = proj.factory.entry_state(stdin=flag)

for i in range(FLAG_LEN):
    byte = flag.get_byte(i)
    state.solver.add(byte >= 0x20)
    state.solver.add(byte <= 0x7e)

simgr = proj.factory.simulation_manager(state)
simgr.explore(find=0x401234, avoid=[0x401300])

if simgr.found:
    solution = simgr.found[0].solver.eval(flag, cast_to=bytes)
    print(f"Flag: {solution}")
```

---

## Recipe 4: Symbolic argv

**Scenario**: Binary takes flag as command-line argument `./challenge FLAG`.

```python
import angr
import claripy

proj = angr.Project('./challenge', auto_load_libs=False)

flag = claripy.BVS("flag", 8 * 32)
state = proj.factory.full_init_state(
    args=['./challenge', flag],
    add_options=angr.options.unicorn
)

for i in range(32):
    b = flag.get_byte(i)
    state.solver.add(b >= 0x20, b <= 0x7e)

simgr = proj.factory.simulation_manager(state)
simgr.explore(find=0x401234, avoid=[0x401300])

if simgr.found:
    print(simgr.found[0].solver.eval(flag, cast_to=bytes))
```

---

## Recipe 5: Symbolic File Input

**Scenario**: Binary reads flag from a file.

```python
import angr
import claripy

proj = angr.Project('./challenge', auto_load_libs=False)

flag = claripy.BVS("flag", 8 * 64)
sim_file = angr.SimFile("flag.txt", content=flag)

state = proj.factory.entry_state(fs={'/tmp/flag.txt': sim_file})

simgr = proj.factory.simulation_manager(state)
simgr.explore(find=0x401234, avoid=[0x401300])

if simgr.found:
    print(simgr.found[0].solver.eval(flag, cast_to=bytes))
```

---

## Recipe 6: Hook scanf

**Scenario**: Custom scanf that angr's default SimProcedure handles incorrectly.

```python
import angr
import claripy

class MyScanf(angr.SimProcedure):
    def run(self, fmt, ptr):
        buf = claripy.BVS("scanf_input", 8 * 32)
        self.state.memory.store(ptr, buf)
        self.state.globals['scanf_buf'] = buf
        return 1

proj = angr.Project('./challenge', auto_load_libs=False)
proj.hook_symbol('__isoc99_scanf', MyScanf())

state = proj.factory.entry_state()
simgr = proj.factory.simulation_manager(state)
simgr.explore(find=0x401234, avoid=[0x401300])

if simgr.found:
    found = simgr.found[0]
    buf = found.globals['scanf_buf']
    print(found.solver.eval(buf, cast_to=bytes))
```

---

## Recipe 7: Hook strcmp to Leak Expected Value

**Scenario**: Binary compares input against computed value. Hook strcmp to extract the expected string.

```python
import angr
import claripy

class LeakStrcmp(angr.SimProcedure):
    def run(self, s1, s2):
        # Read both strings up to null
        str1 = self.state.memory.load(s1, 64)
        str2 = self.state.memory.load(s2, 64)
        self.state.globals['cmp_s1'] = s1
        self.state.globals['cmp_s2'] = s2
        return claripy.If(str1 == str2, claripy.BVV(0, 32), claripy.BVV(1, 32))

proj = angr.Project('./challenge', auto_load_libs=False)
proj.hook_symbol('strcmp', LeakStrcmp())

state = proj.factory.entry_state()
simgr = proj.factory.simulation_manager(state)
simgr.explore(find=0x401234)

if simgr.found:
    found = simgr.found[0]
    s2_addr = found.solver.eval(found.globals['cmp_s2'])
    expected = found.memory.load(s2_addr, 32)
    print(found.solver.eval(expected, cast_to=bytes))
```

---

## Recipe 8: Start from Specific Function (skip initialization)

**Scenario**: Large binary, but the check function is known. Skip everything before it.

```python
import angr
import claripy

proj = angr.Project('./challenge', auto_load_libs=False)

CHECK_FUNC = 0x401500
BUF_ADDR = 0x600000

flag = claripy.BVS("flag", 8 * 32)

state = proj.factory.blank_state(addr=CHECK_FUNC)
state.memory.store(BUF_ADDR, flag)
state.regs.rdi = BUF_ADDR  # first arg = buffer pointer (System V ABI)
state.regs.rsi = 32         # second arg = length

state.stack_push(0xDEADBEEF)  # fake return address

simgr = proj.factory.simulation_manager(state)
simgr.explore(
    find=lambda s: b"OK" in s.posix.dumps(1),
    avoid=lambda s: b"FAIL" in s.posix.dumps(1)
)

if simgr.found:
    print(simgr.found[0].solver.eval(flag, cast_to=bytes))
```

---

## Recipe 9: Veritesting (Merge Paths at Branches)

**Scenario**: Path explosion due to many conditional branches (e.g., per-character check).

```python
import angr

proj = angr.Project('./challenge', auto_load_libs=False)
state = proj.factory.entry_state()
simgr = proj.factory.simulation_manager(state)

simgr.use_technique(angr.exploration_techniques.Veritesting())
simgr.explore(find=0x401234, avoid=[0x401300])

if simgr.found:
    print(simgr.found[0].posix.dumps(0))
```

---

## Recipe 10: Unicorn Concrete Execution Speedup

**Scenario**: Most code is concrete (no symbolic data), only small portion needs symbolic reasoning.

```python
import angr

proj = angr.Project('./challenge', auto_load_libs=False)
state = proj.factory.entry_state(
    add_options=angr.options.unicorn  # enable Unicorn backend
)
simgr = proj.factory.simulation_manager(state)
simgr.explore(find=0x401234, avoid=[0x401300])

if simgr.found:
    print(simgr.found[0].posix.dumps(0))
```

---

## Recipe 11: Multi-Stage Binary (Sequential Checks)

**Scenario**: Binary has multiple stages, each validating part of the flag.

```python
import angr
import claripy

proj = angr.Project('./challenge', auto_load_libs=False)

stages = [
    {"find": 0x401100, "avoid": [0x401150]},
    {"find": 0x401200, "avoid": [0x401250]},
    {"find": 0x401300, "avoid": [0x401350]},
]

flag = claripy.BVS("flag", 8 * 48)
state = proj.factory.entry_state(stdin=flag)

for i, stage in enumerate(stages):
    simgr = proj.factory.simulation_manager(state)
    simgr.explore(find=stage["find"], avoid=stage["avoid"])
    if simgr.found:
        state = simgr.found[0]
        partial = state.solver.eval(flag, cast_to=bytes)
        print(f"Stage {i+1}: {partial}")
    else:
        print(f"Stage {i+1} failed")
        break
```

---

## Recipe 12: Constrain Flag Format (CTF{...})

**Scenario**: Flag format is known: `flag{...}` with alphanumeric content.

```python
import angr
import claripy

proj = angr.Project('./challenge', auto_load_libs=False)

FLAG_LEN = 40
flag = claripy.BVS("flag", 8 * FLAG_LEN)
state = proj.factory.entry_state(stdin=flag)

prefix = b"flag{"
for i, c in enumerate(prefix):
    state.solver.add(flag.get_byte(i) == c)
state.solver.add(flag.get_byte(FLAG_LEN - 1) == ord('}'))

for i in range(len(prefix), FLAG_LEN - 1):
    b = flag.get_byte(i)
    state.solver.add(
        claripy.Or(
            claripy.And(b >= ord('0'), b <= ord('9')),
            claripy.And(b >= ord('a'), b <= ord('z')),
            claripy.And(b >= ord('A'), b <= ord('Z')),
            b == ord('_'), b == ord('-')
        )
    )

simgr = proj.factory.simulation_manager(state)
simgr.explore(find=0x401234, avoid=[0x401300])

if simgr.found:
    print(simgr.found[0].solver.eval(flag, cast_to=bytes))
```

---

## Recipe 13: Replace rand() with Concrete Value

**Scenario**: Binary uses `rand()` making execution non-deterministic.

```python
import angr

class ConcreteRand(angr.SimProcedure):
    def run(self):
        return 42  # deterministic

proj = angr.Project('./challenge', auto_load_libs=False)
proj.hook_symbol('rand', ConcreteRand())
proj.hook_symbol('srand', angr.SIM_PROCEDURES['libc']['srand']())

state = proj.factory.entry_state()
simgr = proj.factory.simulation_manager(state)
simgr.explore(find=0x401234)

if simgr.found:
    print(simgr.found[0].posix.dumps(0))
```

---

## Recipe 14: Timeout + DFS for Deep Binaries

**Scenario**: Binary has deep execution path, BFS runs out of memory.

```python
import angr
import signal

def timeout_handler(signum, frame):
    raise TimeoutError("angr timed out")

signal.signal(signal.SIGALRM, timeout_handler)
signal.alarm(300)  # 5-minute timeout

proj = angr.Project('./challenge', auto_load_libs=False)
state = proj.factory.entry_state()
simgr = proj.factory.simulation_manager(state)

simgr.use_technique(angr.exploration_techniques.DFS())

try:
    simgr.explore(find=0x401234, avoid=[0x401300])
    if simgr.found:
        print(simgr.found[0].posix.dumps(0))
except TimeoutError:
    print("Timed out — try narrowing search space")
```

---

## Recipe 15: Symbolic Memory at Global Buffer

**Scenario**: Input is read into a global buffer at a known address, but through a complex path.

```python
import angr
import claripy

proj = angr.Project('./challenge', auto_load_libs=False)

GLOBAL_BUF = 0x604080
CHECK_START = 0x401200
FLAG_LEN = 24

flag = claripy.BVS("flag", 8 * FLAG_LEN)
state = proj.factory.blank_state(addr=CHECK_START)
state.memory.store(GLOBAL_BUF, flag)

for i in range(FLAG_LEN):
    b = flag.get_byte(i)
    state.solver.add(b >= 0x20, b <= 0x7e)

simgr = proj.factory.simulation_manager(state)
simgr.explore(find=0x401300, avoid=[0x401400])

if simgr.found:
    print(simgr.found[0].solver.eval(flag, cast_to=bytes))
```

---

## Adaptation Checklist

When adapting any recipe to your specific challenge:

1. **Update addresses**: `find=`, `avoid=`, function addresses, buffer addresses
2. **Match input method**: stdin, argv, file, or memory — use the right recipe
3. **Set correct flag length**: Too short = can't find solution; too long = slow
4. **Add format constraints**: Known prefix, charset, null terminator
5. **Hook problematic functions**: Any function causing path explosion or crashes
6. **Choose exploration technique**: BFS (default) for short paths, DFS for deep, Veritesting for branchy
7. **Enable Unicorn**: `add_options=angr.options.unicorn` for speed on mostly-concrete paths
8. **Check architecture**: Ensure `auto_load_libs=False` doesn't break needed functionality
