#!/usr/bin/env bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BASE_DIR="$(dirname "$SCRIPT_DIR")"
BIN_DIR="$SCRIPT_DIR/bin"

mkdir -p "$BIN_DIR"

ARCH="$(uname -m)"
case "$ARCH" in
    x86_64|amd64)
        PD_ARCH="amd64"
        ;;
    aarch64|arm64)
        PD_ARCH="arm64"
        ;;
    *)
        PD_ARCH="amd64"
        ;;
esac

download_pd_tool() {
    local repo="$1"
    local tool_name="$2"
    local api_url="https://api.github.com/repos/projectdiscovery/${repo}/releases/latest"
    
    local latest_url
    latest_url=$(curl -sL "$api_url" | grep -o "https://[^\"]*${tool_name}[^\"]*linux_${PD_ARCH}\.\(zip\|tar\.gz\)" | head -n 1 || true)

    if [ -n "$latest_url" ]; then
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
            echo "✓ Installed $tool_name -> $BIN_DIR/$tool_name"
        fi
        rm -rf "$tmp_dir"
    else
        echo "ℹ️  Could not resolve release for $tool_name (linux/$PD_ARCH). Skipped."
    fi
}

echo "⬇️  Downloading ProjectDiscovery security toolchain (linux/$PD_ARCH)..."
TOOLS=("subfinder:subfinder" "httpx:httpx" "katana:katana" "nuclei:nuclei")
for item in "${TOOLS[@]}"; do
    IFS=":" read -r repo binary <<< "$item"
    download_pd_tool "$repo" "$binary"
done

if [ -f "$BIN_DIR/nuclei" ]; then
    echo "📜 Updating Nuclei community templates..."
    "$BIN_DIR/nuclei" -update-templates -silent 2>/dev/null || true
fi

chmod -R a+rx "$BIN_DIR" 2>/dev/null || true
