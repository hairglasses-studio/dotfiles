#!/usr/bin/env bash
# hg-repo-profile-sync.sh — Sync profile-driven shared assets into managed repos.
# Usage: hg-repo-profile-sync.sh [verify|sync|ensure-missing] [--repos=a,b] [--allow-dirty] [--include-compatibility] [--include-deprecated]
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/lib/hg-core.sh"
source "$SCRIPT_DIR/lib/hg-workspace.sh"

MODE="sync"
REPO_FILTER=""
INCLUDE_COMPATIBILITY=false
INCLUDE_DEPRECATED=false
ALLOW_DIRTY=false
FAILED=0

usage() {
  cat <<'EOF'
Usage: hg-repo-profile-sync.sh [verify|sync|ensure-missing] [--repos=a,b] [--allow-dirty] [--include-compatibility] [--include-deprecated]

Modes:
  verify          Check manifest-managed repo assets for drift without writing.
  sync            Update managed repo assets to match manifest-driven standards.
  ensure-missing  Only create missing baseline assets; do not overwrite clean files.
EOF
}

if [[ $# -gt 0 ]]; then
  case "$1" in
    verify|sync|ensure-missing)
      MODE="$1"
      shift
      ;;
  esac
fi

while [[ $# -gt 0 ]]; do
  case "$1" in
    --repos=*)
      REPO_FILTER="${1#--repos=}"
      shift
      ;;
    --allow-dirty)
      ALLOW_DIRTY=true
      shift
      ;;
    --include-compatibility)
      INCLUDE_COMPATIBILITY=true
      shift
      ;;
    --include-deprecated)
      INCLUDE_DEPRECATED=true
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

repo_names() {
  local selected=()
  hg_workspace_parse_repo_filter "$REPO_FILTER" selected
  if [[ "${#selected[@]}" -gt 0 ]]; then
    printf '%s\n' "${selected[@]}"
    return
  fi

  local jq_filter='.repos[] | select(.baseline_target == true)'
  if ! $INCLUDE_COMPATIBILITY; then
    jq_filter+=' | select((.lifecycle // "canonical") != "compatibility")'
  fi
  if ! $INCLUDE_DEPRECATED; then
    jq_filter+=' | select((.lifecycle // "canonical") != "deprecated")'
  fi
  hg_workspace_repo_names "$jq_filter"
}

repo_file_dirty() {
  local repo_path="$1"
  local target_rel="$2"
  [[ -n "$(git -C "$repo_path" status --porcelain --untracked-files=all -- "$target_rel" 2>/dev/null)" ]]
}

mark_failure() {
  FAILED=1
}

repo_dirty_paths() {
  local repo_path="$1"
  shift

  local dirty=()
  local target_rel
  for target_rel in "$@"; do
    if repo_file_dirty "$repo_path" "$target_rel"; then
      dirty+=("$target_rel")
    fi
  done

  local joined=""
  local path
  for path in "${dirty[@]}"; do
    if [[ -n "$joined" ]]; then
      joined+=", "
    fi
    joined+="$path"
  done
  printf '%s\n' "$joined"
}

skip_dirty_group() {
  local repo="$1"
  local repo_path="$2"
  local label="$3"
  shift 3

  if $ALLOW_DIRTY; then
    return 1
  fi

  local dirty_paths
  dirty_paths="$(repo_dirty_paths "$repo_path" "$@")"
  [[ -n "$dirty_paths" ]] || return 1

  hg_warn "$repo: skipping dirty $label ($dirty_paths)"
  mark_failure
  return 0
}

render_agent_docs_staging() {
  local repo_path="$1"
  local outdir="$2"

  if [[ -f "$repo_path/AGENTS.md" ]]; then
    cp "$repo_path/AGENTS.md" "$outdir/AGENTS.md"
  fi
  if [[ -f "$repo_path/CLAUDE.md" ]]; then
    cp "$repo_path/CLAUDE.md" "$outdir/CLAUDE.md"
  fi

  mkdir -p "$outdir/.github"
  "$SCRIPT_DIR/hg-agent-docs.sh" "$outdir" >/dev/null
}

sync_standard_file() {
  local repo="$1"
  local repo_path="$2"
  local source_file="$3"
  local target_rel="$4"
  local label="$5"

  local target="$repo_path/$target_rel"
  if [[ ! -f "$source_file" ]]; then
    hg_die "$repo: missing source template for $label: $source_file"
  fi

  if [[ ! -f "$target" ]]; then
    case "$MODE" in
      verify)
        hg_warn "$repo: missing $label ($target_rel)"
        mark_failure
        ;;
      sync|ensure-missing)
        mkdir -p "$(dirname "$target")"
        cp -f "$source_file" "$target"
        hg_ok "$repo: created $label"
        ;;
    esac
    return
  fi

  if cmp -s "$source_file" "$target"; then
    return
  fi

  case "$MODE" in
    verify)
      hg_warn "$repo: drift in $label ($target_rel)"
      mark_failure
      ;;
    ensure-missing)
      ;;
    sync)
      if ! $ALLOW_DIRTY && repo_file_dirty "$repo_path" "$target_rel"; then
        hg_warn "$repo: skipping dirty $label ($target_rel)"
        mark_failure
        return
      fi
      cp -f "$source_file" "$target"
      hg_ok "$repo: synced $label"
      ;;
  esac
}

