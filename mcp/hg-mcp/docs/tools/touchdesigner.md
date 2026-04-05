# touchdesigner

> TouchDesigner project control and monitoring

**25 tools**

## Tools

- [`aftrs_td_backup`](#aftrs-td-backup)
- [`aftrs_td_chop_channels`](#aftrs-td-chop-channels)
- [`aftrs_td_chops`](#aftrs-td-chops)
- [`aftrs_td_components`](#aftrs-td-components)
- [`aftrs_td_cue_trigger`](#aftrs-td-cue-trigger)
- [`aftrs_td_custom_pars`](#aftrs-td-custom-pars)
- [`aftrs_td_dat_content`](#aftrs-td-dat-content)
- [`aftrs_td_dats`](#aftrs-td-dats)
- [`aftrs_td_errors`](#aftrs-td-errors)
- [`aftrs_td_execute`](#aftrs-td-execute)
- [`aftrs_td_gpu_memory`](#aftrs-td-gpu-memory)
- [`aftrs_td_network_health`](#aftrs-td-network-health)
- [`aftrs_td_operators`](#aftrs-td-operators)
- [`aftrs_td_parameters`](#aftrs-td-parameters)
- [`aftrs_td_performance`](#aftrs-td-performance)
- [`aftrs_td_preset_recall`](#aftrs-td-preset-recall)
- [`aftrs_td_project_info`](#aftrs-td-project-info)
- [`aftrs_td_pulse`](#aftrs-td-pulse)
- [`aftrs_td_render_settings`](#aftrs-td-render-settings)
- [`aftrs_td_reset`](#aftrs-td-reset)
- [`aftrs_td_status`](#aftrs-td-status)
- [`aftrs_td_textures`](#aftrs-td-textures)
- [`aftrs_td_timelines`](#aftrs-td-timelines)
- [`aftrs_td_tox_export`](#aftrs-td-tox-export)
- [`aftrs_td_variables`](#aftrs-td-variables)

---

## aftrs_td_backup

Create a backup of the current TouchDesigner project.

**Complexity:** moderate

**Tags:** `touchdesigner`, `backup`, `save`, `project`

**Use Cases:**
- Create project backup
- Save before changes

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `destination` | string |  | Backup file path (optional, auto-generates timestamp) |

### Example

```json
{
  "destination": "example"
}
```

---

## aftrs_td_chop_channels

Get channel values from a specific CHOP operator.

**Complexity:** simple

**Tags:** `touchdesigner`, `chop`, `channels`, `values`

**Use Cases:**
- Read channel values
- Monitor CHOP output

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `chop` | string | Yes | CHOP operator path |

### Example

```json
{
  "chop": "example"
}
```

---

## aftrs_td_chops

List CHOP operators in a network with channel counts.

**Complexity:** simple

**Tags:** `touchdesigner`, `chop`, `channels`, `audio`

**Use Cases:**
- List CHOPs
- Find audio/control operators

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `path` | string |  | Network path to scan (default: /project1) |

### Example

```json
{
  "path": "example"
}
```

---

## aftrs_td_components

List container COMPs in a network.

**Complexity:** simple

**Tags:** `touchdesigner`, `components`, `comp`, `containers`

**Use Cases:**
- Browse components
- Navigate project structure

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `path` | string |  | Network path to scan (default: /project1) |

### Example

```json
{
  "path": "example"
}
```

---

## aftrs_td_cue_trigger

Trigger a cue or jump to position on a timeline.

**Complexity:** moderate

**Tags:** `touchdesigner`, `cue`, `timeline`, `trigger`

**Use Cases:**
- Trigger show cues
- Jump to timeline position

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `cue` | string |  | Cue name or frame number |
| `timeline` | string | Yes | Timeline operator path |

### Example

```json
{
  "cue": "example",
  "timeline": "example"
}
```

---

## aftrs_td_custom_pars

List custom parameters on an operator.

**Complexity:** simple

**Tags:** `touchdesigner`, `custom`, `parameters`, `interface`

**Use Cases:**
- View custom parameters
- Check interface settings

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `operator` | string | Yes | Operator path |

### Example

```json
{
  "operator": "example"
}
```

---

## aftrs_td_dat_content

Get or set the content of a DAT operator.

**Complexity:** simple

**Tags:** `touchdesigner`, `dat`, `content`, `text`

**Use Cases:**
- Read DAT content
- Update DAT data

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `content` | string |  | Content to set (omit to read) |
| `dat` | string | Yes | DAT operator path |

### Example

```json
{
  "content": "example",
  "dat": "example"
}
```

---

## aftrs_td_dats

List DAT operators in a network.

**Complexity:** simple

**Tags:** `touchdesigner`, `dat`, `table`, `text`

**Use Cases:**
- List DATs
- Find data operators

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `path` | string |  | Network path to scan (default: /project1) |

### Example

```json
{
  "path": "example"
}
```

---

## aftrs_td_errors

Get current errors and warnings from TouchDesigner.

**Complexity:** simple

**Tags:** `touchdesigner`, `errors`, `warnings`, `debug`

**Use Cases:**
- View error log
- Debug issues

---

## aftrs_td_execute

Execute Python code in TouchDesigner's context.

**Complexity:** moderate

**Tags:** `touchdesigner`, `python`, `scripting`, `execute`

**Use Cases:**
- Run custom Python scripts
- Automate TD operations

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `script` | string | Yes | Python script to execute |

### Example

```json
{
  "script": "example"
}
```

---

## aftrs_td_gpu_memory

Get GPU memory usage breakdown.

**Complexity:** simple

**Tags:** `touchdesigner`, `gpu`, `memory`, `vram`

**Use Cases:**
- Monitor GPU memory
- Prevent VRAM exhaustion

---

## aftrs_td_network_health

Get health score and analysis for a TouchDesigner operator network.

**Complexity:** moderate

**Tags:** `touchdesigner`, `health`, `performance`, `analysis`

**Use Cases:**
- Analyze network health
- Find performance issues

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `path` | string |  | Network path to analyze (default: /project1) |

### Example

```json
{
  "path": "example"
}
```

---

## aftrs_td_operators

List operators in a TouchDesigner network.

**Complexity:** simple

**Tags:** `touchdesigner`, `operators`, `network`

**Use Cases:**
- Browse operator network
- Find specific operators

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `path` | string |  | Network path to list (default: /project1) |
| `type` | string |  | Filter by operator type (e.g., 'TOP', 'CHOP', 'DAT') |

### Example

```json
{
  "path": "example",
  "type": "example"
}
```

---

## aftrs_td_parameters

Get or set parameters on a TouchDesigner operator.

**Complexity:** simple

**Tags:** `touchdesigner`, `parameters`, `control`

**Use Cases:**
- Read operator parameters
- Modify operator settings

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `operator` | string | Yes | Operator path (e.g., /project1/geo1) |
| `param` | string |  | Parameter name to get/set (omit to list all) |
| `value` | string |  | Value to set (omit to just read) |

### Example

```json
{
  "operator": "example",
  "param": "example",
  "value": "example"
}
```

---

## aftrs_td_performance

Get performance profiling data including cook times and frame timing.

**Complexity:** simple

**Tags:** `touchdesigner`, `performance`, `profiling`, `cook`

**Use Cases:**
- Profile performance
- Find bottlenecks

---

## aftrs_td_preset_recall

Recall a parameter preset to an operator.

**Complexity:** moderate

**Tags:** `touchdesigner`, `preset`, `recall`, `parameters`

**Use Cases:**
- Load saved presets
- Recall show settings

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `operator` | string |  | Target operator path (optional) |
| `preset` | string | Yes | Preset file path or name |

### Example

```json
{
  "operator": "example",
  "preset": "example"
}
```

---

## aftrs_td_project_info

Get detailed information about the current TouchDesigner project.

**Complexity:** simple

**Tags:** `touchdesigner`, `project`, `info`, `statistics`

**Use Cases:**
- View project details
- Get operator counts

---

## aftrs_td_pulse

Pulse a parameter on an operator.

**Complexity:** simple

**Tags:** `touchdesigner`, `pulse`, `trigger`, `parameter`

**Use Cases:**
- Trigger pulse parameters
- Reset operators

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `operator` | string | Yes | Operator path |
| `param` | string | Yes | Parameter name to pulse |

### Example

```json
{
  "operator": "example",
  "param": "example"
}
```

---

## aftrs_td_render_settings

Get or set render output settings (resolution, format, etc).

**Complexity:** simple

**Tags:** `touchdesigner`, `render`, `output`, `resolution`

**Use Cases:**
- Check render settings
- Change output resolution

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `fps` | number |  | Output FPS |
| `height` | number |  | Output height |
| `width` | number |  | Output width |

### Example

```json
{
  "fps": 0,
  "height": 0,
  "width": 0
}
```

---

## aftrs_td_reset

Reset an operator to its default state.

**Complexity:** simple

**Tags:** `touchdesigner`, `reset`, `defaults`, `operator`

**Use Cases:**
- Reset operator
- Restore defaults

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `operator` | string | Yes | Operator path to reset |

### Example

```json
{
  "operator": "example"
}
```

---

## aftrs_td_status

Get TouchDesigner project status including FPS, cook time, and error counts.

**Complexity:** simple

**Tags:** `touchdesigner`, `status`, `fps`, `performance`

**Use Cases:**
- Check TD performance
- Monitor project health

---

## aftrs_td_textures

List texture TOPs with resolution and memory usage.

**Complexity:** simple

**Tags:** `touchdesigner`, `textures`, `top`, `memory`

**Use Cases:**
- Find large textures
- Audit texture memory usage

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `path` | string |  | Network path to scan (default: /project1) |

### Example

```json
{
  "path": "example"
}
```

---

## aftrs_td_timelines

List timeline CHOPs with playback status.

**Complexity:** simple

**Tags:** `touchdesigner`, `timeline`, `chop`, `playback`

**Use Cases:**
- View timeline status
- Check playback state

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `path` | string |  | Network path to scan (default: /project1) |

### Example

```json
{
  "path": "example"
}
```

---

## aftrs_td_tox_export

Export an operator/component as a TOX file.

**Complexity:** moderate

**Tags:** `touchdesigner`, `tox`, `export`, `component`

**Use Cases:**
- Export component
- Share TOX file

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `destination` | string |  | Destination file path |
| `operator` | string | Yes | Operator path to export |

### Example

```json
{
  "destination": "example",
  "operator": "example"
}
```

---

## aftrs_td_variables

Get or set project variables.

**Complexity:** simple

**Tags:** `touchdesigner`, `variables`, `project`, `global`

**Use Cases:**
- View project variables
- Set global values

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `name` | string |  | Variable name (omit to list all) |
| `value` | string |  | Value to set (omit to read) |

### Example

```json
{
  "name": "example",
  "value": "example"
}
```

---

