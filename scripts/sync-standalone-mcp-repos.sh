#!/usr/bin/env bash
# sync-standalone-mcp-repos.sh — Mirror canonical dotfiles MCP modules into standalone publish repos.
# Usage: sync-standalone-mcp-repos.sh [bootstrap|sync|check|hygiene] [--repos=a,b] [--allow-dirty] [--refresh-origin] [--repair-bare-main]
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/lib/hg-core.sh"
source "$SCRIPT_DIR/lib/hg-workspace.sh"

MODE="sync"
REPO_FILTER=""
ALLOW_DIRTY=false
REFRESH_ORIGIN=false
REPAIR_BARE_MAIN=false

usage() {
  cat <<'EOF'
Usage: sync-standalone-mcp-repos.sh [bootstrap|sync|check|hygiene] [--repos=a,b] [--allow-dirty] [--refresh-origin] [--repair-bare-main]

Mirror canonical MCP source trees from dotfiles/mcp/* into standalone publish repos
declared as lifecycle=mirror in workspace/manifest.json.

Modes:
  bootstrap  Initialize tree-sync mirrors
  sync       Update tree-sync mirrors
  check      Verify tree-sync mirrors and run manual-projection planners
  hygiene    Inspect bare mirror repo branch hygiene; use --repair-bare-main to
             align stale local refs/heads/main to refs/remotes/origin/main
             after optionally refreshing origin with --refresh-origin
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    bootstrap|sync|check|hygiene)
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
    --refresh-origin)
      REFRESH_ORIGIN=true
      shift
      ;;
    --repair-bare-main)
      REPAIR_BARE_MAIN=true
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

manual_projection_helper() {
  local repo="$1"
  local candidate="$SCRIPT_DIR/hg-${repo}-projection.sh"
  [[ -f "$candidate" ]] || return 1
  printf '%s\n' "$candidate"
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
ISSUES=0
REPAIRED=0

git_ref_value() {
  local repo_path="$1"
  local ref="$2"
  git -C "$repo_path" rev-parse --verify -q "$ref" 2>/dev/null || true
}

short_sha() {
  local value="${1:-}"
  if [[ -z "$value" ]]; then
    printf '%s\n' "-"
  else
    printf '%.12s\n' "$value"
  fi
}

hygiene_check_repo() {
  local repo="$1"
  local repo_path="$2"

  if ! git -C "$repo_path" rev-parse --git-dir >/dev/null 2>&1; then
    hg_warn "$repo: hygiene skipped because path is not a git repo: $repo_path"
    SKIPPED=$((SKIPPED + 1))
    return 0
  fi

  local is_bare
  is_bare="$(git -C "$repo_path" rev-parse --is-bare-repository 2>/dev/null || printf 'false')"
  if [[ "$is_bare" != "true" ]]; then
    hg_info "$repo: hygiene skipped because repo is not bare"
    SKIPPED=$((SKIPPED + 1))
    return 0
  fi

  if $REFRESH_ORIGIN || $REPAIR_BARE_MAIN; then
    git -C "$repo_path" fetch --prune origin '+refs/heads/*:refs/remotes/origin/*' >/dev/null 2>&1 || true
  fi

  local head_ref local_main remote_main status
  head_ref="$(git -C "$repo_path" symbolic-ref -q HEAD 2>/dev/null || true)"
  local_main="$(git_ref_value "$repo_path" refs/heads/main)"
  remote_main="$(git_ref_value "$repo_path" refs/remotes/origin/main)"

  status="unknown"
  if [[ -z "$remote_main" ]]; then
    status="missing_origin_main"
  elif [[ -z "$local_main" ]]; then
    status="missing_local_main"
  elif [[ "$local_main" == "$remote_main" ]]; then
    status="in_sync"
  elif git -C "$repo_path" merge-base --is-ancestor "$local_main" "$remote_main" >/dev/null 2>&1; then
    status="stale_local_main"
  else
    status="divergent_local_main"
  fi

  if $REPAIR_BARE_MAIN && [[ -n "$remote_main" ]] && ([[ "$status" == "missing_local_main" ]] || [[ "$status" == "stale_local_main" ]]); then
    git -C "$repo_path" update-ref refs/heads/main "$remote_main"
    if [[ "$head_ref" != "refs/heads/main" ]]; then
      git -C "$repo_path" symbolic-ref HEAD refs/heads/main >/dev/null 2>&1 || true
    fi
    local_main="$remote_main"
    status="in_sync"
    REPAIRED=$((REPAIRED + 1))
  fi

  case "$status" in
    in_sync)
      hg_ok "$repo: bare main in sync (local=$(short_sha "$local_main") origin=$(short_sha "$remote_main"))"
      ;;
    stale_local_main)
      hg_warn "$repo: bare main stale (local=$(short_sha "$local_main") origin=$(short_sha "$remote_main")); rerun with hygiene --repair-bare-main"
      ISSUES=$((ISSUES + 1))
      ;;
    missing_local_main)
      hg_warn "$repo: bare repo missing refs/heads/main while origin/main=$(short_sha "$remote_main"); rerun with hygiene --repair-bare-main"
      ISSUES=$((ISSUES + 1))
      ;;
    divergent_local_main)
      hg_warn "$repo: bare main diverged (local=$(short_sha "$local_main") origin=$(short_sha "$remote_main")); inspect manually before repair"
      ISSUES=$((ISSUES + 1))
      ;;
    missing_origin_main)
      hg_warn "$repo: bare repo missing refs/remotes/origin/main; fetch/configure origin first"
      ISSUES=$((ISSUES + 1))
      ;;
    *)
      hg_warn "$repo: bare hygiene status unknown"
      ISSUES=$((ISSUES + 1))
      ;;
  esac

  UPDATED=$((UPDATED + 1))
  return 0
}

while IFS= read -r repo; do
  [[ -n "$repo" ]] || continue

  canonical_rel="$(hg_workspace_repo_field "$repo" "mirror_of" "")"
  [[ -n "$canonical_rel" ]] || hg_die "$repo is missing mirror_of in workspace manifest"

  mirror_path="$(hg_workspace_repo_path "$repo")"
  canonical_path="$HG_STUDIO_ROOT/$canonical_rel"

  if [[ "$MODE" == "hygiene" ]]; then
    [[ -e "$mirror_path" ]] || hg_die "mirror repo missing: $mirror_path"
    hygiene_check_repo "$repo" "$mirror_path"
    continue
  fi

  [[ -e "$mirror_path/.git" ]] || hg_die "mirror repo missing: $mirror_path"
  [[ -d "$canonical_path" ]] || hg_die "canonical path missing: $canonical_path"

  sync_strategy="$(mirror_sync_strategy "$repo" "$canonical_path")"
  if [[ "$sync_strategy" != "tree_sync" ]]; then
    if helper="$(manual_projection_helper "$repo")"; then
      helper_mode="plan"
      if [[ "$MODE" == "check" ]]; then
        helper_mode="check"
      fi
      bash "$helper" "$helper_mode" --canonical "$canonical_path" --standalone "$mirror_path"
      hg_warn "$repo: generic $MODE skipped because sync_strategy=$sync_strategy uses repo-specific helper $(basename "$helper")"
    else
      hg_warn "$repo: skipping $MODE because sync_strategy=$sync_strategy requires a dedicated projection workflow"
    fi
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
if [[ "$MODE" == "hygiene" ]]; then
  if $REPAIR_BARE_MAIN; then
    hg_info "$REPAIRED bare mirror repos repaired"
  fi
  if [[ "$ISSUES" -gt 0 ]]; then
    exit 1
  fi
fi
