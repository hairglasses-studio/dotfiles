#!/usr/bin/env bash
set -euo pipefail

# ── Ghostty Shader Picker ──────────────────────────
# Interactive shader audition: swaps custom-shader in your
# Ghostty config, reloads, and lets you keep/skip each one.
#
# Usage:
#   bash pick-shaders.sh                  # audit all shaders
#   bash pick-shaders.sh --category CRT   # audit only CRT shaders
#   bash pick-shaders.sh --resume         # resume from last session
#
# Outputs a curated list of config lines at the end.

SHADERS_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CONFIG_FILE="$HOME/dotfiles/ghostty/config"
PICKS_FILE="/tmp/ghostty-shader-picks.txt"
PROGRESS_FILE="/tmp/ghostty-shader-progress.txt"
CATEGORY_FILTER=""
RESUME=false

# ── Parse args ─────────────────────────────────────
while [[ $# -gt 0 ]]; do
  case "$1" in
    --category) CATEGORY_FILTER="$2"; shift 2 ;;
    --resume)   RESUME=true; shift ;;
    -h|--help)
      echo "Usage: $0 [--category CRT|Cursor|Post-FX|Background|Watercolor] [--resume]"
      exit 0 ;;
    *) echo "Unknown option: $1"; exit 1 ;;
  esac
done

# ── Shader metadata ───────────────────────────────
declare -A SHADER_CAT
declare -A SHADER_COST

classify() {
  local name="$1" cat="$2" cost="$3"
  SHADER_CAT["$name"]="$cat"
  SHADER_COST["$name"]="$cost"
}

# cost: LOW = simple math/single texture, MED = moderate loops/noise, HIGH = nested loops/heavy iteration
# CRT
classify "blue-crt.glsl"               "CRT"        "LOW"
classify "green-crt.glsl"              "CRT"        "LOW"
classify "crt.glsl"                    "CRT"        "LOW"
classify "in-game-crt.glsl"            "CRT"        "MED"
classify "in-game-crt-cursor.glsl"     "CRT"        "MED"
classify "retro-terminal.glsl"         "CRT"        "MED"
classify "crt_glitch.glsl"             "CRT"        "MED"
classify "scanline.glsl"               "CRT"        "LOW"

# Post-FX
classify "dither.glsl"                 "Post-FX"    "LOW"
classify "drunkard.glsl"               "Post-FX"    "LOW"
classify "flicker.glsl"                "Post-FX"    "LOW"
classify "glitchy.glsl"                "Post-FX"    "MED"
classify "glow.glsl"                   "Post-FX"    "LOW"
classify "glow-rgbsplit-twitchy.glsl"  "Post-FX"    "MED"
classify "hexglitch.glsl"              "Post-FX"    "HIGH"
classify "mnoise.glsl"                 "Post-FX"    "MED"
classify "pixels.glsl"                 "Post-FX"    "LOW"
classify "shake.glsl"                  "Post-FX"    "LOW"
classify "tft.glsl"                    "Post-FX"    "LOW"
classify "zoom_and_aberration.glsl"    "Post-FX"    "MED"

# Cursor
classify "blaze_sparks.glsl"           "Cursor"     "MED"
classify "cursor_blaze.glsl"           "Cursor"     "MED"
classify "cursor_blaze_no_trail.glsl"  "Cursor"     "LOW"
classify "cursor_blaze_tapered.glsl"   "Cursor"     "MED"
classify "cursor_border_1.glsl"        "Cursor"     "LOW"
classify "cursor_frozen.glsl"          "Cursor"     "MED"
classify "cursor-glitch.glsl"          "Cursor"     "MED"
classify "cursor_smear.glsl"           "Cursor"     "MED"
classify "cursor_smear_fade.glsl"      "Cursor"     "MED"
classify "cursor_smear_gradient.glsl"  "Cursor"     "MED"
classify "cursor_smear_rainbow.glsl"   "Cursor"     "MED"
classify "cursor_sweep.glsl"           "Cursor"     "LOW"
classify "cursor_synesthaxia.glsl"     "Cursor"     "MED"
classify "cursor_tail.glsl"            "Cursor"     "LOW"
classify "cursor_warp.glsl"            "Cursor"     "LOW"
classify "last_letter_zoom.glsl"       "Cursor"     "LOW"
classify "manga_slash.glsl"            "Cursor"     "MED"
classify "party_sparks.glsl"           "Cursor"     "MED"
classify "ripple_cursor.glsl"          "Cursor"     "LOW"
classify "ripple_rectangle_cursor.glsl" "Cursor"    "LOW"
classify "sparks.glsl"                 "Cursor"     "MED"

