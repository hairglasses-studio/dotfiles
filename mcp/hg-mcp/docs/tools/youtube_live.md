# youtube_live

> YouTube Live streaming integration

**8 tools**

## Tools

- [`aftrs_ytlive_broadcast`](#aftrs-ytlive-broadcast)
- [`aftrs_ytlive_channel`](#aftrs-ytlive-channel)
- [`aftrs_ytlive_chat`](#aftrs-ytlive-chat)
- [`aftrs_ytlive_health`](#aftrs-ytlive-health)
- [`aftrs_ytlive_metrics`](#aftrs-ytlive-metrics)
- [`aftrs_ytlive_search`](#aftrs-ytlive-search)
- [`aftrs_ytlive_status`](#aftrs-ytlive-status)
- [`aftrs_ytlive_video`](#aftrs-ytlive-video)

---

## aftrs_ytlive_broadcast

Get broadcast details.

**Complexity:** simple

**Tags:** `youtube`, `broadcast`, `live`

**Use Cases:**
- List broadcasts
- Get broadcast details

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `action` | string |  | Action: list, get |
| `broadcast_id` | string |  | Broadcast ID for get |
| `status` | string |  | Filter: active, all, completed, upcoming |

### Example

```json
{
  "action": "example",
  "broadcast_id": "example",
  "status": "example"
}
```

---

## aftrs_ytlive_channel

Get current live stream for a channel.

**Complexity:** simple

**Tags:** `youtube`, `channel`, `live`

**Use Cases:**
- Check if channel is live
- Get channel stream

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `channel_id` | string |  | Channel ID (uses configured default if omitted) |

### Example

```json
{
  "channel_id": "example"
}
```

---

## aftrs_ytlive_chat

Get live chat messages.

**Complexity:** simple

**Tags:** `youtube`, `chat`, `live`

**Use Cases:**
- Read live chat
- Monitor chat

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `page_token` | string |  | Pagination token |
| `video_id` | string | Yes | Video/stream ID |

### Example

```json
{
  "page_token": "example",
  "video_id": "example"
}
```

---

## aftrs_ytlive_health

Check YouTube API health.

**Complexity:** simple

**Tags:** `youtube`, `health`, `status`

**Use Cases:**
- Check API connection
- Diagnose issues

---

## aftrs_ytlive_metrics

Get live stream metrics and statistics.

**Complexity:** simple

**Tags:** `youtube`, `metrics`, `analytics`

**Use Cases:**
- Get viewer stats
- Monitor performance

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `video_id` | string | Yes | Video ID |

### Example

```json
{
  "video_id": "example"
}
```

---

## aftrs_ytlive_search

Search for live streams.

**Complexity:** simple

**Tags:** `youtube`, `search`, `live`

**Use Cases:**
- Find live streams
- Search content

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `max_results` | number |  | Max results (default 25) |
| `query` | string | Yes | Search query |

### Example

```json
{
  "max_results": 0,
  "query": "example"
}
```

---

## aftrs_ytlive_status

Get YouTube Live broadcast status and metrics.

**Complexity:** simple

**Tags:** `youtube`, `streaming`, `status`

**Use Cases:**
- Check if broadcasting
- Get viewer count

---

## aftrs_ytlive_video

Get video/stream details with live stats.

**Complexity:** simple

**Tags:** `youtube`, `video`, `stats`

**Use Cases:**
- Get video details
- Check live stats

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `video_id` | string | Yes | Video ID |

### Example

```json
{
  "video_id": "example"
}
```

---

