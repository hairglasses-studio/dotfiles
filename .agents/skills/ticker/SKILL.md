---
name: ticker
description: "Manage the keybind-ticker visual effects bar — restart, debug, tune effects, switch content streams, and iterate on the cyberpunk visual stack. Use when the user mentions the ticker, scrolling bar, keybind bar, bottom bar, visual effects tuning, stock ticker, or wants to adjust scroll speed, glow, scanlines, gradients, water caustic, glitch effects, or content streams on the ticker."
---

# Keybind Ticker (v3)

The keybind-ticker is a standalone GTK4 PangoCairo app rendering a pixel-smooth scrolling bar at the bottom of the Samsung ultrawide (DP-3). 12 content streams rotate with per-stream refresh intervals. Multi-layer cyberpunk visual effects stack.

## Architecture

- **App**: `scripts/keybind-ticker.py` — GTK4 DrawingArea + `add_tick_callback` for 240Hz frame-clock sync
- **Service**: `systemd/dotfiles-keybind-ticker.service` — Wayland-gated, sets `LD_PRELOAD` for layer-shell
- **Layer rule**: `hyprland/hyprland.conf` matches `namespace = ^keybind-ticker$` for blur
- **Playlist**: `ticker/content-playlists/main.txt` — one stream ID per line, read at startup
- **State**: `~/.local/state/keybind-ticker/` — `current-stream`, `paused`, `pinned-stream`

## Operations

### Restart
```bash
systemctl --user restart dotfiles-keybind-ticker.service
```

### Windowed debug mode
```bash
pkill -f keybind-ticker.py; sleep 0.3
GSK_RENDERER=gl python3 ~/hairglasses-studio/dotfiles/scripts/keybind-ticker.py
# with preset: ... --preset cyberpunk
# different monitor: ... --monitor DP-2
```

### Check status
```bash
systemctl --user status dotfiles-keybind-ticker.service
journalctl --user -u dotfiles-keybind-ticker.service --since "1 min ago"
```

### Screenshot
`mcp__dotfiles__hypr_screenshot` with `output: DP-3` and look at the bottom 28px. Or `scripts/capture-window-gif.sh ticker 3` for a 3-second GIF.

## Visual Effects Stack (v3)

| Layer | Effect | Key Config | Notes |
|-------|--------|------------|-------|
| BG-0 | Dark panel | `bg_alpha` | Base fill |
| BG-1 | Water caustic | `water_skip` | Ported from darkwindow/water.glsl |
| BG-2 | Scanlines | `scanline_opacity` | Under text |
| BG-3 | Holo shimmer | `holo_shimmer` | Phase 6 |
| BG-4 | Top border (synthwave sweep) | `synthwave_border` | Hot-pink→purple sweep |
| FG-5 | Text outline (solid or pulse) | `outline_width`, `outline_pulse` | Pulse = lit edge via gradient stroke |
| FG-6 | Wave distortion | `wave_amp`, `wave_freq`, `wave_speed` | Strip sine displacement |
| FG-7 | Typewriter clip | (on stream change) | Reveals text on rotation |
| FG-8 | Phosphor decay trail | `phosphor_trail` | CRT afterimage ring buffer |
| FG-9 | Ghost echo | `ghost_echo` | VHS double-image |
| FG-10 | Neon glow (breathing) | `glow_kernel`, `glow_base_alpha`, `glow_pulse_*` | Blur + sine breathing |
| FG-11 | Drop shadow | `shadow_offset`, `shadow_alpha` | Dark offset |
| FG-12 | Gradient text | `gradient_speed`, `gradient_span` | Flowing neon palette |
| FG-13 | Chromatic aberration | `ca_offset` | During glitch only |
| FG-14 | Glitch strips | `glitch_prob`, `glitch_frames` | Random displacement |
| POST | Edge fade vignette | `edge_fade` | Dark gradient mask on L/R edges |
| POST | Progress bar | `progress_bar` | 1px bottom bar = time in stream |

### Presets

- `ambient` (default) — subtle, breathing, readable
- `cyberpunk` — aggressive, more glitch, phosphor trail, hue sweep
- `minimal` — clean, no wave, no glitch, light effects
- `clean` — bare minimum, no effects

Per-stream preset override via `STREAM_META` dict:
- `fleet` → cyberpunk
- `weather` → ambient
- `music` → minimal

### Tuning

All tunables are constants at the top of `keybind-ticker.py` inside the `PRESETS` dict. Edit, then restart the service.

## Content Streams (12 total)

Each stream has its own refresh interval (see `STREAM_META`). Slow streams (github, music, updates) run on background threads to avoid blocking the render loop.

| Stream | Source | Refresh | Badge | Click action |
|--------|--------|---------|-------|--------------|
| `keybinds` | `hyprctl binds -j` | 5 min | cyan | Copy keybind |
| `system` | sensors / nvidia-smi / free / uptime | 10 s | yellow | — |
| `fleet` | `/tmp/rg-status.json` | 30 s | magenta | — |
| `weather` | `/tmp/bar-weather.txt` (cached) | 30 min | blue | — |
| `github` | `gh api /notifications` (threaded) | 2 min | green | Open URL in browser |
| `notifications` | `~/.local/state/.../history.jsonl` | 1 min | red | — |
| `music` | `playerctl` (threaded) | 10 s | magenta | — |
| `updates` | `/tmp/bar-updates.txt` + `checkupdates` | 30 min | cyan | — |
| `mx-battery` | `/tmp/bar-mx.txt` | 5 min | yellow | — |
| `disk` | `df -h` | 1 min | blue | — |
| `load` | `/proc/loadavg` | 5 s | green | Sparkline |
| `workspace` | `hyprctl activeworkspace/activewindow/workspaces -j` | 5 s | magenta | — |

To add a new stream, create `build_<name>_markup()` returning `(markup_str, segments_list)`, add to `STREAMS`, `STREAM_META`, and `main.txt`.

## Interactive Controls

| Input | Action |
|-------|--------|
| **Scroll wheel** | Adjust speed (10-200 px/s) |
| **Shift + scroll** | Switch streams |
| **Left-click** | Copy segment (keybinds) or open URL in browser (github) |
| **Middle-click** | Toggle pause (persists to `STATE_DIR/paused`) |
| **Right-click** | Context menu: streams / presets / pause / pin |
| **Hover** | Adaptive speed (slows to 20%); tooltip shows stream+preset+segment |

### Pin a stream
Right-click → "Pin current stream" → stops auto-rotation. Persists to `STATE_DIR/pinned-stream`. Unpin from the same menu.

### Priority interrupts
A DBus-driven watcher polls the notification history every 3 s. If a `critical`-urgency notification arrives, the ticker jumps to the notifications stream immediately and enters urgent mode (amplified glitch, CA, for 10 s).

## CLI Flags

- `--layer` — layer-shell mode (bottom-anchored, exclusive zone). Used by systemd.
- `--preset <name>` — start with `ambient|cyberpunk|minimal|clean`
- `--monitor <name>` — target a specific output (default `DP-3`)

## Hyprland Integration

In **layer-shell mode** (default via systemd), the ticker uses `gtk4-layer-shell` to anchor to the bottom of the monitor with exclusive zone. The systemd service sets `LD_PRELOAD=/usr/lib/libgtk4-layer-shell.so`.

In **windowed mode** (no `--layer` flag), the ticker runs as a floating window. Hyprland windowrule pins it to DP-3 at 2560x28 at the bottom.
