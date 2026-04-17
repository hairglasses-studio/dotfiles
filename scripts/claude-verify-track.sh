#!/usr/bin/env bash
# claude-verify-track.sh — PreToolUse hook: record source writes and test
# command invocations for the claude-verify-gate.sh Stop hook.
#
# Populates two per-session logs under ~/.cache/claude-verify-gate/$SESSION_ID:
#   - src-writes.log  — one line per Write/Edit of a source file
#   - test-runs.log   — one line per Bash tool call matching a test-runner pattern
#
# Always returns {"decision":"allow"} — this is pure instrumentation.
#
# Wire as:
#   "PreToolUse": [{"matcher":"Write|Edit|NotebookEdit|Bash",
#                   "hooks":[{"type":"command","command":"<path>/claude-verify-track.sh"}]}]

set -euo pipefail

INPUT="$(cat 2>/dev/null || echo '{}')"
TOOL="$(printf '%s' "$INPUT" | jq -r '.tool_name // empty' 2>/dev/null || true)"
SESSION_ID="$(printf '%s' "$INPUT" | jq -r '.session_id // "default"' 2>/dev/null || echo "default")"

CACHE_DIR="$HOME/.cache/claude-verify-gate/$SESSION_ID"

case "$TOOL" in
  Write|Edit|NotebookEdit)
    FILE_PATH="$(printf '%s' "$INPUT" | jq -r '.tool_input.file_path // .tool_input.path // empty' 2>/dev/null || true)"
    # Only count source-code files (not docs, configs, notes)
    case "$FILE_PATH" in
      *.go|*.ts|*.tsx|*.js|*.jsx|*.py|*.rs|*.java|*.kt|*.swift|*.rb|*.c|*.cpp|*.h|*.hpp)
        # Skip generated/test files — tests themselves don't count as "source
        # needing verification", and generated code is out of scope.
        case "$FILE_PATH" in
          *_test.go|*.pb.go|*_gen.go|*_generated.go|*.test.ts|*.test.js|*test_*.py|*.spec.ts|*.spec.js) ;;
          *)
            mkdir -p "$CACHE_DIR"
            echo "$FILE_PATH" >> "$CACHE_DIR/src-writes.log"
            ;;
        esac
        ;;
    esac
    ;;
  Bash)
    CMD="$(printf '%s' "$INPUT" | jq -r '.tool_input.command // empty' 2>/dev/null || true)"
    # Match common test-runner patterns. Keep conservative — false negatives
    # cost a reminder, false positives suppress a legitimate nudge.
    if echo "$CMD" | grep -qE '(^|[^a-zA-Z_])(go test|cargo test|pytest|npm test|npm run test|yarn test|make test|make check|ci|bun test)([^a-zA-Z_]|$)'; then
      mkdir -p "$CACHE_DIR"
      echo "$CMD" >> "$CACHE_DIR/test-runs.log"
    fi
    ;;
esac

printf '{"decision":"allow"}\n'
