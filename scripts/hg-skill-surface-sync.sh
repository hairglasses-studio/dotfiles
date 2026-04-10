#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/lib/hg-core.sh"

export HG_STUDIO_ROOT HG_DOTFILES
REPO_PATH="${1:-$PWD}"
MODE_ARGS=()

if [[ $# -gt 0 ]]; then
  shift
fi

for arg in "$@"; do
  case "$arg" in
    --check|--dry-run)
      MODE_ARGS+=("$arg")
      ;;
    *)
      hg_die "Unknown argument: $arg"
      ;;
  esac
done

REPO_PATH="$(cd "$REPO_PATH" && pwd -P)"
SELF_PATH="$SCRIPT_DIR/$(basename "${BASH_SOURCE[0]}")"

for target in \
  "$REPO_PATH/scripts/skill-surface-sync.sh" \
  "$REPO_PATH/scripts/hg-skill-surface-sync.sh"
do
  [[ -f "$target" ]] || continue
  if [[ "$(cd "$(dirname "$target")" && pwd -P)/$(basename "$target")" == "$SELF_PATH" ]]; then
    continue
  fi
  exec bash "$target" "$REPO_PATH" "${MODE_ARGS[@]}"
done

for target in \
  "$HG_STUDIO_ROOT/codexkit/scripts/skill-surface-sync.sh" \
  "$HG_STUDIO_ROOT/surfacekit/scripts/skill-surface-sync.sh"
do
  [[ -f "$target" ]] || continue
  exec bash "$target" "$REPO_PATH" "${MODE_ARGS[@]}"
done

hg_die "No managed skill sync entrypoint found for $REPO_PATH"
