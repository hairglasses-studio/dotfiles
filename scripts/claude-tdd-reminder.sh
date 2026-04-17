#!/usr/bin/env bash
# claude-tdd-reminder.sh — PreToolUse hook: inject a TDD reminder when a Go
# source file is about to be written without a corresponding test file being
# touched in the same session.
#
# This is an ADVISORY hook — it never blocks. It returns `allow` with a
# `systemMessage` reminder that surfaces as context to the agent, nudging
# it to write the test first.
#
# Inspired by nizos/tdd-guard and obra/superpowers "verify-before-complete"
# patterns. See dotfiles ROADMAP.md [P2][S] "TDD enforcement hook".
#
# Input: JSON on stdin with tool_name, tool_input, session_id
# Output: JSON to stdout with:
#   - `{"decision":"allow"}` — no reminder needed
#   - `{"decision":"allow","systemMessage":"<reminder>"}` — reminder injected
#
# Behavior:
#   1. Only fire on Write/Edit of *.go files (non-test).
#   2. Look at recent git log in the repo for test file changes in the last
#      hour — if any, assume the agent is on a TDD cycle and stay silent.
#   3. Look at session-local tracking at ~/.cache/claude-tdd/$SESSION_ID for
#      *_test.go paths touched; stay silent if any found.
#   4. Otherwise, emit a reminder.
#   5. Always record the current Go file write in the session cache so a
#      later test-file write suppresses future reminders within the session.
#
# Opt-in: wire via hooks config in settings.json; not installed globally.

set -euo pipefail

INPUT="$(cat)"
TOOL_NAME="$(printf '%s' "$INPUT" | jq -r '.tool_name // empty' 2>/dev/null || true)"

case "$TOOL_NAME" in
  Write|Edit|NotebookEdit) ;;
  *) printf '{"decision":"allow"}\n'; exit 0 ;;
esac

FILE_PATH="$(printf '%s' "$INPUT" | jq -r '.tool_input.file_path // .tool_input.path // empty' 2>/dev/null || true)"
SESSION_ID="$(printf '%s' "$INPUT" | jq -r '.session_id // "default"' 2>/dev/null || echo "default")"

# Filter: only Go source files, not tests or generated code
case "$FILE_PATH" in
  *.go)
    case "$FILE_PATH" in
      *_test.go|*.pb.go|*_generated.go|*_gen.go|*.gen.go|*_mock.go)
        # Test or generated file — record it and stay silent. This satisfies
        # the "test first" condition for subsequent writes in the session.
        CACHE_DIR="$HOME/.cache/claude-tdd/$SESSION_ID"
        mkdir -p "$CACHE_DIR"
        echo "$FILE_PATH" >> "$CACHE_DIR/test-writes.log"
        printf '{"decision":"allow"}\n'
        exit 0
        ;;
    esac
    ;;
  *)
    printf '{"decision":"allow"}\n'
    exit 0
    ;;
esac

# Check session cache: has the agent written any test file this session?
CACHE_DIR="$HOME/.cache/claude-tdd/$SESSION_ID"
if [[ -f "$CACHE_DIR/test-writes.log" && -s "$CACHE_DIR/test-writes.log" ]]; then
  printf '{"decision":"allow"}\n'
  exit 0
fi

# Check git: has a test file been committed or staged in the last hour? If so,
# the agent is likely on a Red-Green-Refactor loop already.
FILE_DIR="$(dirname "$FILE_PATH")"
if [[ -d "$FILE_DIR" ]] && git -C "$FILE_DIR" rev-parse --git-dir >/dev/null 2>&1; then
  RECENT_TESTS="$(git -C "$FILE_DIR" log --since='1 hour ago' --name-only --pretty=format: 2>/dev/null | grep -cE '_test\.go$' || echo 0)"
  # Strip any whitespace or non-digits from the count (grep+wc races)
  RECENT_TESTS="${RECENT_TESTS//[^0-9]/}"
  if [[ -n "$RECENT_TESTS" && "$RECENT_TESTS" -gt 0 ]]; then
    printf '{"decision":"allow"}\n'
    exit 0
  fi
fi

# No test activity detected — record this write and nudge the agent.
mkdir -p "$CACHE_DIR"
echo "$FILE_PATH" >> "$CACHE_DIR/src-writes.log"

REMINDER=$(cat <<'MSG'
TDD reminder: you are writing Go source code without a corresponding test file
touched in this session or committed in the last hour. Consider whether a test
should accompany this change — either a red test written first, or coverage
added alongside. This is advisory only; continue if a test is not appropriate
(e.g., pure refactor, test file about to follow, scaffolding).
MSG
)

# Emit as systemMessage on an allow decision — the agent receives the reminder
# as context but the tool call proceeds normally.
jq -n --arg msg "$REMINDER" '{"decision":"allow","systemMessage":$msg}'
