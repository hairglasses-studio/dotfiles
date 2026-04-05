#!/bin/bash
# hypr-vertical-columns.sh — Force all tiled windows on the focused
# workspace into equal-width, full-height vertical columns (dwindle).
#
# Dwindle creates a right-leaning binary tree. Default ratio = 1.0 gives
# 50/50 at each level → exponentially shrinking columns. To equalize:
#   Window i (0-indexed, left→right) needs parent ratio = 2/(N-i)
#   Delta from default 1.0 = 2/(N-i) - 1.0
set -euo pipefail

ws=$(hyprctl activeworkspace -j | jq -r '.id')
active=$(hyprctl activewindow -j | jq -r '.address')

# Get all tiled windows, active window LAST so it stays as anchor
others=()
while IFS= read -r addr; do
  [[ "$addr" == "$active" ]] && continue
  others+=("$addr")
done < <(hyprctl clients -j | jq -r --argjson ws "$ws" \
  '[.[] | select(.workspace.id == $ws and .floating == false)]
   | sort_by(.at[0], .at[1])
   | .[].address')

count=$(( ${#others[@]} + 1 ))
[[ $count -lt 2 ]] && exit 0

# Phase 1: stash all windows EXCEPT the active one (keeps workspace alive)
for addr in "${others[@]}"; do
  hyprctl dispatch movetoworkspacesilent "special:_vcol,address:$addr"
done
sleep 0.1

# Phase 2: bring others back one-by-one with preselect right
# Active window is already on the workspace as leftmost anchor
for addr in "${others[@]}"; do
  hyprctl dispatch layoutmsg "preselect r"
  sleep 0.1
  hyprctl dispatch movetoworkspacesilent "$ws,address:$addr"
  sleep 0.05
  hyprctl dispatch focuswindow "address:$addr"
  sleep 0.1
done
sleep 0.2

# Phase 3: equalize columns via splitratio deltas
# Re-read positions after retiling (left→right order)
eq_addrs=$(hyprctl clients -j | jq -r --argjson ws "$ws" \
  '[.[] | select(.workspace.id == $ws and .floating == false)]
   | sort_by(.at[0])
   | .[].address')
eq_count=$(echo "$eq_addrs" | grep -c .)

i=0
for addr in $eq_addrs; do
  remaining=$((eq_count - i))
  if [[ $remaining -ge 3 ]]; then
    delta=$(awk "BEGIN {printf \"%.4f\", 2.0/$remaining - 1.0}")
    hyprctl dispatch focuswindow "address:$addr"
    hyprctl dispatch layoutmsg "splitratio $delta"
  fi
  ((i++))
done
