# resolume_plugins

> Resolume plugin management: scan, sync, upload, and download FFGL/ISF plugins via S3

**18 tools**

## Tools

- [`aftrs_resolume_plugins_backup`](#aftrs-resolume-plugins-backup)
- [`aftrs_resolume_plugins_batch_download`](#aftrs-resolume-plugins-batch-download)
- [`aftrs_resolume_plugins_batch_upload`](#aftrs-resolume-plugins-batch-upload)
- [`aftrs_resolume_plugins_download`](#aftrs-resolume-plugins-download)
- [`aftrs_resolume_plugins_health`](#aftrs-resolume-plugins-health)
- [`aftrs_resolume_plugins_import_gdrive`](#aftrs-resolume-plugins-import-gdrive)
- [`aftrs_resolume_plugins_import_gdrive_folder`](#aftrs-resolume-plugins-import-gdrive-folder)
- [`aftrs_resolume_plugins_install`](#aftrs-resolume-plugins-install)
- [`aftrs_resolume_plugins_isf_browse`](#aftrs-resolume-plugins-isf-browse)
- [`aftrs_resolume_plugins_list`](#aftrs-resolume-plugins-list)
- [`aftrs_resolume_plugins_paths`](#aftrs-resolume-plugins-paths)
- [`aftrs_resolume_plugins_resolume_status`](#aftrs-resolume-plugins-resolume-status)
- [`aftrs_resolume_plugins_s3_list`](#aftrs-resolume-plugins-s3-list)
- [`aftrs_resolume_plugins_scan`](#aftrs-resolume-plugins-scan)
- [`aftrs_resolume_plugins_search`](#aftrs-resolume-plugins-search)
- [`aftrs_resolume_plugins_sync`](#aftrs-resolume-plugins-sync)
- [`aftrs_resolume_plugins_uninstall`](#aftrs-resolume-plugins-uninstall)
- [`aftrs_resolume_plugins_upload`](#aftrs-resolume-plugins-upload)

---

## aftrs_resolume_plugins_backup

Backup all local plugins to S3.

**Complexity:** moderate

**Tags:** `resolume`, `plugins`, `backup`, `s3`, `all`

**Use Cases:**
- Full plugin backup
- Disaster recovery prep

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `dry_run` | boolean |  | Preview what would be uploaded |
| `include_juicebar` | boolean |  | Include JuiceBar plugins (default: true) |

### Example

```json
{
  "dry_run": false,
  "include_juicebar": false
}
```

---

## aftrs_resolume_plugins_batch_download

Download multiple plugins from S3 at once.

**Complexity:** moderate

**Tags:** `resolume`, `plugins`, `download`, `batch`, `s3`

**Use Cases:**
- Download multiple plugins
- Bulk restore

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `keys` | string | Yes | Comma-separated S3 keys to download |

### Example

```json
{
  "keys": "example"
}
```

---

## aftrs_resolume_plugins_batch_upload

Upload multiple plugins to S3 at once.

**Complexity:** moderate

**Tags:** `resolume`, `plugins`, `upload`, `batch`, `s3`

**Use Cases:**
- Upload multiple plugins
- Bulk backup

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `is_juicebar` | boolean |  | Mark as JuiceBar backups |
| `paths` | string | Yes | Comma-separated local paths to plugin files |

### Example

```json
{
  "is_juicebar": false,
  "paths": "example"
}
```

---

## aftrs_resolume_plugins_download

Download a plugin from S3 and install to Extra Effects.

**Complexity:** moderate

**Tags:** `resolume`, `plugins`, `download`, `s3`, `install`

**Use Cases:**
- Download cloud plugin
- Restore from backup

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `key` | string | Yes | S3 key of the plugin (from s3_list) |

### Example

```json
{
  "key": "example"
}
```

---

## aftrs_resolume_plugins_health

Check Resolume plugin system health and get recommendations.

**Complexity:** simple

**Tags:** `resolume`, `plugins`, `health`, `status`

**Use Cases:**
- Diagnose plugin issues
- System check

---

## aftrs_resolume_plugins_import_gdrive

Import a plugin from Google Drive and optionally upload to S3.

**Complexity:** moderate

**Tags:** `resolume`, `plugins`, `gdrive`, `import`

**Use Cases:**
- Import from shared Drive
- Download plugin from link

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `file_id` | string | Yes | Google Drive file ID |
| `upload_to_s3` | boolean |  | Also upload to S3 after installing (default: true) |

### Example

```json
{
  "file_id": "example",
  "upload_to_s3": false
}
```

---

## aftrs_resolume_plugins_import_gdrive_folder

Import all plugins from a Google Drive folder.

**Complexity:** moderate

**Tags:** `resolume`, `plugins`, `gdrive`, `import`, `batch`

**Use Cases:**
- Import folder of plugins
- Bulk import from Drive

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `folder_id` | string | Yes | Google Drive folder ID |
| `upload_to_s3` | boolean |  | Also upload to S3 after installing (default: true) |

### Example

```json
{
  "folder_id": "example",
  "upload_to_s3": false
}
```

---

## aftrs_resolume_plugins_install

Install a plugin file to Resolume Extra Effects folder.

**Complexity:** simple

**Tags:** `resolume`, `plugins`, `install`, `local`

**Use Cases:**
- Install new plugin
- Add FFGL effect

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `path` | string | Yes | Path to plugin file (.bundle/.dll/.isf) |

### Example

```json
{
  "path": "example"
}
```

---

## aftrs_resolume_plugins_isf_browse

Browse ISF shader collections from the community.

**Complexity:** simple

**Tags:** `resolume`, `plugins`, `isf`, `shaders`, `browse`

**Use Cases:**
- Discover ISF shaders
- Browse shader collections

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `category` | string |  | Filter by category: effects, generators, glitch, patterns |
| `limit` | number |  | Maximum results to return |

### Example

```json
{
  "category": "example",
  "limit": 0
}
```

---

## aftrs_resolume_plugins_list

List all installed Resolume plugins (JuiceBar + Extra Effects + built-in).

**Complexity:** simple

**Tags:** `resolume`, `plugins`, `list`, `ffgl`, `isf`, `juicebar`

**Use Cases:**
- Browse installed plugins
- Inventory local plugins

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `include_builtin` | boolean |  | Include built-in plugins (default: false) |
| `type` | string |  | Filter by type: ffgl_effect, ffgl_source, isf_shader, juicebar |

### Example

```json
{
  "include_builtin": false,
  "type": "example"
}
```

---

## aftrs_resolume_plugins_paths

Show all Resolume plugin directory paths.

**Complexity:** simple

**Tags:** `resolume`, `plugins`, `paths`, `directories`

**Use Cases:**
- Find plugin folders
- Check paths

---

## aftrs_resolume_plugins_resolume_status

Check if Resolume is running and get connection info.

**Complexity:** simple

**Tags:** `resolume`, `plugins`, `status`, `connection`

**Use Cases:**
- Check Resolume status
- Verify connection

---

## aftrs_resolume_plugins_s3_list

List plugins available in the S3 bucket.

**Complexity:** simple

**Tags:** `resolume`, `plugins`, `s3`, `cloud`, `list`

**Use Cases:**
- Browse cloud plugins
- See available downloads

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `prefix` | string |  | Filter by S3 prefix (e.g., 'ffgl/', 'isf/', 'juicebar-backup/') |

### Example

```json
{
  "prefix": "example"
}
```

---

## aftrs_resolume_plugins_scan

Deep scan all Resolume plugin directories with metadata extraction.

**Complexity:** moderate

**Tags:** `resolume`, `plugins`, `scan`, `metadata`

**Use Cases:**
- Full plugin audit
- Get detailed plugin info

---

## aftrs_resolume_plugins_search

Search for plugins in S3 by name, type, or tags.

**Complexity:** simple

**Tags:** `resolume`, `plugins`, `search`, `s3`, `filter`

**Use Cases:**
- Find specific plugin
- Filter by category

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `query` | string |  | Search query (matches plugin name) |
| `tags` | string |  | Comma-separated tags to filter by |
| `type` | string |  | Filter by type: ffgl_effect, ffgl_source, isf_shader |

### Example

```json
{
  "query": "example",
  "tags": "example",
  "type": "example"
}
```

---

## aftrs_resolume_plugins_sync

Compare local and S3 plugins to identify sync opportunities.

**Complexity:** moderate

**Tags:** `resolume`, `plugins`, `sync`, `compare`

**Use Cases:**
- Check sync status
- Find missing plugins

---

## aftrs_resolume_plugins_uninstall

Remove a plugin from Resolume Extra Effects folder.

**Complexity:** simple

**Tags:** `resolume`, `plugins`, `uninstall`, `remove`

**Use Cases:**
- Remove plugin
- Clean up effects

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `name` | string | Yes | Plugin name (partial match supported) |

### Example

```json
{
  "name": "example"
}
```

---

## aftrs_resolume_plugins_upload

Upload a plugin to S3.

**Complexity:** moderate

**Tags:** `resolume`, `plugins`, `upload`, `s3`

**Use Cases:**
- Backup plugin to cloud
- Share plugin

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `is_juicebar` | boolean |  | Mark as JuiceBar backup |
| `path` | string | Yes | Local path to the plugin file |
| `tags` | string |  | Comma-separated tags |
| `version` | string |  | Version string (e.g., '1.0.0') |

### Example

```json
{
  "is_juicebar": false,
  "path": "example",
  "tags": "example",
  "version": "example"
}
```

---

