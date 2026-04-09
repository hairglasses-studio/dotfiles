#!/usr/bin/env bash
set -euo pipefail

SCRIPT_PATH="$(readlink -f "${BASH_SOURCE[0]:-$0}")"
SCRIPT_DIR="$(cd "$(dirname "$SCRIPT_PATH")" && pwd)"
PLAYLIST="${KITTY_VISUAL_PLAYLIST:-ambient}"

exec "$SCRIPT_DIR/kitty-shader-playlist.sh" spawn "$PLAYLIST" -- "$@"
