# streamdeck

> Elgato Stream Deck physical control surface integration

**10 tools**

## Tools

- [`aftrs_streamdeck_brightness`](#aftrs-streamdeck-brightness)
- [`aftrs_streamdeck_buttons`](#aftrs-streamdeck-buttons)
- [`aftrs_streamdeck_clear`](#aftrs-streamdeck-clear)
- [`aftrs_streamdeck_devices`](#aftrs-streamdeck-devices)
- [`aftrs_streamdeck_health`](#aftrs-streamdeck-health)
- [`aftrs_streamdeck_refresh`](#aftrs-streamdeck-refresh)
- [`aftrs_streamdeck_reset`](#aftrs-streamdeck-reset)
- [`aftrs_streamdeck_set_color`](#aftrs-streamdeck-set-color)
- [`aftrs_streamdeck_set_image`](#aftrs-streamdeck-set-image)
- [`aftrs_streamdeck_status`](#aftrs-streamdeck-status)

---

## aftrs_streamdeck_brightness

Set device brightness.

**Complexity:** simple

**Tags:** `streamdeck`, `brightness`, `display`

**Use Cases:**
- Adjust brightness
- Dim display

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `brightness` | number | Yes | Brightness 0-100 |
| `device` | number |  | Device index (default 0) |

### Example

```json
{
  "brightness": 0,
  "device": 0
}
```

---

## aftrs_streamdeck_buttons

Get button states for a device.

**Complexity:** simple

**Tags:** `streamdeck`, `buttons`, `state`

**Use Cases:**
- Get button layout
- Check button states

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `device` | number |  | Device index (default 0) |

### Example

```json
{
  "device": 0
}
```

---

## aftrs_streamdeck_clear

Clear button(s) to black.

**Complexity:** simple

**Tags:** `streamdeck`, `clear`, `reset`

**Use Cases:**
- Clear button
- Clear all buttons

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `all` | boolean |  | Clear all buttons |
| `button` | number |  | Button index (omit to clear all) |
| `device` | number |  | Device index (default 0) |

### Example

```json
{
  "all": false,
  "button": 0,
  "device": 0
}
```

---

## aftrs_streamdeck_devices

List all connected Stream Deck devices.

**Complexity:** simple

**Tags:** `streamdeck`, `devices`, `list`

**Use Cases:**
- List Stream Decks
- Get device details

---

## aftrs_streamdeck_health

Check Stream Deck system health.

**Complexity:** simple

**Tags:** `streamdeck`, `health`, `status`

**Use Cases:**
- Check connection health
- Diagnose issues

---

## aftrs_streamdeck_refresh

Rescan for connected Stream Deck devices.

**Complexity:** simple

**Tags:** `streamdeck`, `refresh`, `scan`

**Use Cases:**
- Detect new devices
- Reconnect devices

---

## aftrs_streamdeck_reset

Reset device to default logo.

**Complexity:** simple

**Tags:** `streamdeck`, `reset`, `logo`

**Use Cases:**
- Reset to default
- Show Elgato logo

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `device` | number |  | Device index (default 0) |

### Example

```json
{
  "device": 0
}
```

---

## aftrs_streamdeck_set_color

Set a button to a solid color.

**Complexity:** simple

**Tags:** `streamdeck`, `color`, `button`

**Use Cases:**
- Set button color
- Indicate state with color

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `button` | number | Yes | Button index |
| `color` | string | Yes | Color as hex (#FF0000) or name (red, green, blue, etc.) |
| `device` | number |  | Device index (default 0) |

### Example

```json
{
  "button": 0,
  "color": "example",
  "device": 0
}
```

---

## aftrs_streamdeck_set_image

Set a button image from file path.

**Complexity:** simple

**Tags:** `streamdeck`, `image`, `button`

**Use Cases:**
- Set button icon
- Display image on button

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `button` | number | Yes | Button index |
| `device` | number |  | Device index (default 0) |
| `path` | string | Yes | Image file path (PNG recommended) |

### Example

```json
{
  "button": 0,
  "device": 0,
  "path": "example"
}
```

---

## aftrs_streamdeck_status

Get Stream Deck connection status and device info.

**Complexity:** simple

**Tags:** `streamdeck`, `status`, `devices`

**Use Cases:**
- Check connected devices
- Get device info

---

