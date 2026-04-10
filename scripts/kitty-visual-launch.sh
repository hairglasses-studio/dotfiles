#!/usr/bin/env bash
set -euo pipefail

SCRIPT_PATH="$(readlink -f "${BASH_SOURCE[0]:-$0}")"
SCRIPT_DIR="$(cd "$(dirname "$SCRIPT_PATH")" && pwd)"
PLAYLIST="${KITTY_VISUAL_PLAYLIST:-ambient}"
SPAWNER="${KITTY_VISUAL_SPAWNER:-$SCRIPT_DIR/kitty-shader-playlist.sh}"

# Force a fresh top-level Kitty OS window for every visual launch surface so
# Hyprland binds, scratchpads, bars, and controller actions do not collapse
# into an existing instance. Do not set --session=none here: Kitty ignores
# program arguments when --session is supplied, which would break -e launchers.
exec "$SPAWNER" spawn "$PLAYLIST" -- --single-instance=no --start-as=normal "$@"
