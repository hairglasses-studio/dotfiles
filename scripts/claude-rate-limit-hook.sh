#!/usr/bin/env bash
# claude-rate-limit-hook.sh — StopFailure hook for Claude Code.
# Fires when the turn ends due to an API error (rate limit, auth failure, etc.)
# Writes a structured event to ~/.claude/provider-events.jsonl for the
# provider rotation system to pick up on next SelectProvider() call.
#
# Install: add to ~/.claude/settings.json hooks.StopFailure

set -euo pipefail

input="$(cat)"

# Extract fields from StopFailure event
stop_reason="$(printf '%s' "$input" | jq -r '.stop_reason // empty' 2>/dev/null || true)"
error_type="$(printf '%s' "$input" | jq -r '.error.type // empty' 2>/dev/null || true)"
error_msg="$(printf '%s' "$input" | jq -r '.error.message // empty' 2>/dev/null || true)"
response="$(printf '%s' "$input" | jq -r '.assistant_response // empty' 2>/dev/null || true)"

# Nothing to do if no error info
[[ -z "$stop_reason" && -z "$error_type" && -z "$error_msg" ]] && exit 0

# Classify the event
event="stop_failure"
provider="claude"

# Check for rate limit indicators
lower_reason="$(printf '%s' "$stop_reason $error_type $error_msg $response" | tr '[:upper:]' '[:lower:]')"
if printf '%s' "$lower_reason" | grep -qE 'rate.?limit|429|quota.?exhaust|out.?of.?credits|subscription.?usage.?exhaust|extra.?usage.?exhaust'; then
  event="rate_limit"
fi

# Write event to JSONL file
events_file="$HOME/.claude/provider-events.jsonl"
timestamp="$(date -u +"%Y-%m-%dT%H:%M:%SZ")"
cwd="${PWD:-unknown}"

printf '{"event":"%s","provider":"%s","timestamp":"%s","reason":"%s","error_type":"%s","cwd":"%s"}\n' \
  "$event" "$provider" "$timestamp" "$stop_reason${error_type:+ ($error_type)}" "$error_type" "$cwd" \
  >> "$events_file"

# Desktop notification for visibility
if command -v notify-send >/dev/null 2>&1; then
  notify-send -u critical -t 10000 "Claude Rate Limit" "Provider: $provider\nReason: $stop_reason" 2>/dev/null || true
fi

exit 0
