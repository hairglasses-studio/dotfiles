#!/usr/bin/env bash
# quickshell-try.sh — launch the Quickshell prototype bar on DP-3.
#
# Not a systemd service — this is an opt-in prototype. When the
# prototype is ready for promotion, add `dotfiles-quickshell.service`
# and wire into install.sh::desktop_service_units.
#
# Usage:
#   quickshell-try.sh            start the prototype in the foreground
#   quickshell-try.sh --bg       start detached; writes PID to /tmp
#   quickshell-try.sh --stop     kill any running prototype

set -uo pipefail

SCRIPT_DIR="$(cd "$(dirname "$(readlink -f "$0")")" && pwd)"
CONFIG="$SCRIPT_DIR/../quickshell/shell.qml"
PID_FILE="/tmp/quickshell-try.pid"
MODE="${1:-fg}"

command -v quickshell >/dev/null 2>&1 || {
  printf 'quickshell not installed — try: sudo pacman -S quickshell\n' >&2
  exit 1
}

case "$MODE" in
  --stop)
    if [[ -f "$PID_FILE" ]]; then
      pid="$(cat "$PID_FILE")"
      if kill -0 "$pid" 2>/dev/null; then
        kill "$pid" && printf 'stopped pid=%s\n' "$pid"
      fi
      rm -f "$PID_FILE"
    fi
    # Belt-and-braces: kill any stray quickshell instance we spawned
    pkill -f "quickshell.*$CONFIG" 2>/dev/null || true
    exit 0
    ;;

  --bg)
    # Detached mode — save PID, redirect stdout/err to a log under /tmp
    quickshell --config "$CONFIG" \
      >/tmp/quickshell-try.log 2>&1 &
    echo $! > "$PID_FILE"
    printf 'started pid=%s → tail /tmp/quickshell-try.log\n' "$!"
    ;;

  --help|-h)
    sed -n '2,11p' "$0" | sed 's/^# \{0,1\}//'
    ;;

  *)
    exec quickshell --config "$CONFIG"
    ;;
esac
