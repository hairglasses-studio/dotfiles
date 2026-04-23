#!/usr/bin/env bash
# claude-skill-activate.sh — advisory hook that surfaces context-fit skills.
#
# Intended for UserPromptSubmit or early PreToolUse use. It reads a Claude Code
# hook envelope, inspects the current project and target file, and emits a
# one-time systemMessage with relevant repo skills. It never blocks.

set -euo pipefail

CACHE_ROOT="${CLAUDE_SKILL_ACTIVATE_DIR:-$HOME/.cache/claude-skill-activate}"

json_allow() {
  printf '{"decision":"allow"}\n'
}

json_message() {
  local message="$1"
  jq -n --arg msg "$message" '{"decision":"allow","systemMessage":$msg}'
}

has_path() {
  local root="$1" rel="$2"
  [[ -e "$root/$rel" ]]
}

add_suggestion() {
  local item="$1"
  local current="${SUGGESTIONS:-}"
  if [[ "$current" != *"|$item|"* ]]; then
    SUGGESTIONS="${current}|$item|"
  fi
}

suggestions_lines() {
  local raw="${SUGGESTIONS:-}"
  [[ -n "$raw" ]] || return 0
  printf '%s\n' "$raw" |
    tr '|' '\n' |
    sed '/^$/d' |
    sort -u
}

project_key() {
  local root="$1"
  printf '%s' "$root" | sha256sum | awk '{print $1}'
}

main() {
  local input tool session cwd file_path key sentinel message lines
  input="$(cat 2>/dev/null || echo '{}')"
  tool="$(printf '%s' "$input" | jq -r '.tool_name // empty' 2>/dev/null || true)"
  session="$(printf '%s' "$input" | jq -r '.session_id // "default"' 2>/dev/null || echo default)"
  cwd="$(printf '%s' "$input" | jq -r '.cwd // empty' 2>/dev/null || true)"
  file_path="$(printf '%s' "$input" | jq -r '.tool_input.file_path // .tool_input.path // empty' 2>/dev/null || true)"

  [[ -n "$cwd" && -d "$cwd" ]] || cwd="$PWD"
  key="$(project_key "$cwd")"
  sentinel="$CACHE_ROOT/$session/$key"
  if [[ -f "$sentinel" ]]; then
    json_allow
    return 0
  fi

  SUGGESTIONS=""

  if has_path "$cwd" "AGENTS.md" || has_path "$cwd" "CLAUDE.md"; then
    add_suggestion 'common_ground — verify repo assumptions before non-trivial implementation.'
  fi

  if has_path "$cwd" "hyprland" || has_path "$cwd" "hyprdynamicmonitors"; then
    add_suggestion 'dotfiles_ui — use screenshot/reload/verify loops for desktop UI and rice changes.'
    add_suggestion 'pre_risky_config — snapshot before monitor, VRR, or NVIDIA-sensitive Hyprland edits.'
  fi

  if has_path "$cwd" "scripts" || has_path "$cwd" "install.sh"; then
    add_suggestion 'dotfiles_ops — prefer shared scripts and repo-native verification for tooling changes.'
  fi

  if has_path "$cwd" "mcp/dotfiles-mcp" || has_path "$cwd" ".well-known/mcp.json"; then
    add_suggestion 'hg_mcp — use discovery-first MCP health, contract, and session diagnostics workflows.'
  fi

  if has_path "$cwd" "go.mod"; then
    add_suggestion 'go-check — run build, vet, tests, race, and formatting checks for Go changes.'
  fi

  if has_path "$cwd" "kitty/shaders" || [[ "$file_path" == *.glsl ]]; then
    add_suggestion 'shader_forge — compile-check, register, tier, and live-test new DarkWindow shaders.'
  fi

  lines="$(suggestions_lines)"
  if [[ -z "$lines" ]]; then
    json_allow
    return 0
  fi

  mkdir -p "$(dirname "$sentinel")"
  date -u +%Y-%m-%dT%H:%M:%SZ > "$sentinel"

  message=$(
    {
      printf 'Skill auto-activation: this project context matches these workflows:\n'
      while IFS= read -r line; do
        [[ -n "$line" ]] || continue
        printf -- '- %s\n' "$line"
      done <<< "$lines"
      printf '\nUse the relevant skill instructions before editing matching surfaces. This is advisory only.'
    }
  )

  json_message "$message"
}

main "$@"
