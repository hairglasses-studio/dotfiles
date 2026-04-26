#!/usr/bin/env bash
# claude-phase-gate.sh - hook state machine for gated /dev-loop runs.
#
# Enforces the high-level sequence:
#   plan -> human review -> implement -> verify -> ship/stop
#
# UserPromptSubmit starts or approves a gate. PreToolUse blocks mutating work
# before approval and blocks shipping commands before verification. PostToolUse
# records successful verification commands. Stop blocks completion when writes
# happened after the last verification.

set -euo pipefail

INPUT="$(cat 2>/dev/null || echo '{}')"

jq_field() {
  local query="$1"
  printf '%s' "$INPUT" | jq -r "$query // empty" 2>/dev/null || true
}

EVENT="$(jq_field '.hook_event_name')"
TOOL_NAME="$(jq_field '.tool_name')"
PROMPT="$(jq_field '.prompt')"
SESSION_ID="$(jq_field '.session_id')"
CWD="$(jq_field '.cwd')"

if [[ -z "$EVENT" ]]; then
  if [[ -n "$PROMPT" ]]; then
    EVENT="UserPromptSubmit"
  elif printf '%s' "$INPUT" | jq -e 'has("tool_response")' >/dev/null 2>&1; then
    EVENT="PostToolUse"
  elif [[ -n "$TOOL_NAME" ]]; then
    EVENT="PreToolUse"
  else
    EVENT="Stop"
  fi
fi

SESSION_ID="${SESSION_ID:-default}"
SESSION_KEY="$(printf '%s' "$SESSION_ID" | tr -c 'A-Za-z0-9_.-' '_')"
CACHE_ROOT="${CLAUDE_PHASE_GATE_CACHE:-$HOME/.cache/claude-phase-gate}"
CACHE_DIR="$CACHE_ROOT/$SESSION_KEY"
STATE_FILE="$CACHE_DIR/state.json"

now_epoch() {
  date +%s
}

now_iso() {
  date -u +%Y-%m-%dT%H:%M:%SZ
}

json_allow() {
  printf '{"decision":"allow"}\n'
}

json_block() {
  local reason="$1"
  jq -cn --arg reason "$reason" '{decision:"block",reason:$reason}'
}

json_message() {
  local message="$1"
  jq -cn --arg msg "$message" --arg event "$EVENT" '{
    decision:"allow",
    systemMessage:$msg,
    hookSpecificOutput:{
      hookEventName:$event,
      additionalContext:$msg
    }
  }'
}

state_exists() {
  [[ -f "$STATE_FILE" ]]
}

state_active() {
  state_exists && jq -e '.active == true' "$STATE_FILE" >/dev/null 2>&1
}

state_phase() {
  state_exists && jq -r '.phase // "inactive"' "$STATE_FILE" 2>/dev/null || printf 'inactive\n'
}

state_last_write_epoch() {
  state_exists && jq -r '.last_write_seq // .last_write_epoch // 0' "$STATE_FILE" 2>/dev/null || printf '0\n'
}

state_last_verify_epoch() {
  state_exists && jq -r '.last_verify_seq // .last_verify_epoch // 0' "$STATE_FILE" 2>/dev/null || printf '0\n'
}

state_verify_count() {
  state_exists && jq -r '.verify_count // 0' "$STATE_FILE" 2>/dev/null || printf '0\n'
}

write_state() {
  local tmp
  mkdir -p "$CACHE_DIR"
  tmp="$(mktemp "$STATE_FILE.XXXXXX")"
  cat > "$tmp"
  mv -f "$tmp" "$STATE_FILE"
}

init_state() {
  local phase="$1"
  local task="$2"
  local review_required="${3:-true}"
  local now epoch cwd
  now="$(now_iso)"
  epoch="$(now_epoch)"
  cwd="${CWD:-$PWD}"

  jq -n \
    --arg phase "$phase" \
    --arg task "$task" \
    --arg cwd "$cwd" \
    --arg session "$SESSION_ID" \
    --arg now "$now" \
    --argjson epoch "$epoch" \
    --argjson review_required "$review_required" \
    '{
      active:true,
      phase:$phase,
      session_id:$session,
      cwd:$cwd,
      task:$task,
      review_required:$review_required,
      started_at:$now,
      updated_at:$now,
      started_epoch:$epoch,
      last_write_at:"",
      last_write_epoch:0,
      last_write_seq:0,
      last_verify_at:"",
      last_verify_epoch:0,
      last_verify_seq:0,
      write_count:0,
      verify_count:0,
      event_seq:0
    }' | write_state
}

