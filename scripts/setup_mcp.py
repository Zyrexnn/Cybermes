#!/usr/bin/env python3
"""
scripts/setup_mcp.py — Cybermes MCP Server Multi-Client Auto-Installer & Config Injector

Instantly detects and wires Cybermes MCP Server into 11+ AI Clients:
Claude Desktop, Cursor, OpenCode, Windsurf, Cline, Roo Code, Claude Code, Continue, Zed, Kilo, Hermes, Codex.

Usage:
  python scripts/setup_mcp.py [--all] [--local] [--dry-run] [--force] [--status] [--uninstall]
"""

import argparse
import datetime
import json
import os
import platform
import shutil
import sys
from pathlib import Path

# Safe Unicode output on Windows PowerShell/CMD
if hasattr(sys.stdout, "reconfigure"):
    sys.stdout.reconfigure(encoding="utf-8", errors="replace")
if hasattr(sys.stderr, "reconfigure"):
    sys.stderr.reconfigure(encoding="utf-8", errors="replace")

VERSION = "3.1.1"
CYBERMES_ROOT = Path(__file__).resolve().parent.parent


def get_local_binary_path():
    is_win = platform.system() == "Windows"
    ext = ".exe" if is_win else ""
    candidates = [
        CYBERMES_ROOT / "tools" / "bin" / f"cybermes-mcp{ext}",
        CYBERMES_ROOT / f"cybermes-mcp{ext}",
    ]
    for c in candidates:
        if c.is_file():
            return str(c)
    return None


def create_backup(file_path: Path):
    if file_path.is_file():
        ts = datetime.datetime.now().strftime("%Y%m%d-%H%M%S")
        bak = file_path.with_name(f"{file_path.name}.bak-{ts}")
        shutil.copy2(file_path, bak)
        return str(bak)
    return None


def safe_read_json(file_path: Path):
    if not file_path.is_file():
        return None
    try:
        with open(file_path, "r", encoding="utf-8") as f:
            content = f.read().strip()
            if not content:
                return {}
            try:
                return json.loads(content)
            except Exception:
                # Strip comments while preserving strings with quotes
                cleaned = re.sub(r'("(?:\\.|[^"\\])*")|//.*$|/\*[\s\S]*?\*/', lambda m: m.group(1) or '', content, flags=re.MULTILINE)
                cleaned = re.sub(r',\s*([}\]])', r'\1', cleaned)
                return json.loads(cleaned)
    except Exception as e:
        return {"_parseError": str(e)}


def safe_write_json(file_path: Path, data: dict, dry_run: bool):
    if dry_run:
        return True
    file_path.parent.mkdir(parents=True, exist_ok=True)
    with open(file_path, "w", encoding="utf-8") as f:
        json.dump(data, f, indent=2)
        f.write("\n")
    return True


