# Drop CTRL — Hyprland Focus Trio

Firmware keymaps for the Drop CTRL (v1 and v2) that remap the top-right trio (Print Screen / Scroll Lock / Pause) to Hyprland window focus controls.

## Key Remaps

```
[PrtSc]  [ScrLk]  [Pause]
  F13     Enter     F14
 focus←  confirm   focus→
```

| Physical Key | Default | Remapped To | Purpose |
|-------------|---------|-------------|---------|
| Print Screen | `KC_PSCR` | `KC_F13` | Hyprland `movefocus l` |
| Scroll Lock | `KC_SCRL` | `KC_ENT` | Enter / submit prompt |
| Pause/Break | `KC_PAUS` | `KC_F14` | Hyprland `movefocus r` |

These are indices **13, 14, 15** in the VIA layer array.

## Firmware Files

| File | Hardware | VIA vendorProductId |
|------|----------|-------------------|
| `drop-ctrl-v1.json` | Drop CTRL v1 + High-Profile (Microchip MCU) | 81325778 |
| `drop-ctrl-v2.json` | Drop CTRL v2 (Drop MCU) | 899350537 |

## Flashing via VIA

1. Open [usevia.app](https://usevia.app) in Chrome/Chromium (requires WebHID)
2. Connect the Drop CTRL via USB
3. If VIA doesn't auto-detect: Settings > "Show Design Tab" > load the keyboard definition from [the-via/keyboards](https://github.com/the-via/keyboards/tree/master/v3/drop/ctrl)
4. Go to **Save+Load** > **Load Saved Layout** > select the matching JSON (`drop-ctrl-v1.json` or `drop-ctrl-v2.json`)
5. Verify the three keys changed in the VIA UI
6. Done — VIA writes to EEPROM in real-time (no separate flash step)

## Fn Layer (Layer 1)

The Fn layer preserves Drop's defaults plus RGB controls:

| Fn + Key | Function |
|----------|----------|
| Fn + PrtSc | `KC_MUTE` (audio mute) |
| Fn + W | `RGB_TOG` (toggle RGB) |
| Fn + E | `RGB_VAI` (brightness+) |
| Fn + R | `RGB_SPI` (speed+) |
| Fn + T | `RGB_HUI` (hue+) |
| Fn + Y | `RGB_SAI` (saturation+) |
| Fn + S | `RGB_MOD` (next effect) |
| Fn + D | `RGB_VAD` (brightness-) |
| Fn + F | `RGB_SPD` (speed-) |
| Fn + G | `RGB_HUD` (hue-) |
| Fn + H | `RGB_SAD` (saturation-) |
| Fn + Z | `RGB_M_P` (plain) |
| Fn + X | `RGB_M_B` (breathing) |
| Fn + C | `RGB_M_R` (rainbow) |
| Fn + V | `RGB_M_SW` (swirl) |
| Fn + B | `QK_BOOT` (bootloader) |
| Fn + N | `NK_TOGG` (N-key rollover) |
| Fn + Space | `EE_CLR` (EEPROM clear) |
| Fn + Ins | `KC_MPLY` (play/pause) |
| Fn + Home | `KC_MSTP` (stop) |
| Fn + PgUp | `KC_VOLU` (volume+) |
| Fn + Del | `KC_MPRV` (prev track) |
| Fn + End | `KC_MNXT` (next track) |
| Fn + PgDn | `KC_VOLD` (volume-) |

## Cyberpunk RGB Setup

Same as the Keychron — set via the VIA **Lighting tab** after loading the keymap:

| Setting | Value |
|---------|-------|
| Effect | Digital Rain |
| Hue | 128 (cyan) |
| Saturation | 255 |
| Brightness | 220 |
| Speed | 100 |

Cycle effects with **Fn + S** (`RGB_MOD`).

## Hyprland Integration

The F13/F14 binds are in `hyprland/hyprland.conf`:

```ini
bind = , F13, movefocus, l
bind = , F14, movefocus, r
```

These are shared with the Keychron V1 Ultra — both keyboards send the same keycodes.

Screenshot binds moved from `Print` to `$mod+S` (since PrtSc now sends F13):

```ini
bind = $mod, S, exec, wayshot --stdout | wl-copy
bind = $mod SHIFT, S, exec, wayshot -s "$(slurp)" --stdout | wl-copy
bind = $mod CTRL, S, exec, screenshot-crop.sh
```

## Identifying Your Hardware Version

Check VIA's auto-detection, or look at the PCB:
- **v1**: says "Massdrop" or "Drop" with Microchip/Atmel MCU (ATSAM)
- **v2**: newer PCB with Drop's own MCU, shipped from ~2023+

If unsure, try v1 first — it's the more common variant.

## Testing

| Test | Expected |
|------|----------|
| Tap PrtSc (top-left of trio) | Focus moves left |
| Tap Pause (top-right of trio) | Focus moves right |
| Tap ScrLk (middle of trio) | Sends Enter |
| `$mod + S` | Full screenshot to clipboard |
| `$mod + SHIFT + S` | Region screenshot |
| Fn + S | Next RGB effect |
| Fn + W | Toggle RGB on/off |
| `wev -f wl_keyboard:key` | Shows F13/F14 keycodes |
