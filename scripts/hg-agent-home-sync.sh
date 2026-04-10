#!/usr/bin/env bash
# hg-agent-home-sync.sh — Keep user and root provider homes aligned with the managed launcher contract.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$(readlink -f "${BASH_SOURCE[0]}")")" && pwd)"
source "$SCRIPT_DIR/lib/hg-core.sh"
source "$SCRIPT_DIR/lib/hg-agent-launch.sh"

MODE="write"
FAILED=0
USER_HOME_DIR="${HG_AGENT_USER_HOME:-/home/hg}"
ROOT_HOME_DIR="${HG_AGENT_ROOT_HOME:-/root}"
WORKSPACE_ROOT="${HG_AGENT_WORKSPACE_ROOT:-${USER_HOME_DIR}/hairglasses-studio}"
CLAUDE_TEMPLATE="$HG_DOTFILES/templates/provider-home/CLAUDE.md"
GEMINI_TEMPLATE="$HG_DOTFILES/templates/provider-home/GEMINI.md"

usage() {
  cat <<'EOF'
Usage: hg-agent-home-sync.sh [--dry-run|--check|--write]

Syncs provider home docs, shared skill mirrors, workspace-global overlays, and
launcher-enforced defaults for both /home/hg and /root.
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --dry-run) MODE="dry-run" ;;
    --check) MODE="check" ;;
    --write) MODE="write" ;;
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

if [[ "${EUID:-$(id -u)}" -ne 0 ]]; then
  exec sudo -H "$(readlink -f "$0")" "--${MODE}"
fi

hg_require jq rsync

# Seed the provider home docs from the repo template when missing, but let
# hg-workspace-global-sync.sh remain the authoritative renderer for the final
# file because it appends the managed workspace context block.
seed_file_from_source_if_missing() {
  local source="$1"
  local target="$2"
  local label="$3"

  mkdir -p "$(dirname "$target")"
  if [[ -f "$target" ]]; then
    return 0
  fi

  case "$MODE" in
    dry-run)
      hg_warn "Would seed $label: $target"
      ;;
    check)
      hg_warn "Missing $label: $target"
      FAILED=1
      return 0
      ;;
    write)
      cp "$source" "$target"
      hg_ok "Seeded $label: $target"
      ;;
  esac
}

sync_directory_tree() {
  local source_dir="$1"
  local target_dir="$2"
  local label="$3"
  local -a cmp_args=(-rlcni --delete)
  local -a sync_args=(-rlptD --delete)
  mkdir -p "$target_dir"
  case "$MODE" in
    dry-run)
      if rsync "${cmp_args[@]}" "$source_dir/" "$target_dir/" | grep -q .; then
        hg_warn "Would sync $label: $target_dir"
      fi
      ;;
    check)
      if rsync "${cmp_args[@]}" "$source_dir/" "$target_dir/" | grep -q .; then
        hg_warn "Out of date $label: $target_dir"
        FAILED=1
      fi
      ;;
    write)
      rsync "${sync_args[@]}" "$source_dir/" "$target_dir/"
      hg_ok "Synced $label: $target_dir"
      ;;
  esac
}

