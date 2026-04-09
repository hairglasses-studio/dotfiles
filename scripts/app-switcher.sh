#!/usr/bin/env bash
set -euo pipefail

# app-switcher.sh — fallback-aware app switcher for Hyprland
#
# hyprshell captures Alt+Tab itself when installed and running. Until that
# stack is present, keep the old wofi-based switcher available.

SCRIPT_PATH="$(readlink -f "${BASH_SOURCE[0]:-$0}")"
SCRIPT_DIR="$(cd "$(dirname "$SCRIPT_PATH")" && pwd)"
source "$SCRIPT_DIR/lib/launcher.sh"

if launcher_hyprshell_running; then
  exit 0
fi

if ! command -v hyprctl >/dev/null 2>&1 || ! command -v jq >/dev/null 2>&1; then
  printf 'app-switcher: hyprctl and jq are required\n' >&2
  exit 1
fi

clients="$(
  hyprctl clients -j | jq -r '
    .[] |
    select((.address // "") != "") |
    "\(.address) \(.class // "app") — \(.title // "")"
  '
)"

[[ -n "$clients" ]] || exit 0

selection=""
if command -v wofi >/dev/null 2>&1; then
  IFS=$'\t' read -r width height monitor < <(launcher_wofi_geometry)
  args=(--dmenu --prompt "Switch to" --insensitive --width "$width" --height "$height")
  if [[ -n "$monitor" ]]; then
    args+=(--monitor "$monitor")
  fi
  selection="$(printf '%s\n' "$clients" | wofi "${args[@]}" || true)"
elif command -v rofi >/dev/null 2>&1; then
  selection="$(printf '%s\n' "$clients" | rofi -dmenu -i -p "Switch to" || true)"
else
  printf 'app-switcher: no supported launcher found (expected wofi or rofi)\n' >&2
  exit 1
fi

address="${selection%% *}"
[[ -n "$address" ]] || exit 0
exec hyprctl dispatch focuswindow "address:${address}"
