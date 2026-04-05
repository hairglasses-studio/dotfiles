# Google Tasks Module

> Google Tasks integration for task management with lists, due dates, and completion tracking

**13 tools**

## Tools

| Tool | Description |
|------|-------------|
| `aftrs_tasks_lists` | List all Google Tasks lists |
| `aftrs_tasks_list_create` | Create a new task list |
| `aftrs_tasks_list_delete` | Delete a task list |
| `aftrs_tasks_get` | Get tasks from a specific list |
| `aftrs_tasks_today` | Get tasks due today |
| `aftrs_tasks_overdue` | Get overdue tasks |
| `aftrs_tasks_upcoming` | Get tasks due in the next N days |
| `aftrs_task_create` | Create a new task |
| `aftrs_task_update` | Update an existing task |
| `aftrs_task_complete` | Mark a task as completed |
| `aftrs_task_uncomplete` | Reopen a completed task |
| `aftrs_task_delete` | Delete a task |
| `aftrs_tasks_clear_completed` | Clear all completed tasks from a list |

## Authentication

Requires Google OAuth credentials via:
- `GOOGLE_APPLICATION_CREDENTIALS` environment variable pointing to a service account JSON file
- Or `GOOGLE_API_KEY` for API key authentication

## Use Cases

- Track daily tasks and to-dos
- Manage project task lists
- Set due dates and reminders
- Integrate with Google Workspace workflows
- Sync tasks across devices via Android/Google ecosystem
