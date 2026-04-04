#!/usr/bin/env bash
# agent-post-tool-reload.sh — Post-edit reload hook for agent-driven file edits.
# Reads JSON from stdin, extracts a candidate file path, reloads known services,
# and emits non-blocking palette warnings.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/config.sh" 2>/dev/null || exit 0

extract_file_path() {
  local input="$1"
  local match

  match="$(
    printf '%s' "$input" |
      tr '\n' ' ' |
      grep -oE '"(file_path|path)"[[:space:]]*:[[:space:]]*"[^"]*"' |
      head -1 || true
  )"

  [[ -n "$match" ]] || return 1
  printf '%s\n' "$match" | sed -E 's/^"(file_path|path)"[[:space:]]*:[[:space:]]*"//; s/"$//'
}

input="$(cat)"
file_path="$(extract_file_path "$input" || true)"
[[ -n "$file_path" ]] || exit 0

case "$file_path" in
  */hyprland/* | */hypr/*)  config_reload_service hyprland || echo "[hook] hyprland reload failed" >&2 ;;
  */mako/*)                 config_reload_service mako     || echo "[hook] mako reload failed" >&2 ;;
  */eww/*)                  config_reload_service eww      || echo "[hook] eww reload failed" >&2 ;;
  */waybar/*)               config_reload_service waybar   || echo "[hook] waybar reload failed" >&2 ;;
  */sway/*)                 config_reload_service sway     || echo "[hook] sway reload failed" >&2 ;;
  */tmux/*)                 config_reload_service tmux     || echo "[hook] tmux reload failed" >&2 ;;
esac

case "$file_path" in
  *.conf | *.toml | *.scss | *.yuck | *.ini)
    if [[ -f "$file_path" ]]; then
      snazzy_palette="000000 1a1a1a 1a1b26 ff5c57 5af78e f3f99d 57c7ff ff6ac1 9aedfe f1f1f0 686868 eff0eb d4d4d4 264f78 ffffff f8f8f2 282a36"
      hex_colors="$(grep -oiE '#[0-9a-fA-F]{6}' "$file_path" 2>/dev/null | sed 's/#//' | tr '[:upper:]' '[:lower:]' | sort -u)"
      for color in $hex_colors; do
        match=false
        for ok in $snazzy_palette; do
          if [[ "$color" == "$ok" ]]; then
            match=true
            break
          fi
        done
        if [[ "$match" == false ]]; then
          echo "[palette] non-Snazzy color #$color in $file_path" >&2
        fi
      done
    fi
    ;;
esac

exit 0
