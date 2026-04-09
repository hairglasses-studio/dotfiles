#!/usr/bin/env bash
set -euo pipefail

SCRIPT_PATH="$(readlink -f "${BASH_SOURCE[0]:-$0}")"
SCRIPT_DIR="$(cd "$(dirname "$SCRIPT_PATH")" && pwd)"
source "$SCRIPT_DIR/lib/hg-core.sh"

json_mode=false
repo_path="$PWD"

while [[ $# -gt 0 ]]; do
  case "$1" in
    --json)
      json_mode=true
      shift
      ;;
    -*)
      hg_die "Unknown option: $1"
      ;;
    *)
      repo_path="$1"
      shift
      ;;
  esac
done

repo_path="$(cd "$repo_path" 2>/dev/null && pwd)" || hg_die "Repo path not found: $repo_path"
hg_require git

emit_json() {
  hg_require jq
  jq -n \
    --arg repo "$repo_root" \
    --arg branch "$branch" \
    --arg upstream "$upstream" \
    --arg compare_ref "$compare_ref" \
    --arg remote_url "$remote_url" \
    --arg lane "$lane" \
    --arg reason "$reason" \
    --arg ssh_probe "$ssh_probe" \
    --arg ssh_detail "$ssh_detail" \
    --argjson dirty_tracked "$dirty_tracked" \
    --argjson dirty_untracked "$dirty_untracked" \
    --argjson ahead "$ahead" \
    --argjson behind "$behind" \
    '{
      repo: $repo,
      branch: $branch,
      upstream: $upstream,
      compare_ref: $compare_ref,
      remote_url: $remote_url,
      dirty_tracked: $dirty_tracked,
      dirty_untracked: $dirty_untracked,
      ahead: $ahead,
      behind: $behind,
      ssh_probe: $ssh_probe,
      ssh_detail: $ssh_detail,
      lane: $lane,
      reason: $reason
    }'
}

if ! repo_root="$(git -C "$repo_path" rev-parse --show-toplevel 2>/dev/null)"; then
  if $json_mode; then
    repo_root="$repo_path"
    branch=""
    upstream=""
    compare_ref=""
    remote_url=""
    dirty_tracked=0
    dirty_untracked=0
    ahead=0
    behind=0
    ssh_probe="skip"
    ssh_detail="not a git repository"
    lane="blocked"
    reason="not a git repository"
    emit_json
    exit 1
  fi
  hg_die "Not a git repository: $repo_path"
fi

branch="$(git -C "$repo_root" branch --show-current 2>/dev/null || true)"
branch="${branch:-detached}"
upstream="$(git -C "$repo_root" rev-parse --abbrev-ref --symbolic-full-name '@{u}' 2>/dev/null || true)"
compare_ref="$upstream"
if [[ -z "$compare_ref" ]] && git -C "$repo_root" rev-parse --verify origin/main >/dev/null 2>&1; then
  compare_ref="origin/main"
fi

dirty_tracked="$(git -C "$repo_root" status --porcelain --untracked-files=no 2>/dev/null | sed '/^$/d' | wc -l | tr -d ' ')"
dirty_untracked="$(git -C "$repo_root" status --porcelain 2>/dev/null | grep -c '^?? ' || true)"
remote_url="$(git -C "$repo_root" remote get-url origin 2>/dev/null || true)"

ahead=0
behind=0
if [[ -n "$compare_ref" ]] && git -C "$repo_root" rev-parse --verify "$compare_ref" >/dev/null 2>&1; then
  read -r behind ahead < <(git -C "$repo_root" rev-list --left-right --count "${compare_ref}...HEAD")
fi

ssh_probe="skip"
ssh_detail="origin does not use GitHub SSH"
if [[ "$remote_url" == git@github.com:* ]]; then
  if GIT_SSH_COMMAND='ssh -o BatchMode=yes -o StrictHostKeyChecking=accept-new' git ls-remote "$remote_url" HEAD >/dev/null 2>&1; then
    ssh_probe="pass"
    ssh_detail="github ssh auth available"
  else
    ssh_probe="fail"
    ssh_detail="github ssh auth failed"
  fi
fi

lane="in_place"
reason="worktree is clean and compare ref is not ahead"

if [[ -z "$remote_url" ]]; then
  lane="blocked"
  reason="origin remote is not configured"
elif [[ "$dirty_tracked" -gt 0 || "$dirty_untracked" -gt 0 ]]; then
  lane="clean_worktree"
  reason="worktree has uncommitted changes"
elif [[ -z "$compare_ref" ]]; then
  lane="clean_worktree"
  reason="no upstream or origin/main comparison ref is configured"
elif [[ "$behind" -gt 0 ]]; then
  lane="clean_worktree"
  reason="local branch is behind ${compare_ref}"
fi

if $json_mode; then
  emit_json
  exit 0
fi

printf 'repo           %s\n' "$repo_root"
printf 'branch         %s\n' "$branch"
printf 'upstream       %s\n' "${upstream:-none}"
printf 'compare_ref    %s\n' "${compare_ref:-none}"
printf 'remote_url     %s\n' "${remote_url:-none}"
printf 'dirty_tracked  %s\n' "$dirty_tracked"
printf 'dirty_untracked %s\n' "$dirty_untracked"
printf 'ahead          %s\n' "$ahead"
printf 'behind         %s\n' "$behind"
printf 'ssh_probe      %s (%s)\n' "$ssh_probe" "$ssh_detail"
printf 'lane           %s\n' "$lane"
printf 'reason         %s\n' "$reason"
