#!/usr/bin/env bash
# shell-stack-mode.sh — Quickshell-only stack switcher.
#
# After the 2026-04 staged migration, the legacy ironbar/keybind-ticker/
# hyprshell/hypr-dock stack is retired. Only two end-states remain:
#
#   full-cutover  — Quickshell owns every surface (the default).
#   rollback      — Stop Quickshell. The legacy stack is not auto-restarted
#                   here; if a user genuinely needs to roll back to it they
#                   should re-enable the units manually (`systemctl --user
#                   enable --now ironbar.service` etc.) and revert the
#                   relevant install.sh commit. Kept as an escape hatch
#                   while the legacy repo content is still in tree.
#
# `--apply` is accepted as a no-op for backwards compatibility with
# pre-2026-04-26 callers; everything is apply-by-default now.

set -euo pipefail

MODE="status"
JSON_MODE=false
STATE_DIR="${XDG_STATE_HOME:-$HOME/.local/state}/dotfiles/shell-stack"
MODE_FILE="$STATE_DIR/mode"
ENV_FILE="$STATE_DIR/env"

usage() {
  cat <<'EOF'
Usage: shell-stack-mode.sh [--apply] [--json] <status|full-cutover|rollback>

Modes:
  status        Show current shell service state.
  full-cutover  Start Quickshell as the sole owner of bar, ticker, menus,
                dock, notifications, and companion overlays. The default
                end-state on a fresh install.
  rollback     Stop Quickshell. Re-enable legacy units manually if needed
                (see comment at top of script). Kept while the legacy repo
                content is still in tree.

--json applies to status output. --apply is accepted as a no-op (apply is
the default).
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --apply)
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
    status|full-cutover|rollback)
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
  printf '+ %s\n' "$*"
  "$@"
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
  printf 'full-cutover\n'
}

persist_mode() {
  local mode="$1"
  mkdir -p "$STATE_DIR"
  printf '%s\n' "$mode" > "$MODE_FILE"
  {
    printf 'SHELL_STACK_MODE=%q\n' "$mode"
    if [[ -n "${QS_PRIMARY_MONITOR:-}" ]]; then
      printf 'QS_PRIMARY_MONITOR=%q\n' "$QS_PRIMARY_MONITOR"
    fi
  } > "$ENV_FILE"
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

# Status output still surfaces the legacy units so a user can see whether
# any of them are unexpectedly running. The list shrinks as PRs 2/3 prune
# the unit files themselves; for now it tracks what install.sh ships.
service_units() {
  printf '%s\n' \
    dotfiles-quickshell.service \
    swaync.service
}

print_status() {
  local unit state
  printf '%-42s %s\n' "shell-stack-mode" "$(mode_or_default)"
  while IFS= read -r unit; do
    state="$(service_state "$unit")"
    printf '%-42s %s\n' "$unit" "${state:-unknown}"
  done < <(service_units)
}

print_status_json() {
  local unit state first=true
  local saved_mode
  saved_mode="$(mode_or_default)"
  printf '{"mode":"status","shell_mode":"%s","services":[' \
    "$(json_escape "$saved_mode")"
  while IFS= read -r unit; do
    state="$(service_state "$unit")"
    if $first; then
      first=false
    else
      printf ','
    fi
    printf '{"unit":"%s","state":"%s"}' "$(json_escape "$unit")" "$(json_escape "${state:-unknown}")"
  done < <(service_units)
  printf ']}\n'
}

start_unit() { run_cmd systemctl --user start "$1"; }
restart_unit() { run_cmd systemctl --user restart "$1"; }
stop_unit() { run_cmd systemctl --user stop "$1"; }

case "$MODE" in
  status)
    if $JSON_MODE; then
      print_status_json
    else
      print_status
    fi
    ;;
  full-cutover)
    persist_mode "$MODE"
    restart_unit dotfiles-quickshell.service
    ;;
  rollback)
    persist_mode "$MODE"
    stop_unit dotfiles-quickshell.service
    ;;
esac
