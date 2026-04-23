#!/usr/bin/env bash
# claude-phase-gate.sh — opt-in PreToolUse guard for phase-gated dev loops.
#
# Hook mode reads a Claude Code tool-call envelope from stdin and enforces the
# current session phase:
#   plan/review  -> block writes and mutating shell commands
#   implement    -> allow implementation tools
#   verify       -> block further writes, allow test/check commands
#
# CLI mode advances the state machine:
#   claude-phase-gate.sh reset --session <id>
#   claude-phase-gate.sh mark review --session <id>
#   claude-phase-gate.sh mark implement --session <id> --reviewed-by <label>
#   claude-phase-gate.sh mark verify --session <id>
#   claude-phase-gate.sh mark done --session <id>
#   claude-phase-gate.sh status --session <id>
#
# The script is intentionally not wired into repo settings by default. It is a
# hard gate when installed as a PreToolUse hook, but it requires an explicit
# operator review label before entering implementation.

set -euo pipefail

STATE_ROOT="${CLAUDE_PHASE_GATE_DIR:-$HOME/.cache/claude-phase-gate}"

json_allow() {
  printf '{"decision":"allow"}\n'
}

json_block() {
  local reason="$1"
  jq -n --arg reason "$reason" '{"decision":"block","reason":$reason}'
}

session_dir() {
  local session="$1"
  printf '%s/%s\n' "$STATE_ROOT" "$session"
}

phase_file() {
  local session="$1"
  printf '%s/phase\n' "$(session_dir "$session")"
}

review_file() {
  local session="$1"
  printf '%s/reviewed_by\n' "$(session_dir "$session")"
}

current_phase() {
  local session="$1"
  local file
  file="$(phase_file "$session")"
  if [[ -f "$file" ]]; then
    cat "$file"
  else
    printf 'plan\n'
  fi
}

write_phase() {
  local session="$1" phase="$2"
  mkdir -p "$(session_dir "$session")"
  printf '%s\n' "$phase" > "$(phase_file "$session")"
}

parse_session_arg() {
  local default_session="${1:-default}"
  shift || true
  local session="$default_session"
  while (($#)); do
    case "$1" in
      --session)
        session="${2:?missing --session value}"
        shift 2
        ;;
      --session=*)
        session="${1#--session=}"
        shift
        ;;
      *)
        shift
        ;;
    esac
  done
  printf '%s\n' "$session"
}

reviewed_by_arg() {
  local reviewed_by=""
  while (($#)); do
    case "$1" in
      --reviewed-by)
        reviewed_by="${2:?missing --reviewed-by value}"
        shift 2
        ;;
      --reviewed-by=*)
        reviewed_by="${1#--reviewed-by=}"
        shift
        ;;
      *)
        shift
        ;;
    esac
  done
  printf '%s\n' "$reviewed_by"
}

mark_phase() {
  local phase="${1:?missing phase}"
  shift || true
  local session reviewed_by current
  session="$(parse_session_arg default "$@")"
  reviewed_by="$(reviewed_by_arg "$@")"
  current="$(current_phase "$session")"

  case "$phase" in
    plan)
      write_phase "$session" plan
      ;;
    review)
      case "$current" in
        plan|review) write_phase "$session" review ;;
        *) printf 'cannot move from %s to review\n' "$current" >&2; return 1 ;;
      esac
      ;;
    implement)
      if [[ -z "$reviewed_by" ]]; then
        printf 'mark implement requires --reviewed-by <label>\n' >&2
        return 1
      fi
      case "$current" in
        review|implement)
          write_phase "$session" implement
          printf '%s\n' "$reviewed_by" > "$(review_file "$session")"
          ;;
        *) printf 'cannot move from %s to implement\n' "$current" >&2; return 1 ;;
      esac
      ;;
    verify)
      case "$current" in
        implement|verify) write_phase "$session" verify ;;
        *) printf 'cannot move from %s to verify\n' "$current" >&2; return 1 ;;
      esac
      ;;
    done)
      case "$current" in
        verify|done) write_phase "$session" done ;;
        *) printf 'cannot move from %s to done\n' "$current" >&2; return 1 ;;
      esac
      ;;
    *)
      printf 'unknown phase: %s\n' "$phase" >&2
      return 1
      ;;
  esac
}

