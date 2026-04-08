#!/usr/bin/env bash
# hg-skill-check.sh — Verify skill surface consistency across repos
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/lib/hg-core.sh"
source "$SCRIPT_DIR/lib/hg-workspace.sh"

hg_require jq find

ROOT="${1:-$HOME/hairglasses-studio}"
[[ -d "$ROOT" ]] || hg_die "Root directory not found: $ROOT"

SYNC_SCRIPT="$SCRIPT_DIR/hg-skill-surface-sync.sh"
WORKSPACE_SYNC_SCRIPT="$SCRIPT_DIR/hg-workspace-global-sync.sh"

pass=0
warn=0
fail=0

report_issue() {
  local severity="$1"
  local message="$2"
  case "$severity" in
    fail)
      hg_warn "$message"
      fail=$((fail + 1))
      ;;
    warn)
      hg_warn "$message"
      warn=$((warn + 1))
      ;;
    *)
      hg_die "Unknown severity: $severity"
      ;;
  esac
}

repo_issue_severity() {
  local repo="$1"
  local scope
  scope="$(hg_workspace_repo_field "$repo" "scope" "")"
  case "$scope" in
    active_operator|active_first_party)
      printf 'fail\n'
      ;;
    *)
      printf 'warn\n'
      ;;
  esac
}

repo_has_dirty_tracked_changes() {
  local repo="$1"
  git -C "$repo" status --short --untracked-files=no 2>/dev/null | grep -q .
}

# Check repo-level skill surfaces
for repo in "$ROOT"/*/; do
  [[ -e "$repo/.git" ]] || continue
  name="$(basename "$repo")"

  has_surface=0
  surface_json_compatible=0
  has_legacy=0
  has_generated_claude_skills=0
  has_generated_plugin_skills=0
  has_mcp_json=0
  has_mcp_policy=0
  issue_severity="warn"

  [[ -f "$repo/.agents/skills/surface.yaml" ]] && has_surface=1
  if [[ "$has_surface" -eq 1 ]] && jq -e '.version == 1 and (.skills | type == "array")' "$repo/.agents/skills/surface.yaml" >/dev/null 2>&1; then
    surface_json_compatible=1
  fi
  [[ -d "$repo/.claude/commands" ]] && has_legacy=1
  [[ -d "$repo/.claude/skills" ]] && has_generated_claude_skills=1
  if find "$repo/plugins" -path '*/skills/*/SKILL.md' -print -quit 2>/dev/null | grep -q .; then
    has_generated_plugin_skills=1
  fi
  [[ -f "$repo/.mcp.json" ]] && has_mcp_json=1
  [[ -f "$repo/.codex/mcp-profile-policy.json" ]] && has_mcp_policy=1
  if hg_workspace_repo_exists "$name"; then
    issue_severity="$(repo_issue_severity "$name")"
  fi

  if [[ "$has_surface" -eq 0 && ( "$has_generated_claude_skills" -eq 1 || "$has_generated_plugin_skills" -eq 1 ) ]]; then
    report_issue "$issue_severity" "$name: managed generated skills exist without canonical .agents/skills/surface.yaml"
  fi

  # Check surface sync consistency
  if [[ "$has_surface" -eq 1 && -x "$SYNC_SCRIPT" ]]; then
    if [[ "$surface_json_compatible" -eq 0 ]]; then
      report_issue "$issue_severity" "$name: non-portable or YAML-only skill manifest blocks managed mirror sync"
    elif "$SYNC_SCRIPT" "$repo" --check >/dev/null 2>&1; then
      pass=$((pass + 1))
    else
      mismatch_severity="$issue_severity"
      if repo_has_dirty_tracked_changes "$repo"; then
        mismatch_severity="warn"
      fi
      report_issue "$mismatch_severity" "$name: skill surface out of sync (run hg-skill-surface-sync.sh)"
    fi
  fi

  # Check for orphan legacy commands alongside surface.yaml
  if [[ "$has_surface" -eq 1 && "$has_legacy" -eq 1 ]]; then
    legacy_count=$(find "$repo/.claude/commands" -name '*.md' 2>/dev/null | wc -l)
    if [[ "$legacy_count" -gt 0 ]]; then
      report_issue "$issue_severity" "$name: has $legacy_count orphan .claude/commands/ alongside surface.yaml"
    fi
  fi

  # Check MCP policy completeness (skip example-only .mcp.json)
  if [[ "$has_mcp_json" -eq 1 && "$has_mcp_policy" -eq 0 ]]; then
    # Check if .mcp.json has real servers (not just _example_ prefixed)
    real_servers=$(jq -r '.mcpServers // {} | keys[]' "$repo/.mcp.json" 2>/dev/null | grep -cv '^_' || true)
    if [[ "$real_servers" -gt 0 ]]; then
      hg_warn "$name: has .mcp.json ($real_servers servers) but no .codex/mcp-profile-policy.json"
      warn=$((warn + 1))
    fi
  fi
done

# Check global user skill parity
claude_count=$(find "$HOME/.claude/commands" -name '*.md' 2>/dev/null | wc -l)
claude_skill_count=$(find "$HOME/.claude/skills" -mindepth 1 -maxdepth 1 -type d 2>/dev/null | wc -l)
agents_count=$(find "$HOME/.agents/skills" -mindepth 1 -maxdepth 1 -type d 2>/dev/null | wc -l)
codex_count=$(find "$HOME/.codex/skills" -mindepth 1 -maxdepth 1 -type d 2>/dev/null | wc -l)

if [[ "$claude_count" -gt 0 || "$claude_skill_count" -gt 0 ]]; then
  hg_info "Global user skills: $claude_count commands / $claude_skill_count Claude skills / $agents_count Agents / $codex_count Codex"
fi

# Check managed workspace-global source contracts first; that verifies
# downstream repo-local skill and MCP metadata without depending on personal
# home-directory overlays.
if [[ -x "$WORKSPACE_SYNC_SCRIPT" ]]; then
  if "$WORKSPACE_SYNC_SCRIPT" --root "$ROOT" --source-check >/dev/null 2>&1; then
    pass=$((pass + 1))
    hg_ok "Workspace global source contract: up to date"
  else
    hg_warn "Workspace global source contract: out of date"
    warn=$((warn + 1))
  fi
  if "$WORKSPACE_SYNC_SCRIPT" --root "$ROOT" --check >/dev/null 2>&1; then
    pass=$((pass + 1))
    hg_ok "Workspace global home overlays: up to date"
  else
    hg_warn "Workspace global home overlays: out of date (run hg-workspace-global-sync.sh)"
    warn=$((warn + 1))
  fi
elif [[ -x "$SCRIPT_DIR/hg-global-skill-sync.sh" ]]; then
  if "$SCRIPT_DIR/hg-global-skill-sync.sh" --check >/dev/null 2>&1; then
    pass=$((pass + 1))
    hg_ok "Global skill sync: up to date"
  else
    hg_warn "Global skill sync: out of date (run hg-global-skill-sync.sh)"
    warn=$((warn + 1))
  fi
fi

echo ""
hg_info "Skill check: $pass pass, $warn warnings, $fail failures"

if [[ "$fail" -gt 0 ]]; then
  exit 1
fi
