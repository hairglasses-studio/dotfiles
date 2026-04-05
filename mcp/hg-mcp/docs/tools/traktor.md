# traktor

> Traktor Pro 3 library access for tracks, playlists, cue points, and export

**10 tools**

## Tools

- [`aftrs_traktor_cues`](#aftrs-traktor-cues)
- [`aftrs_traktor_export`](#aftrs-traktor-export)
- [`aftrs_traktor_health`](#aftrs-traktor-health)
- [`aftrs_traktor_history`](#aftrs-traktor-history)
- [`aftrs_traktor_library`](#aftrs-traktor-library)
- [`aftrs_traktor_loops`](#aftrs-traktor-loops)
- [`aftrs_traktor_playlist`](#aftrs-traktor-playlist)
- [`aftrs_traktor_playlists`](#aftrs-traktor-playlists)
- [`aftrs_traktor_status`](#aftrs-traktor-status)
- [`aftrs_traktor_track`](#aftrs-traktor-track)

---

## aftrs_traktor_cues

Get cue points for a specific track

**Complexity:** simple

**Tags:** `traktor`, `cues`, `hotcues`, `markers`

**Use Cases:**
- View cue points
- Check hotcues
- Get cue positions

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `file_path` | string | Yes | File path of the track |

### Example

```json
{
  "file_path": "example"
}
```

---

## aftrs_traktor_export

Export Traktor collection to Rekordbox XML format for migration

**Complexity:** moderate

**Tags:** `traktor`, `export`, `rekordbox`, `migration`

**Use Cases:**
- Export to Rekordbox
- Migrate library
- Backup collection

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `output_path` | string | Yes | Output path for the Rekordbox XML file |

### Example

```json
{
  "output_path": "example"
}
```

---

## aftrs_traktor_health

Check Traktor library health and get troubleshooting recommendations

**Complexity:** simple

**Tags:** `traktor`, `health`, `diagnostics`

**Use Cases:**
- Check library health
- Diagnose issues
- Troubleshoot errors

---

## aftrs_traktor_history

Get recently played tracks from Traktor history

**Complexity:** simple

**Tags:** `traktor`, `history`, `recent`, `played`

**Use Cases:**
- View play history
- Find recently played
- Track usage stats

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `limit` | integer |  | Maximum tracks to return (default: 20) |

### Example

```json
{
  "limit": 0
}
```

---

## aftrs_traktor_library

Search tracks in Traktor library by title, artist, album, or genre

**Complexity:** simple

**Tags:** `traktor`, `search`, `tracks`, `library`

**Use Cases:**
- Find tracks by name
- Search artist discography
- Browse by genre

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `limit` | integer |  | Maximum results to return (default: 20) |
| `query` | string | Yes | Search query for title, artist, album, or genre |

### Example

```json
{
  "limit": 0,
  "query": "example"
}
```

---

## aftrs_traktor_loops

Get saved loops for a specific track

**Complexity:** simple

**Tags:** `traktor`, `loops`, `sections`

**Use Cases:**
- View saved loops
- Get loop positions
- Check loop lengths

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `file_path` | string | Yes | File path of the track |

### Example

```json
{
  "file_path": "example"
}
```

---

## aftrs_traktor_playlist

Get contents of a specific playlist with all tracks

**Complexity:** simple

**Tags:** `traktor`, `playlist`, `tracks`

**Use Cases:**
- View playlist tracks
- Export playlist
- Get playlist info

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `name` | string | Yes | Playlist name (use folder/playlist for nested playlists) |

### Example

```json
{
  "name": "example"
}
```

---

## aftrs_traktor_playlists

List all playlists and folders in Traktor library

**Complexity:** simple

**Tags:** `traktor`, `playlists`, `folders`

**Use Cases:**
- Browse playlists
- View playlist hierarchy
- Find playlist by name

---

## aftrs_traktor_status

Get Traktor library status including track count and collection path

**Complexity:** simple

**Tags:** `traktor`, `dj`, `library`, `status`

**Use Cases:**
- Check library status
- Verify collection path
- Get track count

---

## aftrs_traktor_track

Get detailed track information including cue points, loops, and metadata

**Complexity:** simple

**Tags:** `traktor`, `track`, `metadata`, `cues`

**Use Cases:**
- Get track details
- View cue points
- Check BPM/key

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `file_path` | string | Yes | File path of the track |

### Example

```json
{
  "file_path": "example"
}
```

---

