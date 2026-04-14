#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=scripts/lib/runtime-desktop-env.sh
source "$SCRIPT_DIR/lib/runtime-desktop-env.sh"

wait_secs="${IRONBAR_WAIT_SECS:-15}"
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
    printf 'run-ironbar: no live Wayland socket after %ss\n' "$wait_secs" >&2
    exit 1
    ;;
  run|--run)
    ;;
  *)
    printf 'Usage: %s [--check|--print-env]\n' "${0##*/}" >&2
    exit 2
    ;;
esac

if ! wait_for_wayland "$wait_secs"; then
  printf 'run-ironbar: no live Wayland socket after %ss\n' "$wait_secs" >&2
  exit 1
fi

if [[ -n "${XDG_RUNTIME_DIR:-}" ]]; then
  rm -f "${XDG_RUNTIME_DIR%/}/ironbar-ipc.sock"
fi

exec /usr/bin/env ironbar \
  -c "${IRONBAR_CONFIG_PATH:-$HOME/.config/ironbar/config.toml}" \
  -t "${IRONBAR_STYLE_PATH:-$HOME/.config/ironbar/style.css}"
