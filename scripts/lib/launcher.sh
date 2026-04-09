#!/usr/bin/env bash

# Shared launcher helpers for Hyprland-facing fallback surfaces.

launcher_hyprshell_running() {
  command -v hyprshell >/dev/null 2>&1 && pgrep -x hyprshell >/dev/null 2>&1
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
