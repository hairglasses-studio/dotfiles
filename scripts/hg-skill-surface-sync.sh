#!/usr/bin/env bash
set -euo pipefail

STUDIO_ROOT="${HG_STUDIO_ROOT:-$HOME/hairglasses-studio}"
TARGET="$STUDIO_ROOT/surfacekit/scripts/skill-surface-sync.sh"

[[ -f "$TARGET" ]] || {
  echo "surfacekit skill sync entrypoint missing: $TARGET" >&2
  exit 1
}

exec bash "$TARGET" "$@"
