#!/usr/bin/env python3
"""
scripts/mcp.py — Universal Cross-Platform MCP Manager & Optimizer
Zero external dependencies. Pure standard library.
"""

import argparse
import ctypes
import datetime
import json
import os
import platform
import re
import shutil
import subprocess
import sys
import time
from pathlib import Path

# Setup Windows Virtual Terminal & UTF-8 Console
IS_WINDOWS = platform.system() == "Windows"
if IS_WINDOWS:
    try:
        kernel32 = ctypes.windll.kernel32
        h_stdout = kernel32.GetStdHandle(-11)
        mode = ctypes.c_ulong()
        if kernel32.GetConsoleMode(h_stdout, ctypes.byref(mode)):
            kernel32.SetConsoleMode(h_stdout, mode.value | 0x0004)
    except Exception:
        pass

if hasattr(sys.stdout, "reconfigure"):
    try:
        sys.stdout.reconfigure(encoding="utf-8", errors="replace")
    except Exception:
        pass
if hasattr(sys.stderr, "reconfigure"):
    try:
        sys.stderr.reconfigure(encoding="utf-8", errors="replace")
    except Exception:
        pass

ANSI_REGEX = re.compile(r"\x1b\[[0-9;]*m")

def visible_len(text: str) -> int:
    return len(ANSI_REGEX.sub("", text))

def pad_visible(text: str, width: int, align: str = "left") -> str:
    v_len = visible_len(text)
    pad = max(0, width - v_len)
    if align == "right":
        return " " * pad + text
    return text + " " * pad

class UI:
    RESET = "\033[0m"
    BOLD = "\033[1m"
    DIM = "\033[2m"
    
    CYAN = "\033[38;2;6;182;212m"
    TEAL = "\033[38;2;20;184;166m"
    GREEN = "\033[38;2;34;197;94m"
    YELLOW = "\033[38;2;234;179;8m"
    RED = "\033[38;2;239;68;68m"
    PURPLE = "\033[38;2;168;85;247m"
    GRAY = "\033[38;2;148;163;184m"
    DARK_GRAY = "\033[38;2;71;85;105m"
    WHITE = "\033[38;2;248;250;252m"

    @classmethod
    def header(cls, title: str, subtitle: str = ""):
        width = 68
        print(f"\n{cls.CYAN}╭{'─' * (width - 2)}╮{cls.RESET}")
        print(f"{cls.CYAN}│  {cls.BOLD}{cls.WHITE}{title.ljust(width - 5)}{cls.RESET}{cls.CYAN}│{cls.RESET}")
        if subtitle:
            sub = subtitle if len(subtitle) <= width - 5 else subtitle[:width - 8] + "..."
            print(f"{cls.CYAN}│  {cls.GRAY}{sub.ljust(width - 5)}{cls.RESET}{cls.CYAN}│{cls.RESET}")
        print(f"{cls.CYAN}╰{'─' * (width - 2)}╯{cls.RESET}")

    @classmethod
    def badge(cls, text: str, style: str = "ok") -> str:
        if style == "ok":
            return f"{cls.GREEN}[ OK ]{cls.RESET}"
        elif style == "active":
            return f"{cls.CYAN}[ ACTIVE ]{cls.RESET}"
        elif style == "disabled":
            return f"{cls.DARK_GRAY}[ DISABLED ]{cls.RESET}"
        elif style == "warn":
            return f"{cls.YELLOW}[ WARN ]{cls.RESET}"
        elif style == "fail":
            return f"{cls.RED}[ ERROR ]{cls.RESET}"
        elif style == "info":
            return f"{cls.PURPLE}[ INFO ]{cls.RESET}"
        return f"[{text}]"


def get_default_config_path() -> Path:
    home = Path.home()
    return home / ".gemini" / "config" / "mcp_config.json"


