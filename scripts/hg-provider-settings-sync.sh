#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/lib/hg-core.sh"

export HG_STUDIO_ROOT HG_DOTFILES
TARGET="$HG_STUDIO_ROOT/codexkit/scripts/provider-settings-sync.sh"

[[ -f "$TARGET" ]] || {
  echo "codexkit provider sync entrypoint missing: $TARGET" >&2
  exit 1
}

has_include_codex_config=false
for arg in "$@"; do
  if [[ "$arg" == "--include-codex-config" ]]; then
    has_include_codex_config=true
    break
  fi
done

if $has_include_codex_config; then
  exec bash "$TARGET" "$@"
fi

exec bash "$TARGET" "$@" --include-codex-config
