#!/usr/bin/env bash
# hg-git-hygiene.sh — Scan or clean merged branches and extra worktrees for one repo.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/lib/hg-core.sh"

MODE="dry-run"
REPO_PATH=""
FETCH=0
PRUNE_ADMIN=0
DELETE_LOCAL_MERGED=0
DELETE_REMOTE_MERGED=0
DELETE_CLEAN_WORKTREES=0
PRUNE_MANAGED_STATE=0
JSON=0
BRANCH_PREFIXES=()

usage() {
  cat <<'EOF'
Usage: hg-git-hygiene.sh --repo <path> [options]

Scan or safely clean repo-local git branch and worktree drift. The script is
dry-run by default and only removes branches or worktrees when --execute is set.

Options:
  --repo <path>              Absolute or relative repo path. Required.
  --fetch                    Refresh origin refs with git fetch --prune before scanning.
  --prune-admin              Run git worktree prune (only when --execute is set).
  --delete-merged-local      Delete merged local branches other than the default/current branch.
  --delete-merged-remote     Delete merged origin/* branches other than the default branch.
  --delete-clean-worktrees   Remove extra clean worktrees whose branch tip is merged into default.
  --prune-managed-state      Run hg-codex-worktree-prune.sh for the repo after git cleanup.
  --branch-prefix <prefix>   Limit delete candidates to branches matching the prefix. Repeatable.
  --execute                  Apply the requested cleanup actions.
  --json                     Emit JSON instead of colored human-readable text.
  -h, --help                 Show this help.
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --repo)
      [[ $# -ge 2 ]] || hg_die "--repo requires a path"
      REPO_PATH="$2"
      shift 2
      ;;
    --fetch)
      FETCH=1
      shift
      ;;
    --prune-admin)
      PRUNE_ADMIN=1
      shift
      ;;
    --delete-merged-local)
      DELETE_LOCAL_MERGED=1
      shift
      ;;
    --delete-merged-remote)
      DELETE_REMOTE_MERGED=1
      shift
      ;;
    --delete-clean-worktrees)
      DELETE_CLEAN_WORKTREES=1
      shift
      ;;
    --prune-managed-state)
      PRUNE_MANAGED_STATE=1
      shift
      ;;
    --branch-prefix)
      [[ $# -ge 2 ]] || hg_die "--branch-prefix requires a prefix"
      BRANCH_PREFIXES+=("$2")
      shift 2
      ;;
    --execute)
      MODE="execute"
      shift
      ;;
    --json)
      JSON=1
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

hg_require git jq
[[ -n "$REPO_PATH" ]] || hg_die "--repo is required"

REPO_PATH="$(readlink -f "$REPO_PATH")"
git -C "$REPO_PATH" rev-parse --show-toplevel >/dev/null 2>&1 || hg_die "Not a git repo: $REPO_PATH"
PRIMARY_REPO="$(git -C "$REPO_PATH" rev-parse --show-toplevel)"

TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT

LOCAL_BRANCHES_JSONL="$TMP_DIR/local_branches.jsonl"
REMOTE_BRANCHES_JSONL="$TMP_DIR/remote_branches.jsonl"
WORKTREES_JSONL="$TMP_DIR/worktrees.jsonl"
ACTIONS_JSONL="$TMP_DIR/actions.jsonl"
MANAGED_STATE_OUT="$TMP_DIR/managed_state.out"
: >"$LOCAL_BRANCHES_JSONL"
: >"$REMOTE_BRANCHES_JSONL"
: >"$WORKTREES_JSONL"
: >"$ACTIONS_JSONL"
: >"$MANAGED_STATE_OUT"

DEFAULT_BRANCH=""
CURRENT_BRANCH="$(git -C "$PRIMARY_REPO" branch --show-current || true)"

if origin_head="$(git -C "$PRIMARY_REPO" symbolic-ref --quiet refs/remotes/origin/HEAD 2>/dev/null)"; then
  DEFAULT_BRANCH="${origin_head#refs/remotes/origin/}"
elif git -C "$PRIMARY_REPO" show-ref --verify --quiet refs/heads/main; then
  DEFAULT_BRANCH="main"
elif git -C "$PRIMARY_REPO" show-ref --verify --quiet refs/heads/master; then
  DEFAULT_BRANCH="master"
elif [[ -n "$CURRENT_BRANCH" ]]; then
  DEFAULT_BRANCH="$CURRENT_BRANCH"
else
  DEFAULT_BRANCH="HEAD"
fi

if git -C "$PRIMARY_REPO" show-ref --verify --quiet "refs/heads/$DEFAULT_BRANCH"; then
  DEFAULT_REF="refs/heads/$DEFAULT_BRANCH"
elif git -C "$PRIMARY_REPO" show-ref --verify --quiet "refs/remotes/origin/$DEFAULT_BRANCH"; then
  DEFAULT_REF="refs/remotes/origin/$DEFAULT_BRANCH"
else
  DEFAULT_REF="HEAD"
fi

prefix_matches() {
  local name="$1"
  local prefix
  if [[ "${#BRANCH_PREFIXES[@]}" -eq 0 ]]; then
    return 0
  fi
  for prefix in "${BRANCH_PREFIXES[@]}"; do
    [[ "$name" == "$prefix"* ]] && return 0
  done
  return 1
}

record_action() {
  local kind="$1"
  local target="$2"
  local status="$3"
  local message="$4"
  jq -cn \
    --arg kind "$kind" \
    --arg target "$target" \
    --arg status "$status" \
    --arg message "$message" \
    '{kind:$kind,target:$target,status:$status,message:$message}' >>"$ACTIONS_JSONL"
}

if (( FETCH )); then
  if git -C "$PRIMARY_REPO" fetch --prune origin >/dev/null 2>"$TMP_DIR/fetch.err"; then
    record_action "fetch" "origin" "ok" "Refreshed origin refs with git fetch --prune."
  else
    record_action "fetch" "origin" "fail" "$(tr '\n' ' ' <"$TMP_DIR/fetch.err" | sed 's/[[:space:]]\+/ /g; s/^ //; s/ $//')"
  fi
fi

MERGED_LOCAL_NAMES="$(git -C "$PRIMARY_REPO" for-each-ref refs/heads --format='%(refname:short)' --merged "$DEFAULT_REF" 2>/dev/null || true)"
MERGED_REMOTE_NAMES="$(git -C "$PRIMARY_REPO" for-each-ref refs/remotes/origin --format='%(refname:short)' --merged "$DEFAULT_REF" 2>/dev/null || true)"

WORKTREE_BRANCHES="$(git -C "$PRIMARY_REPO" worktree list --porcelain | awk '/^branch refs\/heads\//{sub(/^branch refs\/heads\//, "", $0); print $0}' | sort -u)"

while IFS= read -r branch; do
  [[ -n "$branch" ]] || continue
  current=false
  merged=false
  eligible=false
  checked_out_in_worktree=false
  ahead=0
  behind=0

  [[ "$branch" == "$CURRENT_BRANCH" ]] && current=true
  grep -Fxq "$branch" <<<"$MERGED_LOCAL_NAMES" && merged=true
  grep -Fxq "$branch" <<<"$WORKTREE_BRANCHES" && checked_out_in_worktree=true
  if git -C "$PRIMARY_REPO" rev-parse --verify --quiet "$DEFAULT_REF" >/dev/null 2>&1; then
    ahead="$(git -C "$PRIMARY_REPO" rev-list --count "$DEFAULT_REF..refs/heads/$branch" 2>/dev/null || printf '0')"
    behind="$(git -C "$PRIMARY_REPO" rev-list --count "refs/heads/$branch..$DEFAULT_REF" 2>/dev/null || printf '0')"
  fi
  if [[ "$branch" != "$DEFAULT_BRANCH" ]] && [[ "$current" == false ]] && [[ "$merged" == true ]]; then
    if prefix_matches "$branch"; then
      eligible=true
    fi
  fi

  jq -cn \
    --arg name "$branch" \
    --arg default_branch "$DEFAULT_BRANCH" \
    --argjson current "$current" \
    --argjson merged "$merged" \
    --argjson eligible "$eligible" \
    --argjson checked_out_in_worktree "$checked_out_in_worktree" \
    --argjson ahead "$ahead" \
    --argjson behind "$behind" \
    '{
      name: $name,
      default_branch: $default_branch,
      current: $current,
      merged_into_default: $merged,
      eligible_for_cleanup: $eligible,
      checked_out_in_worktree: $checked_out_in_worktree,
      ahead: $ahead,
      behind: $behind
    }' >>"$LOCAL_BRANCHES_JSONL"
done < <(git -C "$PRIMARY_REPO" for-each-ref refs/heads --format='%(refname:short)' | sort)

while IFS= read -r branch; do
  [[ -n "$branch" ]] || continue
  [[ "$branch" == "origin/HEAD" ]] && continue
  short="${branch#origin/}"
  merged=false
  eligible=false
  grep -Fxq "$branch" <<<"$MERGED_REMOTE_NAMES" && merged=true
  if [[ "$short" != "$DEFAULT_BRANCH" ]] && [[ "$merged" == true ]]; then
    if prefix_matches "$short"; then
      eligible=true
    fi
  fi
  jq -cn \
    --arg name "$short" \
    --arg full_ref "$branch" \
    --argjson merged "$merged" \
    --argjson eligible "$eligible" \
    '{
      name: $name,
      full_ref: $full_ref,
      merged_into_default: $merged,
      eligible_for_cleanup: $eligible
    }' >>"$REMOTE_BRANCHES_JSONL"
done < <(git -C "$PRIMARY_REPO" for-each-ref refs/remotes/origin --format='%(refname:short)' | sort)

flush_worktree() {
  local path="$1"
  local branch="$2"
  local prunable="$3"
  [[ -n "$path" ]] || return 0

  local current=false missing=false dirty=false merged=false eligible=false
  [[ "$path" == "$PRIMARY_REPO" ]] && current=true
  [[ -e "$path" ]] || missing=true

  if [[ "$missing" == false ]]; then
    if [[ -n "$(git -C "$path" status --porcelain --untracked-files=normal 2>/dev/null || true)" ]]; then
      dirty=true
    fi
  fi

  if [[ -n "$branch" && "$branch" != "DETACHED" ]] && git -C "$PRIMARY_REPO" show-ref --verify --quiet "refs/heads/$branch"; then
    if git -C "$PRIMARY_REPO" merge-base --is-ancestor "refs/heads/$branch" "$DEFAULT_REF" >/dev/null 2>&1; then
      merged=true
    fi
  fi

  if [[ "$current" == false ]] && [[ "$missing" == false ]] && [[ "$dirty" == false ]] && [[ "$merged" == true ]]; then
    if [[ -z "$branch" || "$branch" == "DETACHED" ]]; then
      eligible=false
    elif prefix_matches "$branch"; then
      eligible=true
    fi
  fi

  jq -cn \
    --arg path "$path" \
    --arg branch "$branch" \
    --argjson current "$current" \
    --argjson missing "$missing" \
    --argjson dirty "$dirty" \
    --argjson merged "$merged" \
    --argjson eligible "$eligible" \
    --argjson prunable "$prunable" \
    '{
      path: $path,
      branch: $branch,
      current: $current,
      missing: $missing,
      dirty: $dirty,
      merged_into_default: $merged,
      eligible_for_cleanup: $eligible,
      prunable: $prunable
    }' >>"$WORKTREES_JSONL"
}

current_path=""
current_wt_branch=""
current_prunable=false
while IFS= read -r line || [[ -n "$line" ]]; do
  if [[ -z "$line" ]]; then
    flush_worktree "$current_path" "$current_wt_branch" "$current_prunable"
    current_path=""
    current_wt_branch=""
    current_prunable=false
    continue
  fi
  case "$line" in
    worktree\ *)
      current_path="${line#worktree }"
      ;;
    branch\ refs/heads/*)
      current_wt_branch="${line#branch refs/heads/}"
      ;;
    detached)
      current_wt_branch="DETACHED"
      ;;
    prunable*)
      current_prunable=true
      ;;
  esac
done < <(git -C "$PRIMARY_REPO" worktree list --porcelain)
flush_worktree "$current_path" "$current_wt_branch" "$current_prunable"

if (( DELETE_CLEAN_WORKTREES )); then
  while IFS= read -r line; do
    [[ -n "$line" ]] || continue
    path="$(jq -r '.path' <<<"$line")"
    branch="$(jq -r '.branch' <<<"$line")"
    eligible="$(jq -r '.eligible_for_cleanup' <<<"$line")"
    [[ "$eligible" == "true" ]] || continue
    if [[ "$MODE" == "execute" ]]; then
      if git -C "$PRIMARY_REPO" worktree remove "$path" >/dev/null 2>"$TMP_DIR/worktree.err"; then
        record_action "worktree_remove" "$path" "ok" "Removed clean merged worktree for branch $branch."
      else
        record_action "worktree_remove" "$path" "fail" "$(tr '\n' ' ' <"$TMP_DIR/worktree.err" | sed 's/[[:space:]]\+/ /g; s/^ //; s/ $//')"
      fi
    else
      record_action "worktree_remove" "$path" "planned" "Would remove clean merged worktree for branch $branch."
    fi
  done <"$WORKTREES_JSONL"
fi

if (( DELETE_LOCAL_MERGED )); then
  while IFS= read -r line; do
    [[ -n "$line" ]] || continue
    branch="$(jq -r '.name' <<<"$line")"
    eligible="$(jq -r '.eligible_for_cleanup' <<<"$line")"
    checked_out_in_worktree="$(jq -r '.checked_out_in_worktree' <<<"$line")"
    [[ "$eligible" == "true" ]] || continue
    if [[ "$MODE" == "execute" ]]; then
      if git -C "$PRIMARY_REPO" branch -d "$branch" >/dev/null 2>"$TMP_DIR/local.err"; then
        record_action "branch_delete_local" "$branch" "ok" "Deleted merged local branch."
      else
        record_action "branch_delete_local" "$branch" "fail" "$(tr '\n' ' ' <"$TMP_DIR/local.err" | sed 's/[[:space:]]\+/ /g; s/^ //; s/ $//')"
      fi
    else
      message="Would delete merged local branch."
      if [[ "$checked_out_in_worktree" == "true" ]]; then
        message="Would delete merged local branch after removing its attached worktree."
      fi
      record_action "branch_delete_local" "$branch" "planned" "$message"
    fi
  done <"$LOCAL_BRANCHES_JSONL"
fi

if (( DELETE_REMOTE_MERGED )); then
  while IFS= read -r line; do
    [[ -n "$line" ]] || continue
    branch="$(jq -r '.name' <<<"$line")"
    eligible="$(jq -r '.eligible_for_cleanup' <<<"$line")"
    [[ "$eligible" == "true" ]] || continue
    if [[ "$MODE" == "execute" ]]; then
      if git -C "$PRIMARY_REPO" push origin --delete "$branch" >/dev/null 2>"$TMP_DIR/remote.err"; then
        record_action "branch_delete_remote" "$branch" "ok" "Deleted merged remote branch from origin."
      else
        record_action "branch_delete_remote" "$branch" "fail" "$(tr '\n' ' ' <"$TMP_DIR/remote.err" | sed 's/[[:space:]]\+/ /g; s/^ //; s/ $//')"
      fi
    else
      record_action "branch_delete_remote" "$branch" "planned" "Would delete merged remote branch from origin."
    fi
  done <"$REMOTE_BRANCHES_JSONL"
fi

if (( PRUNE_ADMIN )); then
  if [[ "$MODE" == "execute" ]]; then
    if git -C "$PRIMARY_REPO" worktree prune >/dev/null 2>"$TMP_DIR/prune.err"; then
      record_action "worktree_prune_admin" "$PRIMARY_REPO" "ok" "Pruned stale git worktree admin state."
    else
      record_action "worktree_prune_admin" "$PRIMARY_REPO" "fail" "$(tr '\n' ' ' <"$TMP_DIR/prune.err" | sed 's/[[:space:]]\+/ /g; s/^ //; s/ $//')"
    fi
  else
    record_action "worktree_prune_admin" "$PRIMARY_REPO" "planned" "Would run git worktree prune."
  fi
fi

if (( PRUNE_MANAGED_STATE )); then
  if [[ "$MODE" == "execute" ]]; then
    if bash "$SCRIPT_DIR/hg-codex-worktree-prune.sh" --repo "$PRIMARY_REPO" >"$MANAGED_STATE_OUT" 2>&1; then
      record_action "managed_worktree_prune" "$PRIMARY_REPO" "ok" "Pruned Codex-managed worktree state."
    else
      record_action "managed_worktree_prune" "$PRIMARY_REPO" "fail" "$(tr '\n' ' ' <"$MANAGED_STATE_OUT" | sed 's/[[:space:]]\+/ /g; s/^ //; s/ $//')"
    fi
  else
    if bash "$SCRIPT_DIR/hg-codex-worktree-prune.sh" --dry-run --repo "$PRIMARY_REPO" >"$MANAGED_STATE_OUT" 2>&1; then
      record_action "managed_worktree_prune" "$PRIMARY_REPO" "planned" "Would prune Codex-managed worktree state."
    else
      record_action "managed_worktree_prune" "$PRIMARY_REPO" "fail" "$(tr '\n' ' ' <"$MANAGED_STATE_OUT" | sed 's/[[:space:]]\+/ /g; s/^ //; s/ $//')"
    fi
  fi
fi

local_branches_json="$(jq -s '.' "$LOCAL_BRANCHES_JSONL")"
remote_branches_json="$(jq -s '.' "$REMOTE_BRANCHES_JSONL")"
worktrees_json="$(jq -s '.' "$WORKTREES_JSONL")"
actions_json="$(jq -s '.' "$ACTIONS_JSONL")"
managed_preview_json="$(jq -Rcs 'split("\n") | map(select(length > 0))' "$MANAGED_STATE_OUT")"

summary_json="$(
  jq -n \
    --argjson local_branches "$local_branches_json" \
    --argjson remote_branches "$remote_branches_json" \
    --argjson worktrees "$worktrees_json" \
    --argjson actions "$actions_json" '
      {
        local_branch_count: ($local_branches | length),
        local_merged_count: ($local_branches | map(select(.merged_into_default)) | length),
        local_cleanup_candidate_count: ($local_branches | map(select(.eligible_for_cleanup)) | length),
        remote_branch_count: ($remote_branches | length),
        remote_merged_count: ($remote_branches | map(select(.merged_into_default)) | length),
        remote_cleanup_candidate_count: ($remote_branches | map(select(.eligible_for_cleanup)) | length),
        extra_worktree_count: ($worktrees | map(select(.current | not)) | length),
        clean_merged_worktree_count: ($worktrees | map(select(.eligible_for_cleanup)) | length),
        dirty_worktree_count: ($worktrees | map(select(.dirty)) | length),
        blocked_worktree_count: ($worktrees | map(select((.current | not) and (.eligible_for_cleanup | not))) | length),
        action_count: ($actions | length),
        completed_action_count: ($actions | map(select(.status == "ok")) | length),
        planned_action_count: ($actions | map(select(.status == "planned")) | length),
        failed_action_count: ($actions | map(select(.status == "fail")) | length)
      }'
)"

report_json="$(
  jq -n \
    --arg repo "$PRIMARY_REPO" \
    --arg default_branch "$DEFAULT_BRANCH" \
    --arg current_branch "$CURRENT_BRANCH" \
    --arg mode "$MODE" \
    --argjson fetch $FETCH \
    --argjson prune_admin $PRUNE_ADMIN \
    --argjson delete_merged_local $DELETE_LOCAL_MERGED \
    --argjson delete_merged_remote $DELETE_REMOTE_MERGED \
    --argjson delete_clean_worktrees $DELETE_CLEAN_WORKTREES \
    --argjson prune_managed_state $PRUNE_MANAGED_STATE \
    --argjson branch_prefixes "$(printf '%s\n' "${BRANCH_PREFIXES[@]}" | jq -R . | jq -s '.')" \
    --argjson local_branches "$local_branches_json" \
    --argjson remote_branches "$remote_branches_json" \
    --argjson worktrees "$worktrees_json" \
    --argjson actions "$actions_json" \
    --argjson managed_state_preview "$managed_preview_json" \
    --argjson summary "$summary_json" '
      {
        repo: $repo,
        default_branch: $default_branch,
        current_branch: $current_branch,
        mode: $mode,
        options: {
          fetch: $fetch,
          prune_admin: $prune_admin,
          delete_merged_local: $delete_merged_local,
          delete_merged_remote: $delete_merged_remote,
          delete_clean_worktrees: $delete_clean_worktrees,
          prune_managed_state: $prune_managed_state,
          branch_prefixes: $branch_prefixes
        },
        summary: $summary,
        local_branches: $local_branches,
        remote_branches: $remote_branches,
        worktrees: $worktrees,
        actions: $actions,
        managed_state_preview: $managed_state_preview
      }'
)"

if (( JSON )); then
  printf '%s\n' "$report_json"
  exit 0
fi

summary_line="$(jq -r '[.repo, (.summary.local_cleanup_candidate_count|tostring), (.summary.remote_cleanup_candidate_count|tostring), (.summary.clean_merged_worktree_count|tostring), (.summary.failed_action_count|tostring)] | @tsv' <<<"$report_json")"
IFS=$'\t' read -r repo_out local_candidates remote_candidates worktree_candidates failed_actions <<<"$summary_line"

hg_info "Repo: $repo_out"
hg_info "Default branch: $DEFAULT_BRANCH"
hg_info "Mode: $MODE"
hg_info "Cleanup candidates -> local merged: $local_candidates, remote merged: $remote_candidates, worktrees: $worktree_candidates"
if [[ "$failed_actions" != "0" ]]; then
  hg_warn "One or more cleanup actions failed. Re-run with --json for details."
fi

if (( DELETE_LOCAL_MERGED || DELETE_REMOTE_MERGED || DELETE_CLEAN_WORKTREES || PRUNE_ADMIN || PRUNE_MANAGED_STATE || FETCH )); then
  while IFS= read -r action; do
    [[ -n "$action" ]] || continue
    status="$(jq -r '.status' <<<"$action")"
    kind="$(jq -r '.kind' <<<"$action")"
    target="$(jq -r '.target' <<<"$action")"
    message="$(jq -r '.message' <<<"$action")"
    case "$status" in
      ok) hg_ok "$kind :: $target :: $message" ;;
      planned) hg_warn "$kind :: $target :: $message" ;;
      fail) hg_error "$kind :: $target :: $message" ;;
      *) hg_info "$kind :: $target :: $message" ;;
    esac
  done <"$ACTIONS_JSONL"
fi
