---
name: ticker
description: "Manage the keybind-ticker visual effects bar — restart, debug, tune effects, switch content streams, and iterate on the cyberpunk visual stack. Use when the user mentions the ticker, scrolling bar, keybind bar, bottom bar, visual effects tuning, stock ticker, or wants to adjust scroll speed, glow, scanlines, gradients, water caustic, glitch effects, or content streams on the ticker."
---

# Keybind Ticker (v3)

The keybind-ticker is a standalone GTK4 PangoCairo app rendering a pixel-smooth scrolling bar spanning the full width of DP-2 (5120x1440) at the bottom of the display. 39 content streams rotate with per-stream refresh intervals across 3 playlists (main / coding / focus). Multi-layer cyberpunk visual effects stack.

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

**Always use `scripts/ticker-shot.sh` (or `hg ticker shot`).** Full-monitor
screenshots (3840×1080 on DP-2, 5120×1440 in scaled mode) are rejected by
Claude with "image exceeds dimension limits" — the ingestion cap is ~1568 px
on the longer side. The ticker itself is a 28 px strip at the bottom; this
tool captures exactly that region via `grim -g "0,1052 3840x28"`, producing
a ~25 KB PNG that always ingests cleanly.

```bash
# Current on-screen content
hg ticker shot                                    # → /tmp/ticker-shot.png

# Pin a specific stream, wait past the 400 ms wipe, shoot, auto-unpin
hg ticker shot calendar                           # → /tmp/ticker-shot.png
hg ticker shot github-prs /tmp/prs.png            # explicit path

# Passthrough flags (any flag ends the positional parse)
hg ticker shot --monitor DP-3                     # secondary instance
hg ticker shot load --scale 0.5                   # 1920×14 for extra-small file
hg ticker shot --print-geom                       # just echo "X,Y WxH"
```

The `--pin` path (`hg ticker shot <stream>`) saves any existing pin, applies
the requested one, sends SIGUSR1 to reload, sleeps 0.7 s past the wipe,
shoots, then restores the prior pin on EXIT via a trap — so an interrupted
shot never leaves the ticker stuck on a pinned stream.

For motion capture (effects tuning, scroll-speed iteration) use
`scripts/capture-window-gif.sh ticker 3` for a 3 s GIF — same cropping
principle, temporal output.

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
| POST | Progress bar | `progress_bar` | 2px bottom bar with dim rail + bright gradient fill showing dwell elapsed (all presets) |
| POST | Stream-change wipe | (automatic) | 400ms gradient sweep when rotation advances |

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

## Rotation timing

`_current_interval` determines how long each stream stays on screen. The
value is read from `STREAM_META[name].dwell` or `.refresh`, then clamped
to `[MIN_DWELL_S, MAX_DWELL_S]` (currently 12–75 s). The cap sizes to the
native ultrawide: at ~55 px/s scroll speed, content needs ~70 s to
traverse a 3840 px display, so the 75 s cap lets it finish before
rotating. The floor keeps pomodoro/recording streams (refresh=1 s)
readable rather than flashing past.

## Content Streams (39 total)

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
| `cpu` | `/sys/class/hwmon` (k10temp/coretemp/zenpower) + `cpufreq` | 10 s | cyan |
| `gpu` | `/tmp/bar-gpu-full.txt` (nvidia-smi snapshot) | 10 s | green |
| `top-procs` | `ps --sort=-%cpu` | 15 s | orange |
| `uptime` | `/proc/uptime` | 5 min | purple |
| `tmux` | `tmux list-sessions` | 60 s | lime |
| `workspace` | `hyprctl activeworkspace/activewindow/workspaces -j` | 5 s | magenta |
| `network` | `nmcli` + `/proc/net/dev` | 30 s | blue |
| `audio` | `pactl` / playerctl | 10 s | orange |
| `shader` | `hyprshade current` | 30 s | purple |
| `claude-sessions` | `~/.claude/projects/*/*.jsonl` scan (threaded) | 60 s | amber |
| `ci` | `/tmp/bar-ci.txt` (gh workflow runs) | 2 min | green |
| `hacker` | `/tmp/bar-hacker.txt` (zenquotes cached) | 10 min | red |

