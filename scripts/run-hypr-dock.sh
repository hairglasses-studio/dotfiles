#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=scripts/lib/runtime-desktop-env.sh
source "$SCRIPT_DIR/lib/runtime-desktop-env.sh"

wait_secs="${HYPR_DOCK_WAIT_SECS:-15}"
mode="${1:-run}"

usage() {
  printf 'Usage: %s [--check|--print-env|--] [hypr-dock args...]\n' "${0##*/}" >&2
}

case "$mode" in
  --print-env)
    refresh_desktop_runtime_env
    print_desktop_runtime_env
    exit 0
    ;;
  --check)
    if wait_for_wayland "$wait_secs"; then
      print_desktop_runtime_env
      exit 0
    fi
    printf 'run-hypr-dock: no live Wayland socket after %ss\n' "$wait_secs" >&2
    exit 1
    ;;
  --)
    shift
    ;;
  run|--run)
    shift
    ;;
  *)
    usage
    exit 2
    ;;
esac

if ! wait_for_wayland "$wait_secs"; then
  printf 'run-hypr-dock: no live Wayland socket after %ss\n' "$wait_secs" >&2
  exit 1
fi

state_root="${XDG_STATE_HOME:-$HOME/.local/state}/hypr-dock"
mkdir -p "$state_root"
raw_log_path="${HYPR_DOCK_STDERR_LOG_PATH:-$state_root/hypr-dock.stderr.log}"
touch "$raw_log_path"

stderr_fifo="$(mktemp -u "${XDG_RUNTIME_DIR:-/tmp}/hypr-dock-stderr.XXXXXX")"
cleanup() {
  rm -f "$stderr_fifo"
}
trap cleanup EXIT
mkfifo "$stderr_fifo"

{
  while IFS= read -r line; do
    printf '[%s] %s\n' "$(date '+%F %T')" "$line" >>"$raw_log_path"
    case "$line" in
      *'Error reading desktop file: error="open : no such file or directory"'*) continue ;;
    esac
    printf '%s\n' "$line"
  done <"$stderr_fifo"
} &
filter_pid=$!

cmd=(hypr-dock)
if [[ "$#" -gt 0 ]]; then
  cmd+=("$@")
else
  cmd+=(-config "$HOME/.config/hypr-dock/hypr-dock.conf")
fi

set +e
"${cmd[@]}" >"$stderr_fifo" 2>&1
status=$?
set -e

wait "$filter_pid" || true
exit "$status"
