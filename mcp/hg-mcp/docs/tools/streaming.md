# streaming

> NDI source discovery and streaming health monitoring

**4 tools**

## Tools

- [`aftrs_ndi_sources`](#aftrs-ndi-sources)
- [`aftrs_ndi_status`](#aftrs-ndi-status)
- [`aftrs_stream_health`](#aftrs-stream-health)
- [`aftrs_stream_start`](#aftrs-stream-start)

---

## aftrs_ndi_sources

List available NDI sources on the network.

**Complexity:** simple

**Tags:** `ndi`, `sources`, `discovery`, `video`

**Use Cases:**
- Find NDI sources
- Check available video feeds

---

## aftrs_ndi_status

Get detailed status of an NDI source including frame rate and bandwidth.

**Complexity:** simple

**Tags:** `ndi`, `status`, `video`, `performance`

**Use Cases:**
- Check NDI source health
- Monitor video quality

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `source` | string | Yes | NDI source name or partial match |

### Example

```json
{
  "source": "example"
}
```

---

## aftrs_stream_health

Get combined streaming health check including NDI, OBS, and capture devices.

**Complexity:** moderate

**Tags:** `streaming`, `health`, `ndi`, `obs`, `video`

**Use Cases:**
- Check streaming infrastructure
- Pre-stream verification

---

## aftrs_stream_start

Start streaming to a destination (placeholder for future OBS integration).

**Complexity:** moderate

**Tags:** `streaming`, `start`, `obs`, `broadcast`

**Use Cases:**
- Start live stream
- Begin broadcast

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `destination` | string | Yes | Streaming destination (e.g., 'twitch', 'youtube', 'custom') |

### Example

```json
{
  "destination": "example"
}
```

---