### Round 2 live-data streams
| Stream | Source | Refresh | Badge |
|--------|--------|---------|-------|
| `hn-top` | hacker-news.firebaseio.com/topstories | 10 min | orange |
| `github-prs` | `gh pr list` across manifest repos | 5 min | lime |
| `weather-alerts` | api.weather.gov/alerts/active | 15 min | severity-keyed |
| `cve-alerts` | `arch-audit` (Arch Security Team) | 1 h | red/orange/amber by severity |

`cve-alerts` shows an install hint when `arch-audit` (available via `pacman -S arch-audit`) isn't installed, so the stream degrades gracefully.

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
- `main.txt` (default) — all 39 streams cycling
- `coding.txt` — keybinds, system, fleet, ci, claude-sessions, token-burn, dirty-repos, failed-units, kernel-errors, workspace, github, notifications, updates, shader
- `focus.txt` — system, pomodoro, calendar, workspace, music, notifications

## Adding a stream (hybrid architecture)

The ticker supports three stream shapes, ordered by how much code you write:

### 1. Declarative TOML — zero Python

Best for cache-fed streams that render a summary line + optional list
without custom colour logic. Edit `ticker/streams.toml`:

```toml
[token-burn]
type = "cache_single"
path = "/tmp/bar-tokens.txt"
label = "\U000f01cb TOKENS"
color = "#fbbf24"
preset = "cyberpunk"
refresh = 60
empty_message = "no token data"
```

Supported `type` values: `cache_single` (one line, bold span) and
`cache_list` (summary + rotating-font list). See `scripts/lib/ticker_streams/__init__.py`
for the full set of TOML fields (list_limit, fail_keywords,
empty_is_success, summary_color, etc.). Loader at
`keybind-ticker.py::_load_toml_catalogue` registers the resulting
builder into `STREAMS` / `STREAM_META` / `FALLBACK_ORDER`.

### 2. Python plugin — `scripts/lib/ticker_streams/<name>.py`

Best for anything with custom fetch, formatting, or state. Drop a
single-file module following this contract:

```python
"""mystream — one-line summary of what this renders."""
from __future__ import annotations

from html import escape

import ticker_render as tr
from ticker_streams import FONTS  # if you need the font cycle

META = {"name": "mystream", "preset": None, "refresh": 30}
# Optional: META["slow"] = True    — runs on a background thread
# Optional: META["dwell"] = 20     — override dwell cap

_LABEL = "\U000f0000 MYSTREAM"

def build():
    ...
    return tr.dup("".join(parts)), segments_list
```

Filename stem becomes the stream name unless `META["name"]` overrides it
(needed for hyphenated names — `failed_units.py` → `"failed-units"`
stream). The plugin discovery loop (`_load_bundled_plugins` in
`keybind-ticker.py`) imports every `*.py` in that directory at startup
and registers each as a first-class stream. Import errors are logged
and skipped — a broken plugin can never crash the ticker.

Use `tr.badge(label, color)`, `tr.empty(label, color, message)`,
`tr.dup(markup)` from `ticker_render` for the standard chrome.

### 3. User-drop-in plugin — `~/.config/keybind-ticker/plugins/<name>.py`

For experimenting outside the repo. Older contract (`build_markup()`
instead of `build()`); see the Plugins section below.

### Cache-fed streams

Regardless of which shape you pick, a cache-fed stream also needs:

- `scripts/bar-<name>-cache.sh` — writes `/tmp/bar-<name>.txt`
- `systemd/bar-<name>.{service,timer}` pair
- entry in `install.sh::desktop_passive_units`

### Playlist placement

Add the stream name to one or more `ticker/content-playlists/*.txt`.
`--pin <stream>` only applies if the stream is in the active playlist's
`stream_order` — streams absent from `main.txt` (e.g. `recording`) can
only be pinned after switching to a playlist that contains them.

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

