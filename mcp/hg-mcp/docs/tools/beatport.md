# beatport

> Beatport API tools for track search, metadata enrichment, and chart browsing

**12 tools**

## Tools

- [`beatport_auth_status`](#beatport-auth-status)
- [`beatport_batch_enrich`](#beatport-batch-enrich)
- [`beatport_chart_tracks`](#beatport-chart-tracks)
- [`beatport_charts`](#beatport-charts)
- [`beatport_download`](#beatport-download)
- [`beatport_enrich`](#beatport-enrich)
- [`beatport_genres`](#beatport-genres)
- [`beatport_playlist`](#beatport-playlist)
- [`beatport_search`](#beatport-search)
- [`beatport_sync_likes`](#beatport-sync-likes)
- [`beatport_sync_playlist`](#beatport-sync-playlist)
- [`beatport_track`](#beatport-track)

---

## beatport_auth_status

Check Beatport authentication status and token validity

**Complexity:** simple

**Tags:** `beatport`, `auth`, `status`, `tokens`

**Use Cases:**
- Check auth status
- Verify API access

---

## beatport_batch_enrich

Batch enrich multiple CR8 tracks with Beatport metadata. Processes tracks missing BPM/key data.

**Complexity:** complex

**Tags:** `beatport`, `enrich`, `batch`, `cr8`, `metadata`

**Use Cases:**
- Bulk enrich library
- Fill missing metadata

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `dry_run` | boolean |  | Preview without saving to database (default: true) |
| `filter` | string |  | Filter tracks: 'missing_bpm', 'missing_key', 'missing_any' (default: missing_any) |
| `limit` | number |  | Maximum tracks to process (default: 10) |

### Example

```json
{
  "dry_run": false,
  "filter": "example",
  "limit": 0
}
```

---

## beatport_chart_tracks

Get tracks from a specific Beatport chart

**Complexity:** simple

**Tags:** `beatport`, `chart`, `tracks`, `top`

**Use Cases:**
- Get chart tracks
- View top tracks in genre

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `chart_id` | number | Yes | Beatport chart ID |

### Example

```json
{
  "chart_id": 0
}
```

---

## beatport_charts

List Beatport charts, optionally filtered by genre

**Complexity:** simple

**Tags:** `beatport`, `charts`, `trending`, `top`

**Use Cases:**
- Browse genre charts
- Find trending tracks

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `genre` | string |  | Genre slug to filter by (e.g., 'techno', 'house', 'drum-and-bass') |

### Example

```json
{
  "genre": "example"
}
```

---

## beatport_download

Download a single track from Beatport to local storage or S3

**Complexity:** moderate

**Tags:** `beatport`, `download`, `track`, `flac`, `aiff`

**Use Cases:**
- Download single track
- Add track to library

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `format` | string |  | Output format: 'flac', 'aiff', 'wav'. Default: aiff |
| `quality` | string |  | Download quality: 'lossless', 'medium', 'low'. Default: lossless |
| `track_id` | number | Yes | Beatport track ID |
| `upload_s3` | boolean |  | Upload to S3 after download (default: true) |

### Example

```json
{
  "format": "example",
  "quality": "example",
  "track_id": 0,
  "upload_s3": false
}
```

---

## beatport_enrich

Enrich a CR8 track with Beatport metadata. Searches by artist/title and updates DynamoDB with BPM, key, genre, label, etc.

**Complexity:** moderate

**Tags:** `beatport`, `enrich`, `metadata`, `cr8`, `dynamodb`

**Use Cases:**
- Add BPM/key to tracks
- Enrich library metadata

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `dry_run` | boolean |  | Preview without saving to database (default: true) |
| `track_id` | string | Yes | CR8 track ID to enrich |

### Example

```json
{
  "dry_run": false,
  "track_id": "example"
}
```

---

## beatport_genres

List all available Beatport genres

**Complexity:** simple

**Tags:** `beatport`, `genres`, `categories`

**Use Cases:**
- Browse genres
- Get genre list for filtering

---

## beatport_playlist

Get information about a Beatport playlist

**Complexity:** simple

**Tags:** `beatport`, `playlist`, `info`

**Use Cases:**
- View playlist details
- Preview before sync

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `include_tracks` | boolean |  | Include track listing (default: false) |
| `playlist_id` | number | Yes | Beatport playlist ID |

### Example

```json
{
  "include_tracks": false,
  "playlist_id": 0
}
```

---

## beatport_search

Search Beatport for tracks by artist and/or title. Returns track metadata including BPM, key, genre, label.

**Complexity:** simple

**Tags:** `beatport`, `search`, `tracks`, `metadata`, `music`

**Use Cases:**
- Search for track metadata
- Find BPM and key
- Lookup by artist/title

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `limit` | number |  | Maximum results to return (default: 10, max: 50) |
| `query` | string | Yes | Search query (artist name, track title, or both) |

### Example

```json
{
  "limit": 0,
  "query": "example"
}
```

---

## beatport_sync_likes

Like all tracks and follow all artists in a Beatport playlist using browser automation. Runs in background with resume support.

**Complexity:** complex

**Tags:** `beatport`, `likes`, `follows`, `playlist`, `automation`

**Use Cases:**
- Like all playlist tracks
- Follow all artists
- Build Beatport recommendations

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `mode` | string |  | Sync mode: 'both' (default), 'tracks' (likes only), 'artists' (follows only), 'status' (check progress) |
| `playlist_id` | string | Yes | Beatport playlist ID |

### Example

```json
{
  "mode": "example",
  "playlist_id": "example"
}
```

---

## beatport_sync_playlist

Sync a Beatport playlist to CR8 library. Downloads tracks as FLAC, converts to AIFF (Rekordbox-compatible), uploads to S3, and records in DynamoDB.

**Complexity:** complex

**Tags:** `beatport`, `sync`, `playlist`, `download`, `cr8`, `rekordbox`

**Use Cases:**
- Sync Beatport playlist to library
- Download tracks for DJing

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `dry_run` | boolean |  | Preview sync without downloading (default: true) |
| `format` | string |  | Output format: 'flac', 'aiff', 'wav'. Default: aiff (best for Rekordbox) |
| `limit` | number |  | Maximum tracks to sync (default: all) |
| `playlist_id` | number | Yes | Beatport playlist ID to sync |
| `quality` | string |  | Download quality: 'lossless' (FLAC), 'medium' (256 AAC), 'low' (128 AAC). Default: lossless |

### Example

```json
{
  "dry_run": false,
  "format": "example",
  "limit": 0,
  "playlist_id": 0,
  "quality": "example"
}
```

---

## beatport_track

Get detailed metadata for a specific Beatport track by ID

**Complexity:** simple

**Tags:** `beatport`, `track`, `metadata`, `details`

**Use Cases:**
- Get full track details
- Lookup by Beatport ID

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `track_id` | number | Yes | Beatport track ID |

### Example

```json
{
  "track_id": 0
}
```

---

