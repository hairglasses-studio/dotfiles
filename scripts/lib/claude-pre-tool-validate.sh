#!/usr/bin/env bash
# Compatibility shim for legacy Claude hook configs.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
exec "$SCRIPT_DIR/agent-pre-tool-validate.sh" "$@"
