#!/usr/bin/env python3
"""
Cybermes Smart Output Filter & Token Saver (smart_pipe.py)
Captures full raw security tool outputs to disk and streams only high-signal results to stdout.
Prevents context window bloat and token exhaustion in AI agents.
"""

import os
import sys
import re
import argparse
from pathlib import Path

BASE_DIR = Path(__file__).resolve().parent.parent
RECON_DIR = BASE_DIR / "recon"

# Static assets with low security relevance
STATIC_EXTENSIONS = (
    ".png", ".jpg", ".jpeg", ".gif", ".svg", ".ico", ".woff", ".woff2", 
    ".ttf", ".eot", ".css", ".mp4", ".mp3", ".webm"
)

ANSI_ESCAPE = re.compile(r'\x1B(?:[@-Z\\-_]|\[[0-?]*[ -/]*[@-~])')

def clean_line(line: str) -> str:
    """Remove ANSI colors and whitespace."""
    return ANSI_ESCAPE.sub('', line).strip()

def is_high_signal(line: str) -> bool:
    """Determine if a line is likely high signal for security assessment."""
    lower = line.lower()
    
    # Filter out pure static assets
    if any(lower.endswith(ext) or f"{ext}?" in lower for ext in STATIC_EXTENSIONS):
        return False
        
    # High signal indicators
    high_keywords = (
        "200 ok", "[200]", "301", "302", "401", "403", "500", "api", "v1", "v2", "v3",
        "auth", "login", "admin", "token", "secret", "user", "passwd", "password",
        "graphql", "swagger", "openapi", "debug", "config", "env", "backup", ".git",
        "cve-", "vulnerability", "[high]", "[critical]", "[medium]", "idor", "sqli", "xss"
    )
    
    return any(k in lower for k in high_keywords) or "?" in line or "=" in line

def process_stream(input_lines, target_slug: str, tool_name: str, limit: int = 35) -> None:
    """Filter stream, save raw to disk, print high-signal lines."""
    target_recon_dir = RECON_DIR / target_slug
    target_recon_dir.mkdir(parents=True, exist_ok=True)
    
    raw_log_path = target_recon_dir / f"{tool_name}_raw.txt"
    
    all_raw_lines = []
    high_signal_lines = []
    seen = set()

    for line in input_lines:
        raw_clean = clean_line(line)
        if not raw_clean:
            continue
        all_raw_lines.append(raw_clean)
        
        if raw_clean not in seen:
            seen.add(raw_clean)
            if is_high_signal(raw_clean):
                high_signal_lines.append(raw_clean)

    # If high-signal filter was too restrictive, fallback to unique lines
    display_lines = high_signal_lines if high_signal_lines else list(seen)
    truncated_lines = display_lines[:limit]

    # Save full raw output to disk
    with open(raw_log_path, "w", encoding="utf-8", errors="replace") as f:
        f.write("\n".join(all_raw_lines) + "\n")
    try:
        os.chmod(raw_log_path, 0o666)
    except Exception:
        pass

    # Print summary & top high-signal lines
    total_count = len(all_raw_lines)
    shown_count = len(truncated_lines)
    
    print(f"📊 [Smart Filter] {shown_count} high-signal findings shown (from {total_count} total lines).")
    print(f"💾 Full raw output saved: recon/{target_slug}/{tool_name}_raw.txt\n")
    
    for l in truncated_lines:
        print(l)
        
    if shown_count < len(display_lines):
        print(f"\n... (+{len(display_lines) - shown_count} more relevant entries saved in raw log)")

def main():
    parser = argparse.ArgumentParser(description="Cybermes Smart Tool Output Filter")
    parser.add_argument("--target", "-t", default="default_target", help="Target slug name")
    parser.add_argument("--tool", "-n", default="tool", help="Tool name (katana, ffuf, httpx, etc.)")
    parser.add_argument("--limit", "-l", type=int, default=35, help="Max lines to display in context")
    args = parser.parse_args()

    # Read from standard input pipe
    if not sys.stdin.isatty():
        process_stream(sys.stdin, args.target, args.tool, args.limit)
    else:
        print("Usage: <tool_command> | python3 tools/smart_pipe.py --target <SLUG> --tool <TOOL>")

if __name__ == "__main__":
    main()
