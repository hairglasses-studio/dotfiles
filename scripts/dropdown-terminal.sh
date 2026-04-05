#!/usr/bin/env bash
# dropdown-terminal.sh — Yakuake-style toggle for ralphglasses + claude code
# Works on both Sway and Hyprland
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
source "$SCRIPT_DIR/lib/compositor.sh"

APP_ID="dropdown-cyber"
TMUX_SESSION="dropdown"
WIDTH=5120
HEIGHT=486

_launch_terminal() {
    tmux kill-session -t "$TMUX_SESSION" 2>/dev/null
    foot --app-id "$APP_ID" --font='Maple Mono NF CN:size=18' -- bash -c "
        if command -v tte &>/dev/null; then
            echo 'CYBERDECK ONLINE' | tte beams \
                --beam-gradient-stops 57c7ff ff6ac1 \
                --final-gradient-stops 5af78e 57c7ff \
                --beam-delay 2 2>/dev/null
            sleep 0.3
        fi
        tmux new-session -d -s $TMUX_SESSION \
            'ralphglasses --scan-path $HOME/hairglasses-studio'
        tmux split-window -t $TMUX_SESSION -h -c $HOME/hairglasses-studio/dotfiles \
            'claude'
        tmux select-pane -t $TMUX_SESSION:0.0
        tmux attach-session -t $TMUX_SESSION
    " &
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
sway)
    # Sway — use scratchpad
    if swaymsg -t get_tree | grep -q "\"app_id\": \"$APP_ID\""; then
        swaymsg "[app_id=\"$APP_ID\"] scratchpad show"
    else
        _launch_terminal
    fi
    sleep 0.15
    swaymsg "[app_id=\"$APP_ID\"] resize set $WIDTH $HEIGHT, move position 0 0"
    ;;
esac
