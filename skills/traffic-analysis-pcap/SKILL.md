---
name: traffic-analysis-pcap
description: >-
  Traffic analysis and PCAP forensics playbook. Use when analyzing network captures including Wireshark filters, protocol analysis (HTTP/DNS/FTP/SMTP/USB/WiFi), data extraction, covert channel detection, PCAP repair, TLS decryption, and tshark command-line analysis.
---

# SKILL: Traffic Analysis & PCAP — Expert Analysis Playbook

> **AI LOAD INSTRUCTION**: Expert traffic analysis and PCAP forensics techniques. Covers PCAP repair, Wireshark essential filters, protocol-specific analysis (HTTP, HTTPS/TLS, DNS, FTP, SMTP, USB HID, WiFi, ICMP), data extraction (file carving, credential harvesting, covert channels), NetworkMiner, and tshark CLI analysis. Base models miss USB keyboard decode patterns, DNS tunneling detection heuristics, and TLS decryption workflows.

## 0. RELATED ROUTING

Before going deep, consider loading:

- [memory-forensics-volatility](../memory-forensics-volatility/SKILL.md) for correlating memory artifacts with network traffic
- [steganography-techniques](../steganography-techniques/SKILL.md) for analyzing files extracted from traffic captures
- [network-protocol-attacks](../network-protocol-attacks/SKILL.md) for understanding attack patterns visible in captures
- [reverse-shell-techniques](../reverse-shell-techniques/SKILL.md) for identifying shell traffic in captures

---

## 1. PCAP REPAIR

```bash
pcapfix corrupted.pcap -o fixed.pcap           # repair corrupted PCAP
# Magic bytes: d4c3b2a1=pcap(LE), a1b2c3d4=pcap(BE), 0a0d0d0a=pcapng
editcap -F pcap capture.pcapng capture.pcap    # convert pcapng→pcap
mergecap -w merged.pcap file1.pcap file2.pcap  # merge captures
```

---

## 2. WIRESHARK ESSENTIAL FILTERS

### IP / Host Filters

```
ip.addr == 10.0.0.1                  # source or destination
ip.src == 10.0.0.1                   # source only
ip.dst == 10.0.0.1                   # destination only
ip.addr == 10.0.0.0/24              # subnet
!(ip.addr == 10.0.0.1)              # exclude host
```

### Protocol Filters

```
http                                  # all HTTP
dns                                   # all DNS
tcp                                   # all TCP
ftp                                   # all FTP
smtp                                  # all SMTP
tls                                   # all TLS/SSL
icmp                                  # all ICMP
arp                                   # all ARP
```

### TCP / Stream

```
tcp.stream eq 5                       # follow specific TCP stream
tcp.port == 80                        # traffic on port 80
tcp.flags.syn == 1 && tcp.flags.ack == 0   # SYN packets (connection starts)
tcp.analysis.retransmission           # retransmitted packets
tcp.len > 0                           # packets with payload
```

### HTTP

```
http.request.method == "POST"         # POST requests
http.request.method == "GET"          # GET requests
http.response.code == 200             # successful responses
http.response.code >= 400             # error responses
http.request.uri contains "login"     # URI contains string
http.host contains "target.com"       # specific host
http.content_type contains "json"     # JSON responses
http.cookie contains "session"        # session cookies
http.request.full_uri                 # show full URIs (column)
```

### DNS

```
dns.qry.name contains "evil.com"     # specific domain queries
dns.qry.type == 1                    # A records
dns.qry.type == 28                   # AAAA records
dns.qry.type == 16                   # TXT records
dns.flags.response == 1              # DNS responses only
dns.resp.len > 100                   # large DNS responses
```

### TLS

```
tls.handshake.type == 1              # Client Hello
tls.handshake.type == 2              # Server Hello
tls.handshake.extensions.server_name  # SNI (hostname)
tls.handshake.type == 11             # Certificate
```

### Content Search

```
frame contains "password"             # search in raw bytes
frame contains "flag{"                # CTF flag pattern
tcp contains "admin"                  # search in TCP payload
```

---

## 3. PROTOCOL ANALYSIS

### HTTP — Follow Stream & Extract

