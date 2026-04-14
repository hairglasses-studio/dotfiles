#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=scripts/lib/runtime-desktop-env.sh
source "$SCRIPT_DIR/lib/runtime-desktop-env.sh"

wait_secs="${IRONBAR_WAIT_SECS:-15}"
audio_wait_secs="${IRONBAR_AUDIO_WAIT_SECS:-10}"
mode="${1:-run}"

default_audio_sink_ready() {
  if ! command -v wpctl >/dev/null 2>&1; then
    return 0
  fi
  wpctl get-volume @DEFAULT_AUDIO_SINK@ >/dev/null 2>&1
}

wait_for_default_audio_sink() {
  local wait_secs="${1:-10}"
  local waited=0
  while (( waited <= wait_secs )); do
    if default_audio_sink_ready; then
      return 0
    fi
    sleep 1
    ((waited += 1))
  done
  return 1
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

wait_for_default_audio_sink "$audio_wait_secs" || true

state_root="${XDG_STATE_HOME:-$HOME/.local/state}/ironbar"
mkdir -p "$state_root"
raw_log_path="${IRONBAR_STDERR_LOG_PATH:-$state_root/ironbar.stderr.log}"
touch "$raw_log_path"

export IRONBAR_LOG="${IRONBAR_LOG:-warn}"
export IRONBAR_FILE_LOG="${IRONBAR_FILE_LOG:-warn}"

if [[ -n "${XDG_RUNTIME_DIR:-}" ]]; then
  rm -f "${XDG_RUNTIME_DIR%/}/ironbar-ipc.sock"
fi

stderr_fifo="$(mktemp -u "${XDG_RUNTIME_DIR:-/tmp}/ironbar-stderr.XXXXXX")"
cleanup() {
  rm -f "$stderr_fifo"
}
trap cleanup EXIT
mkfifo "$stderr_fifo"

{
  while IFS= read -r line; do
    printf '[%s] %s\n' "$(date '+%F %T')" "$line" >>"$raw_log_path"
    case "$line" in
      *"Unable to locate workspace"*) continue ;;
      *"Unable to locate client"*) continue ;;
      *"Couldn't find sink: "* ) continue ;;
      *"Attempted to update menu at "* ) continue ;;
      *"ignoring workspace event received before initialization"*) continue ;;
      *"[Gdk] MESSAGE: Vulkan: Loader Message:"*) continue ;;
      *"[Gdk] vkQueuePresentKHR(): A swapchain no longer matches the surface properties exactly, but can still be used to present to the surface successfully. (VK_SUBOPTIMAL_KHR)"*) continue ;;
      *"[Gdk] vkAcquireNextImageKHR(): A swapchain no longer matches the surface properties exactly, but can still be used to present to the surface successfully. (VK_SUBOPTIMAL_KHR)"*) continue ;;
    esac
    printf '%s\n' "$line"
  done <"$stderr_fifo"
} &
filter_pid=$!

set +e
/usr/bin/env ironbar \
  -c "${IRONBAR_CONFIG_PATH:-$HOME/.config/ironbar/config.toml}" \
  -t "${IRONBAR_STYLE_PATH:-$HOME/.config/ironbar/style.css}" \
  >"$stderr_fifo" 2>&1
status=$?
set -e

wait "$filter_pid" || true
exit "$status"
