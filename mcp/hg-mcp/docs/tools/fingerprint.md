# fingerprint

> Audio fingerprinting using Chromaprint and AcoustID for track identification

**5 tools**

## Tools

- [`aftrs_fingerprint_batch`](#aftrs-fingerprint-batch)
- [`aftrs_fingerprint_duplicates`](#aftrs-fingerprint-duplicates)
- [`aftrs_fingerprint_generate`](#aftrs-fingerprint-generate)
- [`aftrs_fingerprint_health`](#aftrs-fingerprint-health)
- [`aftrs_fingerprint_match`](#aftrs-fingerprint-match)

---

## aftrs_fingerprint_batch

Generate fingerprints for multiple audio files

**Complexity:** complex

**Tags:** `fingerprint`, `batch`, `bulk`, `generate`

**Use Cases:**
- Fingerprint multiple tracks
- Batch processing
- Library analysis

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `file_paths` | array | Yes | List of audio file paths to fingerprint |

### Example

```json
{
  "file_paths": []
}
```

---

## aftrs_fingerprint_duplicates

Find duplicate tracks in a directory using fingerprint comparison

**Complexity:** complex

**Tags:** `fingerprint`, `duplicates`, `library`, `cleanup`

**Use Cases:**
- Find duplicate tracks
- Clean up library
- Detect same songs

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `directory` | string | Yes | Directory to scan for duplicate tracks |
| `threshold` | number |  | Similarity threshold (0.0-1.0, default: 0.9) |

### Example

```json
{
  "directory": "example",
  "threshold": 0
}
```

---

## aftrs_fingerprint_generate

Generate an audio fingerprint for a track using Chromaprint

**Complexity:** moderate

**Tags:** `fingerprint`, `chromaprint`, `audio`, `identify`

**Use Cases:**
- Generate fingerprint
- Prepare for matching
- Create audio signature

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `file_path` | string | Yes | Path to the audio file to fingerprint |

### Example

```json
{
  "file_path": "example"
}
```

---

## aftrs_fingerprint_health

Check fingerprinting service health and configuration

**Complexity:** simple

**Tags:** `fingerprint`, `health`, `diagnostics`, `status`

**Use Cases:**
- Check fpcalc installation
- Verify AcoustID key
- Diagnose issues

---

## aftrs_fingerprint_match

Match a fingerprint against AcoustID database to identify a track

**Complexity:** moderate

**Tags:** `fingerprint`, `acoustid`, `identify`, `match`

**Use Cases:**
- Identify unknown track
- Find track metadata
- Music recognition

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `file_path` | string | Yes | Path to the audio file to identify |

### Example

```json
{
  "file_path": "example"
}
```

---

