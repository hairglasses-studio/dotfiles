#!/usr/bin/env bash
set -euo pipefail

# app-switcher.sh — fallback-aware app switcher for Hyprland
#
# hyprshell captures Alt+Tab itself when installed and running. Until that
# stack is present, keep the old wofi-based switcher available.

if command -v hyprshell >/dev/null 2>&1 && pgrep -x hyprshell >/dev/null 2>&1; then
  exit 0
fi

hyprctl clients -j | jq -r '.[] | "\(.address) \(.class) — \(.title)"' | \
  wofi --dmenu --prompt "Switch to" --insensitive | \
  awk '{print $1}' | \
  xargs -I{} hyprctl dispatch focuswindow address:{}
