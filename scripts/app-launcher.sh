#!/usr/bin/env bash
set -euo pipefail

# app-launcher.sh — fallback-aware launcher entrypoint for Hyprland
#
# If hyprshell is installed and running, it owns Super+D itself and this script
# should stay inert to avoid double-launching surfaces. Until that stack is
# installed, fall back to whichever launcher is available locally.

SCRIPT_PATH="$(readlink -f "${BASH_SOURCE[0]:-$0}")"
SCRIPT_DIR="$(cd "$(dirname "$SCRIPT_PATH")" && pwd)"
source "$SCRIPT_DIR/lib/launcher.sh"

if launcher_hyprshell_running; then
  exit 0
fi

if command -v wofi >/dev/null 2>&1; then
  IFS=$'\t' read -r width height monitor < <(launcher_wofi_geometry)
  args=(--show drun --width "$width" --height "$height")
  if [[ -n "$monitor" ]]; then
    args+=(--monitor "$monitor")
  fi
  exec wofi "${args[@]}"
fi

if command -v rofi >/dev/null 2>&1; then
  exec rofi -show drun
fi

printf 'app-launcher: no supported launcher found (expected wofi or rofi)\n' >&2
exit 1