def get_client_definitions(use_local=False, local_bin=None):
    home = Path.home()
    is_win = platform.system() == "Windows"
    is_mac = platform.system() == "Darwin"

    appdata = Path(os.environ.get("APPDATA", home / "AppData" / "Roaming")) if is_win else home

    if use_local and local_bin:
        default_def = {
            "command": str(local_bin),
            "args": ["-workspace", str(CYBERMES_ROOT)],
        }
    else:
        default_def = {
            "command": "npx",
            "args": ["-y", "cybermes-mcp"],
        }

    return [
        {
            "id": "claude-desktop",
            "name": "Claude Desktop",
            "paths": [
                appdata / "Claude" / "claude_desktop_config.json"
                if is_win
                else (home / "Library" / "Application Support" / "Claude" / "claude_desktop_config.json" if is_mac else home / ".config" / "Claude" / "claude_desktop_config.json")
            ],
            "type": "json-mcpServers",
            "definition": default_def,
        },
        {
            "id": "cursor",
            "name": "Cursor IDE",
            "paths": [
                home / ".cursor" / "mcp.json",
                Path.cwd() / ".cursor" / "mcp.json",
            ],
            "type": "json-mcpServers",
            "definition": default_def,
        },
        {
            "id": "opencode",
            "name": "OpenCode Interpreter",
            "paths": [
                p for p in [
                    home / ".config" / "opencode" / "opencode.json",
                    home / ".config" / "opencode" / "opencode.jsonc",
                    home / ".config" / "opencode" / "config.json",
                    appdata / "opencode" / "opencode.json" if is_win else None,
                    Path.cwd() / "opencode.json",
                    Path.cwd() / "opencode.jsonc",
                ] if p is not None
            ],
            "type": "json-mcp_servers",
            "definition": default_def,
        },
        {
            "id": "windsurf",
            "name": "Windsurf IDE (Codeium)",
            "paths": [
                home / ".codeium" / "windsurf" / "mcp_config.json",
            ],
            "type": "json-mcpServers",
            "definition": default_def,
        },
        {
            "id": "cline",
            "name": "Cline (VS Code Extension)",
            "paths": [
                appdata / "Code" / "User" / "globalStorage" / "saoudrizwan.claude-dev" / "settings" / "cline_mcp_settings.json"
                if is_win
                else (home / "Library" / "Application Support" / "Code" / "User" / "globalStorage" / "saoudrizwan.claude-dev" / "settings" / "cline_mcp_settings.json" if is_mac else home / ".config" / "Code" / "User" / "globalStorage" / "saoudrizwan.claude-dev" / "settings" / "cline_mcp_settings.json")
            ],
            "type": "json-cline",
            "definition": {
                **default_def,
                "disabled": False,
                "autoApprove": [
                    "cybermes_search_knowledge",
                    "cybermes_list_skills",
                    "cybermes_get_skill",
                    "cybermes_scan_secrets",
                    "cybermes_validate_scope",
                ],
            },
        },
        {
            "id": "roo-code",
            "name": "Roo Code (VS Code Extension)",
            "paths": [
                appdata / "Code" / "User" / "globalStorage" / "rooveterinaryinc.roo-cline" / "settings" / "cline_mcp_settings.json"
                if is_win
                else (home / "Library" / "Application Support" / "Code" / "User" / "globalStorage" / "rooveterinaryinc.roo-cline" / "settings" / "cline_mcp_settings.json" if is_mac else home / ".config" / "Code" / "User" / "globalStorage" / "rooveterinaryinc.roo-cline" / "settings" / "cline_mcp_settings.json")
            ],
            "type": "json-cline",
            "definition": {
                **default_def,
                "disabled": False,
                "autoApprove": [
                    "cybermes_search_knowledge",
                    "cybermes_list_skills",
                    "cybermes_get_skill",
                    "cybermes_scan_secrets",
                    "cybermes_validate_scope",
                ],
            },
        },
        {
            "id": "claude-code",
            "name": "Claude Code CLI",
            "paths": [
                home / ".claude.json",
            ],
            "type": "json-mcpServers",
            "definition": default_def,
        },
        {
            "id": "continue",
            "name": "Continue.dev",
            "paths": [
                home / ".continue" / "config.json",
            ],
            "type": "json-continue",
            "definition": {
                "name": "cybermes",
                "transport": {
                    "type": "stdio",
                    "command": default_def["command"],
                    "args": default_def["args"],
                },
            },
        },
        {
            "id": "zed",
            "name": "Zed Editor",
            "paths": [
                home / ".config" / "zed" / "settings.json",
                appdata / "Zed" / "settings.json" if is_win else None,
            ],
            "type": "json-zed",
            "definition": default_def,
        },
        {
            "id": "kilo",
            "name": "Kilo Code",
            "paths": [
                home / ".kilo" / "mcp.json",
            ],
            "type": "json-mcpServers",
            "definition": default_def,
        },
        {
            "id": "hermes",
            "name": "Hermes Agent",
            "paths": [
                home / ".hermes" / "config.yaml",
                CYBERMES_ROOT / ".hermes" / "config.yaml",
            ],
            "type": "yaml-hermes",
            "definition": default_def,
        },
        {
            "id": "codex",
            "name": "Codex CLI",
            "paths": [
                home / ".codex" / "config.toml",
            ],
            "type": "toml-codex",
            "definition": default_def,
        },
    ]


