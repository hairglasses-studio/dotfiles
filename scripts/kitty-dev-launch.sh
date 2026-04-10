#!/usr/bin/env bash
set -euo pipefail

SCRIPT_PATH="$(readlink -f "${BASH_SOURCE[0]:-$0}")"
SCRIPT_DIR="$(cd "$(dirname "$SCRIPT_PATH")" && pwd)"
LAUNCHER="${KITTY_DEV_LAUNCHER:-$SCRIPT_DIR/kitty-visual-launch.sh}"

# Explicit tmux entrypoint for persistent dev sessions. Keep the default
# launcher on kitty-shell-launch.sh so fresh terminal windows stay shell-first.
if ! command -v tmux >/dev/null 2>&1; then
  exec "$LAUNCHER" "$@"
fi

exec "$LAUNCHER" "$@" -e "$SCRIPT_DIR/tmux-main-session.sh"
