# tailscale

> Tailscale VPN network management and device connectivity

**3 tools**

## Tools

- [`aftrs_tailscale_devices`](#aftrs-tailscale-devices)
- [`aftrs_tailscale_ping`](#aftrs-tailscale-ping)
- [`aftrs_tailscale_status`](#aftrs-tailscale-status)

---

## aftrs_tailscale_devices

List all devices in the Tailscale network.

**Complexity:** simple

**Tags:** `tailscale`, `devices`, `peers`, `network`

**Use Cases:**
- List network devices
- Find device IPs

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `online_only` | boolean |  | Show only online devices |

### Example

```json
{
  "online_only": false
}
```

---

## aftrs_tailscale_ping

Ping a device in the Tailscale network.

**Complexity:** simple

**Tags:** `tailscale`, `ping`, `connectivity`, `latency`

**Use Cases:**
- Test device connectivity
- Measure latency

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `target` | string | Yes | Device hostname or IP to ping |

### Example

```json
{
  "target": "example"
}
```

---

## aftrs_tailscale_status

Get Tailscale network status and connected devices.

**Complexity:** simple

**Tags:** `tailscale`, `vpn`, `network`, `status`

**Use Cases:**
- Check VPN status
- View connected devices

---

