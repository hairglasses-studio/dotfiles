#!/usr/bin/env bash
set -euo pipefail
# en16-session-focus.sh — Focus the Nth Claude Code agent session window.
# Called by mapitall EN16 profile (Grid.toml) when an encoder button is pushed.
# Uses the same Ghostty window discovery as agent-session-picker.sh.
#
# Usage: en16-session-focus.sh <session_number> [page_offset]
#   session_number: 1-12 (encoder position)
#   page_offset:    optional page multiplier (default 0), for >12 sessions

N="${1:-1}"
PAGE="${2:-0}"
INDEX=$(( PAGE * 12 + N - 1 ))

addr=$(hyprctl clients -j | jq -r '
  [.[] | select(.class == "kitty" and (.title | test("^────")))]
  | sort_by(.title) | .['"$INDEX"'].address // empty
')

if [[ -n "$addr" ]]; then
  hyprctl dispatch focuswindow "address:$addr"
else
  notify-send -a "EN16" "No session #$((INDEX + 1))" -t 1500
fi
