#!/usr/bin/env python3
"""
Cybermes System Diagnostics & Health Check (Doctor)
Cross-platform environment inspector and auto-repair utility.
"""

import sys
import os
import platform
import shutil
import subprocess
import argparse
from pathlib import Path

GREEN = "\033[92m"
YELLOW = "\033[93m"
RED = "\033[91m"
CYAN = "\033[96m"
BOLD = "\033[1m"
RESET = "\033[0m"

if sys.platform == "win32":
    try:
        sys.stdout.reconfigure(encoding="utf-8", errors="replace")
        sys.stderr.reconfigure(encoding="utf-8", errors="replace")
    except Exception:
        pass
    os.system("")

def print_header(title: str):
    print(f"\n{BOLD}{CYAN}{'='*60}{RESET}")
    print(f"{BOLD}{CYAN}  {title}{RESET}")
    print(f"{BOLD}{CYAN}{'='*60}{RESET}")

def print_status(status: str, msg: str, detail: str = ""):
    badge_map = {
        "ok": f"{GREEN}[OK]{RESET}",
        "warn": f"{YELLOW}[WARN]{RESET}",
        "fail": f"{RED}[FAIL]{RESET}",
        "fixed": f"{CYAN}[FIXED]{RESET}",
    }
    badge = badge_map.get(status.lower(), f"[{status.upper()}]")
    detail_str = f" ({detail})" if detail else ""
    print(f"  {badge} {msg}{detail_str}")

def check_python() -> tuple[int, int, int]:
    print_header("1. Python Environment")
    passed, warns, fails = 0, 0, 0
    
    py_ver = f"{sys.version_info.major}.{sys.version_info.minor}.{sys.version_info.micro}"
    if sys.version_info >= (3, 10):
        print_status("ok", "Python Version", f"{py_ver} >= 3.10")
        passed += 1
    else:
        print_status("fail", "Python Version", f"{py_ver} (Requires 3.10+)")
        fails += 1

    in_venv = hasattr(sys, 'real_prefix') or (hasattr(sys, 'base_prefix') and sys.base_prefix != sys.prefix)
    if in_venv:
        print_status("ok", "Virtual Environment", "Active")
        passed += 1
    else:
        print_status("warn", "Virtual Environment", "Not active in current shell")
        warns += 1

    core_pkgs = ["requests", "yaml", "jinja2", "rich", "markdown"]
    for pkg in core_pkgs:
        try:
            __import__(pkg)
            print_status("ok", f"Package '{pkg}'", "Installed")
            passed += 1
        except ImportError:
            print_status("warn", f"Package '{pkg}'", "Not installed")
            warns += 1

    return passed, warns, fails

def check_directories(root_dir: Path, auto_fix: bool = False) -> tuple[int, int, int]:
    print_header("2. Workspace Directories & Permissions")
    passed, warns, fails = 0, 0, 0
    dirs = ["reports", "recon", "output", "logs", "targets", "skills", "tools/bin"]

    for d in dirs:
        dir_path = root_dir / d
        if not dir_path.exists():
            if auto_fix:
                try:
                    dir_path.mkdir(parents=True, exist_ok=True)
                    print_status("fixed", f"Directory '{d}'", "Created")
                    passed += 1
                except Exception as e:
                    print_status("fail", f"Directory '{d}'", f"Failed to create: {e}")
                    fails += 1
            else:
                print_status("fail", f"Directory '{d}'", "Missing (Use --fix to create)")
                fails += 1
        else:
            try:
                test_file = dir_path / ".perm_test"
                test_file.write_text("test", encoding="utf-8")
                test_file.unlink()
                print_status("ok", f"Directory '{d}'", "Writable")
                passed += 1
            except Exception as e:
                print_status("fail", f"Directory '{d}'", f"Not writable: {e}")
                fails += 1

    return passed, warns, fails

