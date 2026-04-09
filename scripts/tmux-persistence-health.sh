#!/usr/bin/env bash
set -euo pipefail

SCRIPT_PATH="$(readlink -f "${BASH_SOURCE[0]:-$0}")"
SCRIPT_DIR="$(cd "$(dirname "$SCRIPT_PATH")" && pwd)"
source "$SCRIPT_DIR/lib/tmux-persistence.sh"

json_mode=false
[[ "${1:-}" == "--json" ]] && json_mode=true

errors=0
warnings=0
results=()

while IFS=$'\t' read -r check status detail; do
  [[ -n "${check:-}" ]] || continue
  results+=("{\"check\":\"$check\",\"status\":\"$status\",\"detail\":\"$detail\"}")
  case "$status" in
    fail) errors=$((errors + 1)) ;;
    warn) warnings=$((warnings + 1)) ;;
  esac
  if ! $json_mode; then
    printf '%-20s %-5s %s\n' "$check" "$status" "$detail"
  fi
done < <(tmux_persistence_report)

if $json_mode; then
  printf '{"errors":%d,"warnings":%d,"results":[%s]}\n' "$errors" "$warnings" "$(IFS=,; echo "${results[*]}")"
else
  printf '\nsummary: %d errors, %d warnings\n' "$errors" "$warnings"
fi

exit "$errors"
