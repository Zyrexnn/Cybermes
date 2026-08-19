---
name: volatility-cheatsheet
description: >-
  Volatility 2 and 3 command comparison table with common plugin sequences for specific investigation types.
---

# VOLATILITY CHEATSHEET — Vol2 / Vol3 Command Reference

> Supplementary reference for [memory-forensics-volatility](./SKILL.md). Quick-access command tables.

---

## 1. COMMAND COMPARISON TABLE

### System Information

| Purpose | Volatility 2 | Volatility 3 |
|---|---|---|
| OS identification | `imageinfo` | `windows.info` |
| Kernel debug scan | `kdbgscan` | (auto) |
| OS banners | `imageinfo` | `banners.Banners` |

### Process Analysis

| Purpose | Volatility 2 | Volatility 3 |
|---|---|---|
| Process list (linked) | `pslist` | `windows.pslist` |
| Process list (scan) | `psscan` | `windows.psscan` |
| Process tree | `pstree` | `windows.pstree` |
| Process command line | `cmdline` | `windows.cmdline` |
| Process environment | `envars` | `windows.envars` |
| Process handles | `handles` | `windows.handles` |
| Process privileges | `privs` | `windows.privileges` |
| Process SIDs | `getsids` | `windows.getsids` |

### Memory / DLL

| Purpose | Volatility 2 | Volatility 3 |
|---|---|---|
| Loaded DLLs | `dlllist` | `windows.dlllist` |
| Unlinked DLL detection | `ldrmodules` | `windows.ldrmodules` |
| Code injection | `malfind` | `windows.malfind` |
| VAD tree | `vadinfo` | `windows.vadinfo` |
| Dump process memory | `memdump -p PID` | `windows.memmap --dump --pid PID` |
| Dump DLLs | `dlldump -p PID` | `windows.dlllist --pid PID --dump` |

### Network

| Purpose | Volatility 2 | Volatility 3 |
|---|---|---|
| All connections (Vista+) | `netscan` | `windows.netscan` |
| Active connections (XP) | `connections` | N/A |
| Listening sockets (XP) | `sockets` | N/A |
| Connection scan (XP) | `connscan` | N/A |
| Net stats | N/A | `windows.netstat` |

### Credentials

| Purpose | Volatility 2 | Volatility 3 |
|---|---|---|
| SAM hashes | `hashdump` | `windows.hashdump` |
| LSA secrets | `lsadump` | `windows.lsadump` |
| Cached creds | `cachedump` | `windows.cachedump` |
| Mimikatz | `mimikatz` (plugin) | N/A (use hashdump/lsadump) |

### File System

| Purpose | Volatility 2 | Volatility 3 |
|---|---|---|
| File scan | `filescan` | `windows.filescan` |
| File dump | `dumpfiles -Q OFFSET` | `windows.dumpfiles --virtaddr ADDR` |
| MFT analysis | `mftparser` | `windows.mftscan` |

### Registry

| Purpose | Volatility 2 | Volatility 3 |
|---|---|---|
| List hives | `hivelist` | `windows.registry.hivelist` |
| Print key | `printkey -K "path"` | `windows.registry.printkey --key "path"` |
| User assist | `userassist` | `windows.registry.userassist` |
| Shimcache | `shimcache` | N/A (third-party) |

### History / User Activity

| Purpose | Volatility 2 | Volatility 3 |
|---|---|---|
| CMD history | `cmdscan` | `windows.cmdline` |
| Console output | `consoles` | `windows.consoles` |
| Clipboard | `clipboard` | `windows.clipboard` |
| Screenshots | `screenshot` | N/A |
| IE history | `iehistory` | N/A |

### Rootkit / Hooking

| Purpose | Volatility 2 | Volatility 3 |
|---|---|---|
| SSDT hooks | `ssdt` | `windows.ssdt` |
| IDT hooks | `idt` | N/A |
| Driver scan | `driverscan` | `windows.driverscan` |
| Kernel modules | `modules` | `windows.modules` |
| Moddump | `modscan` | `windows.modscan` |
| Callbacks | `callbacks` | `windows.callbacks` |

