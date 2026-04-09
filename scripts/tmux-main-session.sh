#!/usr/bin/env bash
set -euo pipefail

SCRIPT_PATH="$(readlink -f "${BASH_SOURCE[0]:-$0}")"
SCRIPT_DIR="$(cd "$(dirname "$SCRIPT_PATH")" && pwd)"
source "$SCRIPT_DIR/lib/hg-core.sh"
source "$SCRIPT_DIR/lib/tmux-session.sh"

SESSION_NAME="${HG_DEV_TMUX_SESSION:-main}"
SESSION_CWD="${HG_DEV_TMUX_CWD:-$HG_STUDIO_ROOT}"

if ! command -v tmux >/dev/null 2>&1; then
  cd "$SESSION_CWD"
  exec "${SHELL:-/bin/bash}" -l
fi

tmux_attach_or_create_shell_session "$SESSION_NAME" "$SESSION_CWD"
