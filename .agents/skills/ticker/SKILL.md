---
name: ticker
description: "Manage the Quickshell ticker — pin/unpin streams, switch playlists, pause/resume, tune presets (clean/minimal/ambient/cyberpunk), debug stream output, and inspect ticker state. Use when the user mentions the ticker, scrolling bar, bottom bar, or wants to adjust which stream is showing or how the marquee renders."
---

# Quickshell Ticker

The ticker is a Quickshell QML surface anchored to the bottom of the
primary monitor. `quickshell/services/TickerService.qml` owns playlist
rotation, stream fetching, banner/urgent state, lock + recording
watchers, and a health snapshot. `quickshell/modules/TickerBar.qml`
draws the marquee — a 30 px bar with stream-name pill, scrolling text,
GPU-FBO-rendered scanlines, and ghost-glitch overlay (cyberpunk preset).

## Architecture

- **Service**: `quickshell/services/TickerService.qml` — playlist + stream rotation, presets, banner/urgent state, watcher polling. Source of truth for current/pinned stream, paused, shuffle, preset.
- **View**: `quickshell/modules/TickerBar.qml` — layer-shell PanelWindow on bottom of primary monitor (DP-2 by default; override with `QS_MONITOR=DP-3` env). Subscribes to `TickerService.sweepRequested` for stream-change animation restart.
- **Context menu**: `quickshell/modules/TickerContextMenu.qml` — right-click on ticker pill opens a stream/playlist/preset picker.
- **Catalog**: `ticker/streams.toml` — declarative stream definitions (label, refresh, preset, shell command). 39 streams covering keybinds, system, fleet, weather, github, music, claude-sessions, ci, hacker, calendar, pomodoro, and more.
- **Playlists**: `ticker/content-playlists/{main,coding,focus,lock,recording}.txt` — one stream ID per line; `TickerService` rotates within the active playlist.
- **State** (host-local, written by QS): `~/.local/state/keybind-ticker/` — `active-playlist`, `pinned-stream`, `paused`, `shuffle`, `preset`, `pre-lock-playlist`, `pre-recording-playlist`. The `keybind-ticker` directory name is historical; no `keybind-ticker.py` runs here anymore.
- **IPC**: Quickshell exposes a `ticker` IPC target. `scripts/ticker-control.sh` is the wrapper users invoke; it calls `qs ipc call ticker <command>` under the hood.

## Operations

### Status
```bash
hg ticker status              # human-readable
hg ticker status --json       # machine-readable for scripts
```

### Pin / unpin a specific stream
```bash
hg ticker pin keybinds        # lock onto the keybinds stream
hg ticker unpin               # release pin, resume playlist rotation
hg ticker pin-toggle          # toggle the current stream's pin state
```

### Switch playlist
```bash
hg ticker playlist coding     # main / coding / focus / lock / recording
hg ticker list-playlists      # show available playlists
```

### Pause / shuffle / preset
```bash
hg ticker pause               # toggle pause
hg ticker shuffle on          # on | off | toggle
hg ticker preset cyberpunk    # clean | minimal | ambient | cyberpunk
```

### Banners + urgent
```bash
hg ticker banner "Build PASSED" "#3dffb5"   # flash a colored banner for 4.5s
hg ticker urgent true                        # red pulse mode (auto-clears after 10s)
hg ticker snooze-urgent                      # dismiss urgent mode early
```

### Stream advance / list
```bash
hg ticker next                # advance to next stream in playlist
hg ticker prev                # rewind to previous
hg ticker list-streams        # all 39 stream IDs
```

### Restart Quickshell (picks up TickerBar.qml or TickerService.qml edits)
```bash
hg ticker restart             # systemctl --user restart dotfiles-quickshell.service
hg ticker logs                # tail Quickshell journal for the last 2 minutes
```

### Reload state without restart
```bash
hg ticker reload              # re-read playlist + streams.toml without dropping QS
```

### Recover from monitor DSC fallback
If the Samsung ultrawide drops into a 2560×1440 fallback mode and the
ticker renders clipped, restore native 5120×1440@240 via the
runtime-keyword path:
```bash
hg ticker recover-monitor
```
The helper re-applies the `monitor =` keyword via `hyprctl` and
restarts Quickshell so layer-shell surfaces re-read geometry. If EDID
no longer lists the native mode, it exits with the hardware
power-cycle procedure (see `.claude/rules/nvidia-wayland.md`).

## Adding a new stream

1. Edit `ticker/streams.toml`: add a `[[streams]]` block with `id`,
   `label`, `refresh` (seconds), `preset`, and the `command` to run.
2. Add the stream ID to one of `ticker/content-playlists/*.txt` if
   you want it in active rotation.
3. `hg ticker reload` (or `hg ticker restart` for cleaner state).

## Per-stream behavior on click

`TickerBar.qml` translates click + scroll-wheel events on the marquee
into IPC calls — left-click advances, right-click opens the context
menu, scroll wheel rewinds/advances. Customizations live inline in
the QML rather than as a per-stream `_click_*` callback table.

## Lock / recording playlists

`TickerService.pollWatchers` runs a 2 s shell probe for `pgrep -x
hyprlock` and `pgrep -x wf-recorder|wl-screenrec`. On lock it saves
the current playlist to `~/.local/state/keybind-ticker/pre-lock-playlist`
and switches to `lock.txt`; on unlock it restores. Recording behaves
similarly with `recording.txt`. State is persisted across QS
restarts.

## Debugging

- `journalctl --user -u dotfiles-quickshell.service --since "5 min ago"` — QS reload + QML errors.
- `qs ipc call ticker status` — raw JSON without the human-readable wrapper.
- `qs ipc call ticker listStreams` — the stream catalog as Quickshell sees it.
- `hyprctl layers | grep quickshell` — confirms the bar is layer-shell-anchored.

## Visual presets

`TickerBar.qml`'s Canvas paints scanlines on cyberpunk + urgent presets;
all other presets show a static stripe with no per-frame repaint cost
(GPU-FBO render target, ~5 Hz when animated). Tweak `Canvas.onPaint`
constants to change density, opacity, or wave amplitude.

| Preset | Scanline | Ghost glitch | Sweep cost |
|---|---|---|---|
| `clean` | hidden | hidden | minimal |
| `minimal` | static | hidden | minimal |
| `ambient` | static | hidden | minimal |
| `cyberpunk` | animated 5 Hz | enabled | GPU FBO |

`urgent` mode overlays a red pulse on top of any preset.
