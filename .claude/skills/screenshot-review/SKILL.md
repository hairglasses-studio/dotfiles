---
name: screenshot-review
description: Take a screenshot of the desktop and analyze the visual rice quality
allowed-tools: Bash, Read
---

Take a screenshot of the current desktop and provide a detailed visual analysis.

Steps:
1. Capture the desktop: `screencapture -x /tmp/rice-review.png`
2. Read the screenshot image
3. Analyze and report on:
   - **Window layout** — Are windows tiled correctly by AeroSpace? Any overlap or gaps?
   - **Color consistency** — Does everything use the Snazzy palette (cyan #57c7ff, magenta #ff6ac1, green #5af78e)?
   - **Bar visibility** — Is SketchyBar visible? Are workspaces, clock, and widgets showing?
   - **Font rendering** — Is text crisp and readable? Any rendering artifacts?
   - **Shader effects** — Is a Ghostty shader active? Does it look good?
   - **CRT overlay** — Is RetroVisor running? How does the CRT effect look?
   - **Notifications** — Any error overlays or warnings visible?
   - **Overall aesthetic** — Rate the cyberpunk vibe 1-10

If $ARGUMENTS contains "fix", also suggest and implement improvements.
