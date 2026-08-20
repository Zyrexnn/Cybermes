#!/usr/bin/env bash
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
export CYBERMES_DIR="$SCRIPT_DIR"
export HERMES_HOME="$CYBERMES_DIR/.hermes"
export PATH="$CYBERMES_DIR/tools/bin:$CYBERMES_DIR/bin:$CYBERMES_DIR/venv/bin:$PATH"

if [ -f "$CYBERMES_DIR/venv/bin/activate" ]; then
    source "$CYBERMES_DIR/venv/bin/activate"
fi

echo "✓ Cybermes Environment Activated: $CYBERMES_DIR"
if command -v hermes >/dev/null 2>&1; then
    echo "  Hermes: $(hermes --version 2>&1 | head -n 1)"
fi
echo "  Tools: nmap, subfinder, amass, httpx, katana, gau, ffuf, feroxbuster, nuclei, dalfox, sqlmap, rg"
