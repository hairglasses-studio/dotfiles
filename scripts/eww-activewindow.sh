#!/usr/bin/env bash
# eww-activewindow.sh — Compositor-aware active window title listener for eww bar

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
source "$SCRIPT_DIR/lib/compositor.sh"

_hyprland() {
  _hypr_listen() {
    hyprctl activewindow -j 2>/dev/null | jq -r '.title // ""'
    socat -u "UNIX-CONNECT:$(hypr_socket2)" - 2>/dev/null | while read -r line; do
      case "$line" in
        activewindow\>*|activewindowv2\>*)
          hyprctl activewindow -j 2>/dev/null | jq -r '.title // ""'
          ;;
      esac
    done
  }

  resilient_listen _hypr_listen
}

case "$(compositor_type)" in
  hyprland) _hyprland ;;
  *)        echo "" ;;
esac
