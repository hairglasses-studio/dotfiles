#!/usr/bin/env bash
# controller-screenshot.sh — Gamepad screenshot (handles pipe for AntiMicroX Execute)
grim - | wl-copy && notify-send -a "Gamepad" "Screenshot" "Copied to clipboard" 2>/dev/null
