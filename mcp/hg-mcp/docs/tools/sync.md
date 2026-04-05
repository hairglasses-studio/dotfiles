# sync

> Music sync tools for SoundCloud, Beatport, and Rekordbox

**12 tools**

## Tools

- [`aftrs_timecode_goto`](#aftrs-timecode-goto)
- [`aftrs_timecode_status`](#aftrs-timecode-status)
- [`aftrs_timecode_sync`](#aftrs-timecode-sync)
- [`sync_add_user`](#sync-add-user)
- [`sync_all`](#sync-all)
- [`sync_beatport`](#sync-beatport)
- [`sync_discover_playlists`](#sync-discover-playlists)
- [`sync_health`](#sync-health)
- [`sync_list_users`](#sync-list-users)
- [`sync_rekordbox`](#sync-rekordbox)
- [`sync_soundcloud`](#sync-soundcloud)
- [`sync_status`](#sync-status)

---

## aftrs_timecode_goto

Jump all linked systems to a specific timecode position.

**Complexity:** simple

**Tags:** `timecode`, `goto`, `position`, `jump`

**Use Cases:**
- Jump to show position
- Sync all systems to mark
- Rehearse from point

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `position` | string | Yes | Timecode position in HH:MM:SS:FF format (e.g., 01:30:45:15) |

### Example

```json
{
  "position": "example"
}
```

---

## aftrs_timecode_status

Get current timecode status across all systems. Shows master source, current position, linked systems, and sync status.

**Complexity:** simple

**Tags:** `timecode`, `sync`, `status`, `smpte`, `mtc`

**Use Cases:**
- Check timecode position
- Verify sync status
- Monitor linked systems

---

## aftrs_timecode_sync

Configure timecode synchronization. Set master source and link/unlink systems.

**Complexity:** moderate

**Tags:** `timecode`, `sync`, `link`, `master`

**Use Cases:**
- Set master timecode source
- Link systems for sync
- Configure frame rate

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `action` | string | Yes | Action: set_master, link, unlink, sync_now |
| `drop_frame` | boolean |  | Use drop frame timecode (for 29.97/59.94) |
| `format` | string |  | Timecode format: smpte, mtc, ltc (for set_master) |
| `frame_rate` | number |  | Frame rate: 24, 25, 29.97, 30, etc. |
| `system` | string |  | System name: showkontrol, grandma3, ableton |

### Example

```json
{
  "action": "example",
  "drop_frame": false,
  "format": "example",
  "frame_rate": 0,
  "system": "example"
}
```

---

## sync_add_user

Add a new SoundCloud user for conversion tracking. Downloads likes and all public playlists to S3.

**Complexity:** moderate

**Tags:** `sync`, `soundcloud`, `user`, `add`

**Use Cases:**
- Add new SoundCloud user
- Track new artist

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `display_name` | string |  | Display name for Rekordbox folders |
| `download` | boolean |  | Start downloading immediately (default: true) |
| `username` | string | Yes | SoundCloud username (from URL like soundcloud.com/username) |

### Example

```json
{
  "display_name": "example",
  "download": false,
  "username": "example"
}
```

---

## sync_all

Sync all music sources to Rekordbox. Downloads from SoundCloud and Beatport, then imports to Rekordbox playlists.

**Complexity:** complex

**Tags:** `sync`, `soundcloud`, `beatport`, `rekordbox`, `music`

**Use Cases:**
- Full sync of all music sources
- Sync specific user

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `dry_run` | boolean |  | Preview without making changes (default: true) |
| `user` | string |  | Sync specific user only (hairglasses, freaq-show, rahul, luke-lasley, aidan, fogel, marissughdevelops) |

### Example

```json
{
  "dry_run": false,
  "user": "example"
}
```

---

## sync_beatport

Sync Beatport playlists to local storage.

**Complexity:** moderate

**Tags:** `sync`, `beatport`, `music`, `download`

**Use Cases:**
- Sync Beatport purchases
- Download purchased tracks

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `dry_run` | boolean |  | Preview without making changes |
| `user` | string |  | Beatport username to sync |

### Example

```json
{
  "dry_run": false,
  "user": "example"
}
```

---

## sync_discover_playlists

Discover and list all public playlists for a SoundCloud user.

**Complexity:** simple

**Tags:** `sync`, `soundcloud`, `playlists`, `discover`

**Use Cases:**
- Find user playlists
- Discover new content

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `username` | string | Yes | SoundCloud username |

### Example

```json
{
  "username": "example"
}
```

---

## sync_health

Check health of all sync dependencies (AWS CLI, S3, DynamoDB, ffmpeg). Returns status for each service and circuit breaker states.

**Complexity:** simple

**Tags:** `sync`, `health`, `monitoring`, `aws`, `s3`, `circuit-breaker`

**Use Cases:**
- Check service health
- Diagnose sync issues
- Reset circuit breakers

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `reset_circuit_breaker` | string |  | Reset circuit breaker for a specific service (soundcloud, beatport, rekordbox) or 'all' to reset all |

### Example

```json
{
  "reset_circuit_breaker": "example"
}
```

---

## sync_list_users

List all tracked SoundCloud users and their sync status.

**Complexity:** simple

**Tags:** `sync`, `users`, `list`

**Use Cases:**
- View tracked users
- Check user sync status

---

## sync_rekordbox

Import pending music files to Rekordbox playlists.

**Complexity:** moderate

**Tags:** `sync`, `rekordbox`, `import`, `dj`

**Use Cases:**
- Import tracks to Rekordbox
- Update Rekordbox playlists

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `dry_run` | boolean |  | Preview without making changes |

### Example

```json
{
  "dry_run": false
}
```

---

## sync_soundcloud

Sync SoundCloud likes and playlists to local storage.

**Complexity:** moderate

**Tags:** `sync`, `soundcloud`, `music`, `playlists`

**Use Cases:**
- Sync SoundCloud likes
- Sync SoundCloud playlists

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `dry_run` | boolean |  | Preview without making changes |
| `user` | string |  | SoundCloud username to sync |

### Example

```json
{
  "dry_run": false,
  "user": "example"
}
```

---

## sync_status

Show current sync status for all services.

**Complexity:** simple

**Tags:** `sync`, `status`, `monitoring`

**Use Cases:**
- Check sync status
- View last sync times

---

