# analytics

> Performance analytics and session tracking

**6 tools**

## Tools

- [`aftrs_analytics_report`](#aftrs-analytics-report)
- [`aftrs_session_end`](#aftrs-session-end)
- [`aftrs_session_start`](#aftrs-session-start)
- [`aftrs_session_status`](#aftrs-session-status)
- [`aftrs_sessions_list`](#aftrs-sessions-list)
- [`aftrs_track_log`](#aftrs-track-log)

---

## aftrs_analytics_report

Generate analytics report across all sessions. Shows most played tracks, artists, venue stats, and trends.

**Complexity:** simple

**Tags:** `analytics`, `report`, `stats`, `trends`

**Use Cases:**
- View play history
- Analyze performance trends

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `period` | string |  | Time period: all, week, month, year (default: all) |

### Example

```json
{
  "period": "example"
}
```

---

## aftrs_session_end

End the current performance session. Calculates final metrics and saves session data.

**Complexity:** simple

**Tags:** `session`, `end`, `stop`, `save`

**Use Cases:**
- End DJ set
- Complete performance logging

---

## aftrs_session_start

Start a new performance session for tracking. Logs tracks, transitions, and BPM throughout the set.

**Complexity:** simple

**Tags:** `session`, `start`, `tracking`, `performance`

**Use Cases:**
- Start DJ set tracking
- Begin live performance logging

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `name` | string | Yes | Session name (e.g., 'Friday Night Set') |
| `type` | string |  | Session type: dj, live, hybrid (default: dj) |
| `venue` | string |  | Venue name |

### Example

```json
{
  "name": "example",
  "type": "example",
  "venue": "example"
}
```

---

## aftrs_session_status

Get current session status with live metrics. Shows tracks played, BPM history, and performance stats.

**Complexity:** simple

**Tags:** `session`, `status`, `metrics`, `current`

**Use Cases:**
- Check set progress
- View live metrics

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `session_id` | string |  | Session ID (default: current session) |

### Example

```json
{
  "session_id": "example"
}
```

---

## aftrs_sessions_list

List all recorded performance sessions.

**Complexity:** simple

**Tags:** `sessions`, `list`, `history`

**Use Cases:**
- View past sets
- Find session to analyze

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `limit` | number |  | Maximum sessions to show (default: 10) |

### Example

```json
{
  "limit": 0
}
```

---

## aftrs_track_log

Log a track play in the current session. Automatically calculates transitions.

**Complexity:** simple

**Tags:** `track`, `log`, `play`, `record`

**Use Cases:**
- Log track in set
- Record play history

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `artist` | string | Yes | Track artist |
| `bpm` | number | Yes | Track BPM |
| `energy` | number |  | Energy level 1-10 |
| `key` | string |  | Track key (e.g., '5A', 'Dm') |
| `source` | string |  | Source: rekordbox, serato, traktor, ableton |
| `title` | string | Yes | Track title |

### Example

```json
{
  "artist": "example",
  "bpm": 0,
  "energy": 0,
  "key": "example",
  "source": "example",
  "title": "example"
}
```

---

