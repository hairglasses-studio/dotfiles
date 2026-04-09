#!/usr/bin/env bash
set -euo pipefail

SCRIPT_PATH="$(readlink -f "${BASH_SOURCE[0]:-$0}")"
SCRIPT_DIR="$(cd "$(dirname "$SCRIPT_PATH")" && pwd)"
source "$SCRIPT_DIR/lib/hg-core.sh"
source "$SCRIPT_DIR/lib/tmux-session.sh"

SESSION_NAME="${HG_DROPDOWN_TMUX_SESSION:-dropdown}"
STUDIO_ROOT="${HG_DROPDOWN_STUDIO_ROOT:-$HG_STUDIO_ROOT}"
DOTFILES_ROOT="${HG_DROPDOWN_DOTFILES_ROOT:-$HG_DOTFILES}"

if ! command -v tmux >/dev/null 2>&1; then
  cd "$STUDIO_ROOT"
  exec env HG_AGENT_SESSION_QUIET=1 ralphglasses --scan-path "$STUDIO_ROOT"
fi

tmux_bootstrap_dropdown_session "$SESSION_NAME" "$STUDIO_ROOT" "$DOTFILES_ROOT"
exec tmux attach-session -t "$SESSION_NAME"
