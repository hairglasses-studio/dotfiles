#!/usr/bin/env bash
# mod-pomo.sh — hg pomo module
# Thin wrapper around scripts/hg-pomo for the pomodoro ticker stream.

_POMO="$HG_DOTFILES/scripts/hg-pomo"

pomo_description() {
  echo "Pomodoro timer for the ticker"
}

pomo_commands() {
  cat <<'CMDS'
start	Start a work session [minutes, default 25]
break	Start a break [minutes, default 5]
pause	Pause current session
resume	Resume paused session
stop	Clear pomodoro state
status	Show current state as JSON
CMDS
}

pomo_run() {
  local cmd="${1:-status}"
  shift || true
  exec bash "$_POMO" "$cmd" "$@"
}
