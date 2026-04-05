#!/usr/bin/env bash
# controller-profile-switch.sh — Toggle between AntiMicroX profiles
# Usage: controller-profile-switch.sh <profile-name>
set -euo pipefail
PROFILE_DIR="$HOME/.config/antimicrox"
name="${1:-hyprland-desktop}"
profile="$PROFILE_DIR/${name}.gamecontroller.amgp"
[[ -f "$profile" ]] || exit 1
pkill antimicrox 2>/dev/null
sleep 0.3
antimicrox --tray --hidden --profile "$profile" &
disown
notify-send -a "Gamepad" "Profile: $name" 2>/dev/null
