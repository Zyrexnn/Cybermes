#!/usr/bin/env bash
# =============================================================================
# Cybermes Native Host Setup & Installer
# Automatically prepares local environment, Python venv, dependencies & MCP tools
# =============================================================================

set -e

CYBERMES_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$CYBERMES_DIR"

echo "========================================================"
echo "  🛡️  Cybermes Native Host Installation & Setup"
echo "========================================================"
echo "Directory: $CYBERMES_DIR"
echo ""

# 1. Check Python 3.10+
if ! command -v python3 >/dev/null 2>&1; then
    echo "❌ Error: Python 3 is not installed. Please install Python 3.10 or higher."
    exit 1
fi

PY_VER=$(python3 -c 'import sys; print(f"{sys.version_info.major}.{sys.version_info.minor}")')
echo "✓ Found Python $PY_VER"

# 2. Check Node.js / npm (Optional for MCP servers)
if command -v node >/dev/null 2>&1 && command -v npm >/dev/null 2>&1; then
    echo "✓ Found Node.js $(node --version) & npm $(npm --version)"
    echo "  Installing MCP servers (Puppeteer & Filesystem)..."
    npm install -g @modelcontextprotocol/server-puppeteer @modelcontextprotocol/server-filesystem 2>/dev/null || \
    npm install --prefix "$CYBERMES_DIR" @modelcontextprotocol/server-puppeteer @modelcontextprotocol/server-filesystem 2>/dev/null || true
else
    echo "⚠️  Node.js/npm not found. Browser MCP automation will be disabled unless Node is installed."
fi

# 3. Create Python Virtual Environment
echo ""
echo "📦 Setting up Python virtual environment (venv)..."
if [ ! -d "$CYBERMES_DIR/venv" ]; then
    python3 -m venv "$CYBERMES_DIR/venv"
fi

source "$CYBERMES_DIR/venv/bin/activate"
pip install --upgrade pip --quiet

echo "📥 Installing Python dependencies from requirements.txt..."
pip install -r "$CYBERMES_DIR/requirements.txt" --quiet

# 4. Install Playwright browser dependencies (if available)
if command -v playwright >/dev/null 2>&1; then
    echo "🌐 Installing Playwright Chromium browser..."
    playwright install chromium 2>/dev/null || true
fi

# 5. Create standard directories
echo ""
echo "📁 Initializing workspace directory structure..."
mkdir -p "$CYBERMES_DIR/reports" \
         "$CYBERMES_DIR/recon" \
         "$CYBERMES_DIR/output" \
         "$CYBERMES_DIR/logs" \
         "$CYBERMES_DIR/targets" \
         "$CYBERMES_DIR/tools/bin" \
         "$CYBERMES_DIR/.hermes/skills"

# Set ACL and permissions so created files/reports are readable by user
if command -v setfacl >/dev/null 2>&1; then
    setfacl -R -d -m u::rwx,g::rwx,o::rwx "$CYBERMES_DIR/reports" "$CYBERMES_DIR/recon" "$CYBERMES_DIR/output" "$CYBERMES_DIR/logs" "$CYBERMES_DIR/skills" "$CYBERMES_DIR/targets" 2>/dev/null || true
    setfacl -R -m u::rwx,g::rwx,o::rwx "$CYBERMES_DIR/reports" "$CYBERMES_DIR/recon" "$CYBERMES_DIR/output" "$CYBERMES_DIR/logs" "$CYBERMES_DIR/skills" "$CYBERMES_DIR/targets" 2>/dev/null || true
fi
chmod -R a+rwX "$CYBERMES_DIR/reports" "$CYBERMES_DIR/recon" "$CYBERMES_DIR/output" "$CYBERMES_DIR/logs" "$CYBERMES_DIR/skills" "$CYBERMES_DIR/targets" 2>/dev/null || true

