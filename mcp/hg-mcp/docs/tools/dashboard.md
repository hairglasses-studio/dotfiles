# dashboard

> Unified status dashboard and health monitoring

**4 tools**

## Tools

- [`aftrs_dashboard_alerts`](#aftrs-dashboard-alerts)
- [`aftrs_dashboard_full`](#aftrs-dashboard-full)
- [`aftrs_dashboard_quick`](#aftrs-dashboard-quick)
- [`aftrs_dashboard_watch`](#aftrs-dashboard-watch)

---

## aftrs_dashboard_alerts

View and manage active system alerts.

**Complexity:** simple

**Tags:** `dashboard`, `alerts`, `warnings`

**Use Cases:**
- View system alerts
- Acknowledge issues

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `action` | string |  | Action: 'list' (default), 'resolve', or 'clear' |
| `alert_id` | string |  | Alert ID to resolve (required for 'resolve' action) |
| `include_resolved` | boolean |  | Include resolved alerts in list (default: false) |

### Example

```json
{
  "action": "example",
  "alert_id": "example",
  "include_resolved": false
}
```

---

## aftrs_dashboard_full

Get detailed status of all systems with health scores and latency.

**Complexity:** moderate

**Tags:** `dashboard`, `status`, `health`, `detailed`

**Use Cases:**
- Detailed health check
- Troubleshooting
- System inventory

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `category` | string |  | Filter by category: audio, video, lighting, network, automation, ai |

### Example

```json
{
  "category": "example"
}
```

---

## aftrs_dashboard_quick

Get a quick one-line status overview of all connected systems.

**Complexity:** simple

**Tags:** `dashboard`, `status`, `health`, `quick`

**Use Cases:**
- Quick system overview
- Pre-show health check

---

## aftrs_dashboard_watch

Start or stop background health monitoring with automatic alerts.

**Complexity:** simple

**Tags:** `dashboard`, `monitoring`, `watch`, `alerts`

**Use Cases:**
- Continuous monitoring
- Proactive alerting

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `action` | string |  | Action: 'start', 'stop', or 'status' (default: status) |
| `interval` | number |  | Polling interval in seconds (default: 30, min: 10, max: 300) |

### Example

```json
{
  "action": "example",
  "interval": 0
}
```

---

