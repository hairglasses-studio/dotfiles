#!/usr/bin/env bash
# hg-go-sync.sh — Sync Go version across all hairglasses-studio Go repos.
# Usage: hg-go-sync.sh [--dry-run] [--tidy]
# Reads target version from dotfiles/make/go-version.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/lib/hg-core.sh"

STUDIO="$HOME/hairglasses-studio"
VERSION_FILE="$SCRIPT_DIR/../make/go-version"
TARGET_VERSION=$(cat "$VERSION_FILE" | tr -d '[:space:]')
DRY_RUN=false
TIDY=false

for arg in "$@"; do
  case "$arg" in
    --dry-run) DRY_RUN=true ;;
    --tidy)    TIDY=true ;;
  esac
done

hg_info "Target Go version: $TARGET_VERSION"
echo ""

CHANGED=0
SKIPPED=0

for d in "$STUDIO"/*/; do
  [[ -f "$d/go.mod" ]] || continue
  name=$(basename "$d")
  current=$(grep -m1 '^go ' "$d/go.mod" | awk '{print $2}')

  if [[ "$current" == "$TARGET_VERSION" ]]; then
    printf "%s%-25s %s (current)%s\n" "$HG_DIM" "$name" "$current" "$HG_RESET"
    continue
  fi

  if $DRY_RUN; then
    printf "%s%-25s %s → %s (would update)%s\n" "$HG_YELLOW" "$name" "$current" "$TARGET_VERSION" "$HG_RESET"
    CHANGED=$((CHANGED + 1))
  else
    cd "$d"
    sed -i "s/^go .*/go $TARGET_VERSION/" go.mod
    if $TIDY; then
      go mod tidy 2>&1 | tail -3 || hg_warn "$name: go mod tidy had issues"
    fi
    printf "%s%-25s %s → %s%s\n" "$HG_GREEN" "$name" "$current" "$TARGET_VERSION" "$HG_RESET"
    CHANGED=$((CHANGED + 1))
  fi
done

echo ""
if $DRY_RUN; then
  hg_info "$CHANGED repos would be updated"
else
  hg_ok "$CHANGED repos updated to Go $TARGET_VERSION"
fi
