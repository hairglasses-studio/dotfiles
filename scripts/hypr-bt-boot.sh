#!/usr/bin/env bash
# hypr-bt-boot.sh — Boot-time bluetooth device connection with notifications
# Called from hyprland exec-once. Waits for BT adapter, then connects devices.
set -euo pipefail

source "$(cd "$(dirname "$0")/lib" && pwd)/notify.sh"

# Devices to connect at boot (MAC -> friendly name)
declare -A DEVICES=(
  [${BT_HEADPHONES:-AC:BF:71:C8:DB:95}]="Headphones"
  [${BT_MX_MASTER:-D2:8E:C5:DE:9F:C8}]="MX Master 4"
)

# Wait for BT adapter
sleep 5

connected=()
failed=()

for mac in "${!DEVICES[@]}"; do
  name="${DEVICES[$mac]}"
  if bluetoothctl connect "$mac" &>/dev/null; then
    connected+=("$name")
  else
    failed+=("$name")
  fi
done

if (( ${#connected[@]} > 0 )); then
  hg_notify_low "Bluetooth" "Connected: ${connected[*]}"
fi
if (( ${#failed[@]} > 0 )); then
  hg_notify "Bluetooth" "Failed: ${failed[*]}"
fi