sync_gemini_settings() {
  local repo="$1"
  local repo_path="$2"
  local settings_rel=".gemini/settings.json"
  local legacy_rel=".gemini/config.yaml"
  local sync_args=("$repo_path")

  if $ALLOW_DIRTY; then
    sync_args+=("--allow-dirty")
  fi

  case "$MODE" in
    verify)
      if ! "$SCRIPT_DIR/hg-gemini-settings-sync.sh" "$repo_path" --check >/dev/null; then
        hg_warn "$repo: Gemini settings out of sync"
        mark_failure
      fi
      ;;
    ensure-missing)
      if [[ ! -f "$repo_path/$settings_rel" ]]; then
        "$SCRIPT_DIR/hg-gemini-settings-sync.sh" "${sync_args[@]}" >/dev/null
        hg_ok "$repo: created Gemini settings"
      fi
      ;;
    sync)
      if skip_dirty_group "$repo" "$repo_path" "Gemini settings" "$settings_rel" "$legacy_rel"; then
        return
      fi
      "$SCRIPT_DIR/hg-gemini-settings-sync.sh" "${sync_args[@]}" >/dev/null
      hg_ok "$repo: synced Gemini settings"
      ;;
  esac
}

verify_agent_docs() {
  local repo="$1"
  local repo_path="$2"

  if [[ ! -f "$repo_path/AGENTS.md" && ! -f "$repo_path/CLAUDE.md" ]]; then
    return
  fi

  local tmpdir
  tmpdir="$(mktemp -d)"
  render_agent_docs_staging "$repo_path" "$tmpdir"

  local rel
  for rel in AGENTS.md CLAUDE.md GEMINI.md .github/copilot-instructions.md; do
    if [[ ! -f "$repo_path/$rel" ]]; then
      hg_warn "$repo: missing generated agent doc $rel"
      mark_failure
      continue
    fi
    if ! cmp -s "$tmpdir/$rel" "$repo_path/$rel"; then
      hg_warn "$repo: drift in generated agent doc $rel"
      mark_failure
    fi
  done

  rm -rf "$tmpdir"
}

sync_agent_docs() {
  local repo="$1"
  local repo_path="$2"

  if [[ ! -f "$repo_path/AGENTS.md" && ! -f "$repo_path/CLAUDE.md" ]]; then
    return
  fi

  if [[ "$MODE" == "verify" ]]; then
    verify_agent_docs "$repo" "$repo_path"
    return
  fi

  if skip_dirty_group \
    "$repo" \
    "$repo_path" \
    "generated agent docs" \
    AGENTS.md \
    CLAUDE.md \
    GEMINI.md \
    .github/copilot-instructions.md; then
    return
  fi

  if [[ "$MODE" == "ensure-missing" ]]; then
    local missing=0
    local rel
    for rel in AGENTS.md CLAUDE.md GEMINI.md .github/copilot-instructions.md; do
      if [[ ! -f "$repo_path/$rel" ]]; then
        missing=1
        break
      fi
    done

    [[ "$missing" -eq 1 ]] || return

    local tmpdir
    tmpdir="$(mktemp -d)"
    render_agent_docs_staging "$repo_path" "$tmpdir"
    for rel in AGENTS.md CLAUDE.md GEMINI.md .github/copilot-instructions.md; do
      if [[ -f "$repo_path/$rel" ]]; then
        continue
      fi
      mkdir -p "$(dirname "$repo_path/$rel")"
      cp -f "$tmpdir/$rel" "$repo_path/$rel"
    done
    rm -rf "$tmpdir"
    hg_ok "$repo: created missing agent docs"
    return
  fi

  "$SCRIPT_DIR/hg-agent-docs.sh" "$repo_path" >/dev/null
  hg_ok "$repo: synced agent docs"
}

sync_provider_role_surface() {
  local repo="$1"
  local repo_path="$2"
  [[ -d "$repo_path/.codex/agents" ]] || return 0

  case "$MODE" in
    verify)
      if ! "$SCRIPT_DIR/hg-provider-role-sync.sh" "$repo_path" --check >/dev/null; then
        hg_warn "$repo: drift in generated provider role surfaces"
        mark_failure
      fi
      ;;
    sync|ensure-missing)
      if skip_dirty_group \
        "$repo" \
        "$repo_path" \
        "generated provider role surfaces" \
        ".claude/agents" \
        ".gemini/commands"; then
        return
      fi
      "$SCRIPT_DIR/hg-provider-role-sync.sh" "$repo_path" >/dev/null
      hg_ok "$repo: synced provider role surfaces"
      ;;
  esac
}

