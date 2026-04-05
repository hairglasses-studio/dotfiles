# vault

> Obsidian vault knowledge management and documentation

**11 tools**

## Tools

- [`aftrs_project_notes`](#aftrs-project-notes)
- [`aftrs_runbook_search`](#aftrs-runbook-search)
- [`aftrs_session_get`](#aftrs-session-get)
- [`aftrs_session_list`](#aftrs-session-list)
- [`aftrs_session_summary`](#aftrs-session-summary)
- [`aftrs_setlist_get`](#aftrs-setlist-get)
- [`aftrs_setlist_list`](#aftrs-setlist-list)
- [`aftrs_show_history`](#aftrs-show-history)
- [`aftrs_show_log`](#aftrs-show-log)
- [`aftrs_vault_save`](#aftrs-vault-save)
- [`aftrs_vault_search`](#aftrs-vault-search)

---

## aftrs_project_notes

Get notes for a specific project.

**Complexity:** simple

**Tags:** `vault`, `project`, `notes`, `documentation`

**Use Cases:**
- View project documentation
- Get project context

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

## aftrs_runbook_search

Search for operational runbooks.

**Complexity:** simple

**Tags:** `vault`, `runbook`, `operations`, `howto`

**Use Cases:**
- Find troubleshooting guides
- Get setup instructions

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `query` | string | Yes | Search query |

### Example

```json
{
  "query": "example"
}
```

---

## aftrs_session_get

Get the content of a specific session.

**Complexity:** simple

**Tags:** `vault`, `session`, `details`, `view`

**Use Cases:**
- View session details
- Review past session

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `path` | string | Yes | Session path (from session_list) |

### Example

```json
{
  "path": "example"
}
```

---

## aftrs_session_list

List studio sessions for a date or recent sessions.

**Complexity:** simple

**Tags:** `vault`, `session`, `list`, `history`

**Use Cases:**
- View today's sessions
- Find past sessions

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `date` | string |  | Date in YYYY-MM-DD format (default: today) |
| `limit` | number |  | Number of recent sessions to return if no date specified |

### Example

```json
{
  "date": "example",
  "limit": 0
}
```

---

## aftrs_session_summary

Get aggregated session summary for a period.

**Complexity:** simple

**Tags:** `vault`, `session`, `summary`, `analytics`

**Use Cases:**
- Weekly activity report
- Session statistics

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `period` | string |  | Period: 'daily', 'weekly', or 'monthly' (default: weekly) |

### Example

```json
{
  "period": "example"
}
```

---

## aftrs_setlist_get

Get details for a specific setlist.

**Complexity:** simple

**Tags:** `vault`, `setlist`, `performance`, `cues`

**Use Cases:**
- Load setlist for show
- Review past performance

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `name` | string | Yes | Setlist name |

### Example

```json
{
  "name": "example"
}
```

---

## aftrs_setlist_list

List all performance setlists.

**Complexity:** simple

**Tags:** `vault`, `setlist`, `performance`, `show`

**Use Cases:**
- Browse setlists
- Find past performances

---

## aftrs_show_history

Get past show history.

**Complexity:** simple

**Tags:** `vault`, `show`, `history`, `archive`

**Use Cases:**
- Review past shows
- Find historical data

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `limit` | number |  | Maximum number of shows to return (default: 10) |

### Example

```json
{
  "limit": 0
}
```

---

## aftrs_show_log

Log an event for the current show.

**Complexity:** simple

**Tags:** `vault`, `show`, `log`, `events`

**Use Cases:**
- Log show events
- Track issues during performance

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `description` | string | Yes | Event description |
| `details` | string |  | Additional details (optional) |
| `event_type` | string | Yes | Event type: 'start', 'end', 'cue', 'issue', 'note' |

### Example

```json
{
  "description": "example",
  "details": "example",
  "event_type": "example"
}
```

---

## aftrs_vault_save

Save a document to the vault.

**Complexity:** simple

**Tags:** `vault`, `save`, `write`, `notes`

**Use Cases:**
- Save session notes
- Create documentation

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `content` | string | Yes | Document content (markdown) |
| `path` | string | Yes | Document path within vault (e.g., 'sessions/2024-01-15') |

### Example

```json
{
  "content": "example",
  "path": "example"
}
```

---

## aftrs_vault_search

Search the vault for documents matching a query.

**Complexity:** simple

**Tags:** `vault`, `search`, `documentation`, `notes`

**Use Cases:**
- Find project notes
- Search runbooks

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `query` | string | Yes | Search query |

### Example

```json
{
  "query": "example"
}
```

---

