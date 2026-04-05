# Keyboard Firmware

## Keychron V1 Ultra 8K (ANSI, Encoder)

Firmware: `keychron-v1-ultra-8k.json` — Keychron Launcher keymap export with custom encoder mappings and RGB controls.

### Encoder Map

| Layer | Rotate Left | Rotate Right | Press |
|-------|------------|-------------|-------|
| 0 (Base) | `KC_F13` → Hyprland focus left | `KC_F14` → Hyprland focus right | `KC_ENT` (Enter) |
| 1 (Fn) | `RGB_RMOD` → prev effect | `RGB_MOD` → next effect | `RGB_MOD` → next effect |
| 2 (Mac Base) | `KC_VOLD` → volume down | `KC_VOLU` → volume up | _(unassigned)_ |
| 3 (Mac Fn) | `RGB_RMOD` → prev effect | `RGB_MOD` → next effect | _(unassigned)_ |

### Hyprland Integration

F13/F14 are spare HID keycodes that don't conflict with anything. The encoder sends them without modifiers, and Hyprland binds them to `movefocus`:

```ini
# In hyprland/hyprland.conf
bind = , F13, movefocus, l
bind = , F14, movefocus, r
```

This lets you rotate the encoder to cycle window focus across tiled terminals (e.g. multiple Claude Code sessions) and press the encoder to send Enter.

### Flashing

1. Open [launcher.keychron.com](https://launcher.keychron.com) in Chrome/Chromium (requires WebHID)
2. Connect the V1 Ultra 8K via USB
3. Click **Import** and select `keychron-v1-ultra-8k.json`
4. Click **Flash**
5. Run `hyprctl reload` (or `SUPER+SHIFT+R`) to pick up the F13/F14 binds

### Cyberpunk RGB Lighting

RGB effect/color/speed are stored in the keyboard's EEPROM (not in the JSON). After flashing the keymap, configure lighting via the Keychron Launcher **Lighting tab** or physical shortcuts:

#### Recommended: "Tron Hacker" Profile

| Setting | Value |
|---------|-------|
| Effect | Digital Rain |
| Hue | 128 (cyan) |
| Saturation | 255 |
| Brightness | 220 |
| Speed | 100 (slow drip) |

#### Other Cyberpunk Effects

| Effect | Best Hue | Vibe |
|--------|----------|------|
| Digital Rain | 128 (cyan) or 85 (green) | Matrix / Tron Legacy |
| Typing Heatmap | 128 (cyan) | Thermal vision |
| Reactive Multinexus | 213 (magenta) | Neon grid pulses on keypress |
| Splash | 213 (magenta) | Rainbow ripple bursts |
| Cycle Spiral | any (full spectrum) | Hypnotic vortex |
| Pixel Rain | any | Neon cityscape flicker |

#### HSV Quick Reference (QMK 0-255 scale)

| Color | Hue | Use Case |
|-------|-----|----------|
| Cyan | 128 | Tron neon, daily driver |
| Neon Green | 85 | Classic Matrix hacker |
| Magenta | 213 | Hot pink neon |
| Electric Purple | 191 | Deep cyberpunk |
| Blade Runner Orange | 15 | Warm amber accent |

#### Physical Shortcuts

- `Fn + Tab` — toggle RGB on/off
- `Fn + rotate encoder` — cycle through effects
- `Fn + Q-row` — adjust hue, saturation, brightness, speed (see Fn layer)

### Fn Layer RGB Keys (Layer 1)

```
Tab  → RGB_TOG (toggle)
Q    → RGB_MOD (next effect)    A → RGB_RMOD (prev effect)
W    → RGB_VAI (brightness+)    S → RGB_VAD (brightness-)
E    → RGB_HUI (hue+)           D → RGB_HUD (hue-)
R    → RGB_SAI (saturation+)    F → RGB_SAD (saturation-)
T    → RGB_SPI (speed+)         G → RGB_SPD (speed-)
```

### Testing Checklist

| Test | Expected |
|------|----------|
| Rotate encoder (no Fn) | Terminal focus shifts left/right |
| Press encoder | Sends Enter |
| Normal typing | No interference from F13/F14 |
| `Fn + rotate encoder` | RGB effect changes live |
| `Fn + press encoder` | Steps to next RGB effect |
| `Fn + Tab` | Toggles RGB on/off |
| Volume keys (F11/F12) | Still work |
| `wev -f wl_keyboard:key` | Shows F13/F14 on encoder rotate |

### Keycode Reference

| Decimal | Hex | QMK Name | Used For |
|---------|-----|----------|----------|
| 40 | 0x28 | KC_ENT | Encoder press (layer 0) |
| 104 | 0x68 | KC_F13 | Encoder rotate left (layer 0) |
| 105 | 0x69 | KC_F14 | Encoder rotate right (layer 0) |
| 30752 | 0x7820 | RGB_TOG | Toggle RGB |
| 30753 | 0x7821 | RGB_MOD | Next RGB effect |
| 30754 | 0x7822 | RGB_RMOD | Prev RGB effect |

---

## Drop CTRL (TKL, v1 + v2)

Firmware: `drop-ctrl-v1.json` (Microchip MCU) and `drop-ctrl-v2.json` (Drop MCU) — VIA keymap exports with the top-right trio remapped to match the Keychron encoder behavior.

### Key Remaps

The Print Screen / Scroll Lock / Pause trio mirrors the Keychron encoder layout:

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

### Hyprland Integration

Same F13/F14 binds as the Keychron — both keyboards send identical keycodes:

```ini
bind = , F13, movefocus, l
bind = , F14, movefocus, r
```

Since Print Screen no longer sends `KC_PSCR`, screenshot binds use `$mod+S`:

```ini
bind = $mod, S, exec, wayshot --stdout | wl-copy
bind = $mod SHIFT, S, exec, wayshot -s "$(slurp)" --stdout | wl-copy
bind = $mod CTRL, S, exec, screenshot-crop.sh
```

### Flashing via VIA

See [`drop-ctrl.md`](drop-ctrl.md) for detailed flashing instructions, Fn layer reference, cyberpunk RGB setup, and hardware version identification.

### Testing Checklist

| Test | Expected |
|------|----------|
| Tap PrtSc (top-left of trio) | Focus moves left |
| Tap Pause (top-right of trio) | Focus moves right |
| Tap ScrLk (middle of trio) | Sends Enter |
| `$mod + S` | Full screenshot to clipboard |
| `$mod + SHIFT + S` | Region screenshot |
| `wev -f wl_keyboard:key` | Shows F13/F14 keycodes |
