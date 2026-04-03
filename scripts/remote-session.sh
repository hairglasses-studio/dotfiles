#!/usr/bin/env bash
# remote-session.sh — Expose fleet status via a simple HTTP endpoint
# Usage: remote-session.sh [start|stop|status]
#
# Serves fleet diagnostics on port 9876 (LAN-only by default).
# WARNING: This is a diagnostic endpoint for LAN use only.
# Do NOT expose to the internet without authentication.
#
# Environment:
#   REMOTE_SESSION_BIND — bind address (default: 127.0.0.1, set to 0.0.0.0 for LAN)
#   REMOTE_SESSION_PORT — listen port (default: 9876)
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/lib/hg-core.sh"

hg_require python3

BIND="${REMOTE_SESSION_BIND:-127.0.0.1}"
PORT="${REMOTE_SESSION_PORT:-9876}"
SERVE_DIR="/tmp/remote-session"
PID_FILE="/tmp/remote-session.pid"
FLEET_SOURCE="/tmp/ralphglasses-active.json"
AUDIT_LOG="$HOME/.local/state/hg/fleet-audit.json"

# ── Helpers ───────────────────────────────────────

_prepare_serve_dir() {
  mkdir -p "$SERVE_DIR"

  # Symlink fleet status if source exists
  if [[ -f "$FLEET_SOURCE" ]]; then
    ln -sf "$FLEET_SOURCE" "$SERVE_DIR/fleet.json"
  else
    echo '{"status":"no fleet data","ts":"'"$(date -Iseconds)"'"}' > "$SERVE_DIR/fleet.json"
  fi

  # Generate audit summary
  if [[ -f "$AUDIT_LOG" ]]; then
    cp "$AUDIT_LOG" "$SERVE_DIR/audit.json"
  else
    echo '{"status":"no audit data","ts":"'"$(date -Iseconds)"'"}' > "$SERVE_DIR/audit.json"
  fi

  # Health endpoint
  echo '{"status":"ok","ts":"'"$(date -Iseconds)"'","bind":"'"$BIND"'","port":'"$PORT"'}' > "$SERVE_DIR/health"
}

_is_running() {
  [[ -f "$PID_FILE" ]] && kill -0 "$(cat "$PID_FILE")" 2>/dev/null
}

# ── Commands ──────────────────────────────────────

cmd_start() {
  if _is_running; then
    hg_warn "Already running (PID $(cat "$PID_FILE"))"
    return 0
  fi

  _prepare_serve_dir

  hg_info "Starting HTTP server on $BIND:$PORT (serving $SERVE_DIR)"

  python3 -m http.server "$PORT" \
    --bind "$BIND" \
    --directory "$SERVE_DIR" \
    &>/dev/null &

  echo $! > "$PID_FILE"
  hg_ok "Server started (PID $!)"
}

cmd_stop() {
  if ! _is_running; then
    hg_warn "Not running"
    return 0
  fi

  PID=$(cat "$PID_FILE")
  kill "$PID" 2>/dev/null && rm -f "$PID_FILE"
  hg_ok "Server stopped (PID $PID)"
}

cmd_status() {
  if _is_running; then
    PID=$(cat "$PID_FILE")
    hg_ok "Running (PID $PID) on $BIND:$PORT"
    echo ""
    echo "  Endpoints:"
    echo "    http://$BIND:$PORT/fleet.json"
    echo "    http://$BIND:$PORT/audit.json"
    echo "    http://$BIND:$PORT/health"
  else
    hg_warn "Not running"
    return 1
  fi
}

# ── Main ──────────────────────────────────────────

case "${1:-status}" in
  start)  cmd_start  ;;
  stop)   cmd_stop   ;;
  status) cmd_status ;;
  *)      hg_die "Usage: remote-session.sh [start|stop|status]" ;;
esac