def inject_config(client: dict, file_path: Path, dry_run: bool):
    ctype = client["type"]
    cdef = client["definition"]

    if ctype in ("json-mcpServers", "json-cline"):
        data = safe_read_json(file_path)
        if data and "_parseError" in data:
            return {"status": "error", "details": f"Invalid JSON: {data['_parseError']}"}
        if data is None:
            data = {}

        data.setdefault("mcpServers", {})
        if data["mcpServers"].get("cybermes") == cdef:
            return {"status": "unchanged", "details": "Already up-to-date"}

        data["mcpServers"]["cybermes"] = cdef
        if not dry_run:
            bak = create_backup(file_path)
            safe_write_json(file_path, data, False)
            return {"status": "injected", "details": f"Updated (backup: {Path(bak).name})" if bak else "Created config"}
        return {"status": "dry-run", "details": "Would inject mcpServers.cybermes"}

    if ctype == "json-mcp_servers":
        data = safe_read_json(file_path)
        if data and "_parseError" in data:
            return {"status": "error", "details": f"Invalid JSON: {data['_parseError']}"}
        if data is None:
            data = {}

        opencode_entry = {
            "type": "local",
            "command": [cdef["command"]] + cdef["args"],
            "enabled": True,
        }

        if "mcp" in data or ("mcp_servers" not in data and "mcpServers" not in data):
            data.setdefault("mcp", {})
            if data["mcp"].get("cybermes") == opencode_entry:
                return {"status": "unchanged", "details": "Already up-to-date in mcp"}
            data["mcp"]["cybermes"] = opencode_entry
        else:
            data.setdefault("mcp_servers", {})
            if data["mcp_servers"].get("cybermes") == cdef:
                return {"status": "unchanged", "details": "Already up-to-date in mcp_servers"}
            data["mcp_servers"]["cybermes"] = cdef

        if not dry_run:
            bak = create_backup(file_path)
            safe_write_json(file_path, data, False)
            return {"status": "injected", "details": f"Updated (backup: {Path(bak).name})" if bak else "Created config"}
        return {"status": "dry-run", "details": "Would inject mcp.cybermes"}

    if ctype == "json-continue":
        data = safe_read_json(file_path)
        if data and "_parseError" in data:
            return {"status": "error", "details": f"Invalid JSON: {data['_parseError']}"}
        if data is None:
            data = {}

        exp = data.setdefault("experimental", {})
        srvs = exp.setdefault("modelContextProtocolServers", [])

        idx = next((i for i, s in enumerate(srvs) if s.get("name") == "cybermes"), -1)
        if idx >= 0 and srvs[idx] == cdef:
            return {"status": "unchanged", "details": "Already up-to-date"}

        if idx >= 0:
            srvs[idx] = cdef
        else:
            srvs.append(cdef)

        if not dry_run:
            bak = create_backup(file_path)
            safe_write_json(file_path, data, False)
            return {"status": "injected", "details": f"Updated (backup: {Path(bak).name})" if bak else "Created config"}
        return {"status": "dry-run", "details": "Would inject experimental.modelContextProtocolServers"}

    if ctype == "json-zed":
        data = safe_read_json(file_path)
        if data and "_parseError" in data:
            return {"status": "error", "details": f"Invalid JSON: {data['_parseError']}"}
        if data is None:
            data = {}

        ctx = data.setdefault("context_servers", {})
        if ctx.get("cybermes") == cdef:
            return {"status": "unchanged", "details": "Already up-to-date"}

        ctx["cybermes"] = cdef
        if not dry_run:
            bak = create_backup(file_path)
            safe_write_json(file_path, data, False)
            return {"status": "injected", "details": f"Updated (backup: {Path(bak).name})" if bak else "Created config"}
        return {"status": "dry-run", "details": "Would inject context_servers.cybermes"}

    if ctype == "yaml-hermes":
        content = file_path.read_text(encoding="utf-8") if file_path.is_file() else ""
        if "cybermes:" in content and ("cybermes-mcp" in content or "@zyrexnn/cybermes-mcp" in content):
            return {"status": "unchanged", "details": "Already up-to-date in YAML"}

        cmd_json = json.dumps(cdef["command"])
        args_json = json.dumps(cdef["args"])
        if "mcp_servers:" in content:
            block = f"\n  cybermes:\n    command: {cmd_json}\n    args: {args_json}\n"
        else:
            block = f"\nmcp_servers:\n  cybermes:\n    command: {cmd_json}\n    args: {args_json}\n"

        if not dry_run:
            bak = create_backup(file_path)
            file_path.parent.mkdir(parents=True, exist_ok=True)
            with open(file_path, "a", encoding="utf-8") as f:
                f.write(block)
            return {"status": "injected", "details": f"Appended (backup: {Path(bak).name})" if bak else "Created config"}
        return {"status": "dry-run", "details": "Would append YAML block"}

    if ctype == "toml-codex":
        content = file_path.read_text(encoding="utf-8") if file_path.is_file() else ""
        if "[mcp_servers.cybermes]" in content:
            return {"status": "unchanged", "details": "Already up-to-date in TOML"}

        cmd_json = json.dumps(cdef["command"])
        args_json = ", ".join(json.dumps(a) for a in cdef["args"])
        block = f"\n[mcp_servers.cybermes]\ncommand = {cmd_json}\nargs = [{args_json}]\n"

        if not dry_run:
            bak = create_backup(file_path)
            file_path.parent.mkdir(parents=True, exist_ok=True)
            with open(file_path, "a", encoding="utf-8") as f:
                f.write(block)
            return {"status": "injected", "details": f"Appended (backup: {Path(bak).name})" if bak else "Created config"}
        return {"status": "dry-run", "details": "Would append TOML block"}

    return {"status": "skipped", "details": "Unsupported format"}


