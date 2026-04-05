#!/usr/bin/env bash
# eww-mode.sh — Compositor-aware mode/submap listener for eww bar

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
source "$SCRIPT_DIR/lib/compositor.sh"

_hyprland() {
  _hypr_listen() {
    echo "default"
    socat -u "UNIX-CONNECT:$(hypr_socket2)" - 2>/dev/null | while read -r line; do
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

  resilient_listen _hypr_listen
}

case "$(compositor_type)" in
  hyprland) _hyprland ;;
  *)        echo "default" ;;
esac