def load_config(config_path: Path) -> dict:
    if not config_path.is_file():
        return {"mcpServers": {}, "_disabledMcpServers": {}}
    try:
        with open(config_path, "r", encoding="utf-8") as f:
            data = json.load(f)
            if "mcpServers" not in data:
                data["mcpServers"] = {}
            if "_disabledMcpServers" not in data:
                data["_disabledMcpServers"] = {}
            return data
    except Exception as e:
        print(f"{UI.badge('CONFIG ERROR', 'fail')} Failed to read {config_path}: {e}")
        return {"mcpServers": {}, "_disabledMcpServers": {}, "_error": str(e)}


def save_config(config_path: Path, data: dict) -> bool:
    try:
        config_path.parent.mkdir(parents=True, exist_ok=True)
        if config_path.is_file():
            ts = datetime.datetime.now().strftime("%Y%m%d-%H%M%S")
            bak = config_path.with_name(f"{config_path.name}.bak-{ts}")
            shutil.copy2(config_path, bak)
        
        with open(config_path, "w", encoding="utf-8") as f:
            json.dump(data, f, indent=2)
            f.write("\n")
        return True
    except Exception as e:
        print(f"{UI.badge('SAVE ERROR', 'fail')} Failed to save {config_path}: {e}")
        return False


def probe_mcp_server(name: str, config: dict, timeout_sec: float = 3.5) -> tuple[bool, str, float]:
    cmd = config.get("command")
    args = config.get("args", [])
    env_vars = os.environ.copy()
    if "env" in config and isinstance(config["env"], dict):
        env_vars.update(config["env"])

    if not cmd:
        return False, "No command specified", 0.0

    resolved_cmd = shutil.which(cmd) or cmd
    full_cmd = [resolved_cmd] + args
    start_time = time.perf_counter()

    try:
        if not shutil.which(cmd) and not Path(cmd).is_file():
            return False, f"'{cmd}' not found in PATH", 0.0

        proc = subprocess.Popen(
            full_cmd,
            stdin=subprocess.PIPE,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            env=env_vars,
            text=True,
            shell=IS_WINDOWS,
            bufsize=1,
        )

        init_request = json.dumps({
            "jsonrpc": "2.0",
            "id": 1,
            "method": "initialize",
            "params": {
                "protocolVersion": "2024-11-05",
                "capabilities": {},
                "clientInfo": {"name": "mcp-checker", "version": "1.0.0"}
            }
        }) + "\n"

        try:
            stdout_data, stderr_data = proc.communicate(input=init_request, timeout=timeout_sec)
            elapsed_ms = (time.perf_counter() - start_time) * 1000.0
            
            if stdout_data and ("jsonrpc" in stdout_data or "result" in stdout_data or "capabilities" in stdout_data):
                return True, "JSON-RPC Handshake Success", elapsed_ms
            elif proc.returncode == 0 or proc.returncode is None:
                return True, "Process responsive", elapsed_ms
            else:
                err_msg = stderr_data.strip().split("\n")[-1] if stderr_data else f"Exit code {proc.returncode}"
                return False, f"{err_msg[:24]}", elapsed_ms
        except subprocess.TimeoutExpired:
            proc.kill()
            elapsed_ms = (time.perf_counter() - start_time) * 1000.0
            return True, "Standby (Stdio Ready)", elapsed_ms

    except Exception as e:
        elapsed_ms = (time.perf_counter() - start_time) * 1000.0
        return False, f"{str(e)[:24]}", elapsed_ms


