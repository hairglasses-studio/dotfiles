# stems

> AI-powered stem separation using Demucs for vocals, drums, bass, and other

**7 tools**

## Tools

- [`aftrs_stems_health`](#aftrs-stems-health)
- [`aftrs_stems_job`](#aftrs-stems-job)
- [`aftrs_stems_list`](#aftrs-stems-list)
- [`aftrs_stems_models`](#aftrs-stems-models)
- [`aftrs_stems_queue`](#aftrs-stems-queue)
- [`aftrs_stems_separate`](#aftrs-stems-separate)
- [`aftrs_stems_status`](#aftrs-stems-status)

---

## aftrs_stems_health

Check stem separation service health and get troubleshooting recommendations

**Complexity:** simple

**Tags:** `stems`, `health`, `diagnostics`, `troubleshooting`

**Use Cases:**
- Diagnose issues
- Check demucs installation
- Verify GPU setup

---

## aftrs_stems_job

Get status of a stem separation job

**Complexity:** simple

**Tags:** `stems`, `job`, `status`, `progress`

**Use Cases:**
- Check job status
- Get job result
- Monitor progress

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `job_id` | string | Yes | Job ID to check status for |

### Example

```json
{
  "job_id": "example"
}
```

---

## aftrs_stems_list

List available stems for a track or all separated tracks

**Complexity:** simple

**Tags:** `stems`, `list`, `browse`, `library`

**Use Cases:**
- Browse available stems
- Find separated tracks
- List stem files

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `track_name` | string |  | Track name to list stems for (optional - lists all if not specified) |

### Example

```json
{
  "track_name": "example"
}
```

---

## aftrs_stems_models

List available Demucs models for stem separation

**Complexity:** simple

**Tags:** `stems`, `models`, `demucs`, `config`

**Use Cases:**
- View available models
- Choose separation quality
- Model comparison

---

## aftrs_stems_queue

Queue a track for background stem separation

**Complexity:** moderate

**Tags:** `stems`, `queue`, `background`, `async`

**Use Cases:**
- Queue for later
- Background processing
- Batch separation

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `file_path` | string | Yes | Path to the audio file to separate |

### Example

```json
{
  "file_path": "example"
}
```

---

## aftrs_stems_separate

Separate an audio track into stems (vocals, drums, bass, other)

**Complexity:** complex

**Tags:** `stems`, `demucs`, `separate`, `vocals`, `drums`

**Use Cases:**
- Extract vocals
- Isolate drums
- Create acapella
- Get instrumental

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `file_path` | string | Yes | Path to the audio file to separate |

### Example

```json
{
  "file_path": "example"
}
```

---

## aftrs_stems_status

Get stem separation service status including GPU availability and active jobs

**Complexity:** simple

**Tags:** `stems`, `demucs`, `status`, `gpu`

**Use Cases:**
- Check service status
- Verify GPU availability
- View active jobs

---