### Timeline

| Purpose | Volatility 2 | Volatility 3 |
|---|---|---|
| Generate timeline | `timeliner` | `timeliner.Timeliner` |

---

## 2. COMMON INVESTIGATION SEQUENCES

### Malware Triage (Quick)

```bash
# Vol3
vol -f mem.raw windows.info
vol -f mem.raw windows.pstree
vol -f mem.raw windows.malfind
vol -f mem.raw windows.netscan
```

### Full Malware Investigation

```bash
# Vol3
vol -f mem.raw windows.info
vol -f mem.raw windows.pslist
vol -f mem.raw windows.psscan         # compare with pslist
vol -f mem.raw windows.pstree         # check parent-child
vol -f mem.raw windows.cmdline        # command lines
vol -f mem.raw windows.netscan        # C2 connections
vol -f mem.raw windows.malfind        # code injection
vol -f mem.raw windows.dlllist --pid SUSPICIOUS_PID
vol -f mem.raw windows.handles --pid SUSPICIOUS_PID
vol -f mem.raw windows.filescan | grep -i "suspicious_name"
vol -f mem.raw windows.dumpfiles --virtaddr OFFSET
```

### Credential Extraction

```bash
# Vol3
vol -f mem.raw windows.hashdump
vol -f mem.raw windows.lsadump
vol -f mem.raw windows.cachedump
```

### Incident Response Timeline

```bash
# Vol3
vol -f mem.raw windows.info
vol -f mem.raw timeliner.Timeliner
vol -f mem.raw windows.cmdline
vol -f mem.raw windows.registry.userassist
vol -f mem.raw windows.netscan
vol -f mem.raw windows.filescan
```

### CTF Challenge

```bash
# Vol3
vol -f mem.raw windows.info           # or banners.Banners
vol -f mem.raw windows.filescan | grep -iE "flag|secret|password|key"
vol -f mem.raw windows.cmdline        # look for typed commands
vol -f mem.raw windows.pslist         # find interesting processes
vol -f mem.raw windows.netscan        # find interesting connections
vol -f mem.raw windows.dumpfiles --virtaddr OFFSET  # extract files
vol -f mem.raw windows.hashdump       # extract hashes
```

### Linux Rootkit Detection

```bash
# Vol3
vol -f mem.lime linux.pslist
vol -f mem.lime linux.check_syscall
vol -f mem.lime linux.check_afinfo
vol -f mem.lime linux.tty_check
vol -f mem.lime linux.bash
vol -f mem.lime linux.elfs            # find injected ELFs
```

---

## 3. PROFILE MANAGEMENT (Vol2)

```bash
# List available profiles
vol.py --info | grep -i "Profile"

# Common Windows profiles
Win7SP1x64, Win10x64_19041, Win10x64_17763, WinXPSP3x86, Win2016x64_14393

# Determine correct profile
vol.py -f mem.raw imageinfo
# Use "Suggested Profile(s)" from output

# Custom Linux profile
# On target system:
cd volatility/tools/linux && make
zip profile.zip module.dwarf /boot/System.map-$(uname -r)
cp profile.zip volatility/plugins/overlays/linux/
```

---

## 4. USEFUL GREP PATTERNS

```bash
# Find interesting files
vol -f mem.raw windows.filescan | grep -iE '\.(txt|doc|xls|pdf|kdbx|key|pem|conf|ini|bat|ps1)$'

# Find web-related files
vol -f mem.raw windows.filescan | grep -iE '\.(php|asp|jsp|html)$'

# Find executables in unusual locations
vol -f mem.raw windows.filescan | grep -iE '\\(temp|tmp|appdata|downloads)\\.*\.exe'

# Network connections to external IPs
vol -f mem.raw windows.netscan | grep -v "127.0.0.1\|0.0.0.0\|::1\|::"
```
