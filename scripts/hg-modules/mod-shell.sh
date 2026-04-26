#!/usr/bin/env bash
# mod-shell.sh — Quickshell stack controls.

shell_description() {
  echo "Shell stack — Quickshell ownership status, full-cutover, rollback"
}

shell_commands() {
  cat <<'CMDS'
status	Show current shell service state
full-cutover	Start Quickshell as the sole owner of every desktop surface
rollback	Stop Quickshell (escape hatch; re-enable legacy units manually if needed)
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
    status|full-cutover|rollback)
      _shell_stack_mode "$cmd" "$@"
      ;;
    *) hg_die "Unknown shell command: $cmd. Run 'hg shell --help'." ;;
  esac
}
