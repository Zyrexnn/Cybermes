#!/usr/bin/env bash
# Universal MCP Manager Launcher for Linux & macOS
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

if [ -f "$SCRIPT_DIR/venv/bin/python" ]; then
    PYTHON_EXE="$SCRIPT_DIR/venv/bin/python"
elif command -v python3 >/dev/null 2>&1; then
    PYTHON_EXE="python3"
elif command -v python >/dev/null 2>&1; then
    PYTHON_EXE="python"
else
    echo "[-] Error: Python 3 is required to run MCP Manager." >&2
    exit 1
fi

exec "$PYTHON_EXE" "$SCRIPT_DIR/scripts/mcp.py" "$@"
