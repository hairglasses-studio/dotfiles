#!/usr/bin/env bash
# hyprshell-trigger.sh — invoke hyprshell launcher surfaces from mouse actions
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$(readlink -f "${BASH_SOURCE[0]}")")" && pwd)"
source "$SCRIPT_DIR/lib/compositor.sh"

action="${1:-}"

ydotool_socket_path() {
  printf '%s\n' "${YDOTOOL_SOCKET:-${XDG_RUNTIME_DIR:-/run/user/$(id -u)}/.ydotool_socket}"
}

send_keys() {
  command -v ydotool >/dev/null 2>&1 || return 1
  YDOTOOL_SOCKET="$(ydotool_socket_path)" ydotool key "$@"
}

hyprshell_running() {
  pgrep -u "$USER" -f '[h]yprshell' >/dev/null 2>&1
}

trigger_overview() {
  if hyprshell_running && send_keys 125:1 32:1 32:0 125:0; then
    return 0
  fi

  exec "$SCRIPT_DIR/app-launcher.sh"
}

trigger_switcher() {
  if hyprshell_running && send_keys 56:1 15:1 15:0 56:0; then
    return 0
  fi

  exec "$SCRIPT_DIR/app-switcher.sh"
}

case "$action" in
  overview)
    trigger_overview
    ;;
  switch)
    trigger_switcher
    ;;
  *)
    printf 'Usage: %s {overview|switch}\n' "$(basename "$0")" >&2
    exit 2
    ;;
esac
