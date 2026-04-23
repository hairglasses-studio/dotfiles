#!/usr/bin/env bash
# menu-control.sh - Quickshell menu IPC with rollback fallbacks.

set -euo pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
dotfiles_dir="$(cd "$script_dir/.." && pwd)"
config_path="${QUICKSHELL_CONFIG_PATH:-$dotfiles_dir/quickshell/shell.qml}"
quickshell_bin="${QUICKSHELL_BIN:-quickshell}"
ipc_timeout="${QUICKSHELL_IPC_TIMEOUT:-0.8}"

usage() {
  cat <<'EOF'
Usage: menu-control.sh <apps|windows|power|emoji|agents|clipboard|close|status>

Prefers Quickshell IPC. If Quickshell is unavailable, falls back to the
previous wofi/wlogout/clipse-based menu surface.
EOF
}

qs_call() {
  command -v "$quickshell_bin" >/dev/null 2>&1 || return 1
  timeout "$ipc_timeout" "$quickshell_bin" ipc --path "$config_path" --newest call menus "$@" >/dev/null 2>&1
}

qs_call_output() {
  command -v "$quickshell_bin" >/dev/null 2>&1 || return 1
  timeout "$ipc_timeout" "$quickshell_bin" ipc --path "$config_path" --newest call menus "$@"
}

fallback_apps() {
  if command -v wofi >/dev/null 2>&1; then
    exec wofi --show drun
  fi
  if command -v rofi >/dev/null 2>&1; then
    exec rofi -show drun
  fi
  return 1
}

fallback_windows() {
  command -v hyprctl >/dev/null 2>&1 || return 1
  command -v jq >/dev/null 2>&1 || return 1
  local clients selection address
  clients="$(hyprctl clients -j | jq -r '.[] | select((.address // "") != "") | "\(.address) \(.class // "app") — \(.title // "")"')"
  [[ -n "$clients" ]] || return 0
  if command -v wofi >/dev/null 2>&1; then
    selection="$(printf '%s\n' "$clients" | wofi --dmenu --prompt "Switch to" --insensitive || true)"
  elif command -v rofi >/dev/null 2>&1; then
    selection="$(printf '%s\n' "$clients" | rofi -dmenu -i -p "Switch to" || true)"
  else
    return 1
  fi
  address="${selection%% *}"
  [[ -n "$address" ]] || return 0
  exec hyprctl dispatch focuswindow "address:${address}"
}

fallback_agents() {
  command -v hyprctl >/dev/null 2>&1 || return 1
  command -v jq >/dev/null 2>&1 || return 1
  command -v wofi >/dev/null 2>&1 || return 1
  local windows selected addr
  windows="$(hyprctl clients -j | jq -r '.[] | select(.class == "kitty" and (.title | test("^────|[Cc]laude|[Cc]odex"))) | "\(.address)\t\(.title)"' | sort -t$'\t' -k2)"
  if [[ -z "$windows" ]]; then
    notify-send -a "Agent Sessions" "No agent sessions found" 2>/dev/null || true
    return 0
  fi
  selected="$(printf '%s\n' "$windows" | cut -f2 | wofi --dmenu -p "Agent Sessions" --width 600 --height 400 || true)"
  [[ -n "$selected" ]] || return 0
  addr="$(printf '%s\n' "$windows" | grep -F "$selected" | head -1 | cut -f1)"
  [[ -n "$addr" ]] || return 0
  exec hyprctl dispatch focuswindow "address:$addr"
}

fallback_clipboard() {
  if command -v clipse >/dev/null 2>&1; then
    exec "${TERMINAL:-kitty}" --class=clipse -e clipse
  fi
  return 1
}

action="${1:-apps}"
case "$action" in
  apps|windows|power|emoji|agents|clipboard)
    qs_call open "$action" && exit 0
    case "$action" in
      apps) fallback_apps ;;
      windows) fallback_windows ;;
      power) exec wlogout -p layer-shell ;;
      emoji) exec wofi-emoji ;;
      agents) fallback_agents ;;
      clipboard) fallback_clipboard ;;
    esac
    ;;
  close)
    qs_call close || true
    ;;
  status)
    if status_output="$(qs_call_output status 2>/dev/null)"; then
      printf '%s\n' "$status_output"
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
