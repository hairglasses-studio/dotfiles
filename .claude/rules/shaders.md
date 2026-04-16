---
paths:
  - "kitty/shaders/**"
  - "wallpaper-shaders/**"
---

# Shader Rules

- All `.glsl` files must be self-contained — DarkWindow has NO `#include` support
- Use `// #include "lib/X.glsl"` comments for the preprocessor (`shader-build.sh` inlines them)
- Test compilation with `glslangValidator -S frag <file>`
- DarkWindow shaders use `x_Time` for animation; Hyprland uses `time` uniform — shaders need `NeedsConstantDamage` for per-frame updates
- DarkWindow uniforms: `_wTexture` (sampler2D), `_wResolution` (vec2), `x_Time` (float)
- Wallpaper shaders (in `wallpaper-shaders/`) use Shadertoy API: `mainImage(out vec4, in vec2)` with `iResolution`, `iTime`
- Terminal shaders should keep backgrounds dark for text readability
- The shader catalog lives at `kitty/shaders/darkwindow/` (139 `.glsl` files)
