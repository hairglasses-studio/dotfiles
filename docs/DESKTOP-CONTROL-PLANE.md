# Desktop Control-Plane Integration

The desktop control-plane for Manjaro/Wayland uses `Hyprland` as the primary compositor, with `ironbar` as the rollback menubar and a staged `Quickshell` migration path for bar, ticker, menus, and notifications.

## The Stack

1. **Compositor**: `Hyprland` (Wayland). Configuration lives in `hyprland/`.
   - Dynamically managed via `hyprctl` for layout restoration and workspace switching.
   - Fallback exists for `Sway`.

2. **Menubar**: `ironbar` with `Quickshell` pilot/cutover
   - Configured in `ironbar/config.toml` and `ironbar/style.css`.
   - Started by `systemd --user` and reinforced by the Hyprland boot path.
   - Reads cache-backed status surfaces such as `/tmp/rg-status.json` instead of calling slow fleet tooling inline.
   - `dotfiles-quickshell.service` runs the QML shell and is controlled through `hg shell <status|pilot|bar-cutover|ticker-cutover|menu-cutover|notification-cutover|full-pilot|full-cutover|rollback>`; `shell-stack-mode.sh --json status` exposes the same state for agents.
   - Stack mode is persisted under `$XDG_STATE_HOME/dotfiles/shell-stack/` and consumed by `run-quickshell.sh` as `QS_BAR_CUTOVER`, `QS_TICKER_CUTOVER`, `QS_MENU_CUTOVER`, and `QUICKSHELL_NOTIFICATION_OWNER`; cutover modes restart Quickshell so the env is re-read.
   - Hyprland startup runs `shell-stack-boot.sh`, which applies the persisted owner mode instead of unconditionally starting ironbar, swaync, and the Python ticker.
   - Reload and palette propagation scripts read the same state through `scripts/lib/shell-stack.sh`, so cutover modes do not revive ironbar, swaync, or the Python ticker.
   - Quickshell now owns native MPRIS and StatusNotifier tray widgets in the pilot bar; Ironbar remains the rollback owner until `bar-cutover`/`full-cutover`.

3. **Notifications**: `swaync` (SwayNotificationCenter) and `mako` (fallback), with Quickshell history bridge.
   - Tightly integrated with the desktop to show alerts for long-running MCP tasks or system health checks.
   - D-Bus activation for `org.freedesktop.Notifications` routes through `notification-daemon-launch.sh`, which starts Quickshell when the persisted shell-stack mode makes it the notification owner and otherwise falls back to swaync.
   - Notification keybinds route through `notification-control.sh`; it prefers Quickshell IPC targets for center/DND/clear actions and falls back to swaync or the local bridge.
   - MCP notification tools use the same wrapper before falling back to raw `swaync-client`, keeping agent actions aligned with the active shell owner.
   - `notification-history-listener.py` remains the D-Bus observer; `notification-bridge.py` exposes the local history log to Quickshell without claiming `org.freedesktop.Notifications`.
   - The Quickshell shell renders notification history, DND, clear/close actions, and optional Quickshell-owned popups. `notification-cutover` stops swaync and enables Quickshell's `NotificationServer`; rollback restores swaync.

4. **Ticker**: Quickshell-owned QML ticker with legacy rollback.
   - `TickerService.qml` owns playlist rotation, pause/pin/shuffle/preset state, urgent mode, banner text, health snapshots, and cache/live stream commands.
   - `ticker-control.sh` is the stable control wrapper used by `hg ticker`, keybinds, lock/record watchers, and plugin hooks; it prefers Quickshell IPC and falls back to legacy DBus/state.
   - `keybind-ticker.py` remains the rollback owner when Quickshell is unavailable or `dotfiles-keybind-ticker.service` is explicitly restored.

5. **Launchers & Menus**: Quickshell-owned modal menus with fallback wrappers.
   - `MenuService.qml` and `MenuOverlay.qml` own app launch, window switch, power, emoji, agent-session, and clipboard-history menus.
   - `menu-control.sh` is the stable keybind wrapper; it prefers the Quickshell `menus` IPC target and falls back to `wofi`, `wlogout`, `clipse`, or the old Hyprland client picker if needed.
   - `menu-cutover` stops `dotfiles-hyprshell.service`; rollback restores it.

## MCP Integration (`dotfiles-mcp`)

The control plane is uniquely orchestrated by an MCP server (`dotfiles-mcp`), allowing LLM agents to visually and behaviorally manipulate the desktop:
- **State Reading**: Agents can read window layouts (`hypr_list_windows`), capture screenshots (`hypr_screenshot_window`), and inspect UI state.
- **State Writing**: Agents can click UI elements (`hypr_click`), type text (`hypr_type_text`), and manage window sizes and placements.
- **Service Reloads**: Actions like `dotfiles_cascade_reload` ensure that changes to `ironbar`, `quickshell`, `hyprland`, or `swaync` are reloaded safely and sequentially.

## IPC and Event Handling

Many desktop behaviors rely on Unix sockets or IPC:
- Hyprland's IPC socket is watched to trigger automatic workspace renames and desktop state refreshes when windows move.
- D-Bus is used for media control.
- Quickshell exposes `shell`, `menus`, and `ticker` IPC targets for notification controls, modal menus, and ticker state/actions used by desktop keybind wrappers.

## Quickshell Layout

The Quickshell config is repo-native and intentionally does not vendor external shells. It uses a small module split modeled after the strongest Quickshell-heavy references:
- `quickshell/services/` owns process-backed data, ticker/menu state, notification bridge state, and persisted stack flags.
- `quickshell/modules/` owns Wayland layer surfaces: top bars, ticker, ticker context menu, modal menu overlay, notification center/popups, and quick settings.
- `quickshell/components/` owns small reusable visual primitives such as badges and panel buttons.
