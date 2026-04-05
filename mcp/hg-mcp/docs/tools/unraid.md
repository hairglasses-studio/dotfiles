# unraid

> UNRAID server management including Docker, VMs, and array control

**14 tools**

## Tools

- [`aftrs_docker_containers`](#aftrs-docker-containers)
- [`aftrs_docker_logs`](#aftrs-docker-logs)
- [`aftrs_docker_restart`](#aftrs-docker-restart)
- [`aftrs_docker_start`](#aftrs-docker-start)
- [`aftrs_docker_stop`](#aftrs-docker-stop)
- [`aftrs_unraid_array_status`](#aftrs-unraid-array-status)
- [`aftrs_unraid_diagnostics`](#aftrs-unraid-diagnostics)
- [`aftrs_unraid_disks`](#aftrs-unraid-disks)
- [`aftrs_unraid_mover`](#aftrs-unraid-mover)
- [`aftrs_unraid_plugins`](#aftrs-unraid-plugins)
- [`aftrs_unraid_shares`](#aftrs-unraid-shares)
- [`aftrs_unraid_status`](#aftrs-unraid-status)
- [`aftrs_vm_control`](#aftrs-vm-control)
- [`aftrs_vm_list`](#aftrs-vm-list)

---

## aftrs_docker_containers

List all Docker containers with status.

**Complexity:** simple

**Tags:** `docker`, `containers`, `list`, `status`

**Use Cases:**
- View running containers
- Check container status

---

## aftrs_docker_logs

View Docker container logs.

**Complexity:** simple

**Tags:** `docker`, `logs`, `container`, `debug`

**Use Cases:**
- Debug container
- Check service logs

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `lines` | number |  | Number of lines (default 50) |
| `name` | string | Yes | Container name |

### Example

```json
{
  "lines": 0,
  "name": "example"
}
```

---

## aftrs_docker_restart

Restart a Docker container.

**Complexity:** simple

**Tags:** `docker`, `restart`, `container`

**Use Cases:**
- Restart service
- Apply config changes

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `name` | string | Yes | Container name to restart |

### Example

```json
{
  "name": "example"
}
```

---

## aftrs_docker_start

Start a Docker container.

**Complexity:** simple

**Tags:** `docker`, `start`, `container`

**Use Cases:**
- Start stopped container
- Launch service

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `name` | string | Yes | Container name to start |

### Example

```json
{
  "name": "example"
}
```

---

## aftrs_docker_stop

Stop a Docker container.

**Complexity:** simple

**Tags:** `docker`, `stop`, `container`

**Use Cases:**
- Stop running container
- Shutdown service

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `name` | string | Yes | Container name to stop |

### Example

```json
{
  "name": "example"
}
```

---

## aftrs_unraid_array_status

Get detailed array status including parity and mover.

**Complexity:** simple

**Tags:** `unraid`, `array`, `parity`, `status`

**Use Cases:**
- Check array health
- Monitor parity status

---

## aftrs_unraid_diagnostics

Generate UNRAID diagnostics bundle.

**Complexity:** moderate

**Tags:** `unraid`, `diagnostics`, `debug`, `support`

**Use Cases:**
- Generate diagnostics
- Prepare support request

---

## aftrs_unraid_disks

List array disks with status, temperature, and health.

**Complexity:** simple

**Tags:** `unraid`, `disks`, `array`, `health`

**Use Cases:**
- Check disk health
- Monitor temperatures

---

## aftrs_unraid_mover

Trigger the mover to transfer cache to array.

**Complexity:** moderate

**Tags:** `unraid`, `mover`, `cache`, `array`

**Use Cases:**
- Manual mover run
- Clear cache

---

## aftrs_unraid_plugins

List installed UNRAID plugins.

**Complexity:** simple

**Tags:** `unraid`, `plugins`, `list`

**Use Cases:**
- View installed plugins
- Check for updates

---

## aftrs_unraid_shares

List user shares with usage information.

**Complexity:** simple

**Tags:** `unraid`, `shares`, `storage`

**Use Cases:**
- Check share usage
- View storage allocation

---

## aftrs_unraid_status

Get UNRAID server status including array, CPU, memory, and disk usage.

**Complexity:** simple

**Tags:** `unraid`, `status`, `server`, `health`

**Use Cases:**
- Check server status
- Monitor resources

---

## aftrs_vm_control

Control virtual machine (start/stop/pause/resume).

**Complexity:** moderate

**Tags:** `vm`, `control`, `start`, `stop`

**Use Cases:**
- Start/stop VMs
- Manage VM state

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `action` | string | Yes | Action: start, stop, pause, resume, force_stop |
| `name` | string | Yes | VM name |

### Example

```json
{
  "action": "example",
  "name": "example"
}
```

---

## aftrs_vm_list

List virtual machines with status.

**Complexity:** simple

**Tags:** `vm`, `virtual`, `list`, `status`

**Use Cases:**
- View VMs
- Check VM status

---