status_json() {
  local session="$1"
  local phase reviewed_by=""
  phase="$(current_phase "$session")"
  if [[ -f "$(review_file "$session")" ]]; then
    reviewed_by="$(cat "$(review_file "$session")")"
  fi
  jq -n \
    --arg session "$session" \
    --arg phase "$phase" \
    --arg reviewed_by "$reviewed_by" \
    '{session:$session, phase:$phase, reviewed_by:$reviewed_by}'
}

is_phase_gate_command() {
  local command="$1"
  [[ "$command" == *"claude-phase-gate.sh"* ]]
}

is_write_tool() {
  case "$1" in
    Write|Edit|MultiEdit|NotebookEdit) return 0 ;;
    *) return 1 ;;
  esac
}

is_test_command() {
  local command="$1"
  printf '%s' "$command" | grep -qE '(^|[^a-zA-Z_])(go test|cargo test|pytest|npm test|npm run test|yarn test|make test|make check|bats|bash -n|shellcheck)([^a-zA-Z_]|$)'
}

is_mutating_command() {
  local command="$1"
  printf '%s' "$command" | grep -qE '(^|[;&|[:space:]])(apply_patch|rm|mv|cp|chmod|chown|ln|mkdir|touch|tee|git[[:space:]]+(add|commit|push|reset|checkout|merge|rebase|pull)|npm[[:space:]]+install|go[[:space:]]+get|cargo[[:space:]]+add|make[[:space:]]+(install|sync))([^a-zA-Z0-9_-]|$)'
}

hook_mode() {
  local input tool session phase command=""
  input="$(cat 2>/dev/null || echo '{}')"
  tool="$(printf '%s' "$input" | jq -r '.tool_name // empty' 2>/dev/null || true)"
  session="$(printf '%s' "$input" | jq -r '.session_id // "default"' 2>/dev/null || echo default)"
  phase="$(current_phase "$session")"

  if [[ "$tool" == "Bash" ]]; then
    command="$(printf '%s' "$input" | jq -r '.tool_input.command // empty' 2>/dev/null || true)"
    if is_phase_gate_command "$command"; then
      json_allow
      return 0
    fi
  fi

  case "$phase" in
    plan|review)
      if is_write_tool "$tool"; then
        json_block "Phase gate: $phase phase blocks implementation tools until review is complete and implement is marked."
        return 0
      fi
      if [[ "$tool" == "Bash" ]] && is_mutating_command "$command"; then
        json_block "Phase gate: $phase phase blocks mutating shell commands until review is complete and implement is marked."
        return 0
      fi
      ;;
    verify)
      if is_write_tool "$tool"; then
        json_block "Phase gate: verify phase blocks further edits; run checks or move back to implement deliberately."
        return 0
      fi
      if [[ "$tool" == "Bash" ]] && is_mutating_command "$command" && ! is_test_command "$command"; then
        json_block "Phase gate: verify phase allows checks but blocks mutating shell commands."
        return 0
      fi
      ;;
  esac

  json_allow
}

main() {
  local cmd="${1:-hook}"
  case "$cmd" in
    hook)
      hook_mode
      ;;
    reset)
      shift || true
      local session
      session="$(parse_session_arg default "$@")"
      rm -rf "$(session_dir "$session")"
      status_json "$session"
      ;;
    mark)
      shift || true
      mark_phase "$@"
      local session
      session="$(parse_session_arg default "$@")"
      status_json "$session"
      ;;
    status)
      shift || true
      status_json "$(parse_session_arg default "$@")"
      ;;
    *)
      hook_mode
      ;;
  esac
}

main "$@"
