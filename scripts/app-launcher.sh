#!/usr/bin/env bash
set -euo pipefail

# app-launcher.sh — fallback-aware launcher entrypoint for Hyprland
#
# Default to stable local launcher surfaces. Hyprshell overview is still
# available for explicit opt-in callers, but daily Mod+D should not depend on
# a native overview path that may silently no-op.

SCRIPT_PATH="$(readlink -f "${BASH_SOURCE[0]:-$0}")"
SCRIPT_DIR="$(cd "$(dirname "$SCRIPT_PATH")" && pwd)"
source "$SCRIPT_DIR/lib/launcher.sh"

if launcher_prefer_hyprshell; then
  if launcher_hyprshell_socat '"OpenOverview"' && launcher_wait_hyprshell_layer 'hyprshell_(overview|launcher)'; then
    exit 0
  fi
fi

if command -v wofi >/dev/null 2>&1; then
  IFS=$'\t' read -r width height monitor < <(launcher_wofi_geometry)
  args=(--show drun --width "$width" --height "$height")
  if [[ -n "$monitor" ]]; then
    args+=(--monitor "$monitor")
  fi
  wofi_dir="$(launcher_wofi_config_dir)"
  if [[ -d "$wofi_dir" ]]; then
    cd "$wofi_dir"
  fi
  exec wofi "${args[@]}"
fi

if command -v rofi >/dev/null 2>&1; then
  exec rofi -show drun
fi

printf 'app-launcher: no supported launcher found (expected wofi or rofi)\n' >&2
exit 1
