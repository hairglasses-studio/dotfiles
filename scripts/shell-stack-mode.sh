#!/usr/bin/env bash
# shell-stack-mode.sh — staged shell migration service switcher.
#
# Defaults to dry-run. Pass --apply to change systemd state.

set -euo pipefail

APPLY=false
MODE="status"
JSON_MODE=false

usage() {
  cat <<'EOF'
Usage: shell-stack-mode.sh [--apply] [--json] <status|pilot|bar-cutover|ticker-cutover|full-pilot|rollback>

Modes:
  status          Show current shell service state.
  pilot           Start Quickshell; keep ironbar, ticker, and swaync live.
  bar-cutover     Start Quickshell; stop ironbar; keep ticker and swaync live.
  ticker-cutover  Start Quickshell; stop keybind ticker; keep ironbar and swaync live.
  full-pilot      Start Quickshell; stop ironbar and keybind ticker; keep swaync live.
  rollback        Stop Quickshell; start ironbar, keybind ticker, and swaync.

Without --apply, prints the commands it would run.
--json is supported for status output.
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --apply)
      APPLY=true
      shift
      ;;
    --json)
      JSON_MODE=true
      shift
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    status|pilot|bar-cutover|ticker-cutover|full-pilot|rollback)
      MODE="$1"
      shift
      ;;
    *)
      printf 'Unknown argument: %s\n' "$1" >&2
      usage >&2
      exit 2
      ;;
  esac
done

run_cmd() {
  if $APPLY; then
    printf '+ %s\n' "$*"
    "$@"
  else
    printf '[dry] %q' "$1"
    shift
    for arg in "$@"; do
      printf ' %q' "$arg"
    done
    printf '\n'
  fi
}

service_state() {
  local unit="$1"
  if [[ "$unit" == "swaync.service" ]]; then
    if pgrep -x swaync >/dev/null 2>&1; then
      printf 'active-process\n'
      return
    fi
  fi
  systemctl --user is-active "$unit" 2>/dev/null || true
}

json_escape() {
  local value="$1"
  value=${value//\\/\\\\}
  value=${value//\"/\\\"}
  value=${value//$'\n'/\\n}
  value=${value//$'\r'/\\r}
  value=${value//$'\t'/\\t}
  printf '%s' "$value"
}

print_status() {
  local unit state
  for unit in \
    dotfiles-quickshell.service \
    ironbar.service \
    dotfiles-keybind-ticker.service \
    swaync.service \
    dotfiles-notification-history.service; do
    state="$(service_state "$unit")"
    printf '%-42s %s\n' "$unit" "${state:-unknown}"
  done
}

print_status_json() {
  local unit state first=true
  printf '{"mode":"status","services":['
  for unit in \
    dotfiles-quickshell.service \
    ironbar.service \
    dotfiles-keybind-ticker.service \
    swaync.service \
    dotfiles-notification-history.service; do
    state="$(service_state "$unit")"
    if $first; then
      first=false
    else
      printf ','
    fi
    printf '{"unit":"%s","state":"%s"}' "$(json_escape "$unit")" "$(json_escape "${state:-unknown}")"
  done
  printf ']}\n'
}

start_unit() { run_cmd systemctl --user start "$1"; }
stop_unit() { run_cmd systemctl --user stop "$1"; }

start_swaync() {
  if systemctl --user list-unit-files swaync.service >/dev/null 2>&1; then
    start_unit swaync.service
  else
    run_cmd bash -lc 'pgrep -x swaync >/dev/null || nohup swaync >/tmp/swaync.log 2>&1 &'
  fi
}

case "$MODE" in
  status)
    if $JSON_MODE; then
      print_status_json
    else
      print_status
    fi
    ;;
  pilot)
    start_unit dotfiles-quickshell.service
    start_unit ironbar.service
    start_unit dotfiles-keybind-ticker.service
    start_unit dotfiles-notification-history.service
    ;;
  bar-cutover)
    start_unit dotfiles-quickshell.service
    stop_unit ironbar.service
    start_unit dotfiles-keybind-ticker.service
    start_unit dotfiles-notification-history.service
    ;;
  ticker-cutover)
    start_unit dotfiles-quickshell.service
    start_unit ironbar.service
    stop_unit dotfiles-keybind-ticker.service
    start_unit dotfiles-notification-history.service
    ;;
  full-pilot)
    start_unit dotfiles-quickshell.service
    stop_unit ironbar.service
    stop_unit dotfiles-keybind-ticker.service
    start_unit dotfiles-notification-history.service
    ;;
  rollback)
    stop_unit dotfiles-quickshell.service
    start_unit ironbar.service
    start_unit dotfiles-keybind-ticker.service
    start_swaync
    start_unit dotfiles-notification-history.service
    ;;
esac
