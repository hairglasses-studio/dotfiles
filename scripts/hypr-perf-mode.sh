#!/usr/bin/env bash
# hypr-perf-mode.sh — toggle Hyprland quality vs. performance profile at runtime.
#
# Uses hyprctl keyword to flip visual effects without editing config files.
# Survives until next Hyprland reload, so safe for A/B testing.
#
# Modes:
#   quality    — current committed settings (blur + shadows + animations + shaders)
#   performance — 240Hz-tuned: reduced blur, no shadows, VFR on, shaders permitted
#                 but DarkWindow animation paused via vfr gate
#   auto       — toggle between the two
#
# Usage:
#   hypr-perf-mode.sh [quality|performance|auto|status]
#
# Keybind (see hyprland.conf): $mod CTRL ALT, Q

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/hg-core.sh
source "$SCRIPT_DIR/lib/hg-core.sh"

STATE_FILE="${XDG_STATE_HOME:-$HOME/.local/state}/hypr-perf-mode"
mkdir -p "$(dirname "$STATE_FILE")"

current_mode() {
    [[ -f "$STATE_FILE" ]] && cat "$STATE_FILE" || echo quality
}

apply_quality() {
    hyprctl --batch "\
        keyword decoration:blur:enabled true; \
        keyword decoration:blur:size 6; \
        keyword decoration:blur:passes 2; \
        keyword decoration:blur:vibrancy 0.22; \
        keyword decoration:shadow:enabled true; \
        keyword decoration:shadow:range 36; \
        keyword animations:enabled true; \
        keyword misc:vfr false; \
        keyword debug:overlay false; \
    " >/dev/null
    echo quality > "$STATE_FILE"
    hg_notify_low "Perf Mode" "Quality (blur + shadow + shaders, vfr off)"
    hg_ok "quality mode applied"
}

apply_performance() {
    # VFR on: compositor pauses rendering when idle. This freezes DarkWindow
    # animated shaders but massively cuts GPU load at 240Hz.
    hyprctl --batch "\
        keyword decoration:blur:enabled true; \
        keyword decoration:blur:size 4; \
        keyword decoration:blur:passes 1; \
        keyword decoration:blur:vibrancy 0.15; \
        keyword decoration:shadow:enabled false; \
        keyword animations:enabled true; \
        keyword misc:vfr true; \
        keyword debug:overlay true; \
    " >/dev/null
    echo performance > "$STATE_FILE"
    hg_notify_low "Perf Mode" "Performance (reduced blur, no shadow, vfr on, overlay on)"
    hg_ok "performance mode applied — watch nvtop for idle GPU drop"
}

show_status() {
    local mode
    mode="$(current_mode)"
    echo "mode: $mode"
    echo "---"
    echo "live hyprctl options:"
    for opt in decoration:blur:enabled decoration:blur:size decoration:blur:passes \
               decoration:shadow:enabled misc:vfr misc:vrr debug:overlay \
               render:direct_scanout render:explicit_sync debug:damage_tracking; do
        printf "  %-32s = %s\n" "$opt" \
            "$(hyprctl getoption "$opt" -j 2>/dev/null | jq -r '.int // .str // .custom // "?"' 2>/dev/null || echo '?')"
    done
    echo "---"
    echo "monitor VRR state:"
    hyprctl monitors -j 2>/dev/null | jq -r '.[] | "  \(.name): \(.refreshRate)Hz  vrr=\(.vrr // "off")"' 2>/dev/null || true
}

case "${1:-auto}" in
    quality|q)      apply_quality ;;
    performance|p|perf) apply_performance ;;
    auto|toggle)
        if [[ "$(current_mode)" == quality ]]; then
            apply_performance
        else
            apply_quality
        fi
        ;;
    status|s)       show_status ;;
    *)              hg_die "Usage: hypr-perf-mode.sh [quality|performance|auto|status]" ;;
esac
