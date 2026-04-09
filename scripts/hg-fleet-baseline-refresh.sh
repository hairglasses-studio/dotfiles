#!/usr/bin/env bash
# hg-fleet-baseline-refresh.sh — Refresh the cached workspace health matrix.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/lib/hg-core.sh"

DEFAULT_STUDIO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
LOCAL_DIR="${HG_FLEET_LOCAL_DIR:-${HG_STUDIO_ROOT:-$DEFAULT_STUDIO_ROOT}}"

if [ "${1:-}" != "" ]; then
  LOCAL_DIR="$1"
  shift
fi

hg_info "Refreshing fleet baseline cache from $LOCAL_DIR"
python3 "$SCRIPT_DIR/hg-fleet-baseline-refresh.py" --local-dir "$LOCAL_DIR" "$@"
hg_ok "Workspace health matrix refreshed: $LOCAL_DIR/docs/agent-parity/workspace-health-matrix.json"
