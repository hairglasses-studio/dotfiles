#!/usr/bin/env bash
# health-watchdog.sh — periodic liveness checks for dotfiles-mcp, the event
# bus, and ticker cache producers. Invoked by dotfiles-health-watchdog.timer
# at 30s cadence.
#
# Passes:
#   1. MCP server: run `run-dotfiles-mcp.sh --contract-print` with a 10s
#      timeout. On failure, append a {type:"mcp_dead"} record to
#      ~/.claude/recovery-events.jsonl and notify-send -u critical.
#   2. Event bus: check dotfiles-event-bus.service is active, AND that
#      ~/.local/state/dotfiles/events.jsonl has been written within
#      BUS_STALE_THRESHOLD_S. Restart the service and record
#      {type:"event_bus_dead"} on failure. This is the trust anchor for
#      /heal and canary Tier 6 — if the bus dies, those skills would
#      otherwise report "all clear" when the workstation is silently
#      drifting.
#   3. Ticker cache files under /tmp/bar-*.txt: if age > 2x the producer
#      timer's OnUnitActiveSec, record {type:"ticker_stale"} and
#      systemctl --user restart <producer>.
#
# Best-effort: every step is wrapped so a transient failure doesn't kill
# subsequent checks. All structured events write NDJSON lines so downstream
# tools (claude-session-context, /heal, canary) can tail them cheaply.

set -o pipefail

DOTFILES_DIR="${DOTFILES_DIR:-$HOME/hairglasses-studio/dotfiles}"
EVENTS_LOG="${HOME}/.claude/recovery-events.jsonl"
RUN_MCP="$DOTFILES_DIR/scripts/run-dotfiles-mcp.sh"
BUS_HEARTBEAT="${XDG_STATE_HOME:-$HOME/.local/state}/dotfiles/bus.heartbeat"
# The bus touches bus.heartbeat every 60s from a dedicated async task, so
# the mtime is tick-driven (not event-driven). 5 minutes is enough slack
# to ride out a single slow poll cycle while still catching a wedged
# daemon within a couple of watchdog runs.
BUS_STALE_THRESHOLD_S=300

# Ensure event log dir exists (~/.claude/ should exist already; mkdir -p is safe).
mkdir -p "$(dirname "$EVENTS_LOG")"

now_epoch() { date +%s; }
now_rfc() { date -u +'%Y-%m-%dT%H:%M:%SZ'; }

log_event() {
  local json="$1"
  printf '%s\n' "$json" >> "$EVENTS_LOG" 2>/dev/null || true
}

notify_critical() {
  local title="$1" body="$2"
  command -v notify-send >/dev/null 2>&1 || return 0
  notify-send -u critical -a health-watchdog "$title" "$body" >/dev/null 2>&1 || true
}

# ---------------------------------------------------------------------------
# Pass 1: MCP liveness
# ---------------------------------------------------------------------------

check_mcp() {
  [[ -x "$RUN_MCP" ]] || return 0
  # Use `timeout` from coreutils; treat exit 124 (timed out) and any non-zero
  # as dead. --contract-print is lightweight (no stdio MCP handshake) and
  # fails fast if the binary panics or the registry is corrupt.
  # 10s allows for a cold rebuild (measured ~2s); warm runs return in ~30ms.
  local out rc
  out="$(timeout --preserve-status 10s "$RUN_MCP" --contract-print 2>&1)"
  rc=$?
  if [[ $rc -ne 0 ]]; then
    local fingerprint
    fingerprint="$(printf '%s' "$out" | head -c 200 | tr -d '\n\r' | sed 's/"/\\"/g')"
    local payload
    payload=$(printf '{"type":"mcp_dead","at":"%s","rc":%d,"fingerprint":"%s"}' \
                 "$(now_rfc)" "$rc" "$fingerprint")
    log_event "$payload"
    notify_critical "dotfiles-mcp probe failed (rc=$rc)" \
      "See ~/.claude/recovery-events.jsonl — restart Claude Code or run $RUN_MCP --contract-print"
  fi
}

# ---------------------------------------------------------------------------
# Pass 2: event-bus liveness
# ---------------------------------------------------------------------------