# 6. Copy skills to Hermes home if needed
if [ -d "$CYBERMES_DIR/skills" ]; then
    cp -r "$CYBERMES_DIR"/skills/* "$CYBERMES_DIR/.hermes/skills/" 2>/dev/null || true
fi

# 7. Setup .env and .hermes/config.yaml from examples
if [ ! -f "$CYBERMES_DIR/.env" ] && [ -f "$CYBERMES_DIR/.env.example" ]; then
    cp "$CYBERMES_DIR/.env.example" "$CYBERMES_DIR/.env"
    echo "✓ Generated default .env file from .env.example"
fi

if [ ! -f "$CYBERMES_DIR/.hermes/config.yaml" ] && [ -f "$CYBERMES_DIR/.hermes/config.yaml.example" ]; then
    cp "$CYBERMES_DIR/.hermes/config.yaml.example" "$CYBERMES_DIR/.hermes/config.yaml"
    echo "✓ Initialized .hermes/config.yaml from .hermes/config.yaml.example"
fi

# 8. Compile High-Performance Go Core Tools (if Go compiler present)
if command -v go >/dev/null 2>&1; then
    echo "⚡ Compiling Cybermes High-Performance Go Tools into tools/bin/..."
    go build -ldflags="-s -w" -o "$CYBERMES_DIR/tools/bin/smart_pipe" "$CYBERMES_DIR/cmd/smart_pipe" 2>/dev/null || true
    go build -ldflags="-s -w" -o "$CYBERMES_DIR/tools/bin/secret_scan" "$CYBERMES_DIR/cmd/secret_scan" 2>/dev/null || true
    go build -ldflags="-s -w" -o "$CYBERMES_DIR/tools/bin/search_knowledge" "$CYBERMES_DIR/cmd/search_knowledge" 2>/dev/null || true
    go build -ldflags="-s -w" -o "$CYBERMES_DIR/tools/bin/aggregate_reports" "$CYBERMES_DIR/cmd/aggregate_reports" 2>/dev/null || true
    echo "✓ Built smart_pipe, secret_scan, search_knowledge, aggregate_reports binaries"
fi

# 9. Set execute permissions
chmod +x "$CYBERMES_DIR/hermes" \
         "$CYBERMES_DIR/bin/hermes" \
         "$CYBERMES_DIR/env.sh" 2>/dev/null || true

if [ -d "$CYBERMES_DIR/tools/bin" ]; then
    chmod +x "$CYBERMES_DIR"/tools/bin/* 2>/dev/null || true
fi

# 10. Configure dynamically in .hermes/config.yaml & sync .env
python3 - <<PYEOF
import os
import re
from pathlib import Path

cybermes_dir = "$CYBERMES_DIR"
config_path = Path(cybermes_dir) / ".hermes/config.yaml"
env_path = Path(cybermes_dir) / ".env"

env_vars = {}
if env_path.exists():
    for line in env_path.read_text(encoding="utf-8").splitlines():
        line = line.strip()
        if line and not line.startswith("#") and "=" in line:
            k, v = line.split("=", 1)
            env_vars[k.strip()] = v.strip().strip("'\"")

api_key = env_vars.get("ROUTER_API_KEY") or env_vars.get("OPENROUTER_API_KEY") or "your_api_key_here"
base_url = env_vars.get("ROUTER_BASE_URL") or env_vars.get("OPENROUTER_BASE_URL") or "http://localhost:20128/v1"
model_name = env_vars.get("HERMES_DEFAULT_MODEL") or "hermes"

if config_path.exists():
    content = config_path.read_text(encoding="utf-8")
    # Replace absolute workspace paths with current CYBERMES_DIR
    content = re.sub(r'root:\s*.*', f'root: {cybermes_dir}', content)
    content = re.sub(r'targets:\s*.*', f'targets: {cybermes_dir}/targets', content)
    content = re.sub(r'recon:\s*.*', f'recon: {cybermes_dir}/recon', content)
    content = re.sub(r'output:\s*.*', f'output: {cybermes_dir}/output', content)
    content = re.sub(r'reports:\s*.*', f'reports: {cybermes_dir}/reports', content)
    content = re.sub(r'logs:\s*.*', f'logs: {cybermes_dir}/logs', content)
    content = re.sub(r'knowledge:\s*.*', f'knowledge: {cybermes_dir}/knowledge', content)
    content = re.sub(r'wordlists:\s*.*', f'wordlists: {cybermes_dir}/tools/wordlists', content)
    content = re.sub(r'directory:\s*.*skills', f'directory: {cybermes_dir}/skills', content)

    # Sync model and provider credentials if available in .env
    content = re.sub(r'default:\s*.*', f'default: {model_name}', content)
    if api_key and api_key != "your_api_key_here":
        content = re.sub(r'api_key:\s*.*', f'api_key: {api_key}', content)
    if base_url:
        content = re.sub(r'base_url:\s*.*', f'base_url: {base_url}', content)

    config_path.write_text(content, encoding="utf-8")

# Sync to .hermes/.env
hermes_env = Path(cybermes_dir) / ".hermes/.env"
if env_path.exists():
    hermes_env.write_text(env_path.read_text(encoding="utf-8"), encoding="utf-8")
PYEOF

echo ""
echo "========================================================"
echo "  ✅ Cybermes Native Setup Complete!"
echo "========================================================"
echo ""
echo "To start using Cybermes natively on your host machine:"
echo ""
echo "  1. Activate the environment:"
echo "     source env.sh"
echo ""
echo "  2. Edit your .env file with your API keys:"
echo "     nano .env"
echo ""
echo "  3. Run Cybermes CLI or Gateway:"
echo "     ./hermes \"Assess http://127.0.0.1:8888\""
echo "     ./hermes gateway run"
echo ""