skill_manifest_supports_sync() {
  local manifest="$1"
  jq -e '.version == 1 and (.skills | type == "array")' "$manifest" >/dev/null 2>&1
}

skill_surface_plugin_root() {
  local repo="$1"
  local manifest="$2"
  jq -r --arg repo "$repo" '.plugin_root // $repo' "$manifest"
}

sync_skill_surface() {
  local repo="$1"
  local repo_path="$2"
  local manifest="$repo_path/.agents/skills/surface.yaml"
  [[ -f "$manifest" ]] || return 0

  if ! skill_manifest_supports_sync "$manifest"; then
    hg_warn "$repo: skipping YAML-only skill surface manifest"
    return
  fi

  local plugin_root
  plugin_root="$(skill_surface_plugin_root "$repo" "$manifest")"

  case "$MODE" in
    verify)
      if ! "$SCRIPT_DIR/hg-skill-surface-sync.sh" "$repo_path" --check >/dev/null; then
        hg_warn "$repo: drift in generated skill surface"
        mark_failure
      fi
      ;;
    sync|ensure-missing)
      if skip_dirty_group \
        "$repo" \
        "$repo_path" \
        "generated skill surface" \
        ".claude/skills" \
        "plugins/$plugin_root/skills"; then
        return
      fi
      "$SCRIPT_DIR/hg-skill-surface-sync.sh" "$repo_path" >/dev/null
      hg_ok "$repo: synced skill surface"
      ;;
  esac
}

sync_codex_mcp_block() {
  local repo="$1"
  local repo_path="$2"
  [[ -f "$repo_path/.mcp.json" ]] || return 0
  [[ -f "$repo_path/.codex/mcp-profile-policy.json" ]] || return 0
  [[ -f "$repo_path/.codex/config.toml" ]] || return 0

  case "$MODE" in
    verify)
      local diff_output
      diff_output="$("$SCRIPT_DIR/hg-codex-mcp-sync.sh" "$repo_path" --dry-run 2>/dev/null || true)"
      if [[ -n "$diff_output" ]]; then
        hg_warn "$repo: generated Codex MCP block is out of sync"
        mark_failure
      fi
      ;;
    sync|ensure-missing)
      if skip_dirty_group "$repo" "$repo_path" "generated Codex MCP block" ".codex/config.toml"; then
        return
      fi
      "$SCRIPT_DIR/hg-codex-mcp-sync.sh" "$repo_path" >/dev/null
      hg_ok "$repo: synced generated Codex MCP block"
      ;;
  esac
}

run_workflow_sync() {
  local repo="$1"
  local args=("--repos=$repo")
  case "$MODE" in
    verify) args=("--dry-run" "--ensure-missing" "--repos=$repo") ;;
    ensure-missing) args=("--ensure-missing" "--repos=$repo") ;;
  esac
  if $ALLOW_DIRTY; then
    args+=("--allow-dirty")
  fi

  if [[ "$MODE" == "verify" ]]; then
    local output
    output="$("$SCRIPT_DIR/hg-workflow-sync.sh" "${args[@]}" 2>/dev/null || true)"
    if grep -Eq 'would create|would update' <<<"$output"; then
      hg_warn "$repo: workflow drift detected"
      mark_failure
    fi
    return
  fi

  "$SCRIPT_DIR/hg-workflow-sync.sh" "${args[@]}" >/dev/null
}

while IFS= read -r repo; do
  [[ -n "$repo" ]] || continue
  repo_path="$(hg_workspace_repo_path "$repo")"
  [[ -d "$repo_path/.git" ]] || {
    hg_warn "$repo: repo path missing"
    mark_failure
    continue
  }

  run_workflow_sync "$repo"

  sync_standard_file \
    "$repo" \
    "$repo_path" \
    "$SCRIPT_DIR/../templates/codex-config.standard.toml" \
    ".codex/config.toml" \
    "Codex config"

  sync_gemini_settings "$repo" "$repo_path"

  sync_agent_docs "$repo" "$repo_path"
  sync_provider_role_surface "$repo" "$repo_path"
  sync_skill_surface "$repo" "$repo_path"
  sync_codex_mcp_block "$repo" "$repo_path"
done < <(repo_names)

if [[ "$MODE" == "verify" && "$FAILED" -ne 0 ]]; then
  exit 1
fi

hg_ok "Profile sync complete ($MODE)"
