---
name: ticker
description: "Manage the keybind-ticker visual effects bar — restart, debug, tune effects, switch content streams, and iterate on the cyberpunk visual stack. Use when the user mentions the ticker, scrolling bar, keybind bar, bottom bar, visual effects tuning, stock ticker, or wants to adjust scroll speed, glow, scanlines, gradients, water caustic, glitch effects, or content streams on the ticker."
---

# Keybind Ticker (v3)

The keybind-ticker is a standalone GTK4 PangoCairo app rendering a pixel-smooth scrolling bar spanning the full width of DP-2 (5120x1440) at the bottom of the display. 29 content streams rotate with per-stream refresh intervals across 3 playlists (main / coding / focus). Multi-layer cyberpunk visual effects stack.

## Architecture

- **App**: `scripts/keybind-ticker.py` — GTK4 DrawingArea + `add_tick_callback` for 240Hz frame-clock sync
- **Service**: `systemd/dotfiles-keybind-ticker.service` — Wayland-gated, sets `LD_PRELOAD` for layer-shell
- **Layer rule**: `hyprland/hyprland.conf` matches `namespace = ^keybind-ticker$` for blur
- **Playlists**: `ticker/content-playlists/{main,coding,focus}.txt` — one stream ID per line, picked at startup via `--playlist` or `STATE_DIR/active-playlist`
- **State**: `~/.local/state/keybind-ticker/` — `current-stream`, `paused`, `pinned-stream`, `active-playlist`, `pomodoro.json`

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

## Content Streams (29 total)

Each stream has its own refresh interval (see `STREAM_META`). Slow streams (github, music, updates, ci, claude-sessions, hacker) run on background threads. Cache-fed streams read `/tmp/bar-<name>.txt` written by systemd timer-paired oneshot units (`systemd/bar-<name>.{service,timer}`).

### Core streams
| Stream | Source | Refresh | Badge |
|--------|--------|---------|-------|
| `keybinds` | `hyprctl binds -j` | 5 min | cyan |
| `system` | sensors / nvidia-smi / free / uptime | 10 s | yellow |
| `fleet` | `/tmp/rg-status.json` | 30 s | magenta |
| `weather` | `/tmp/bar-weather.txt` (cached) | 30 min | blue |
| `github` | `gh api /notifications` (threaded) | 2 min | green |
| `notifications` | `~/.local/state/.../history.jsonl` | 1 min | red |
| `music` | `playerctl` (threaded) | 10 s | magenta |
| `updates` | `/tmp/bar-updates.txt` + `checkupdates` | 30 min | cyan |
| `mx-battery` | `/tmp/bar-mx.txt` | 5 min | yellow |
| `disk` | `df -h` | 1 min | blue |
| `load` | `/proc/loadavg` | 5 s | green |
| `workspace` | `hyprctl activeworkspace/activewindow/workspaces -j` | 5 s | magenta |
| `network` | `nmcli` + `/proc/net/dev` | 30 s | blue |
| `audio` | `pactl` / playerctl | 10 s | orange |
| `shader` | `hyprshade current` | 30 s | purple |
| `claude-sessions` | `~/.claude/projects/*/*.jsonl` scan (threaded) | 60 s | amber |
| `ci` | `/tmp/bar-ci.txt` (gh workflow runs) | 2 min | green |
| `hacker` | `/tmp/bar-hacker.txt` (zenquotes cached) | 10 min | red |

### Phase 1 streams (April 2026)
| Stream | Source | Refresh | Badge |
|--------|--------|---------|-------|
| `calendar` | `/tmp/bar-calendar.txt` (gcalcli agenda 24h) | 10 min | blue |
| `pomodoro` | `~/.local/state/keybind-ticker/pomodoro.json` (via `hg pomo`) | 1 s | red |
| `token-burn` | `/tmp/bar-tokens.txt` (Claude JSONL sum) | 60 s | amber |
| `dirty-repos` | `/tmp/bar-dirty.txt` (workspace git status -s) | 5 min | orange |
| `failed-units` | `systemctl --failed` (live subprocess) | 60 s | red |
| `arch-news` | `/tmp/bar-archnews.txt` (RSS feed) | 1 h | blue |
| `smart-disk` | `/tmp/bar-smart.txt` (smartctl -H -j via sudo -n) | 1 h | purple |
| `wifi-quality` | `iw dev wlp11s0 link` (live subprocess) | 30 s | blue |
| `container-status` | `docker ps` (live subprocess) | 30 s | cyan |
| `net-throughput` | `/proc/net/dev` stateful delta | 5 s | green |
| `kernel-errors` | `journalctl -p err -k -n 5` (live subprocess) | 60 s | red |

