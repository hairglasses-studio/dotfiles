#!/usr/bin/env bash
set -euo pipefail

SCRIPT_PATH="$(readlink -f "${BASH_SOURCE[0]:-$0}")"
SCRIPT_DIR="$(cd "$(dirname "$SCRIPT_PATH")" && pwd)"
LAUNCHER="${KITTY_DEV_LAUNCHER:-$SCRIPT_DIR/kitty-visual-launch.sh}"

if ! command -v tmux >/dev/null 2>&1; then
  exec "$LAUNCHER" "$@"
fi

exec "$LAUNCHER" "$@" -e "$SCRIPT_DIR/tmux-main-session.sh"
