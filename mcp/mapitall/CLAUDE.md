# mapitall

Cross-platform Go event daemon for real-time controller mapping. Replaces makima (Rust, Linux-only). Reads the same unified TOML profiles that dotfiles-mcp manages, captures device events via mcpkit/device, and dispatches to all 11 output types.

## Build & Test
```bash
go build ./cmd/mapitall
go vet ./...
go test ./... -count=1
make check   # all three above
```

## Architecture

- `cmd/mapitall/` — CLI entry point (run, status, reload subcommands), SIGHUP handler
- `internal/daemon/` — Core loop: init providers, load profiles, run pipeline, hot-plug, active-app tracking
- `internal/pipeline/` — Event → Resolve → Transform → Dispatch hot path, custom modifier interception
- `internal/output/` — OutputTarget interface + implementations per output type
- `internal/ipc/` — Unix socket JSON-RPC 2.0 server + client
- `internal/reload/` — File watcher with debounced profile reload
- `internal/service/` — Platform service lifecycle (systemd, launchd)

## Dependencies

- `github.com/hairglasses-studio/mapping` — Shared types, engine, profile parsing, LUT transforms
- `github.com/hairglasses-studio/mcpkit` — Device provider framework (evdev, ALSA MIDI input)

## Output Types

| Type | Status | File | Platform |
|------|--------|------|----------|
| command | Done | output/command.go | All |
| osc | Done | output/osc.go | All |
| set_var | Done | output/variable.go | All |
| toggle_var | Done | output/variable.go | All |
| layer | Done | output/layer.go | All (ActiveLayer integration) |
| key | Done | key_{linux,darwin,windows}.go | Linux (uinput), macOS (CGO), Windows (SendInput) |
| movement | Done | movement_{linux,darwin,windows}.go | Linux (uinput), macOS (CGO), Windows (SendInput) |
| websocket | Done | websocket.go | All |
| midi_out | Done | midi_{linux,darwin,windows}.go | Linux (ALSA), macOS (CoreMIDI), Windows (winmm) |
| dbus | Done | dbus_linux.go + dbus_stub.go | Linux only |
| sequence | Done | output/sequence.go | All |

## Input Providers (mcpkit/device)

| Provider | Platform | Method |
|----------|----------|--------|
| evdev | Linux | Direct binary `input_event` reads from `/dev/input/eventN` |
| ALSA MIDI | Linux | Raw reads from `/dev/snd/midiC*D*` |
| IOKit HID | macOS | Done (CGO IOHIDManager polling) |
| CoreMIDI | macOS | Done (CGO pipe-based C→Go bridge) |
| XInput | Windows | Done (syscall xinput1_4.dll) |
| WinMM MIDI | Windows | Done (syscall winmm.dll callback) |

## Runtime Features

- **Custom modifiers** — Pipeline intercepts modifier buttons, updates EngineState instead of dispatching
- **Device grab** — `EVIOCGRAB` ioctl for exclusive evdev access (configured via `grab_device` in profile settings)
- **Hot-plug** — 3-second polling for device connect/disconnect, auto-starts pipeline goroutines
- **Active-app tracking** — 200ms polling for window class (Hyprland/Sway/X11 on Linux, lsappinfo on macOS)
- **SIGHUP reload** — `kill -HUP` triggers profile reload without restart
- **Layer switching** — `SetActiveLayer` per device via OutputAction

## IPC

JSON-RPC 2.0 over Unix socket. Methods: `status`, `reload`, `list_devices`, `list_profiles`, `get_state`, `set_variable`.

CLI subcommands (`mapitall status`, `mapitall reload`, etc.) use the IPC client.

## Service Files

- `mapitall.service` — systemd unit (Type=notify)
- `50-mapitall.rules` — udev rule for /dev/uinput access
- `com.hairglasses.mapitall.plist` — macOS launchd agent

## CGO

macOS outputs (key, movement, midi) use CGO for CoreGraphics, Carbon, and CoreMIDI frameworks. Linux outputs are pure Go (syscall/ioctl). Windows outputs use pure syscall (user32.dll SendInput, winmm.dll).
