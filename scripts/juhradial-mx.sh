#!/usr/bin/env bash
# juhradial-mx.sh — start the juhradial daemon and overlay from repo-managed paths
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/lib/juhradial.sh"

quiet=false
overlay_only=false
restart=false

for arg in "$@"; do
  case "$arg" in
    --quiet) quiet=true ;;
    --overlay-only) overlay_only=true ;;
    --restart) restart=true ;;
    *)
      printf 'Unknown option: %s\n' "$arg" >&2
      exit 2
      ;;
  esac
done

log() {
  $quiet || printf '[juhradial-mx] %s\n' "$*"
}

overlay_script="$(juhradial_overlay_script)"
if [[ ! -f "$overlay_script" ]]; then
  printf 'juhradial overlay not installed at %s\n' "$overlay_script" >&2
  printf 'Run %s/scripts/juhradial-install.sh first.\n' "$(juhradial_dotfiles_dir)" >&2
  exit 1
fi

if ! $overlay_only; then
  if $restart; then
    juhradial_systemctl restart juhradialmx-daemon.service >/dev/null
    log "Restarted juhradialmx-daemon.service"
  elif ! juhradial_systemctl is-active juhradialmx-daemon.service >/dev/null 2>&1; then
    juhradial_systemctl start --no-block juhradialmx-daemon.service >/dev/null
    log "Started juhradialmx-daemon.service"
  fi
fi

if $restart; then
  pkill -f 'juhradial-overlay(\.py)?' >/dev/null 2>&1 || true
  sleep 0.2
fi

if juhradial_overlay_running; then
  log "Overlay already running"
  exit 0
fi

nohup python3 "$overlay_script" >/dev/null 2>&1 &
log "Started juhradial overlay"
