#!/usr/bin/env bash
# dropdown-terminal.sh — Yakuake-style toggle for ralphglasses + claude code
# Hyprland special workspace as scratchpad
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
source "$SCRIPT_DIR/lib/compositor.sh"

APP_ID="dropdown-cyber"
LAUNCHER="${KITTY_DROPDOWN_LAUNCHER:-$SCRIPT_DIR/kitty-visual-launch.sh}"

_launch_terminal() {
    "$LAUNCHER" --class="$APP_ID" -o font_size=18 -e "$SCRIPT_DIR/dropdown-session.sh" &
}

case "$(compositor_type)" in
hyprland)
    # Hyprland — use special workspace as scratchpad
    if hyprctl clients -j | jq -e ".[] | select(.class == \"$APP_ID\")" > /dev/null 2>&1; then
        hyprctl dispatch togglespecialworkspace dropdown
    else
        _launch_terminal
        sleep 0.3
        hyprctl dispatch movetoworkspacesilent "special:dropdown,class:^($APP_ID)$"
        hyprctl dispatch togglespecialworkspace dropdown
    fi
    ;;
esac
