---
name: ticker
description: "Manage the keybind-ticker visual effects bar — restart, debug, tune effects, switch content streams, and iterate on the cyberpunk visual stack. Use when the user mentions the ticker, scrolling bar, keybind bar, bottom bar, visual effects tuning, stock ticker, or wants to adjust scroll speed, glow, scanlines, gradients, water caustic, glitch effects, or content streams on the ticker."
---

# Keybind Ticker

The keybind-ticker is a standalone GTK4 PangoCairo app that renders a pixel-smooth scrolling bar at the bottom of the Samsung ultrawide (DP-3). It displays keybindings with a cyberpunk visual effects stack (water caustic, neon glow, gradient, scanlines, glitch).

## Architecture

- **App**: `scripts/keybind-ticker.py` — Python GTK4 DrawingArea with `add_tick_callback` for 240Hz frame-clock sync
- **Service**: `systemd/dotfiles-keybind-ticker.service` — Wayland-gated, hyprctl-ready, auto-restarts
- **Window rule**: `hyprland/hyprland.conf` matches `title = ^(keybind-ticker)$` — pinned floating on DP-3, `decorate = false`
- **Layer rule**: `hyprland/hyprland.conf` matches `namespace = ^keybind-ticker$` for frosted glass blur in `--layer` mode
- **Ironbar**: The old bash-based ticker was removed from `ironbar/config.toml`. The standalone Python app replaced it.

## Operations

### Restart the ticker
```bash
systemctl --user restart dotfiles-keybind-ticker.service
```

### Launch in windowed debug mode (tiled, visible errors)
```bash
pkill -f keybind-ticker.py; sleep 0.3
GSK_RENDERER=gl python3 ~/hairglasses-studio/dotfiles/scripts/keybind-ticker.py
```

### Check if running
```bash
systemctl --user status dotfiles-keybind-ticker.service
# or
pgrep -fa keybind-ticker
```

### Screenshot the ticker bar region
Use `mcp__dotfiles__hypr_screenshot` with `output: DP-3` and look at the bottom 28px.

## Visual Effects Stack (compositing order)

All effects render in the BACKGROUND, text stays crisp on top:

| Layer | Effect | Config Constant | Notes |
|-------|--------|-----------------|-------|
| 0 | Dark panel BG | `BG = (0.02, 0.03, 0.05, 0.82)` | Alpha < 1 for frosted glass |
| 0.5 | Water caustic | `WATER_SKIP = 4` | Ported from `kitty/shaders/darkwindow/water.glsl` |
| 1 | CRT scanlines | `SCANLINE_OPACITY = 0.08` | Under text for readability |
| 2 | Top border | Animated gradient | 1px neon gradient line |
| 3 | Neon glow | `GLOW_KERNEL=17`, `GLOW_BASE_ALPHA=0.35`, `GLOW_PULSE_*` | Breathing halo via sine wave |
| 4 | Drop shadow | `SHADOW_OFFSET=2`, `SHADOW_ALPHA=0.30` | Dark offset for depth |
| 5 | Gradient text | `GRADIENT_SPEED=40`, `GRADIENT_SPAN=800` | Flowing neon palette |
| 6 | Chromatic aberration | `CA_OFFSET=3` | During glitch only |
| 7 | Glitch strips | `GLITCH_PROB=0.004`, `GLITCH_FRAMES=4` | ~1/sec random displacement |

### Tuning effects

All tunables are constants at the top of `keybind-ticker.py`. To adjust:

1. Edit the constant in `scripts/keybind-ticker.py`
2. Restart: `pkill -f keybind-ticker.py && systemctl --user restart dotfiles-keybind-ticker.service`
3. Screenshot to verify: use `mcp__dotfiles__hypr_screenshot output=DP-3`

**Common adjustments:**
- Scroll speed: `SPEED` (px/sec, default 55)
- Glow intensity: `GLOW_BASE_ALPHA` (0.0–1.0)
- Glow breathing: `GLOW_PULSE_AMP` and `GLOW_PULSE_PERIOD`
- Scanline visibility: `SCANLINE_OPACITY` (0.0 = off, 0.08 = subtle)
- Glitch frequency: `GLITCH_PROB` (0.0 = off, 0.004 = ~1/sec at 240Hz)
- Water caustic: set `WATER_SKIP` higher for less CPU, or remove the layer entirely
- Background transparency: `BG` alpha value (lower = more frosted glass visible)

### Disabling an effect

Set its opacity/alpha/probability to 0:
- No glow: `GLOW_BASE_ALPHA = 0`
- No scanlines: `SCANLINE_OPACITY = 0`
- No glitch: `GLITCH_PROB = 0`
- No water caustic: comment out the water caustic layer in `_draw()`
- No shadow: `SHADOW_ALPHA = 0`

## Font System

The ticker cycles through 10 Maple Mono NF CN weight variants via Pango markup (`<span font_desc="...">`). The `FONTS` list in the config section controls the rotation. Text colors come from a Cairo LinearGradient (not Pango foreground), which allows the animated flowing gradient effect.

## Content Streams (planned)

The ticker currently shows only keybinds. Multi-stream support is planned:
- `keybinds` — current, from `hyprctl binds -j`
- `system` — CPU/GPU/RAM/temp metrics
- `fleet` — ralphglasses fleet status from `/tmp/rg-status.json`
- `weather` — from `scripts/bar-weather.sh`
- `github` — notifications via `gh api /notifications`
- `music` — MPRIS now-playing via `playerctl`
- `notifications` — desktop notification history

When implementing new streams, add a `build_<stream>_markup()` function that returns Pango markup (same format as `build_ticker_markup()`). The `_rebuild()` method selects which stream to display.

## Hyprland Integration

The window rule pins the ticker to DP-3 as a floating window:
```
windowrule {
    name = keybind-ticker
    match:title = ^(keybind-ticker)$
    monitor = DP-3
    float = yes
    size = 2560 28
    move = 0 692
    pin = yes
    decorate = false
}
```

If the ticker gets accidentally moved (e.g., Mod+[+/-]), restart the service to reset position.
