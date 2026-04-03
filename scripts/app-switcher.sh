#!/usr/bin/env bash
# app-switcher.sh — wofi-based window switcher for Hyprland
hyprctl clients -j | jq -r '.[] | "\(.address) \(.class) — \(.title)"' | \
  wofi --dmenu --prompt "Switch to" --insensitive | \
  awk '{print $1}' | \
  xargs -I{} hyprctl dispatch focuswindow address:{}
