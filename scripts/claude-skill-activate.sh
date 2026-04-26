#!/usr/bin/env bash
# claude-skill-activate.sh - PreToolUse context-to-skill injector.
#
# Detects common dotfiles work surfaces and injects the relevant skill guidance
# once per session. This is advisory: it always allows the tool call.

set -euo pipefail

INPUT="$(cat 2>/dev/null || echo '{}')"

jq_field() {
  local query="$1"
  printf '%s' "$INPUT" | jq -r "$query // empty" 2>/dev/null || true
}

TOOL_NAME="$(jq_field '.tool_name')"
SESSION_ID="$(jq_field '.session_id')"
FILE_PATH="$(jq_field '.tool_input.file_path // .tool_input.path')"
COMMAND="$(jq_field '.tool_input.command')"
CWD="$(jq_field '.cwd')"

SESSION_ID="${SESSION_ID:-default}"
SESSION_KEY="$(printf '%s' "$SESSION_ID" | tr -c 'A-Za-z0-9_.-' '_')"
CACHE_ROOT="${CLAUDE_SKILL_ACTIVATE_CACHE:-$HOME/.cache/claude-skill-activate}"
CACHE_DIR="$CACHE_ROOT/$SESSION_KEY"
SEEN_FILE="$CACHE_DIR/seen.log"

REPO_ROOT="$(git rev-parse --show-toplevel 2>/dev/null || pwd)"

json_allow() {
  printf '{"decision":"allow"}\n'
}

json_message() {
  local message="$1"
  jq -cn --arg msg "$message" '{
    decision:"allow",
    systemMessage:$msg,
    hookSpecificOutput:{
      hookEventName:"PreToolUse",
      additionalContext:$msg
    }
  }'
}

relative_path() {
  local path="$1"
  [[ -z "$path" ]] && return 0
  case "$path" in
    "$REPO_ROOT"/*) printf '%s\n' "${path#"$REPO_ROOT"/}" ;;
    "$HOME"/hairglasses-studio/dotfiles/*) printf '%s\n' "${path#"$HOME"/hairglasses-studio/dotfiles/}" ;;
    *) printf '%s\n' "$path" ;;
  esac
}

lower_text() {
  tr '[:upper:]' '[:lower:]'
}

seen_skill() {
  local key="$1"
  [[ -f "$SEEN_FILE" ]] && grep -qxF "$key" "$SEEN_FILE"
}

mark_seen() {
  local key="$1"
  mkdir -p "$CACHE_DIR"
  printf '%s\n' "$key" >> "$SEEN_FILE"
}

append_skill() {
  local key="$1"
  local existing
  for existing in "${SKILLS[@]}"; do
    [[ "$existing" == "$key" ]] && return 0
  done
  if ! seen_skill "$key"; then
    SKILLS+=("$key")
  fi
}

detect_context() {
  local rel lower command_lower
  rel="$(relative_path "$FILE_PATH")"
  lower="$(printf '%s %s %s' "$rel" "$COMMAND" "${CWD:-}" | lower_text)"
  command_lower="$(printf '%s' "$COMMAND" | lower_text)"

  case "$rel" in
    *.go)
      append_skill "go"
      ;;
  esac

  if printf '%s' "$command_lower" | grep -qE '(^|[[:space:];&|])(go[[:space:]]+(test|build|vet|run|list)|gofmt|gofumpt|goimports)([[:space:];&|]|$)|\.go([[:space:];&|]|$)'; then
    append_skill "go"
  fi

  case "$rel" in
    mcp/*|*.mcp.json|.well-known/mcp.json|snapshots/contract/*.json|scripts/run-dotfiles-mcp.sh|.claude/settings.json)
      append_skill "mcp"
      ;;
  esac

  if printf '%s' "$lower" | grep -qE '(^|/)(mcp|dotfiles-mcp)(/|$)|mcpkit|mcp__|\.mcp\.json|snapshots/contract|run-dotfiles-mcp'; then
    append_skill "mcp"
  fi

  case "$rel" in
    *.glsl|*.frag|kitty/shaders/*|*/darkwindow/*|scripts/shader-*.sh|scripts/validate-darkwindow-shaders.sh)
      append_skill "shader"
      ;;
  esac

  if printf '%s' "$lower" | grep -qE '\.(glsl|frag)([[:space:];&|]|$)|darkwindow|glslangvalidator|shader-(wallpaper|preview|playlist|consistency)'; then
    append_skill "shader"
  fi
}

skill_message() {
  local parts=()
  local skill
  for skill in "${SKILLS[@]}"; do
    case "$skill" in
      go)
        # shellcheck disable=SC2016 # Literal backticks are guidance text.
        parts+=('Go context: activate `go-conventions` and `mcpkit-go` guidance. For MCP handlers, return `handler.CodedErrorResult(...), nil` for classified errors, use handler result builders/param extractors, and verify with `go test ./... -count=1` or the repo-specific check.')
        ;;
      mcp)
        # shellcheck disable=SC2016 # Literal backticks are guidance text.
        parts+=('MCP context: activate `.agents/skills/hg_mcp/SKILL.md` plus `mcpkit-go` when Go code is involved. Check tool/resource contract snapshots, avoid ad-hoc MCP config edits, and run the matching MCP smoke/contract validation before shipping.')
        ;;
      shader)
        # shellcheck disable=SC2016 # Literal backticks are guidance text.
        parts+=('Shader context: activate `.agents/skills/shader_forge/SKILL.md`. Preserve terminal readability, compile-check GLSL, update playlists/registry together, and run `scripts/check-shader-consistency.sh` after shader changes.')
        ;;
    esac
  done

  (IFS=$'\n'; printf '%s\n' "${parts[*]}")
}

main() {
  case "$TOOL_NAME" in
    Read|Write|Edit|NotebookEdit|Bash) ;;
    *) json_allow; return ;;
  esac

  SKILLS=()
  detect_context

  if [[ "${#SKILLS[@]}" -eq 0 ]]; then
    json_allow
    return
  fi

  local skill
  for skill in "${SKILLS[@]}"; do
    mark_seen "$skill"
  done

  json_message "Skill auto-activation:
$(skill_message)"
}

main "$@"
