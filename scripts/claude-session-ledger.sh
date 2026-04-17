#!/usr/bin/env bash
# claude-session-ledger.sh — Session ledger handoff: write structured YAML
# at Stop time, read and inject at SessionStart time.
#
# Enables continuity across Claude Code session boundaries (compaction,
# restart, new chat): the next session sees what the previous session
# decided, shipped, and intended to do next.
#
# Inspired by Continuous-Claude-v3 (3.7K stars) session handoff pattern.
# Complementary to claude-tdd-reminder.sh and claude-verify-gate.sh.
#
# Mode is chosen by $1 (first arg): `write` on Stop, `read` on SessionStart.
#
# Ledger file: ~/.cache/claude-session-ledger/project-<basename>/latest.yaml
# Project keyed by $(basename $PWD) so different repos don't collide.
#
# Wire as:
#   "Stop":         [{"hooks":[{"type":"command","command":"<path> write"}]}]
#   "SessionStart": [{"hooks":[{"type":"command","command":"<path> read"}]}]

set -euo pipefail

MODE="${1:-read}"
INPUT="$(cat 2>/dev/null || echo '{}')"

SESSION_ID="$(printf '%s' "$INPUT" | jq -r '.session_id // "default"' 2>/dev/null || echo "default")"
PROJECT="$(basename "$PWD")"
LEDGER_DIR="$HOME/.cache/claude-session-ledger/project-$PROJECT"
LEDGER="$LEDGER_DIR/latest.yaml"

case "$MODE" in
  write)
    mkdir -p "$LEDGER_DIR"

    # Gather state cheaply — all shell primitives, no network.
    TIMESTAMP="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
    BRANCH="$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo unknown)"
    HEAD_SHA="$(git rev-parse --short HEAD 2>/dev/null || echo unknown)"
    RECENT_COMMITS="$(git log --oneline -5 2>/dev/null | sed 's/^/    - /' || true)"
    DIRTY_COUNT="$(git status --short 2>/dev/null | wc -l | tr -d ' ')"
    DIRTY_COUNT="${DIRTY_COUNT:-0}"

    # Pull summarized context from sibling hook caches if they exist.
    TDD_CACHE="$HOME/.cache/claude-tdd/$SESSION_ID"
    VERIFY_CACHE="$HOME/.cache/claude-verify-gate/$SESSION_ID"
    SRC_WRITES=0
    TEST_WRITES=0
    TEST_RUNS=0
    if [[ -f "$TDD_CACHE/src-writes.log" ]]; then
      SRC_WRITES="$(wc -l < "$TDD_CACHE/src-writes.log" | tr -d ' ')"
    fi
    if [[ -f "$TDD_CACHE/test-writes.log" ]]; then
      TEST_WRITES="$(wc -l < "$TDD_CACHE/test-writes.log" | tr -d ' ')"
    fi
    if [[ -f "$VERIFY_CACHE/test-runs.log" ]]; then
      TEST_RUNS="$(wc -l < "$VERIFY_CACHE/test-runs.log" | tr -d ' ')"
    fi

    # Write the ledger. YAML keeps it human-readable for users inspecting
    # ~/.cache between sessions.
    {
      echo "# Claude Code session ledger — $PROJECT"
      echo "session_id: \"$SESSION_ID\""
      echo "project: \"$PROJECT\""
      echo "cwd: \"$PWD\""
      echo "stopped_at: \"$TIMESTAMP\""
      echo "git:"
      echo "  branch: \"$BRANCH\""
      echo "  head: \"$HEAD_SHA\""
      echo "  dirty_files: $DIRTY_COUNT"
      echo "  recent_commits:"
      if [[ -n "$RECENT_COMMITS" ]]; then
        echo "$RECENT_COMMITS"
      fi
      echo "activity:"
      echo "  source_writes: $SRC_WRITES"
      echo "  test_writes: $TEST_WRITES"
      echo "  test_runs: $TEST_RUNS"
    } > "$LEDGER"

    # Hooks may output JSON for decisions; a Stop hook returning {} is a no-op.
    printf '{}\n'
    ;;

  read)
    if [[ ! -f "$LEDGER" ]]; then
      # No prior session — nothing to inject.
      printf '{}\n'
      exit 0
    fi

    # Pull ledger, wrap it as additionalContext so the agent sees it as
    # part of the session-start injection.
    CONTENT="$(cat "$LEDGER")"
    jq -n --arg content "$CONTENT" '{
      "hookSpecificOutput": {
        "hookEventName": "SessionStart",
        "additionalContext": ("Previous session ledger for this project:\n\n```yaml\n" + $content + "\n```\n\nReview before continuing; previous session may have left work in a partial state.")
      }
    }'
    ;;

  *)
    echo "usage: $0 {write|read}" >&2
    exit 1
    ;;
esac
