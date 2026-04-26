#!/usr/bin/env bash
# shell-stack.sh - read shell-stack mode state.

shell_stack_state_dir() {
  printf '%s\n' "${XDG_STATE_HOME:-$HOME/.local/state}/dotfiles/shell-stack"
}

shell_stack_env_file() {
  printf '%s/env\n' "$(shell_stack_state_dir)"
}

shell_stack_load() {
  SHELL_STACK_MODE="${SHELL_STACK_MODE:-full-cutover}"

  local env_file
  env_file="$(shell_stack_env_file)"
  if [[ -r "$env_file" ]]; then
    # shellcheck disable=SC1090
    source "$env_file"
  fi
}

shell_stack_quickshell_wanted() {
  [[ "${SHELL_STACK_MODE:-full-cutover}" != "rollback" ]]
}
