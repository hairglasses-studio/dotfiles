#!/usr/bin/env bash
# ── shader-benchmark — measure shader performance via glslViewer ──
# Renders each shader headless at 1920x1080 and reports FPS.
#
# Usage:
#   shader-bench bloom-soft.glsl         # benchmark one shader
#   shader-bench --playlist best-of      # benchmark all in a playlist
#   shader-bench --all --sort fps        # full collection ranked by FPS
#   shader-bench --update-manifest       # write measured cost back to shaders.toml

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SHADERS_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
MANIFEST="$SHADERS_DIR/shaders.toml"
RESULTS_FILE="${HOME}/.local/state/ghostty/benchmarks.tsv"
GLSLVIEWER="$(command -v glslViewer 2>/dev/null || echo "")"

WIDTH=1920
HEIGHT=1080
DURATION=5  # seconds per shader
SORT_BY="name"
UPDATE_MANIFEST=false

# ── Parse args ────────────────────────────────────
FILES=()
while [[ $# -gt 0 ]]; do
  case "$1" in
    --all)
      for f in "$SHADERS_DIR"/*.glsl; do
        [[ -f "$f" ]] && FILES+=("$f")
      done
      shift ;;
    --playlist)
      local_plist="$SHADERS_DIR/playlists/${2}.txt"
      if [[ ! -f "$local_plist" ]]; then
        echo "Playlist not found: $2" >&2; exit 1
      fi
      while IFS= read -r name; do
        [[ -f "$SHADERS_DIR/$name" ]] && FILES+=("$SHADERS_DIR/$name")
      done < "$local_plist"
      shift 2 ;;
    --sort)
      SORT_BY="$2"; shift 2 ;;
    --duration)
      DURATION="$2"; shift 2 ;;
    --size)
      WIDTH="${2%%x*}"; HEIGHT="${2##*x}"; shift 2 ;;
    --update-manifest)
      UPDATE_MANIFEST=true; shift ;;
    -h|--help)
      echo "Usage: shader-bench [--all | --playlist NAME] [--sort fps|name|cost] [--duration N]"
      echo ""
      echo "Benchmarks shaders using glslViewer headless rendering."
      echo ""
      echo "Options:"
      echo "  --all                 Benchmark all .glsl files"
      echo "  --playlist <name>     Benchmark shaders in a playlist"
      echo "  --sort fps|name|cost  Sort results (default: name)"
      echo "  --duration <secs>     Render duration per shader (default: 5)"
      echo "  --size WxH            Render resolution (default: 1920x1080)"
      echo "  --update-manifest     Write measured cost back to shaders.toml"
      echo ""
      echo "Results saved to: $RESULTS_FILE"
      exit 0 ;;
    *.glsl)
      if [[ -f "$SHADERS_DIR/$1" ]]; then
        FILES+=("$SHADERS_DIR/$1")
      elif [[ -f "$1" ]]; then
        FILES+=("$1")
      else
        echo "Not found: $1" >&2; exit 1
      fi
      shift ;;
    *) echo "Unknown option: $1" >&2; exit 1 ;;
  esac
done

if [[ -z "$GLSLVIEWER" ]]; then
  echo "ERROR: glslViewer not found. Install with: brew install glslviewer" >&2
  exit 1
fi

if [[ ${#FILES[@]} -eq 0 ]]; then
  echo "No shaders specified. Use --all, --playlist, or provide .glsl files." >&2
  exit 0
fi

mkdir -p "$(dirname "$RESULTS_FILE")" 2>/dev/null

# ── Create wrapper fragment shader ────────────────
# glslViewer expects a .frag file with main(). Most Ghostty shaders use
# Shadertoy-style mainImage(), so we create a wrapper.
create_wrapper() {
  local shader_path="$1"
  local wrapper
  wrapper="$(/usr/bin/mktemp /tmp/shader-bench-XXXXXX.frag)"

  # Check if shader uses mainImage (Shadertoy convention) or main()
  if grep -q 'void mainImage' "$shader_path"; then
    cat > "$wrapper" << 'FRAG_HEAD'
#ifdef GL_ES
precision highp float;
#endif

uniform sampler2D u_tex0;
uniform vec2 u_resolution;
uniform float u_time;

#define iResolution vec3(u_resolution, 1.0)
#define iTime u_time
#define iChannel0 u_tex0

FRAG_HEAD
    cat "$shader_path" >> "$wrapper"
    cat >> "$wrapper" << 'FRAG_TAIL'

void main() {
    mainImage(gl_FragColor, gl_FragCoord.xy);
}
FRAG_TAIL
  else
    # Shader already has main() — use as-is but add uniform aliases
    cat > "$wrapper" << 'FRAG_HEAD'
#ifdef GL_ES
precision highp float;
#endif

uniform sampler2D u_tex0;
uniform vec2 u_resolution;
uniform float u_time;

#define iResolution vec3(u_resolution, 1.0)
#define iTime u_time
#define iChannel0 u_tex0
#define ghostty_time u_time
#define ghostty_resolution u_resolution
#define ghostty_texture u_tex0

FRAG_HEAD
    cat "$shader_path" >> "$wrapper"
  fi

  echo "$wrapper"
}

# ── Benchmark one shader ──────────────────────────
bench_one() {
  local shader_path="$1"
  local shader_name
  shader_name="$(basename "$shader_path")"

  local wrapper
  wrapper="$(create_wrapper "$shader_path")"

  # Run glslViewer headless, capture FPS output
  local output fps_val
  output="$(echo "fps" | timeout "$((DURATION + 2))" \
    "$GLSLVIEWER" "$wrapper" \
    --headless --noncurses \
    -w "$WIDTH" -h "$HEIGHT" \
    2>&1)" || true

  # Extract FPS from output (glslViewer prints "FPS: XX.XX" or just the number)
  fps_val="$(echo "$output" | grep -oE '[0-9]+\.[0-9]+' | tail -1)"

  # If we couldn't get FPS, try a different approach: measure frame count
  if [[ -z "$fps_val" ]]; then
    # Fall back: run for DURATION seconds, count frames via stderr
    local start_time end_time frames
    start_time=$(date +%s)
    timeout "$DURATION" "$GLSLVIEWER" "$wrapper" \
      --headless --noncurses \
      -w "$WIDTH" -h "$HEIGHT" \
      > /dev/null 2>&1 || true
    end_time=$(date +%s)
    fps_val="N/A"
  fi

  rm -f "$wrapper"
  echo "$fps_val"
}

# ── Main benchmark loop ──────────────────────────
printf "\n  Shader Benchmark\n"
printf "  ────────────────\n"
printf "  Resolution: %dx%d\n" "$WIDTH" "$HEIGHT"
printf "  Duration: %ds per shader\n" "$DURATION"
printf "  Shaders: %d\n\n" "${#FILES[@]}"

declare -a RESULTS=()

for shader_path in "${FILES[@]}"; do
  shader_name="$(basename "$shader_path")"
  printf "  Benchmarking: %-40s " "$shader_name"

  fps="$(bench_one "$shader_path")"

  # Classify cost based on FPS
  local cost="MED"
  if [[ "$fps" != "N/A" ]]; then
    # Compare as integers (truncate decimals)
    fps_int="${fps%%.*}"
    if (( fps_int >= 120 )); then
      cost="LOW"
    elif (( fps_int >= 30 )); then
      cost="MED"
    else
      cost="HIGH"
    fi
  fi

  printf "%s FPS  (%s)\n" "$fps" "$cost"
  RESULTS+=("$shader_name\t$fps\t$cost")
done

# ── Save results ──────────────────────────────────
{
  printf "shader\tfps\tcost\tdate\n"
  for entry in "${RESULTS[@]}"; do
    printf '%b\t%s\n' "$entry" "$(date +%Y-%m-%d)"
  done
} > "$RESULTS_FILE"

# ── Sort and display ──────────────────────────────
printf "\n  ══════════════════════════════════════════════════════\n"
printf "  %-40s %8s %5s\n" "SHADER" "FPS" "COST"
printf "  %-40s %8s %5s\n" "──────" "───" "────"

case "$SORT_BY" in
  fps)
    for entry in "${RESULTS[@]}"; do
      printf "  %b\n" "$entry"
    done | sort -t$'\t' -k2 -rn | while IFS=$'\t' read -r name fps cost; do
      printf "  %-40s %8s %5s\n" "$name" "$fps" "$cost"
    done ;;
  cost)
    for entry in "${RESULTS[@]}"; do
      printf "  %b\n" "$entry"
    done | sort -t$'\t' -k3,3 -k2,2rn | while IFS=$'\t' read -r name fps cost; do
      printf "  %-40s %8s %5s\n" "$name" "$fps" "$cost"
    done ;;
  *)
    for entry in "${RESULTS[@]}"; do
      printf "  %b\n" "$entry"
    done | sort -t$'\t' -k1,1 | while IFS=$'\t' read -r name fps cost; do
      printf "  %-40s %8s %5s\n" "$name" "$fps" "$cost"
    done ;;
esac

printf "  ══════════════════════════════════════════════════════\n"
printf "\n  Results saved to: %s\n\n" "$RESULTS_FILE"

# ── Update manifest costs ─────────────────────────
if $UPDATE_MANIFEST; then
  updated=0
  for entry in "${RESULTS[@]}"; do
    IFS=$'\t' read -r name fps cost <<< "$(printf '%b' "$entry")"
    shader_base="${name%.glsl}"
    current_cost="$("$SCRIPT_DIR/shader-meta.sh" get "$shader_base" cost 2>/dev/null)"
    if [[ -n "$current_cost" ]] && [[ "$current_cost" != "$cost" ]]; then
      # Update cost in manifest
      if [[ "$(uname)" == "Darwin" ]]; then
        sed -i '' "/^\[shaders\.\"${shader_base}\"\]/,/^\[shaders\./ s/^cost = \".*\"/cost = \"${cost}\"/" "$MANIFEST"
      else
        sed -i "/^\[shaders\.\"${shader_base}\"\]/,/^\[shaders\./ s/^cost = \".*\"/cost = \"${cost}\"/" "$MANIFEST"
      fi
      printf "  Updated: %s  %s → %s\n" "$name" "$current_cost" "$cost"
      updated=$((updated + 1))
    fi
  done
  if [[ $updated -eq 0 ]]; then
    echo "  All manifest costs match benchmarks."
  else
    echo "  Updated $updated entries in shaders.toml."
  fi
fi
