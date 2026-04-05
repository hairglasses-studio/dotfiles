# twitch

> Twitch streaming platform integration

**10 tools**

## Tools

- [`aftrs_twitch_chat`](#aftrs-twitch-chat)
- [`aftrs_twitch_clip`](#aftrs-twitch-clip)
- [`aftrs_twitch_health`](#aftrs-twitch-health)
- [`aftrs_twitch_markers`](#aftrs-twitch-markers)
- [`aftrs_twitch_mod`](#aftrs-twitch-mod)
- [`aftrs_twitch_poll`](#aftrs-twitch-poll)
- [`aftrs_twitch_prediction`](#aftrs-twitch-prediction)
- [`aftrs_twitch_raid`](#aftrs-twitch-raid)
- [`aftrs_twitch_status`](#aftrs-twitch-status)
- [`aftrs_twitch_stream`](#aftrs-twitch-stream)

---

## aftrs_twitch_chat

Send messages to Twitch chat.

**Complexity:** simple

**Tags:** `twitch`, `chat`, `message`

**Use Cases:**
- Send announcement
- Interact with viewers

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `message` | string | Yes | Message to send |

### Example

```json
{
  "message": "example"
}
```

---

## aftrs_twitch_clip

Create clips from current stream.

**Complexity:** simple

**Tags:** `twitch`, `clip`, `highlight`

**Use Cases:**
- Create clip from stream

---

## aftrs_twitch_health

Check Twitch API health.

**Complexity:** simple

**Tags:** `twitch`, `health`, `status`

**Use Cases:**
- Check API connection
- Diagnose issues

---

## aftrs_twitch_markers

Add stream markers.

**Complexity:** simple

**Tags:** `twitch`, `marker`, `highlight`

**Use Cases:**
- Mark highlight moment

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `description` | string |  | Marker description |

### Example

```json
{
  "description": "example"
}
```

---

## aftrs_twitch_mod

Moderate chat users.

**Complexity:** simple

**Tags:** `twitch`, `moderation`, `ban`, `timeout`

**Use Cases:**
- Timeout user
- Ban user
- Unban user

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `action` | string | Yes | Action: timeout, ban, unban |
| `duration` | number |  | Timeout seconds (default 600) |
| `reason` | string |  | Reason |
| `user` | string | Yes | Username |

### Example

```json
{
  "action": "example",
  "duration": 0,
  "reason": "example",
  "user": "example"
}
```

---

## aftrs_twitch_poll

Create and manage polls.

**Complexity:** moderate

**Tags:** `twitch`, `poll`, `engagement`

**Use Cases:**
- Create viewer poll
- End poll
- List polls

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `action` | string | Yes | Action: create, end, list |
| `choices` | string |  | Comma-separated choices |
| `duration` | number |  | Duration seconds |
| `poll_id` | string |  | Poll ID for end |
| `title` | string |  | Poll title |

### Example

```json
{
  "action": "example",
  "choices": "example",
  "duration": 0,
  "poll_id": "example",
  "title": "example"
}
```

---

## aftrs_twitch_prediction

Create and manage predictions.

**Complexity:** moderate

**Tags:** `twitch`, `prediction`, `channel-points`

**Use Cases:**
- Create prediction
- Resolve prediction
- Cancel prediction

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `action` | string | Yes | Action: create, resolve, cancel, list |
| `duration` | number |  | Duration seconds |
| `outcomes` | string |  | Comma-separated outcomes |
| `prediction_id` | string |  | Prediction ID |
| `title` | string |  | Prediction title |
| `winning_outcome_id` | string |  | Winning outcome ID |

### Example

```json
{
  "action": "example",
  "duration": 0,
  "outcomes": "example",
  "prediction_id": "example",
  "title": "example",
  "winning_outcome_id": "example"
}
```

---

## aftrs_twitch_raid

Start or cancel a raid.

**Complexity:** simple

**Tags:** `twitch`, `raid`, `host`

**Use Cases:**
- Start raid to another channel
- Cancel raid

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `action` | string | Yes | Action: start, cancel |
| `channel` | string |  | Target channel |

### Example

```json
{
  "action": "example",
  "channel": "example"
}
```

---

## aftrs_twitch_status

Get Twitch stream status including live state, viewer count, and stream info.

**Complexity:** simple

**Tags:** `twitch`, `streaming`, `status`

**Use Cases:**
- Check if stream is live
- Get viewer count

---

## aftrs_twitch_stream

Control and update stream information.

**Complexity:** simple

**Tags:** `twitch`, `stream`, `title`, `game`

**Use Cases:**
- Update stream title
- Change game category

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `action` | string |  | Action: info, update_title, update_game |
| `game` | string |  | Game name to search/set |
| `title` | string |  | New stream title |

### Example

```json
{
  "action": "example",
  "game": "example",
  "title": "example"
}
```

---

