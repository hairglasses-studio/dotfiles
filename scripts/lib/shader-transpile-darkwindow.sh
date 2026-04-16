#!/usr/bin/env bash
# shader-transpile-darkwindow.sh — Convert Ghostty GLSL to Hypr-DarkWindow windowShader format
# Usage: shader-transpile-darkwindow.sh <input.glsl> [output.glsl]
# If output is omitted, prints to stdout.
#
# DarkWindow API:
#   void windowShader(inout vec4 color)
#   Available: x_Time, x_PixelPos, x_CursorPos, x_WindowSize
#   Sampling: x_Texture(uv), x_TextureOffset(offset)

set -euo pipefail

transpile_darkwindow() {
  local src="$1"
  local content
  content="$(< "$src")"

  # Strip #version and precision lines (DarkWindow injects its own)
  content="$(echo "$content" | sed -E \
    -e '/^#version/d' \
    -e '/^precision\s+/d' \
  )"

  # Replace Ghostty entry point with DarkWindow windowShader. Use a unique
  # out-parameter name (_wShaderOut) instead of `color` so it can't collide
  # with local `vec3 color` declarations common in the Ghostty source set.
  content="$(echo "$content" | sed -E \
    -e 's/void\s+mainImage\s*\(\s*out\s+vec4\s+fragColor\s*,\s*in\s+vec2\s+fragCoord\s*\)/void windowShader(inout vec4 _wShaderOut)/' \
  )"

  # Replace texture sampling: texture(iChannel0, uv) -> x_Texture(uv)
  content="$(echo "$content" | sed -E \
    -e 's/texture\(\s*iChannel0\s*,\s*/x_Texture(/g' \
    -e 's/\biChannel0\b/x_Texture/g' \
  )"

  # Replace time uniforms
  content="$(echo "$content" | sed -E \
    -e 's/\biTime\b/x_Time/g' \
    -e 's/\bghostty_time\b/x_Time/g' \
  )"

  # Replace resolution (vec3 in Ghostty -> vec2 x_WindowSize in DarkWindow)
  content="$(echo "$content" | sed -E \
    -e 's/\biResolution\.xy\b/x_WindowSize/g' \
    -e 's/\biResolution\.x\b/x_WindowSize.x/g' \
    -e 's/\biResolution\.y\b/x_WindowSize.y/g' \
    -e 's/\biResolution\b/vec3(x_WindowSize, 1.0)/g' \
  )"

  # Replace fragCoord with x_PixelPos
  content="$(echo "$content" | sed -E \
    -e 's/\bfragCoord\b/x_PixelPos/g' \
  )"

  # Replace output variable — must match the parameter name above. Keeping
  # it distinct from `color` avoids shadowing local `vec3 color` variables.
  content="$(echo "$content" | sed -E \
    -e 's/\bfragColor\b/_wShaderOut/g' \
  )"

  # Replace iFrame with time-based approximation
  content="$(echo "$content" | sed -E \
    -e 's/\biFrame\b/int(x_Time * 60.0)/g' \
  )"

  # Replace iMouse with zero
  content="$(echo "$content" | sed -E \
    -e 's/\biMouse\b/vec4(0.0)/g' \
  )"

  # Cursor uniforms — DarkWindow has x_CursorPos
  content="$(echo "$content" | sed -E \
    -e 's/\biCurrentCursor\b/x_CursorPos/g' \
    -e 's/\biPreviousCursor\b/x_CursorPos/g' \
    -e 's/\biCursorVisible\b/true/g' \
    -e 's/\biTimeCursorChange\b/x_Time/g' \
    -e 's/\biCurrentCursorColor\b/vec4(1.0)/g' \
    -e 's/\biPreviousCursorColor\b/vec4(1.0)/g' \
  )"

  # Focus uniforms
  content="$(echo "$content" | sed -E \
    -e 's/\biFocus\b/true/g' \
    -e 's/\biTimeFocus\b/0.0/g' \
  )"

  # Color uniforms (Snazzy palette defaults)
  content="$(echo "$content" | sed -E \
    -e 's/\biBackgroundColor\b/vec4(0.0, 0.0, 0.0, 1.0)/g' \
    -e 's/\biForegroundColor\b/vec4(0.945, 0.945, 0.941, 1.0)/g' \
    -e 's/\biCursorColor\b/vec4(0.341, 0.780, 1.0, 1.0)/g' \
    -e 's/\biSelectionForegroundColor\b/vec4(0.0, 0.0, 0.0, 1.0)/g' \
    -e 's/\biSelectionBackgroundColor\b/vec4(0.341, 0.780, 1.0, 1.0)/g' \
  )"

  # Remove uniform declarations (DarkWindow provides everything implicitly)
  content="$(echo "$content" | sed -E \
    -e '/^\s*uniform\s+/d' \
  )"

  echo "$content"
}

if [[ "${1:-}" == "--check" ]]; then
  src="$2"
  warnings=""
  grep -qE '\biPalette\b' "$src" && warnings="${warnings}iPalette "
  grep -qE '\biCurrentCursor\b|\biPreviousCursor\b' "$src" && warnings="${warnings}cursor(partial) "
  if [[ -n "$warnings" ]]; then
    echo "WARN:${warnings% }"
  else
    echo "OK"
  fi
  exit 0
fi

src="${1:?Usage: shader-transpile-darkwindow.sh <input.glsl> [output.glsl]}"
dst="${2:-}"

if [[ -n "$dst" ]]; then
  transpile_darkwindow "$src" > "$dst"
else
  transpile_darkwindow "$src"
fi
