# backup

> Project backup automation and restore capabilities

**4 tools**

## Tools

- [`aftrs_backup_list`](#aftrs-backup-list)
- [`aftrs_backup_projects`](#aftrs-backup-projects)
- [`aftrs_backup_restore`](#aftrs-backup-restore)
- [`aftrs_backup_status`](#aftrs-backup-status)

---

## aftrs_backup_list

List all backups for a project.

**Complexity:** simple

**Tags:** `backup`, `list`, `history`, `versions`

**Use Cases:**
- View backup versions
- List available backups

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `project` | string | Yes | Project name |

### Example

```json
{
  "project": "example"
}
```

---

## aftrs_backup_projects

Create a backup of a project directory.

**Complexity:** moderate

**Tags:** `backup`, `archive`, `project`, `tar`

**Use Cases:**
- Backup project files
- Create project archive

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `exclude` | string |  | Comma-separated patterns to exclude |
| `project` | string | Yes | Project name |
| `source` | string | Yes | Source directory path |

### Example

```json
{
  "exclude": "example",
  "project": "example",
  "source": "example"
}
```

---

## aftrs_backup_restore

Restore a backup to a target directory.

**Complexity:** moderate

**Tags:** `backup`, `restore`, `recover`, `extract`

**Use Cases:**
- Restore from backup
- Recover project files

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `backup_id` | string | Yes | Backup ID to restore |
| `project` | string | Yes | Project name |
| `target` | string | Yes | Target directory for restore |

### Example

```json
{
  "backup_id": "example",
  "project": "example",
  "target": "example"
}
```

---

## aftrs_backup_status

Get backup status for all projects.

**Complexity:** simple

**Tags:** `backup`, `status`, `list`, `overview`

**Use Cases:**
- Check backup status
- View backup history

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `project` | string |  | Filter by project name |

### Example

```json
{
  "project": "example"
}
```

---

