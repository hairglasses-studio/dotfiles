---
description: "Hyprland window and workspace management. $ARGUMENTS: (empty)=list windows, 'workspaces'=list workspaces, 'focus <class>'=focus window, 'ws <id>'=switch workspace, 'monitors'=display config"
user_invocable: true
---

Parse `$ARGUMENTS`:
- **(empty)** or **"windows"**: Call `mcp__dotfiles__hypr_list_windows` — all windows with title, class, workspace
- **"workspaces"**: Call `mcp__dotfiles__hypr_list_workspaces` — workspaces with window counts
- **"focus <class>"**: Call `mcp__dotfiles__hypr_focus_window` with `class=<class>`
- **"ws <id>"**: Call `mcp__dotfiles__hypr_switch_workspace` with `id=<id>`
- **"monitors"**: Call `mcp__dotfiles__hypr_get_monitors` — resolution, refresh, position, scale
- **"reload"**: Call `mcp__dotfiles__hypr_reload_config` — reload config + check errors
- **"close <class>"**: Call `mcp__dotfiles__hypr_close_window` with `class=<class>`
- **"fullscreen"**: Call `mcp__dotfiles__hypr_fullscreen_window` — toggle fullscreen
- **"float"**: Call `mcp__dotfiles__hypr_toggle_floating` — toggle floating state
