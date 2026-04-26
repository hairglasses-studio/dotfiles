#!/usr/bin/env bash
# ticker-control.sh — Quickshell ticker IPC wrapper.

set -euo pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
dotfiles_dir="$(cd "$script_dir/.." && pwd)"
config_path="${QUICKSHELL_CONFIG_PATH:-$dotfiles_dir/quickshell/shell.qml}"
quickshell_bin="${QUICKSHELL_BIN:-quickshell}"
ipc_timeout="${QUICKSHELL_IPC_TIMEOUT:-0.8}"

usage() {
  cat <<'EOF'
Usage: ticker-control.sh <status|next|prev|pin|unpin|pin-toggle|pause|shuffle|playlist|preset|reload|banner|urgent|snooze-urgent|list-streams|list-playlists>

Dispatches to Quickshell's `ticker` IPC target. Quickshell must be running.
EOF
}

qs_call() {
  command -v "$quickshell_bin" >/dev/null 2>&1 || return 1
  timeout "$ipc_timeout" "$quickshell_bin" ipc --path "$config_path" --newest call ticker "$@" >/dev/null 2>&1
}

qs_call_output() {
  command -v "$quickshell_bin" >/dev/null 2>&1 || return 1
  timeout "$ipc_timeout" "$quickshell_bin" ipc --path "$config_path" --newest call ticker "$@"
}

print_human_status() {
  local json="$1"
  jq -r '
    "service   : \(.service // "quickshell")",
    "playlist  : \(.playlist // "main")",
    "current   : \(.current_stream // .stream // "")",
    "pinned    : \((.pinned // "") | if . == "" then "(none)" else . end)",
    "paused    : \(if .paused then "yes" else "no" end)",
    "shuffle   : \(if .shuffle then "on" else "off" end)",
    "urgent    : \(if .urgent then "yes" else "no" end)"
  ' <<<"$json"
}

cmd="${1:-status}"
shift || true

case "$cmd" in
  status)
    json="$(qs_call_output status 2>/dev/null)" || {
      printf 'ticker-control: Quickshell not responding to ticker IPC\n' >&2
      exit 1
    }
    if [[ "${1:-}" == "--json" ]]; then
      printf '%s\n' "$json"
    else
      print_human_status "$json"
    fi
    ;;
  next) qs_call next ;;
  prev) qs_call prev ;;
  pin)
    stream="${1:?usage: ticker-control.sh pin <stream>}"
    qs_call pin "$stream" && printf 'pinned to %s\n' "$stream"
    ;;
  unpin) qs_call unpin && printf 'unpinned\n' ;;
  pin-toggle) qs_call pinToggle ;;
  pause) qs_call pauseToggle ;;
  shuffle)
    mode="${1:-toggle}"
    case "$mode" in
      on|off|toggle|"") qs_call shuffle "$mode" ;;
      *) printf 'usage: ticker-control.sh shuffle [on|off|toggle]\n' >&2; exit 2 ;;
    esac
    ;;
  playlist)
    name="${1:?usage: ticker-control.sh playlist <name>}"
    qs_call playlist "$name" && printf 'playlist switched to %s\n' "$name"
    ;;
  preset)
    name="${1:?usage: ticker-control.sh preset <name>}"
    qs_call preset "$name" && printf 'preset switched to %s\n' "$name"
    ;;
  reload) qs_call reload ;;
  banner)
    text="${1:?usage: ticker-control.sh banner <text> [color]}"
    color="${2:-#29f0ff}"
    qs_call banner "$text" "$color"
    ;;
  urgent)
    enabled="${1:-true}"
    qs_call urgent "$enabled"
    ;;
  snooze-urgent) qs_call snoozeUrgent ;;
  list-streams) qs_call_output listStreams ;;
  list-playlists)
    for f in "$dotfiles_dir/ticker/content-playlists"/*.txt; do
      printf '%s\n' "$(basename "$f" .txt)"
    done
    ;;
  -h|--help|help) usage ;;
  *) usage >&2; exit 2 ;;
esac
