---
name: screenshot-review
description: Take a screenshot of the desktop and analyze the visual rice quality
allowed-tools: Bash, Read, mcp__hyprland__hypr_screenshot, mcp__hyprland__hypr_list_windows, mcp__hyprland__hypr_list_workspaces
---

Take a screenshot of the current desktop and provide a detailed visual analysis.

Steps:
1. Use the `hypr_screenshot` MCP tool to capture the desktop (or fall back to `grim /tmp/rice-review.png`)
2. Read the screenshot image
3. Analyze and report on:
   - **Window layout** — Are windows tiled correctly? Any overlap or gaps?
   - **Color consistency** — Does everything use the Snazzy palette (cyan #57c7ff, magenta #ff6ac1, green #5af78e)?
   - **Bar visibility** — Is the eww bar visible? Are workspaces, clock, and metrics showing?
   - **Font rendering** — Is text crisp and readable? Any rendering artifacts?
   - **Shader effects** — Is a Ghostty shader active? Does it look good?
   - **Wallpaper** — Is a wallpaper or shader wallpaper visible?
   - **Notifications** — Any error overlays or warnings visible?
   - **Overall aesthetic** — Rate the cyberpunk vibe 1-10

If $ARGUMENTS contains "fix", also suggest and implement improvements.
