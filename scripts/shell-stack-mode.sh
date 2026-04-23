#!/usr/bin/env bash
# shell-stack-mode.sh — staged shell migration service switcher.
#
# Defaults to dry-run. Pass --apply to change systemd/state.

set -euo pipefail

APPLY=false
MODE="status"
JSON_MODE=false
STATE_DIR="${XDG_STATE_HOME:-$HOME/.local/state}/dotfiles/shell-stack"
MODE_FILE="$STATE_DIR/mode"
ENV_FILE="$STATE_DIR/env"

usage() {
  cat <<'EOF'
Usage: shell-stack-mode.sh [--apply] [--json] <status|pilot|bar-cutover|ticker-cutover|notification-cutover|full-pilot|full-cutover|rollback>

Modes:
  status          Show current shell service state.
  pilot           Start Quickshell; keep ironbar, ticker, and swaync live.
  bar-cutover     Start Quickshell; stop ironbar; keep ticker and swaync live.
  ticker-cutover  Start Quickshell; stop keybind ticker; keep ironbar and swaync live.
  notification-cutover
                  Start Quickshell notification owner; stop swaync.
  full-pilot      Start Quickshell; stop ironbar and keybind ticker; keep swaync live.
  full-cutover    Start Quickshell as bar, ticker, and notification owner.
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
    status|pilot|bar-cutover|ticker-cutover|notification-cutover|full-pilot|full-cutover|rollback)
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

mode_or_default() {
  if [[ -f "$MODE_FILE" ]]; then
    local saved
    saved="$(<"$MODE_FILE")"
    [[ -n "$saved" ]] && {
      printf '%s\n' "$saved"
      return
    }
  fi
  printf 'pilot\n'
}

mode_flag() {
  local mode="$1" flag="$2"
  case "$flag" in
    bar)
      [[ "$mode" == "bar-cutover" || "$mode" == "full-pilot" || "$mode" == "full-cutover" ]]
      ;;
    ticker)
      [[ "$mode" == "ticker-cutover" || "$mode" == "full-pilot" || "$mode" == "full-cutover" ]]
      ;;
    notifications)
      [[ "$mode" == "notification-cutover" || "$mode" == "full-cutover" ]]
      ;;
    *)
      return 1
      ;;
  esac
}

persist_mode() {
  local mode="$1"
  local bar=0 ticker=0 notifications=0
  mode_flag "$mode" bar && bar=1
  mode_flag "$mode" ticker && ticker=1
  mode_flag "$mode" notifications && notifications=1

  if $APPLY; then
    mkdir -p "$STATE_DIR"
    printf '%s\n' "$mode" > "$MODE_FILE"
    {
      printf 'SHELL_STACK_MODE=%q\n' "$mode"
      printf 'QS_BAR_CUTOVER=%s\n' "$bar"
      printf 'QS_TICKER_CUTOVER=%s\n' "$ticker"
      printf 'QUICKSHELL_NOTIFICATION_OWNER=%s\n' "$notifications"
      if [[ -n "${QS_PRIMARY_MONITOR:-}" ]]; then
        printf 'QS_PRIMARY_MONITOR=%q\n' "$QS_PRIMARY_MONITOR"
      fi
    } > "$ENV_FILE"
  else
    printf '[dry] write %q and %q for mode %q\n' "$MODE_FILE" "$ENV_FILE" "$mode"
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
  printf '%-42s %s\n' "shell-stack-mode" "$(mode_or_default)"
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
  local saved_mode
  saved_mode="$(mode_or_default)"
  printf '{"mode":"status","shell_mode":"%s","notification_owner":%s,"services":[' \
    "$(json_escape "$saved_mode")" \
    "$(mode_flag "$saved_mode" notifications && printf true || printf false)"
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

stop_swaync() {
  if ! $APPLY; then
    stop_unit swaync.service
    return
  fi
  if systemctl --user list-unit-files swaync.service >/dev/null 2>&1; then
    stop_unit swaync.service
  else
    run_cmd pkill -x swaync
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
    persist_mode "$MODE"
    start_unit dotfiles-quickshell.service
    start_unit ironbar.service
    start_unit dotfiles-keybind-ticker.service
    start_swaync
    start_unit dotfiles-notification-history.service
    ;;
  bar-cutover)
    persist_mode "$MODE"
    start_unit dotfiles-quickshell.service
    stop_unit ironbar.service
    start_unit dotfiles-keybind-ticker.service
    start_swaync
    start_unit dotfiles-notification-history.service
    ;;
  ticker-cutover)
    persist_mode "$MODE"
    start_unit dotfiles-quickshell.service
    start_unit ironbar.service
    stop_unit dotfiles-keybind-ticker.service
    start_swaync
    start_unit dotfiles-notification-history.service
    ;;
  notification-cutover)
    persist_mode "$MODE"
    start_unit dotfiles-quickshell.service
    start_unit ironbar.service
    start_unit dotfiles-keybind-ticker.service
    stop_swaync
    start_unit dotfiles-notification-history.service
    ;;
  full-pilot)
    persist_mode "$MODE"
    start_unit dotfiles-quickshell.service
    stop_unit ironbar.service
    stop_unit dotfiles-keybind-ticker.service
    start_swaync
    start_unit dotfiles-notification-history.service
    ;;
  full-cutover)
    persist_mode "$MODE"
    start_unit dotfiles-quickshell.service
    stop_unit ironbar.service
    stop_unit dotfiles-keybind-ticker.service
    stop_swaync
    start_unit dotfiles-notification-history.service
    ;;
  rollback)
    persist_mode "$MODE"
    stop_unit dotfiles-quickshell.service
    start_unit ironbar.service
    start_unit dotfiles-keybind-ticker.service
    start_swaync
    start_unit dotfiles-notification-history.service
    ;;
esac
