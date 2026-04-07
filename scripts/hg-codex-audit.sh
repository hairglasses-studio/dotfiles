#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "warning: hg-codex-audit.sh is deprecated; use hg-agent-parity-audit.sh" >&2
exec "$SCRIPT_DIR/hg-agent-parity-audit.sh" "$@"
