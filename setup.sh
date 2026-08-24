#!/usr/bin/env bash
set -e

CYBERMES_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$CYBERMES_DIR"

echo "========================================================"
echo "  🛡️  Cybermes Automated Setup & Installer"
echo "========================================================"
echo "Directory: $CYBERMES_DIR"
echo ""

# 1. Python check
if ! command -v python3 >/dev/null 2>&1; then
    echo "❌ Error: Python 3 is not installed. Please install Python 3.10+."
    exit 1
fi
PY_VER=$(python3 -c 'import sys; print(f"{sys.version_info.major}.{sys.version_info.minor}")')
echo "✓ Found Python $PY_VER"

# 2. Workspace directories
mkdir -p "$CYBERMES_DIR/reports" \
         "$CYBERMES_DIR/recon" \
         "$CYBERMES_DIR/output" \
         "$CYBERMES_DIR/logs" \
         "$CYBERMES_DIR/targets" \
         "$CYBERMES_DIR/tools/bin" \
         "$CYBERMES_DIR/.hermes/skills"

# 3. Setup Python Virtual Environment
if [ ! -d "$CYBERMES_DIR/venv" ]; then
    echo "📦 Creating virtual environment (venv)..."
    python3 -m venv "$CYBERMES_DIR/venv"
fi

source "$CYBERMES_DIR/venv/bin/activate"
pip install --upgrade pip --quiet
echo "📥 Installing Python dependencies from requirements.txt..."
pip install -r "$CYBERMES_DIR/requirements.txt" --quiet

# 4. Playwright browser setup
if command -v playwright >/dev/null 2>&1; then
    echo "🌐 Installing Playwright Chromium browser..."
    playwright install chromium 2>/dev/null || true
fi

# 5. Node.js / MCP servers (optional)
if command -v npm >/dev/null 2>&1; then
    echo "📦 Setting up MCP servers..."
    npm install --prefix "$CYBERMES_DIR" @modelcontextprotocol/server-puppeteer @modelcontextprotocol/server-filesystem --silent 2>/dev/null || true
fi

# 6. Synchronize skills
if [ -d "$CYBERMES_DIR/skills" ]; then
    cp -r "$CYBERMES_DIR"/skills/* "$CYBERMES_DIR/.hermes/skills/" 2>/dev/null || true
fi

# 7. Environment files initialization
if [ ! -f "$CYBERMES_DIR/.env" ] && [ -f "$CYBERMES_DIR/.env.example" ]; then
    cp "$CYBERMES_DIR/.env.example" "$CYBERMES_DIR/.env"
    echo "✓ Initialized .env from .env.example"
fi

if [ -f "$CYBERMES_DIR/.env" ] && [ ! -f "$CYBERMES_DIR/.hermes/.env" ]; then
    cp "$CYBERMES_DIR/.env" "$CYBERMES_DIR/.hermes/.env" 2>/dev/null || true
fi

# Sanitize directory traps if any exist
if [ -d "$CYBERMES_DIR/.hermes/config.yaml" ]; then
    rm -rf "$CYBERMES_DIR/.hermes/config.yaml"
fi
if [ -d "$CYBERMES_DIR/.hermes/auth.json" ]; then
    rm -rf "$CYBERMES_DIR/.hermes/auth.json"
fi

if [ ! -f "$CYBERMES_DIR/.hermes/config.yaml" ] && [ -f "$CYBERMES_DIR/.hermes/config.yaml.example" ]; then
    cp "$CYBERMES_DIR/.hermes/config.yaml.example" "$CYBERMES_DIR/.hermes/config.yaml"
    echo "✓ Initialized .hermes/config.yaml"
fi

if [ ! -f "$CYBERMES_DIR/.hermes/auth.json" ]; then
    echo "{}" > "$CYBERMES_DIR/.hermes/auth.json"
fi

# 8. Compile Go Core Tools (if Go is installed)
if command -v go >/dev/null 2>&1; then
    echo "⚡ Compiling Cybermes Go tools..."
    go build -ldflags="-s -w" -o "$CYBERMES_DIR/tools/bin/smart_pipe" "$CYBERMES_DIR/cmd/smart_pipe" 2>/dev/null || true
    go build -ldflags="-s -w" -o "$CYBERMES_DIR/tools/bin/secret_scan" "$CYBERMES_DIR/cmd/secret_scan" 2>/dev/null || true
    go build -ldflags="-s -w" -o "$CYBERMES_DIR/tools/bin/search_knowledge" "$CYBERMES_DIR/cmd/search_knowledge" 2>/dev/null || true
    go build -ldflags="-s -w" -o "$CYBERMES_DIR/tools/bin/aggregate_reports" "$CYBERMES_DIR/cmd/aggregate_reports" 2>/dev/null || true
    echo "✓ Built smart_pipe, secret_scan, search_knowledge, aggregate_reports"
fi

# 9. Download ProjectDiscovery Security Toolchain
if [ -f "$CYBERMES_DIR/tools/update_tools.sh" ]; then
    bash "$CYBERMES_DIR/tools/update_tools.sh" || true
fi

# 10. Fix permissions
chmod +x "$CYBERMES_DIR/cybermes" \
         "$CYBERMES_DIR/hermes" \
         "$CYBERMES_DIR/env.sh" 2>/dev/null || true

if [ -d "$CYBERMES_DIR/tools/bin" ]; then
    chmod +x "$CYBERMES_DIR"/tools/bin/* 2>/dev/null || true
fi

if command -v setfacl >/dev/null 2>&1; then
    setfacl -R -d -m u::rwx,g::rwx,o::rwx "$CYBERMES_DIR/reports" "$CYBERMES_DIR/recon" "$CYBERMES_DIR/output" "$CYBERMES_DIR/logs" 2>/dev/null || true
fi

echo ""
echo "========================================================"
echo "  ✅ Cybermes Setup Complete!"
echo "========================================================"
echo ""
echo "Quick Start:"
echo "  1. Configure your API keys:  nano .env"
echo "  2. Verify environment:       python3 tools/doctor.py"
echo "  3. Configure LLM model:      ./cybermes model"
echo "  4. Run assessment:           ./cybermes \"Assess https://example.com\""
echo ""
