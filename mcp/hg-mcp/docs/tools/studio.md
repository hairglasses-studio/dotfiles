# studio

> Studio automation including UNRAID, network, and hardware management

**5 tools**

## Tools

- [`aftrs_hardware_status`](#aftrs-hardware-status)
- [`aftrs_network_scan`](#aftrs-network-scan)
- [`aftrs_studio_health`](#aftrs-studio-health)
- [`aftrs_unraid_docker`](#aftrs-unraid-docker)
- [`aftrs_unraid_vms`](#aftrs-unraid-vms)

---

## aftrs_hardware_status

Get aggregated hardware status including GPUs, capture cards, and audio interfaces.

**Complexity:** moderate

**Tags:** `hardware`, `gpu`, `capture`, `audio`

**Use Cases:**
- Check hardware health
- Pre-show verification

---

## aftrs_network_scan

Scan the studio network for devices.

**Complexity:** moderate

**Tags:** `network`, `scan`, `devices`, `discovery`

**Use Cases:**
- Find network devices
- Check device connectivity

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `subnet` | string |  | Subnet to scan (default: auto-detect) |

### Example

```json
{
  "subnet": "example"
}
```

---

## aftrs_studio_health

Get comprehensive studio health score across all systems.

**Complexity:** moderate

**Tags:** `health`, `studio`, `overview`, `consolidated`

**Use Cases:**
- Full health check
- Morning verification

---

## aftrs_unraid_docker

List Docker containers on UNRAID with their status.

**Complexity:** simple

**Tags:** `unraid`, `docker`, `containers`

**Use Cases:**
- List containers
- Check container status

---

## aftrs_unraid_vms

List virtual machines on UNRAID with their status.

**Complexity:** simple

**Tags:** `unraid`, `vm`, `virtualization`

**Use Cases:**
- List VMs
- Check VM status

---

