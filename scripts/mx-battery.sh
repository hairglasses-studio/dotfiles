#!/usr/bin/env bash
# mx-battery.sh — Waybar custom module for MX Master 4 battery
set -euo pipefail

# Parse battery from solaar
battery=$(solaar show "MX Master 4" 2>/dev/null | strings | grep -oP 'Battery: \K\d+' | head -1 || true)

# Fallback: bluetoothctl
if [[ -z "$battery" ]]; then
    battery=$(bluetoothctl info ${BT_MX_MASTER:-D2:8E:C5:DE:9F:CC} 2>/dev/null | grep -oP 'Battery Percentage:.*\(\K\d+' || true)
fi

if [[ -n "$battery" ]]; then
    class="normal"
    (( battery <= 20 )) && class="critical"
    (( battery > 20 && battery <= 30 )) && class="warning"
    printf '{"text": " %s%%", "tooltip": "MX Master 4: %s%%", "class": "%s", "percentage": %s}\n' \
        "$battery" "$battery" "$class" "$battery"
else
    printf '{"text": "", "tooltip": "MX Master 4: disconnected", "class": "disconnected"}\n'
fi
