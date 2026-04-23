# Desktop Control-Plane Integration

The desktop control-plane for Manjaro/Wayland uses `Hyprland` as the primary compositor, with `ironbar` as the primary menubar and a staged `Quickshell` pilot for the bar/ticker/notification migration path. Repo-managed systemd user services own the companion layers.

## The Stack

1. **Compositor**: `Hyprland` (Wayland). Configuration lives in `hyprland/`.
   - Dynamically managed via `hyprctl` for layout restoration and workspace switching.
   - Fallback exists for `Sway`.

2. **Menubar**: `ironbar` with `Quickshell` pilot
   - Configured in `ironbar/config.toml` and `ironbar/style.css`.
   - Started by `systemd --user` and reinforced by the Hyprland boot path.
   - Reads cache-backed status surfaces such as `/tmp/rg-status.json` instead of calling slow fleet tooling inline.
   - `dotfiles-quickshell.service` runs the QML pilot in parallel and is controlled through `hg shell <status|pilot|bar-cutover|ticker-cutover|full-pilot|rollback>`; `shell-stack-mode.sh --json status` exposes the same state for agents.

3. **Notifications**: `swaync` (SwayNotificationCenter) and `mako` (fallback), with Quickshell history bridge.
   - Tightly integrated with the desktop to show alerts for long-running MCP tasks or system health checks.
   - `notification-history-listener.py` remains the D-Bus observer; `notification-bridge.py` exposes the local history log to Quickshell without claiming `org.freedesktop.Notifications`.
   - The Quickshell pilot renders a click-to-toggle notification history panel from that bridge, so notification-center replacement can be tested before daemon ownership moves.

4. **Ticker**: Python `keybind-ticker` with Quickshell bridge.
   - `keybind-ticker.py` remains the live owner of the bottom ticker during pilot mode.
   - `ticker-bridge.py` exposes the existing stream catalog as NDJSON for Quickshell QML consumption.
   - The Quickshell ticker pilot discovers the stream catalog, rotates streams, and supports manual stream advance from the stream badge.

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
