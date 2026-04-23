#!/usr/bin/env bash
# shell-stack-boot.sh - apply persisted shell owner state at session startup.

set -euo pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$script_dir/lib/shell-stack.sh"

apply=true
mode=""

usage() {
  cat <<'EOF'
Usage: shell-stack-boot.sh [--dry-run] [--mode <mode>]

Reads $XDG_STATE_HOME/dotfiles/shell-stack/env and applies the persisted
Quickshell migration owner state. Used by Hyprland exec-once so login does not
restart replaced owners after bar/ticker/notification cutover.
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --dry-run)
      apply=false
      ;;
    --mode)
      mode="${2:-}"
      [[ -n "$mode" ]] || {
        usage >&2
        exit 2
      }
      shift
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      usage >&2
      exit 2
      ;;
  esac
  shift
done

shell_stack_load
mode="${mode:-${SHELL_STACK_MODE:-pilot}}"

if $apply; then
  exec "$script_dir/shell-stack-mode.sh" --apply "$mode"
fi

exec "$script_dir/shell-stack-mode.sh" "$mode"
