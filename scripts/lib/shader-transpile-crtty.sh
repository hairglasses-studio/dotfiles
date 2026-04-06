#!/usr/bin/env bash
# shader-transpile-crtty.sh — Convert Ghostty GLSL to CRTty GLSL 330 core format
# Usage: shader-transpile-crtty.sh <input.glsl> [output.glsl]
# If output is omitted, prints to stdout.

set -euo pipefail

CRTTY_HEADER='#version 330 core
in vec2 v_uv;
out vec4 o_color;
uniform sampler2D u_input;
uniform float u_time;
uniform vec2 u_resolution;
'

transpile_crtty() {
  local src="$1"
  local content
  content="$(< "$src")"

  # Strip existing #version lines (CRTty requires 330 core)
  content="$(echo "$content" | sed '/^#version/d')"

  # Replace Ghostty entry point with CRTty main()
  # Handle multi-line signature variations
  content="$(echo "$content" | sed -E \
    -e 's/void\s+mainImage\s*\(\s*out\s+vec4\s+fragColor\s*,\s*in\s+vec2\s+fragCoord\s*\)/void main()/' \
  )"

  # Replace uniforms: iChannel0 -> u_input
  content="$(echo "$content" | sed -E \
    -e 's/texture\(\s*iChannel0/texture(u_input/g' \
    -e 's/iChannel0/u_input/g' \
  )"

  # Replace time uniforms
  content="$(echo "$content" | sed -E \
    -e 's/\biTime\b/u_time/g' \
    -e 's/\bghostty_time\b/u_time/g' \
  )"

  # Replace iResolution (vec3 in Ghostty -> vec2 in CRTty)
  # iResolution.xy -> u_resolution, iResolution.x -> u_resolution.x
  content="$(echo "$content" | sed -E \
    -e 's/\biResolution\.xy\b/u_resolution/g' \
    -e 's/\biResolution\.x\b/u_resolution.x/g' \
    -e 's/\biResolution\.y\b/u_resolution.y/g' \
    -e 's/\biResolution\b/vec3(u_resolution, 1.0)/g' \
  )"

  # Replace fragCoord with gl_FragCoord.xy
  content="$(echo "$content" | sed -E \
    -e 's/\bfragCoord\b/gl_FragCoord.xy/g' \
  )"

  # Replace output variable
  content="$(echo "$content" | sed -E \
    -e 's/\bfragColor\b/o_color/g' \
  )"

  # Replace iFrame with time-based approximation
  content="$(echo "$content" | sed -E \
    -e 's/\biFrame\b/int(u_time * 60.0)/g' \
  )"

  # Replace iMouse with zero (no mouse in CRTty)
  content="$(echo "$content" | sed -E \
    -e 's/\biMouse\b/vec4(0.0)/g' \
  )"

  # Stub out cursor uniforms (CRTty has no cursor data)
  content="$(echo "$content" | sed -E \
    -e 's/\biCurrentCursor\b/vec2(0.0)/g' \
    -e 's/\biPreviousCursor\b/vec2(0.0)/g' \
    -e 's/\biCursorVisible\b/false/g' \
    -e 's/\biTimeCursorChange\b/0.0/g' \
    -e 's/\biCurrentCursorColor\b/vec4(1.0)/g' \
    -e 's/\biPreviousCursorColor\b/vec4(1.0)/g' \
  )"

  # Stub out focus uniforms
  content="$(echo "$content" | sed -E \
    -e 's/\biFocus\b/true/g' \
    -e 's/\biTimeFocus\b/0.0/g' \
  )"

  # Stub out color uniforms
  content="$(echo "$content" | sed -E \
    -e 's/\biBackgroundColor\b/vec4(0.0, 0.0, 0.0, 1.0)/g' \
    -e 's/\biForegroundColor\b/vec4(0.945, 0.945, 0.941, 1.0)/g' \
    -e 's/\biCursorColor\b/vec4(0.341, 0.780, 1.0, 1.0)/g' \
    -e 's/\biSelectionForegroundColor\b/vec4(0.0, 0.0, 0.0, 1.0)/g' \
    -e 's/\biSelectionBackgroundColor\b/vec4(0.341, 0.780, 1.0, 1.0)/g' \
  )"

  # Remove uniform declarations that are now provided by CRTty header
  content="$(echo "$content" | sed -E \
    -e '/^\s*uniform\s+sampler2D\s+iChannel0/d' \
    -e '/^\s*uniform\s+float\s+(iTime|ghostty_time|u_time)/d' \
    -e '/^\s*uniform\s+vec[23]\s+iResolution/d' \
    -e '/^\s*uniform\s+int\s+iFrame/d' \
    -e '/^\s*uniform\s+vec[24]\s+iMouse/d' \
    -e '/^\s*uniform\s+vec2\s+(iCurrentCursor|iPreviousCursor)/d' \
    -e '/^\s*uniform\s+bool\s+(iCursorVisible|iFocus)/d' \
    -e '/^\s*uniform\s+float\s+(iTimeCursorChange|iTimeFocus)/d' \
    -e '/^\s*uniform\s+vec4\s+(iCurrentCursorColor|iPreviousCursorColor)/d' \
    -e '/^\s*uniform\s+vec4\s+(iBackgroundColor|iForegroundColor|iCursorColor)/d' \
    -e '/^\s*uniform\s+vec4\s+(iSelectionForegroundColor|iSelectionBackgroundColor)/d' \
  )"

  # Prepend CRTty header
  printf '%s\n%s\n' "$CRTTY_HEADER" "$content"
}

if [[ "${1:-}" == "--check" ]]; then
  # Dry-run mode: report which shaders would need manual review
  src="$2"
  warnings=""
  grep -qE '\biCurrentCursor\b|\biPreviousCursor\b' "$src" && warnings="${warnings}cursor "
  grep -qE '\biFrame\b' "$src" && warnings="${warnings}iFrame "
  grep -qE '\biMouse\b' "$src" && warnings="${warnings}iMouse "
  grep -qE '\biPalette\b' "$src" && warnings="${warnings}iPalette "
  if [[ -n "$warnings" ]]; then
    echo "WARN:${warnings% }"
  else
    echo "OK"
  fi
  exit 0
fi

src="${1:?Usage: shader-transpile-crtty.sh <input.glsl> [output.glsl]}"
dst="${2:-}"

if [[ -n "$dst" ]]; then
  transpile_crtty "$src" > "$dst"
else
  transpile_crtty "$src"
fi
