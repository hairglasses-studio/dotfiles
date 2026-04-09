#!/usr/bin/env bash
# hg-codex-worktree-prune.sh — Prune orphaned managed worktree state and optional old worktrees.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$(readlink -f "${BASH_SOURCE[0]}")")" && pwd)"
source "$SCRIPT_DIR/lib/hg-agent-launch.sh"

MODE="write"
REPO_PATH=""
KEEP_LATEST=0

usage() {
  cat <<'EOF'
Usage: hg-codex-worktree-prune.sh [--dry-run] [--repo <path>] [--keep-latest <n>]

By default this removes orphaned metadata and asks git to prune stale admin
state for every source repo referenced by the managed worktree metadata.

Options:
  --dry-run          Report changes without deleting anything.
  --repo <path>      Limit pruning to one source repo.
  --keep-latest <n>  When combined with --repo, remove older managed worktrees
                     beyond the newest N entries for that repo.
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --dry-run)
      MODE="dry-run"
      shift
      ;;
    --repo)
      [[ $# -ge 2 ]] || hg_die "--repo requires a path"
      REPO_PATH="$2"
      shift 2
      ;;
    --keep-latest)
      [[ $# -ge 2 ]] || hg_die "--keep-latest requires a value"
      KEEP_LATEST="$2"
      shift 2
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

hg_agent_require_tools

[[ -d "$HG_AGENT_STATE_DIR" ]] || exit 0

remove_entry() {
  local repo_path="$1"
  local worktree_path="$2"
  local branch_name="$3"
  local state_file="$4"

  if [[ "$MODE" == "dry-run" ]]; then
    hg_warn "Would prune managed worktree: $worktree_path"
    return
  fi

  if [[ -d "$worktree_path" ]]; then
    git -C "$repo_path" worktree remove --force "$worktree_path" >/dev/null 2>&1 || rm -rf "$worktree_path"
  fi
  rm -f "$state_file"
  git -C "$repo_path" branch -D "$branch_name" >/dev/null 2>&1 || true
  hg_ok "Pruned managed worktree: $worktree_path"
}

declare -A seen_repo=()
mapfile -t state_files < <(find "$HG_AGENT_STATE_DIR" -maxdepth 1 -type f -name '*.json' | sort)

for state_file in "${state_files[@]}"; do
  repo_path="$(jq -r '.source_repo' "$state_file")"
  worktree_path="$(jq -r '.worktree_path' "$state_file")"
  branch_name="$(jq -r '.branch_name' "$state_file")"

  if [[ -n "$REPO_PATH" && "$(cd "$repo_path" 2>/dev/null && pwd -P)" != "$(cd "$REPO_PATH" && pwd -P)" ]]; then
    continue
  fi

  if [[ ! -d "$repo_path/.git" && ! -f "$repo_path/.git" ]]; then
    if [[ "$MODE" == "dry-run" ]]; then
      hg_warn "Would remove orphaned metadata: $state_file"
    else
      rm -f "$state_file"
      hg_ok "Removed orphaned metadata: $state_file"
    fi
    continue
  fi

  if [[ ! -d "$worktree_path" ]]; then
    if [[ "$MODE" == "dry-run" ]]; then
      hg_warn "Would remove stale metadata for missing worktree: $worktree_path"
    else
      rm -f "$state_file"
      hg_ok "Removed stale metadata for missing worktree: $worktree_path"
    fi
    continue
  fi

  seen_repo["$repo_path"]=1
done

for repo_path in "${!seen_repo[@]}"; do
  if [[ "$MODE" == "dry-run" ]]; then
    hg_warn "Would run git worktree prune in $repo_path"
  else
    git -C "$repo_path" worktree prune >/dev/null
    hg_ok "Pruned git worktree admin state: $repo_path"
  fi
done

if [[ -n "$REPO_PATH" && "$KEEP_LATEST" -gt 0 ]]; then
  mapfile -t repo_state_files < <(
    for state_file in "${state_files[@]}"; do
      repo_path="$(jq -r '.source_repo' "$state_file")"
      [[ "$(cd "$repo_path" 2>/dev/null && pwd -P)" == "$(cd "$REPO_PATH" && pwd -P)" ]] || continue
      created_at="$(jq -r '.created_at' "$state_file")"
      printf '%s\t%s\n' "$created_at" "$state_file"
    done | sort -r
  )

  keep_count=0
  for entry in "${repo_state_files[@]}"; do
    state_file="${entry#*$'\t'}"
    keep_count=$((keep_count + 1))
    if [[ "$keep_count" -le "$KEEP_LATEST" ]]; then
      continue
    fi
    repo_path="$(jq -r '.source_repo' "$state_file")"
    worktree_path="$(jq -r '.worktree_path' "$state_file")"
    branch_name="$(jq -r '.branch_name' "$state_file")"
    remove_entry "$repo_path" "$worktree_path" "$branch_name" "$state_file"
  done
fi
