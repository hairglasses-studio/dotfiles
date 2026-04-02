#!/usr/bin/env bash
# controller-wallpaper-next.sh — Gamepad wallpaper cycle
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
"$SCRIPT_DIR/../hg" wallpaper next