# Background (animated — these render every frame)
classify "animated-gradient-shader.glsl" "Background" "MED"
classify "clouds.glsl"                 "Background" "MED"
classify "cubes.glsl"                  "Background" "HIGH"
classify "electric.glsl"               "Background" "HIGH"
classify "electric-modes.glsl"         "Background" "HIGH"
classify "galaxy.glsl"                 "Background" "HIGH"
classify "gears-and-belts.glsl"        "Background" "MED"
classify "gradient-background.glsl"    "Background" "LOW"
classify "inside-the-matrix.glsl"      "Background" "MED"
classify "just-snow.glsl"              "Background" "MED"
classify "matrix-hallway.glsl"         "Background" "MED"
classify "sparks-from-fire.glsl"       "Background" "HIGH"
classify "splatter-fractal.glsl"       "Background" "MED"
classify "starfield-colors.glsl"       "Background" "MED"
classify "starfield.glsl"              "Background" "MED"
classify "water.glsl"                  "Background" "MED"

# Watercolor
classify "graded-wash-bg.glsl"         "Watercolor" "MED"
classify "salt-bg.glsl"                "Watercolor" "MED"
classify "splatter-bg.glsl"            "Watercolor" "MED"
classify "variegated-wash-bg.glsl"     "Watercolor" "MED"
classify "wet-on-wet-bg.glsl"          "Watercolor" "MED"

