#!/usr/bin/env bash
set -euo pipefail

# app-switcher.sh — fallback-aware app switcher for Hyprland
#
# When hyprshell is installed and running, ask its daemon to open or close the
# switcher directly. Until that stack is present, keep the old wofi-based
# switcher available.

SCRIPT_PATH="$(readlink -f "${BASH_SOURCE[0]:-$0}")"
SCRIPT_DIR="$(cd "$(dirname "$SCRIPT_PATH")" && pwd)"
source "$SCRIPT_DIR/lib/launcher.sh"

action="${1:-open}"
case "$action" in
  open|"")
    hyprshell_payload='{"OpenSwitch":{"reverse":false}}'
    ;;
  reverse)
    hyprshell_payload='{"OpenSwitch":{"reverse":true}}'
    ;;
  close)
    hyprshell_payload='"CloseSwitch"'
    ;;
  *)
    printf 'Usage: %s {open|reverse|close}\n' "$(basename "$0")" >&2
    exit 2
    ;;
esac

if launcher_hyprshell_socat "$hyprshell_payload"; then
  case "$action" in
    open|reverse)
      if launcher_wait_hyprshell_layer 'hyprshell_switch'; then
        exit 0
      fi
      ;;
    close)
      exit 0
      ;;
  esac
fi

if [[ "$action" == "close" ]]; then
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
  wofi_dir="$(launcher_wofi_config_dir)"
  if [[ -d "$wofi_dir" ]]; then
    cd "$wofi_dir"
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
