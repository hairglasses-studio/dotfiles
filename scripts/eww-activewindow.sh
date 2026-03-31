#!/usr/bin/env bash
# eww-activewindow.sh — Compositor-aware active window title listener for eww bar

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
source "$SCRIPT_DIR/lib/compositor.sh"

_sway() {
  while true; do
    swaymsg -t get_tree | jq -r 'recurse(.nodes[]?, .floating_nodes[]?) | select(.focused) | .name // ""' | head -1
    swaymsg -t subscribe '["window"]' > /dev/null 2>&1
  done
}

_hyprland() {
  hyprctl activewindow -j 2>/dev/null | jq -r '.title // ""'
  socat -u "UNIX-CONNECT:/tmp/hypr/$HYPRLAND_INSTANCE_SIGNATURE/.socket2.sock" - 2>/dev/null | while read -r line; do
    case "$line" in
      activewindow\>*|activewindowv2\>*)
        hyprctl activewindow -j 2>/dev/null | jq -r '.title // ""'
        ;;
    esac
  done
}

case "$(compositor_type)" in
  hyprland) _hyprland ;;
  sway)     _sway ;;
  *)        echo "" ;;
esac
