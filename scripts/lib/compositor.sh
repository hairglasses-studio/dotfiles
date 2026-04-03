#!/usr/bin/env bash
# compositor.sh — Shared compositor detection and IPC abstraction
# Source this file: source "$(dirname "$0")/lib/compositor.sh"

# Returns: hyprland, sway, aerospace, or unknown
compositor_type() {
  if [[ -n "${HYPRLAND_INSTANCE_SIGNATURE:-}" ]]; then
    echo "hyprland"
  elif [[ -n "${SWAYSOCK:-}" ]]; then
    echo "sway"
  elif [[ "$(uname -s)" == "Darwin" ]]; then
    echo "aerospace"
  else
    echo "unknown"
  fi
}

# Send a dispatch/command to the active compositor
# Usage: compositor_msg "workspace 3" or compositor_msg "reload"
compositor_msg() {
  local cmd="$1"
  case "$(compositor_type)" in
    hyprland) hyprctl dispatch "$cmd" 2>/dev/null ;;
    sway)     swaymsg "$cmd" 2>/dev/null ;;
    aerospace) aerospace "$cmd" 2>/dev/null ;;
  esac
}

# Query compositor state. Returns JSON.
# Usage: compositor_query workspaces|windows|activewindow|outputs
compositor_query() {
  local what="$1"
  case "$(compositor_type)" in
    hyprland)
      case "$what" in
        workspaces)   hyprctl workspaces -j 2>/dev/null ;;
        windows)      hyprctl clients -j 2>/dev/null ;;
        activewindow) hyprctl activewindow -j 2>/dev/null ;;
        outputs)      hyprctl monitors -j 2>/dev/null ;;
      esac
      ;;
    sway)
      case "$what" in
        workspaces)   swaymsg -t get_workspaces 2>/dev/null ;;
        windows)      swaymsg -t get_tree 2>/dev/null ;;
        activewindow) swaymsg -t get_tree 2>/dev/null | jq '[recurse(.nodes[]?, .floating_nodes[]?) | select(.focused)] | .[0]' 2>/dev/null ;;
        outputs)      swaymsg -t get_outputs 2>/dev/null ;;
      esac
      ;;
  esac
}

# Get primary output name
compositor_output() {
  case "$(compositor_type)" in
    hyprland) hyprctl monitors -j 2>/dev/null | jq -r '.[0].name // empty' 2>/dev/null ;;
    sway)     swaymsg -t get_outputs 2>/dev/null | jq -r '.[0].name // empty' 2>/dev/null ;;
  esac
}

# Get Hyprland socket2 path (supports XDG_RUNTIME_DIR and /tmp fallback)
hypr_socket2() {
  local xdg="${XDG_RUNTIME_DIR:-/run/user/$(id -u)}"
  local sig="$HYPRLAND_INSTANCE_SIGNATURE"
  for path in "$xdg/hypr/$sig/.socket2.sock" "/tmp/hypr/$sig/.socket2.sock"; do
    [[ -S "$path" ]] && echo "$path" && return
  done
  echo "/tmp/hypr/$sig/.socket2.sock"
}

# Resilient listener wrapper — restarts the given command on exit with backoff.
# Usage: resilient_listen <command> [args...]
# The command is expected to block (e.g. socat, pactl subscribe, cava).
# On exit, waits with exponential backoff (1s → 30s cap) before restart.
resilient_listen() {
  local backoff=1
  while true; do
    "$@" && backoff=1 || true
    sleep "$backoff"
    backoff=$(( backoff < 30 ? backoff * 2 : 30 ))
  done
}

# Subscribe to compositor events (blocking, pipe-friendly)
# Usage: compositor_subscribe workspace | while read -r line; do ...; done
compositor_subscribe() {
  local events="$1"
  case "$(compositor_type)" in
    hyprland)
      socat -u "UNIX-CONNECT:$(hypr_socket2)" - 2>/dev/null
      ;;
    sway)
      swaymsg -t subscribe "[\"$events\"]" 2>/dev/null
      ;;
  esac
}

# Reload the active compositor config
compositor_reload() {
  case "$(compositor_type)" in
    hyprland) hyprctl reload 2>/dev/null ;;
    sway)     swaymsg reload 2>/dev/null ;;
  esac
}

# Switch workspace by number
compositor_workspace() {
  local num="$1"
  case "$(compositor_type)" in
    hyprland) hyprctl dispatch workspace "$num" 2>/dev/null ;;
    sway)     swaymsg workspace number "$num" 2>/dev/null ;;
  esac
}
