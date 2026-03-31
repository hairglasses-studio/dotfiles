---
paths:
  - "ghostty/shaders/**"
  - "wallpaper-shaders/**"
---

# Shader Rules

- All `.glsl` files must be self-contained — Ghostty has NO `#include` support
- Use `// #include "lib/X.glsl"` comments for the preprocessor (`shader-build.sh` inlines them)
- Test compilation with `glslangValidator -S frag <file>`
- Shaders using `ghostty_time`, `iTime`, or `u_time` uniforms need `custom-shader-animation = true`
- Wallpaper shaders (in `wallpaper-shaders/`) use Shadertoy API: `mainImage(out vec4, in vec2)` with `iResolution`, `iTime`
- Terminal shaders should keep backgrounds dark for text readability
- The manifest at `ghostty/shaders/shaders.toml` is the single source of truth for metadata
