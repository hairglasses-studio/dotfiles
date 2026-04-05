---
description: "Manage audio devices and volume. $ARGUMENTS can be: (empty)=show status, a number like '75'=set volume %, 'mute'=toggle mute, 'devices'=list devices, a device name=switch to that device"
user_invocable: true
---

Parse `$ARGUMENTS` to determine the action:

- **(empty)**: Call `mcp__dotfiles__audio_status` and display current sink/source, volume, mute state
- **Number (e.g. "75")**: Call `mcp__dotfiles__audio_volume` with `volume=$ARGUMENTS`
- **"+5" or "-10"**: Call `mcp__dotfiles__audio_volume` with relative adjustment
- **"mute"**: Call `mcp__dotfiles__audio_mute` with `action=toggle`
- **"devices"**: Call `mcp__dotfiles__audio_devices` and list all sinks/sources
- **Any other string**: Assume it's a device name. Call `mcp__dotfiles__audio_device_switch` with `device=$ARGUMENTS`

After any write action, call `mcp__dotfiles__notify_send` with title="Audio" and body showing the new state.
