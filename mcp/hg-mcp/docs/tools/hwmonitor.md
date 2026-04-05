# hwmonitor

> Hardware monitoring for CPU, GPU temperature and power consumption

**4 tools**

## Tools

- [`aftrs_cpu_temperature`](#aftrs-cpu-temperature)
- [`aftrs_gpu_temperature`](#aftrs-gpu-temperature)
- [`aftrs_power_consumption`](#aftrs-power-consumption)
- [`aftrs_thermal_alert`](#aftrs-thermal-alert)

---

## aftrs_cpu_temperature

Get CPU temperature and information.

**Complexity:** simple

**Tags:** `cpu`, `temperature`, `monitoring`, `hardware`

**Use Cases:**
- Monitor CPU heat
- Check thermal status

---

## aftrs_gpu_temperature

Get GPU temperature and information.

**Complexity:** simple

**Tags:** `gpu`, `temperature`, `monitoring`, `hardware`, `nvidia`, `amd`

**Use Cases:**
- Monitor GPU heat
- Check graphics card status

---

## aftrs_power_consumption

Get power consumption information for CPU and GPU.

**Complexity:** simple

**Tags:** `power`, `consumption`, `watts`, `energy`

**Use Cases:**
- Monitor power usage
- Check energy consumption

---

## aftrs_thermal_alert

Get thermal status and any active temperature alerts.

**Complexity:** moderate

**Tags:** `thermal`, `alert`, `warning`, `critical`, `temperature`

**Use Cases:**
- Check for overheating
- Configure thermal alerts

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `set_cpu_critical` | string |  | Set CPU critical threshold (Celsius) |
| `set_cpu_warning` | string |  | Set CPU warning threshold (Celsius) |
| `set_gpu_critical` | string |  | Set GPU critical threshold (Celsius) |
| `set_gpu_warning` | string |  | Set GPU warning threshold (Celsius) |

### Example

```json
{
  "set_cpu_critical": "example",
  "set_cpu_warning": "example",
  "set_gpu_critical": "example",
  "set_gpu_warning": "example"
}
```

---

