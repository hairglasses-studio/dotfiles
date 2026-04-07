#!/usr/bin/env bash
# hg-go-sync.sh — Sync Go version across manifest-backed Go repos.
# Usage: hg-go-sync.sh [--dry-run] [--tidy] [--repos=a,b,c] [--include-compatibility] [--include-deprecated]
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/lib/hg-core.sh"
source "$SCRIPT_DIR/lib/hg-workspace.sh"

VERSION_FILE="$SCRIPT_DIR/../make/go-version"
TARGET_VERSION="$(tr -d '[:space:]' < "$VERSION_FILE")"
DRY_RUN=false
TIDY=false
INCLUDE_COMPATIBILITY=false
INCLUDE_DEPRECATED=false
REPO_FILTER=""

for arg in "$@"; do
  case "$arg" in
    --dry-run) DRY_RUN=true ;;
    --tidy)    TIDY=true ;;
    --include-compatibility) INCLUDE_COMPATIBILITY=true ;;
    --include-deprecated) INCLUDE_DEPRECATED=true ;;
    --repos=*) REPO_FILTER="${arg#--repos=}" ;;
  esac
done

repo_names() {
  local selected=()
  hg_workspace_parse_repo_filter "$REPO_FILTER" selected
  if [[ "${#selected[@]}" -gt 0 ]]; then
    printf '%s\n' "${selected[@]}"
    return
  fi

  local jq_filter='.repos[] | select((.language | tostring | contains("Go")) or (.go_work_member // false))'
  if ! $INCLUDE_COMPATIBILITY; then
    jq_filter+=' | select((.lifecycle // "canonical") != "compatibility")'
  fi
  if ! $INCLUDE_DEPRECATED; then
    jq_filter+=' | select((.lifecycle // "canonical") != "deprecated")'
  fi
  hg_workspace_repo_names "$jq_filter"
}

hg_info "Target Go version: $TARGET_VERSION"
echo ""

CHANGED=0

while IFS= read -r name; do
  [[ -n "$name" ]] || continue
  repo_path="$(hg_workspace_repo_path "$name")"
  [[ -f "$repo_path/go.mod" ]] || continue

  current="$(grep -m1 '^go ' "$repo_path/go.mod" | awk '{print $2}')"
  if [[ "$current" == "$TARGET_VERSION" ]]; then
    printf "%s%-25s %s (current)%s\n" "$HG_DIM" "$name" "$current" "$HG_RESET"
    continue
  fi

  if $DRY_RUN; then
    printf "%s%-25s %s → %s (would update)%s\n" "$HG_YELLOW" "$name" "$current" "$TARGET_VERSION" "$HG_RESET"
    CHANGED=$((CHANGED + 1))
    continue
  fi

  (
    cd "$repo_path"
    sed -i "s/^go .*/go $TARGET_VERSION/" go.mod
    if $TIDY; then
      go mod tidy 2>&1 | tail -3 || hg_warn "$name: go mod tidy had issues"
    fi
  )
  printf "%s%-25s %s → %s%s\n" "$HG_GREEN" "$name" "$current" "$TARGET_VERSION" "$HG_RESET"
  CHANGED=$((CHANGED + 1))
done < <(repo_names)

echo ""
if $DRY_RUN; then
  hg_info "$CHANGED repos would be updated"
else
  hg_ok "$CHANGED repos updated to Go $TARGET_VERSION"
fi
