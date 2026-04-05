---
description: "Notification center management. $ARGUMENTS: (empty)=count, 'dnd'=toggle DND, 'dismiss'=clear all, 'panel'=toggle panel, 'send <msg>'=send notification"
user_invocable: true
---

Parse `$ARGUMENTS`:
- **(empty)**: Call `mcp__dotfiles__notify_count` — current notification count
- **"dnd"**: Call `mcp__dotfiles__notify_dnd` with `action=toggle` — toggle Do Not Disturb
- **"dnd on/off"**: Call `mcp__dotfiles__notify_dnd` with `action=on/off`
- **"dismiss"**: Call `mcp__dotfiles__notify_dismiss` with `scope=all` — clear all notifications
- **"panel"**: Call `mcp__dotfiles__notify_panel` with `action=toggle` — toggle notification panel
- **"send <msg>"**: Call `mcp__dotfiles__notify_send` with `title=Claude, body=<msg>`
- **"history"**: Call `mcp__dotfiles__notify_history` — notification status snapshot
