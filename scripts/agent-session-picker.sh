#!/usr/bin/env bash
set -euo pipefail
# agent-session-picker.sh — Focus an active agent session by name via wofi.
# Queries Hyprland IPC for Kitty windows with divider-styled titles,
# presents them sorted A-Z, and focuses the selected one.

windows=$(hyprctl clients -j | jq -r '
  .[] | select(.class == "kitty" and (.title | test("^────")))
  | "\(.address)\t\(.title)"
' | sort -t$'\t' -k2)

if [[ -z "$windows" ]]; then
  notify-send -a "Agent Sessions" "No agent sessions found"
  exit 0
fi

selected=$(echo "$windows" | cut -f2 | wofi --dmenu -p "Agent Sessions" --width 600 --height 400)

if [[ -n "$selected" ]]; then
  addr=$(echo "$windows" | grep -F "$selected" | head -1 | cut -f1)
  hyprctl dispatch focuswindow "address:$addr"
fi