def remove_config(client: dict, file_path: Path, dry_run: bool):
    if not file_path.is_file():
        return {"status": "not_found", "details": "File does not exist"}

    ctype = client["type"]
    if ctype in ("json-mcpServers", "json-cline"):
        data = safe_read_json(file_path)
        if not data or "_parseError" in data or not data.get("mcpServers", {}).get("cybermes"):
            return {"status": "unchanged", "details": "Cybermes not present"}
        del data["mcpServers"]["cybermes"]
        if not dry_run:
            bak = create_backup(file_path)
            safe_write_json(file_path, data, False)
            return {"status": "removed", "details": f"Cleaned (backup: {Path(bak).name})" if bak else "Removed"}
        return {"status": "dry-run", "details": "Would remove mcpServers.cybermes"}

    if ctype == "json-mcp_servers":
        data = safe_read_json(file_path)
        if not data or "_parseError" in data:
            return {"status": "unchanged", "details": "Cybermes not present"}
        has_removed = False
        if "mcp" in data and "cybermes" in data["mcp"]:
            del data["mcp"]["cybermes"]
            has_removed = True
        if "mcp_servers" in data and "cybermes" in data["mcp_servers"]:
            del data["mcp_servers"]["cybermes"]
            has_removed = True
        if not has_removed:
            return {"status": "unchanged", "details": "Cybermes not present"}
        if not dry_run:
            bak = create_backup(file_path)
            safe_write_json(file_path, data, False)
            return {"status": "removed", "details": f"Cleaned (backup: {Path(bak).name})" if bak else "Removed"}
        return {"status": "dry-run", "details": "Would remove mcp.cybermes"}

    return {"status": "skipped", "details": "Manual cleanup recommended for YAML/TOML"}


def check_status(client: dict, file_path: Path):
    if not file_path or not file_path.is_file():
        return {"installed": False, "configured": False, "details": "Not detected"}

    ctype = client["type"]
    if ctype in ("json-mcpServers", "json-cline"):
        data = safe_read_json(file_path)
        has_cb = bool(data and data.get("mcpServers", {}).get("cybermes"))
        return {"installed": True, "configured": has_cb, "details": "Configured" if has_cb else "Detected (Missing Cybermes)"}

    if ctype == "json-mcp_servers":
        data = safe_read_json(file_path)
        has_cb = bool(data and ((data.get("mcp") and "cybermes" in data["mcp"]) or (data.get("mcp_servers") and "cybermes" in data["mcp_servers"])))
        return {"installed": True, "configured": has_cb, "details": "Configured" if has_cb else "Detected (Missing Cybermes)"}

    if ctype in ("yaml-hermes", "toml-codex"):
        content = file_path.read_text(encoding="utf-8")
        has_cb = "cybermes" in content
        return {"installed": True, "configured": has_cb, "details": "Configured" if has_cb else "Detected (Missing Cybermes)"}

    return {"installed": True, "configured": False, "details": "Detected"}


