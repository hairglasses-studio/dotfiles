#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=scripts/lib/runtime-desktop-env.sh
source "$SCRIPT_DIR/lib/runtime-desktop-env.sh"

wait_secs="${QUICKSHELL_WAIT_SECS:-30}"
mode="${1:-run}"
dotfiles_dir="$(cd "$SCRIPT_DIR/.." && pwd)"
shell_state_env="${XDG_STATE_HOME:-$HOME/.local/state}/dotfiles/shell-stack/env"
if [[ -f "$shell_state_env" ]]; then
  # shellcheck disable=SC1090
  source "$shell_state_env"
fi
config_path="${QUICKSHELL_CONFIG_PATH:-$dotfiles_dir/quickshell/shell.qml}"
palette_script="${QUICKSHELL_PALETTE_SCRIPT:-$SCRIPT_DIR/palette-propagate.sh}"
colors_path="${QUICKSHELL_COLORS_PATH:-$dotfiles_dir/quickshell/styles/Colors.qml}"
target_monitor="${QUICKSHELL_MONITOR:-${QS_PRIMARY_MONITOR:-DP-2}}"

ensure_palette_tokens() {
  [[ -x "$palette_script" ]] || return 0
  "$palette_script" --no-reload >/dev/null 2>&1 || return 0
  [[ -f "$colors_path" ]]
}

case "$mode" in
  --print-env)
    refresh_desktop_runtime_env
    print_desktop_runtime_env
    printf 'QUICKSHELL_CONFIG_PATH=%s\n' "$config_path"
    printf 'QUICKSHELL_MONITOR=%s\n' "$target_monitor"
    printf 'SHELL_STACK_MODE=%s\n' "${SHELL_STACK_MODE:-pilot}"
    printf 'QS_BAR_CUTOVER=%s\n' "${QS_BAR_CUTOVER:-0}"
    printf 'QS_TICKER_CUTOVER=%s\n' "${QS_TICKER_CUTOVER:-0}"
    printf 'QS_MENU_CUTOVER=%s\n' "${QS_MENU_CUTOVER:-0}"
    printf 'QS_DOCK_CUTOVER=%s\n' "${QS_DOCK_CUTOVER:-0}"
    printf 'QS_COMPANION_CUTOVER=%s\n' "${QS_COMPANION_CUTOVER:-0}"
    printf 'QUICKSHELL_NOTIFICATION_OWNER=%s\n' "${QUICKSHELL_NOTIFICATION_OWNER:-0}"
    exit 0
    ;;
  --check)
    if ! command -v quickshell >/dev/null 2>&1; then
      printf 'run-quickshell: quickshell not found on PATH\n' >&2
      exit 1
    fi
    if ! wait_for_wayland "$wait_secs"; then
      printf 'run-quickshell: no live Wayland socket after %ss\n' "$wait_secs" >&2
      exit 1
    fi
    if ! ensure_palette_tokens; then
      printf 'run-quickshell: missing generated palette tokens: %s\n' "$colors_path" >&2
      exit 1
    fi
    [[ -f "$config_path" ]] || {
      printf 'run-quickshell: missing config: %s\n' "$config_path" >&2
      exit 1
    }
    exit 0
    ;;
  run|--run)
    ;;
  *)
    printf 'Usage: %s [--check|--print-env]\n' "${0##*/}" >&2
    exit 2
    ;;
esac

if ! command -v quickshell >/dev/null 2>&1; then
  printf 'run-quickshell: quickshell not found on PATH\n' >&2
  exit 1
fi

if ! wait_for_wayland "$wait_secs"; then
  printf 'run-quickshell: no live Wayland socket after %ss\n' "$wait_secs" >&2
  exit 1
fi

ensure_palette_tokens || true

export QS_MONITOR="$target_monitor"
export SHELL_STACK_MODE="${SHELL_STACK_MODE:-pilot}"
export QS_BAR_CUTOVER="${QS_BAR_CUTOVER:-0}"
export QS_TICKER_CUTOVER="${QS_TICKER_CUTOVER:-0}"
export QS_MENU_CUTOVER="${QS_MENU_CUTOVER:-0}"
export QS_DOCK_CUTOVER="${QS_DOCK_CUTOVER:-0}"
export QS_COMPANION_CUTOVER="${QS_COMPANION_CUTOVER:-0}"
export QUICKSHELL_NOTIFICATION_OWNER="${QUICKSHELL_NOTIFICATION_OWNER:-0}"
exec /usr/bin/env quickshell --path "$config_path"
