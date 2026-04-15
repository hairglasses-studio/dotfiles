#!/usr/bin/env bash
# hg-agent-launch.sh — Shared launch and managed worktree helpers for CLI agents.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/hg-core.sh"

HG_AGENT_USER_HOME="${HG_AGENT_USER_HOME:-${HOME:-/home/hg}}"
HG_AGENT_MANAGED_ROOT="${HG_AGENT_MANAGED_ROOT:-${HG_AGENT_USER_HOME}/.codex/worktrees}"
HG_AGENT_STATE_DIR="${HG_AGENT_MANAGED_ROOT}/.state"
HG_AGENT_CODEX_HOME_ROOT="${HG_AGENT_CODEX_HOME_ROOT:-${HG_AGENT_USER_HOME}/.codex}"
HG_AGENT_BREAK_GLASS_ENV="${HG_AGENT_BREAK_GLASS_ENV:-CODEX_GLOBAL_WRAPPER_DISABLE}"

# Force Claude-native subagents to Sonnet by default so Opus main-thread cost
# does not fan out to every file-read / git-peek. Highest-priority model
# resolution per https://code.claude.com/docs/en/sub-agents. User can override
# by exporting CLAUDE_CODE_SUBAGENT_MODEL in the launching shell.
export CLAUDE_CODE_SUBAGENT_MODEL="${CLAUDE_CODE_SUBAGENT_MODEL:-claude-sonnet-4-6}"

hg_agent_require_tools() {
  hg_require git jq mktemp tar awk sed tr
}

hg_agent_realpath() {
  readlink -f "$1"
}

hg_agent_slugify() {
  local raw="${1:-repo}"
  printf '%s' "$raw" \
    | tr '[:upper:]' '[:lower:]' \
    | sed -E 's/[^a-z0-9._-]+/-/g; s/^-+//; s/-+$//; s/-{2,}/-/g'
}

hg_agent_timestamp_id() {
  printf '%s-%s' "$(date -u +%Y%m%dT%H%M%SZ)" "$(tr -dc 'a-f0-9' </dev/urandom | head -c 6)"
}

hg_agent_repo_root() {
  git rev-parse --show-toplevel 2>/dev/null || true
}

hg_agent_repo_relpath() {
  local repo_root="$1"
  local cwd="$2"
  local rel="."
  if [[ "$cwd" == "$repo_root" ]]; then
    printf '.\n'
    return
  fi
  if [[ "$cwd" == "$repo_root/"* ]]; then
    rel="${cwd#"$repo_root"/}"
  fi
  printf '%s\n' "$rel"
}

hg_agent_toml_set_top_level_string() {
  local path="$1"
  local key="$2"
  local value="$3"
  local tmp
  tmp="$(mktemp)"
  awk -v key="$key" -v value="$value" '
    BEGIN {
      done = 0
      inserted_before_section = 0
    }
    {
      if (!inserted_before_section && $0 ~ /^\[/) {
        if (!done) {
          printf "%s = \"%s\"\n", key, value
          done = 1
        }
        inserted_before_section = 1
      }

      if ($0 ~ ("^" key "[[:space:]]*=")) {
        if (!done) {
          printf "%s = \"%s\"\n", key, value
          done = 1
        }
        next
      }

      print
    }
    END {
      if (!done) {
        printf "%s = \"%s\"\n", key, value
      }
    }
  ' "$path" >"$tmp"
  mv "$tmp" "$path"
}

hg_agent_align_codex_config_file() {
  local path="$1"
  mkdir -p "$(dirname "$path")"
  if [[ ! -f "$path" ]]; then
    if [[ -f "$HG_DOTFILES/codex/config.toml" ]]; then
      cp "$HG_DOTFILES/codex/config.toml" "$path"
    else
      printf 'approval_policy = "never"\nsandbox_mode = "danger-full-access"\n' >"$path"
    fi
  fi
  hg_agent_toml_set_top_level_string "$path" "approval_policy" "never"
  hg_agent_toml_set_top_level_string "$path" "sandbox_mode" "danger-full-access"
}

hg_agent_align_codex_live_configs() {
  hg_agent_align_codex_config_file "${HG_AGENT_CODEX_HOME_ROOT}/config.toml"
  if [[ "${HG_AGENT_CODEX_HOME_ROOT}/config.toml" != "${HG_AGENT_USER_HOME}/.codex/config.toml" ]]; then
    hg_agent_align_codex_config_file "${HG_AGENT_USER_HOME}/.codex/config.toml"
  fi
}

hg_agent_find_state_file_for_worktree() {
  local worktree_path="$1"
  local state_file
  [[ -d "$HG_AGENT_STATE_DIR" ]] || return 1
  while IFS= read -r -d '' state_file; do
    if jq -e --arg path "$worktree_path" '.worktree_path == $path' "$state_file" >/dev/null 2>&1; then
      printf '%s\n' "$state_file"
      return 0
    fi
  done < <(find "$HG_AGENT_STATE_DIR" -maxdepth 1 -type f -name '*.json' -print0 2>/dev/null)
  return 1
}

hg_agent_inside_managed_worktree() {
  local repo_root="$1"
  [[ "$repo_root" == "$HG_AGENT_MANAGED_ROOT/"* ]] || return 1
  hg_agent_find_state_file_for_worktree "$repo_root" >/dev/null 2>&1
}

hg_agent_copy_untracked_files() {
  local repo_root="$1"
  local target_root="$2"
  local list_file="$3"
  git -C "$repo_root" ls-files --others --exclude-standard -z >"$list_file"
  if [[ ! -s "$list_file" ]]; then
    return 0
  fi
  (
    cd "$repo_root"
    tar --null --files-from="$list_file" -cf - \
      | tar -xf - -C "$target_root"
  )
}

hg_agent_write_metadata() {
  local state_file="$1"
  local provider="$2"
  local source_repo="$3"
  local worktree_path="$4"
  local branch_name="$5"
  local source_head="$6"
  local relative_cwd="$7"
  mkdir -p "$(dirname "$state_file")"
  jq -n \
    --arg provider "$provider" \
    --arg source_repo "$source_repo" \
    --arg worktree_path "$worktree_path" \
    --arg branch_name "$branch_name" \
    --arg source_head "$source_head" \
    --arg relative_cwd "$relative_cwd" \
    --arg created_at "$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
    --arg launched_by "${SUDO_USER:-${USER:-root}}" \
    '{
      provider: $provider,
      source_repo: $source_repo,
      worktree_path: $worktree_path,
      branch_name: $branch_name,
      source_head: $source_head,
      relative_cwd: $relative_cwd,
      created_at: $created_at,
      launched_by: $launched_by
    }' >"$state_file"
}

