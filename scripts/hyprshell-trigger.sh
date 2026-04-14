#!/usr/bin/env bash
# hyprshell-trigger.sh — invoke hyprshell launcher surfaces from mouse actions
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$(readlink -f "${BASH_SOURCE[0]}")")" && pwd)"

action="${1:-}"

case "$action" in
  overview)
    export DOTFILES_LAUNCHER_PREFER_HYPRSHELL=1
    exec "$SCRIPT_DIR/app-launcher.sh"
    ;;
  switch)
    export DOTFILES_LAUNCHER_PREFER_HYPRSHELL=1
    exec "$SCRIPT_DIR/app-switcher.sh"
    ;;
  *)
    printf 'Usage: %s {overview|switch}\n' "$(basename "$0")" >&2
    exit 2
    ;;
esac
