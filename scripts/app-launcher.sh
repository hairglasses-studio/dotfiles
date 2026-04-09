#!/usr/bin/env bash
set -euo pipefail

# app-launcher.sh — fallback-aware launcher entrypoint for Hyprland
#
# If hyprshell is installed and running, it owns Super+D itself and this script
# should stay inert to avoid double-launching surfaces. Until that stack is
# installed, fall back to whichever launcher is available locally.

if command -v hyprshell >/dev/null 2>&1 && pgrep -x hyprshell >/dev/null 2>&1; then
  exit 0
fi

if command -v wofi >/dev/null 2>&1; then
  exec wofi --show drun
fi

if command -v rofi >/dev/null 2>&1; then
  exec rofi -show drun
fi

printf 'app-launcher: no supported launcher found (expected wofi or rofi)\n' >&2
exit 1
