---
name: shader-browse
description: "Browse, preview, and manage the 138 Ghostty shaders + wallpaper shaders. $ARGUMENTS: (empty)=status, 'list'=all shaders, 'set <name>'=apply, 'next'=cycle, 'random'=random, 'test'=compile all, 'build'=preprocess, 'playlist'=manage playlists, 'wallpaper'=live wallpapers"
allowed-tools: mcp__dotfiles__shader_list, mcp__dotfiles__shader_set, mcp__dotfiles__shader_cycle, mcp__dotfiles__shader_random, mcp__dotfiles__shader_status, mcp__dotfiles__shader_meta, mcp__dotfiles__shader_test, mcp__dotfiles__shader_build, mcp__dotfiles__shader_playlist, mcp__dotfiles__shader_get_state, mcp__dotfiles__wallpaper_set, mcp__dotfiles__wallpaper_random, mcp__dotfiles__wallpaper_list, mcp__dotfiles__hypr_screenshot
---

Full shader pipeline management for Ghostty terminal shaders and live wallpaper shaders.

Parse `$ARGUMENTS`:

- **(empty)** or **"status"**: Call `mcp__dotfiles__shader_status` — current shader, animation state, playlist position, auto-rotate status
- **"list"** or **"list <category>"**: Call `mcp__dotfiles__shader_list` — all shaders, optionally filtered by category (crt, cursor, background, post-fx, watercolor)
- **"set <name>"**: Call `mcp__dotfiles__shader_set` with `shader=<name>` — apply shader via atomic config write. Then `hypr_screenshot` to preview.
- **"next"** or **"cycle"**: Call `mcp__dotfiles__shader_cycle` — advance to next in playlist. Then screenshot.
- **"random"**: Call `mcp__dotfiles__shader_random` — pick and apply random shader. Then screenshot.
- **"current"**: Call `mcp__dotfiles__shader_get_state` — read active shader from Ghostty config
- **"test"** or **"test <name>"**: Call `mcp__dotfiles__shader_test` — compile-test via glslangValidator (single or all)
- **"build"**: Call `mcp__dotfiles__shader_build` — preprocess shaders (inline includes)
- **"meta <name>"**: Call `mcp__dotfiles__shader_meta` — full manifest metadata (category, cost, source, playlists)
- **"playlist"** or **"playlist <name>"**: Call `mcp__dotfiles__shader_playlist` — list playlists or pick random from one
- **"wallpaper"**: Call `mcp__dotfiles__wallpaper_list` — list available wallpaper shaders
- **"wallpaper set <name>"**: Call `mcp__dotfiles__wallpaper_set` with `shader=<name>`
- **"wallpaper random"**: Call `mcp__dotfiles__wallpaper_random`

After applying any shader, take a screenshot with `hypr_screenshot` to show the result.
