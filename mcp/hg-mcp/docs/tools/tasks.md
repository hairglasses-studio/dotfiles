# tasks

> Async task management for long-running operations

**5 tools**

## Tools

- [`aftrs_task_cancel`](#aftrs-task-cancel)
- [`aftrs_task_cleanup`](#aftrs-task-cleanup)
- [`aftrs_task_list`](#aftrs-task-list)
- [`aftrs_task_status`](#aftrs-task-status)
- [`aftrs_task_summary`](#aftrs-task-summary)

---

## aftrs_task_cancel

Cancel a pending or running task.

**Complexity:** simple

**Tags:** `task`, `async`, `cancel`, `stop`

**Use Cases:**
- Cancel long-running task
- Stop processing

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `task_id` | string | Yes | Task ID to cancel |

### Example

```json
{
  "task_id": "example"
}
```

---

## aftrs_task_cleanup

Remove old completed/failed tasks from memory.

**Complexity:** simple

**Tags:** `task`, `cleanup`, `maintenance`

**Use Cases:**
- Clean up old tasks
- Free memory

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `older_than_hours` | number |  | Remove tasks older than N hours (default: 24) |

### Example

```json
{
  "older_than_hours": 0
}
```

---

## aftrs_task_list

List all async tasks. Optionally filter by status or type.

**Complexity:** simple

**Tags:** `task`, `async`, `list`, `status`

**Use Cases:**
- View running tasks
- Check task queue

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `status` | string |  | Filter by status: pending, working, completed, failed, cancelled |
| `type` | string |  | Filter by task type (e.g., stems, inference, backup) |

### Example

```json
{
  "status": "example",
  "type": "example"
}
```

---

## aftrs_task_status

Get detailed status of a specific task.

**Complexity:** simple

**Tags:** `task`, `async`, `status`, `progress`

**Use Cases:**
- Check task progress
- Get task result

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `task_id` | string | Yes | Task ID to check |

### Example

```json
{
  "task_id": "example"
}
```

---

## aftrs_task_summary

Get a summary of all tasks by status.

**Complexity:** simple

**Tags:** `task`, `summary`, `overview`

**Use Cases:**
- View task queue health
- Monitor async operations

---

