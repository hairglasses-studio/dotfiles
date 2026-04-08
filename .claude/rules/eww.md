---
paths:
  - "eww/**"
---

eww bar conventions:
- Use Snazzy palette colors (bg=#000000, fg=#f1f1f0, cyan=#57c7ff, magenta=#ff6ac1, green=#5af78e, yellow=#f3f99d, red=#ff5c57)
- Font: Maple Mono NF CN
- Validate .yuck files: balanced parentheses (PreToolUse hook checks this)
- Validate .scss files: `sassc --style compressed` (PreToolUse hook checks this)
- After editing: `eww reload` fires automatically via PostToolUse hook
