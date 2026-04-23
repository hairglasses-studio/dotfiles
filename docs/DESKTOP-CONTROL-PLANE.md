# Desktop Control-Plane Integration

The desktop control-plane for Manjaro/Wayland uses `Hyprland` as the primary compositor, with `ironbar` as the primary menubar and a staged `Quickshell` pilot for the bar/ticker/notification migration path. Repo-managed systemd user services own the companion layers.

## The Stack

1. **Compositor**: `Hyprland` (Wayland). Configuration lives in `hyprland/`.
   - Dynamically managed via `hyprctl` for layout restoration and workspace switching.
   - Fallback exists for `Sway`.

2. **Menubar**: `ironbar` with `Quickshell` pilot/cutover
   - Configured in `ironbar/config.toml` and `ironbar/style.css`.
   - Started by `systemd --user` and reinforced by the Hyprland boot path.
   - Reads cache-backed status surfaces such as `/tmp/rg-status.json` instead of calling slow fleet tooling inline.
   - `dotfiles-quickshell.service` runs the QML shell and is controlled through `hg shell <status|pilot|bar-cutover|ticker-cutover|notification-cutover|full-pilot|full-cutover|rollback>`; `shell-stack-mode.sh --json status` exposes the same state for agents.
   - Stack mode is persisted under `$XDG_STATE_HOME/dotfiles/shell-stack/` and consumed by `run-quickshell.sh` as `QS_BAR_CUTOVER`, `QS_TICKER_CUTOVER`, and `QUICKSHELL_NOTIFICATION_OWNER`; cutover modes restart Quickshell so the env is re-read.
   - Quickshell now owns native MPRIS and StatusNotifier tray widgets in the pilot bar; Ironbar remains the rollback owner until `bar-cutover`/`full-cutover`.

3. **Notifications**: `swaync` (SwayNotificationCenter) and `mako` (fallback), with Quickshell history bridge.
   - Tightly integrated with the desktop to show alerts for long-running MCP tasks or system health checks.
   - D-Bus activation for `org.freedesktop.Notifications` routes through `notification-daemon-launch.sh`, which starts Quickshell when the persisted shell-stack mode makes it the notification owner and otherwise falls back to swaync.
   - Notification keybinds route through `notification-control.sh`; it prefers Quickshell IPC targets for center/DND/clear actions and falls back to swaync or the local bridge.
   - `notification-history-listener.py` remains the D-Bus observer; `notification-bridge.py` exposes the local history log to Quickshell without claiming `org.freedesktop.Notifications`.
   - The Quickshell shell renders notification history, DND, clear/close actions, and optional Quickshell-owned popups. `notification-cutover` stops swaync and enables Quickshell's `NotificationServer`; rollback restores swaync.

4. **Ticker**: Python `keybind-ticker` with Quickshell bridge.
   - `keybind-ticker.py` remains the live owner of the bottom ticker during pilot mode.
   - `ticker-bridge.py` exposes the existing stream catalog as NDJSON for Quickshell QML consumption and supports long-running `--watch` mode.
   - The Quickshell ticker discovers the stream catalog, rotates streams, supports manual stream advance from the stream badge, and moves to the bottom exclusive layer in ticker/full cutover modes.

5. **Launchers & Menus**: `hyprshell`, `wofi`, and `wlogout`.
   - `hyprshell` remains the primary launcher and overview surface.
   - `wlogout` is the power menu layer.

## MCP Integration (`dotfiles-mcp`)

The control plane is uniquely orchestrated by an MCP server (`dotfiles-mcp`), allowing LLM agents to visually and behaviorally manipulate the desktop:
- **State Reading**: Agents can read window layouts (`hypr_list_windows`), capture screenshots (`hypr_screenshot_window`), and inspect UI state.
- **State Writing**: Agents can click UI elements (`hypr_click`), type text (`hypr_type_text`), and manage window sizes and placements.
- **Service Reloads**: Actions like `dotfiles_cascade_reload` ensure that changes to `ironbar`, `quickshell`, `hyprland`, or `swaync` are reloaded safely and sequentially.

## IPC and Event Handling

Many desktop behaviors rely on Unix sockets or IPC:
- Hyprland's IPC socket is watched to trigger automatic workspace renames and desktop state refreshes when windows move.
- D-Bus is used for media control.
- Quickshell exposes the `shell` IPC target for notification center, quick settings, DND, and status calls used by desktop keybind wrappers.

## Quickshell Layout

The Quickshell config is repo-native and intentionally does not vendor external shells. It uses a small module split modeled after the strongest Quickshell-heavy references:
- `quickshell/services/` owns process-backed data, ticker bridge state, notification bridge state, and persisted stack flags.
- `quickshell/modules/` owns Wayland layer surfaces: top bars, ticker, notification center/popups, and quick settings.
- `quickshell/components/` owns small reusable visual primitives such as badges and panel buttons.