def main():
    parser = argparse.ArgumentParser(description=f"Cybermes MCP Universal Multi-Client Auto-Installer v{VERSION}")
    parser.add_argument("--dry-run", action="store_true", help="Simulate execution without modifying files")
    parser.add_argument("--local", action="store_true", help="Use local compiled binary (tools/bin/cybermes-mcp)")
    parser.add_argument("--force", action="store_true", help="Generate config files even if client not detected")
    parser.add_argument("--status", action="store_true", help="Display client discovery and configuration matrix")
    parser.add_argument("--uninstall", action="store_true", help="Cleanly remove Cybermes MCP configuration")
    parser.add_argument("--clients", type=str, help="Comma-separated client IDs to target")
    parser.add_argument("--opencode", action="store_true", help="Target only OpenCode Interpreter")
    parser.add_argument("--gemini", action="store_true", help="Target only Antigravity / Gemini")
    parser.add_argument("--cursor", action="store_true", help="Target only Cursor IDE")
    parser.add_argument("--claude", action="store_true", help="Target only Claude Desktop")
    parser.add_argument("--windsurf", action="store_true", help="Target only Windsurf IDE")
    parser.add_argument("--cline", action="store_true", help="Target only Cline")
    parser.add_argument("--roo", action="store_true", help="Target only Roo Code")
    parser.add_argument("--continue", dest="continue_dev", action="store_true", help="Target only Continue.dev")
    parser.add_argument("--zed", action="store_true", help="Target only Zed Editor")
    parser.add_argument("--kilo", action="store_true", help="Target only Kilo Code")
    parser.add_argument("--hermes", action="store_true", help="Target only Hermes Agent")
    parser.add_argument("--codex", action="store_true", help="Target only Codex CLI")

    args = parser.parse_args()

    local_bin = get_local_binary_path() if args.local else None
    clients = get_client_definitions(use_local=args.local, local_bin=local_bin)

    direct_targets = []
    if args.opencode:
        direct_targets.append("opencode")
    if args.gemini:
        direct_targets.extend(["gemini", "antigravity"])
    if args.cursor:
        direct_targets.append("cursor")
    if args.claude:
        direct_targets.extend(["claude", "claude-desktop", "claude-code"])
    if args.windsurf:
        direct_targets.append("windsurf")
    if args.cline:
        direct_targets.append("cline")
    if args.roo:
        direct_targets.append("roo-cline")
    if args.continue_dev:
        direct_targets.append("continue")
    if args.zed:
        direct_targets.append("zed")
    if args.kilo:
        direct_targets.append("kilo")
    if args.hermes:
        direct_targets.append("hermes")
    if args.codex:
        direct_targets.append("codex")

    if direct_targets:
        filter_list = direct_targets
    elif args.clients:
        filter_list = [c.strip().lower() for c in args.clients.split(",")]
    else:
        filter_list = None

    if args.status:
        print(f"\n📊 Cybermes MCP Server — Client Discovery Matrix v{VERSION}")
        print("=" * 70)
        for client in clients:
            valid_paths = [p for p in client["paths"] if p]
            target_path = next((p for p in valid_paths if p.is_file()), valid_paths[0] if valid_paths else None)
            st = check_status(client, target_path)
            mark = "✓" if st["configured"] else ("!" if st["installed"] else "-")
            state_str = "[CONFIGURED]" if st["configured"] else ("[NOT WIRED]" if st["installed"] else "[NOT DETECTED]")
            print(f"  {mark} {client['name']:<26} {state_str:<16} {target_path}")
        print("=" * 70)
        print("💡 Run `python scripts/setup_mcp.py` to auto-inject into all un-wired clients.\n")
        return 0

    if args.uninstall:
        print(f"\n🗑️  Cybermes MCP Server — Uninstaller v{VERSION}")
        print("=" * 70)
        if args.dry_run:
            print("🔍 DRY RUN MODE ACTIVATED — No configuration files will be modified.\n")

        removed = 0
        for client in clients:
            if filter_list and client["id"] not in filter_list and client["name"].lower() not in filter_list:
                continue
            valid_paths = [p for p in client["paths"] if p]
            target_path = next((p for p in valid_paths if p.is_file()), None)
            if not target_path:
                continue
            res = remove_config(client, target_path, args.dry_run)
            if res["status"] in ("removed", "dry-run"):
                removed += 1
                print(f"  ✓ {client['name']:<26} -> [{res['status'].upper()}] {target_path} ({res['details']})")
        print("=" * 70)
        print(f"✨ Cleanup completed! Removed from {removed} client(s).\n")
        return 0

    # Default: Install / Auto-Inject
    print(f"\n🛡️  Cybermes MCP Server — Universal Auto-Installer v{VERSION}")
    print("=" * 70)
    if args.dry_run:
        print("🔍 DRY RUN MODE ACTIVATED — No configuration files will be modified.\n")

    injected = 0
    detected = 0

    for client in clients:
        if filter_list and client["id"] not in filter_list and client["name"].lower() not in filter_list:
            continue

        valid_paths = [p for p in client["paths"] if p]
        target_path = next((p for p in valid_paths if p.is_file()), None)
        if not target_path:
            if args.force and valid_paths:
                target_path = valid_paths[0]
            else:
                continue

        detected += 1
        res = inject_config(client, target_path, args.dry_run)

        if res["status"] in ("injected", "dry-run"):
            injected += 1
            print(f"  ✓ {client['name']:<26} -> [{res['status'].upper()}] {target_path} ({res['details']})")
        elif res["status"] == "unchanged":
            print(f"  = {client['name']:<26} -> [UNCHANGED] {target_path} ({res['details']})")
        elif res["status"] == "error":
            print(f"  ✗ {client['name']:<26} -> [ERROR] {target_path} ({res['details']})")

    print("=" * 70)
    print(f"🎉 Auto-installer finished! Evaluated: {detected}, Updated: {injected}")
    print("💡 Note: Restart your AI client (Cursor, Claude, Windsurf, etc.) to reload MCP tools.\n")
    return 0


if __name__ == "__main__":
    sys.exit(main())
