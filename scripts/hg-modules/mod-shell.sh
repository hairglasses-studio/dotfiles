#!/usr/bin/env bash
# mod-shell.sh — staged shell migration controls

shell_description() {
  echo "Shell migration — Quickshell pilot, cutover, rollback"
}

shell_commands() {
  cat <<'CMDS'
status	Show current shell service state
pilot	Start Quickshell while keeping ironbar, ticker, and swaync live
bar-cutover	Start Quickshell and stop ironbar
ticker-cutover	Start Quickshell and stop keybind ticker
full-pilot	Start Quickshell and stop ironbar + keybind ticker
rollback	Stop Quickshell and restore ironbar + keybind ticker + swaync
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
    status|pilot|bar-cutover|ticker-cutover|full-pilot|rollback)
      _shell_stack_mode "$cmd" "$@"
      ;;
    *) hg_die "Unknown shell command: $cmd. Run 'hg shell --help'." ;;
  esac
}
