# Desktop Control-Plane Integration

The desktop control-plane on Manjaro/Wayland uses `Hyprland` as the compositor and `Quickshell` as the sole owner of bar, ticker, dock, menus, and notifications. The 2026-04 staged migration from ironbar/hyprshell/hypr-dock/keybind-ticker is complete; the legacy stack is retired.

## The Stack

1. **Compositor**: `Hyprland` (Wayland). Configuration lives in `hyprland/`.
   - Dynamically managed via `hyprctl` for layout restoration and workspace switching.
   - Fallback exists for `Sway`.

2. **Shell**: `Quickshell` owns every desktop surface — bars, ticker, dock, menus, companion overlays, and notifications.
   - Started by `dotfiles-quickshell.service` (systemd `--user`) and reinforced by the Hyprland boot path.
   - Reads cache-backed status surfaces such as `/tmp/rg-status.json` and `/tmp/bar-*.txt` instead of calling slow fleet tooling inline.
   - Controlled through `hg shell <status|full-cutover|rollback>`; `shell-stack-mode.sh --json status` exposes the same state for agents. `full-cutover` is the default — `rollback` is an escape hatch that stops Quickshell, intended only while the legacy stack is still vendored in tree.
   - Stack mode is persisted under `$XDG_STATE_HOME/dotfiles/shell-stack/env` as a single line `SHELL_STACK_MODE=<mode>`; the runtime no longer carries per-surface cutover flags.
   - Hyprland startup runs `shell-stack-boot.sh`, which defaults to `full-cutover` when no mode is persisted.

3. **Notifications**: Quickshell's `NotificationServer` is the freedesktop.org Notifications daemon.
   - D-Bus activation for `org.freedesktop.Notifications` routes through `notification-daemon-launch.sh`, which ensures `dotfiles-quickshell.service` is up and waits for Quickshell to claim the bus name. Falls back to `swaync` only if Quickshell does not claim within the timeout — a safety net while swaync still ships.
   - Notification keybinds route through `notification-control.sh`; it prefers Quickshell IPC targets for center/DND/clear actions and falls back to swaync or the local bridge.
   - `notification-bridge.py` exposes the local history log to Quickshell. Quickshell's `NotificationServer` appends owned notifications directly via `appendEntry` → bridge → `history.jsonl`.
   - Notification action buttons render in `NotificationCenter.qml` and `NotificationPopups.qml` and round-trip through the original `NotificationAction.invoke()` over D-Bus.

4. **Ticker**: Quickshell-owned QML ticker.
   - `TickerService.qml` owns playlist rotation, pause/pin/shuffle/preset state, urgent mode, banner text, health snapshots, lock/record watcher polling, and cache/live stream commands.
   - `ticker-control.sh` is the stable control wrapper used by `hg ticker`, keybinds, and plugin hooks; it dispatches via Quickshell IPC.

5. **Launchers & Menus**: Quickshell-owned modal menus.
   - `MenuService.qml` and `MenuOverlay.qml` own app launch, window switch, power, emoji, agent-session, and clipboard-history menus.
   - `menu-control.sh` is the stable keybind wrapper; it dispatches through the Quickshell `menus` IPC target.

6. **Dock**: Quickshell-owned bottom dock.
   - `DockService.qml`, `Dock.qml`, and `quickshell-dock-data.sh` own pinned app groups plus unpinned Hyprland windows.
   - Left-click focuses an existing group or launches it; right/middle-click launches a new instance.

## MCP Integration (`dotfiles-mcp`)

The control plane is uniquely orchestrated by an MCP server (`dotfiles-mcp`), allowing LLM agents to visually and behaviorally manipulate the desktop:
- **State Reading**: Agents can read window layouts (`hypr_list_windows`), capture screenshots (`hypr_screenshot_window`), and inspect UI state.
- **State Writing**: Agents can click UI elements (`hypr_click`), type text (`hypr_type_text`), and manage window sizes and placements.
- **Service Reloads**: Actions like `dotfiles_cascade_reload` ensure that changes to `ironbar`, `quickshell`, `hyprland`, or `swaync` are reloaded safely and sequentially.

## IPC and Event Handling

Many desktop behaviors rely on Unix sockets or IPC:
- Hyprland's IPC socket is watched to trigger automatic workspace renames and desktop state refreshes when windows move.
- D-Bus is used for media control.
- Quickshell exposes `shell`, `menus`, `dock`, and `ticker` IPC targets for notification controls, modal menus, dock state/actions, and ticker state/actions used by desktop keybind wrappers.

## Quickshell Layout

The Quickshell config is repo-native and intentionally does not vendor external shells. It uses a small module split modeled after the strongest Quickshell-heavy references:
- `quickshell/services/` owns process-backed data, dock/ticker/menu state, notification bridge state, and persisted stack flags.
- `quickshell/modules/` owns Wayland layer surfaces: top bars, dock, ticker, ticker context menu, modal menu overlay, notification center/popups, and quick settings.
- `quickshell/components/` owns small reusable visual primitives such as badges and panel buttons.
