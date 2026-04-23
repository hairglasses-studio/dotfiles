#!/usr/bin/env bash
# quickshell-try.sh — launch the Quickshell pilot bar.
#
# Not a systemd service — this is an opt-in prototype. When the
# The managed service now uses the same runner; this helper stays useful
# for foreground/background manual iteration while the pilot surface runs
# in parallel with the current stack.
#
# Usage:
#   quickshell-try.sh            start the prototype in the foreground
#   quickshell-try.sh --bg       start detached; writes PID to /tmp
#   quickshell-try.sh --stop     kill any running prototype

set -uo pipefail

SCRIPT_DIR="$(cd "$(dirname "$(readlink -f "$0")")" && pwd)"
PID_FILE="/tmp/quickshell-try.pid"
MODE="${1:-fg}"
RUNNER="$SCRIPT_DIR/run-quickshell.sh"

[[ -x "$RUNNER" ]] || {
  printf 'run-quickshell.sh missing or not executable: %s\n' "$RUNNER" >&2
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
    pkill -f "quickshell.*quickshell/shell.qml" 2>/dev/null || true
    exit 0
    ;;

  --bg)
    # Detached mode — save PID, redirect stdout/err to a log under /tmp
    "$RUNNER" \
      >/tmp/quickshell-try.log 2>&1 &
    echo $! > "$PID_FILE"
    printf 'started pid=%s → tail /tmp/quickshell-try.log\n' "$!"
    ;;

  --help|-h)
    sed -n '2,11p' "$0" | sed 's/^# \{0,1\}//'
    ;;

  *)
    exec "$RUNNER"
    ;;
esac
