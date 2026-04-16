#!/usr/bin/env bash
# ironbar-keybind-ticker.sh — Scrolling stock-ticker for keybinds
# Outputs a sliding window of keybind text at ~50fps for smooth scrolling.
# Colors are handled by ironbar CSS, not Pango markup.
set -euo pipefail

format_mods() {
  local m=$1 out=""
  (( m & 64 )) && out+="Super+"
  (( m & 1 ))  && out+="Shift+"
  (( m & 4 ))  && out+="Ctrl+"
  (( m & 8 ))  && out+="Alt+"
  printf '%s' "$out"
}

# Build the full ticker string from live keybinds
ticker=""
while IFS=$'\t' read -r desc mask key; do
  mods=$(format_mods "$mask")
  ticker+="  ${desc}  ${mods}${key}  ·"
done < <(
  hyprctl binds -j 2>/dev/null | jq -r '
    [.[] | select(.has_description == true and .submap == "" and .mouse == false)
     | "\(.description)\t\(.modmask)\t\(.key)"] | .[]'
)

[[ -z "$ticker" ]] && { echo "No keybinds"; exit 1; }

# Double for seamless wrap
ticker+="$ticker"
half=$(( ${#ticker} / 2 ))
offset=0

# Render at ~233fps (just under DP-3's 239.76Hz) but advance text slowly.
# Scroll 1 char every 12 frames ≈ 19 chars/sec — readable stock-ticker pace.
frame=0
advance=12
while true; do
  printf '%s\n' "${ticker:$offset:350}"
  (( ++frame % advance == 0 )) && offset=$(( (offset + 1) % half ))
  sleep 0.0043
done
