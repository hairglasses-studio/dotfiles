# dante

> Dante audio network discovery, routing, and monitoring

**7 tools**

## Tools

- [`aftrs_dante_devices`](#aftrs-dante-devices)
- [`aftrs_dante_discover`](#aftrs-dante-discover)
- [`aftrs_dante_health`](#aftrs-dante-health)
- [`aftrs_dante_latency`](#aftrs-dante-latency)
- [`aftrs_dante_route`](#aftrs-dante-route)
- [`aftrs_dante_routes`](#aftrs-dante-routes)
- [`aftrs_dante_status`](#aftrs-dante-status)

---

## aftrs_dante_devices

List all known Dante devices with channels and status

**Complexity:** simple

**Tags:** `dante`, `devices`, `list`

**Use Cases:**
- List all devices
- View device channels
- Check device status

---

## aftrs_dante_discover

Discover Dante devices on the network using mDNS

**Complexity:** moderate

**Tags:** `dante`, `discover`, `mdns`, `devices`

**Use Cases:**
- Find Dante devices
- Refresh device list
- Network scan

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `timeout` | integer |  | Discovery timeout in seconds (default: 3) |

### Example

```json
{
  "timeout": 0
}
```

---

## aftrs_dante_health

Check Dante network health and get troubleshooting recommendations

**Complexity:** simple

**Tags:** `dante`, `health`, `diagnostics`, `troubleshooting`

**Use Cases:**
- Diagnose network issues
- Check device health
- Get troubleshooting tips

---

## aftrs_dante_latency

Check Dante network latency and performance metrics

**Complexity:** simple

**Tags:** `dante`, `latency`, `performance`, `monitoring`

**Use Cases:**
- Check network latency
- Monitor performance
- Diagnose issues

---

## aftrs_dante_route

Create or delete an audio route between Dante devices

**Complexity:** complex

**Tags:** `dante`, `route`, `routing`, `patch`

**Use Cases:**
- Create audio route
- Delete audio route
- Patch channels

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `action` | string | Yes | Action to perform: create or delete |
| `rx_channel` | integer | Yes | Receiving channel number |
| `rx_device` | string | Yes | Receiving device name |
| `tx_channel` | integer |  | Transmitting channel number |
| `tx_device` | string |  | Transmitting device name |

### Example

```json
{
  "action": "example",
  "rx_channel": 0,
  "rx_device": "example",
  "tx_channel": 0,
  "tx_device": "example"
}
```

---

## aftrs_dante_routes

List all audio routes between Dante devices

**Complexity:** simple

**Tags:** `dante`, `routes`, `routing`, `audio`

**Use Cases:**
- View audio routing
- Check channel connections
- Audit routing

---

## aftrs_dante_status

Get Dante network status including device count and master clock

**Complexity:** simple

**Tags:** `dante`, `audio`, `network`, `status`

**Use Cases:**
- Check network status
- Verify Dante connectivity
- View master clock

---