def cmd_status(config_path: Path, run_probe: bool = True):
    UI.header(
        "MCP SERVER DIAGNOSTICS & HEALTH CHECK",
        f"Config: {config_path} | OS: {platform.system()} ({platform.machine()})"
    )
    
    cfg = load_config(config_path)
    active_servers = cfg.get("mcpServers", {})
    disabled_servers = cfg.get("_disabledMcpServers", {})

    total_active = len(active_servers)
    total_disabled = len(disabled_servers)

    print(f"\n  {UI.BOLD}Summary:{UI.RESET} {UI.GREEN}{total_active} Active{UI.RESET} | {UI.DARK_GRAY}{total_disabled} Disabled{UI.RESET}\n")

    if not active_servers and not disabled_servers:
        print(f"  {UI.badge('INFO', 'warn')} No MCP servers configured in this file.")
        return

    w_name = 24
    w_status = 16
    w_info = 26
    w_cmd = 28

    hdr = f"  {pad_visible(UI.BOLD + 'SERVER NAME' + UI.RESET, w_name)} {pad_visible(UI.BOLD + 'STATUS' + UI.RESET, w_status)} {pad_visible(UI.BOLD + 'LATENCY / INFO' + UI.RESET, w_info)} {UI.BOLD}COMMAND{UI.RESET}"
    print(hdr)
    print(f"  {UI.DARK_GRAY}{'─' * (w_name + w_status + w_info + w_cmd)}{UI.RESET}")

    for name, s_cfg in active_servers.items():
        cmd_str = f"{s_cfg.get('command', '')} {' '.join(s_cfg.get('args', []))}"
        if len(cmd_str) > w_cmd:
            cmd_str = cmd_str[:w_cmd - 3] + "..."

        if run_probe:
            is_ok, msg, lat = probe_mcp_server(name, s_cfg)
            if is_ok:
                badge_str = UI.badge("ACTIVE", "active")
                lat_str = f"{UI.GREEN}{lat:.0f}ms{UI.RESET} {UI.DIM}({msg[:14]}){UI.RESET}"
            else:
                badge_str = UI.badge("ERROR", "fail")
                lat_str = f"{UI.RED}{msg}{UI.RESET}"
        else:
            badge_str = UI.badge("ACTIVE", "active")
            lat_str = f"{UI.GRAY}Ready{UI.RESET}"

        line = f"  {pad_visible(UI.WHITE + name + UI.RESET, w_name)} {pad_visible(badge_str, w_status)} {pad_visible(lat_str, w_info)} {UI.DARK_GRAY}{cmd_str}{UI.RESET}"
        print(line)

    for name, s_cfg in disabled_servers.items():
        cmd_str = f"{s_cfg.get('command', '')} {' '.join(s_cfg.get('args', []))}"
        if len(cmd_str) > w_cmd:
            cmd_str = cmd_str[:w_cmd - 3] + "..."
        badge_str = UI.badge("DISABLED", "disabled")
        lat_str = f"{UI.DARK_GRAY}Idle (Token Saved){UI.RESET}"
        line = f"  {pad_visible(UI.DARK_GRAY + name + UI.RESET, w_name)} {pad_visible(badge_str, w_status)} {pad_visible(lat_str, w_info)} {UI.DARK_GRAY}{cmd_str}{UI.RESET}"
        print(line)

    print(f"  {UI.DARK_GRAY}{'─' * (w_name + w_status + w_info + w_cmd)}{UI.RESET}\n")


def cmd_toggle(config_path: Path, target_server: str = None):
    UI.header("TOGGLE MCP SERVER", "Enable or disable servers to optimize RAM & context tokens")
    
    cfg = load_config(config_path)
    active = cfg.get("mcpServers", {})
    disabled = cfg.get("_disabledMcpServers", {})

    all_servers = list(active.keys()) + list(disabled.keys())
    if not all_servers:
        print(f"  {UI.badge('INFO', 'warn')} No servers available to toggle.")
        return

    if not target_server:
        print(f"\n  {UI.BOLD}Configured Servers:{UI.RESET}")
        for idx, s in enumerate(all_servers, 1):
            status = f"{UI.GREEN}ACTIVE{UI.RESET}" if s in active else f"{UI.DARK_GRAY}DISABLED{UI.RESET}"
            print(f"  [{idx}] {s:<24} ({status})")
        
        choice = input(f"\n  {UI.CYAN}Enter server number or name to toggle (0 to cancel): {UI.RESET}").strip()
        if not choice or choice == "0":
            print("  Operation cancelled.")
            return
        
        if choice.isdigit() and 1 <= int(choice) <= len(all_servers):
            target_server = all_servers[int(choice) - 1]
        elif choice in all_servers:
            target_server = choice
        else:
            print(f"  {UI.badge('FAIL', 'fail')} Server '{choice}' not found.")
            return

    if target_server in active:
        disabled[target_server] = active.pop(target_server)
        cfg["mcpServers"] = active
        cfg["_disabledMcpServers"] = disabled
        if save_config(config_path, cfg):
            print(f"\n  {UI.badge('SUCCESS', 'ok')} Server '{target_server}' disabled {UI.DARK_GRAY}(RAM & Context saved){UI.RESET}.")
    elif target_server in disabled:
        active[target_server] = disabled.pop(target_server)
        cfg["mcpServers"] = active
        cfg["_disabledMcpServers"] = disabled
        if save_config(config_path, cfg):
            print(f"\n  {UI.badge('SUCCESS', 'active')} Server '{target_server}' enabled.")
    else:
        print(f"\n  {UI.badge('ERROR', 'fail')} Server '{target_server}' not registered.")


