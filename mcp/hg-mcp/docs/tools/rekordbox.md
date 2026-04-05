# rekordbox

> Rekordbox DJ library integration with OneLibrary cloud sync support

**12 tools**

## Tools

- [`aftrs_rekordbox_cloud_status`](#aftrs-rekordbox-cloud-status)
- [`aftrs_rekordbox_import_shared`](#aftrs-rekordbox-import-shared)
- [`aftrs_rekordbox_list_shared`](#aftrs-rekordbox-list-shared)
- [`aftrs_rekordbox_playlists`](#aftrs-rekordbox-playlists)
- [`aftrs_rekordbox_recent`](#aftrs-rekordbox-recent)
- [`aftrs_rekordbox_s3_sync`](#aftrs-rekordbox-s3-sync)
- [`aftrs_rekordbox_search`](#aftrs-rekordbox-search)
- [`aftrs_rekordbox_share_playlist`](#aftrs-rekordbox-share-playlist)
- [`aftrs_rekordbox_stats`](#aftrs-rekordbox-stats)
- [`aftrs_rekordbox_track_info`](#aftrs-rekordbox-track-info)
- [`aftrs_rekordbox_usb_status`](#aftrs-rekordbox-usb-status)
- [`aftrs_rekordbox_usb_validate`](#aftrs-rekordbox-usb-validate)

---

## aftrs_rekordbox_cloud_status

Check Google Drive cloud sync status for Rekordbox library.

**Complexity:** simple

**Tags:** `rekordbox`, `cloud`, `gdrive`, `sync`

**Use Cases:**
- Check sync status
- Monitor cloud library

---

## aftrs_rekordbox_import_shared

Import a shared playlist from another user.

**Complexity:** moderate

**Tags:** `rekordbox`, `import`, `shared`, `playlist`

**Use Cases:**
- Import collaborator playlists
- Sync shared sets

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `dry_run` | boolean |  | Preview import without making changes |
| `from_user` | string | Yes | Source user who shared the playlist |
| `playlist` | string | Yes | Shared playlist name to import |

### Example

```json
{
  "dry_run": false,
  "from_user": "example",
  "playlist": "example"
}
```

---

## aftrs_rekordbox_list_shared

List playlists shared with you from other users.

**Complexity:** simple

**Tags:** `rekordbox`, `shared`, `playlists`, `import`

**Use Cases:**
- See available shared playlists
- Check for updates

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `from_user` | string |  | Filter by source user (optional) |

### Example

```json
{
  "from_user": "example"
}
```

---

## aftrs_rekordbox_playlists

List all Rekordbox playlists and folders with track counts.

**Complexity:** simple

**Tags:** `rekordbox`, `playlists`, `folders`, `organization`

**Use Cases:**
- Browse playlists
- View folder structure

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `tree` | boolean |  | Show as hierarchical tree (default: true) |

### Example

```json
{
  "tree": false
}
```

---

## aftrs_rekordbox_recent

List recently added or modified tracks.

**Complexity:** simple

**Tags:** `rekordbox`, `recent`, `new`, `tracks`

**Use Cases:**
- See new additions
- Track recent imports

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `days` | number |  | Number of days to look back (default: 7) |
| `limit` | number |  | Max results (default: 50) |

### Example

```json
{
  "days": 0,
  "limit": 0
}
```

---

## aftrs_rekordbox_s3_sync

Sync Rekordbox library to S3 bucket for multi-machine access.

**Complexity:** moderate

**Tags:** `rekordbox`, `s3`, `sync`, `backup`

**Use Cases:**
- Backup library
- Multi-machine sync

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `direction` | string |  | Sync direction: up, down, both (default: up) |
| `dry_run` | boolean |  | Preview changes without syncing |

### Example

```json
{
  "direction": "example",
  "dry_run": false
}
```

---

## aftrs_rekordbox_search

Search tracks in Rekordbox library by title, artist, BPM, or key.

**Complexity:** moderate

**Tags:** `rekordbox`, `search`, `tracks`, `find`

**Use Cases:**
- Find tracks for sets
- Filter by BPM/key

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `bpm_max` | number |  | Maximum BPM |
| `bpm_min` | number |  | Minimum BPM |
| `key` | string |  | Musical key (e.g., 8A, 1B, Am, C) |
| `limit` | number |  | Max results (default: 20) |
| `query` | string |  | Search query (title, artist) |

### Example

```json
{
  "bpm_max": 0,
  "bpm_min": 0,
  "key": "example",
  "limit": 0,
  "query": "example"
}
```

---

## aftrs_rekordbox_share_playlist

Share a playlist with another user via S3. Exports playlist to XML and uploads.

**Complexity:** moderate

**Tags:** `rekordbox`, `share`, `playlist`, `multi-user`

**Use Cases:**
- Share playlists between DJs
- Collaborate on sets

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `include_tracks` | boolean |  | Include track files in share (default: false, metadata only) |
| `playlist` | string | Yes | Playlist name to share |
| `to_user` | string | Yes | Target user (e.g., luke-lasley, hairglasses) |

### Example

```json
{
  "include_tracks": false,
  "playlist": "example",
  "to_user": "example"
}
```

---

## aftrs_rekordbox_stats

Get Rekordbox library statistics: track count, playlists, cues, analysis data.

**Complexity:** simple

**Tags:** `rekordbox`, `stats`, `library`, `dj`

**Use Cases:**
- View library overview
- Check collection size

---

## aftrs_rekordbox_track_info

Get detailed track information including cue points, loops, and beat grid.

**Complexity:** moderate

**Tags:** `rekordbox`, `track`, `cues`, `beatgrid`, `details`

**Use Cases:**
- View cue points
- Check beat grid
- See track metadata

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `path` | string |  | Or path to audio file |
| `track_id` | number |  | Track ID from search results |

### Example

```json
{
  "path": "example",
  "track_id": 0
}
```

---

## aftrs_rekordbox_usb_status

Check status of connected USB drives with Rekordbox exports.

**Complexity:** simple

**Tags:** `rekordbox`, `usb`, `export`, `status`

**Use Cases:**
- Check USB exports
- Verify gig prep

---

## aftrs_rekordbox_usb_validate

Validate USB structure for XDJ-1000 MK2 compatibility.

**Complexity:** moderate

**Tags:** `rekordbox`, `usb`, `validate`, `xdj`

**Use Cases:**
- Verify USB compatibility
- Check for issues

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `path` | string |  | USB mount path (e.g., /Volumes/DJ_USB) |

### Example

```json
{
  "path": "example"
}
```

---

