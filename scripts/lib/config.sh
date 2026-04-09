#!/usr/bin/env bash
# config.sh — Shared atomic config operations
# Source this file: source "$(dirname "$0")/lib/config.sh"

source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/notify.sh"
source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/compositor.sh"

config_action_lane() {
  local verb="${1:-}" component="${2:-}"

  case "${verb}:${component}" in
    reload:hyprland|reload:hypr|reload:hyprshell|reload:swaync|reload:eww|reload:ironbar|reload:tmux)
      printf 'safe_reload\n'
      ;;
    reload:hypr-dock|reload:hyprdock|reload:hyprdynamicmonitors|reload:monitors|reload:hyprland-autoname-workspaces|reload:autoname)
      printf 'service_reload\n'
      ;;
    restart:hyprshell|restart:hypr-dock|restart:hyprdock|restart:hyprdynamicmonitors|restart:monitors|restart:hyprland-autoname-workspaces|restart:autoname)
      printf 'explicit_restart\n'
      ;;
    *)
      return 1
      ;;
  esac
}

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

# Trigger Hyprshell's built-in watched-file reload without restarting the process.
config_reload_hyprshell() {
  local config_path="${HYPRSHELL_CONFIG_PATH:-$HOME/.config/hyprshell/config.toml}"
  local style_path="${HYPRSHELL_STYLE_PATH:-$HOME/.config/hyprshell/styles.css}"
  local touched=0 path

  pgrep -x hyprshell >/dev/null 2>&1 || return 1

  for path in "$config_path" "$style_path"; do
    [[ -e "$path" ]] || continue
    touch "$path"
    touched=1
  done

  (( touched > 0 ))
}

# Reload the service associated with a config directory
# Usage: config_reload_service <component_name> [--quiet]
# Returns: 0 on success, non-zero on failure
config_reload_service() {
  local component="$1"
  shift || true
  local quiet=false rc=0
  [[ "${1:-}" == "--quiet" ]] && quiet=true
  case "$component" in
    hyprland|hypr) compositor_reload;      rc=$? ;;
    hyprshell)     config_reload_hyprshell; rc=$? ;;
    hypr-dock|hyprdock) systemctl --user restart dotfiles-hypr-dock.service; rc=$? ;;
    hyprdynamicmonitors|monitors)
      hypr_ensure_runtime_state
      systemctl --user restart dotfiles-hyprdynamicmonitors.service
      rc=$?
      ;;
    hyprland-autoname-workspaces|autoname) systemctl --user restart dotfiles-hyprland-autoname-workspaces.service; rc=$? ;;
    ironbar)       ironbar reload 2>/dev/null; rc=$? ;;
    swaync)        swaync-client --reload-config; rc=$? ;;
    eww)           eww reload 2>/dev/null; rc=$? ;;
    tmux)          tmux source-file ~/.tmux.conf 2>/dev/null; rc=$? ;;
    # ghostty and tattoy auto-reload via file watching
  esac
  if ! $quiet; then
    if (( rc == 0 )); then
      hg_notify_low "Config" "Reloaded $component"
    else
      hg_notify_critical "Config" "Failed to reload $component (exit $rc)"
    fi
  fi
  return $rc
}

# Restart a service-backed component explicitly.
# Usage: config_restart_service <component_name> [--quiet]
config_restart_service() {
  local component="$1"
  shift || true
  local quiet=false rc=0
  [[ "${1:-}" == "--quiet" ]] && quiet=true
  case "$component" in
    hyprshell)     systemctl --user restart dotfiles-hyprshell.service; rc=$? ;;
    hypr-dock|hyprdock) systemctl --user restart dotfiles-hypr-dock.service; rc=$? ;;
    hyprdynamicmonitors|monitors)
      hypr_ensure_runtime_state
      systemctl --user restart dotfiles-hyprdynamicmonitors.service
      rc=$?
      ;;
    hyprland-autoname-workspaces|autoname) systemctl --user restart dotfiles-hyprland-autoname-workspaces.service; rc=$? ;;
    *)
      rc=1
      ;;
  esac
  if ! $quiet; then
    if (( rc == 0 )); then
      hg_notify_low "Config" "Restarted $component"
    else
      hg_notify_critical "Config" "Failed to restart $component (exit $rc)"
    fi
  fi
  return $rc
}

# Reload multiple services in parallel
# Usage: config_reload_parallel hyprland swaync eww
config_reload_parallel() {
  local pids=() component
  for component in "$@"; do
    config_reload_service "$component" 2>/dev/null &
    pids+=($!)
  done
  local rc=0
  for pid in "${pids[@]}"; do
    wait "$pid" || rc=1
  done
  return $rc
}

# Restart multiple services in parallel.
# Usage: config_restart_parallel hyprshell hypr-dock autoname
config_restart_parallel() {
  local pids=() component
  for component in "$@"; do
    config_restart_service "$component" --quiet 2>/dev/null &
    pids+=($!)
  done
  local rc=0
  for pid in "${pids[@]}"; do
    wait "$pid" || rc=1
  done
  return $rc
}