```
Right-click packet → Follow → TCP Stream
# Shows full HTTP request/response conversation

# File extraction:
# File → Export Objects → HTTP → Save All

# Useful filters for credential hunting:
http.request.method == "POST" && frame contains "password"
http.request.method == "POST" && frame contains "login"
http.authbasic                        # Basic auth (base64 encoded)
```

### HTTPS / TLS Decryption

```bash
# Method 1: SSLKEYLOGFILE (pre-master secrets from browser)
# Set environment variable BEFORE opening browser:
export SSLKEYLOGFILE=/tmp/sslkeys.log
firefox https://target.com

# Wireshark: Edit → Preferences → Protocols → TLS
# → (Pre)-Master-Secret log filename: /tmp/sslkeys.log

# Method 2: Server private key (for RSA key exchange only)
# Wireshark: Edit → Preferences → Protocols → TLS → RSA keys list
# → Add: IP, Port, Protocol, Key file (.pem)
```

### DNS — Tunneling Detection

```bash
# Indicators of DNS tunneling:
# 1. Unusually long subdomain names (>30 chars)
# 2. High volume of TXT record queries/responses
# 3. Consistent query patterns to same domain
# 4. Base32/Base64-like subdomain strings
# 5. High query frequency from single host

# Wireshark filter for suspicious DNS:
dns.qry.name.len > 50                # long query names
dns.qry.type == 16                   # TXT records (common for tunneling)
dns.resp.len > 512                   # large DNS responses

# tshark extraction:
tshark -r capture.pcap -Y "dns.qry.type==16" -T fields -e dns.qry.name
```

### FTP — Credential & File Extraction

```bash
# FTP credentials (plaintext)
# Filter: ftp.request.command == "USER" || ftp.request.command == "PASS"

# FTP file transfer reconstruction:
# FTP uses separate data channel (usually port 20 or dynamic)
# Follow TCP stream of data connection to extract file

# tshark:
tshark -r capture.pcap -Y "ftp.request.command==USER || ftp.request.command==PASS" -T fields -e ftp.request.arg
```

### SMTP — Email Content Extraction

```bash
# Follow TCP stream → MAIL FROM/RCPT TO/DATA sections
# Attachments: base64 in MIME → decode Content-Transfer-Encoding blocks
# Filters:
smtp.req.command == "AUTH"            # authentication (often base64)
smtp contains "Content-Disposition: attachment"   # attachments
```

### USB — Keyboard HID Capture Decode

```bash
# USB HID keyboard traffic: interrupt transfers with 8-byte data
# Filter: usb.transfer_type == 0x01

# Extract keystrokes:
tshark -r usb.pcap -Y "usb.capdata && usb.data_len == 8" -T fields -e usb.capdata > keystrokes.txt

# HID keycode layout: byte[0]=modifier, byte[2]=keycode
# 0x04=a..0x1d=z, 0x1e=1..0x27=0, 0x28=Enter, 0x2c=Space
# Use Python/online HID decoder to convert keycodes → text
```

### WiFi — WPA Handshake

```bash
# Capture: airodump-ng --bssid AP_MAC -w capture wlan0mon
# Convert + crack: hcxpcapngtool -o hash.hc22000 capture.pcap
hashcat -m 22000 hash.hc22000 wordlist.txt
# Deauth detection: wlan.fc.type_subtype == 0x0c
```

### ICMP — Data Exfiltration

```bash
# ICMP payload analysis
# Normal ping: 32 or 64 bytes of pattern data
# Exfiltration: meaningful data in ICMP payload

# Filter:
icmp && data.len > 48                 # unusual ICMP payload size
icmp.type == 8                        # echo requests

# Extract ICMP payloads:
tshark -r capture.pcap -Y "icmp.type==8" -T fields -e data.data
```

---

## 4. DATA EXTRACTION

### File Carving

```bash
# Wireshark: File → Export Objects
# Supported: HTTP, SMB, TFTP, IMF (email), DICOM

# Manual from reassembled stream:
# Follow TCP Stream → Show as Raw → Save As

# binwalk on exported stream data
binwalk -e exported_stream.bin
foremost -i exported_stream.bin -o carved/
```

### Credential Harvesting

