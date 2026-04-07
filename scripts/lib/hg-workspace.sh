#!/usr/bin/env bash
# hg-workspace.sh — Shared helpers for manifest-backed repo selection.

# Do not reuse SCRIPT_DIR here; callers often rely on their own copy after sourcing.
HG_WORKSPACE_LIB_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$HG_WORKSPACE_LIB_DIR/hg-core.sh"

hg_workspace_manifest() {
  printf '%s\n' "$HG_STUDIO_ROOT/workspace/manifest.json"
}

hg_workspace_require_manifest() {
  hg_require jq
  local manifest
  manifest="$(hg_workspace_manifest)"
  [[ -f "$manifest" ]] || hg_die "workspace manifest not found: $manifest"
}

hg_workspace_repo_names() {
  local jq_filter="${1:-.repos[]}"
  hg_workspace_require_manifest
  jq -r "$jq_filter | .name" "$(hg_workspace_manifest)"
}

hg_workspace_default_repo_names() {
  hg_workspace_repo_names '.repos[] | select((.lifecycle // "canonical") != "compatibility" and (.lifecycle // "canonical") != "deprecated")'
}

hg_workspace_repo_exists() {
  local repo="$1"
  hg_workspace_require_manifest
  jq -e --arg repo "$repo" '.repos[] | select(.name == $repo)' "$(hg_workspace_manifest)" >/dev/null
}

hg_workspace_repo_path() {
  local repo="$1"
  printf '%s\n' "$HG_STUDIO_ROOT/$repo"
}

hg_workspace_repo_field() {
  local repo="$1"
  local field="$2"
  local default_value="${3:-}"
  hg_workspace_require_manifest
  jq -r \
    --arg repo "$repo" \
    --arg field "$field" \
    --arg def "$default_value" \
    'first(.repos[] | select(.name == $repo) | .[$field]) // $def' \
    "$(hg_workspace_manifest)"
}

hg_workspace_repo_bool() {
  local repo="$1"
  local field="$2"
  [[ "$(hg_workspace_repo_field "$repo" "$field" "false")" == "true" ]]
}

hg_workspace_repo_json() {
  local repo="$1"
  hg_workspace_require_manifest
  jq -c --arg repo "$repo" 'first(.repos[] | select(.name == $repo))' "$(hg_workspace_manifest)"
}

hg_workspace_parse_repo_filter() {
  local filter="$1"
  local -n out_ref="$2"
  out_ref=()
  if [[ -z "$filter" ]]; then
    return 0
  fi

  local entry
  IFS=',' read -r -a out_ref <<<"$filter"
  for entry in "${out_ref[@]}"; do
    [[ -n "$entry" ]] || hg_die "empty repo name in --repos filter"
    hg_workspace_repo_exists "$entry" || hg_die "repo not found in manifest: $entry"
  done
}
