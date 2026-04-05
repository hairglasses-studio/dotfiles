# setlist

> DJ setlist planning and management

**7 tools**

## Tools

- [`aftrs_setlist_add_track`](#aftrs-setlist-add-track)
- [`aftrs_setlist_analyze`](#aftrs-setlist-analyze)
- [`aftrs_setlist_create`](#aftrs-setlist-create)
- [`aftrs_setlist_export`](#aftrs-setlist-export)
- [`aftrs_setlist_remove_track`](#aftrs-setlist-remove-track)
- [`aftrs_setlist_reorder`](#aftrs-setlist-reorder)
- [`aftrs_setlist_view`](#aftrs-setlist-view)

---

## aftrs_setlist_add_track

Add a track to a setlist.

**Complexity:** simple

**Tags:** `setlist`, `track`, `add`

**Use Cases:**
- Add track to set plan
- Build setlist

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `artist` | string | Yes | Track artist |
| `bpm` | number |  | Track BPM |
| `energy` | number |  | Energy level 1-10 |
| `key` | string |  | Track key (e.g., '5A', 'Dm') |
| `notes` | string |  | Transition notes |
| `setlist_id` | string | Yes | Setlist ID |
| `title` | string | Yes | Track title |

### Example

```json
{
  "artist": "example",
  "bpm": 0,
  "energy": 0,
  "key": "example",
  "notes": "example",
  "setlist_id": "example",
  "title": "example"
}
```

---

## aftrs_setlist_analyze

Analyze a setlist for BPM flow, key compatibility, and energy progression.

**Complexity:** simple

**Tags:** `setlist`, `analyze`, `flow`, `bpm`, `key`

**Use Cases:**
- Check set flow
- Find transition issues

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `setlist_id` | string | Yes | Setlist ID |

### Example

```json
{
  "setlist_id": "example"
}
```

---

## aftrs_setlist_create

Create a new setlist for planning a DJ set or performance.

**Complexity:** simple

**Tags:** `setlist`, `create`, `plan`, `dj`

**Use Cases:**
- Create new DJ set plan
- Plan performance tracklist

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `date` | string |  | Performance date (YYYY-MM-DD) |
| `description` | string |  | Setlist description |
| `name` | string | Yes | Setlist name |
| `venue` | string |  | Venue name |

### Example

```json
{
  "date": "example",
  "description": "example",
  "name": "example",
  "venue": "example"
}
```

---

## aftrs_setlist_export

Export a setlist to various formats.

**Complexity:** simple

**Tags:** `setlist`, `export`, `m3u`, `playlist`

**Use Cases:**
- Export to playlist
- Share setlist

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `format` | string |  | Export format: text, m3u, json (default: text) |
| `setlist_id` | string | Yes | Setlist ID |

### Example

```json
{
  "format": "example",
  "setlist_id": "example"
}
```

---

## aftrs_setlist_remove_track

Remove a track from a setlist by position.

**Complexity:** simple

**Tags:** `setlist`, `track`, `remove`, `delete`

**Use Cases:**
- Remove track from set
- Edit setlist

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `position` | number | Yes | Track position to remove (1-indexed) |
| `setlist_id` | string | Yes | Setlist ID |

### Example

```json
{
  "position": 0,
  "setlist_id": "example"
}
```

---

## aftrs_setlist_reorder

Move a track to a new position in the setlist.

**Complexity:** simple

**Tags:** `setlist`, `reorder`, `move`

**Use Cases:**
- Reorder set tracks
- Adjust flow

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `from_position` | number | Yes | Current track position |
| `setlist_id` | string | Yes | Setlist ID |
| `to_position` | number | Yes | New track position |

### Example

```json
{
  "from_position": 0,
  "setlist_id": "example",
  "to_position": 0
}
```

---

## aftrs_setlist_view

View a setlist with all tracks and details.

**Complexity:** simple

**Tags:** `setlist`, `view`, `details`

**Use Cases:**
- View setlist tracks
- Check set plan

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `setlist_id` | string | Yes | Setlist ID |

### Example

```json
{
  "setlist_id": "example"
}
```

---

