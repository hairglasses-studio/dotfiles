#!/usr/bin/env bash
# mx-battery-notify.sh — Desktop notification if MX Master 4 battery is low
set -euo pipefail

THRESHOLD=20

battery=$(solaar show "MX Master 4" 2>/dev/null | strings | grep -oP 'Battery: \K\d+' | head -1 || true)
[[ -z "$battery" ]] && battery=$(bluetoothctl info ${BT_MX_MASTER:-D2:8E:C5:DE:9F:C8} 2>/dev/null | grep -oP 'Battery Percentage:.*\(\K\d+' || true)
[[ -z "$battery" ]] && exit 0

if (( battery <= THRESHOLD )); then
    notify-send -u critical "MX Master 4" "Battery low: ${battery}%"
fi
