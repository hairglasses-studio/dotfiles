#!/usr/bin/env bash
# juhradial-mx.sh — start the juhradial daemon and overlay from repo-managed paths
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$(readlink -f "${BASH_SOURCE[0]}")")" && pwd)"
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
overlay_unit="juhradial-overlay"
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
  juhradial_systemctl stop "${overlay_unit}.service" >/dev/null 2>&1 || true
  pkill -f 'juhradial-overlay(\.py)?' >/dev/null 2>&1 || true
  sleep 0.2
fi

if juhradial_overlay_running; then
  log "Overlay already running"
  exit 0
fi

juhradial_export_graphical_env
if command -v systemd-run >/dev/null 2>&1; then
  desktop_file="$(juhradial_desktop_file)"
  env \
    XDG_RUNTIME_DIR="$(juhradial_runtime_dir)" \
    DBUS_SESSION_BUS_ADDRESS="$(juhradial_user_bus)" \
    systemd-run \
      --user \
      --unit="$overlay_unit" \
      --quiet \
      --collect \
      /usr/bin/env \
      XDG_RUNTIME_DIR="$XDG_RUNTIME_DIR" \
      DBUS_SESSION_BUS_ADDRESS="$DBUS_SESSION_BUS_ADDRESS" \
      DISPLAY="${DISPLAY:-}" \
      WAYLAND_DISPLAY="${WAYLAND_DISPLAY:-}" \
      HYPRLAND_INSTANCE_SIGNATURE="${HYPRLAND_INSTANCE_SIGNATURE:-}" \
      XDG_CURRENT_DESKTOP="${XDG_CURRENT_DESKTOP:-}" \
      XDG_SESSION_TYPE="${XDG_SESSION_TYPE:-}" \
      GIO_LAUNCHED_DESKTOP_FILE="$desktop_file" \
      GIO_LAUNCHED_DESKTOP_FILE_PID="$$" \
      python3 "$overlay_script" >/dev/null
else
  nohup python3 "$overlay_script" </dev/null >/dev/null 2>&1 &
fi

log "Started juhradial overlay"