```bash
# Plaintext: ftp || telnet || http.authbasic || smtp || pop || imap
# NTLM: ntlmssp.auth.username → extract challenge/response from NTLMSSP messages
# Hash format: user::domain:challenge:NTProofStr:blob → hashcat -m 5600
```

### Covert Channel Detection

Indicators: DNS with long subdomains, ICMP with large payloads, HTTP with encoded headers, regular beacon intervals (C2). Use `tshark -q -z io,stat,1` and `-z conv,tcp` for statistical anomaly detection.

---

## 5. NETWORKMINER

```bash
# Automated PCAP analysis: sudo apt install networkminer
# Open PCAP → auto-extracts: Files, Images, Credentials, Sessions, DNS
# Files tab: carved from HTTP/SMB/FTP | Credentials tab: plaintext creds
```

---

## 6. TSHARK COMMAND-LINE ANALYSIS

```bash
tshark -r capture.pcap -Y "http.request" -T fields -e http.host -e http.request.uri
tshark -r capture.pcap -Y "dns.flags.response==0" -T fields -e dns.qry.name | sort -u
tshark -r capture.pcap -Y "http.request.method==POST" -T fields -e http.file_data
tshark -r capture.pcap -q -z io,stat,1                # I/O graph
tshark -r capture.pcap -q -z conv,tcp                  # TCP conversations
tshark -r capture.pcap -q -z endpoints,ip              # IP endpoints
tshark -r capture.pcap -q -z io,phs                    # protocol hierarchy
tshark -r capture.pcap -q -z follow,tcp,ascii,0        # follow stream 0
tshark -r capture.pcap --export-objects http,/tmp/exported/
```

---

## 7. DECISION TREE

```
PCAP file for analysis
│
├── File won't open?
│   ├── Check magic bytes: xxd | head (§1)
│   ├── Repair: pcapfix (§1)
│   └── Convert: editcap pcapng→pcap (§1)
│
├── What's in the capture? (Quick overview)
│   ├── tshark -q -z io,phs (protocol hierarchy) (§6)
│   ├── tshark -q -z conv,tcp (conversations) (§6)
│   └── tshark -q -z endpoints,ip (endpoints) (§6)
│
├── HTTP traffic?
│   ├── Export objects: File → Export Objects → HTTP (§4)
│   ├── Credential hunt: POST + password/login filters (§3)
│   ├── Follow streams: interesting request/response pairs (§3)
│   └── Encrypted (HTTPS)? → need SSLKEYLOGFILE or RSA key (§3)
│
├── DNS traffic?
│   ├── Long subdomains? → DNS tunneling (§3)
│   ├── High TXT record volume? → DNS exfiltration (§3)
│   ├── Extract all queries: tshark -Y dns -T fields -e dns.qry.name (§6)
│   └── DNS rebinding? → check for alternating A record responses
│
├── FTP / Telnet / SMTP?
│   ├── Extract credentials (plaintext) (§3)
│   ├── Reconstruct file transfers (follow data stream) (§3)
│   └── Email content and attachments (base64 decode) (§3)
│
├── USB traffic?
│   ├── Keyboard HID → decode keystrokes (§3)
│   ├── Storage → extract transferred files
│   └── Check transfer_type and data_len fields
│
├── WiFi traffic?
│   ├── WPA handshake → crack with hashcat (§3)
│   ├── Deauth frames → detect attack (§3)
│   └── Probe requests → device fingerprinting
│
├── ICMP traffic?
│   ├── Large/variable payloads → data exfiltration (§3)
│   ├── Regular pattern → ICMP tunnel (§3)
│   └── Extract payloads: tshark -Y icmp -T fields -e data.data
│
├── Suspicious patterns?
│   ├── Regular beacon interval → C2 communication (§4)
│   ├── Unusual port/protocol combos → covert channel (§4)
│   ├── High volume to single external IP → data exfil (§4)
│   └── Encrypted traffic without SNI → suspicious tunnel
│
└── Need automated extraction?
    ├── NetworkMiner for files/creds/images (§5)
    ├── tshark --export-objects for HTTP/SMB files (§6)
    └── binwalk/foremost on exported streams (§4)
```
