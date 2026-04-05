---
description: "Clipboard operations (Wayland). $ARGUMENTS: (empty)=read text, 'write <text>'=copy to clipboard, 'image'=read image from clipboard"
user_invocable: true
---

Parse `$ARGUMENTS`:
- **(empty)** or **"read"**: Call `mcp__dotfiles__clipboard_read` — read current clipboard text
- **"write <text>"**: Call `mcp__dotfiles__clipboard_write` with `text=<text>` — copy to clipboard
- **"image"**: Call `mcp__dotfiles__clipboard_read_image` — read image as base64 PNG