update_state() {
  local filter="$1"
  local now epoch tmp
  state_exists || return 0
  now="$(now_iso)"
  epoch="$(now_epoch)"
  tmp="$(mktemp "$STATE_FILE.XXXXXX")"
  jq --arg now "$now" --argjson epoch "$epoch" "$filter" "$STATE_FILE" > "$tmp"
  mv -f "$tmp" "$STATE_FILE"
}

record_write() {
  # shellcheck disable=SC2016 # jq filter uses jq variables, not shell vars.
  update_state '.event_seq=(.event_seq // 0) + 1 | .phase="implement" | .updated_at=$now | .last_write_at=$now | .last_write_epoch=$epoch | .last_write_seq=.event_seq | .write_count=(.write_count // 0) + 1'
}

record_verify() {
  # shellcheck disable=SC2016 # jq filter uses jq variables, not shell vars.
  update_state '.event_seq=(.event_seq // 0) + 1 | .phase="verify" | .updated_at=$now | .last_verify_at=$now | .last_verify_epoch=$epoch | .last_verify_seq=.event_seq | .verify_count=(.verify_count // 0) + 1'
}

mark_awaiting_review() {
  # shellcheck disable=SC2016 # jq filter uses jq variables, not shell vars.
  update_state '.phase="awaiting_review" | .updated_at=$now'
}

mark_approved() {
  # shellcheck disable=SC2016 # jq filter uses jq variables, not shell vars.
  update_state '.phase="implement" | .approved_at=$now | .updated_at=$now'
}

mark_complete() {
  # shellcheck disable=SC2016 # jq filter uses jq variables, not shell vars.
  update_state '.active=false | .phase="complete" | .completed_at=$now | .updated_at=$now'
}

reset_state() {
  if state_exists; then
    # shellcheck disable=SC2016 # jq filter uses jq variables, not shell vars.
    update_state '.active=false | .phase="reset" | .reset_at=$now | .updated_at=$now'
  fi
}

lower_text() {
  tr '[:upper:]' '[:lower:]'
}

prompt_starts_dev_loop() {
  local lower="$1"
  printf '%s' "$lower" | grep -qE '(^|[[:space:]])/dev-loop([[:space:]]|$)'
}

prompt_read_only_dev_loop() {
  local lower="$1"
  printf '%s' "$lower" | grep -qE '(^|[[:space:]])/dev-loop[[:space:]]+(status|orient|research)([[:space:]]|$)'
}

prompt_approves_gate() {
  local lower="$1"
  printf '%s' "$lower" | grep -qE '(approve(d)?|lgtm|looks good).*(dev-loop|phase[- ]gate|plan)|(dev-loop|phase[- ]gate|plan).*(approve(d)?|lgtm|looks good)'
}

prompt_resets_gate() {
  local lower="$1"
  printf '%s' "$lower" | grep -qE '(reset|cancel|clear).*(dev-loop|phase[- ]gate)|(dev-loop|phase[- ]gate).*(reset|cancel|clear)'
}

bash_command() {
  jq_field '.tool_input.command'
}

is_verify_command() {
  local cmd="$1"
  printf '%s' "$cmd" | grep -qiE '(^|[[:space:];&|])(go[[:space:]]+test|cargo[[:space:]]+test|pytest|python[[:space:]]+-m[[:space:]]+pytest|npm[[:space:]]+(run[[:space:]]+)?test|pnpm[[:space:]]+test|yarn[[:space:]]+test|bun[[:space:]]+test|bats|make[[:space:]]+(ci|test|check|lint)|/pipeline[[:space:]]+check|bash[[:space:]]+scripts/validate-config-syntax\.sh|git[[:space:]]+diff[[:space:]]+--check)([[:space:];&|]|$)'
}

is_ship_command() {
  local cmd="$1"
  printf '%s' "$cmd" | grep -qiE '(^|[[:space:];&|])(git[[:space:]]+(commit|push|tag)|gh[[:space:]]+pr[[:space:]]+(create|merge)|gh[[:space:]]+release[[:space:]]+create|make[[:space:]]+ship)([[:space:];&|]|$)'
}

is_mutating_bash() {
  local cmd="$1"
  if is_verify_command "$cmd"; then
    return 1
  fi
  printf '%s' "$cmd" | grep -qiE '(^|[[:space:];&|])(apply_patch|git[[:space:]]+(add|commit|push|tag|merge|rebase|reset|checkout|switch|stash)|gh[[:space:]]+(pr[[:space:]]+(create|merge|edit|close|reopen)|release[[:space:]]+create)|rm[[:space:]]+|mv[[:space:]]+|cp[[:space:]]+|mkdir[[:space:]]+|touch[[:space:]]+|chmod[[:space:]]+|chown[[:space:]]+|ln[[:space:]]+|install[[:space:]]+|sed[[:space:]].*-i|perl[[:space:]].*-pi|tee[[:space:]]|npm[[:space:]]+(install|i|add)|pnpm[[:space:]]+(install|add)|yarn[[:space:]]+(add|install)|cargo[[:space:]]+add|go[[:space:]]+get)([[:space:];&|]|$)'
}

