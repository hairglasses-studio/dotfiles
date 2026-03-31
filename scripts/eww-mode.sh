#!/usr/bin/env bash
# eww-mode.sh — Compositor-aware mode/submap listener for eww bar

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
source "$SCRIPT_DIR/lib/compositor.sh"

_sway() {
  echo "default"
  swaymsg -t subscribe '["mode"]' | jq -r --unbuffered '.change'
}

_hyprland() {
  echo "default"
  socat -u "UNIX-CONNECT:/tmp/hypr/$HYPRLAND_INSTANCE_SIGNATURE/.socket2.sock" - 2>/dev/null | while read -r line; do
    case "$line" in
      submap\>\>)
        echo "default"
        ;;
      submap\>\>*)
        echo "${line#submap>>}"
        ;;
    esac
  done
}

case "$(compositor_type)" in
  hyprland) _hyprland ;;
  sway)     _sway ;;
  *)        echo "default" ;;
esac
