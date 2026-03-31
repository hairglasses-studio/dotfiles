#!/usr/bin/env bash
# screenshot-crop.sh — Crop-select screenshot, save locally, copy path to clipboard
# Usage: Bind to a keybind (e.g., Super+Shift+Print)
# Saves to ~/Pictures/screenshots/ and copies filepath to clipboard for sharing

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
source "$SCRIPT_DIR/lib/compositor.sh" 2>/dev/null

SCREENSHOT_DIR="$HOME/Pictures/screenshots"
mkdir -p "$SCREENSHOT_DIR"

FILENAME="$(date +%Y%m%d_%H%M%S).png"
FILEPATH="$SCREENSHOT_DIR/$FILENAME"

# Crop-select region with slurp, capture with grim
REGION=$(slurp 2>/dev/null)
[[ -z "$REGION" ]] && exit 1

grim -g "$REGION" "$FILEPATH" 2>/dev/null || exit 1

# Copy filepath to clipboard for sharing
echo -n "$FILEPATH" | wl-copy 2>/dev/null

# Notification
notify-send -a "Screenshot" -i "$FILEPATH" "Screenshot saved" "$FILEPATH" 2>/dev/null

echo "$FILEPATH"
