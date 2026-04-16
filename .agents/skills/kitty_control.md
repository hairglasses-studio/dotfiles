---
name: kitty_control
description: Manage kitty terminal sessions — tabs, windows, themes, fonts, scrollback, images — via remote control MCP tools.
triggers:
  - kitty
  - terminal tab
  - terminal window
  - terminal theme
  - terminal font
  - scrollback
  - kitty layout
---

# Kitty Terminal Remote Control

Manage the kitty terminal emulator via its remote control protocol. 23 MCP tools cover the full kitty `kitten @` surface. No public Claude Code skill exists for kitty session management — this fills that gap.

## MCP Tools

### Session inspection
- `kitty_status` — running instances, OS window count, listen socket
- `kitty_list_windows` — all kitty windows with ID, title, PID, cwd, foreground process
- `kitty_list_tabs` — all tabs with window counts, active status, layout

### Window/tab management
- `kitty_launch` — spawn new window with optional command, cwd, title, location (vsplit, hsplit, tab, window)
- `kitty_focus_window` — focus by window ID
- `kitty_focus_tab` — focus by tab ID
- `kitty_close_window` — close by window ID
- `kitty_close_tab` — close by tab ID
- `kitty_set_title` — set window title
- `kitty_set_tab_title` — set tab title

### Layout and sizing
- `kitty_set_layout` — change tab layout (stack, tall, fat, grid, horizontal, vertical)
- `kitty_resize_window` — resize kitty window within the tab layout
- `kitty_resize_os_window` — resize the OS-level kitty window
- `kitty_last_used_layout` — switch to previously active layout

### Appearance
- `kitty_set_font_size` — change font size (absolute or delta)
- `kitty_set_theme` — apply a named kitty theme
- `kitty_set_opacity` — change background opacity (0.0-1.0)
- `kitty_load_config` — hot-reload kitty.conf without restarting

### I/O
- `kitty_send_text` — send text to a window (as if typed)
- `kitty_send_key` — send key events (Enter, Ctrl+C, etc.)
- `kitty_get_text` — capture text from a window's screen or scrollback
- `kitty_run_remote` — run arbitrary `kitten @` commands
- `kitty_show_image` — display an image inline using kitty graphics protocol

## Workflows

### Create a multi-pane dev session

1. `kitty_status` — confirm kitty is running
2. `kitty_list_tabs` — find or create the target tab
3. `kitty_launch` with `location=vsplit` — split current window vertically
4. `kitty_launch` with `location=hsplit` — split the new pane horizontally
5. `kitty_send_text` to each window — start the appropriate process (editor, server, logs)
6. `kitty_set_tab_title` — label the tab for identification

### Theme and font cycling

The dotfiles have shader and theme playlists managed by `kitty-shader-playlist`:
- `shader_cycle` — advance the DarkWindow shader playlist
- `shader_set` — apply a specific shader
- `shader_status` — current shader, theme, playlist position

For direct kitty appearance changes:
1. `kitty_set_theme` — apply a named theme (e.g., "Snazzy", "Tokyo Night")
2. `kitty_set_font_size` — adjust (delta: `+2`, `-1`, or absolute: `14`)
3. `kitty_set_opacity` — e.g., `0.92` for slight transparency

### Capture and analyze terminal content

1. `kitty_get_text` — capture visible screen or scrollback buffer
2. Parse the captured text for errors, logs, or specific patterns
3. Use `kitty_send_text` or `kitty_send_key` to respond (e.g., send a fix command)

### Display images and diagrams

`kitty_show_image` renders images inline using kitty's graphics protocol:
- Architecture diagrams during code review
- Screenshots for comparison during rice iteration
- Chart/graph output from CLI tools

## Key Context

- kitty config: `kitty/kitty.conf` (symlinked to `~/.config/kitty/kitty.conf`)
- `allow_remote_control = yes` is set — all `kitten @` commands work
- DarkWindow shaders: 139 GLSL shaders in `kitty/shaders/darkwindow/`
- Theme playlists: managed by `scripts/kitty-shader-playlist.sh`
- Scrollback in Neovim: `Ctrl+Shift+H` opens kitty-scrollback.nvim
- Font: Maple Mono NF CN (standardized across the rice)

## Conventions

- Always check `kitty_status` before sending commands — kitty may have been restarted
- Use `kitty_list_windows` to resolve window IDs rather than guessing
- Prefer `kitty_launch` with `location` parameter over manual splitting
- After theme/font changes, the shader playlist may need re-syncing via `shader_cycle`
