#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/lib/hg-core.sh"

export HG_STUDIO_ROOT HG_DOTFILES
TARGET="$HG_STUDIO_ROOT/surfacekit/scripts/provider-settings-sync.sh"

[[ -f "$TARGET" ]] || {
  echo "surfacekit provider sync entrypoint missing: $TARGET" >&2
  exit 1
}

exec bash "$TARGET" "$@"