needs_review() {
  local phase
  phase="$(state_phase)"
  [[ "$phase" == "plan" || "$phase" == "awaiting_review" ]]
}

needs_verify() {
  local last_write last_verify
  last_write="$(state_last_write_epoch)"
  last_verify="$(state_last_verify_epoch)"
  [[ "$last_write" =~ ^[0-9]+$ ]] || last_write=0
  [[ "$last_verify" =~ ^[0-9]+$ ]] || last_verify=0
  (( last_write > last_verify ))
}

needs_ship_verify() {
  local verify_count
  verify_count="$(state_verify_count)"
  [[ "$verify_count" =~ ^[0-9]+$ ]] || verify_count=0
  (( verify_count == 0 )) || needs_verify
}

handle_prompt() {
  local lower phase task
  lower="$(printf '%s' "$PROMPT" | lower_text)"

  if prompt_resets_gate "$lower"; then
    reset_state
    json_message "Dev-loop phase gate reset for this session."
    return
  fi

  if state_active && prompt_approves_gate "$lower"; then
    phase="$(state_phase)"
    if [[ "$phase" == "plan" || "$phase" == "awaiting_review" ]]; then
      mark_approved
      json_message "Dev-loop phase gate approved. Implementation tools are now allowed. Run verification after the last write before commit, push, PR creation, or final stop."
      return
    fi
  fi

  if prompt_starts_dev_loop "$lower" && ! prompt_read_only_dev_loop "$lower"; then
    task="$(printf '%s' "$PROMPT" | head -1 | cut -c1-180)"
    init_state "plan" "$task" true
    json_message "Dev-loop phase gate active. Phase: plan. Produce the implementation plan and stop for human review. Writes and mutating shell commands are blocked until the user approves the plan (for example: 'approve dev-loop plan')."
    return
  fi

  json_allow
}

handle_pre_tool() {
  local cmd
  state_active || { json_allow; return; }

  case "$TOOL_NAME" in
    Write|Edit|NotebookEdit)
      if needs_review; then
        json_block "Dev-loop phase gate: write tools are blocked until the plan is approved. Ask the user to reply 'approve dev-loop plan' after reviewing the plan."
        return
      fi
      record_write
      json_allow
      ;;
    Bash)
      cmd="$(bash_command)"
      if [[ -z "$cmd" ]]; then
        json_allow
        return
      fi
      if needs_review && is_mutating_bash "$cmd"; then
        json_block "Dev-loop phase gate: mutating shell commands are blocked until the plan is approved. Ask the user to reply 'approve dev-loop plan' after reviewing the plan."
        return
      fi
      if is_ship_command "$cmd" && needs_ship_verify; then
        json_block "Dev-loop phase gate: shipping commands are blocked until verification runs after the last write."
        return
      fi
      if ! is_ship_command "$cmd" && is_mutating_bash "$cmd"; then
        record_write
      fi
      json_allow
      ;;
    *)
      json_allow
      ;;
  esac
}

post_tool_failed() {
  printf '%s' "$INPUT" | jq -e '
    (.tool_response.success == false)
    or ((.tool_response.exit_code? // 0) != 0)
  ' >/dev/null 2>&1
}

handle_post_tool() {
  local cmd
  state_active || { json_allow; return; }

  if [[ "$TOOL_NAME" == "Bash" ]]; then
    cmd="$(bash_command)"
    if [[ -n "$cmd" ]] && is_verify_command "$cmd"; then
      if post_tool_failed; then
        json_message "Dev-loop phase gate: verification command finished with a failure signal, so the gate still requires a passing verification before ship/stop."
        return
      fi
      record_verify
      json_message "Dev-loop phase gate: verification recorded. Commit, push, PR creation, and final stop are allowed until another write occurs."
      return
    fi
  fi

  json_allow
}

handle_stop() {
  state_active || { json_allow; return; }

  case "$(state_phase)" in
    plan)
      mark_awaiting_review
      json_message "Dev-loop phase gate: plan phase is complete. Waiting for human review; reply 'approve dev-loop plan' to unlock implementation."
      return
      ;;
    awaiting_review)
      json_allow
      return
      ;;
  esac

  if needs_verify; then
    json_block "Dev-loop phase gate: this session wrote or mutated files after the last verification. Run the repo verification command before stopping."
    exit 2
  fi

  mark_complete
  json_allow
}

case "$EVENT" in
  UserPromptSubmit) handle_prompt ;;
  PreToolUse) handle_pre_tool ;;
  PostToolUse) handle_post_tool ;;
  Stop) handle_stop ;;
  *) json_allow ;;
esac
