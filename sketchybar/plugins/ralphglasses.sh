#!/bin/bash
# SketchyBar plugin: ralphglasses fleet observability
# Single mcp-call, parsed by rg_parse.py, populates all rg.* items.

PLUGIN_DIR="$(cd "$(dirname "$0")" && pwd)"
RG="$HOME/hairglasses-studio/ralphglasses"

if [ ! -d "$RG" ]; then
    for item in rg.fleet rg.loops rg.cost rg.models rg.repos rg.iters; do
        sketchybar --set $item drawing=off 2>/dev/null
    done
    exit 0
fi

TMPF=$(mktemp /tmp/rg-sb.XXXXXX)
cd "$RG" && go run . mcp-call ralphglasses_fleet_status 2>/dev/null | python3 "$PLUGIN_DIR/rg_parse.py" > "$TMPF" 2>/dev/null

if [ ! -s "$TMPF" ]; then
    sketchybar --set rg.fleet label="offline" icon.color=0xff686868 label.color=0xff686868
    rm -f "$TMPF"
    exit 0
fi

FLEET_LABEL=$(sed -n '1p' "$TMPF")
FLEET_COLOR=$(sed -n '2p' "$TMPF")
LOOPS_LABEL=$(sed -n '3p' "$TMPF")
COST_LABEL=$(sed -n '4p' "$TMPF")
MODELS_LABEL=$(sed -n '5p' "$TMPF")
REPOS_LABEL=$(sed -n '6p' "$TMPF")
ITERS_LABEL=$(sed -n '7p' "$TMPF")
rm -f "$TMPF"

sketchybar \
    --set rg.fleet  drawing=on label="$FLEET_LABEL"  icon.color="$FLEET_COLOR" label.color="$FLEET_COLOR" \
    --set rg.loops  drawing=on label="$LOOPS_LABEL" \
    --set rg.cost   drawing=on label="$COST_LABEL" \
    --set rg.models drawing=on label="$MODELS_LABEL" \
    --set rg.repos  drawing=on label="$REPOS_LABEL" \
    --set rg.iters  drawing=on label="$ITERS_LABEL"
