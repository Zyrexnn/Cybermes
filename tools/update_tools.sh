#!/usr/bin/env bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BASE_DIR="$(dirname "$SCRIPT_DIR")"
BIN_DIR="$SCRIPT_DIR/bin"

mkdir -p "$BIN_DIR"

# Detect OS (Linux / Darwin/macOS)
RAW_OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
case "$RAW_OS" in
    darwin*)
        PD_OS="macOS"
        ;;
    linux*)
        PD_OS="linux"
        ;;
    *)
        PD_OS="linux"
        ;;
esac

# Detect Architecture (amd64 / arm64)
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
    
    # Match both linux and macOS patterns (e.g., _linux_amd64, _macOS_arm64, _darwin_amd64)
    local latest_url
    latest_url=$(curl -sL "$api_url" | grep -io "https://[^\"]*${tool_name}[^\"]*\(linux\|macos\|darwin\)[^\"]*${PD_ARCH}\.\(zip\|tar\.gz\)" | grep -i "${PD_OS}" | head -n 1 || true)
    
    if [ -z "$latest_url" ]; then
        latest_url=$(curl -sL "$api_url" | grep -io "https://[^\"]*${tool_name}[^\"]*${PD_OS}[^\"]*${PD_ARCH}\.\(zip\|tar\.gz\)" | head -n 1 || true)
    fi

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
        echo "ℹ️  Could not resolve release binary for $tool_name (${PD_OS}/${PD_ARCH}). Skipped."
    fi
}

WITH_NUCLEI=false
for arg in "$@"; do
    if [ "$arg" == "--with-nuclei" ] || [ "$arg" == "-nuclei" ]; then
        WITH_NUCLEI=true
    fi
done

echo "⬇️  Downloading ProjectDiscovery security toolchain (${PD_OS}/${PD_ARCH})..."
TOOLS=("subfinder:subfinder" "httpx:httpx" "katana:katana")

if [ "$WITH_NUCLEI" = true ]; then
    echo "📦 Nuclei included as requested (--with-nuclei)..."
    TOOLS+=("nuclei:nuclei")
else
    echo "ℹ️  Nuclei skipped by default (On-Demand / Optional). Pass --with-nuclei to install."
fi

for item in "${TOOLS[@]}"; do
    IFS=":" read -r repo binary <<< "$item"
    download_pd_tool "$repo" "$binary"
done

if [ "$WITH_NUCLEI" = true ] && [ -f "$BIN_DIR/nuclei" ]; then
    echo "📜 Updating Nuclei community templates..."
    "$BIN_DIR/nuclei" -update-templates -silent 2>/dev/null || true
fi

chmod -R a+rx "$BIN_DIR" 2>/dev/null || true
