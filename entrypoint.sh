#!/usr/bin/env bash
set -e

# Set universal read/write/execute creation mask for container processes
umask 0000

# Fix permissions on host-mounted directories so host user never gets permission denied
fix_permissions() {
    chmod -R a+rwX /workspace/reports /workspace/recon /workspace/output /workspace/logs /workspace/skills /workspace/targets 2>/dev/null || true
    if command -v setfacl >/dev/null 2>&1; then
        # Apply default and access ACLs for automatic inheritance on all new files
        setfacl -R -d -m u::rwx,g::rwx,o::rwx /workspace/reports /workspace/recon /workspace/output /workspace/logs /workspace/skills /workspace/targets 2>/dev/null || true
        setfacl -R -m u::rwx,g::rwx,o::rwx /workspace/reports /workspace/recon /workspace/output /workspace/logs /workspace/skills /workspace/targets 2>/dev/null || true
    fi
}

# Ensure working directories exist with open permissions and default ACL inheritance
mkdir -p /workspace/reports /workspace/recon /workspace/output /workspace/logs /workspace/targets
fix_permissions

# Register exit/interrupt traps for graceful cleanup
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
