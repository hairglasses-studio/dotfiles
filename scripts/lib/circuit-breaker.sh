#!/usr/bin/env bash
# circuit-breaker.sh — Circuit breaker for autonomous dev-loop and fleet-autopilot
# Source this file from loop scripts to get N-failure stop, rate limit detection,
# and budget ceiling enforcement.
#
# Usage:
#   source "$(dirname "$0")/lib/circuit-breaker.sh"
#   cb_init "/path/to/.dev-loop-state.json"
#   cb_check_before_iteration   # exits 1 if breaker is open
#   ... do work ...
#   cb_record_result "shipped"  # or "no_change" or "blocked"
#   cb_check_rate_limit         # detects API rate limiting from recent errors

# ── Config (override before sourcing if needed) ──────────
CB_MAX_NO_SHIP=${CB_MAX_NO_SHIP:-3}          # Consecutive no-ship iterations before stop
CB_MAX_FAILURES=${CB_MAX_FAILURES:-3}        # Consecutive failures before stop
CB_RATE_LIMIT_WINDOW=${CB_RATE_LIMIT_WINDOW:-300}  # Seconds to check for rate limit pattern
CB_BUDGET_CEILING=${CB_BUDGET_CEILING:-20.0}       # USD budget ceiling
CB_COST_PER_ITER=${CB_COST_PER_ITER:-0.50}         # Estimated cost per iteration

_CB_STATE_FILE=""
_CB_FAILURE_COUNT=0
_CB_ITER=0

cb_init() {
  _CB_STATE_FILE="${1:-.dev-loop-state.json}"

  if [[ ! -f "$_CB_STATE_FILE" ]]; then
    printf '{"iter":0,"no_ship_streak":0,"started":"%s","last_task":"","failure_count":0}\n' \
      "$(date -Iseconds)" > "$_CB_STATE_FILE"
  fi

  _CB_ITER=$(python3 -c "import json; print(json.load(open('$_CB_STATE_FILE')).get('iter', 0))" 2>/dev/null || echo 0)
  _CB_FAILURE_COUNT=$(python3 -c "import json; print(json.load(open('$_CB_STATE_FILE')).get('failure_count', 0))" 2>/dev/null || echo 0)
}

cb_check_before_iteration() {
  local no_ship failure_count budget_used

  no_ship=$(python3 -c "import json; print(json.load(open('$_CB_STATE_FILE')).get('no_ship_streak', 0))" 2>/dev/null || echo 0)
  failure_count=$(python3 -c "import json; print(json.load(open('$_CB_STATE_FILE')).get('failure_count', 0))" 2>/dev/null || echo 0)

  # Layer 1: No-ship streak
  if [[ "$no_ship" -ge "$CB_MAX_NO_SHIP" ]]; then
    hg_error "Circuit breaker OPEN: $no_ship consecutive iterations without shipping (max: $CB_MAX_NO_SHIP)"
    return 1
  fi

  # Layer 2: Failure streak
  if [[ "$failure_count" -ge "$CB_MAX_FAILURES" ]]; then
    hg_error "Circuit breaker OPEN: $failure_count consecutive failures (max: $CB_MAX_FAILURES)"
    return 1
  fi

  # Layer 3: Budget ceiling
  budget_used=$(python3 -c "
import json
d = json.load(open('$_CB_STATE_FILE'))
print(d.get('iter', 0) * $CB_COST_PER_ITER)
" 2>/dev/null || echo 0)

  if python3 -c "exit(0 if $budget_used < $CB_BUDGET_CEILING else 1)" 2>/dev/null; then
    : # under budget
  else
    hg_error "Circuit breaker OPEN: estimated budget \$$budget_used >= ceiling \$$CB_BUDGET_CEILING"
    return 1
  fi

  return 0
}

cb_record_result() {
  local result="${1:-no_change}"

  python3 -c "
import json
d = json.load(open('$_CB_STATE_FILE'))
d['iter'] = d.get('iter', 0) + 1
if '$result' == 'shipped':
    d['no_ship_streak'] = 0
    d['failure_count'] = 0
elif '$result' == 'blocked' or '$result' == 'failed':
    d['no_ship_streak'] = d.get('no_ship_streak', 0) + 1
    d['failure_count'] = d.get('failure_count', 0) + 1
else:
    d['no_ship_streak'] = d.get('no_ship_streak', 0) + 1
d['last_task'] = '$(date -Iseconds)'
json.dump(d, open('$_CB_STATE_FILE', 'w'), indent=2)
print(f'iter={d[\"iter\"]} no_ship={d[\"no_ship_streak\"]} failures={d.get(\"failure_count\", 0)}')
" 2>/dev/null
}

cb_check_rate_limit() {
  # Check if recent git push or API calls show rate-limiting patterns
  # Look for "rate limit" or "429" in recent output
  local recent_log="${1:-}"

  if [[ -n "$recent_log" ]]; then
    if echo "$recent_log" | grep -qi "rate.limit\|429\|too many requests\|abuse detection"; then
      hg_warn "Rate limit detected — backing off"
      return 1
    fi
  fi

  return 0
}

cb_status() {
  local no_ship failure_count iter budget_est
  iter=$(python3 -c "import json; print(json.load(open('$_CB_STATE_FILE')).get('iter', 0))" 2>/dev/null || echo 0)
  no_ship=$(python3 -c "import json; print(json.load(open('$_CB_STATE_FILE')).get('no_ship_streak', 0))" 2>/dev/null || echo 0)
  failure_count=$(python3 -c "import json; print(json.load(open('$_CB_STATE_FILE')).get('failure_count', 0))" 2>/dev/null || echo 0)
  budget_est=$(python3 -c "print($iter * $CB_COST_PER_ITER)" 2>/dev/null || echo "?")

  printf 'Circuit breaker: iter=%s no_ship=%s/%s failures=%s/%s budget=$%s/$%s\n' \
    "$iter" "$no_ship" "$CB_MAX_NO_SHIP" "$failure_count" "$CB_MAX_FAILURES" \
    "$budget_est" "$CB_BUDGET_CEILING"
}
