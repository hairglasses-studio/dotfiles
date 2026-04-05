# gdrive

> Google Drive tools for managing VJ clips and media files

**14 tools**

## Tools

- [`aftrs_gdrive_download`](#aftrs-gdrive-download)
- [`aftrs_gdrive_download_folder`](#aftrs-gdrive-download-folder)
- [`aftrs_gdrive_download_videos`](#aftrs-gdrive-download-videos)
- [`aftrs_gdrive_info`](#aftrs-gdrive-info)
- [`aftrs_gdrive_list`](#aftrs-gdrive-list)
- [`aftrs_gdrive_quota`](#aftrs-gdrive-quota)
- [`aftrs_gdrive_rclone_sync`](#aftrs-gdrive-rclone-sync)
- [`aftrs_gdrive_search`](#aftrs-gdrive-search)
- [`aftrs_gdrive_shared_drives`](#aftrs-gdrive-shared-drives)
- [`aftrs_gdrive_vj_packs`](#aftrs-gdrive-vj-packs)
- [`aftrs_gdrive_vj_sync`](#aftrs-gdrive-vj-sync)
- [`aftrs_gdrive_vj_zip_download`](#aftrs-gdrive-vj-zip-download)
- [`aftrs_gdrive_vj_zip_packs`](#aftrs-gdrive-vj-zip-packs)
- [`aftrs_gdrive_vj_zip_search`](#aftrs-gdrive-vj-zip-search)

---

## aftrs_gdrive_download

Download a single file from Google Drive to local storage.

**Complexity:** moderate

**Tags:** `gdrive`, `download`, `file`

**Use Cases:**
- Download VJ clip
- Get single media file

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `destination` | string |  | Local destination path (directory or full path). Defaults to current directory. |
| `file_id` | string | Yes | File ID to download |

### Example

```json
{
  "destination": "example",
  "file_id": "example"
}
```

---

## aftrs_gdrive_download_folder

Download all files from a Google Drive folder. Can filter by type (video/image) for VJ clips.

**Complexity:** complex

**Tags:** `gdrive`, `download`, `folder`, `batch`, `vj`, `clips`

**Use Cases:**
- Download VJ clip folder for show
- Batch download media

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `destination` | string | Yes | Local destination directory |
| `file_type` | string |  | Filter: 'video' for VJ clips, 'image' for stills, or empty for all files |
| `folder_id` | string | Yes | Folder ID to download |
| `recursive` | boolean |  | Include subfolders (default: false) |

### Example

```json
{
  "destination": "example",
  "file_type": "example",
  "folder_id": "example",
  "recursive": false
}
```

---

## aftrs_gdrive_download_videos

Quick download of all video files from a folder - optimized for VJ clip preparation.

**Complexity:** complex

**Tags:** `gdrive`, `download`, `video`, `vj`, `clips`, `batch`

**Use Cases:**
- Prepare VJ clips for tonight's show
- Batch download video files

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `destination` | string | Yes | Local destination directory (e.g., ~/VJ/Sets/Tonight) |
| `folder_id` | string | Yes | Folder ID containing VJ clips |
| `recursive` | boolean |  | Include videos from subfolders |

### Example

```json
{
  "destination": "example",
  "folder_id": "example",
  "recursive": false
}
```

---

## aftrs_gdrive_info

Get detailed information about a file or folder including size, dates, and path.

**Complexity:** simple

**Tags:** `gdrive`, `info`, `metadata`

**Use Cases:**
- Check file details
- Verify file before download

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `file_id` | string | Yes | File or folder ID |
| `include_path` | boolean |  | Include full folder path (slightly slower) |

### Example

```json
{
  "file_id": "example",
  "include_path": false
}
```

---

## aftrs_gdrive_list

List files and folders in a Google Drive directory. Returns file names, sizes, and IDs for further operations.

**Complexity:** simple

**Tags:** `gdrive`, `list`, `files`, `folders`

**Use Cases:**
- Browse VJ clip folders
- Find media files

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `folder_id` | string |  | Folder ID to list (use 'root' for My Drive root, or paste folder ID from URL) |
| `limit` | number |  | Maximum number of files to return (default: 100) |
| `page_token` | string |  | Page token for pagination (from previous response) |

### Example

```json
{
  "folder_id": "example",
  "limit": 0,
  "page_token": "example"
}
```

---

## aftrs_gdrive_quota

Check Google Drive storage quota and available space.

**Complexity:** simple

**Tags:** `gdrive`, `quota`, `storage`, `space`

**Use Cases:**
- Check available storage
- Monitor drive usage

---

## aftrs_gdrive_rclone_sync

Fast sync of VJ packs using rclone. Much faster than API for large files. Supports syncing specific packs or all packs.

**Complexity:** complex

**Tags:** `gdrive`, `vj`, `sync`, `rclone`, `fast`, `packs`

**Use Cases:**
- Fast sync VJ packs for show
- Update clip library with rclone

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `destination` | string |  | Destination folder (defaults to ~/Documents/Resolume Arena/Media) |
| `dry_run` | boolean |  | Preview sync without transferring files |
| `pack` | string | Yes | Pack name(s) to sync: 'hackerglasses', 'hairglasses,fetz,masks', or 'all' |

### Example

```json
{
  "destination": "example",
  "dry_run": false,
  "pack": "example"
}
```

---

## aftrs_gdrive_search

Search for files in Google Drive by name, type, or folder. Great for finding VJ clips by name.

**Complexity:** simple

**Tags:** `gdrive`, `search`, `find`, `video`, `clips`

**Use Cases:**
- Find VJ clips by name
- Search for video files

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `file_type` | string |  | Filter by type: 'video', 'image', 'audio', 'folder', or leave empty for all |
| `folder_id` | string |  | Limit search to specific folder ID |
| `limit` | number |  | Maximum results (default: 50) |
| `query` | string |  | Search query (file name contains this text) |

### Example

```json
{
  "file_type": "example",
  "folder_id": "example",
  "limit": 0,
  "query": "example"
}
```

---

## aftrs_gdrive_shared_drives

List available shared drives (team drives) that you have access to.

**Complexity:** simple

**Tags:** `gdrive`, `shared`, `team`, `drives`

**Use Cases:**
- Find shared VJ clip libraries
- Access team drives

---

## aftrs_gdrive_vj_packs

List available VJ clip packs that can be synced with rclone. Shows pack names and folder mappings.

**Complexity:** simple

**Tags:** `gdrive`, `vj`, `packs`, `list`, `catalog`

**Use Cases:**
- Browse available VJ packs
- Choose packs for tonight's show

---

## aftrs_gdrive_vj_sync

Sync VJ clips from a Google Drive folder to local Resolume media folder. Only downloads new/changed files.

**Complexity:** complex

**Tags:** `gdrive`, `vj`, `sync`, `resolume`, `clips`

**Use Cases:**
- Sync VJ library before show
- Update Resolume clips

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `destination` | string |  | Local Resolume media folder (defaults to ~/Documents/Resolume Arena/Media) |
| `recursive` | boolean |  | Include subfolders |
| `source_folder_id` | string | Yes | Google Drive folder ID containing VJ clips |

### Example

```json
{
  "destination": "example",
  "recursive": false,
  "source_folder_id": "example"
}
```

---

## aftrs_gdrive_vj_zip_download

Download and extract a VJ clip zip pack. Uses rclone for fast downloads and unzip for extraction.

**Complexity:** complex

**Tags:** `gdrive`, `vj`, `zip`, `download`, `extract`

**Use Cases:**
- Download VJ clip pack
- Install new visuals for show

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `destination` | string |  | Destination folder (defaults to ~/Documents/Resolume Arena/Media) |
| `dry_run` | boolean |  | Preview download without transferring files |
| `extract` | boolean |  | Extract zip after download (default: true) |
| `key` | string | Yes | Pack key from catalog (e.g., 'supreme_cyphers_v3', 'beeple_brainfeeder') |

### Example

```json
{
  "destination": "example",
  "dry_run": false,
  "extract": false,
  "key": "example"
}
```

---

## aftrs_gdrive_vj_zip_packs

List available VJ clip zip pack archives. Shows downloadable zip files containing VJ loops and visuals.

**Complexity:** simple

**Tags:** `gdrive`, `vj`, `zip`, `packs`, `catalog`

**Use Cases:**
- Browse VJ zip archives
- Find downloadable clip packs

---

## aftrs_gdrive_vj_zip_search

Search VJ clip zip packs by name, description, or tags.

**Complexity:** simple

**Tags:** `gdrive`, `vj`, `zip`, `search`

**Use Cases:**
- Find specific VJ packs
- Search by style or artist

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `query` | string | Yes | Search query (e.g., 'beeple', 'loops', '3d') |

### Example

```json
{
  "query": "example"
}
```

---

