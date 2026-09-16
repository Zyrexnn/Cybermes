---
name: embedded-router-firmware-audit
description: Comprehensive security auditing methodology for embedded routers, CPE gateways, and IoT firmware images. Aligned with OWASP IoT Security Testing Guide (ISTG) and the 9-stage Firmware Security Testing Methodology (FSTM). Covers filesystem extraction, hardcoded secret mining, dangerous daemon checks, and web management console assessment.
---

# Embedded Router & Firmware Security Auditing Playbook

## 1. Overview & Framework Alignment
Embedded network equipment (SOHO routers, wireless access points, fiber CPE gateways) often run embedded Linux environments with legacy busybox utilities, customized CGI web servers, and third-party vendor SDKs.

This playbook synthesizes the **OWASP IoT Security Testing Guide (ISTG)** and the industry-standard **Firmware Security Testing Methodology (FSTM)** into a systematic, repeatable audit lifecycle.

---

## 2. The 9-Stage FSTM Audit Process

```text
[Stage 1: Info Gathering] ──> [Stage 2: Firmware Acquisition] ──> [Stage 3: Binary Inspection]
                                                                        │
┌───────────────────────────────────────────────────────────────────────┘
▼
[Stage 4: Extraction]    ──> [Stage 5: Static Analysis]       ──> [Stage 6: Emulation]
                                      │
┌─────────────────────────────────────┘
▼
[Stage 7: Dynamic Testing] ──> [Stage 8: Runtime Audit]        ──> [Stage 9: Remediation & Reporting]
```

### Stage 1 & 2: Information Gathering & Acquisition
- **Sources**: Vendor support download portals, OTA update mirrors, or hardware extraction (SPI Flash / UART dump).
- **Integrity Verification**: Verify SHA256 checksums and check whether the image is cryptographically signed (GPG/RSA header).

### Stage 3 & 4: Binary Inspection & Extraction
- **Entropy Analysis**: Identify encrypted vs compressed sections using `binwalk -E firmware.bin`.
- **Filesystem Carving**:
  ```bash
  binwalk -Me firmware.bin
  # Common filesystem targets: squashfs-root, jffs2-root, ubifs
  ```

### Stage 5: Static Filesystem Audit (OWASP IoT Top 10)
Run the Cybermes firmware static analyzer:
```bash
python3 scripts/audit_firmware_fs.py --rootfs _firmware.bin.extracted/squashfs-root/
```

1. **Credential & Secret Discovery (OWASP I1)**:
   - Check `/etc/shadow`, `/etc/passwd`, and `/etc/default/`.
   - Identify hardcoded DES / MD5 crypt hashes or empty password entries (`root::0:0:`).
   - Search for bundled private keys (`*.pem`, `*.key`, `id_rsa`) or hardcoded encryption passphrases.
2. **Insecure Network Services & Init Daemons (OWASP I2)**:
   - Inspect `/etc/init.d/`, `/etc/rc.local`, and `/etc/rc.d/`.
   - Identify daemons launching with insecure parameters:
     - `telnetd -l /bin/sh` (Root shell without authentication).
     - `dropbear -B` (Allow blank passwords).
     - `httpd -u` (Debug or development mode).
3. **Outdated Third-Party Components & SDKs (OWASP I5)**:
   - Check versions of `busybox`, `dnsmasq`, `miniupnpd`, and `lighttpd`.
   - Cross-reference known CVEs for kernel and C library versions (`uClibc`, `musl`, `glibc`).

### Stage 6 & 7: Emulation & Dynamic Web Management Audit
- **Emulation**: Emulate architecture (MIPS/ARM) using QEMU (`qemu-mips-static` or firmadyne/FAT).
- **Web Management Interface Audit**:
  - CGI scripts located under `/www/`, `/usr/www/`, or `/cgi-bin/`.
  - Common weaknesses: Unauthenticated command injection in diagnostic tools (`ping.cgi`, `traceroute.cgi`), path traversal in file download endpoints (`download.cgi?file=...`).

---

## 3. Remediation & Hardening Guidelines
1. **Remove Hardcoded Factory Credentials**:
   - Force unique, randomized passwords generated during factory provisioning (printed on device label), or require immediate password setup upon first boot.
2. **Disable Debug Interfaces in Production**:
   - Strip `telnetd`, test scripts, and debug endpoints from production build targets.
   - Disable serial console login without physical button activation or authentication.
3. **Cryptographic Signature Verification**:
   - Implement Hardware Root of Trust (Secure Boot).
   - Ensure the bootloader (U-Boot) verifies firmware digital signatures prior to executing the kernel.
