---
description: "Hyprland window, workspace, and desktop management. $ARGUMENTS: (empty)=list windows, 'workspaces'=list, 'focus <class>'=focus window, 'ws <id>'=switch workspace, 'monitors'=display config, 'move <class> <x> <y>'=move window, 'resize <class> <w> <h>'=resize, 'click <x> <y>'=click, 'type <text>'=type, 'key <combo>'=send keys, 'screenshot'=capture"
user_invocable: true
allowed-tools: mcp__dotfiles__hypr_list_windows, mcp__dotfiles__hypr_list_workspaces, mcp__dotfiles__hypr_get_monitors, mcp__dotfiles__hypr_screenshot, mcp__dotfiles__hypr_screenshot_monitors, mcp__dotfiles__hypr_screenshot_window, mcp__dotfiles__hypr_focus_window, mcp__dotfiles__hypr_switch_workspace, mcp__dotfiles__hypr_reload_config, mcp__dotfiles__hypr_click, mcp__dotfiles__hypr_type_text, mcp__dotfiles__hypr_key, mcp__dotfiles__hypr_set_monitor, mcp__dotfiles__hypr_move_window, mcp__dotfiles__hypr_resize_window, mcp__dotfiles__hypr_close_window, mcp__dotfiles__hypr_toggle_floating, mcp__dotfiles__hypr_minimize_window, mcp__dotfiles__hypr_fullscreen_window
---

Parse `$ARGUMENTS`:

**Query:**
- **(empty)** or **"windows"**: Call `mcp__dotfiles__hypr_list_windows` — all windows with title, class, workspace
- **"workspaces"**: Call `mcp__dotfiles__hypr_list_workspaces` — workspaces with window counts
- **"monitors"**: Call `mcp__dotfiles__hypr_get_monitors` — resolution, refresh, position, scale

**Window Management:**
- **"focus <class>"**: Call `mcp__dotfiles__hypr_focus_window` with `class=<class>`
- **"ws <id>"**: Call `mcp__dotfiles__hypr_switch_workspace` with `id=<id>`
- **"close <class>"**: Call `mcp__dotfiles__hypr_close_window` with `class=<class>`
- **"float"** or **"float <class>"**: Call `mcp__dotfiles__hypr_toggle_floating` — toggle floating state
- **"fullscreen"** or **"fullscreen <class>"**: Call `mcp__dotfiles__hypr_fullscreen_window` — toggle fullscreen/maximize
- **"minimize <class>"**: Call `mcp__dotfiles__hypr_minimize_window` — minimize to special:minimized workspace
- **"move <class> <x> <y>"**: Call `mcp__dotfiles__hypr_move_window` with `class=<class>, x=<x>, y=<y>` — move to exact pixel coordinates
- **"resize <class> <w> <h>"**: Call `mcp__dotfiles__hypr_resize_window` with `class=<class>, width=<w>, height=<h>` — resize to exact dimensions

**Input Simulation:**
- **"click <x> <y>"**: Call `mcp__dotfiles__hypr_click` with `x=<x>, y=<y>` — click at coordinates via ydotool
- **"type <text>"**: Call `mcp__dotfiles__hypr_type_text` with `text=<text>` — type text at cursor via wtype
- **"key <combo>"**: Call `mcp__dotfiles__hypr_key` with `keys=<combo>` — send key events via ydotool

**Config & Display:**
- **"reload"**: Call `mcp__dotfiles__hypr_reload_config` — reload config + check errors
- **"monitor <name> <resolution>"**: Call `mcp__dotfiles__hypr_set_monitor` — configure monitor resolution, position, or scale

**Screenshots:**
- **"screenshot"**: Call `mcp__dotfiles__hypr_screenshot` — capture all monitors
- **"screenshot <class>"**: Call `mcp__dotfiles__hypr_screenshot_window` with `class=<class>` — capture specific window
- **"screenshot monitors"**: Call `mcp__dotfiles__hypr_screenshot_monitors` — separate per-monitor captures

Present window lists as tables with columns: Class, Title, Workspace, Address.
