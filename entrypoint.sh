#!/usr/bin/env bash
set -e
umask 000
export PATH="/workspace/tools/bin:/usr/local/bin:$PATH"

if [ "$#" -eq 0 ]; then
  exec hermes gateway run
fi

if command -v "$1" >/dev/null 2>&1; then
  exec "$@"
else
  exec hermes "$@"
fi
