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

## The `misc:vfr = false` decision (load-bearing)

Hyprland's `misc:vfr = true` pauses rendering on idle — saves power but
freezes time-based shader animation (every `x_Time` reference stops advancing).

The current committed setting is `misc:vfr = false`: compositor renders every
frame regardless of damage. This is the trade-off that keeps all 139 DarkWindow
shaders animating correctly at rest.

Cost: ~20-40% idle GPU on RTX at 240Hz.

Escape hatch: `scripts/hypr-perf-mode.sh performance` flips `vfr=true` at
runtime for battery/thermal scenarios. Animated shaders freeze until a damage
event (mouse move, keystroke, window change) drives the next frame.

When editing shaders:
- Prefer shaders that animate via input-driven damage (cursor position, window
  focus) over pure time-based animation when writing for "mid" or "heavy" tiers
- Document in the shader's header whether it needs `vfr=false` to be visibly
  active

## Tier playlists (perf classification)

`kitty/shaders/bin/shader-tier.sh generate` writes three perf-tier playlists to
`kitty/shaders/playlists/`:

- `tier-cheap.txt` (~45 shaders, <2.2KB) — safe for ambient rotation at 240Hz
- `tier-mid.txt`   (~46 shaders, 2.2–4.6KB) — good for focused windows
- `tier-heavy.txt` (~48 shaders, ≥4.6KB) — showcase / perf-mode only

Classification uses file size + loop count as a static heuristic (no live GPU
measurement yet). Regenerate when shaders are added/removed:

```bash
kitty/shaders/bin/shader-tier.sh generate
kitty/shaders/bin/shader-tier.sh verify
```

Future extension: `shader_benchmark` MCP tool + nvtop delta for real GPU-cost
tiering. Tracked in ROADMAP Phase 3 deferred items.
