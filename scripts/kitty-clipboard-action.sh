#!/usr/bin/env bash
# kitty-clipboard-action.sh — mouse-friendly copy/paste helpers with Kitty focus awareness
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$(readlink -f "${BASH_SOURCE[0]}")")" && pwd)"
source "$SCRIPT_DIR/lib/compositor.sh"

action="${1:-}"

active_class() {
  compositor_query activewindow 2>/dev/null | jq -r '.class // ""' 2>/dev/null | tr '[:upper:]' '[:lower:]'
}

ydotool_socket_path() {
  printf '%s\n' "${YDOTOOL_SOCKET:-${XDG_RUNTIME_DIR:-/run/user/$(id -u)}/.ydotool_socket}"
}

send_keys() {
  command -v ydotool >/dev/null 2>&1 || {
    printf 'ydotool is required for clipboard actions\n' >&2
    exit 1
  }
  YDOTOOL_SOCKET="$(ydotool_socket_path)" ydotool key "$@"
}

is_kitty=false
[[ "$(active_class)" == "kitty" ]] && is_kitty=true

case "$action" in
  copy)
    if $is_kitty; then
      send_keys 29:1 42:1 46:1 46:0 42:0 29:0
    else
      send_keys 29:1 46:1 46:0 29:0
    fi
    ;;
  paste)
    send_keys 29:1 47:1 47:0 29:0
    ;;
  paste-selection)
    if $is_kitty; then
      send_keys 29:1 42:1 47:1 47:0 42:0 29:0
    fi
    ;;
  *)
    printf 'Usage: %s {copy|paste|paste-selection}\n' "$(basename "$0")" >&2
    exit 2
    ;;
esac
