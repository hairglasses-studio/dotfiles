#!/usr/bin/env bash
# dropdown-terminal.sh — Yakuake-style toggle for ralphglasses + claude code

APP_ID="dropdown-cyber"
TMUX_SESSION="dropdown"

# Check if the dropdown window already exists in Sway
if swaymsg -t get_tree | grep -q "\"app_id\": \"$APP_ID\""; then
    # Window exists — toggle scratchpad visibility
    swaymsg "[app_id=\"$APP_ID\"] scratchpad show"
else
    # Kill any orphaned tmux session
    tmux kill-session -t "$TMUX_SESSION" 2>/dev/null

    # Launch foot with the dropdown app_id, running a tmux split layout
    foot --app-id "$APP_ID" --font='Maple Mono NF CN:size=18' -- bash -c "
        tmux new-session -d -s $TMUX_SESSION \
            'ralphglasses --scan-path /home/hg/hairglasses-studio'
        tmux split-window -t $TMUX_SESSION -h -c /home/hg/dotfiles \
            'claude'
        tmux select-pane -t $TMUX_SESSION:0.0
        tmux attach-session -t $TMUX_SESSION
    "
fi
