#!/usr/bin/env bash
set -e

umask 0000

# Fix permissions on workspace and output directories
fix_permissions() {
    chmod -R a+rwX /workspace/reports /workspace/recon /workspace/output /workspace/logs /workspace/skills /workspace/targets 2>/dev/null || true
    if command -v setfacl >/dev/null 2>&1; then
        setfacl -R -d -m u::rwx,g::rwx,o::rwx /workspace/reports /workspace/recon /workspace/output /workspace/logs /workspace/skills /workspace/targets 2>/dev/null || true
        setfacl -R -m u::rwx,g::rwx,o::rwx /workspace/reports /workspace/recon /workspace/output /workspace/logs /workspace/skills /workspace/targets 2>/dev/null || true
    fi
}

mkdir -p /workspace/reports /workspace/recon /workspace/output /workspace/logs /workspace/targets /root/.hermes/skills
fix_permissions

# Initialize configuration files if missing
if [ ! -f /root/.hermes/config.yaml ] && [ -f /workspace/.hermes/config.yaml.example ]; then
    cp /workspace/.hermes/config.yaml.example /root/.hermes/config.yaml
fi

if [ ! -f /root/.hermes/auth.json ]; then
    echo "{}" > /root/.hermes/auth.json
fi

if [ ! -f /root/.hermes/.env ] && [ -f /workspace/.env ]; then
    cp /workspace/.env /root/.hermes/.env
fi

if [ -d /workspace/skills ]; then
    cp -r /workspace/skills/* /root/.hermes/skills/ 2>/dev/null || true
fi

trap fix_permissions EXIT INT TERM

export PATH="/opt/hermes-venv/bin:/workspace/tools/bin:/usr/local/bin:$PATH"

if [ "$#" -eq 0 ]; then
    exec hermes gateway run
fi

if command -v "$1" >/dev/null 2>&1; then
    exec "$@"
else
    exec hermes "$@"
fi
