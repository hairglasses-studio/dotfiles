#!/usr/bin/env bash
# hg-agent-parity-sync.sh — Sync tri-provider agent parity across manifest-managed repos.
# Usage: hg-agent-parity-sync.sh [--dry-run|--check|--write] [--repos=a,b] [--include-compatibility]
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/lib/hg-core.sh"
source "$SCRIPT_DIR/lib/hg-workspace.sh"
source "$SCRIPT_DIR/lib/hg-agent-parity.sh"

MODE="write"
REPO_FILTER=""
INCLUDE_COMPATIBILITY=false
FAILED=0
CREATED=0
UPDATED=0
CURRENT=0

usage() {
  cat <<'EOF'
Usage: hg-agent-parity-sync.sh [--dry-run|--check|--write] [--repos=a,b] [--include-compatibility]

Modes:
  --dry-run  Show parity drift without writing files.
  --check    Exit non-zero when parity drift exists.
  --write    Apply the managed parity contract. This is the default.

This sync covers:
- generated agent compatibility docs
- root .claude/settings.json
- root .gemini/settings.json
- optional Gemini extension scaffolds for hook-heavy repos
- generated skill mirrors
- generated Codex MCP blocks
- manifest-managed workflows
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --dry-run) MODE="dry-run" ;;
    --check) MODE="check" ;;
    --write) MODE="write" ;;
    --repos=*) REPO_FILTER="${1#--repos=}" ;;
    --include-compatibility) INCLUDE_COMPATIBILITY=true ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      hg_die "Unknown argument: $1"
      ;;
  esac
  shift
done

hg_parity_require_tools

repo_names() {
  local selected=()
  hg_workspace_parse_repo_filter "$REPO_FILTER" selected
  if [[ "${#selected[@]}" -gt 0 ]]; then
    printf '%s\n' "${selected[@]}"
    return
  fi

  local jq_filter='.repos[] | select(.baseline_target == true)'
  if ! $INCLUDE_COMPATIBILITY; then
    jq_filter+=' | select((.scope // "active_first_party") != "compatibility_only")'
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

report_current() {
  local repo="$1"
  local label="$2"
  printf "%s%-20s %s (current)%s\n" "$HG_DIM" "$repo" "$label" "$HG_RESET"
  CURRENT=$((CURRENT + 1))
}

report_missing_or_drift() {
  local repo="$1"
  local label="$2"
  printf "%s%-20s %s%s\n" "$HG_YELLOW" "$repo" "$label" "$HG_RESET"
  mark_failure
}

write_text_file() {
  local repo="$1"
  local repo_path="$2"
  local target_rel="$3"
  local expected="$4"
  local label="$5"

  local target="$repo_path/$target_rel"
  if [[ -f "$target" ]] && diff -u <(printf '%s\n' "$expected") "$target" >/dev/null 2>&1; then
    report_current "$repo" "$label"
    return 0
  fi

  case "$MODE" in
    dry-run)
      if [[ -f "$target" ]]; then
        printf "%s%-20s %s (would update)%s\n" "$HG_YELLOW" "$repo" "$label" "$HG_RESET"
      else
        printf "%s%-20s %s (would create)%s\n" "$HG_YELLOW" "$repo" "$label" "$HG_RESET"
      fi
      mark_failure
      return 0
      ;;
    check)
      if [[ -f "$target" ]]; then
        report_missing_or_drift "$repo" "$label (drift)"
      else
        report_missing_or_drift "$repo" "$label (missing)"
      fi
      return 0
      ;;
  esac

  if repo_file_dirty "$repo_path" "$target_rel"; then
    hg_warn "$repo: skipping dirty $label ($target_rel)"
    mark_failure
    return 0
  fi

  mkdir -p "$(dirname "$target")"
  printf '%s\n' "$expected" > "$target"
  if [[ -f "$target" && -s "$target" ]]; then
    if [[ -f "$target" ]]; then
      if git -C "$repo_path" ls-files --error-unmatch "$target_rel" >/dev/null 2>&1; then
        printf "%s%-20s %s (updated)%s\n" "$HG_GREEN" "$repo" "$label" "$HG_RESET"
        UPDATED=$((UPDATED + 1))
      else
        printf "%s%-20s %s (created)%s\n" "$HG_GREEN" "$repo" "$label" "$HG_RESET"
        CREATED=$((CREATED + 1))
      fi
    fi
  fi
}

verify_or_sync_agent_docs() {
  local repo="$1"
  local repo_path="$2"
  if [[ ! -f "$repo_path/AGENTS.md" && ! -f "$repo_path/CLAUDE.md" ]]; then
    return 0
  fi

  local tmpdir
  tmpdir="$(mktemp -d)"
  if [[ -f "$repo_path/AGENTS.md" ]]; then
    cp "$repo_path/AGENTS.md" "$tmpdir/AGENTS.md"
  fi
  if [[ -f "$repo_path/CLAUDE.md" ]]; then
    cp "$repo_path/CLAUDE.md" "$tmpdir/CLAUDE.md"
  fi
  mkdir -p "$tmpdir/.github"
  "$SCRIPT_DIR/hg-agent-docs.sh" "$tmpdir" >/dev/null

  local rel
  for rel in AGENTS.md CLAUDE.md GEMINI.md .github/copilot-instructions.md; do
    local label="agent-doc:${rel}"
    if [[ -f "$tmpdir/$rel" ]]; then
      write_text_file "$repo" "$repo_path" "$rel" "$(cat "$tmpdir/$rel")" "$label"
    fi
  done

  rm -rf "$tmpdir"
}

