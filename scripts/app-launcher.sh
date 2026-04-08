#!/usr/bin/env bash
set -euo pipefail

# app-launcher.sh — fallback-aware launcher entrypoint for Hyprland
#
# If hyprshell is installed and running, it owns Super+D itself and this script
# should stay inert to avoid double-launching surfaces. Until that stack is
# installed, fall back to wofi.

if command -v hyprshell >/dev/null 2>&1 && pgrep -x hyprshell >/dev/null 2>&1; then
  exit 0
fi

exec wofi --show drun