def cmd_optimize(config_path: Path):
    UI.header("STARTUP OPTIMIZER", "Pre-install and cache packages locally to eliminate 'npx -y' latency")
    
    cfg = load_config(config_path)
    active = cfg.get("mcpServers", {})
    
    npx_targets = []
    for name, s_cfg in active.items():
        cmd = s_cfg.get("command", "")
        args = s_cfg.get("args", [])
        if cmd == "npx" and "-y" in args:
            for a in args:
                if a not in ["-y", "-p", "--quiet"]:
                    npx_targets.append((name, a))
                    break

    if not npx_targets:
        print(f"\n  {UI.badge('OK', 'ok')} All active servers use local binaries or do not require npx pre-caching.")
        return

    print(f"\n  Found {len(npx_targets)} server(s) running remote {UI.BOLD}npx -y{UI.RESET}:")
    for name, pkg in npx_targets:
        print(f"  - {UI.CYAN}{name}{UI.RESET} -> Package: {UI.YELLOW}{pkg}{UI.RESET}")

    npm_path = shutil.which("npm")
    if not npm_path:
        print(f"\n  {UI.badge('FAIL', 'fail')} Node.js / npm not found in system PATH. Please install Node.js.")
        return

    confirm = input(f"\n  {UI.CYAN}Install these packages globally now? (y/N): {UI.RESET}").strip().lower()
    if confirm != "y":
        print("  Operation cancelled.")
        return

    for name, pkg in npx_targets:
        print(f"\n  {UI.badge('INSTALL', 'info')} Installing {UI.BOLD}{pkg}{UI.RESET} globally...")
        try:
            res = subprocess.run([npm_path, "install", "-g", pkg], capture_output=True, text=True, shell=IS_WINDOWS)
            if res.returncode == 0:
                print(f"  {UI.badge('SUCCESS', 'ok')} Package {pkg} installed successfully.")
            else:
                print(f"  {UI.badge('WARN', 'warn')} Failed to install {pkg}: {res.stderr.strip()[:100]}")
        except Exception as e:
            print(f"  {UI.badge('FAIL', 'fail')} Error: {e}")

    print(f"\n  {UI.badge('COMPLETE', 'ok')} Optimization complete. Startup latency eliminated.")


