#!/usr/bin/env bash
# mx-battery-notify.sh — Desktop notification if MX Master 4 battery is low
set -euo pipefail

THRESHOLD=20

# Read MX Master battery via bluetoothctl
battery=""
while IFS= read -r line; do
    if [[ "$line" =~ Percentage:\ 0x([0-9a-fA-F]+) ]]; then
        battery=$(( 16#${BASH_REMATCH[1]} ))
    fi
done < <(bluetoothctl info 2>/dev/null | grep -i "battery\|percentage" || true)
[[ -z "$battery" ]] && exit 0

if (( battery <= THRESHOLD )); then
    notify-send -u critical "MX Master 4" "Battery low: ${battery}%"
fi
