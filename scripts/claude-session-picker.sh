#!/usr/bin/env bash
# claude-session-picker.sh — Focus a Claude Code session by name via wofi
# Queries Hyprland IPC for Ghostty windows with divider-styled titles,
# presents them sorted A-Z, and focuses the selected one.

windows=$(hyprctl clients -j | jq -r '
  .[] | select(.class == "com.mitchellh.ghostty" and (.title | test("^────")))
  | "\(.address)\t\(.title)"
' | sort -t$'\t' -k2)

if [[ -z "$windows" ]]; then
  notify-send -a "Claude Sessions" "No Claude Code sessions found"
  exit 0
fi

selected=$(echo "$windows" | cut -f2 | wofi --dmenu -p "Claude Sessions" --width 600 --height 400)

if [[ -n "$selected" ]]; then
  addr=$(echo "$windows" | grep -F "$selected" | head -1 | cut -f1)
  hyprctl dispatch focuswindow "address:$addr"
fi
