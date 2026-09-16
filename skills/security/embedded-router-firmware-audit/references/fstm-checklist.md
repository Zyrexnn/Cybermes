# Firmware Security Testing Methodology (FSTM) Checklist

This checklist aligns testing procedures with the **OWASP IoT Top 10** and the **9-Stage Firmware Security Testing Methodology (FSTM)**.

---

## Stage-by-Stage Verification Checklist

### Stage 1: Information Gathering
- [ ] Identify device model, CPU architecture (MIPS, ARM, x86), and chipset vendor.
- [ ] Document available network interfaces (Ethernet, Wi-Fi, optical PON) and exposed ports.
- [ ] Review vendor security advisories and end-of-life (EOL) status.

### Stage 2: Firmware Acquisition
- [ ] Download latest and historical firmware binaries directly from official vendor support channels.
- [ ] Verify file hashes (SHA-256) against vendor-provided checksums if available.
- [ ] Inspect binary header for digital signature blocks.

### Stage 3: Binary Characteristics & Entropy
- [ ] Analyze binary structure with `file` and `hexdump -C -n 128 firmware.bin`.
- [ ] Perform entropy analysis using `binwalk -E firmware.bin` to distinguish between encrypted and compressed blocks.

### Stage 4: Filesystem Extraction
- [ ] Extract filesystem partitions using `binwalk -Me firmware.bin` or `unsquashfs`.
- [ ] Verify root filesystem integrity (presence of `/etc`, `/bin`, `/sbin`, `/www` or `/usr/www`).

### Stage 5: Static Analysis (OWASP IoT Top 10)
- [ ] **I1: Weak / Hardcoded Passwords**:
  - Run `python3 scripts/audit_firmware_fs.py --rootfs <extracted_dir>`.
  - Check `/etc/shadow`, `/etc/passwd` for static root hashes or empty passwords.
- [ ] **I2: Insecure Network Services**:
  - Review `/etc/init.d/` daemon scripts for unauthenticated `telnetd` or insecure `dropbear` flags.
- [ ] **I3: Insecure Ecosystem Interfaces**:
  - Review `/www` or `/cgi-bin` scripts for unauthenticated diagnostic APIs (`ping.cgi`, `traceroute.cgi`).
- [ ] **I4: Lack of Secure Update Mechanism**:
  - Inspect upgrade routines (e.g. `sysupgrade`, `fw_update`) to confirm cryptographic signature validation.
- [ ] **I5: Insecure / Outdated Components**:
  - Check versions of BusyBox, OpenSSL/WolfSSL, dnsmasq, and Linux kernel.

### Stage 6 & 7: Emulation & Dynamic Assessment
- [ ] If required, emulate core user-space binaries using QEMU (`qemu-mips-static -L <rootfs> <binary>`).
- [ ] Test web management input sanitization using safe parameter boundary testing.

### Stage 8 & 9: Verification, Remediation & Deliverables
- [ ] Document findings according to Cybermes reporting guidelines in `reports/<target_slug>/`.
- [ ] Map all identified flaws to standard CWE identifiers (e.g. CWE-798 for hardcoded credentials, CWE-78 for OS command injection).
