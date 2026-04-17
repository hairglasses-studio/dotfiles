#!/usr/bin/env bash
# claude-verify-gate.sh — Stop hook: nudge the agent to run tests before
# declaring a task complete.
#
# This is an ADVISORY hook — it never blocks the stop. It tracks per-session
# whether any test-like command has run and, if the session wrote source code
# without running tests, emits a systemMessage reminder at stop time.
#
# Inspired by obra/superpowers "verify-before-complete" pattern. Companion to
# claude-tdd-reminder.sh (PreToolUse): TDD reminder covers the write side,
# verify-gate covers the done side.
#
# Input: JSON on stdin with session_id, tool_name=null (Stop event).
# Output: JSON to stdout, either:
#   - `{}` — no reminder needed, stop proceeds
#   - `{"systemMessage":"<reminder>"}` — reminder surfaced before stop
#
# Wire as:
#   "Stop": [{"hooks": [{"type":"command","command":"<path>/claude-verify-gate.sh"}]}]
#
# Track events in ~/.cache/claude-verify-gate/$SESSION_ID:
#   - src-writes.log (populated by companion PreToolUse hook OR by Bash hook)
#   - test-runs.log  (populated by Bash wrapper that matches test commands)
#
# This script reads those logs and decides. If you want the full loop,
# also wire a PreToolUse Bash hook that appends test-command invocations
# to test-runs.log (see example in dotfiles/templates/agent-parity/).

set -euo pipefail

INPUT="$(cat 2>/dev/null || echo '{}')"
SESSION_ID="$(printf '%s' "$INPUT" | jq -r '.session_id // "default"' 2>/dev/null || echo "default")"

CACHE_DIR="$HOME/.cache/claude-verify-gate/$SESSION_ID"
SRC_LOG="$CACHE_DIR/src-writes.log"
TEST_LOG="$CACHE_DIR/test-runs.log"

# If no cache exists for this session, nothing to say.
if [[ ! -f "$SRC_LOG" ]]; then
  printf '{}\n'
  exit 0
fi

# Count source writes vs test runs
SRC_COUNT=$(wc -l < "$SRC_LOG" 2>/dev/null | tr -d ' ')
SRC_COUNT="${SRC_COUNT:-0}"

TEST_COUNT=0
if [[ -f "$TEST_LOG" ]]; then
  TEST_COUNT=$(wc -l < "$TEST_LOG" 2>/dev/null | tr -d ' ')
  TEST_COUNT="${TEST_COUNT:-0}"
fi

# If source was written but no tests ran, remind.
if [[ "$SRC_COUNT" -gt 0 && "$TEST_COUNT" -eq 0 ]]; then
  # Cleanup cache after emitting — the reminder fires once per session-stop.
  # Leave src-writes intact so subsequent stops in the same session don't
  # double-nudge: create a sentinel instead.
  if [[ -f "$CACHE_DIR/reminded" ]]; then
    # Already reminded this session, stay silent.
    printf '{}\n'
    exit 0
  fi
  touch "$CACHE_DIR/reminded"

  REMINDER=$(cat <<'MSG'
Verify-before-complete: this session wrote source code but no test command
was observed (e.g., `go test`, `npm test`, `pytest`, `make test`). Before
calling the task done, consider running the repo's test suite to catch
regressions. This is advisory only — if tests are not applicable (e.g.,
pure docs work, infra change), continue.
MSG
  )
  jq -n --arg msg "$REMINDER" '{"systemMessage":$msg}'
  exit 0
fi

printf '{}\n'
