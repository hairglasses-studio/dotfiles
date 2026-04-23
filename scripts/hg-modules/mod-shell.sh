#!/usr/bin/env bash
# mod-shell.sh — staged shell migration controls

shell_description() {
  echo "Shell migration — Quickshell pilot, cutover, rollback"
}

shell_commands() {
  cat <<'CMDS'
status	Show current shell service state
pilot	Start Quickshell while keeping ironbar, ticker, hyprshell, swaync, and companions live
bar-cutover	Start Quickshell and stop ironbar
ticker-cutover	Start Quickshell and stop keybind ticker + ticker watcher services
menu-cutover	Start Quickshell menus and stop hyprshell launcher/switcher
dock-cutover	Start Quickshell dock and stop hypr-dock
companion-cutover	Start Quickshell companion overlays and stop standalone companion services
notification-cutover	Start Quickshell notification owner/history and stop swaync + history listener
full-pilot	Start Quickshell and stop ironbar + keybind ticker + hyprshell + dock + companion overlays
full-cutover	Start Quickshell as bar + ticker + menu + dock + notification/history + companion owner
rollback	Stop Quickshell and restore ironbar + keybind ticker + hyprshell + dock + swaync + companions
CMDS
}

_shell_stack_mode() {
  local script="$HG_DOTFILES/scripts/shell-stack-mode.sh"
  if [[ ! -x "$script" ]]; then
    script="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/shell-stack-mode.sh"
  fi
  "$script" "$@"
}

shell_run() {
  local cmd="${1:-}"
  shift || true

  case "$cmd" in
    status|pilot|bar-cutover|ticker-cutover|menu-cutover|dock-cutover|companion-cutover|notification-cutover|full-pilot|full-cutover|rollback)
      _shell_stack_mode "$cmd" "$@"
      ;;
    *) hg_die "Unknown shell command: $cmd. Run 'hg shell --help'." ;;
  esac
}
