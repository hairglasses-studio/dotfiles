#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PLAYLIST="${KITTY_VISUAL_PLAYLIST:-ambient}"

exec "$SCRIPT_DIR/kitty-shader-playlist.sh" spawn "$PLAYLIST" -- "$@"
