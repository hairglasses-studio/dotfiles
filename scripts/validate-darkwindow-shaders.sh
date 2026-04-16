#!/usr/bin/env bash
# validate-darkwindow-shaders.sh — Offline compile validator for Hypr-DarkWindow shaders
#
# Reproduces the plugin's TestCompilation + SpecialVariables::EditSource wrapping
# logic, then feeds each combined fragment shader to glslangValidator.
#
# Usage:
#   validate-darkwindow-shaders.sh [SHADER_DIR] [--baseline N] [--out DIR]
#
# Exit 0 if all shaders pass (or if pass count >= baseline). Exit 1 otherwise.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DOTFILES="$(cd "$SCRIPT_DIR/.." && pwd)"

SHADER_DIR="${DOTFILES}/kitty/shaders/darkwindow"
OUT_DIR="/tmp/shader-validation"
BASELINE=0

while [[ $# -gt 0 ]]; do
  case "$1" in
    --baseline) BASELINE="$2"; shift 2 ;;
    --out)      OUT_DIR="$2"; shift 2 ;;
    -*)         echo "Unknown flag: $1" >&2; exit 2 ;;
    *)          SHADER_DIR="$1"; shift ;;
  esac
done

command -v glslangValidator >/dev/null 2>&1 || {
  echo "error: glslangValidator not found" >&2
  exit 127
}

mkdir -p "$OUT_DIR"

# DarkWindow TestCompilation wrapper template — mirrors CustomShader.cpp
WRAP_TEMPLATE=$(cat <<'TMPL'
#version 300 es
precision highp float;

in vec2 v_texcoord;
uniform sampler2D tex;
uniform float alpha;

layout(location = 0) out vec4 fragColor;

uniform float x_Time;
uniform vec2  x_WindowSize;
uniform vec2  x_CursorPos;
uniform float x_MonitorScale;

vec4 x_Texture(vec2 texCoord) {
    return texture(tex, texCoord) * alpha;
}

vec4 x_TextureOffset(vec2 pixelOffset) {
    return x_Texture(v_texcoord - pixelOffset / x_WindowSize);
}

__CUSTOM_SOURCE__

void main() {
    vec4 pixColor = texture(tex, v_texcoord) * alpha;
    windowShader(pixColor);
    fragColor = pixColor;
}
TMPL
)

# Apply SpecialVariables::EditSource replacements and inject into the wrapper
wrap_source_file() {
  local src_file="$1"
  local tmp_src="$OUT_DIR/_tmp_src.glsl"
  perl -pe 's/\bx_PixelPos\b/(v_texcoord * x_WindowSize)/g; s/\bx_Tex\b/tex/g; s/\bx_TexCoord\b/v_texcoord/g' "$src_file" > "$tmp_src"
  awk -v src_file="$tmp_src" '
    /__CUSTOM_SOURCE__/ {
      while ((getline line < src_file) > 0) print line
      close(src_file)
      next
    }
    { print }
  ' <<< "$WRAP_TEMPLATE"
}

: > "$OUT_DIR/results.txt"
: > "$OUT_DIR/errors.log"
pass=0
fail=0

for f in "$SHADER_DIR"/*.glsl; do
  [[ -f "$f" ]] || continue
  name="$(basename "$f" .glsl)"
  wrapped="$OUT_DIR/${name}.frag"
  wrap_source_file "$f" > "$wrapped"
  if err=$(glslangValidator -S frag "$wrapped" 2>&1); then
    pass=$((pass + 1))
    echo "PASS $name" >> "$OUT_DIR/results.txt"
  else
    fail=$((fail + 1))
    echo "FAIL $name" >> "$OUT_DIR/results.txt"
    printf '=== %s ===\n%s\n\n' "$name" "$err" >> "$OUT_DIR/errors.log"
  fi
done

total=$((pass + fail))
echo "validated: $total  pass=$pass  fail=$fail"

if [[ "$BASELINE" -gt 0 && "$pass" -lt "$BASELINE" ]]; then
  echo "error: only $pass shaders pass, baseline requires $BASELINE" >&2
  exit 1
fi

if [[ "$fail" -gt 0 && "$BASELINE" -eq 0 ]]; then
  exit 1
fi

exit 0
