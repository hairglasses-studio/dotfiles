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
| 4.5 | Text outline | `OUTLINE_WIDTH=0.8` | Dark stroke via `layout_path()` |
| 5 | Wave distortion | `WAVE_AMP=1.5`, `WAVE_FREQ=0.015`, `WAVE_SPEED=2.0` | Sine wave strip displacement |
| 6 | Gradient text | `GRADIENT_SPEED=40`, `GRADIENT_SPAN=800` | Flowing neon palette |
| 7 | Chromatic aberration | `CA_OFFSET=3` | During glitch only |
| 8 | Glitch strips | `GLITCH_PROB=0.004`, `GLITCH_FRAMES=4` | ~1/sec random displacement |

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
- Text outline: `OUTLINE_WIDTH` (0.0 = off, 0.8 = subtle, 1.0+ = bold)
- Wave distortion: `WAVE_AMP` (px, 0 = off, 1.5 = gentle, 2.0 = noticeable)
- Background transparency: `BG` alpha value (lower = more frosted glass visible)

### Disabling an effect

Set its opacity/alpha/probability to 0:
- No glow: `GLOW_BASE_ALPHA = 0`
- No scanlines: `SCANLINE_OPACITY = 0`
- No glitch: `GLITCH_PROB = 0`
- No water caustic: comment out the water caustic layer in `_draw()`
- No shadow: `SHADOW_ALPHA = 0`
- No outline: `OUTLINE_WIDTH = 0`
- No wave: `WAVE_AMP = 0`

## Font System

The ticker cycles through 10 Maple Mono NF CN weight variants via Pango markup (`<span font_desc="...">`). The `FONTS` list in the config section controls the rotation. Text colors come from a Cairo LinearGradient (not Pango foreground), which allows the animated flowing gradient effect.

## Content Streams

7 streams rotate every 5 minutes (REFRESH_S). Playlist state persists across restarts via `~/.local/state/keybind-ticker/current-stream`.

| Stream | Source | Badge Color | Notes |
|--------|--------|-------------|-------|
| `keybinds` | `hyprctl binds -j` | cyan | Click-to-copy via wl-copy |
| `system` | sensors, nvidia-smi, free, /proc/uptime | yellow | CPU/GPU/RAM/uptime |
| `fleet` | `/tmp/rg-status.json` | magenta | ralphglasses fleet status |
| `weather` | `scripts/bar-weather.sh` | blue | Cache-fed |
| `github` | `gh api /notifications` | green | PR/issue/release/discussion icons |
| `notifications` | `~/.local/state/.../history.jsonl` | red | Urgency icons, last 30 entries |
| `music` | `playerctl` | magenta | MPRIS now-playing with position |

To add a new stream, create `build_<name>_markup()` returning `(markup_str, segments_list)` and add it to `STREAMS` and `STREAM_ORDER`.

## Interactive Controls

- **Scroll wheel**: adjust scroll speed (10-200 px/s)
- **Click**: on keybinds stream, copies the keybind combo to clipboard via `wl-copy`
- **Hover tooltip**: shows current stream, speed, and hovered keybind (if applicable)

## Hyprland Integration

In **layer-shell mode** (default via systemd), the ticker uses `gtk4-layer-shell` to anchor to the bottom of DP-3 with exclusive zone. The systemd service sets `LD_PRELOAD=/usr/lib/libgtk4-layer-shell.so`.

In **windowed mode** (no `--layer` flag), the windowrule pins it:
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
