#!/usr/bin/env bash
# dropdown-terminal.sh — Yakuake-style toggle for ralphglasses + claude code

APP_ID="dropdown-cyber"
TMUX_SESSION="dropdown"
WIDTH=5120
HEIGHT=486

# Check if the dropdown window already exists in Sway
if swaymsg -t get_tree | grep -q "\"app_id\": \"$APP_ID\""; then
    # Window exists — toggle scratchpad visibility
    swaymsg "[app_id=\"$APP_ID\"] scratchpad show"
else
    # Kill any orphaned tmux session
    tmux kill-session -t "$TMUX_SESSION" 2>/dev/null

    # Launch foot with the dropdown app_id, running a tmux split layout
    foot --app-id "$APP_ID" --font='Maple Mono NF CN:size=18' -- bash -c "
        # Cyberdeck greeting
        if command -v tte &>/dev/null; then
            echo 'CYBERDECK ONLINE' | tte beams \
                --beam-gradient-stops 57c7ff ff6ac1 \
                --final-gradient-stops 5af78e 57c7ff \
                --beam-delay 2 2>/dev/null
            sleep 0.3
        fi
        tmux new-session -d -s $TMUX_SESSION \
            'ralphglasses --scan-path /home/hg/hairglasses-studio'
        tmux split-window -t $TMUX_SESSION -h -c /home/hg/dotfiles \
            'claude'
        tmux select-pane -t $TMUX_SESSION:0.0
        tmux attach-session -t $TMUX_SESSION
    " &
fi

# Resize after scratchpad show (workaround for swaywm/sway#8493)
sleep 0.15
swaymsg "[app_id=\"$APP_ID\"] resize set $WIDTH $HEIGHT, move position 0 0"