verify_or_sync_skill_surface() {
  local repo="$1"
  local repo_path="$2"
  [[ -f "$repo_path/.agents/skills/surface.yaml" ]] || return 0

  case "$MODE" in
    dry-run|check)
      if ! "$SCRIPT_DIR/hg-skill-surface-sync.sh" "$repo_path" --check >/dev/null 2>&1; then
        report_missing_or_drift "$repo" "skill-surface"
      else
        report_current "$repo" "skill-surface"
      fi
      ;;
    write)
      "$SCRIPT_DIR/hg-skill-surface-sync.sh" "$repo_path" >/dev/null
      printf "%s%-20s %s (synced)%s\n" "$HG_GREEN" "$repo" "skill-surface" "$HG_RESET"
      UPDATED=$((UPDATED + 1))
      ;;
  esac
}

verify_or_sync_codex_mcp() {
  local repo="$1"
  local repo_path="$2"
  [[ -f "$repo_path/.mcp.json" ]] || return 0
  [[ -f "$repo_path/.codex/mcp-profile-policy.json" ]] || return 0
  [[ -f "$repo_path/.codex/config.toml" ]] || return 0

  case "$MODE" in
    dry-run|check)
      local output
      output="$("$SCRIPT_DIR/hg-codex-mcp-sync.sh" "$repo_path" --dry-run 2>/dev/null || true)"
      if [[ -n "$output" ]]; then
        report_missing_or_drift "$repo" "codex-mcp-block"
      else
        report_current "$repo" "codex-mcp-block"
      fi
      ;;
    write)
      "$SCRIPT_DIR/hg-codex-mcp-sync.sh" "$repo_path" >/dev/null
      printf "%s%-20s %s (synced)%s\n" "$HG_GREEN" "$repo" "codex-mcp-block" "$HG_RESET"
      UPDATED=$((UPDATED + 1))
      ;;
  esac
}

verify_or_sync_workflows() {
  local repo="$1"
  local args=("--repos=$repo")
  case "$MODE" in
    dry-run|check)
      args=("--dry-run" "--ensure-missing" "--repos=$repo")
      ;;
  esac
  $INCLUDE_COMPATIBILITY && args+=("--include-compatibility")

  if [[ "$MODE" == "write" ]]; then
    "$SCRIPT_DIR/hg-workflow-sync.sh" --ensure-missing "${args[@]}" >/dev/null
    printf "%s%-20s %s (synced)%s\n" "$HG_GREEN" "$repo" "workflows" "$HG_RESET"
    UPDATED=$((UPDATED + 1))
    return 0
  fi

  local output
  output="$("$SCRIPT_DIR/hg-workflow-sync.sh" "${args[@]}" 2>/dev/null || true)"
  if grep -Eq 'would create|would update' <<<"$output"; then
    report_missing_or_drift "$repo" "workflows"
  else
    report_current "$repo" "workflows"
  fi
}

verify_or_sync_provider_settings() {
  local repo="$1"
  local repo_path="$2"
  local claude_expected gemini_expected extension_expected extension_rel

  claude_expected="$(hg_parity_render_claude_settings "$repo_path")"
  write_text_file "$repo" "$repo_path" ".claude/settings.json" "$claude_expected" "claude-settings"

  gemini_expected="$(hg_parity_render_gemini_settings "$repo_path")"
  write_text_file "$repo" "$repo_path" ".gemini/settings.json" "$gemini_expected" "gemini-settings"

  if hg_parity_repo_requires_gemini_extension "$repo_path" "$repo"; then
    extension_rel="$(hg_parity_gemini_extension_relpath "$repo")"
    extension_expected="$(hg_parity_render_gemini_extension "$repo_path" "$repo")"
    write_text_file "$repo" "$repo_path" "$extension_rel" "$extension_expected" "gemini-extension"
  else
    report_current "$repo" "gemini-extension (not required)"
  fi
}

hg_info "Tri-provider parity sync — manifest-backed baseline repos"
[[ "$MODE" == "write" ]] || hg_warn "Mode: $MODE"
echo ""

while IFS= read -r repo; do
  [[ -n "$repo" ]] || continue
  repo_path="$(hg_workspace_repo_path "$repo")"
  [[ -d "$repo_path/.git" ]] || {
    hg_warn "$repo: repo path missing"
    mark_failure
    continue
  }

  verify_or_sync_agent_docs "$repo" "$repo_path"
  verify_or_sync_provider_settings "$repo" "$repo_path"
  verify_or_sync_skill_surface "$repo" "$repo_path"
  verify_or_sync_codex_mcp "$repo" "$repo_path"
  verify_or_sync_workflows "$repo"
done < <(repo_names)

echo ""
if [[ "$MODE" == "write" ]]; then
  hg_ok "Parity sync complete — ${CREATED} created, ${UPDATED} updated, ${CURRENT} current"
else
  hg_info "Parity audit complete — ${CURRENT} current"
fi

exit "$FAILED"
