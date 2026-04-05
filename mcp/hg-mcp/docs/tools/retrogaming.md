# retrogaming

> PS2 emulation and retro gaming support

**5 tools**

## Tools

- [`aftrs_capture_status`](#aftrs-capture-status)
- [`aftrs_ps2_games`](#aftrs-ps2-games)
- [`aftrs_ps2_savestate`](#aftrs-ps2-savestate)
- [`aftrs_ps2_status`](#aftrs-ps2-status)
- [`aftrs_retro_visualizer`](#aftrs-retro-visualizer)

---

## aftrs_capture_status

Get video capture device status for game capture.

**Complexity:** simple

**Tags:** `capture`, `video`, `recording`, `streaming`

**Use Cases:**
- Check capture card status
- Verify recording setup

---

## aftrs_ps2_games

List available PS2 games in the configured games directory.

**Complexity:** simple

**Tags:** `ps2`, `games`, `library`, `list`

**Use Cases:**
- Browse game library
- Find specific games

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `search` | string |  | Filter games by name |

### Example

```json
{
  "search": "example"
}
```

---

## aftrs_ps2_savestate

List save states for PS2 games.

**Complexity:** simple

**Tags:** `ps2`, `savestate`, `save`, `load`

**Use Cases:**
- Find save states
- Manage game saves

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `game` | string |  | Filter by game name (optional) |

### Example

```json
{
  "game": "example"
}
```

---

## aftrs_ps2_status

Get PCSX2 emulator status including running state and current game.

**Complexity:** simple

**Tags:** `ps2`, `pcsx2`, `emulator`, `status`

**Use Cases:**
- Check if emulator is running
- Get current game info

---

## aftrs_retro_visualizer

Control audio visualizer for retro gaming streams (placeholder).

**Complexity:** moderate

**Tags:** `visualizer`, `audio`, `effects`, `streaming`

**Use Cases:**
- Control stream visuals
- Audio reactive effects

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `action` | string |  | Action: 'start', 'stop', 'status' |

### Example

```json
{
  "action": "example"
}
```

---

