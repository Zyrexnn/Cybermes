#!/usr/bin/env python3
"""
Cybermes Smart Output Filter & Token Saver (smart_pipe.py)
Captures full raw security tool outputs to disk and streams only prioritized,
high-signal results to stdout. Prevents context window bloat and token exhaustion in AI agents.
"""

import os
import sys
import re
import math
import argparse
from pathlib import Path
from collections import Counter

BASE_DIR = Path(__file__).resolve().parent.parent
RECON_DIR = BASE_DIR / "recon"

# Static assets with very low security relevance
STATIC_EXTENSIONS = (
    ".png", ".jpg", ".jpeg", ".gif", ".svg", ".ico", ".webp",
    ".woff", ".woff2", ".ttf", ".eot", ".otf",
    ".css", ".mp4", ".mp3", ".webm", ".avi", ".mov"
)

ANSI_ESCAPE = re.compile(r'\x1B(?:[@-Z\\-_]|\[[0-?]*[ -/]*[@-~])')
UUID_REGEX = re.compile(r'[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}', re.I)
STATUS_SIZE_REGEX = re.compile(r'\[(200|301|302|401|403|500)\]\s*\[(\d+)\s*(?:bytes|b)?\]', re.I)

def clean_line(line: str) -> str:
    """Remove ANSI color codes and extraneous whitespace."""
    return ANSI_ESCAPE.sub('', line).strip()

def calculate_entropy(text: str) -> float:
    """Calculate Shannon entropy to detect high-entropy keys/hashes."""
    if not text or len(text) < 16:
        return 0.0
    counter = Counter(text)
    length = len(text)
    return -sum((count / length) * math.log2(count / length) for count in counter.values())

def score_signal(line: str) -> int:
    """
    Score relevance of an output line (higher = more critical to security audit).
    Returns an integer score from 0 (noise) to 100+ (critical signal).
    """
    lower = line.lower()
    
    # Immediately drop static assets
    if any(lower.endswith(ext) or f"{ext}?" in lower or f"{ext}#" in lower for ext in STATIC_EXTENSIONS):
        return 0
        
    score = 10
    
    # 1. Critical vulnerability / security findings
    critical_markers = ("[critical]", "[high]", "cve-", "rce", "sql injection", "sqli", "idor", "ssrf", "xxe", "auth bypass")
    if any(m in lower for m in critical_markers):
        score += 80
        
    # 2. Sensitive files, configuration & secrets
    secret_markers = (".env", ".git", "swagger", "openapi", "graphql", "id_rsa", "password", "secret_key", "bearer ", "token=", "jwt")
    if any(m in lower for m in secret_markers):
        score += 60
        
    # 3. HTTP status codes & APIs
    if "200 ok" in lower or "[200]" in lower:
        score += 25
        if "/api/" in lower or "/v1/" in lower or "/v2/" in lower:
            score += 25
    elif any(code in lower for code in ("[401]", "[403]", "401 unauthorized", "403 forbidden")):
        score += 20
        if "/admin" in lower or "/api/" in lower or "/internal" in lower:
            score += 25
    elif any(code in lower for code in ("[500]", "500 internal server error")):
        score += 15
        
    # 4. Parameters and Dynamic Endpoints
    if "?" in line and "=" in line:
        score += 20
    if UUID_REGEX.search(line):
        score += 20
        
    # 5. Shannon entropy check for possible leaked tokens/secrets
    if any(k in lower for k in ("key", "secret", "tok", "pass")):
        if calculate_entropy(line) > 3.8:
            score += 30
            
    return score

def process_stream(input_lines, target_slug: str, tool_name: str, limit: int = 40) -> None:
    """Filter input stream, save complete raw log to disk, and print prioritized high-signal items."""
    target_recon_dir = RECON_DIR / target_slug
    target_recon_dir.mkdir(parents=True, exist_ok=True)
    
    raw_log_path = target_recon_dir / f"{tool_name}_raw.txt"
    
    all_raw_lines = []
    scored_lines = []
    seen = set()

    for line in input_lines:
        raw_clean = clean_line(line)
        if not raw_clean:
            continue
        all_raw_lines.append(raw_clean)
        
        if raw_clean not in seen:
            seen.add(raw_clean)
            score = score_signal(raw_clean)
            if score > 0:
                scored_lines.append((score, raw_clean))

    # Sort lines descending by score (highest signal first)
    scored_lines.sort(key=lambda x: x[0], reverse=True)
    
    # Select display lines up to limit
    display_lines = [item[1] for item in scored_lines[:limit]]
    if not display_lines and seen:
        display_lines = list(seen)[:limit]

    # Save complete raw output to disk
    with open(raw_log_path, "w", encoding="utf-8", errors="replace") as f:
        f.write("\n".join(all_raw_lines) + "\n")
    try:
        os.chmod(raw_log_path, 0o666)
    except Exception:
        pass

    total_count = len(all_raw_lines)
    shown_count = len(display_lines)
    
    print(f"📊 [Smart Filter] {shown_count} high-signal findings prioritized (from {total_count} total raw lines).")
    print(f"💾 Full raw output preserved: recon/{target_slug}/{tool_name}_raw.txt\n")
    
    for l in display_lines:
        print(l)
        
    if len(scored_lines) > shown_count:
        print(f"\n... (+{len(scored_lines) - shown_count} more filtered entries archived in raw log)")

def main():
    parser = argparse.ArgumentParser(description="Cybermes Smart Output Filter & Token Optimizer")
    parser.add_argument("--target", "-t", default="default_target", help="Target slug identifier")
    parser.add_argument("--tool", "-n", default="tool", help="Security tool name (katana, ffuf, httpx, etc.)")
    parser.add_argument("--limit", "-l", type=int, default=40, help="Max prioritized lines to display in context")
    args = parser.parse_args()

    if not sys.stdin.isatty():
        process_stream(sys.stdin, args.target, args.tool, args.limit)
    else:
        print("Usage: <tool_command> | python3 tools/smart_pipe.py --target <SLUG> --tool <TOOL>")

if __name__ == "__main__":
    main()
