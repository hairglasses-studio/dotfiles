# Chains Module

> Workflow chain execution for multi-step automated operations

## Tools (8)

### aftrs_chain_list

List available workflow chains.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| category | string | No | Filter by category (show, stream, backup, etc.) |

**Example:**
```json
{"category": "show"}
```

---

### aftrs_chain_get

Get details of a specific chain including steps and parameters.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| chain_id | string | Yes | Chain ID to retrieve |

**Example:**
```json
{"chain_id": "show_startup"}
```

---

### aftrs_chain_execute

Execute a workflow chain with optional parameters.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| chain_id | string | Yes | Chain ID to execute |
| params | object | No | Parameters for the chain |

**Example:**
```json
{
  "chain_id": "stream_start",
  "params": {
    "platform": "twitch"
  }
}
```

---

### aftrs_chain_status

Get the status of a chain execution.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| execution_id | string | Yes | Execution ID to check |

**Example:**
```json
{"execution_id": "exec_abc123"}
```

---

### aftrs_chain_approve

Approve or reject a pending gate in a chain execution.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| execution_id | string | Yes | Execution ID |
| step_id | string | Yes | Gate step ID |
| approved | boolean | Yes | Whether to approve (true) or reject (false) |
| comment | string | No | Optional comment |

**Example:**
```json
{
  "execution_id": "exec_abc123",
  "step_id": "step_2",
  "approved": true,
  "comment": "Systems verified"
}
```

---

### aftrs_chain_cancel

Cancel a running chain execution.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| execution_id | string | Yes | Execution ID to cancel |

**Example:**
```json
{"execution_id": "exec_abc123"}
```

---

### aftrs_chain_pending

List all pending gate approvals across executions.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| - | - | - | No parameters required |

**Example:**
```json
{}
```

---

### aftrs_chain_history

Get recent chain execution history.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| limit | integer | No | Maximum results (default: 20) |

**Example:**
```json
{"limit": 10}
```

---

## Built-in Chains

### show_startup
**Category:** show
**Description:** Complete show startup sequence
**Steps:**
1. Power on studio systems
2. Check all system health
3. Load TouchDesigner project
4. Test video outputs

### stream_start
**Category:** stream
**Description:** Start streaming workflow
**Steps:**
1. Verify OBS is ready
2. Start OBS stream
3. Verify stream health
4. Notify Discord channel

### stream_end
**Category:** stream
**Description:** End streaming workflow
**Steps:**
1. Stop OBS stream
2. Save recording
3. Notify Discord channel

### backup_projects
**Category:** backup
**Description:** Backup all project files
**Steps:**
1. Create timestamped backup
2. Sync to cloud storage
3. Verify backup integrity

### dj_set_prep
**Category:** show
**Description:** Prepare for DJ set
**Steps:**
1. Check Serato status
2. Check audio routing
3. Check lighting system
4. Run pre-show lighting test

### lighting_test
**Category:** show
**Description:** Test all lighting fixtures
**Steps:**
1. Test all DMX fixtures
2. Test WLED devices
3. Verify grandMA3 connection

---

## Chain Step Types

| Type | Description |
|------|-------------|
| `tool` | Execute a single MCP tool |
| `chain` | Execute another chain (nested) |
| `parallel` | Execute multiple steps concurrently |
| `branch` | Conditional execution based on expression |
| `gate` | Pause for manual approval |
| `delay` | Wait for specified duration |

## Error Handling

Each step can specify an `on_error` action:

| Action | Description |
|--------|-------------|
| `stop` | Stop chain execution on error |
| `continue` | Log error and continue to next step |
| `retry` | Retry the step (up to max_retries) |
