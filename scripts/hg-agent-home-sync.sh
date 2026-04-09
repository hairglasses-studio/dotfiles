#!/usr/bin/env bash
# hg-agent-home-sync.sh — Keep user and root provider homes aligned with the managed launcher contract.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
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

sync_file_from_source() {
  local source="$1"
  local target="$2"
  local label="$3"

  mkdir -p "$(dirname "$target")"
  if [[ -f "$target" ]] && cmp -s "$source" "$target"; then
    return 0
  fi

  case "$MODE" in
    dry-run)
      hg_warn "Would sync $label: $target"
      ;;
    check)
      hg_warn "Out of date $label: $target"
      FAILED=1
      return 0
      ;;
    write)
      cp "$source" "$target"
      hg_ok "Synced $label: $target"
      ;;
  esac
}

sync_directory_tree() {
  local source_dir="$1"
  local target_dir="$2"
  local label="$3"
  mkdir -p "$target_dir"
  case "$MODE" in
    dry-run)
      if rsync -ani --delete "$source_dir/" "$target_dir/" | grep -q .; then
        hg_warn "Would sync $label: $target_dir"
      fi
      ;;
    check)
      if rsync -ani --delete "$source_dir/" "$target_dir/" | grep -q .; then
        hg_warn "Out of date $label: $target_dir"
        FAILED=1
      fi
      ;;
    write)
      rsync -a --delete "$source_dir/" "$target_dir/"
      hg_ok "Synced $label: $target_dir"
      ;;
  esac
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
    | .general.defaultApprovalMode = "yolo"
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
  if ! HOME="$USER_HOME_DIR" HG_STUDIO_ROOT="$WORKSPACE_ROOT" \
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
  if ! HOME="$USER_HOME_DIR" HG_STUDIO_ROOT="$WORKSPACE_ROOT" \
    HG_CLAUDE_COMMANDS_DIR="$command_source_home/.claude/commands" \
  HG_CLAUDE_SKILLS_DIR="$command_source_home/.claude/skills" \
  HG_AGENTS_SKILLS_DIR="$home_dir/.agents/skills" \
  HG_CODEX_SKILLS_DIR="$home_dir/.codex/skills" \
    bash "$HG_DOTFILES/scripts/hg-global-skill-sync.sh" "${mode_args[@]}"; then
    FAILED=1
  fi
}

sync_file_from_source "$CLAUDE_TEMPLATE" "$USER_HOME_DIR/.claude/CLAUDE.md" "Claude home doc"
sync_file_from_source "$CLAUDE_TEMPLATE" "$ROOT_HOME_DIR/.claude/CLAUDE.md" "root Claude home doc"
sync_file_from_source "$GEMINI_TEMPLATE" "$USER_HOME_DIR/.gemini/GEMINI.md" "Gemini home doc"
sync_file_from_source "$GEMINI_TEMPLATE" "$ROOT_HOME_DIR/.gemini/GEMINI.md" "root Gemini home doc"

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
