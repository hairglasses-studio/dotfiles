#!/usr/bin/env bash
set -euo pipefail

SCRIPT_PATH="$(readlink -f "${BASH_SOURCE[0]:-$0}")"
SCRIPT_DIR="$(cd "$(dirname "$SCRIPT_PATH")" && pwd)"
source "$SCRIPT_DIR/lib/hg-core.sh"
source "$SCRIPT_DIR/lib/tmux-persistence.sh"

quiet=false
[[ "${1:-}" == "--quiet" ]] && quiet=true

if tmux_persistence_bootstrap; then
  if ! $quiet; then
    hg_ok "tmux persistence bootstrap ready"
  fi
else
  if ! $quiet; then
    hg_error "tmux persistence bootstrap failed"
  fi
  exit 1
fi
