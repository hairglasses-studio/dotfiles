#!/usr/bin/env bash
# notification-control.sh - route notification shortcuts through Quickshell.

set -euo pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
dotfiles_dir="$(cd "$script_dir/.." && pwd)"
config_path="${QUICKSHELL_CONFIG_PATH:-$dotfiles_dir/quickshell/shell.qml}"
quickshell_bin="${QUICKSHELL_BIN:-quickshell}"
swaync_client="${SWAYNC_CLIENT_BIN:-swaync-client}"
bridge="${NOTIFICATION_BRIDGE:-$script_dir/notification-bridge.py}"
ipc_timeout="${QUICKSHELL_IPC_TIMEOUT:-0.8}"

usage() {
  cat <<'EOF'
Usage: notification-control.sh <toggle-center|show-center|hide-center|toggle-dnd|dnd-on|dnd-off|dismiss-all|clear-history|toggle-quick-settings|status>

Prefers Quickshell IPC when a shell instance is running. Falls back to swaync
or the local notification bridge for rollback modes.
EOF
}

qs_call() {
  command -v "$quickshell_bin" >/dev/null 2>&1 || return 1
  timeout "$ipc_timeout" "$quickshell_bin" ipc --path "$config_path" --newest call shell "$@" >/dev/null 2>&1
}

qs_call_output() {
  command -v "$quickshell_bin" >/dev/null 2>&1 || return 1
  timeout "$ipc_timeout" "$quickshell_bin" ipc --path "$config_path" --newest call shell "$@"
}

swaync_call() {
  command -v "$swaync_client" >/dev/null 2>&1 || return 1
  "$swaync_client" "$@" >/dev/null 2>&1
}

bridge_call() {
  python3 "$bridge" "$@" >/dev/null
}

action="${1:-}"
case "$action" in
  toggle-center)
    qs_call toggleNotifications || swaync_call -t -sw
    ;;
  show-center)
    qs_call showNotifications || swaync_call -op -sw
    ;;
  hide-center)
    qs_call hideNotifications || swaync_call -cp -sw
    ;;
  toggle-dnd)
    qs_call toggleDnd || bridge_call --dnd toggle
    ;;
  dnd-on)
    qs_call setDnd true || bridge_call --dnd true
    ;;
  dnd-off)
    qs_call setDnd false || bridge_call --dnd false
    ;;
  dismiss-all)
    qs_call closeNotifications || bridge_call --close-all
    ;;
  clear-history)
    qs_call clearNotifications || bridge_call --clear-history
    ;;
  toggle-quick-settings)
    qs_call toggleQuickSettings || true
    ;;
  status)
    if status_output="$(qs_call_output status 2>/dev/null)"; then
      printf '%s\n' "$status_output"
    else
      python3 "$bridge" --limit 5
    fi
    ;;
  -h|--help|help)
    usage
    ;;
  *)
    usage >&2
    exit 2
    ;;
esac
