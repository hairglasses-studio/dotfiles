---
description: "Screen capture, annotation, and OCR. $ARGUMENTS: (empty)=full screenshot, 'window'=active window, 'ocr'=capture+extract text, 'annotate'=capture+edit, 'monitors'=per-monitor"
user_invocable: true
---

Parse `$ARGUMENTS`:
- **(empty)**: Call `mcp__dotfiles__hypr_screenshot` — full desktop capture
- **"window"** or **"window <class>"**: Call `mcp__dotfiles__hypr_screenshot_window` — capture specific window
- **"ocr"**: Call `mcp__dotfiles__screen_ocr` — capture region + extract text via tesseract
- **"annotate"**: Call `mcp__dotfiles__screen_screenshot_annotated` — capture + open in swappy
- **"monitors"**: Call `mcp__dotfiles__hypr_screenshot_monitors` — separate screenshot per monitor
- **"color"**: Call `mcp__dotfiles__screen_color_pick` — pick color from screen via hyprpicker
