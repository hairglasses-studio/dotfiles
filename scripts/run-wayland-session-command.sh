#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=scripts/lib/runtime-desktop-env.sh
source "$SCRIPT_DIR/lib/runtime-desktop-env.sh"

wait_secs="${WAYLAND_COMMAND_WAIT_SECS:-15}"

usage() {
  printf 'Usage: %s [--check|--print-env|--] [command...]\n' "${0##*/}" >&2
}

mode="${1:-run}"
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
    printf '%s: no live Wayland socket after %ss\n' "${0##*/}" "$wait_secs" >&2
    exit 1
    ;;
  --)
    shift
    ;;
  run|--run)
    shift
    ;;
esac

if [[ "$#" -eq 0 ]]; then
  usage
  exit 2
fi

if ! wait_for_wayland "$wait_secs"; then
  printf '%s: no live Wayland socket after %ss\n' "${0##*/}" "$wait_secs" >&2
  exit 1
fi

exec "$@"
