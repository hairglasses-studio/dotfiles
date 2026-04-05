# consolidated

> Token-efficient aggregated tools combining multiple data sources

**8 tools**

## Tools

- [`aftrs_equipment_audit`](#aftrs-equipment-audit)
- [`aftrs_investigate_show`](#aftrs-investigate-show)
- [`aftrs_morning_check`](#aftrs-morning-check)
- [`aftrs_performance_overview`](#aftrs-performance-overview)
- [`aftrs_pre_stream_check`](#aftrs-pre-stream-check)
- [`aftrs_show_preflight`](#aftrs-show-preflight)
- [`aftrs_stream_dashboard`](#aftrs-stream-dashboard)
- [`aftrs_studio_health_full`](#aftrs-studio-health-full)

---

## aftrs_equipment_audit

Full equipment audit: all connected devices + reliability scores + issue history. Equipment inventory.

**Complexity:** moderate

**Tags:** `equipment`, `audit`, `inventory`, `reliability`

**Use Cases:**
- Equipment review
- Maintenance planning

---

## aftrs_investigate_show

Full show investigation: current status + history + similar shows + known patterns. Comprehensive diagnostic.

**Complexity:** moderate

**Tags:** `investigate`, `show`, `history`, `patterns`

**Use Cases:**
- Deep dive into show issues
- Post-mortem analysis

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `show_name` | string |  | Name of the show to investigate |

### Example

```json
{
  "show_name": "example"
}
```

---

## aftrs_morning_check

Morning studio checkup: all systems + UNRAID + network + recent issues. Start-of-day verification.

**Complexity:** moderate

**Tags:** `morning`, `daily`, `check`, `startup`

**Use Cases:**
- Start of day check
- Daily verification

---

## aftrs_performance_overview

Visual performance overview: TouchDesigner + Resolume + DMX status combined. Saves ~65% tokens vs individual calls.

**Complexity:** moderate

**Tags:** `performance`, `visuals`, `touchdesigner`, `resolume`, `dmx`

**Use Cases:**
- Check visual systems
- Performance monitoring

---

## aftrs_pre_stream_check

Pre-stream checklist: NDI + capture + network + encoding status. Everything needed before going live.

**Complexity:** moderate

**Tags:** `stream`, `precheck`, `broadcast`, `live`

**Use Cases:**
- Before going live
- Stream readiness

---

## aftrs_show_preflight

Pre-show checklist: verifies all systems are ready for performance. Saves ~50% tokens vs manual checks.

**Complexity:** moderate

**Tags:** `show`, `preflight`, `checklist`, `verification`

**Use Cases:**
- Pre-show verification
- System readiness check

---

## aftrs_stream_dashboard

Streaming dashboard: OBS + NDI + capture devices in one view. Saves ~55% tokens vs individual calls.

**Complexity:** moderate

**Tags:** `streaming`, `ndi`, `obs`, `dashboard`

**Use Cases:**
- Monitor streaming setup
- Verify broadcast ready

---

## aftrs_studio_health_full

Get comprehensive studio health: TouchDesigner + Resolume + DMX + NDI + UNRAID in one call. Saves ~60% tokens vs individual calls.

**Complexity:** moderate

**Tags:** `health`, `status`, `consolidated`, `studio`

**Use Cases:**
- Full system check
- Morning studio verification

---

