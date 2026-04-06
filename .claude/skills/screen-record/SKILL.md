---
description: "Screen recording via wf-recorder. $ARGUMENTS: (empty)=status, 'start'=begin recording, 'stop'=end recording"
user_invocable: true
allowed-tools: mcp__dotfiles__screen_record_start, mcp__dotfiles__screen_record_stop, mcp__dotfiles__screen_record_status
---

Parse `$ARGUMENTS`:

- **(empty)** or **"status"**: Call `mcp__dotfiles__screen_record_status` — show if recording, file path, duration
- **"start"**: Call `mcp__dotfiles__screen_record_start` — begin recording the screen. Report the output file path.
- **"stop"**: Call `mcp__dotfiles__screen_record_stop` — end recording. Report the saved file path and duration.

Present status clearly:
```
## Screen Recording

Status: {recording/idle}
File: {path}
Duration: {time}
```
