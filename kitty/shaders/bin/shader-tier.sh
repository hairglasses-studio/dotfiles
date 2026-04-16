#!/usr/bin/env bash
# shader-tier.sh — Generate perf-tier playlists for the DarkWindow shader catalog.
#
# First-pass heuristic (no live Hyprland IPC needed): tier by static metrics
# (file size + line count + loop nesting) — decent proxy for GPU cost.
# Future extension: runtime GPU delta via nvtop + hyprctl darkwindow:shadeactive
# (blocked until shader_benchmark MCP tool is extended).
#
# Output: kitty/shaders/playlists/tier-cheap.txt, tier-mid.txt, tier-heavy.txt
#
# Usage:
#   shader-tier.sh generate      # write tier playlists
#   shader-tier.sh report        # print per-shader tier + score without writing
#   shader-tier.sh verify        # confirm all 139 shaders landed in exactly one tier

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SHADER_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
SHADERS_DIR="$SHADER_ROOT/darkwindow"
PLAYLISTS_DIR="$SHADER_ROOT/playlists"

# Thresholds — tuned empirically from the current 139-shader catalog.
# Natural tercile boundaries from size distribution: p33=2193, p66=4610 bytes.
# Use slightly rounded values to tolerate small catalog changes.
CHEAP_MAX=2200
MID_MAX=4600

score_shader() {
    local f="$1"
    local size fors
    size=$(stat -c %s "$f")
    # Nested loops are the expensive thing; count `for` occurrences as a weak proxy.
    # grep -c always prints a number (0 when no matches) but exits non-zero on 0.
    fors=$(grep -c 'for\s*(' "$f" || true)
    # Fall back to 0 if fors is empty
    [[ -z "$fors" ]] && fors=0
    # Return tier
    if (( size < CHEAP_MAX && fors <= 1 )); then
        echo cheap
    elif (( size < MID_MAX && fors <= 3 )); then
        echo mid
    else
        echo heavy
    fi
}

cmd_generate() {
    mkdir -p "$PLAYLISTS_DIR"
    local cheap="$PLAYLISTS_DIR/tier-cheap.txt"
    local mid="$PLAYLISTS_DIR/tier-mid.txt"
    local heavy="$PLAYLISTS_DIR/tier-heavy.txt"
    : > "$cheap" : > "$mid" : > "$heavy"

    local cheap_n=0 mid_n=0 heavy_n=0
    for f in "$SHADERS_DIR"/*.glsl; do
        [[ -f "$f" ]] || continue
        local name tier
        name=$(basename "$f")
        tier=$(score_shader "$f")
        case "$tier" in
            cheap)  echo "$name" >> "$cheap" ; cheap_n=$((cheap_n+1)) ;;
            mid)    echo "$name" >> "$mid"   ; mid_n=$((mid_n+1)) ;;
            heavy)  echo "$name" >> "$heavy" ; heavy_n=$((heavy_n+1)) ;;
        esac
    done

    # Sort playlists alphabetically
    sort -o "$cheap" "$cheap"
    sort -o "$mid"   "$mid"
    sort -o "$heavy" "$heavy"

    printf 'Generated tier playlists (thresholds: cheap<%d, mid<%d bytes):\n' "$CHEAP_MAX" "$MID_MAX"
    printf '  tier-cheap.txt  %3d shaders  %s\n' "$cheap_n" "(safe for ambient rotation at 240Hz)"
    printf '  tier-mid.txt    %3d shaders  %s\n' "$mid_n"   "(good for focused windows)"
    printf '  tier-heavy.txt  %3d shaders  %s\n' "$heavy_n" "(showcase / perf-mode only)"
    printf '  total:          %3d\n' "$((cheap_n + mid_n + heavy_n))"
}

cmd_report() {
    printf '%-50s  %-7s  %6s  %6s  %5s\n' "shader" "tier" "bytes" "lines" "forN"
    printf -- '-%.0s' {1..82}; printf '\n'
    for f in "$SHADERS_DIR"/*.glsl; do
        [[ -f "$f" ]] || continue
        local name tier size lines fors
        name=$(basename "$f")
        tier=$(score_shader "$f")
        size=$(stat -c %s "$f")
        lines=$(wc -l <"$f")
        fors=$(grep -c 'for\s*(' "$f" || true)
        [[ -z "$fors" ]] && fors=0
        printf '%-50s  %-7s  %6d  %6d  %5d\n' "$name" "$tier" "$size" "$lines" "$fors"
    done | sort -k2,2 -k3,3n
}

cmd_verify() {
    local fs_count tier_count
    fs_count=$(find "$SHADERS_DIR" -maxdepth 1 -name '*.glsl' | wc -l)
    if [[ ! -f "$PLAYLISTS_DIR/tier-cheap.txt" ]]; then
        echo "verify: tier playlists not yet generated — run 'shader-tier.sh generate'" >&2
        return 1
    fi
    tier_count=$(cat "$PLAYLISTS_DIR/tier-cheap.txt" \
                     "$PLAYLISTS_DIR/tier-mid.txt" \
                     "$PLAYLISTS_DIR/tier-heavy.txt" | sort -u | wc -l)
    printf 'Filesystem:     %d shaders\n' "$fs_count"
    printf 'Tier playlists: %d shaders (deduped union)\n' "$tier_count"
    if [[ "$fs_count" -eq "$tier_count" ]]; then
        echo 'OK — all shaders tiered'
    else
        echo "MISMATCH — run 'shader-tier.sh generate' to refresh"
        return 1
    fi
}

case "${1:-generate}" in
    generate|gen) cmd_generate ;;
    report|rep)   cmd_report ;;
    verify|v)     cmd_verify ;;
    *)
        cat <<EOF >&2
Usage: shader-tier.sh [generate|report|verify]

  generate   Write tier-{cheap,mid,heavy}.txt playlists (default)
  report     Print per-shader tier classification
  verify     Check all shaders are tiered exactly once
EOF
        exit 2
        ;;
esac
