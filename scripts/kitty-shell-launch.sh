#!/usr/bin/env bash
set -euo pipefail

SCRIPT_PATH="$(readlink -f "${BASH_SOURCE[0]:-$0}")"
SCRIPT_DIR="$(cd "$(dirname "$SCRIPT_PATH")" && pwd)"
LAUNCHER="${KITTY_SHELL_LAUNCHER:-$SCRIPT_DIR/kitty-visual-launch.sh}"

# Force a normal top-level Kitty OS window for the default terminal entrypoint
# so launcher invocations do not collapse back into tmux-backed sessions or a
# future startup session restore policy. Keep --session=none scoped here: Kitty
# ignores program arguments when --session is set, so the shared visual
# launcher cannot safely apply it for -e entrypoints.
exec "$LAUNCHER" --session=none "$@"
