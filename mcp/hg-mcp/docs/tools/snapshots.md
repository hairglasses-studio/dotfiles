# snapshots

> Show state snapshots for capturing and recalling system states

**4 tools**

## Tools

- [`aftrs_snapshot_capture`](#aftrs-snapshot-capture)
- [`aftrs_snapshot_diff`](#aftrs-snapshot-diff)
- [`aftrs_snapshot_list`](#aftrs-snapshot-list)
- [`aftrs_snapshot_recall`](#aftrs-snapshot-recall)

---

## aftrs_snapshot_capture

Capture current state of all systems as a snapshot. Saves Ableton, Resolume, grandMA3, OBS, and Showkontrol states.

**Complexity:** moderate

**Tags:** `snapshot`, `capture`, `state`, `backup`

**Use Cases:**
- Save show state before changes
- Create restore point
- Backup current settings

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `description` | string |  | Optional description |
| `name` | string | Yes | Name for this snapshot |
| `systems` | string |  | Comma-separated systems to capture (default: all). Options: ableton, resolume, grandma3, obs, showkontrol |
| `tags` | string |  | Comma-separated tags for organization |

### Example

```json
{
  "description": "example",
  "name": "example",
  "systems": "example",
  "tags": "example"
}
```

---

## aftrs_snapshot_diff

Compare two snapshots and show differences.

**Complexity:** simple

**Tags:** `snapshot`, `diff`, `compare`

**Use Cases:**
- Compare show states
- Track changes
- Audit modifications

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `snapshot1` | string | Yes | First snapshot ID |
| `snapshot2` | string | Yes | Second snapshot ID |

### Example

```json
{
  "snapshot1": "example",
  "snapshot2": "example"
}
```

---

## aftrs_snapshot_list

List all available snapshots with details.

**Complexity:** simple

**Tags:** `snapshot`, `list`, `view`

**Use Cases:**
- View available snapshots
- Find snapshot to restore

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `snapshot_id` | string |  | Get details of a specific snapshot |

### Example

```json
{
  "snapshot_id": "example"
}
```

---

## aftrs_snapshot_recall

Restore system states from a previously captured snapshot.

**Complexity:** moderate

**Tags:** `snapshot`, `recall`, `restore`, `state`

**Use Cases:**
- Restore show state
- Rollback changes
- Quick scene reset

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `snapshot_id` | string | Yes | Snapshot ID to recall |
| `systems` | string |  | Comma-separated systems to restore (default: all in snapshot) |

### Example

```json
{
  "snapshot_id": "example",
  "systems": "example"
}
```

---

