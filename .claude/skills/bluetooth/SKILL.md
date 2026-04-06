---
description: "Manage Bluetooth devices. $ARGUMENTS: (empty)=list devices, 'scan'=discover nearby, 'connect <name>'=connect, 'disconnect <name>'=disconnect, 'pair <name>'=full pairing flow, 'info <name>'=device details, 'remove <name>'=forget, 'trust <name>'=trust/untrust, 'battery'=battery levels, 'power on/off'=toggle adapter"
user_invocable: true
allowed-tools: mcp__dotfiles__bt_list_devices, mcp__dotfiles__bt_device_info, mcp__dotfiles__bt_scan, mcp__dotfiles__bt_pair, mcp__dotfiles__bt_connect, mcp__dotfiles__bt_disconnect, mcp__dotfiles__bt_remove, mcp__dotfiles__bt_trust, mcp__dotfiles__bt_power, mcp__dotfiles__bt_discover_and_connect
---

Parse `$ARGUMENTS`:

- **(empty)**: Call `mcp__dotfiles__bt_list_devices` — show all paired devices with connection status and battery levels
- **"scan"**: Call `mcp__dotfiles__bt_scan` — discover nearby devices (8s timeout). Display name, address, RSSI.
- **"connect <name>"**: Call `mcp__dotfiles__bt_connect` with `device=<name>` — connect with BLE retry logic, resolves names against all known devices
- **"disconnect <name>"**: Call `mcp__dotfiles__bt_disconnect` with `device=<name>`
- **"pair <name>"**: Call `mcp__dotfiles__bt_discover_and_connect` with `device=<name>` — full flow: scan→remove stale bonds→pair (with agent)→trust→connect (with retry)
- **"info <name>"**: Call `mcp__dotfiles__bt_device_info` with `device=<name>` — detailed info: battery, profiles, trust, UUIDs
- **"remove <name>"**: Call `mcp__dotfiles__bt_remove` with `device=<name>` — forget a paired device
- **"trust <name>"**: Call `mcp__dotfiles__bt_trust` with `device=<name>` — toggle trust (auto-connect on proximity)
- **"battery"**: Call `mcp__dotfiles__bt_list_devices` and display only battery levels in a compact table
- **"power on"** or **"power off"**: Call `mcp__dotfiles__bt_power` with `action=on/off` — toggle BT adapter power

Present device lists as tables with columns: Name, Status, Battery, Type.
