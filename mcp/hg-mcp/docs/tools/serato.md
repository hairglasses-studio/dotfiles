# serato

> Serato DJ Pro library integration for track info, crates, and history

**12 tools**

## Tools

- [`aftrs_serato_crate`](#aftrs-serato-crate)
- [`aftrs_serato_crates`](#aftrs-serato-crates)
- [`aftrs_serato_health`](#aftrs-serato-health)
- [`aftrs_serato_history`](#aftrs-serato-history)
- [`aftrs_serato_library_path`](#aftrs-serato-library-path)
- [`aftrs_serato_now_playing`](#aftrs-serato-now-playing)
- [`aftrs_serato_recent_tracks`](#aftrs-serato-recent-tracks)
- [`aftrs_serato_search`](#aftrs-serato-search)
- [`aftrs_serato_session`](#aftrs-serato-session)
- [`aftrs_serato_stats`](#aftrs-serato-stats)
- [`aftrs_serato_status`](#aftrs-serato-status)
- [`aftrs_serato_track`](#aftrs-serato-track)

---

## aftrs_serato_crate

Get tracks from a specific crate

**Complexity:** simple

**Tags:** `serato`, `dj`, `crate`, `tracks`

**Use Cases:**
- Browse crate contents
- Get track list from crate

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `name` | string | Yes | Crate name (use / for nested crates) |

### Example

```json
{
  "name": "example"
}
```

---

## aftrs_serato_crates

List all crates in the Serato library

**Complexity:** simple

**Tags:** `serato`, `dj`, `crates`, `library`

**Use Cases:**
- Browse music collection
- List available crates

---

## aftrs_serato_health

Check Serato library health and get recommendations

**Complexity:** simple

**Tags:** `serato`, `dj`, `health`, `diagnostics`

**Use Cases:**
- Diagnose library issues
- Check configuration

---

## aftrs_serato_history

Get recent play history sessions from Serato

**Complexity:** simple

**Tags:** `serato`, `dj`, `history`, `sessions`

**Use Cases:**
- Review past DJ sets
- Export setlists
- Track play statistics

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `limit` | integer |  | Maximum number of sessions to return (default: 10) |

### Example

```json
{
  "limit": 0
}
```

---

## aftrs_serato_library_path

Get the configured Serato library path

**Complexity:** simple

**Tags:** `serato`, `dj`, `config`, `library`

**Use Cases:**
- Check library location
- Verify configuration

---

## aftrs_serato_now_playing

Get the currently playing track from the most recent Serato session

**Complexity:** simple

**Tags:** `serato`, `dj`, `now-playing`, `track`

**Use Cases:**
- Get current track info
- Display now playing
- Sync visuals to music

---

## aftrs_serato_recent_tracks

Get recently played tracks across all sessions

**Complexity:** simple

**Tags:** `serato`, `dj`, `recent`, `tracks`

**Use Cases:**
- Get recently played tracks
- Quick history overview

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

## aftrs_serato_search

Search for tracks in the Serato library

**Complexity:** simple

**Tags:** `serato`, `dj`, `search`, `tracks`

**Use Cases:**
- Find tracks by name
- Search music library

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `limit` | integer |  | Maximum results (default: 50) |
| `query` | string | Yes | Search query (matches filename) |

### Example

```json
{
  "limit": 0,
  "query": "example"
}
```

---

## aftrs_serato_session

Get detailed track list from a specific history session

**Complexity:** simple

**Tags:** `serato`, `dj`, `history`, `session`, `tracks`

**Use Cases:**
- Get full setlist from a session
- Export track list

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `name` | string | Yes | Session name (filename without .session extension) |

### Example

```json
{
  "name": "example"
}
```

---

## aftrs_serato_stats

Get library statistics including track counts and crate info

**Complexity:** simple

**Tags:** `serato`, `dj`, `stats`, `library`

**Use Cases:**
- Get library overview
- Track collection statistics

---

## aftrs_serato_status

Get Serato DJ library status including connection state, crate count, and history sessions

**Complexity:** simple

**Tags:** `serato`, `dj`, `status`, `library`

**Use Cases:**
- Check if Serato library is accessible
- Get library overview

---

## aftrs_serato_track

Get detailed information about a specific track

**Complexity:** simple

**Tags:** `serato`, `dj`, `track`, `metadata`

**Use Cases:**
- Get track details
- View BPM and key information

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `path` | string | Yes | Full path to the track file |

### Example

```json
{
  "path": "example"
}
```

---

