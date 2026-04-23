#!/usr/bin/env bash
# shell-stack.sh - read staged Quickshell migration state.

shell_stack_state_dir() {
  printf '%s\n' "${XDG_STATE_HOME:-$HOME/.local/state}/dotfiles/shell-stack"
}

shell_stack_env_file() {
  printf '%s/env\n' "$(shell_stack_state_dir)"
}

shell_stack_load() {
  SHELL_STACK_MODE="${SHELL_STACK_MODE:-pilot}"
  QS_BAR_CUTOVER="${QS_BAR_CUTOVER:-0}"
  QS_TICKER_CUTOVER="${QS_TICKER_CUTOVER:-0}"
  QUICKSHELL_NOTIFICATION_OWNER="${QUICKSHELL_NOTIFICATION_OWNER:-0}"

  local env_file
  env_file="$(shell_stack_env_file)"
  if [[ -r "$env_file" ]]; then
    # shellcheck disable=SC1090
    source "$env_file"
  fi
}

shell_stack_bar_cutover() {
  [[ "${QS_BAR_CUTOVER:-0}" == "1" ]]
}

shell_stack_ticker_cutover() {
  [[ "${QS_TICKER_CUTOVER:-0}" == "1" ]]
}

shell_stack_notification_cutover() {
  [[ "${QUICKSHELL_NOTIFICATION_OWNER:-0}" == "1" ]]
}

shell_stack_quickshell_wanted() {
  [[ "${SHELL_STACK_MODE:-pilot}" != "rollback" ]]
}