check_event_bus() {
  local active
  active="$(systemctl --user is-active dotfiles-event-bus.service 2>/dev/null)"
  if [[ "$active" != "active" ]]; then
    local payload
    payload=$(printf '{"type":"event_bus_dead","at":"%s","state":"%s"}' \
                 "$(now_rfc)" "$active")
    log_event "$payload"
    notify_critical "dotfiles-event-bus is $active" \
      "Restarting — see ~/.claude/recovery-events.jsonl for details"
    systemctl --user restart dotfiles-event-bus.service >/dev/null 2>&1 || true
    return
  fi
  # Service is active — check the heartbeat file the bus touches once a
  # minute. Absent heartbeat = bus may still be warming up on first boot;
  # the file lands within ~60s so skip without flagging until then.
  [[ -f "$BUS_HEARTBEAT" ]] || return
  local age
  age="$(file_age_seconds "$BUS_HEARTBEAT")"
  [[ -n "$age" ]] || return
  if (( age > BUS_STALE_THRESHOLD_S )); then
    local payload
    payload=$(printf '{"type":"event_bus_stale","at":"%s","age_s":%d,"threshold_s":%d}' \
                 "$(now_rfc)" "$age" "$BUS_STALE_THRESHOLD_S")
    log_event "$payload"
    systemctl --user restart dotfiles-event-bus.service >/dev/null 2>&1 || true
  fi
}

# ---------------------------------------------------------------------------
# Pass 3: ticker cache producers
# ---------------------------------------------------------------------------

# Map /tmp/bar-<stream>.txt back to a producer timer.
# Most streams use bar-<stream>.timer. We look up OnUnitActiveSec from the
# timer unit file and flag staleness if age > 2x that interval.

ticker_producer_timer() {
  local stream="$1"
  # Non-convention files that should NOT be staleness-checked. Each entry
  # is a cache-file stream basename whose refresh cadence is driven by
  # external state (geolocation, user interaction, etc.) rather than a
  # systemd timer. Returning empty tells the caller to skip this file.
  case "$stream" in
    # weather-coords is a geolocation cache — it only rewrites when the
    # user's location changes. A stationary workstation legitimately
    # goes weeks without an update; a staleness restart of bar-weather
    # would fire uselessly.
    weather-coords) return ;;
  esac
  # Heuristic: the producer service has the same basename as the cache file.
  # bar-gpu-full.txt is written by bar-gpu.service (same producer, richer
  # output) — strip the "-full" suffix in that one case.
  local base="$stream"
  base="${base%-full}"
  printf 'bar-%s.timer\n' "$base"
}

# Parse OnUnitActiveSec= from a timer file and return a seconds count.
# Accepts plain integers (10), explicit seconds (30s), minutes (15min),
# hours (1h). Returns 0 when unparseable so the caller can skip.
timer_interval_seconds() {
  local timer_unit="$1"
  local unit_path
  unit_path="$(systemctl --user show "$timer_unit" -p FragmentPath --value 2>/dev/null)"
  [[ -n "$unit_path" && -r "$unit_path" ]] || return 0
  local raw
  raw="$(grep -m1 '^OnUnitActiveSec=' "$unit_path" | cut -d= -f2 | tr -d '[:space:]')"
  [[ -n "$raw" ]] || return 0
  case "$raw" in
    *h) printf '%d\n' "$(( ${raw%h} * 3600 ))" ;;
    *min) printf '%d\n' "$(( ${raw%min} * 60 ))" ;;
    *s) printf '%d\n' "${raw%s}" ;;
    *[!0-9]*) return 0 ;;
    *) printf '%d\n' "$raw" ;;
  esac
}

file_age_seconds() {
  local path="$1"
  local mtime now
  mtime="$(stat -c %Y "$path" 2>/dev/null)"
  [[ -n "$mtime" ]] || return 0
  now="$(now_epoch)"
  printf '%d\n' $(( now - mtime ))
}

check_ticker_cache() {
  shopt -s nullglob
  for path in /tmp/bar-*.txt; do
    local stream base timer interval age threshold
    base="$(basename "$path" .txt)"
    stream="${base#bar-}"
    timer="$(ticker_producer_timer "$stream")"
    interval="$(timer_interval_seconds "$timer")"
    # Skip when we can't resolve the producer or its cadence.
    [[ -n "$interval" && "$interval" -gt 0 ]] || continue
    age="$(file_age_seconds "$path")"
    [[ -n "$age" ]] || continue
    threshold=$(( interval * 2 ))
    if (( age > threshold )); then
      local payload
      payload=$(printf '{"type":"ticker_stale","at":"%s","stream":"%s","age_s":%d,"interval_s":%d,"producer_unit":"%s"}' \
                   "$(now_rfc)" "$stream" "$age" "$interval" "$timer")
      log_event "$payload"
      # Restart the producer — idempotent on a healthy timer.
      systemctl --user restart "$timer" >/dev/null 2>&1 || true
    fi
  done
}

# ---------------------------------------------------------------------------
# Main
# ---------------------------------------------------------------------------

check_mcp
check_event_bus
check_ticker_cache

# Always exit 0 so systemd does not record repeated "failure" events — the
# interesting signal goes to the events log, not to the journal.
exit 0
