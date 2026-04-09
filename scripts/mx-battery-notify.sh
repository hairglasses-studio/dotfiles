#!/usr/bin/env bash
# mx-battery-notify.sh — Desktop notification if MX Master 4 battery is low
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/lib/juhradial.sh"

THRESHOLD=20

status="$(juhradial_battery_status 2>/dev/null || true)"
[[ -z "$status" ]] && exit 0

read -r battery charging <<<"$status"
[[ -z "$battery" ]] && exit 0

if (( battery <= THRESHOLD )); then
    notify-send -u critical "MX Master 4" "Battery low: ${battery}%"
fi
