---
name: hypr_layouts
description: Orchestrate multi-window Hyprland layouts — dev setups, research environments, presentation modes — via live IPC dispatch and kitty terminal spawning.
triggers:
  - layout
  - dev setup
  - window arrangement
  - workspace layout
  - project layout
  - open project
  - arrange windows
---

# Hyprland Layout Orchestration

Compose multi-window layouts by combining Hyprland dispatchers, kitty terminal spawning, and workspace management. This skill fills a gap no public Claude Code skill covers — live IPC-driven window arrangement.

## MCP Tools

### Query state
- `hypr_list_windows` — all windows with address, class, title, workspace, position, size
- `hypr_list_workspaces` — workspaces with window count, monitor, focused status
- `hypr_get_monitors` — monitor resolution, position, scale, refresh rate
- `hypr_get_active_window` — currently focused window
- `hypr_get_active_workspace` — currently focused workspace

### Manipulate windows
- `hypr_focus_window` — focus by address or class
- `hypr_move_window` — move to exact pixel coordinates
- `hypr_resize_window` — resize to exact pixel dimensions
- `hypr_close_window` — close by address or class
- `hypr_toggle_floating` — toggle floating state
- `hypr_fullscreen_window` — toggle fullscreen/maximize
- `hypr_minimize_window` — send to special:minimized workspace
- `hypr_switch_workspace` — switch to workspace by ID
- `hypr_dispatch` — raw dispatcher for anything not covered above

### Spawn terminals
- `kitty_launch` — launch a new kitty window with optional command, cwd, title
- `kitty_set_title` — set window title for identification
- `kitty_set_layout` — change kitty's internal layout (stack, tall, fat, grid)

### Save/restore
- `hypr_layout_save` — snapshot current window positions to a named layout
- `hypr_layout_restore` — restore a saved layout
- `hypr_layout_list` — list saved layouts

## Workflow

### Creating a layout from scratch

1. **Survey**: `hypr_get_monitors` to know resolution and scale of each monitor
2. **Plan**: Calculate window positions based on monitor geometry. With hy3 tiling, new windows auto-tile — focus the target workspace first, then spawn
3. **Execute in order**:
   a. `hypr_switch_workspace` to the target workspace
   b. `kitty_launch` (or `hypr_dispatch` with `exec`) to spawn each window
   c. Wait briefly between spawns for hy3 to tile
   d. Use `hypr_move_window` / `hypr_resize_window` only if precise pixel placement is needed (floating windows)
4. **Verify**: `hypr_list_windows` to confirm all windows landed correctly
5. **Save**: `hypr_layout_save` with a descriptive name for future recall

### Common layout recipes

**Dev (2 workspaces, ultrawide)**
- WS 6: editor (left 60%) + terminal (right 40%) — hy3 auto-tiles this
- WS 7: browser + docs side-by-side
- WS 1 (portrait): monitoring/logs

**Research (single workspace, ultrawide)**  
- 3-column: browser left, notes center, terminal right
- All on WS 6 (ultrawide), hy3 equalizes with `$mod T`

**Presentation (focused)**
- Fullscreen browser on WS 6
- Speaker notes on WS 1 (portrait monitor)
- Terminal on WS 7 for demos

### Restoring a saved layout
1. `hypr_layout_list` to see available layouts
2. `hypr_layout_restore` with the layout name
3. `hypr_list_windows` to verify restoration

## Monitor Context

Two fixed monitors (from `hyprland/monitors.conf`):
- **DP-3**: Samsung LC49G95T ultrawide, 5120x1440@240Hz, scale 2 (effective 2560x720), workspaces 6-10
- **DP-2**: XEC ES-G32C1Q portrait, 2560x1440@180Hz, scale 2 (effective 720x1280), workspaces 1-5, rotated 270

## Key Conventions

- Use `split-workspace` dispatcher (not bare `workspace`) for per-monitor workspace switching
- hy3 is the active layout — windows auto-tile on spawn, use `hy3:equalize` to even them out
- For floating windows, set position relative to monitor origin (DP-3 starts at x=1810, DP-2 at x=4370)
- Keybind `$mod, R` toggles between hy3 and dwindle layouts