hg_agent_exec_provider() {
  local raw_bin="$1"
  shift
  exec "$raw_bin" "$@"
}

hg_agent_launch_main() {
  local provider="$1"
  local raw_bin="$2"
  local launcher_path="$3"
  shift 3

  local -a fixed_args=()
  while [[ $# -gt 0 && "$1" != "--" ]]; do
    fixed_args+=("$1")
    shift
  done
  if [[ "${1:-}" == "--" ]]; then
    shift
  fi
  local -a user_args=("$@")

  hg_agent_require_tools

  if [[ "${!HG_AGENT_BREAK_GLASS_ENV:-0}" == "1" ]]; then
    hg_agent_exec_provider "$raw_bin" "${user_args[@]}"
  fi

  hg_agent_align_codex_live_configs

  local cwd repo_root relative_cwd
  cwd="$(pwd -P)"
  repo_root="$(hg_agent_repo_root)"

  if [[ -z "$repo_root" ]]; then
    hg_agent_exec_provider "$raw_bin" "${fixed_args[@]}" "${user_args[@]}"
  fi

  repo_root="$(cd "$repo_root" && pwd -P)"
  relative_cwd="$(hg_agent_repo_relpath "$repo_root" "$cwd")"

  if hg_agent_inside_managed_worktree "$repo_root"; then
    if [[ "$relative_cwd" != "." ]]; then
      cd "$repo_root/$relative_cwd"
    else
      cd "$repo_root"
    fi
    hg_agent_exec_provider "$raw_bin" "${fixed_args[@]}" "${user_args[@]}"
  fi

  git -C "$repo_root" rev-parse --verify HEAD >/dev/null 2>&1 \
    || hg_die "Cannot create managed worktree without a valid HEAD: $repo_root"

  mkdir -p "$HG_AGENT_STATE_DIR"

  local repo_slug worktree_id branch_name worktree_parent worktree_path state_file
  repo_slug="$(hg_agent_slugify "$(basename "$repo_root")")"
  worktree_id="$(hg_agent_timestamp_id)"
  branch_name="codex/wt/${repo_slug}/${worktree_id}"
  worktree_parent="${HG_AGENT_MANAGED_ROOT}/${repo_slug}"
  worktree_path="${worktree_parent}/${worktree_id}"
  state_file="${HG_AGENT_STATE_DIR}/${repo_slug}-${worktree_id}.json"

  local patch_file untracked_list created_worktree=0 created_branch=0 cleanup_needed=0
  patch_file="$(mktemp)"
  untracked_list="$(mktemp)"

  cleanup() {
    if [[ "$cleanup_needed" -eq 0 ]]; then
      rm -f "$patch_file" "$untracked_list"
      return
    fi

    rm -f "$state_file"
    if [[ "$created_worktree" -eq 1 ]] && [[ -d "$worktree_path" ]]; then
      git -C "$repo_root" worktree remove --force "$worktree_path" >/dev/null 2>&1 || rm -rf "$worktree_path"
    fi
    if [[ "$created_branch" -eq 1 ]]; then
      git -C "$repo_root" branch -D "$branch_name" >/dev/null 2>&1 || true
    fi
    rm -f "$patch_file" "$untracked_list"
  }
  trap cleanup EXIT INT TERM

  git -C "$repo_root" diff --binary HEAD >"$patch_file"
  mkdir -p "$worktree_parent"
  git -C "$repo_root" worktree add -b "$branch_name" "$worktree_path" HEAD >/dev/null
  created_worktree=1
  created_branch=1
  cleanup_needed=1

  if [[ -s "$patch_file" ]]; then
    git -C "$worktree_path" apply --binary "$patch_file"
  fi

  hg_agent_copy_untracked_files "$repo_root" "$worktree_path" "$untracked_list"
  hg_agent_write_metadata \
    "$state_file" \
    "$provider" \
    "$repo_root" \
    "$worktree_path" \
    "$branch_name" \
    "$(git -C "$repo_root" rev-parse HEAD)" \
    "$relative_cwd"

  if [[ "$relative_cwd" != "." ]]; then
    mkdir -p "$worktree_path/$relative_cwd"
    cd "$worktree_path/$relative_cwd"
  else
    cd "$worktree_path"
  fi

  cleanup_needed=0
  rm -f "$patch_file" "$untracked_list"
  hg_agent_exec_provider "$raw_bin" "${fixed_args[@]}" "${user_args[@]}"
}