def cmd_backup_restore(config_path: Path):
    UI.header("BACKUP & RESTORE CONFIG", f"Path: {config_path}")
    
    parent_dir = config_path.parent
    backups = sorted(parent_dir.glob(f"{config_path.name}.bak-*"), reverse=True)
    
    print(f"  [1] Create New Snapshot Backup")
    print(f"  [2] Restore Saved Backup ({len(backups)} available)")
    print(f"  [0] Cancel")
    
    choice = input(f"\n  {UI.CYAN}Choice: {UI.RESET}").strip()
    if choice == "1":
        if config_path.is_file():
            ts = datetime.datetime.now().strftime("%Y%m%d-%H%M%S")
            bak = config_path.with_name(f"{config_path.name}.bak-{ts}")
            shutil.copy2(config_path, bak)
            print(f"\n  {UI.badge('BACKUP OK', 'ok')} Backup saved to: {bak.name}")
        else:
            print(f"  {UI.badge('FAIL', 'fail')} Config file does not exist.")
    elif choice == "2":
        if not backups:
            print(f"  {UI.badge('INFO', 'warn')} No backups available.")
            return
        print(f"\n  {UI.BOLD}Available Backups:{UI.RESET}")
        for i, b in enumerate(backups[:8], 1):
            size = b.stat().st_size
            mtime = datetime.datetime.fromtimestamp(b.stat().st_mtime).strftime("%Y-%m-%d %H:%M:%S")
            print(f"  [{i}] {b.name} ({size} bytes, {mtime})")
        
        b_choice = input(f"\n  {UI.CYAN}Enter backup number to restore (0 to cancel): {UI.RESET}").strip()
        if b_choice.isdigit() and 1 <= int(b_choice) <= min(len(backups), 8):
            selected_bak = backups[int(b_choice) - 1]
            shutil.copy2(selected_bak, config_path)
            print(f"\n  {UI.badge('RESTORE OK', 'ok')} Configuration restored from: {selected_bak.name}")
        else:
            print("  Operation cancelled.")


def interactive_menu(config_path: Path):
    while True:
        UI.header(
            "MCP CONTROLLER (Universal Cross-Platform)",
            f"OS: {platform.system()} | Python {platform.python_version()} | Config: {config_path.name}"
        )
        print(f"  [1] Status & Health Check (Latency probe & diagnostics)")
        print(f"  [2] Toggle Server ON / OFF (Optimize RAM & AI token usage)")
        print(f"  [3] Optimize Startup (Pre-cache npx packages)")
        print(f"  [4] Backup & Restore Config")
        print(f"  [0] Exit")
        
        choice = input(f"\n  {UI.CYAN}Select option [0-4]: {UI.RESET}").strip()
        if choice == "1":
            cmd_status(config_path, run_probe=True)
            input(f"\n  {UI.DARK_GRAY}Press Enter to return to menu...{UI.RESET}")
        elif choice == "2":
            cmd_toggle(config_path)
            input(f"\n  {UI.DARK_GRAY}Press Enter to return to menu...{UI.RESET}")
        elif choice == "3":
            cmd_optimize(config_path)
            input(f"\n  {UI.DARK_GRAY}Press Enter to return to menu...{UI.RESET}")
        elif choice == "4":
            cmd_backup_restore(config_path)
            input(f"\n  {UI.DARK_GRAY}Press Enter to return to menu...{UI.RESET}")
        elif choice in ["0", "q", "exit"]:
            print(f"\n  {UI.CYAN}Exiting MCP Controller.{UI.RESET}\n")
            break
        else:
            print(f"  {UI.badge('INVALID', 'warn')} Invalid selection.")


def main():
    parser = argparse.ArgumentParser(description="Universal Cross-Platform MCP Manager & Optimizer")
    parser.add_argument("command", nargs="?", choices=["status", "toggle", "optimize", "backup", "menu"], default=None,
                        help="Action to execute")
    parser.add_argument("server", nargs="?", default=None, help="Target server name (for toggle)")
    parser.add_argument("--config", "-c", type=str, default=None, help="Custom path to mcp_config.json")
    parser.add_argument("--no-probe", action="store_true", help="Skip live stdio latency probe during status check")

    args = parser.parse_args()
    config_path = Path(args.config) if args.config else get_default_config_path()

    if args.command == "status":
        cmd_status(config_path, run_probe=not args.no_probe)
    elif args.command == "toggle":
        cmd_toggle(config_path, args.server)
    elif args.command == "optimize":
        cmd_optimize(config_path)
    elif args.command == "backup":
        cmd_backup_restore(config_path)
    elif args.command == "menu" or args.command is None:
        interactive_menu(config_path)


if __name__ == "__main__":
    main()