# ── Build shader list ─────────────────────────────
SHADERS=()
for f in "$SHADERS_DIR"/*.glsl; do
  name="$(basename "$f")"
  if [[ -n "$CATEGORY_FILTER" ]]; then
    cat="${SHADER_CAT[$name]:-unknown}"
    [[ "$cat" != "$CATEGORY_FILTER" ]] && continue
  fi
  SHADERS+=("$name")
done

TOTAL=${#SHADERS[@]}

# ── Resume support ────────────────────────────────
START_INDEX=0
if $RESUME && [[ -f "$PROGRESS_FILE" ]]; then
  START_INDEX=$(cat "$PROGRESS_FILE")
  echo "Resuming from shader $((START_INDEX + 1))/$TOTAL"
fi

# Initialize picks file
if ! $RESUME; then
  > "$PICKS_FILE"
fi

# ── Backup original config line ───────────────────
ORIGINAL_SHADER=$(grep '^custom-shader\s*=' "$CONFIG_FILE" | head -1 || echo "")
ORIGINAL_ANIM=$(grep '^custom-shader-animation\s*=' "$CONFIG_FILE" | head -1 || echo "")

restore_config() {
  if [[ -n "$ORIGINAL_SHADER" ]]; then
    sed -i '' "s|^custom-shader = .*|$ORIGINAL_SHADER|" "$CONFIG_FILE"
  fi
  if [[ -n "$ORIGINAL_ANIM" ]]; then
    sed -i '' "s|^custom-shader-animation = .*|$ORIGINAL_ANIM|" "$CONFIG_FILE"
  fi
}

trap restore_config EXIT

# ── Helper: swap shader in config ─────────────────
swap_shader() {
  local shader_path="$1"
  local needs_anim="$2"
  sed -i '' "s|^custom-shader = .*|custom-shader = $shader_path|" "$CONFIG_FILE"
  sed -i '' "s|^custom-shader-animation = .*|custom-shader-animation = $needs_anim|" "$CONFIG_FILE"
}

needs_animation() {
  local name="$1"
  local cat="${SHADER_CAT[$name]:-unknown}"
  # Background and Watercolor shaders need animation=true to see the effect
  case "$cat" in
    Background|Watercolor) echo "true" ;;
    *) echo "false" ;;
  esac
}

# ── Cost color ────────────────────────────────────
cost_color() {
  case "$1" in
    LOW)  printf "\033[0;32mLOW\033[0m" ;;
    MED)  printf "\033[0;33mMED\033[0m" ;;
    HIGH) printf "\033[0;31mHIGH\033[0m" ;;
    *)    printf "$1" ;;
  esac
}

# ── Main loop ─────────────────────────────────────
printf "\n"
printf "  \033[1mGhostty Shader Picker\033[0m\n"
printf "  ─────────────────────\n"
printf "  Total shaders to review: %d\n" "$TOTAL"
if [[ -n "$CATEGORY_FILTER" ]]; then
  printf "  Category filter: %s\n" "$CATEGORY_FILTER"
fi
printf "\n"
printf "  Controls:\n"
printf "    \033[1my\033[0m = keep    \033[1mn\033[0m = skip    \033[1ms\033[0m = skip (save for later)\n"
printf "    \033[1mq\033[0m = quit (saves progress)    \033[1mr\033[0m = re-view current\n"
printf "\n"
printf "  The shader will be hot-swapped in your config.\n"
printf "  Ghostty auto-reloads on config change.\n"
printf "\n"

KEPT=0
SKIPPED=0

for (( i=START_INDEX; i<TOTAL; i++ )); do
  name="${SHADERS[$i]}"
  cat="${SHADER_CAT[$name]:-unknown}"
  cost="${SHADER_COST[$name]:-?}"
  anim=$(needs_animation "$name")

  # Swap shader in config (Ghostty hot-reloads)
  swap_shader "$SHADERS_DIR/$name" "$anim"

  printf "  [\033[1m%d/%d\033[0m] \033[1;36m%-40s\033[0m  %s  cost:" "$((i+1))" "$TOTAL" "$name" "$cat"
  cost_color "$cost"
  printf "\n"

  while true; do
    printf "  → (y/n/s/r/q): "
    read -rsn1 choice
    printf "%s\n" "$choice"
    case "$choice" in
      y|Y)
        echo "$name" >> "$PICKS_FILE"
        KEPT=$((KEPT + 1))
        printf "  \033[0;32m✓ kept\033[0m\n\n"
        break ;;
      n|N)
        SKIPPED=$((SKIPPED + 1))
        printf "  \033[0;31m✗ skipped\033[0m\n\n"
        break ;;
      s|S)
        echo "# MAYBE: $name" >> "$PICKS_FILE"
        SKIPPED=$((SKIPPED + 1))
        printf "  \033[0;33m~ saved for later\033[0m\n\n"
        break ;;
      r|R)
        # Re-trigger reload by touching the config
        swap_shader "$SHADERS_DIR/$name" "$anim"
        printf "  (reloaded)\n"
        ;;
      q|Q)
        echo "$i" > "$PROGRESS_FILE"
        printf "\n  Progress saved at %d/%d. Run with --resume to continue.\n" "$((i+1))" "$TOTAL"
        printf "  Picks so far saved to: %s\n\n" "$PICKS_FILE"
        exit 0 ;;
      *)
        printf "  (press y/n/s/r/q)\n" ;;
    esac
  done

  # Save progress
  echo "$((i+1))" > "$PROGRESS_FILE"
done

# ── Results ───────────────────────────────────────
printf "\n"
printf "  \033[1m══════════════════════════════════════════\033[0m\n"
printf "  \033[1mResults:\033[0m  \033[0;32m%d kept\033[0m  \033[0;31m%d skipped\033[0m  / %d total\n" "$KEPT" "$SKIPPED" "$TOTAL"
printf "  \033[1m══════════════════════════════════════════\033[0m\n"
printf "\n"

printf "  \033[1mCurated config lines:\033[0m\n"
printf "  ─────────────────────\n"
while IFS= read -r line; do
  [[ "$line" == \#* ]] && continue
  [[ -z "$line" ]] && continue
  printf "  custom-shader = %s/%s\n" "$SHADERS_DIR" "$line"
done < "$PICKS_FILE"
printf "\n"

printf "  Full picks list: %s\n" "$PICKS_FILE"
printf "  Copy the config lines above into your ghostty/config.\n\n"

# Clean up progress
rm -f "$PROGRESS_FILE"