restore_workspace_skill_surface_ownership() {
  local manifest_path repo_rel repo_path skill_dir plugin_skill_dir
  manifest_path="${HG_WORKSPACE_MANIFEST:-$WORKSPACE_ROOT/workspace/manifest.json}"
  [[ -f "$manifest_path" ]] || return 0

  while IFS= read -r repo_rel; do
    [[ -n "$repo_rel" ]] || continue
    repo_path="$WORKSPACE_ROOT/$repo_rel"

    skill_dir="$repo_path/.claude/skills"
    [[ -d "$skill_dir" ]] && chown -R hg:hg "$skill_dir" 2>/dev/null || true

    for plugin_skill_dir in "$repo_path"/plugins/*/skills; do
      [[ -d "$plugin_skill_dir" ]] || continue
      chown -R hg:hg "$plugin_skill_dir" 2>/dev/null || true
    done
  done < <(
    jq -r '
      .repos[]
      | select(.baseline_target == true)
      | (if ((.path // "") | length) > 0 then .path else .name end)
    ' "$manifest_path"
  )
}

normalize_gemini_settings() {
  local target="$1"
  local tmp base
  mkdir -p "$(dirname "$target")"
  base="$(mktemp)"
  tmp="$(mktemp)"
  if [[ -f "$target" ]]; then
    cp "$target" "$base"
  else
    printf '{}\n' >"$base"
  fi
  jq '
    .security = (.security // {})
    | .security.enablePermanentToolApproval = true
    | .security.autoAddToPolicyByDefault = true
    | .general = (.general // {})
    | .general.defaultApprovalMode = "auto_edit"
    | .experimental = (.experimental // {})
    | .experimental.worktrees = true
  ' "$base" >"$tmp"

  if [[ -f "$target" ]] && cmp -s "$tmp" "$target"; then
    rm -f "$base" "$tmp"
    return 0
  fi

  case "$MODE" in
    dry-run)
      hg_warn "Would normalize Gemini settings: $target"
      ;;
    check)
      hg_warn "Out of date Gemini settings: $target"
      rm -f "$base" "$tmp"
      FAILED=1
      return 0
      ;;
    write)
      mv "$tmp" "$target"
      hg_ok "Normalized Gemini settings: $target"
      ;;
  esac
  rm -f "$base" "$tmp"
}

normalize_codex_settings() {
  local target="$1"
  case "$MODE" in
    dry-run)
      if [[ ! -f "$target" ]] || ! grep -Eq '^sandbox_mode = "danger-full-access"$' "$target" || ! grep -Eq '^approval_policy = "never"$' "$target"; then
        hg_warn "Would normalize Codex config: $target"
      fi
      ;;
    check)
      if [[ ! -f "$target" ]] || ! grep -Eq '^sandbox_mode = "danger-full-access"$' "$target" || ! grep -Eq '^approval_policy = "never"$' "$target"; then
        hg_warn "Out of date Codex config: $target"
        FAILED=1
      fi
      ;;
    write)
      hg_agent_align_codex_config_file "$target"
      hg_ok "Normalized Codex config: $target"
      ;;
  esac
}

run_workspace_sync() {
  local home_dir="$1"
  local command_source_home="$2"
  local -a mode_args=()
  case "$MODE" in
    dry-run) mode_args=(--dry-run) ;;
    check) mode_args=(--check) ;;
  esac
  if ! HOME="$home_dir" HG_STUDIO_ROOT="$WORKSPACE_ROOT" \
    bash "$HG_DOTFILES/scripts/hg-workspace-global-sync.sh" \
    "${mode_args[@]}" \
    --root "$WORKSPACE_ROOT" \
    --claude-json "$home_dir/.claude.json" \
    --claude-home-doc "$home_dir/.claude/CLAUDE.md" \
    --claude-project-key "$WORKSPACE_ROOT" \
    --claude-skills-dir "$home_dir/.claude/skills" \
    --claude-commands-dir "$command_source_home/.claude/commands" \
    --agents-skills-dir "$home_dir/.agents/skills" \
    --codex-skills-dir "$home_dir/.codex/skills" \
    --codex-config "$home_dir/.codex/config.toml" \
    --gemini-home-doc "$home_dir/.gemini/GEMINI.md" \
    --gemini-projects "$home_dir/.gemini/projects.json" \
    --gemini-settings "$home_dir/.gemini/settings.json"; then
    FAILED=1
  fi
}

run_skill_sync() {
  local home_dir="$1"
  local command_source_home="$2"
  local -a mode_args=()
  case "$MODE" in
    dry-run) mode_args=(--dry-run) ;;
    check) mode_args=(--check) ;;
  esac
  if ! HOME="$home_dir" HG_STUDIO_ROOT="$WORKSPACE_ROOT" \
    HG_CLAUDE_COMMANDS_DIR="$command_source_home/.claude/commands" \
  HG_CLAUDE_SKILLS_DIR="$command_source_home/.claude/skills" \
  HG_AGENTS_SKILLS_DIR="$home_dir/.agents/skills" \
  HG_CODEX_SKILLS_DIR="$home_dir/.codex/skills" \
    bash "$HG_DOTFILES/scripts/hg-global-skill-sync.sh" "${mode_args[@]}"; then
    FAILED=1
  fi
}

seed_file_from_source_if_missing "$CLAUDE_TEMPLATE" "$USER_HOME_DIR/.claude/CLAUDE.md" "Claude home doc"
seed_file_from_source_if_missing "$CLAUDE_TEMPLATE" "$ROOT_HOME_DIR/.claude/CLAUDE.md" "root Claude home doc"
seed_file_from_source_if_missing "$GEMINI_TEMPLATE" "$USER_HOME_DIR/.gemini/GEMINI.md" "Gemini home doc"
seed_file_from_source_if_missing "$GEMINI_TEMPLATE" "$ROOT_HOME_DIR/.gemini/GEMINI.md" "root Gemini home doc"

sync_directory_tree "$USER_HOME_DIR/.claude/commands" "$ROOT_HOME_DIR/.claude/commands" "root Claude commands"
sync_directory_tree "$USER_HOME_DIR/.claude/skills" "$ROOT_HOME_DIR/.claude/skills" "root Claude skills"

run_skill_sync "$USER_HOME_DIR" "$USER_HOME_DIR"
run_skill_sync "$ROOT_HOME_DIR" "$USER_HOME_DIR"

run_workspace_sync "$USER_HOME_DIR" "$USER_HOME_DIR"
run_workspace_sync "$ROOT_HOME_DIR" "$USER_HOME_DIR"

normalize_codex_settings "$USER_HOME_DIR/.codex/config.toml"
normalize_codex_settings "$ROOT_HOME_DIR/.codex/config.toml"
normalize_gemini_settings "$USER_HOME_DIR/.gemini/settings.json"
normalize_gemini_settings "$ROOT_HOME_DIR/.gemini/settings.json"

if [[ "$MODE" == "write" ]]; then
  restore_workspace_skill_surface_ownership
  chown -R hg:hg \
    "$USER_HOME_DIR/.claude" \
    "$USER_HOME_DIR/.claude.json" \
    "$USER_HOME_DIR/.agents" \
    "$USER_HOME_DIR/.codex" \
    "$USER_HOME_DIR/.gemini" 2>/dev/null || true
fi

if [[ "$MODE" == "check" && "$FAILED" -ne 0 ]]; then
  exit 1
fi
