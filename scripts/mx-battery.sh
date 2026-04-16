#!/usr/bin/env bash
# mx-battery.sh — Waybar custom module for MX Master 4 battery
set -euo pipefail

# Read MX Master battery via bluetoothctl
battery=""
charging="false"
while IFS= read -r line; do
    if [[ "$line" =~ Percentage:\ 0x([0-9a-fA-F]+) ]]; then
        battery=$(( 16#${BASH_REMATCH[1]} ))
    fi
done < <(bluetoothctl info 2>/dev/null | grep -i "battery\|percentage" || true)

if [[ -n "$battery" ]]; then
    class="normal"
    (( battery <= 20 )) && class="critical"
    (( battery > 20 && battery <= 30 )) && class="warning"
    tooltip="MX Master 4: ${battery}%"
    [[ "$charging" == "true" ]] && tooltip+=" (charging)"
    printf '{"text":" %s%%","tooltip":"%s","class":"%s","percentage":%s}\n' \
        "$battery" "$tooltip" "$class" "$battery"
else
    printf '{"text": "", "tooltip": "MX Master 4: disconnected", "class": "disconnected"}\n'
fi
