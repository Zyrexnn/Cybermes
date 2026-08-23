#!/usr/bin/env python3
"""
=============================================================================
Cybermes Windows Compatibility & Environment Diagnostic Tool
Verifies system dependencies, Python environment, tools, and configurations.
=============================================================================
"""

import sys
import os
import platform
import subprocess
import shutil
from pathlib import Path

# Fix Windows console UTF-8 output encoding
if sys.platform == "win32":
    try:
        sys.stdout.reconfigure(encoding="utf-8", errors="replace")
        sys.stderr.reconfigure(encoding="utf-8", errors="replace")
    except Exception:
        pass
    os.system("")

GREEN = "\033[92m"
YELLOW = "\033[93m"
RED = "\033[91m"
CYAN = "\033[96m"
BOLD = "\033[1m"
RESET = "\033[0m"

def print_header(title):
    print(f"\n{BOLD}{CYAN}{'='*60}{RESET}")
    print(f"{BOLD}{CYAN}  {title}{RESET}")
    print(f"{BOLD}{CYAN}{'='*60}{RESET}")

def check_mark(status, msg, detail=""):
    if status is True:
        print(f"  {GREEN}[OK]{RESET} {msg} {f'({detail})' if detail else ''}")
    elif status == "warn":
        print(f"  {YELLOW}[WARN]{RESET} {msg} {f'({detail})' if detail else ''}")
    else:
        print(f"  {RED}[FAIL]{RESET} {msg} {f'({detail})' if detail else ''}")

def main():
    root_dir = Path(__file__).resolve().parent.parent
    os.chdir(root_dir)

    print_header("Cybermes Windows System & Diagnostics Check")
    print(f"Root Directory: {root_dir}")

    passed = 0
    warnings = 0
    failed = 0

    # 1. OS & Architecture Check
    print_header("1. Operating System & Architecture")
    is_windows = platform.system() == "Windows"
    os_ver = f"{platform.system()} {platform.release()} ({platform.version()})"
    is_64bit = sys.maxsize > 2**32

    if is_windows:
        check_mark(True, "Operating System", os_ver)
        passed += 1
    else:
        check_mark(True, "Host OS detected", os_ver)
        passed += 1

    if is_64bit:
        check_mark(True, "Architecture", "64-bit")
        passed += 1
    else:
        check_mark(False, "Architecture", "32-bit (64-bit recommended for security binaries)")
        failed += 1

    # 2. Python Environment
    print_header("2. Python Environment")
    py_ver = f"{sys.version_info.major}.{sys.version_info.minor}.{sys.version_info.micro}"
    if sys.version_info >= (3, 10):
        check_mark(True, "Python Version", f"{py_ver} (>= 3.10 required)")
        passed += 1
    else:
        check_mark(False, "Python Version", f"{py_ver} (Needs 3.10+)")
        failed += 1

    # Check venv
    in_venv = hasattr(sys, 'real_prefix') or (hasattr(sys, 'base_prefix') and sys.base_prefix != sys.prefix)
    venv_exists = (root_dir / "venv").exists()
    if in_venv:
        check_mark(True, "Active Environment", f"Virtualenv Active ({sys.prefix})")
        passed += 1
    elif venv_exists:
        check_mark("warn", "Virtualenv Directory Found", "Not active in current shell. Use '. .\\env.ps1' or hermes.bat")
        warnings += 1
    else:
        check_mark("warn", "Virtualenv Directory Missing", "Run '.\\setup_windows.ps1' to generate venv")
        warnings += 1

    # Check packages
    core_pkgs = ["requests", "yaml", "jinja2", "rich", "pydantic"]
    for pkg in core_pkgs:
        try:
            __import__(pkg)
            check_mark(True, f"Python Package '{pkg}'", "Installed")
            passed += 1
        except ImportError:
            check_mark("warn", f"Python Package '{pkg}'", "Not installed in current interpreter")
            warnings += 1

    # 3. Workspace Folders & Permissions
    print_header("3. Workspace Directories & Permissions")
    required_dirs = ["reports", "recon", "output", "logs", "targets", "skills", "tools/bin"]
    for d in required_dirs:
        dir_path = root_dir / d
        if not dir_path.exists():
            try:
                dir_path.mkdir(parents=True, exist_ok=True)
                check_mark(True, f"Directory: {d}", "Created")
                passed += 1
            except Exception as e:
                check_mark(False, f"Directory: {d}", f"Failed to create: {e}")
                failed += 1
        else:
            try:
                test_file = dir_path / ".perm_check"
                test_file.write_text("test", encoding="utf-8")
                test_file.unlink()
                check_mark(True, f"Directory: {d}", "Writable")
                passed += 1
            except Exception as e:
                check_mark(False, f"Directory: {d}", f"Not writable: {e}")
                failed += 1

    # 4. Configuration & Environment Files
    print_header("4. Configuration Files")
    env_file = root_dir / ".env"
    config_file = root_dir / ".hermes/config.yaml"

    if env_file.exists():
        check_mark(True, "Environment File (.env)", "Present")
        passed += 1
    else:
        check_mark("warn", "Environment File (.env)", "Missing (Copy from .env.example)")
        warnings += 1

    if config_file.exists():
        check_mark(True, "Hermes Config (.hermes/config.yaml)", "Present")
        passed += 1
    else:
        check_mark("warn", "Hermes Config (.hermes/config.yaml)", "Missing (Run setup_windows.ps1)")
        warnings += 1

    # 5. Security Toolchain Availability
    print_header("5. Security Toolchain & Tools Availability")
    tools = [
        ("git", "Version Control"),
        ("node", "Node.js (for Browser MCP)"),
        ("npm", "Node Package Manager"),
        ("docker", "Docker Container Runtime"),
        ("nmap", "Network Scanner"),
        ("subfinder", "Subdomain Discovery"),
        ("httpx", "HTTP Prober"),
        ("katana", "Web Crawler"),
        ("gau", "URL Extractor"),
        ("ffuf", "Web Fuzzer"),
        ("nuclei", "Vulnerability Scanner"),
        ("sqlmap", "SQL Injection Tool"),
        ("dalfox", "XSS Scanner"),
        ("rg", "Ripgrep Search Engine")
    ]

    tools_bin = str(root_dir / "tools" / "bin")
    current_path = os.environ.get("PATH", "")
    if tools_bin not in current_path:
        os.environ["PATH"] = f"{tools_bin};{current_path}"

    for tool_name, desc in tools:
        found = shutil.which(tool_name) or shutil.which(f"{tool_name}.exe")
        if found:
            check_mark(True, f"Tool: {tool_name}", f"{desc} -> Found")
            passed += 1
        else:
            check_mark("warn", f"Tool: {tool_name}", f"{desc} -> Not in PATH (Optional / Fallback available)")
            warnings += 1

    # Summary
    print_header("Diagnostics Summary")
    print(f"  {GREEN}Passed Checks:{RESET}   {passed}")
    print(f"  {YELLOW}Warnings/Tips:{RESET}   {warnings}")
    print(f"  {RED}Failures:{RESET}        {failed}")
    print()

    if failed == 0:
        print(f"  {GREEN}{BOLD}[OK] System is ready to run Cybermes!{RESET}")
    else:
        print(f"  {RED}{BOLD}[!] Some requirements need attention. Please review above.{RESET}")
    print()

if __name__ == "__main__":
    main()
