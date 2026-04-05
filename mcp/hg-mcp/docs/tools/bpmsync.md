# bpmsync

> Unified BPM synchronization across Ableton, Resolume, grandMA3, and other systems

**6 tools**

## Tools

- [`aftrs_sync_bpm_health`](#aftrs-sync-bpm-health)
- [`aftrs_sync_bpm_link`](#aftrs-sync-bpm-link)
- [`aftrs_sync_bpm_master`](#aftrs-sync-bpm-master)
- [`aftrs_sync_bpm_push`](#aftrs-sync-bpm-push)
- [`aftrs_sync_bpm_status`](#aftrs-sync-bpm-status)
- [`aftrs_sync_tap_tempo`](#aftrs-sync-tap-tempo)

---

## aftrs_sync_bpm_health

Check BPM sync health and get troubleshooting recommendations

**Complexity:** simple

**Tags:** `bpm`, `health`, `diagnostics`

**Use Cases:**
- Check sync health
- Diagnose sync issues
- Verify configuration

---

## aftrs_sync_bpm_link

Link or unlink a system to receive BPM updates from master

**Complexity:** simple

**Tags:** `bpm`, `link`, `sync`, `system`

**Use Cases:**
- Link Resolume to master
- Unlink grandMA3
- Configure sync targets

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `linked` | boolean |  | True to link, false to unlink |
| `system` | string | Yes | System to link/unlink: ableton, resolume, grandma3 |

### Example

```json
{
  "linked": false,
  "system": "example"
}
```

---

## aftrs_sync_bpm_master

Set the master BPM source (ableton, resolume, or manual)

**Complexity:** simple

**Tags:** `bpm`, `master`, `source`, `config`

**Use Cases:**
- Set Ableton as master
- Switch to manual tempo
- Change sync source

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `source` | string | Yes | Master source: ableton, resolume, or manual |

### Example

```json
{
  "source": "example"
}
```

---

## aftrs_sync_bpm_push

Push a BPM value to all linked systems, or sync from master

**Complexity:** moderate

**Tags:** `bpm`, `push`, `sync`, `tempo`

**Use Cases:**
- Set tempo to 128 BPM
- Sync all systems from master
- Force tempo sync

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `bpm` | number |  | BPM to push (20-999). Omit to sync from current master. |

### Example

```json
{
  "bpm": 0
}
```

---

## aftrs_sync_bpm_status

Get current BPM across all systems and sync status

**Complexity:** simple

**Tags:** `bpm`, `sync`, `tempo`, `status`

**Use Cases:**
- Check BPM across systems
- View sync status
- Detect tempo drift

---

## aftrs_sync_tap_tempo

Record a tap for tap tempo calculation across all linked systems

**Complexity:** simple

**Tags:** `bpm`, `tap`, `tempo`

**Use Cases:**
- Tap to set tempo
- Manual tempo detection
- Live tempo adjustment

---

