# vj_clips

> VJ video clip management: scan, sync, upload, and download clips via S3 with user/playlist organization

**11 tools**

## Tools

- [`aftrs_vj_clips_delete`](#aftrs-vj-clips-delete)
- [`aftrs_vj_clips_download`](#aftrs-vj-clips-download)
- [`aftrs_vj_clips_download_playlist`](#aftrs-vj-clips-download-playlist)
- [`aftrs_vj_clips_health`](#aftrs-vj-clips-health)
- [`aftrs_vj_clips_list`](#aftrs-vj-clips-list)
- [`aftrs_vj_clips_playlists`](#aftrs-vj-clips-playlists)
- [`aftrs_vj_clips_s3_list`](#aftrs-vj-clips-s3-list)
- [`aftrs_vj_clips_s3_playlists`](#aftrs-vj-clips-s3-playlists)
- [`aftrs_vj_clips_sync`](#aftrs-vj-clips-sync)
- [`aftrs_vj_clips_upload`](#aftrs-vj-clips-upload)
- [`aftrs_vj_clips_upload_playlist`](#aftrs-vj-clips-upload-playlist)

---

## aftrs_vj_clips_delete

Delete a clip from S3.

**Complexity:** simple

**Tags:** `vj`, `clips`, `delete`, `s3`

**Use Cases:**
- Remove cloud clip
- Clean up storage

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `key` | string | Yes | S3 key of the clip to delete |

### Example

```json
{
  "key": "example"
}
```

---

## aftrs_vj_clips_download

Download a clip from S3.

**Complexity:** moderate

**Tags:** `vj`, `clips`, `download`, `s3`

**Use Cases:**
- Download cloud clip
- Restore from backup

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `key` | string | Yes | S3 key of the clip |
| `playlist` | string |  | Local playlist to save to (default: from S3 path) |

### Example

```json
{
  "key": "example",
  "playlist": "example"
}
```

---

## aftrs_vj_clips_download_playlist

Download all clips in an S3 playlist.

**Complexity:** moderate

**Tags:** `vj`, `clips`, `download`, `playlist`, `batch`

**Use Cases:**
- Download entire playlist
- Get shared pack

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `playlist` | string | Yes | Playlist name to download |
| `user` | string | Yes | User who owns the playlist |

### Example

```json
{
  "playlist": "example",
  "user": "example"
}
```

---

## aftrs_vj_clips_health

Check VJ clips system health.

**Complexity:** simple

**Tags:** `vj`, `clips`, `health`, `status`

**Use Cases:**
- Diagnose issues
- System check

---

## aftrs_vj_clips_list

List local VJ clips organized by playlist.

**Complexity:** simple

**Tags:** `vj`, `clips`, `list`, `resolume`, `media`

**Use Cases:**
- Browse local clips
- Inventory media

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `playlist` | string |  | Filter by playlist name |

### Example

```json
{
  "playlist": "example"
}
```

---

## aftrs_vj_clips_playlists

List all local playlists with clip counts and sizes.

**Complexity:** simple

**Tags:** `vj`, `clips`, `playlists`, `packs`

**Use Cases:**
- View playlist overview
- Check media organization

---

## aftrs_vj_clips_s3_list

List VJ clips in S3 bucket.

**Complexity:** simple

**Tags:** `vj`, `clips`, `s3`, `cloud`, `list`

**Use Cases:**
- Browse cloud clips
- See available downloads

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `playlist` | string |  | Filter by playlist name |
| `user` | string |  | Filter by user (default: current user) |

### Example

```json
{
  "playlist": "example",
  "user": "example"
}
```

---

## aftrs_vj_clips_s3_playlists

List playlists in S3 bucket.

**Complexity:** simple

**Tags:** `vj`, `clips`, `s3`, `playlists`

**Use Cases:**
- View cloud playlists
- Browse shared content

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `user` | string |  | Filter by user (default: all users) |

### Example

```json
{
  "user": "example"
}
```

---

## aftrs_vj_clips_sync

Compare local and S3 clips to identify sync opportunities.

**Complexity:** moderate

**Tags:** `vj`, `clips`, `sync`, `compare`

**Use Cases:**
- Check sync status
- Find missing clips

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `user` | string |  | User to compare against (default: current user) |

### Example

```json
{
  "user": "example"
}
```

---

## aftrs_vj_clips_upload

Upload a clip to S3.

**Complexity:** moderate

**Tags:** `vj`, `clips`, `upload`, `s3`

**Use Cases:**
- Backup clip to cloud
- Share clip

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `path` | string | Yes | Local path to the clip file |
| `playlist` | string |  | Playlist/pack name (default: from path or 'default') |
| `tags` | string |  | Comma-separated tags |
| `user` | string |  | User/owner (default: current user) |

### Example

```json
{
  "path": "example",
  "playlist": "example",
  "tags": "example",
  "user": "example"
}
```

---

## aftrs_vj_clips_upload_playlist

Upload all clips in a playlist to S3.

**Complexity:** moderate

**Tags:** `vj`, `clips`, `upload`, `playlist`, `batch`

**Use Cases:**
- Backup entire playlist
- Share pack

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `playlist` | string | Yes | Playlist name to upload |
| `user` | string |  | User/owner (default: current user) |

### Example

```json
{
  "playlist": "example",
  "user": "example"
}
```

---

