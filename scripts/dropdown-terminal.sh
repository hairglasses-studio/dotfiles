#!/usr/bin/env bash
# dropdown-terminal.sh — Yakuake-style toggle for ralphglasses + claude code
# Hyprland special workspace as scratchpad
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
source "$SCRIPT_DIR/lib/compositor.sh"

APP_ID="dropdown-cyber"
TMUX_SESSION="dropdown"
WIDTH=5120
HEIGHT=486

_launch_terminal() {
    tmux kill-session -t "$TMUX_SESSION" 2>/dev/null
    ghostty --class="$APP_ID" --font-size=18 -e bash -c "
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
esac
