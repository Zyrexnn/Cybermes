#!/usr/bin/env bash
set -e
umask 0000

# Fix permissions on host-mounted directories so cloner/user never gets permission denied
fix_permissions() {
    chmod -R a+rwX /workspace/reports /workspace/recon /workspace/output /workspace/logs /workspace/skills /workspace/targets 2>/dev/null || true
    if command -v setfacl >/dev/null 2>&1; then
        setfacl -R -d -m u::rwx,g::rwx,o::rwx /workspace/reports /workspace/recon /workspace/output /workspace/logs /workspace/skills /workspace/targets 2>/dev/null || true
        setfacl -R -m u::rwx,g::rwx,o::rwx /workspace/reports /workspace/recon /workspace/output /workspace/logs /workspace/skills /workspace/targets 2>/dev/null || true
    fi
}

trap fix_permissions EXIT INT TERM

# Ensure working directories exist with open permissions
mkdir -p /workspace/reports /workspace/recon /workspace/output /workspace/logs /workspace/targets
fix_permissions

# Background permission keeper: continuously ensures newly written files stay readable
(
    while true; do
        sleep 2
        chmod -R a+rwX /workspace/reports /workspace/recon /workspace/output /workspace/logs 2>/dev/null || true
    done
) &
BG_PID=$!

cleanup() {
    kill $BG_PID 2>/dev/null || true
    fix_permissions
}
trap cleanup EXIT INT TERM

export PATH="/workspace/tools/bin:/usr/local/bin:$PATH"

if [ "$#" -eq 0 ]; then
    hermes gateway run
    exit $?
fi

if command -v "$1" >/dev/null 2>&1; then
    "$@"
    exit $?
else
    hermes "$@"
    exit $?
fi
