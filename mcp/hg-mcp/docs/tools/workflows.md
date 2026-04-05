# workflows

> Automation workflow tools for show management and studio operations

**14 tools**

## Tools

- [`aftrs_backup_all`](#aftrs-backup-all)
- [`aftrs_cue_sequence`](#aftrs-cue-sequence)
- [`aftrs_panic_mode`](#aftrs-panic-mode)
- [`aftrs_show_shutdown`](#aftrs-show-shutdown)
- [`aftrs_show_startup`](#aftrs-show-startup)
- [`aftrs_sync_assets`](#aftrs-sync-assets)
- [`aftrs_test_sequence`](#aftrs-test-sequence)
- [`aftrs_workflow_compose`](#aftrs-workflow-compose)
- [`aftrs_workflow_condition`](#aftrs-workflow-condition)
- [`aftrs_workflow_delete`](#aftrs-workflow-delete)
- [`aftrs_workflow_list`](#aftrs-workflow-list)
- [`aftrs_workflow_run`](#aftrs-workflow-run)
- [`aftrs_workflow_status`](#aftrs-workflow-status)
- [`aftrs_workflow_validate`](#aftrs-workflow-validate)

---

## aftrs_backup_all

Backup all project files to NAS. Includes TouchDesigner, Resolume, and vault.

**Complexity:** moderate

**Tags:** `workflow`, `backup`, `automation`

**Use Cases:**
- Full backup
- Archive projects

---

## aftrs_cue_sequence

Execute a sequence of cues. Provide cues as comma-separated 'number:name:action' items.

**Complexity:** moderate

**Tags:** `workflow`, `cue`, `sequence`, `automation`

**Use Cases:**
- Run cue list
- Automate show sections

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `cues` | string | Yes | Comma-separated cues in format 'number:name:action' (e.g., '1:Intro:play,2:Main:fade') |

### Example

```json
{
  "cues": "example"
}
```

---

## aftrs_panic_mode

EMERGENCY: Immediately stops all outputs - lighting blackout, stops playback, mutes audio, sends alert.

**Complexity:** simple

**Tags:** `workflow`, `panic`, `emergency`, `stop`

**Use Cases:**
- Emergency stop
- Crisis response

---

## aftrs_show_shutdown

Execute the graceful show shutdown sequence. Safely powers down all systems.

**Complexity:** moderate

**Tags:** `workflow`, `show`, `shutdown`, `automation`

**Use Cases:**
- End a show
- Safe shutdown

---

## aftrs_show_startup

Execute the automated show startup sequence. Brings all systems online in the correct order.

**Complexity:** moderate

**Tags:** `workflow`, `show`, `startup`, `automation`

**Use Cases:**
- Start a show
- Initialize all systems

---

## aftrs_sync_assets

Sync project assets to NAS storage.

**Complexity:** moderate

**Tags:** `workflow`, `sync`, `assets`, `automation`

**Use Cases:**
- Sync media
- Update NAS

---

## aftrs_test_sequence

Test all systems in sequence. Validates network, software, lighting, and streaming.

**Complexity:** moderate

**Tags:** `workflow`, `test`, `verify`, `automation`

**Use Cases:**
- System verification
- Pre-show test

---

## aftrs_workflow_compose

Create a workflow from a tool sequence. Use -> for sequential steps, comma for parallel. Example: 'tool1 -> tool2, tool3 -> tool4'

**Complexity:** moderate

**Tags:** `workflow`, `compose`, `create`, `automation`

**Use Cases:**
- Create custom workflows
- Build automation sequences

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `description` | string |  | Workflow description |
| `name` | string | Yes | Workflow name |
| `sequence` | string | Yes | Tool sequence: 'tool1 -> tool2' (sequential) or 'tool1, tool2' (parallel) |

### Example

```json
{
  "description": "example",
  "name": "example",
  "sequence": "example"
}
```

---

## aftrs_workflow_condition

Add a conditional execution rule to a workflow step. Step only runs if condition is met.

**Complexity:** moderate

**Tags:** `workflow`, `condition`, `if`, `branch`

**Use Cases:**
- Add conditional logic
- Create branching workflows

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `operator` | string | Yes | Comparison: eq, ne, gt, lt, gte, lte, contains, exists |
| `step_id` | string | Yes | ID of the step to add condition to |
| `value` | string |  | Value to compare against (not needed for 'exists') |
| `variable` | string | Yes | Variable to check (e.g., 'step1.success', 'env.SHOW_TYPE') |
| `workflow_id` | string | Yes | ID of the workflow |

### Example

```json
{
  "operator": "example",
  "step_id": "example",
  "value": "example",
  "variable": "example",
  "workflow_id": "example"
}
```

---

## aftrs_workflow_delete

Delete a custom workflow. Built-in workflows cannot be deleted.

**Complexity:** simple

**Tags:** `workflow`, `delete`, `remove`

**Use Cases:**
- Remove custom workflows
- Clean up old workflows

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `workflow_id` | string | Yes | ID of the workflow to delete |

### Example

```json
{
  "workflow_id": "example"
}
```

---

## aftrs_workflow_list

List all available workflows. Optionally filter by category (show, backup, maintenance, test).

**Complexity:** simple

**Tags:** `workflow`, `list`, `automation`

**Use Cases:**
- See available workflows
- Browse automation options

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `category` | string |  | Filter by category: show, backup, maintenance, test |

### Example

```json
{
  "category": "example"
}
```

---

## aftrs_workflow_run

Execute a workflow by ID. Returns execution status and can be monitored with aftrs_workflow_status.

**Complexity:** moderate

**Tags:** `workflow`, `run`, `execute`, `automation`

**Use Cases:**
- Run automation sequences
- Execute predefined workflows

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `dry_run` | boolean |  | If true, simulate without executing |
| `workflow_id` | string | Yes | ID of the workflow to run |

### Example

```json
{
  "dry_run": false,
  "workflow_id": "example"
}
```

---

## aftrs_workflow_status

Get the status of a workflow execution.

**Complexity:** simple

**Tags:** `workflow`, `status`, `monitor`

**Use Cases:**
- Check workflow progress
- Monitor execution

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `execution_id` | string | Yes | ID of the execution to check |

### Example

```json
{
  "execution_id": "example"
}
```

---

## aftrs_workflow_validate

Validate a workflow for correctness before execution. Checks dependencies, cycles, and required fields.

**Complexity:** simple

**Tags:** `workflow`, `validate`, `check`, `verify`

**Use Cases:**
- Verify workflow correctness
- Debug workflow issues

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `workflow_id` | string | Yes | ID of the workflow to validate |

### Example

```json
{
  "workflow_id": "example"
}
```

---

