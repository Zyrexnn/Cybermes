#!/usr/bin/env bash
set -e
export PATH="/workspace/tools/bin:/usr/local/bin:$PATH"

if [ "$#" -eq 0 ]; then
  exec hermes
fi

if command -v "$1" >/dev/null 2>&1; then
  exec "$@"
else
  exec hermes "$@"
fi
