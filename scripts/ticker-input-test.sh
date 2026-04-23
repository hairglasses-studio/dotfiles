#!/usr/bin/env bash
# ticker-input-test.sh — smoke-test the ticker control surface.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CONTROL="$SCRIPT_DIR/ticker-control.sh"

json_status() {
  "$CONTROL" status --json
}

prop() {
  local key="$1"
  json_status | jq -r --arg key "$key" 'if has($key) then .[$key] else "" end'
}

assert_eq() {
  local want="$1" got="$2" label="$3"
  if [[ "$got" != "$want" ]]; then
    printf 'FAIL  %s: expected %q, got %q\n' "$label" "$want" "$got" >&2
    exit 1
  fi
  printf 'ok    %s: %q\n' "$label" "$got"
}

if ! json_status >/dev/null; then
  printf 'ticker control surface not reachable\n' >&2
  exit 2
fi

prev_pin="$(prop pinned)"
prev_shuffle="$(prop shuffle)"
prev_playlist="$(prop playlist)"

cleanup() {
  if [[ -n "$prev_pin" ]]; then
    "$CONTROL" pin "$prev_pin" >/dev/null 2>&1 || true
  else
    "$CONTROL" unpin >/dev/null 2>&1 || true
  fi
  if [[ "$prev_shuffle" == "true" ]]; then
    "$CONTROL" shuffle on >/dev/null 2>&1 || true
  else
    "$CONTROL" shuffle off >/dev/null 2>&1 || true
  fi
  [[ -n "$prev_playlist" ]] && "$CONTROL" playlist "$prev_playlist" >/dev/null 2>&1 || true
}
trap cleanup EXIT

echo '━━ Pin / Unpin round-trip ━━'
"$CONTROL" pin load >/dev/null
sleep 0.4
assert_eq "load" "$(prop pinned)" "pin reflected in status"
assert_eq "load" "$(prop current_stream)" "pin advanced current stream"
"$CONTROL" unpin >/dev/null
sleep 0.2
assert_eq "" "$(prop pinned)" "unpin cleared pinned"

echo '━━ PinToggle ━━'
"$CONTROL" pin-toggle >/dev/null
sleep 0.3
pinned_after="$(prop pinned)"
if [[ -z "$pinned_after" ]]; then
  printf 'FAIL  pin-toggle should have pinned a stream\n' >&2
  exit 1
fi
printf 'ok    pin-toggle pinned: %q\n' "$pinned_after"
"$CONTROL" pin-toggle >/dev/null
sleep 0.2
assert_eq "" "$(prop pinned)" "pin-toggle unpinned"

echo '━━ Shuffle on/off ━━'
"$CONTROL" shuffle on >/dev/null
sleep 0.3
assert_eq "true" "$(prop shuffle)" "shuffle on reflected"
"$CONTROL" shuffle off >/dev/null
sleep 0.3
assert_eq "false" "$(prop shuffle)" "shuffle off reflected"

echo '━━ Next advances ━━'
before="$(prop current_stream)"
"$CONTROL" next >/dev/null
sleep 0.3
after="$(prop current_stream)"
if [[ "$before" == "$after" ]]; then
  printf 'FAIL  next did not advance (still on %q)\n' "$before" >&2
  exit 1
fi
printf 'ok    next advanced: %q -> %q\n' "$before" "$after"

echo '━━ Urgent / Snooze ━━'
"$CONTROL" urgent true >/dev/null
sleep 0.2
assert_eq "true" "$(prop urgent)" "urgent true reflected"
"$CONTROL" snooze-urgent >/dev/null
sleep 0.2
assert_eq "false" "$(prop urgent)" "snooze cleared urgent"

echo
echo 'ALL PASS'