def check_config(root_dir: Path, auto_fix: bool = False) -> tuple[int, int, int]:
    print_header("3. Configuration & Credentials")
    passed, warns, fails = 0, 0, 0
    env_file = root_dir / ".env"
    hermes_dir = root_dir / ".hermes"
    config_file = hermes_dir / "config.yaml"
    auth_file = hermes_dir / "auth.json"
    hermes_env_file = hermes_dir / ".env"

    # 1. Root .env validation
    if env_file.is_file():
        print_status("ok", "Environment File (.env)", "Present")
        passed += 1
        # Sanity check OpenRouter Base URL
        try:
            env_content = env_file.read_text(encoding="utf-8", errors="ignore")
            for line in env_content.splitlines():
                line = line.strip()
                if line.startswith("OPENROUTER_BASE_URL=") and not line.startswith("#"):
                    val = line.split("=", 1)[1].strip()
                    if "localhost" in val or "127.0.0.1" in val:
                        print_status("warn", "OPENROUTER_BASE_URL", f"Points to {val} (overrides official endpoint)")
                        warns += 1
        except Exception:
            pass
    elif env_file.is_dir():
        print_status("fail", "Environment File (.env)", "Path is a directory, expected file")
        fails += 1
    else:
        if auto_fix and (root_dir / ".env.example").exists():
            shutil.copy(root_dir / ".env.example", env_file)
            print_status("fixed", "Environment File (.env)", "Generated from .env.example")
            passed += 1
        else:
            print_status("warn", "Environment File (.env)", "Missing (Copy from .env.example)")
            warns += 1

    # 2. .hermes/config.yaml validation (Docker directory trap protection)
    if config_file.is_dir():
        print_status("fail", "Config File (.hermes/config.yaml)", "Directory trap detected!")
        if auto_fix:
            shutil.rmtree(config_file)
            if (hermes_dir / "config.yaml.example").exists():
                shutil.copy(hermes_dir / "config.yaml.example", config_file)
            print_status("fixed", "Config File (.hermes/config.yaml)", "Purged directory & restored file")
            passed += 1
        else:
            fails += 1
    elif config_file.is_file():
        print_status("ok", "Config File (.hermes/config.yaml)", "Present")
        passed += 1
    else:
        if auto_fix and (hermes_dir / "config.yaml.example").exists():
            hermes_dir.mkdir(parents=True, exist_ok=True)
            shutil.copy(hermes_dir / "config.yaml.example", config_file)
            print_status("fixed", "Config File (.hermes/config.yaml)", "Generated from example")
            passed += 1
        else:
            print_status("warn", "Config File (.hermes/config.yaml)", "Missing")
            warns += 1

    # 3. .hermes/auth.json validation
    if auth_file.is_dir():
        print_status("fail", "Auth File (.hermes/auth.json)", "Directory trap detected!")
        if auto_fix:
            shutil.rmtree(auth_file)
            auth_file.write_text("{}", encoding="utf-8")
            print_status("fixed", "Auth File (.hermes/auth.json)", "Purged directory & restored valid JSON")
            passed += 1
        else:
            fails += 1
    elif auth_file.is_file():
        print_status("ok", "Auth File (.hermes/auth.json)", "Present")
        passed += 1
    else:
        if auto_fix:
            hermes_dir.mkdir(parents=True, exist_ok=True)
            auth_file.write_text("{}", encoding="utf-8")
            print_status("fixed", "Auth File (.hermes/auth.json)", "Initialized with empty JSON")
            passed += 1
        else:
            print_status("warn", "Auth File (.hermes/auth.json)", "Missing")
            warns += 1

    # 4. .hermes/.env synchronization
    if auto_fix and env_file.is_file() and not hermes_env_file.exists():
        try:
            shutil.copy(env_file, hermes_env_file)
            print_status("fixed", "Hermes Internal Env (.hermes/.env)", "Synced from root .env")
        except Exception:
            pass

    return passed, warns, fails

def check_tools(root_dir: Path, auto_fix: bool = False) -> tuple[int, int, int]:
    print_header("4. Security Toolchain Availability")
    passed, warns, fails = 0, 0, 0
    
    tools_bin = str(root_dir / "tools" / "bin")
    if tools_bin not in os.environ.get("PATH", ""):
        sep = ";" if sys.platform == "win32" else ":"
        os.environ["PATH"] = f"{tools_bin}{sep}{os.environ.get('PATH', '')}"

    tools = [
        ("smart_pipe", "Stream Output Filter"),
        ("secret_scan", "Secret Scanner"),
        ("search_knowledge", "Offline Knowledge Search"),
        ("aggregate_reports", "Report Aggregator"),
        ("subfinder", "Subdomain Discovery"),
        ("httpx", "HTTP Prober"),
        ("katana", "Web Crawler"),
        ("nuclei", "Vulnerability Scanner"),
    ]

    missing_tools = []
    for name, desc in tools:
        found = shutil.which(name) or (sys.platform == "win32" and shutil.which(f"{name}.exe"))
        if found:
            print_status("ok", f"Tool: {name}", desc)
            passed += 1
        else:
            print_status("warn", f"Tool: {name}", f"{desc} not found")
            warns += 1
            missing_tools.append(name)

    if missing_tools and auto_fix:
        print(f"\n  {CYAN}⚡ Attempting automatic toolchain download...{RESET}")
        if sys.platform == "win32":
            updater = root_dir / "tools" / "update_tools.ps1"
            if updater.exists():
                subprocess.run(["powershell", "-ExecutionPolicy", "Bypass", "-File", str(updater)], check=False)
        else:
            updater = root_dir / "tools" / "update_tools.sh"
            if updater.exists():
                subprocess.run(["bash", str(updater)], check=False)

    return passed, warns, fails

def main():
    parser = argparse.ArgumentParser(description="Cybermes Diagnostics & Health Check")
    parser.add_argument("--fix", action="store_true", help="Automatically repair missing directories, configs, and toolchain")
    args = parser.parse_args()

    root_dir = Path(__file__).resolve().parent.parent
    os.chdir(root_dir)

    print_header("🛡️  Cybermes System Diagnostics & Health Check")
    print(f"System: {platform.system()} {platform.release()} ({platform.machine()})")
    print(f"Directory: {root_dir}")
    if args.fix:
        print(f"Mode: {CYAN}Auto-Repair Enabled (--fix){RESET}")

    p1, w1, f1 = check_python()
    p2, w2, f2 = check_directories(root_dir, auto_fix=args.fix)
    p3, w3, f3 = check_config(root_dir, auto_fix=args.fix)
    p4, w4, f4 = check_tools(root_dir, auto_fix=args.fix)

    total_pass = p1 + p2 + p3 + p4
    total_warn = w1 + w2 + w3 + w4
    total_fail = f1 + f2 + f3 + f4

    print_header("Diagnostics Summary")
    print(f"  {GREEN}Passed Checks:{RESET}   {total_pass}")
    print(f"  {YELLOW}Warnings:{RESET}        {total_warn}")
    print(f"  {RED}Failures:{RESET}        {total_fail}\n")

    if total_fail == 0:
        print(f"  {GREEN}{BOLD}✓ System is healthy and ready to run Cybermes!{RESET}\n")
    else:
        print(f"  {RED}{BOLD}! Some requirements need attention. Run with --fix or check setup guide.{RESET}\n")

if __name__ == "__main__":
    main()
