---
description: "Manage tmux sessions and workspaces. $ARGUMENTS can be: (empty)=list sessions, 'new <name>'=create session, 'kill <name>'=kill session, 'capture'=capture current pane, 'send <keys>'=send keys to pane, 'workspace <spec>'=create multi-pane layout"
user_invocable: true
---

Parse `$ARGUMENTS` to determine the action:

- **(empty)**: Call `mcp__dotfiles__tmux_list_sessions` and display all sessions with window counts
- **"new <name>"**: Call `mcp__dotfiles__tmux_new_session` with `name=<name>`
- **"kill <name>"**: Call `mcp__dotfiles__tmux_kill_session` with `name=<name>`
- **"capture"** or **"capture <target>"**: Call `mcp__dotfiles__tmux_capture_pane` with optional `target=<target>`. Display captured output.
- **"send <keys>"**: Call `mcp__dotfiles__tmux_send_keys` with `keys=<keys>` to the current pane
- **"windows"** or **"windows <session>"**: Call `mcp__dotfiles__tmux_list_windows` with optional `session=<session>`
- **"panes"**: Call `mcp__dotfiles__tmux_list_panes` to show all panes across sessions
- **"search <text>"**: Call `mcp__dotfiles__tmux_search_panes` with `text=<text>` to find text across all panes
- **"workspace <spec>"**: Call `mcp__dotfiles__tmux_workspace` with the spec to create a multi-pane development layout

Display results in a clean, formatted table.
