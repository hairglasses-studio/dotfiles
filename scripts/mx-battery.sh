#!/usr/bin/env bash
# mx-battery.sh — Waybar custom module for MX Master 4 battery
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/lib/juhradial.sh"

status="$(juhradial_battery_status 2>/dev/null || true)"
battery=""
charging="false"
if [[ -n "$status" ]]; then
    read -r battery charging <<<"$status"
fi

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