## New surfaces (Round 2 Phase D)

### window-label
`scripts/window-label.py` — per-window floating label overlay. 28px
layer-shell strip anchored TOP+LEFT+RIGHT on DP-2, subscribes to
`$XDG_RUNTIME_DIR/hypr/*/.socket2.sock` for `activewindow>>` events, and
renders `<class> · <title>` that fades in on focus change, holds 3s,
then fades out. Runs as `dotfiles-window-label.service`.

### fleet-sparkline
`scripts/fleet-sparkline.py` — compact 200×80 sparkline overlay (top-centre
of DP-2, margin_top=60) showing CPU / MEM / NET / GPU in a 2×2 grid. Samples
/proc/stat, /proc/meminfo, /proc/net/dev, and /tmp/bar-gpu.txt every 1 s,
keeps 60 samples. Runs as `dotfiles-fleet-sparkline.service`.

### Known gotcha: layer-shell adjacent surfaces
Two gtk4-layer-shell surfaces whose edges touch (surface A bottom = surface B
top) can trigger a GTK clip that erases up to ~half of surface B. Symptom:
the TOP portion of the lower surface paints transparent even though
`hyprctl layers` reports the correct size. **Mitigation**: leave ≥16px of
vertical gap between any two layer-shell surfaces on the same edge. That's
why `fleet-sparkline` uses `margin_top=60` (16px below `window-label` which
ends at y=32).

### Known gotcha: DSC-fallback clipping
If the Samsung ultrawide drops from its native 5120×1440@240 Hz scale=2 mode
into a DSC-fallback mode (3840×1080@120 Hz scale=1, often after a live VRR
config change or driver hiccup), layer-shell surfaces render with their top
portion clipped — the ticker shows ~12 px of duplicated/leaked content
above its 28 px strip. Symptom is visible as ghost text floating above
the bar, or taking a 40 px-tall grim capture including the strip + some
margin above it.

**Soft recovery** (try first — works when EDID still lists the native
mode in `hyprctl monitors -j | jq '.[] | select(.name=="DP-2").availableModes'`):

```bash
hg ticker recover-monitor        # DP-2 → 5120x1440@239.76 scale=2
```

The helper re-applies the monitors.conf setting at runtime via
`hyprctl keyword monitor` and restarts the three layer-shell services
(ticker, window-label, fleet-sparkline) so each re-reads geometry. It's
idempotent — running while already native is a no-op.

**Hardware recovery** (required when EDID no longer lists the native mode
— `recover-monitor` exits 2 and prints the procedure in that case):
power-cycle the Samsung per `.claude/rules/nvidia-wayland.md` §146.

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

## Plugins (Round 2 Phase E1)

Drop-in plugins extend the stream list without modifying `keybind-ticker.py`.
Each plugin is a single Python file at `~/.config/keybind-ticker/plugins/<name>.py`:

```python
"""example plugin"""
META = {"preset": None, "refresh": 5}

def build_markup():
    badge = '<span background="#29f0ff" foreground="#05070d" font_desc="Maple Mono NF CN Bold 10"> ECHO </span>  '
    body  = '<span font_desc="Maple Mono NF CN 11">  hello  ·</span>'
    return badge + body + badge + body, []
```

Each plugin must define a `META` dict (`preset` + `refresh`) and a
`build_markup()` returning `(markup_str, segments_list)`. Plugins register
as `plugin:<filename>` streams. Import failures are logged and skipped — a
broken plugin can never crash the ticker.

## Per-segment right-click menu (Phase E2)

The context menu shows a **Segment** section when a right-click lands on a
text segment:
- **Copy: <snippet>** — copies the segment text to the Wayland clipboard via `wl-copy`.
- **Open URL** — if the segment contains a `http(s)://` URL, `xdg-open`s it.
- **Dismiss <stream>** — advances past the current stream once; it will
  reappear on the next playlist rotation.

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
