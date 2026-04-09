#!/usr/bin/env bash
# sync-standalone-mcp-repos.sh — Mirror canonical dotfiles MCP modules into standalone publish repos.
# Usage: sync-standalone-mcp-repos.sh [bootstrap|sync|check] [--repos=a,b] [--allow-dirty]
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/lib/hg-core.sh"
source "$SCRIPT_DIR/lib/hg-workspace.sh"

MODE="sync"
REPO_FILTER=""
ALLOW_DIRTY=false

usage() {
  cat <<'EOF'
Usage: sync-standalone-mcp-repos.sh [bootstrap|sync|check] [--repos=a,b] [--allow-dirty]

Mirror canonical MCP source trees from dotfiles/mcp/* into standalone publish repos
declared as lifecycle=mirror in workspace/manifest.json.
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    bootstrap|sync|check)
      MODE="$1"
      shift
      ;;
    --repos=*)
      REPO_FILTER="${1#--repos=}"
      shift
      ;;
    --allow-dirty)
      ALLOW_DIRTY=true
      shift
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      hg_die "Unknown argument: $1"
      ;;
  esac
done

hg_require git

mirror_parity_manifest() {
  printf '%s\n' "${HG_MCP_MIRROR_MANIFEST:-$HG_DOTFILES/mcp/mirror-parity.json}"
}

abspath() {
  local target="$1"
  if [[ -d "$target" ]]; then
    (cd "$target" && pwd)
  else
    local dir base
    dir="$(cd "$(dirname "$target")" && pwd)"
    base="$(basename "$target")"
    printf '%s/%s\n' "$dir" "$base"
  fi
}

git_path_dirty() {
  local target="$1"
  local repo_root abs_target abs_root rel

  if [[ -d "$target" ]]; then
    repo_root="$(git -C "$target" rev-parse --show-toplevel 2>/dev/null)" || return 1
  else
    repo_root="$(git -C "$(dirname "$target")" rev-parse --show-toplevel 2>/dev/null)" || return 1
  fi

  abs_target="$(abspath "$target")"
  abs_root="$(cd "$repo_root" && pwd)"
  rel="${abs_target#$abs_root/}"
  if [[ "$rel" == "$abs_target" ]]; then
    rel="."
  fi

  [[ -n "$(git -C "$repo_root" status --porcelain --untracked-files=all -- "$rel" 2>/dev/null)" ]]
}

mirror_sync_strategy() {
  local repo="$1"
  local canonical_path="$2"
  local manifest canonical_rel
  manifest="$(mirror_parity_manifest)"
  [[ -f "$manifest" ]] || {
    printf 'tree_sync\n'
    return
  }

  canonical_rel="${canonical_path#$HG_DOTFILES/}"
  jq -r \
    --arg repo "$repo" \
    --arg canonical "$canonical_rel" \
    'first(.mirrors[] | select(.standalone_repo == $repo or .canonical_path == $canonical) | (.sync_strategy // "tree_sync")) // "tree_sync"' \
    "$manifest"
}

repo_names() {
  local selected=()
  hg_workspace_parse_repo_filter "$REPO_FILTER" selected
  if [[ "${#selected[@]}" -gt 0 ]]; then
    local repo
    for repo in "${selected[@]}"; do
      if [[ "$(hg_workspace_repo_field "$repo" "lifecycle" "")" != "mirror" ]]; then
        hg_die "$repo is not a manifest mirror repo"
      fi
      printf '%s\n' "$repo"
    done
    return
  fi

  hg_workspace_repo_names '.repos[] | select((.lifecycle // "") == "mirror" and (.mirror_of // "") != "")'
}

UPDATED=0
SKIPPED=0

while IFS= read -r repo; do
  [[ -n "$repo" ]] || continue

  canonical_rel="$(hg_workspace_repo_field "$repo" "mirror_of" "")"
  [[ -n "$canonical_rel" ]] || hg_die "$repo is missing mirror_of in workspace manifest"

  mirror_path="$(hg_workspace_repo_path "$repo")"
  canonical_path="$HG_STUDIO_ROOT/$canonical_rel"

  [[ -e "$mirror_path/.git" ]] || hg_die "mirror repo missing: $mirror_path"
  [[ -d "$canonical_path" ]] || hg_die "canonical path missing: $canonical_path"

  sync_strategy="$(mirror_sync_strategy "$repo" "$canonical_path")"
  if [[ "$sync_strategy" != "tree_sync" ]]; then
    hg_warn "$repo: skipping $MODE because sync_strategy=$sync_strategy requires a dedicated projection workflow"
    SKIPPED=$((SKIPPED + 1))
    continue
  fi

  if ! $ALLOW_DIRTY && [[ "$MODE" != "check" ]]; then
    if git_path_dirty "$canonical_path"; then
      hg_warn "$repo: skipping dirty canonical path $canonical_rel"
      continue
    fi
    if git_path_dirty "$mirror_path"; then
      hg_warn "$repo: skipping dirty mirror repo"
      continue
    fi
  fi

  bash "$SCRIPT_DIR/mcp-mirror.sh" "$MODE" --canonical "$canonical_path" --mirror "$mirror_path"
  hg_ok "$repo: $MODE complete"
  UPDATED=$((UPDATED + 1))
done < <(repo_names)

hg_info "$UPDATED mirror repos processed"
if [[ "$SKIPPED" -gt 0 ]]; then
  hg_info "$SKIPPED mirror repos skipped due to non-tree sync strategy"
fi
