#!/usr/bin/env bash
set -euo pipefail

# app-switcher.sh — stable Alt+Tab switcher entrypoint.

SCRIPT_PATH="$(readlink -f "${BASH_SOURCE[0]:-$0}")"
SCRIPT_DIR="$(cd "$(dirname "$SCRIPT_PATH")" && pwd)"

case "${1:-open}" in
  open|reverse|"")
    exec "$SCRIPT_DIR/menu-control.sh" windows
    ;;
  close)
    exec "$SCRIPT_DIR/menu-control.sh" close
    ;;
  *)
    printf 'Usage: %s {open|reverse|close}\n' "$(basename "$0")" >&2
    exit 2
    ;;
esac
