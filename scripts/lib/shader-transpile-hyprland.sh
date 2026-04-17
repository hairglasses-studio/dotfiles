#!/usr/bin/env bash
# shader-transpile-hyprland.sh — Convert non-Hyprland GLSL → hypr-shader-preview format
#
# Produces a WebGL 2 / GLSL ES 300 fragment shader for
# https://github.com/h-banii/hypr-shader-preview.
#
# Handles two input dialects via a macro shim on top of Hyprland's
# (tex, time, v_texcoord) uniform set:
#
#   darkwindow   — kitty DarkWindow: `void windowShader(inout vec4)`,
#                  `x_Time`, `x_Texture()`, `x_PixelPos`, `x_WindowSize`.
#   shadertoy    — Shadertoy / wallpaper-shaders: `void mainImage(out vec4, in vec2)`,
#                  `iTime`, `iResolution`, `iMouse`, `iChannel0..3`.
#
# Usage:
#   shader-transpile-hyprland.sh <input> [output]
#   shader-transpile-hyprland.sh --type darkwindow <input> [output]
#   shader-transpile-hyprland.sh --type shadertoy  <input> [output]
#   shader-transpile-hyprland.sh --check <input>
#
# If output is omitted, prints to stdout. Type defaults to `auto`.

set -euo pipefail

detect_type() {
  local src="$1"
  if   grep -qE '^\s*void\s+windowShader\s*\(' "$src" 2>/dev/null; then echo "darkwindow"
  elif grep -qE '^\s*void\s+mainImage\s*\('    "$src" 2>/dev/null; then echo "shadertoy"
  else echo "unknown"
  fi
}

strip_directives() {
  sed -E \
    -e '/^\s*#version\b/d' \
    -e '/^\s*precision\s+/d' \
    "$1"
}

emit_darkwindow() {
  local src="$1"
  cat <<'HEADER'
#version 300 es
// Transpiled from Kitty DarkWindow → Hyprland (hypr-shader-preview) format.
// Shim: x_Time / x_Texture() / x_PixelPos / x_WindowSize are aliased to the
// Hyprland (tex, time, v_texcoord) uniform set.

precision mediump float;

in  vec2 v_texcoord;
out vec4 fragColor;
uniform sampler2D tex;
uniform float     time;

#define x_Time               time
#define x_PixelPos           (v_texcoord * vec2(textureSize(tex, 0)))
#define x_CursorPos          vec2(0.0)
#define x_WindowSize         vec2(textureSize(tex, 0))
#define x_Texture(uv)        texture(tex, (uv))
#define x_TextureOffset(off) texture(tex, v_texcoord + (off) / vec2(textureSize(tex, 0)))

HEADER
  strip_directives "$src"
  cat <<'FOOTER'

// Hyprland entrypoint — bridges WebGL's main() to DarkWindow's windowShader().
void main() {
    vec4 _wBase = texture(tex, v_texcoord);
    windowShader(_wBase);
    fragColor = _wBase;
}
FOOTER
}

emit_shadertoy() {
  local src="$1"
  cat <<'HEADER'
#version 300 es
// Transpiled from Shadertoy → Hyprland (hypr-shader-preview) format.
// Shim: iTime / iResolution / iMouse / iChannel0..3 aliased to Hyprland uniforms.

precision mediump float;

in  vec2 v_texcoord;
out vec4 fragColor;
uniform sampler2D tex;
uniform float     time;

#define iResolution vec3(vec2(textureSize(tex, 0)), 1.0)
#define iTime       time
#define iTimeDelta  (1.0/60.0)
#define iFrame      int(time * 60.0)
#define iMouse      vec4(0.0)
#define iChannel0   tex
#define iChannel1   tex
#define iChannel2   tex
#define iChannel3   tex

HEADER
  strip_directives "$src"
  cat <<'FOOTER'

// Hyprland entrypoint — bridges WebGL's main() to Shadertoy's mainImage().
void main() {
    vec2 fragCoord = v_texcoord * vec2(textureSize(tex, 0));
    mainImage(fragColor, fragCoord);
}
FOOTER
}

transpile() {
  local src="$1" type="$2"
  [[ "$type" == "auto" ]] && type="$(detect_type "$src")"
  case "$type" in
    darkwindow) emit_darkwindow "$src" ;;
    shadertoy)  emit_shadertoy  "$src" ;;
    unknown|*)
      echo "shader-transpile-hyprland: cannot detect shader type for $src — pass --type darkwindow|shadertoy" >&2
      return 2
      ;;
  esac
}

check_compatibility() {
  local src="$1"
  local warnings=""
  local type
  type="$(detect_type "$src")"

  case "$type" in
    darkwindow)
      grep -qE '\bx_PixelPos\s*\*'           "$src" && warnings="${warnings}pixelpos-arith(scale-sensitive) "
      grep -qE '\biPalette\b|\bx_Palette\b'  "$src" && warnings="${warnings}palette-uniform(unsupported) "
      grep -qE '\bgl_FragColor\b'            "$src" && warnings="${warnings}legacy-gl_FragColor "
      ;;
    shadertoy)
      grep -qE '\biChannel[1-3]\b'           "$src" && warnings="${warnings}multi-channel(iChannel1..3-collapsed-to-tex) "
      grep -qE '\bsampler(Cube|3D)\b'        "$src" && warnings="${warnings}cube-or-3d-sampler(unsupported) "
      grep -qE '\bmain\s*\('                 "$src" && warnings="${warnings}conflicting-main "
      ;;
    unknown)
      warnings="${warnings}no-known-entrypoint(windowShader|mainImage) "
      ;;
  esac

  if [[ -n "$warnings" ]]; then
    echo "type=$type WARN:${warnings% }"
  else
    echo "type=$type OK"
  fi
}

force_type="auto"
action="transpile"
src=""
dst=""

while [[ $# -gt 0 ]]; do
  case "$1" in
    --type)  force_type="$2"; shift 2 ;;
    --check) action="check"; shift ;;
    -h|--help) sed -n '2,21p' "$0"; exit 0 ;;
    -*) echo "unknown flag: $1" >&2; exit 2 ;;
    *)  if [[ -z "$src" ]]; then src="$1"
        elif [[ -z "$dst" ]]; then dst="$1"
        else echo "extra arg: $1" >&2; exit 2
        fi
        shift ;;
  esac
done

[[ -n "$src" ]] || { echo "Usage: shader-transpile-hyprland.sh [--type darkwindow|shadertoy] <input> [output]" >&2; exit 2; }

if [[ "$action" == "check" ]]; then
  check_compatibility "$src"
  exit 0
fi

if [[ -n "$dst" ]]; then
  transpile "$src" "$force_type" > "$dst"
else
  transpile "$src" "$force_type"
fi
