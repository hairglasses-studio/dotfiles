#!/usr/bin/env bash
# ironbar-keybind-ticker.sh — Scrolling stock-ticker for keybinds
# Outputs a sliding window of the keybind string, one line per frame.
# Designed for ironbar script module in watch mode.
set -euo pipefail

# Decode modmask to human-readable (matches hypr-keybinds.sh format_mods)
format_mods() {
  local m=$1 out=""
  (( m & 64 )) && out+="Super+"
  (( m & 1 ))  && out+="Shift+"
  (( m & 4 ))  && out+="Ctrl+"
  (( m & 8 ))  && out+="Alt+"
  printf '%s' "$out"
}

build_ticker() {
  hyprctl binds -j 2>/dev/null | jq -r '
    [.[] | select(.has_description == true and .submap == "" and .mouse == false)
     | "\(.description)  \(.modmask):\(.key)"]
    | .[]
  ' | while IFS= read -r line; do
    desc="${line%%  *}"
    raw="${line##*  }"
    mask="${raw%%:*}"
    key="${raw##*:}"
    mods=$(format_mods "$mask")
    printf '  %s  %s%s  ' "$desc" "$mods" "$key"
  done
}

# Build the full ticker string once at startup
ticker=$(build_ticker)
# Append a separator and duplicate for seamless wrapping
sep="                    "
ticker="${ticker}${sep}"
len=${#ticker}

# Sliding window: ~350 chars fills the ultrawide at 11px mono
window=350
offset=0

# Double the string so substring extraction never goes out of bounds
doubled="${ticker}${ticker}"

while true; do
  printf '%s\n' "${doubled:$offset:$window}"
  offset=$(( (offset + 1) % len ))
  sleep 0.06
done
