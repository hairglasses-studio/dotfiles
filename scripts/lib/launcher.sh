#!/usr/bin/env bash

# Shared launcher helpers for Hyprland-facing fallback surfaces.

launcher_lib_dir="$(cd "$(dirname "$(readlink -f "${BASH_SOURCE[0]:-$0}")")" && pwd)"
runtime_desktop_env_sh="$launcher_lib_dir/runtime-desktop-env.sh"

if [[ -f "$runtime_desktop_env_sh" ]]; then
  # shellcheck source=/dev/null
  source "$runtime_desktop_env_sh"
fi

launcher_refresh_desktop_env() {
  if declare -F refresh_desktop_runtime_env >/dev/null 2>&1; then
    refresh_desktop_runtime_env
  fi
}

launcher_truthy() {
  local value="${1:-}"
  case "${value,,}" in
    1|true|yes|on)
      return 0
      ;;
  esac
  return 1
}

launcher_prefer_hyprshell() {
  launcher_truthy "${DOTFILES_LAUNCHER_PREFER_HYPRSHELL:-0}"
}

launcher_hyprshell_socket_path() {
  launcher_refresh_desktop_env

  if [[ -n "${XDG_RUNTIME_DIR:-}" && -n "${HYPRLAND_INSTANCE_SIGNATURE:-}" ]]; then
    local instance_socket="${XDG_RUNTIME_DIR%/}/hypr/${HYPRLAND_INSTANCE_SIGNATURE}/hyprshell.sock"
    if [[ -e "$instance_socket" ]]; then
      printf '%s\n' "$instance_socket"
      return 0
    fi
  fi

  if [[ -n "${XDG_RUNTIME_DIR:-}" ]]; then
    local runtime_socket="${XDG_RUNTIME_DIR%/}/hyprshell.sock"
    if [[ -e "$runtime_socket" ]]; then
      printf '%s\n' "$runtime_socket"
      return 0
    fi
  fi

  return 1
}

launcher_hyprshell_running() {
  launcher_refresh_desktop_env
  command -v hyprshell >/dev/null 2>&1 \
    && pgrep -x hyprshell >/dev/null 2>&1 \
    && launcher_hyprshell_socket_path >/dev/null 2>&1
}

launcher_hyprshell_socat() {
  local payload="$1"

  launcher_refresh_desktop_env
  command -v hyprshell >/dev/null 2>&1 || return 1
  pgrep -x hyprshell >/dev/null 2>&1 || return 1
  launcher_hyprshell_socket_path >/dev/null 2>&1 || return 1

  hyprshell socat "$payload" >/dev/null 2>&1
}

launcher_hyprshell_layer_visible() {
  local namespace_pattern="${1:-hyprshell_(overview|launcher|switch)}"

  launcher_refresh_desktop_env
  command -v hyprctl >/dev/null 2>&1 || return 1
  command -v jq >/dev/null 2>&1 || return 1

  hyprctl layers -j 2>/dev/null | jq -e --arg pattern "$namespace_pattern" '
    .. | objects | select((.namespace // "") | test($pattern))
  ' >/dev/null 2>&1
}

launcher_wait_hyprshell_layer() {
  local namespace_pattern="$1"
  local attempts="${2:-10}"
  local sleep_secs="${3:-0.05}"
  local i

  for (( i = 0; i < attempts; i += 1 )); do
    if launcher_hyprshell_layer_visible "$namespace_pattern"; then
      return 0
    fi
    sleep "$sleep_secs"
  done

  return 1
}

launcher_wofi_geometry() {
  local width=860
  local height=620
  local monitor=""

  if command -v hyprctl >/dev/null 2>&1 && command -v jq >/dev/null 2>&1; then
    local focused name logical_width logical_height scale
    focused="$(
      hyprctl -j monitors 2>/dev/null | jq -r '
        map(select(.focused)) | first |
        [(.name // ""), (.width // 0), (.height // 0), (.scale // 1)] | @tsv
      ' 2>/dev/null || true
    )"

    if [[ -n "$focused" ]]; then
      IFS=$'\t' read -r name logical_width logical_height scale <<<"$focused"
      if [[ -n "$name" && "$logical_width" != "0" && "$logical_height" != "0" ]]; then
        read -r width height < <(
          awk -v w="$logical_width" -v h="$logical_height" -v s="$scale" '
            function clamp(v, lo, hi) {
              return v < lo ? lo : (v > hi ? hi : v)
            }
            BEGIN {
              effective_w = w * s
              effective_h = h * s
              printf "%d %d\n",
                clamp(int((effective_w * 0.18) + 0.5), 760, 1120),
                clamp(int((effective_h * 0.58) + 0.5), 520, 860)
            }
          '
        )
        monitor="$name"
      fi
    fi
  fi

  printf '%s\t%s\t%s\n' "$width" "$height" "$monitor"
}

launcher_wofi_config_dir() {
  printf '%s\n' "${XDG_CONFIG_HOME:-$HOME/.config}/wofi"
}
