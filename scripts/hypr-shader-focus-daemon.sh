#!/usr/bin/env bash
# hypr-shader-focus-daemon.sh — focus-driven hyprshade switcher.
#
# Listens to Hyprland socket2 `activewindow>>` events and picks a screen shader
# based on the focused window's app class. Lightweight, single-process,
# reacts within ~100ms of focus change.
#
# Rules live in a simple shader map; unknown classes fall through to the
# time-of-day schedule owned by hyprshade.
#
# Usage:
#   hypr-shader-focus-daemon.sh              # run in foreground (for systemd)
#   hypr-shader-focus-daemon.sh --dry-run    # print decisions, don't apply
#
# Env:
#   HYPR_SHADER_FOCUS_DISABLE=1              # bypass entirely (no-op)
#
# Exit codes:
#   0 — clean shutdown
#   1 — missing dependency (socat, hyprshade, hyprctl)
#   2 — Hyprland socket missing (not running or not yet ready)

set -euo pipefail

DRY_RUN=false
[[ "${1:-}" == "--dry-run" ]] && DRY_RUN=true

if [[ "${HYPR_SHADER_FOCUS_DISABLE:-0}" == "1" ]]; then
    echo "focus daemon disabled via HYPR_SHADER_FOCUS_DISABLE=1"
    exit 0
fi

for cmd in hyprctl hyprshade socat; do
    command -v "$cmd" >/dev/null 2>&1 || { echo "$cmd not found" >&2; exit 1; }
done

SOCK="$XDG_RUNTIME_DIR/hypr/${HYPRLAND_INSTANCE_SIGNATURE:-$(ls "$XDG_RUNTIME_DIR/hypr" 2>/dev/null | head -1)}/.socket2.sock"
[[ -S "$SOCK" ]] || { echo "socket2 not found: $SOCK" >&2; exit 2; }

# Shader map — class prefix → shader name from ~/.config/hypr/shaders/.
# Returns empty to fall through to hyprshade's time-of-day schedule.
shader_for_class() {
    case "${1,,}" in
        firefox|chrom*|brave*)    echo "" ;;             # reader: no scanlines
        code|cursor|claude*|electron*) echo "" ;;        # dev: no scanlines
        kitty|foot|alacritty|ghostty|wezterm) echo "cyberpunk-vignette" ;;
        mpv|vlc|celluloid)        echo "" ;;             # video: no post-fx
        discord|slack|*chat*)     echo "" ;;             # social: plain
        *)                         echo "" ;;
    esac
}

apply_shader() {
    local want="$1" current
    current="$(hyprshade current 2>/dev/null || true)"
    if [[ "$want" == "$current" ]]; then
        return 0
    fi
    if $DRY_RUN; then
        echo "dry-run: would switch '$current' → '${want:-<none>}'"
        return 0
    fi
    if [[ -z "$want" ]]; then
        hyprshade auto >/dev/null 2>&1 || true     # restore time-of-day schedule
    else
        hyprshade on "$want" >/dev/null 2>&1 || true
    fi
}

# Consume events; activewindow event format: `activewindow>>class,title`
socat -U - UNIX-CONNECT:"$SOCK" | while IFS= read -r line; do
    case "$line" in
        activewindow\>\>*)
            payload="${line#activewindow>>}"
            class="${payload%%,*}"
            shader="$(shader_for_class "$class")"
            apply_shader "$shader"
            ;;
    esac
done
