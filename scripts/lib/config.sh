#!/usr/bin/env bash
# config.sh — Shared atomic config operations
# Source this file: source "$(dirname "$0")/lib/config.sh"

source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/notify.sh"
source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/compositor.sh"

# Atomic write: writes content to file via mktemp+mv (no partial reads)
# Usage: config_atomic_write <file> <content>
config_atomic_write() {
  local file="$1" content="$2"
  local tmp
  tmp="$(mktemp "${file}.XXXXXX")"
  printf '%s' "$content" > "$tmp"
  mv -f "$tmp" "$file"
}

# Atomic sed replacement on a file
# Usage: config_sed_replace <file> <sed_expression>
config_sed_replace() {
  local file="$1" expr="$2"
  local tmp
  tmp="$(mktemp "${file}.XXXXXX")"
  sed -e "$expr" "$file" > "$tmp"
  mv -f "$tmp" "$file"
}

# Backup a config file before modification
# Usage: config_backup <file>
config_backup() {
  local file="$1"
  local backup_dir="$HOME/.dotfiles-backup/$(date +%Y%m%d)"
  mkdir -p "$backup_dir"
  cp -a "$file" "$backup_dir/$(basename "$file").$(date +%H%M%S)" 2>/dev/null
}

# Reload the service associated with a config directory
# Usage: config_reload_service <component_name>
# Returns: 0 on success, non-zero on failure
config_reload_service() {
  local component="$1" rc=0
  case "$component" in
    hyprland|hypr) compositor_reload;      rc=$? ;;
    swaync)        swaync-client --reload-config; rc=$? ;;
    eww)           eww reload 2>/dev/null; rc=$? ;;
    waybar)        pkill -SIGUSR2 waybar;  rc=$? ;;
    tmux)          tmux source-file ~/.tmux.conf 2>/dev/null; rc=$? ;;
    # ghostty and tattoy auto-reload via file watching
  esac
  if (( rc == 0 )); then
    hg_notify_low "Config" "Reloaded $component"
  else
    hg_notify_critical "Config" "Failed to reload $component (exit $rc)"
  fi
  return $rc
}
