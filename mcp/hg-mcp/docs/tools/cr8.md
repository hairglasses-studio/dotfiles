# cr8

> CR8 music library management tools for DynamoDB/S3 sync, analysis, and maintenance

**23 tools**

## Tools

- [`cr8_analysis_stats`](#cr8-analysis-stats)
- [`cr8_dynamodb_import`](#cr8-dynamodb-import)
- [`cr8_import_s3_tracks`](#cr8-import-s3-tracks)
- [`cr8_migrate_paths`](#cr8-migrate-paths)
- [`cr8_migration_status`](#cr8-migration-status)
- [`cr8_playlist_list`](#cr8-playlist-list)
- [`cr8_queue_analysis`](#cr8-queue-analysis)
- [`cr8_queue_status`](#cr8-queue-status)
- [`cr8_rekordbox_status`](#cr8-rekordbox-status)
- [`cr8_run_analysis`](#cr8-run-analysis)
- [`cr8_s3_reorganize`](#cr8-s3-reorganize)
- [`cr8_s3_structure`](#cr8-s3-structure)
- [`cr8_s3_to_local`](#cr8-s3-to-local)
- [`cr8_status`](#cr8-status)
- [`cr8_supabase_export`](#cr8-supabase-export)
- [`cr8_sync_likes`](#cr8-sync-likes)
- [`cr8_sync_music`](#cr8-sync-music)
- [`cr8_sync_rekordbox`](#cr8-sync-rekordbox)
- [`cr8_sync_status`](#cr8-sync-status)
- [`cr8_track_search`](#cr8-track-search)
- [`cr8_update_paths`](#cr8-update-paths)
- [`cr8_verify_migration`](#cr8-verify-migration)
- [`cr8_verify_sync`](#cr8-verify-sync)

---

## cr8_analysis_stats

Get audio analysis coverage statistics (BPM, key detection rates)

**Complexity:** simple

**Tags:** `cr8`, `analysis`, `bpm`, `key`, `coverage`

**Use Cases:**
- Check analysis coverage
- Find unanalyzed tracks
- Monitor BPM/key detection

---

## cr8_dynamodb_import

Import JSON data into DynamoDB tables (tracks, playlists, etc.)

**Complexity:** complex

**Tags:** `cr8`, `dynamodb`, `import`, `json`, `migrate`

**Use Cases:**
- Import track data
- Restore from backup
- Migrate from JSON

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `dry_run` | boolean |  | Preview import without writing (default: true) |
| `json_file` | string |  | Path to JSON file (default: exports/{table}.json) |
| `table` | string | Yes | Target table: 'tracks', 'playlists', 'playlist_tracks' |

### Example

```json
{
  "dry_run": false,
  "json_file": "example",
  "table": "example"
}
```

---

## cr8_import_s3_tracks

Import orphaned S3 audio files into DynamoDB. Creates track records for files not in database.

**Complexity:** complex

**Tags:** `cr8`, `import`, `s3`, `tracks`, `orphaned`

**Use Cases:**
- Import S3 files to DB
- Create track records
- Sync S3 with DynamoDB

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `dry_run` | boolean |  | Preview without importing (default: true) |
| `limit` | number |  | Maximum tracks to import (default: all) |

### Example

```json
{
  "dry_run": false,
  "limit": 0
}
```

---

## cr8_migrate_paths

Migrate tracks from flat S3 paths (downloads/user/file.m4a) to full DJ Crates structure (downloads/dj_crates/user/playlist/file.m4a). Moves S3 objects and updates DynamoDB records.

**Complexity:** complex

**Tags:** `cr8`, `migrate`, `paths`, `s3`, `reorganize`

**Use Cases:**
- Fix track paths
- Migrate to new structure
- Reorganize S3 files

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `dry_run` | boolean |  | Preview moves without executing (default: true) |
| `limit` | number |  | Maximum tracks to migrate (default: all) |
| `verbose` | boolean |  | Show detailed output (default: false) |

### Example

```json
{
  "dry_run": false,
  "limit": 0,
  "verbose": false
}
```

---

## cr8_migration_status

Get comprehensive migration status including DynamoDB tables, S3 bucket structure, and rclone jobs

**Complexity:** simple

**Tags:** `cr8`, `migration`, `status`, `dynamodb`, `s3`

**Use Cases:**
- Check migration progress
- Monitor DynamoDB tables
- View S3 bucket structure

---

## cr8_playlist_list

List CR8 playlists with minimal output. Returns ID, service, track count, and name.

**Complexity:** simple

**Tags:** `cr8`, `playlist`, `list`, `lightweight`

**Use Cases:**
- List playlists quickly
- Get playlist IDs
- Low-token playlist lookup

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `all` | boolean |  | Include disabled playlists (default: false) |
| `json` | boolean |  | Output as compact JSON (default: false) |
| `service` | string |  | Filter by service: soundcloud, youtube (optional) |

### Example

```json
{
  "all": false,
  "json": false,
  "service": "example"
}
```

---

## cr8_queue_analysis

Queue unanalyzed tracks for BPM/key analysis

**Complexity:** moderate

**Tags:** `cr8`, `analysis`, `queue`, `bpm`, `key`

**Use Cases:**
- Queue tracks for analysis
- Trigger BPM detection
- Process unanalyzed music

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `dry_run` | boolean |  | Preview without queuing (default: false) |
| `limit` | number |  | Maximum tracks to queue (default: all) |

### Example

```json
{
  "dry_run": false,
  "limit": 0
}
```

---

## cr8_queue_status

Get detailed analysis queue status: pending/processing/completed counts, recent failures, and ETA

**Complexity:** simple

**Tags:** `cr8`, `queue`, `status`, `analysis`, `progress`

**Use Cases:**
- Check queue progress
- View processing rate
- Debug failures

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `failures` | boolean |  | Show recent failure details (default: false) |

### Example

```json
{
  "failures": false
}
```

---

## cr8_rekordbox_status

Get Rekordbox library status, local file counts, and sync progress between CR8 playlists and Rekordbox

**Complexity:** simple

**Tags:** `cr8`, `rekordbox`, `status`, `sync`, `dj`

**Use Cases:**
- Check Rekordbox sync status
- View library stats
- See which playlists need syncing

---

## cr8_run_analysis

Run analysis worker to process queued tracks. Downloads from S3, analyzes with librosa, updates DynamoDB with BPM/key.

**Complexity:** complex

**Tags:** `cr8`, `analysis`, `worker`, `bpm`, `key`, `librosa`

**Use Cases:**
- Process analysis queue
- Analyze tracks for BPM/key
- Run batch analysis

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `batch` | number |  | Number of tracks to process (default: 10) |
| `dry_run` | boolean |  | Preview without updating database (default: false) |

### Example

```json
{
  "batch": 0,
  "dry_run": false
}
```

---

## cr8_s3_reorganize

Reorganize S3 bucket structure - move files to new path hierarchy

**Complexity:** complex

**Tags:** `cr8`, `s3`, `reorganize`, `structure`, `migrate`

**Use Cases:**
- Reorganize S3 paths
- Migrate to new structure
- Clean up bucket

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `dry_run` | boolean |  | Preview moves without executing (default: true) |

### Example

```json
{
  "dry_run": false
}
```

---

## cr8_s3_structure

Get S3 bucket structure with file counts and sizes per prefix

**Complexity:** simple

**Tags:** `cr8`, `s3`, `structure`, `storage`, `files`

**Use Cases:**
- View S3 organization
- Check storage usage
- Audit bucket structure

---

## cr8_s3_to_local

Sync audio files from S3 to local ~/Music/CR8 folder using rclone. Downloads new and updated tracks.

**Complexity:** moderate

**Tags:** `cr8`, `s3`, `local`, `sync`, `rclone`, `download`

**Use Cases:**
- Download music from S3
- Sync cloud to local
- Update local library

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `dry_run` | boolean |  | Preview sync without downloading (default: true) |
| `prefix` | string |  | S3 prefix to sync: 'downloads', 'users', or specific path (default: downloads) |

### Example

```json
{
  "dry_run": false,
  "prefix": "example"
}
```

---

## cr8_status

Get complete CR8 system status: tracks, analysis coverage, queue depth, S3 storage, and ETA for analysis completion

**Complexity:** moderate

**Tags:** `cr8`, `status`, `consolidated`, `overview`, `dashboard`

**Use Cases:**
- Get system overview
- Check analysis progress
- Monitor queue status

---

## cr8_supabase_export

Export data from Supabase to JSON files for migration

**Complexity:** moderate

**Tags:** `cr8`, `supabase`, `export`, `json`, `backup`

**Use Cases:**
- Export from Supabase
- Create migration data
- Backup database

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `table` | string |  | Table to export: 'tracks', 'playlists', 'all' (default: all) |

### Example

```json
{
  "table": "example"
}
```

---

## cr8_sync_likes

Sync new tracks from all playlists (SoundCloud likes, YouTube, etc.) to S3. Fetches new tracks, downloads them, and queues for analysis.

**Complexity:** complex

**Tags:** `cr8`, `sync`, `likes`, `playlist`, `soundcloud`, `youtube`, `s3`

**Use Cases:**
- Sync new playlist likes
- Download new tracks to S3
- Keep playlists updated

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `dry_run` | boolean |  | Preview without downloading (default: false) |
| `limit` | number |  | Max new tracks to process per playlist (default: all) |
| `playlist_id` | number |  | Sync specific playlist by ID (optional) |
| `service` | string |  | Filter by service: soundcloud, youtube (optional) |
| `username` | string |  | Sync playlists for specific username (optional) |

### Example

```json
{
  "dry_run": false,
  "limit": 0,
  "playlist_id": 0,
  "service": "example",
  "username": "example"
}
```

---

## cr8_sync_music

Sync local music folder to S3 using rclone (filters DJ sets >25MB)

**Complexity:** complex

**Tags:** `cr8`, `sync`, `rclone`, `music`, `s3`

**Use Cases:**
- Sync music to S3
- Backup music library
- Upload new tracks

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `dry_run` | boolean |  | Preview sync without copying (default: true) |
| `source` | string | Yes | Source path: 'music' (~/Music) or 'gdrive' (Google Drive) or custom path |

### Example

```json
{
  "dry_run": false,
  "source": "example"
}
```

---

## cr8_sync_rekordbox

Sync CR8 playlists to Rekordbox. Maps S3 paths to local files and creates/updates Rekordbox playlists.

**Complexity:** complex

**Tags:** `cr8`, `rekordbox`, `sync`, `playlist`, `dj`

**Use Cases:**
- Sync playlists to Rekordbox
- Update DJ library
- Auto-sync new tracks

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `download` | boolean |  | Download missing files from S3 (default: false) |
| `dry_run` | boolean |  | Preview sync without making changes (default: true) |
| `playlist_id` | number |  | Sync specific playlist by ID (optional) |
| `sync_local` | boolean |  | Sync S3 to local folder first using rclone (default: false) |

### Example

```json
{
  "download": false,
  "dry_run": false,
  "playlist_id": 0,
  "sync_local": false
}
```

---

## cr8_sync_status

Get sync state and history for playlists. Shows progress, resume state, and recent sync logs.

**Complexity:** simple

**Tags:** `cr8`, `sync`, `status`, `resume`, `logs`

**Use Cases:**
- Check sync progress
- View sync history
- Debug failed syncs
- Reset sync state

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `json` | boolean |  | Output as JSON (default: false) |
| `logs` | boolean |  | Include recent sync logs (default: false) |
| `playlist_id` | number |  | Filter by playlist ID (optional) |
| `reset` | number |  | Reset sync state for playlist ID (allows re-sync from start) |

### Example

```json
{
  "json": false,
  "logs": false,
  "playlist_id": 0,
  "reset": 0
}
```

---

## cr8_track_search

Search tracks by text, BPM range, key, or analysis status

**Complexity:** simple

**Tags:** `cr8`, `search`, `tracks`, `bpm`, `key`, `query`

**Use Cases:**
- Find tracks by artist
- Search by BPM range
- Find tracks in key

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `bpm` | string |  | BPM filter: single value (e.g., '128') or range (e.g., '120-130') |
| `camelot` | string |  | Camelot key filter (e.g., '8A', '5B') |
| `key` | string |  | Musical key filter (e.g., 'Am', 'C') |
| `limit` | number |  | Maximum results (default: 50) |
| `query` | string |  | Search text (matches artist or title) |
| `unanalyzed` | boolean |  | Only show unanalyzed tracks (default: false) |

### Example

```json
{
  "bpm": "example",
  "camelot": "example",
  "key": "example",
  "limit": 0,
  "query": "example",
  "unanalyzed": false
}
```

---

## cr8_update_paths

Update DynamoDB track paths from old structure to new unified structure

**Complexity:** moderate

**Tags:** `cr8`, `paths`, `update`, `migrate`, `structure`

**Use Cases:**
- Update track paths
- Migrate to new structure
- Fix path mismatches

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `dry_run` | boolean |  | Preview changes without applying (default: true) |
| `verbose` | boolean |  | Show each update (default: false) |

### Example

```json
{
  "dry_run": false,
  "verbose": false
}
```

---

## cr8_verify_migration

Comprehensive migration verification - checks Supabase vs DynamoDB data integrity

**Complexity:** complex

**Tags:** `cr8`, `migration`, `verify`, `supabase`, `dynamodb`

**Use Cases:**
- Verify migration completeness
- Check data integrity
- Compare source and target

---

## cr8_verify_sync

Verify S3 and DynamoDB are in sync. Returns matched files, orphaned S3 files, and missing files.

**Complexity:** moderate

**Tags:** `cr8`, `sync`, `verify`, `s3`, `dynamodb`, `consistency`

**Use Cases:**
- Verify data consistency
- Find orphaned S3 files
- Identify missing files

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `sample_size` | number |  | Number of orphaned/missing files to sample (default: 20) |

### Example

```json
{
  "sample_size": 0
}
```

---

