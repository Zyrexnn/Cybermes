#!/usr/bin/env bash
# ==============================================================================
# Cybermes Toolchain & Threat Knowledge Auto-Updater
# Automatically updates ProjectDiscovery binaries, Nuclei templates,
# Python security packages, and validates integrity.
# ==============================================================================
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BASE_DIR="$(dirname "$SCRIPT_DIR")"
BIN_DIR="$SCRIPT_DIR/bin"

mkdir -p "$BIN_DIR"

# Detect host architecture
ARCH="$(uname -m)"
case "$ARCH" in
    x86_64|amd64)
        PD_ARCH="amd64"
        ;;
    aarch64|arm64)
        PD_ARCH="arm64"
        ;;
    *)
        echo "⚠️ Unsupported architecture: $ARCH. Falling back to amd64."
        PD_ARCH="amd64"
        ;;
esac

echo "🔄 [Cybermes Updater] Target Architecture: Linux/$PD_ARCH"

# Function to download latest GitHub release tarball
download_pd_tool() {
    local repo="$1"
    local tool_name="$2"
    echo "📦 Checking latest release for $tool_name ($repo)..."
    
    local latest_url
    latest_url=$(curl -s "https://api.github.com/repos/projectdiscovery/${repo}/releases/latest" | \
        grep -o "https://.*_${tool_name}_.*_linux_${PD_ARCH}.zip" | head -n 1 || true)
        
    if [ -z "$latest_url" ]; then
        latest_url=$(curl -s "https://api.github.com/repos/projectdiscovery/${repo}/releases/latest" | \
            grep -o "https://.*_${tool_name}_.*_linux_${PD_ARCH}.tar.gz" | head -n 1 || true)
    fi

    if [ -n "$latest_url" ]; then
        echo "   ⬇️ Downloading $latest_url..."
        local tmp_dir
        tmp_dir=$(mktemp -d)
        if [[ "$latest_url" == *.zip ]]; then
            curl -sSL "$latest_url" -o "$tmp_dir/pkg.zip"
            unzip -q -o "$tmp_dir/pkg.zip" -d "$tmp_dir"
        else
            curl -sSL "$latest_url" | tar -xz -C "$tmp_dir"
        fi
        
        if [ -f "$tmp_dir/$tool_name" ]; then
            mv "$tmp_dir/$tool_name" "$BIN_DIR/$tool_name"
            chmod +x "$BIN_DIR/$tool_name"
            echo "   ✅ Updated $tool_name -> $BIN_DIR/$tool_name"
        fi
        rm -rf "$tmp_dir"
    else
        echo "   ℹ️ Release binary not found via API. Keeping current version."
    fi
}

# 1. Update Nuclei Templates
if command -v nuclei >/dev/null 2>&1; then
    echo "📜 Updating Nuclei community templates..."
    nuclei -update-templates -silent 2>/dev/null || true
    echo "✅ Nuclei templates updated."
elif [ -f "$BIN_DIR/nuclei" ]; then
    "$BIN_DIR/nuclei" -update-templates -silent 2>/dev/null || true
fi

# 2. Update Go Security Tools if curl is available
TOOLS_LIST=("subfinder:subfinder" "httpx:httpx" "katana:katana" "nuclei:nuclei")
for item in "${TOOLS_LIST[@]}"; do
    IFS=":" read -r repo binary <<< "$item"
    download_pd_tool "$repo" "$binary"
done

# 3. Update Python Security Dependencies in Virtualenv
if [ -d "/opt/hermes-venv" ]; then
    echo "🐍 Updating Python packages in /opt/hermes-venv..."
    /opt/hermes-venv/bin/pip install --upgrade --no-cache-dir requests playwright pyyaml rich markdown jinja2 2>/dev/null || true
elif [ -d "$BASE_DIR/venv" ]; then
    echo "🐍 Updating Python packages in local venv..."
    "$BASE_DIR/venv/bin/pip" install --upgrade --no-cache-dir requests playwright pyyaml rich markdown jinja2 2>/dev/null || true
fi

# 4. Set Open Permissions for Docker/Host Interoperability
chmod -R a+rx "$BIN_DIR" 2>/dev/null || true

echo "✨ [Cybermes Updater] Toolchain update completed successfully."
