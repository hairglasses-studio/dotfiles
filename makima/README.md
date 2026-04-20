# makima device profiles

Profile TOML for [makima](https://github.com/cyber-sushi/makima), the
evdev-based button-remap daemon. One file per physical device, named
by the device's `udev` product string (`DEVICE_NAME` in
`dotfiles/udev/50-makima.rules`).

## Current profiles

- `mx-master-4.toml` — Logitech MX Master 4 (Bluetooth LE). Maps the
  thumb buttons + gesture button to `hg ticker` verbs so the mouse
  side buttons drive the scrolling ticker directly.

## How to add a device

1. Pair and connect the device; find its evdev node under
   `/dev/input/by-id/` or via `libinput list-devices`.
2. Grab the device's `NAME`/`PRODUCT` from `udevadm info`.
3. Add an entry to `dotfiles/udev/50-makima.rules` pointing at this
   directory, or symlink the profile TOML into `/etc/makima/` if the
   rules infrastructure is disabled.
4. Restart `makima.service` — the daemon watches the profile dir and
   picks up new files without a reload.

## How to discover button codes

The MX Master 4 profile below was stubbed from documentation — the
exact `BTN_*` codes for the back / forward / gesture buttons may vary
by firmware revision. To verify live:

```bash
# While the mouse is connected, press each button and watch:
sudo evtest /dev/input/by-id/usb-Logitech_MX_Master_4-event-mouse
# Or, via the dotfiles MCP:
mcp__dotfiles__input_detect_controllers
mcp__dotfiles__input_controller_test
```

Update the `[[bindings]]` blocks in the profile with the actual codes
observed, then `systemctl --user restart makima.service`.
