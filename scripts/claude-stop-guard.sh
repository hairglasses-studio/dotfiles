#!/usr/bin/env bash
# claude-stop-guard.sh — Laziness detection hook for Claude Code Stop events.
# Catches ownership-dodging, permission-seeking, premature stopping, and
# session-length excuses. Exit 2 blocks the stop and injects a continuation
# message. Based on stellaraccident's stop-phrase-guard pattern (GitHub #42796).
#
# Install: add to ~/.claude/settings.json hooks.Stop

set -euo pipefail

input="$(cat)"
response="$(printf '%s' "$input" | jq -r '.assistant_response // empty' 2>/dev/null || true)"

[[ -z "$response" ]] && exit 0

lower="$(printf '%s' "$response" | tr '[:upper:]' '[:lower:]')"

# --- Ownership-dodging patterns ---
if printf '%s' "$lower" | grep -qE \
  'not caused by (my|our) changes|existing issue|pre-existing|was already broken|not related to (my|our|this)'; then
  echo '{"decision":"block","reason":"Ownership-dodging detected. Own the problem and fix it."}' >&2
  exit 2
fi

# --- Permission-seeking patterns ---
if printf '%s' "$lower" | grep -qE \
  'should i continue|want me to keep going|shall i proceed|would you like me to|do you want me to continue|may i proceed'; then
  echo '{"decision":"block","reason":"Do not ask for permission. Continue working until the task is complete."}' >&2
  exit 2
fi

# --- Premature stopping patterns ---
if printf '%s' "$lower" | grep -qE \
  'good stopping point|natural checkpoint|good place to stop|pause here|stop here for now|leave (it|this) for now'; then
  echo '{"decision":"block","reason":"Do not stop prematurely. Complete the full task."}' >&2
  exit 2
fi

# --- Known-limitation labeling ---
if printf '%s' "$lower" | grep -qE \
  'known limitation|future work|out of scope|beyond the scope|for a future'; then
  echo '{"decision":"block","reason":"Do not label work as out of scope. Complete what was asked."}' >&2
  exit 2
fi

# --- Session-length excuses ---
if printf '%s' "$lower" | grep -qE \
  'continue in a new session|session is getting long|pick this up later|start a new conversation'; then
  echo '{"decision":"block","reason":"Do not cite session length. Continue working."}' >&2
  exit 2
fi

exit 0
