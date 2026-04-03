#!/usr/bin/env bash
# audit-dashboard.sh — View MCP tool invocation audit log
# Usage: audit-dashboard [recent|errors|hot|summary]

set -euo pipefail

AUDIT_LOG="${AUDIT_LOG:-${XDG_STATE_HOME:-$HOME/.local/state}/mcpkit/audit.jsonl}"

# Snazzy palette ANSI codes
CYAN='\033[38;2;87;199;255m'
MAGENTA='\033[38;2;255;106;193m'
GREEN='\033[38;2;90;247;142m'
YELLOW='\033[38;2;243;249;157m'
RED='\033[38;2;255;92;87m'
GRAY='\033[38;2;104;104;104m'
FG='\033[38;2;241;241;240m'
BOLD='\033[1m'
RESET='\033[0m'

check_deps() {
  if ! command -v jq &>/dev/null; then
    echo -e "${RED}error:${RESET} jq is required but not found" >&2
    exit 1
  fi
}

check_log() {
  if [[ ! -f "$AUDIT_LOG" ]]; then
    echo -e "${GRAY}No audit log found at ${AUDIT_LOG}${RESET}"
    exit 0
  fi
  if [[ ! -s "$AUDIT_LOG" ]]; then
    echo -e "${GRAY}Audit log is empty${RESET}"
    exit 0
  fi
}

cmd_recent() {
  echo -e "${BOLD}${CYAN}MCP Audit Log — Recent Invocations${RESET}"
  echo -e "${GRAY}$(printf '%.0s─' {1..60})${RESET}"
  tail -n 20 "$AUDIT_LOG" | jq -r '
    [
      (.timestamp // .ts // "?" | if length > 19 then .[0:19] else . end),
      (.tool // .name // "?"),
      (.tier // "?"),
      (if .duration_ms then "\(.duration_ms)ms"
       elif .duration then "\(.duration)"
       else "?" end),
      (if .error then "ERR" else "OK" end)
    ] | @tsv' 2>/dev/null | while IFS=$'\t' read -r ts tool tier dur status; do
      if [[ "$status" == "ERR" ]]; then
        status_color="$RED"
      else
        status_color="$GREEN"
      fi
      printf "${GRAY}%s${RESET}  ${CYAN}%-30s${RESET}  ${MAGENTA}%-8s${RESET}  ${YELLOW}%8s${RESET}  ${status_color}%s${RESET}\n" \
        "$ts" "$tool" "$tier" "$dur" "$status"
    done
}

cmd_errors() {
  echo -e "${BOLD}${RED}MCP Audit Log — Errors${RESET}"
  echo -e "${GRAY}$(printf '%.0s─' {1..60})${RESET}"
  local count
  count=$(jq -s '[.[] | select(.error)] | length' "$AUDIT_LOG" 2>/dev/null || echo 0)
  if [[ "$count" == "0" ]]; then
    echo -e "${GREEN}No errors found${RESET}"
    return
  fi
  jq -rs '[.[] | select(.error)] | .[-20:][] |
    [
      (.timestamp // .ts // "?" | if length > 19 then .[0:19] else . end),
      (.tool // .name // "?"),
      (.error | tostring | if length > 60 then .[0:60] + "..." else . end)
    ] | @tsv' "$AUDIT_LOG" 2>/dev/null | while IFS=$'\t' read -r ts tool err; do
      printf "${GRAY}%s${RESET}  ${CYAN}%-25s${RESET}  ${RED}%s${RESET}\n" "$ts" "$tool" "$err"
    done
}

cmd_hot() {
  echo -e "${BOLD}${MAGENTA}MCP Audit Log — Top 10 Tools${RESET}"
  echo -e "${GRAY}$(printf '%.0s─' {1..60})${RESET}"
  jq -rs 'group_by(.tool // .name) | map({tool: .[0].tool // .[0].name // "?", count: length}) | sort_by(-.count) | .[0:10][] |
    "\(.count)\t\(.tool)"' "$AUDIT_LOG" 2>/dev/null | while IFS=$'\t' read -r count tool; do
      local max_width=30
      # Simple bar chart
      local bar_len=$((count > max_width ? max_width : count))
      local bar
      bar=$(printf '%.0s█' $(seq 1 "$bar_len"))
      printf "${YELLOW}%5s${RESET}  ${CYAN}%-30s${RESET}  ${GREEN}%s${RESET}\n" "$count" "$tool" "$bar"
    done
}

cmd_summary() {
  echo -e "${BOLD}${GREEN}MCP Audit Log — Summary${RESET}"
  echo -e "${GRAY}$(printf '%.0s─' {1..60})${RESET}"
  local today
  today=$(date '+%Y-%m-%d')
  jq -s --arg today "$today" '
    . as $all |
    ($all | length) as $total |
    ($all | [.[] | select(.error)] | length) as $errs |
    ($all | [.[] | select((.timestamp // .ts // "") | startswith($today))] | length) as $today_calls |
    ($all | [.[] | select(.duration_ms) | .duration_ms]) as $durations |
    ($durations | if length > 0 then (add / length | floor) else 0 end) as $avg_dur |
    ($all | [.[] | .tier // "unknown"] | sort | group_by(.) | map({tier: .[0], count: length}) | sort_by(-.count)) as $tiers |
    {
      total: $total,
      errors: $errs,
      error_rate: (if $total > 0 then (($errs * 100 / $total * 10 | floor) / 10) else 0 end),
      today: $today_calls,
      avg_duration_ms: $avg_dur,
      tiers: $tiers
    }
  ' "$AUDIT_LOG" 2>/dev/null | jq -r '
    "  Total calls:     \(.total)",
    "  Today:           \(.today)",
    "  Errors:          \(.errors) (\(.error_rate)%)",
    "  Avg duration:    \(.avg_duration_ms)ms",
    "",
    "  By tier:",
    (.tiers[] | "    \(.tier): \(.count)")
  ' | while IFS= read -r line; do
    if [[ "$line" == *"Errors:"* && "$line" != *"(0%)"* ]]; then
      echo -e "  ${RED}${line}${RESET}"
    elif [[ "$line" == *"By tier:"* ]]; then
      echo -e "  ${MAGENTA}${line}${RESET}"
    elif [[ "$line" == "    "* ]]; then
      echo -e "  ${CYAN}${line}${RESET}"
    else
      echo -e "  ${FG}${line}${RESET}"
    fi
  done
}

main() {
  check_deps
  check_log

  case "${1:-recent}" in
    recent)  cmd_recent ;;
    errors)  cmd_errors ;;
    hot)     cmd_hot ;;
    summary) cmd_summary ;;
    *)
      echo -e "${BOLD}Usage:${RESET} audit-dashboard [recent|errors|hot|summary]"
      echo ""
      echo -e "  ${CYAN}recent${RESET}   Show last 20 invocations (default)"
      echo -e "  ${RED}errors${RESET}   Show only failed invocations"
      echo -e "  ${MAGENTA}hot${RESET}      Show most-called tools (top 10)"
      echo -e "  ${GREEN}summary${RESET}  Show totals, error rate, avg duration, by-tier breakdown"
      exit 1
      ;;
  esac
}

main "$@"
