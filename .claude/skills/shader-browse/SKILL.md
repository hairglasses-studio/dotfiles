---
name: shader-browse
description: Browse, preview, and manage the 132+ Ghostty shader collection
allowed-tools: Bash, Read, mcp__shader__shader_list, mcp__shader__shader_set, mcp__shader__shader_random, mcp__shader__shader_test, mcp__shader__shader_get_state, mcp__hyprland__hypr_screenshot
---

Interactive shader browser for the Ghostty terminal shader collection.

Commands (via $ARGUMENTS):
- `list [category]` — List all shaders, optionally filtered by category (crt, cursor, background, post-fx, watercolor)
- `current` — Show the currently active shader
- `set <name>` — Apply a shader by name and take a screenshot to preview
- `random` — Apply a random shader and screenshot
- `test [name]` — Compile-test a shader (or all shaders if no name)
- `cycle` — Apply the next shader in rotation
- (no args) — Show current shader and list categories with counts

Use MCP tools when available (shader_list, shader_set, etc.), fall back to shell scripts otherwise.
After applying a shader, take a screenshot to show the result.
