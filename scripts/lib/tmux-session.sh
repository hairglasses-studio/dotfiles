#!/usr/bin/env bash
# tmux-session.sh — Shared helpers for persistent tmux-backed terminal entrypoints

tmux_session_exists() {
  local session_name="${1:?session name required}"
  tmux has-session -t "$session_name" 2>/dev/null
}

tmux_attach_or_create_shell_session() {
  local session_name="${1:?session name required}"
  local session_cwd="${2:?session cwd required}"

  if tmux_session_exists "$session_name"; then
    exec tmux attach-session -t "$session_name"
  fi

  exec tmux new-session -s "$session_name" -c "$session_cwd"
}

tmux_bootstrap_dropdown_session() {
  local session_name="${1:?session name required}"
  local studio_root="${2:?studio root required}"
  local dotfiles_root="${3:?dotfiles root required}"
  local ralph_cmd claude_cmd

  if tmux_session_exists "$session_name"; then
    return 0
  fi

  printf -v ralph_cmd 'HG_AGENT_SESSION_QUIET=1 ralphglasses --scan-path %q' "$studio_root"
  claude_cmd="claude"

  tmux new-session -d -s "$session_name" -c "$studio_root" "$ralph_cmd"
  tmux split-window -t "$session_name" -h -c "$dotfiles_root" "$claude_cmd"
  tmux select-pane -t "${session_name}:0.0"
}
