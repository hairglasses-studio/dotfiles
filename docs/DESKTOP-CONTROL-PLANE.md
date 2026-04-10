# Desktop Control-Plane Integration

The desktop control-plane for Manjaro/Wayland uses `Hyprland` as the primary compositor, managed via an array of scripts and MCP tools.

## The Stack

1. **Compositor**: `Hyprland` (Wayland). Configuration lives in `hyprland/`. 
   - Dynamically managed via `hyprctl` for layout restoration and workspace switching.
   - Fallback exists for `Sway`.

2. **Bars & Widgets**: `eww`
   - Configured in `eww/eww.yuck` and SCSS files.
   - Interacts with system state via custom shell scripts that emit JSON streams (for CPU, memory, notifications).
   - Eww variables are updated via `eww update` to reflect live state (like active shader, currently playing track, etc).

3. **Notifications**: `swaync` (SwayNotificationCenter) and `mako` (fallback).
   - Tightly integrated with the desktop to show alerts for long-running MCP tasks or system health checks.

4. **Launchers & Menus**: `wofi` and `wlogout`.
   - Themed dynamically to match the current color palette (Snazzy-on-black).

## MCP Integration (`dotfiles-mcp`)

The control plane is uniquely orchestrated by an MCP server (`dotfiles-mcp`), allowing LLM agents to visually and behaviorally manipulate the desktop:
- **State Reading**: Agents can read window layouts (`hypr_list_windows`), capture screenshots (`hypr_screenshot_window`), and inspect UI state.
- **State Writing**: Agents can click UI elements (`hypr_click`), type text (`hypr_type_text`), and manage window sizes and placements.
- **Service Reloads**: Actions like `dotfiles_cascade_reload` ensure that changes to `eww`, `hyprland`, or `mako` are reloaded safely and sequentially.

## IPC and Event Handling

Many desktop behaviors rely on Unix sockets or IPC:
- Hyprland's IPC socket is watched to trigger automatic workspace renames or bar updates when windows move.
- D-Bus is used for battery metrics (`juhradial`) and media control.
