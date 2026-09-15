#!/usr/bin/env python3
"""
Cybermes Embedded Router Firmware Filesystem Static Auditor
===========================================================
Defensive static analysis utility to inspect extracted firmware filesystems
(e.g., squashfs, jffs2) against OWASP IoT Top 10 and FSTM Stage 5 checks.

Usage:
  python3 scripts/audit_firmware_fs.py --rootfs /path/to/extracted_rootfs [--json]
"""

import argparse
import json
import os
import re
import sys
from typing import Any, Dict, List


SUSPICIOUS_PASS_HASHES = [
    ("DES/MD5_EMPTY_OR_COMMON", re.compile(r"^[a-zA-Z0-9_.\-]+:(\$1\$[a-zA-Z0-9./]{8}\$[a-zA-Z0-9./]{22}|admin|password|123456):")),
    ("EMPTY_PASSWORD", re.compile(r"^[a-zA-Z0-9_.\-]+::")),
]

DANGEROUS_SERVICES = [
    ("TELNETD_ENABLED", re.compile(r"\btelnetd\b")),
    ("DROPBEAR_NO_AUTH", re.compile(r"\bdropbear\b.*-B\b")),
    ("HTTPD_NO_AUTH", re.compile(r"\bhttpd\b.*-u\b")),
]


def audit_firmware_rootfs(rootfs: str) -> Dict[str, Any]:
    """Scan extracted firmware filesystem for insecure defaults and exposed credentials."""
    findings: List[Dict[str, Any]] = []

    if not os.path.isdir(rootfs):
        return {"error": f"Path '{rootfs}' is not a valid directory"}

    # 1. Audit /etc/shadow and /etc/passwd
    for rel_path in ["etc/shadow", "etc/passwd", "etc/config/shadow"]:
        full_path = os.path.join(rootfs, rel_path)
        if os.path.isfile(full_path):
            try:
                with open(full_path, "r", encoding="utf-8", errors="ignore") as f:
                    for line_num, line in enumerate(f, 1):
                        line = line.strip()
                        for check_name, pattern in SUSPICIOUS_PASS_HASHES:
                            if pattern.search(line):
                                findings.append({
                                    "check": "HARDCODED_OR_WEAK_ACCOUNT",
                                    "file": rel_path,
                                    "line": line_num,
                                    "severity": "CRITICAL" if "EMPTY" in check_name else "HIGH",
                                    "detail": f"Account with weak/empty password signature found: {line.split(':')[0]}"
                                })
            except Exception as e:
                pass

    # 2. Audit init scripts (/etc/init.d/) for unauthenticated daemon launches
    init_dir = os.path.join(rootfs, "etc/init.d")
    if os.path.isdir(init_dir):
        for fname in os.listdir(init_dir):
            fpath = os.path.join(init_dir, fname)
            if os.path.isfile(fpath):
                try:
                    with open(fpath, "r", encoding="utf-8", errors="ignore") as f:
                        content = f.read()
                        for check_name, pattern in DANGEROUS_SERVICES:
                            if pattern.search(content):
                                findings.append({
                                    "check": check_name,
                                    "file": f"etc/init.d/{fname}",
                                    "severity": "HIGH",
                                    "detail": f"Insecure service launch configuration detected: {check_name}"
                                })
                except Exception:
                    pass

    # 3. Check for private keys (*.pem, *.key) inside the filesystem
    key_pattern = re.compile(r"(\.pem|\.key|id_rsa|id_dsa|id_ecdsa)$", re.IGNORECASE)
    for root, _, files in os.walk(rootfs):
        for file in files:
            if key_pattern.search(file):
                rel_file = os.path.relpath(os.path.join(root, file), rootfs)
                findings.append({
                    "check": "HARDCODED_PRIVATE_KEY",
                    "file": rel_file,
                    "severity": "HIGH",
                    "detail": "Cryptographic key file bundled directly into filesystem image"
                })

    return {
        "rootfs": rootfs,
        "total_findings": len(findings),
        "findings": findings,
    }


def main():
    parser = argparse.ArgumentParser(description="Cybermes Embedded Firmware Static Auditor")
    parser.add_argument("--rootfs", required=True, help="Directory path to extracted root filesystem")
    parser.add_argument("--json", action="store_true", help="Output summary in JSON format")

    args = parser.parse_args()
    results = audit_firmware_rootfs(args.rootfs)

    if args.json:
        print(json.dumps(results, indent=2))
        return

    print("=" * 65)
    print("  📟 Cybermes Firmware Filesystem Static Audit Report")
    print("=" * 65)
    print(f"RootFS Path:  {results.get('rootfs', 'N/A')}")
    print(f"Total Issues: {results.get('total_findings', 0)}\n")

    if not results.get("findings"):
        print("✅ No critical default credentials or exposed keys detected.")
    else:
        for idx, f in enumerate(results["findings"], 1):
            print(f"[{idx}] Severity: {f['severity']}")
            print(f"    Check:    {f['check']}")
            print(f"    File:     {f['file']}" + (f":{f['line']}" if "line" in f else ""))
            print(f"    Details:  {f['detail']}")
            print()


if __name__ == "__main__":
    main()