### Playlists
- `main.txt` (default) — all 29 streams cycling
- `coding.txt` — keybinds, system, fleet, ci, claude-sessions, token-burn, dirty-repos, failed-units, kernel-errors, workspace, github, notifications, updates, shader
- `focus.txt` — system, pomodoro, calendar, workspace, music, notifications

To add a new stream, create `build_<name>_markup()` returning `(markup_str, segments_list)`, add to `STREAMS`, `STREAM_META`, `FALLBACK_ORDER`, and the relevant playlist(s). Cache-fed streams also need `scripts/bar-<name>-cache.sh` + `systemd/bar-<name>.{service,timer}` + entry in `install.sh::desktop_passive_units`.

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
- `--monitor <name>` — target a specific output (default `DP-2`, the 5120x1440 ultrawide)
- `--playlist <name>` — pick a playlist (main / coding / focus), or set `STATE_DIR/active-playlist` for persistence

## Hyprland Integration

In **layer-shell mode** (default via systemd), the ticker uses `gtk4-layer-shell` to anchor to the bottom of the monitor with exclusive zone. The systemd service sets `LD_PRELOAD=/usr/lib/libgtk4-layer-shell.so`.

In **windowed mode** (no `--layer` flag), the ticker runs as a floating window. Hyprland windowrule pins it to DP-3 at 2560x28 at the bottom.

## Multi-instance & Surfaces

### Secondary instance on DP-3
`systemd/dotfiles-keybind-ticker@.service` is a template unit. Instance name is `MONITOR_PLAYLIST` (underscore-delimited). `install.sh` enables `@DP-3_focus.service` by default so the portrait utility monitor gets its own focus-mode ticker. Each instance uses its own `--state-dir` (`~/.local/state/keybind-ticker-<instance>`), so pin / playlist / pause state doesn't collide between instances.

Add a third instance:
```bash
systemctl --user enable --now dotfiles-keybind-ticker@DP-4_coding.service
```

### Lock-screen swap
`scripts/ticker-lockwatch.sh` (ran as `dotfiles-ticker-lockwatch.service`) polls `pgrep -x hyprlock` every 2s. On lock, saves the current `active-playlist` to `pre-lock-playlist` and swaps to `lock.txt` (minimal: system + weather + notifications). On unlock, restores.

### Recording swap
`scripts/ticker-recordwatch.sh` (ran as `dotfiles-ticker-recordwatch.service`) watches `pgrep -x wf-recorder|wl-screenrec`. While active, writes `/tmp/bar-recording.txt` with `tool\tpid\tstart_epoch` and swaps the ticker to `recording.txt` (featuring the `recording` stream which renders duration + mic state + disk free). Notifies via `notify-send` on start/stop.

## Companion binaries (Phase 4)

All four share the 28px layer-shell form factor and the Hairglasses Neon palette. Each is a standalone Python + GTK4 script in `scripts/`.

| Script | Purpose | Key CLI |
|--------|---------|---------|
| `toast-ticker.py` | DBus-triggered slide-in toasts at `io.hairglasses.toast` — call `ShowToast(message, color)` to flash a 3s banner. | `--monitor`, `--duration` |
| `rsvp-ticker.py`  | Rapid Serial Visual Presentation reader. Consumes stdin, a file, or `--clipboard`. Adjustable 60–1200 WPM via scroll / arrows. | `--wpm`, `--clipboard` |
| `lyrics-ticker.py`| Now-playing banner. Polls `playerctl metadata` every 2s. Future: sync to `.lrc`. | `--monitor` |
| `subtitle-ticker.py`| Mute indicator. Shows "MUTED — audio detached" when default sink is muted. Placeholder for future whisper-live integration. | `--monitor` |

## Tmux status integration (Phase 3)

`scripts/ticker-headless.py` imports the stream builders, strips Pango markup, and prints plain text. Use it in `tmux.conf`:

```tmux
set -g status-right '#(python3 ~/hairglasses-studio/dotfiles/scripts/ticker-headless.py --stream ci --limit 60)'
```

Flags: `--stream <name>` single-shot, `--playlist <name>` cycle-per-minute, `--list` enumerate, `--limit N` truncate.

## `hg ticker` shell subcommand

`scripts/hg-modules/mod-ticker.sh` exposes shell control. Examples:
```bash
hg ticker status                # service / playlist / current / pinned / paused
hg ticker pin calendar          # pin a stream
hg ticker playlist focus        # switch playlist
hg ticker show ci               # plain-text one-shot
hg ticker list-streams          # enumerate
hg ticker list-playlists        # enumerate
hg ticker pause                 # toggle
hg ticker restart               # kick the service
```
