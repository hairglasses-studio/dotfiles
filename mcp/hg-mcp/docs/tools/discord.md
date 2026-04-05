# discord

> Discord bot integration for team communication

**47 tools**

## Tools

- [`aftrs_discord_admin_assign_role`](#aftrs-discord-admin-assign-role)
- [`aftrs_discord_admin_audit_log`](#aftrs-discord-admin-audit-log)
- [`aftrs_discord_admin_ban`](#aftrs-discord-admin-ban)
- [`aftrs_discord_admin_create_category`](#aftrs-discord-admin-create-category)
- [`aftrs_discord_admin_create_channel`](#aftrs-discord-admin-create-channel)
- [`aftrs_discord_admin_create_event`](#aftrs-discord-admin-create-event)
- [`aftrs_discord_admin_create_role`](#aftrs-discord-admin-create-role)
- [`aftrs_discord_admin_delete_channel`](#aftrs-discord-admin-delete-channel)
- [`aftrs_discord_admin_delete_event`](#aftrs-discord-admin-delete-event)
- [`aftrs_discord_admin_delete_role`](#aftrs-discord-admin-delete-role)
- [`aftrs_discord_admin_edit_channel`](#aftrs-discord-admin-edit-channel)
- [`aftrs_discord_admin_edit_role`](#aftrs-discord-admin-edit-role)
- [`aftrs_discord_admin_kick`](#aftrs-discord-admin-kick)
- [`aftrs_discord_admin_list_bans`](#aftrs-discord-admin-list-bans)
- [`aftrs_discord_admin_list_events`](#aftrs-discord-admin-list-events)
- [`aftrs_discord_admin_list_members`](#aftrs-discord-admin-list-members)
- [`aftrs_discord_admin_list_roles`](#aftrs-discord-admin-list-roles)
- [`aftrs_discord_admin_lock_channel`](#aftrs-discord-admin-lock-channel)
- [`aftrs_discord_admin_member_info`](#aftrs-discord-admin-member-info)
- [`aftrs_discord_admin_pin`](#aftrs-discord-admin-pin)
- [`aftrs_discord_admin_purge`](#aftrs-discord-admin-purge)
- [`aftrs_discord_admin_revoke_role`](#aftrs-discord-admin-revoke-role)
- [`aftrs_discord_admin_send_buttons`](#aftrs-discord-admin-send-buttons)
- [`aftrs_discord_admin_send_select`](#aftrs-discord-admin-send-select)
- [`aftrs_discord_admin_server_info`](#aftrs-discord-admin-server-info)
- [`aftrs_discord_admin_set_nickname`](#aftrs-discord-admin-set-nickname)
- [`aftrs_discord_admin_slowmode`](#aftrs-discord-admin-slowmode)
- [`aftrs_discord_admin_timeout`](#aftrs-discord-admin-timeout)
- [`aftrs_discord_admin_unban`](#aftrs-discord-admin-unban)
- [`aftrs_discord_admin_update_event`](#aftrs-discord-admin-update-event)
- [`aftrs_discord_admin_voice_deafen`](#aftrs-discord-admin-voice-deafen)
- [`aftrs_discord_admin_voice_disconnect`](#aftrs-discord-admin-voice-disconnect)
- [`aftrs_discord_admin_voice_move`](#aftrs-discord-admin-voice-move)
- [`aftrs_discord_admin_voice_mute`](#aftrs-discord-admin-voice-mute)
- [`aftrs_discord_admin_voice_status`](#aftrs-discord-admin-voice-status)
- [`aftrs_discord_channels`](#aftrs-discord-channels)
- [`aftrs_discord_delete`](#aftrs-discord-delete)
- [`aftrs_discord_edit`](#aftrs-discord-edit)
- [`aftrs_discord_history`](#aftrs-discord-history)
- [`aftrs_discord_notify`](#aftrs-discord-notify)
- [`aftrs_discord_react`](#aftrs-discord-react)
- [`aftrs_discord_send`](#aftrs-discord-send)
- [`aftrs_discord_status`](#aftrs-discord-status)
- [`aftrs_discord_studio_event`](#aftrs-discord-studio-event)
- [`aftrs_discord_thread`](#aftrs-discord-thread)
- [`aftrs_discord_thread_create`](#aftrs-discord-thread-create)
- [`aftrs_discord_webhook`](#aftrs-discord-webhook)

---

## aftrs_discord_admin_assign_role

Assign a role to a member.

**Complexity:** moderate

**Tags:** `discord`, `admin`, `role`, `assign`

**Use Cases:**
- Grant access
- Add team member

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `role_id` | string | Yes | ID of the role to assign |
| `user_id` | string | Yes | ID of the user |

### Example

```json
{
  "role_id": "example",
  "user_id": "example"
}
```

---

## aftrs_discord_admin_audit_log

Get the server audit log.

**Complexity:** moderate

**Tags:** `discord`, `admin`, `audit`, `log`, `security`

**Use Cases:**
- Track changes
- Investigate incidents

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `action_type` | number |  | Filter by action type (see Discord docs for codes) |
| `limit` | number |  | Number of entries to retrieve (default: 50, max: 100) |
| `user_id` | string |  | Filter by user who performed the action |

### Example

```json
{
  "action_type": 0,
  "limit": 0,
  "user_id": "example"
}
```

---

## aftrs_discord_admin_ban

Ban a member from the server.

**Complexity:** complex

**Tags:** `discord`, `admin`, `ban`, `moderation`

**Use Cases:**
- Permanently remove users
- Block bad actors

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `delete_days` | number |  | Days of messages to delete (0-7, default: 0) |
| `reason` | string |  | Reason for the ban (logged in audit) |
| `user_id` | string | Yes | ID of the user to ban |

### Example

```json
{
  "delete_days": 0,
  "reason": "example",
  "user_id": "example"
}
```

---

## aftrs_discord_admin_create_category

Create a new channel category.

**Complexity:** moderate

**Tags:** `discord`, `admin`, `category`, `create`

**Use Cases:**
- Organize channels by project
- Create team sections

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `name` | string | Yes | Category name |
| `position` | number |  | Position in channel list (optional) |

### Example

```json
{
  "name": "example",
  "position": 0
}
```

---

## aftrs_discord_admin_create_channel

Create a new text or voice channel in the server.

**Complexity:** moderate

**Tags:** `discord`, `admin`, `channel`, `create`

**Use Cases:**
- Create project channels
- Add voice rooms

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `category_id` | string |  | Parent category ID (optional) |
| `name` | string | Yes | Channel name |
| `topic` | string |  | Channel topic (text channels only) |
| `type` | string |  | Channel type: 'text' (default), 'voice', 'news', 'forum' |

### Example

```json
{
  "category_id": "example",
  "name": "example",
  "topic": "example",
  "type": "example"
}
```

---

## aftrs_discord_admin_create_event

Create a scheduled event.

**Complexity:** moderate

**Tags:** `discord`, `admin`, `event`, `schedule`

**Use Cases:**
- Schedule jam sessions
- Plan meetings

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `channel_id` | string |  | Voice/Stage channel ID (for voice events) |
| `description` | string |  | Event description |
| `end_time` | string |  | End time in RFC3339 format (required for external events) |
| `location` | string |  | External location (makes this an external event) |
| `name` | string | Yes | Event name |
| `start_time` | string | Yes | Start time in RFC3339 format (e.g., 2024-01-15T14:00:00Z) |

### Example

```json
{
  "channel_id": "example",
  "description": "example",
  "end_time": "example",
  "location": "example",
  "name": "example",
  "start_time": "example"
}
```

---

## aftrs_discord_admin_create_role

Create a new role.

**Complexity:** moderate

**Tags:** `discord`, `admin`, `role`, `create`

**Use Cases:**
- Create team roles
- Add department tags

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `color` | string |  | Role color in hex (e.g., '#FF0000' or '0xFF0000') |
| `hoist` | boolean |  | Display separately in member list (default: false) |
| `mentionable` | boolean |  | Allow @mentions of this role (default: false) |
| `name` | string | Yes | Role name |

### Example

```json
{
  "color": "example",
  "hoist": false,
  "mentionable": false,
  "name": "example"
}
```

---

## aftrs_discord_admin_delete_channel

Delete a channel from the server.

**Complexity:** moderate

**Tags:** `discord`, `admin`, `channel`, `delete`

**Use Cases:**
- Remove archived channels
- Cleanup

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `channel_id` | string | Yes | ID of the channel to delete |

### Example

```json
{
  "channel_id": "example"
}
```

---

## aftrs_discord_admin_delete_event

Delete a scheduled event.

**Complexity:** moderate

**Tags:** `discord`, `admin`, `event`, `delete`

**Use Cases:**
- Cancel events
- Remove old events

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `event_id` | string | Yes | ID of the event to delete |

### Example

```json
{
  "event_id": "example"
}
```

---

## aftrs_discord_admin_delete_role

Delete a role from the server.

**Complexity:** moderate

**Tags:** `discord`, `admin`, `role`, `delete`

**Use Cases:**
- Remove obsolete roles
- Cleanup

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `role_id` | string | Yes | ID of the role to delete |

### Example

```json
{
  "role_id": "example"
}
```

---

## aftrs_discord_admin_edit_channel

Edit channel settings (name, topic, slowmode).

**Complexity:** moderate

**Tags:** `discord`, `admin`, `channel`, `edit`

**Use Cases:**
- Update channel settings
- Set slowmode

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `channel_id` | string | Yes | ID of the channel to edit |
| `name` | string |  | New channel name |
| `slowmode` | number |  | Slowmode delay in seconds (0-21600) |
| `topic` | string |  | New channel topic |

### Example

```json
{
  "channel_id": "example",
  "name": "example",
  "slowmode": 0,
  "topic": "example"
}
```

---

## aftrs_discord_admin_edit_role

Edit an existing role.

**Complexity:** moderate

**Tags:** `discord`, `admin`, `role`, `edit`

**Use Cases:**
- Update role settings
- Change color

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `color` | string |  | New role color in hex |
| `name` | string |  | New role name |
| `role_id` | string | Yes | ID of the role to edit |

### Example

```json
{
  "color": "example",
  "name": "example",
  "role_id": "example"
}
```

---

## aftrs_discord_admin_kick

Kick a member from the server.

**Complexity:** complex

**Tags:** `discord`, `admin`, `kick`, `moderation`

**Use Cases:**
- Remove rule violators
- Enforce policies

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `reason` | string |  | Reason for the kick (logged in audit) |
| `user_id` | string | Yes | ID of the user to kick |

### Example

```json
{
  "reason": "example",
  "user_id": "example"
}
```

---

## aftrs_discord_admin_list_bans

List banned users.

**Complexity:** simple

**Tags:** `discord`, `admin`, `ban`, `list`

**Use Cases:**
- Review bans
- Audit moderation

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `limit` | number |  | Maximum bans to return (default: 100) |

### Example

```json
{
  "limit": 0
}
```

---

## aftrs_discord_admin_list_events

List scheduled events.

**Complexity:** simple

**Tags:** `discord`, `admin`, `event`, `list`

**Use Cases:**
- View upcoming events
- Check schedule

---

## aftrs_discord_admin_list_members

List members in the server.

**Complexity:** simple

**Tags:** `discord`, `admin`, `member`, `list`

**Use Cases:**
- Browse members
- Audit server

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `limit` | number |  | Maximum members to return (default: 100, max: 1000) |

### Example

```json
{
  "limit": 0
}
```

---

## aftrs_discord_admin_list_roles

List all roles in the server.

**Complexity:** simple

**Tags:** `discord`, `admin`, `role`, `list`

**Use Cases:**
- View role hierarchy
- Find role IDs

---

## aftrs_discord_admin_lock_channel

Lock or unlock a channel (prevent/allow messages).

**Complexity:** moderate

**Tags:** `discord`, `admin`, `channel`, `lock`, `moderation`

**Use Cases:**
- Temporarily disable channel
- Prevent spam

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `channel_id` | string | Yes | ID of the channel to lock/unlock |
| `locked` | boolean | Yes | true to lock, false to unlock |

### Example

```json
{
  "channel_id": "example",
  "locked": false
}
```

---

## aftrs_discord_admin_member_info

Get detailed information about a member.

**Complexity:** simple

**Tags:** `discord`, `admin`, `member`, `info`

**Use Cases:**
- Check user details
- View roles

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `user_id` | string | Yes | ID of the user |

### Example

```json
{
  "user_id": "example"
}
```

---

## aftrs_discord_admin_pin

Pin or unpin a message.

**Complexity:** simple

**Tags:** `discord`, `admin`, `pin`, `message`

**Use Cases:**
- Highlight important messages
- Save announcements

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `channel_id` | string | Yes | ID of the channel |
| `message_id` | string | Yes | ID of the message |
| `pinned` | boolean | Yes | true to pin, false to unpin |

### Example

```json
{
  "channel_id": "example",
  "message_id": "example",
  "pinned": false
}
```

---

## aftrs_discord_admin_purge

Bulk delete messages from a channel (2-100 messages, max 14 days old).

**Complexity:** complex

**Tags:** `discord`, `admin`, `purge`, `delete`, `moderation`

**Use Cases:**
- Clean up spam
- Remove off-topic messages

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `channel_id` | string | Yes | ID of the channel |
| `count` | number | Yes | Number of messages to delete (2-100) |

### Example

```json
{
  "channel_id": "example",
  "count": 0
}
```

---

## aftrs_discord_admin_revoke_role

Remove a role from a member.

**Complexity:** moderate

**Tags:** `discord`, `admin`, `role`, `revoke`

**Use Cases:**
- Revoke access
- Remove from team

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `role_id` | string | Yes | ID of the role to remove |
| `user_id` | string | Yes | ID of the user |

### Example

```json
{
  "role_id": "example",
  "user_id": "example"
}
```

---

## aftrs_discord_admin_send_buttons

Send a message with action buttons.

**Complexity:** complex

**Tags:** `discord`, `admin`, `buttons`, `interactive`

**Use Cases:**
- Approval workflows
- Quick actions

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `buttons` | string | Yes | JSON array of buttons: [{"label": "...", "custom_id": "...", "style": "primary|secondary|success|danger"}] |
| `channel_id` | string |  | Channel to send to (uses default if not specified) |
| `content` | string | Yes | Message content |

### Example

```json
{
  "buttons": "example",
  "channel_id": "example",
  "content": "example"
}
```

---

## aftrs_discord_admin_send_select

Send a message with a select menu.

**Complexity:** complex

**Tags:** `discord`, `admin`, `select`, `dropdown`, `interactive`

**Use Cases:**
- Multi-choice forms
- Selection menus

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `channel_id` | string |  | Channel to send to (uses default if not specified) |
| `content` | string | Yes | Message content |
| `custom_id` | string | Yes | Custom ID for the select menu |
| `options` | string | Yes | JSON array of options: [{"label": "...", "value": "...", "description": "..."}] |
| `placeholder` | string |  | Placeholder text for the select menu |

### Example

```json
{
  "channel_id": "example",
  "content": "example",
  "custom_id": "example",
  "options": "example",
  "placeholder": "example"
}
```

---

## aftrs_discord_admin_server_info

Get detailed server information.

**Complexity:** simple

**Tags:** `discord`, `admin`, `server`, `info`

**Use Cases:**
- Audit server settings
- Check server stats

---

## aftrs_discord_admin_set_nickname

Set a member's nickname.

**Complexity:** moderate

**Tags:** `discord`, `admin`, `nickname`, `member`

**Use Cases:**
- Enforce naming conventions
- Set display names

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `nickname` | string | Yes | New nickname (empty string to clear) |
| `user_id` | string | Yes | ID of the user |

### Example

```json
{
  "nickname": "example",
  "user_id": "example"
}
```

---

## aftrs_discord_admin_slowmode

Set slowmode delay for a channel.

**Complexity:** simple

**Tags:** `discord`, `admin`, `slowmode`, `moderation`

**Use Cases:**
- Rate limit discussions
- Prevent spam

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `channel_id` | string | Yes | ID of the channel |
| `seconds` | number | Yes | Slowmode delay in seconds (0 to disable, max: 21600) |

### Example

```json
{
  "channel_id": "example",
  "seconds": 0
}
```

---

## aftrs_discord_admin_timeout

Timeout a member (temporary mute).

**Complexity:** moderate

**Tags:** `discord`, `admin`, `timeout`, `mute`, `moderation`

**Use Cases:**
- Temporary silence
- Cool down period

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `duration_minutes` | number | Yes | Timeout duration in minutes (max: 40320 = 28 days) |
| `reason` | string |  | Reason for the timeout |
| `user_id` | string | Yes | ID of the user to timeout |

### Example

```json
{
  "duration_minutes": 0,
  "reason": "example",
  "user_id": "example"
}
```

---

## aftrs_discord_admin_unban

Remove a ban from a user.

**Complexity:** moderate

**Tags:** `discord`, `admin`, `unban`, `moderation`

**Use Cases:**
- Restore access
- Second chance

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `user_id` | string | Yes | ID of the user to unban |

### Example

```json
{
  "user_id": "example"
}
```

---

## aftrs_discord_admin_update_event

Update an existing scheduled event.

**Complexity:** moderate

**Tags:** `discord`, `admin`, `event`, `update`

**Use Cases:**
- Reschedule events
- Update details

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `description` | string |  | New event description |
| `event_id` | string | Yes | ID of the event to update |
| `name` | string |  | New event name |
| `start_time` | string |  | New start time in RFC3339 format |

### Example

```json
{
  "description": "example",
  "event_id": "example",
  "name": "example",
  "start_time": "example"
}
```

---

## aftrs_discord_admin_voice_deafen

Server deafen/undeafen a member in voice.

**Complexity:** moderate

**Tags:** `discord`, `admin`, `voice`, `deafen`

**Use Cases:**
- Full audio isolation
- Private discussions

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `deafened` | boolean | Yes | true to deafen, false to undeafen |
| `user_id` | string | Yes | ID of the user |

### Example

```json
{
  "deafened": false,
  "user_id": "example"
}
```

---

## aftrs_discord_admin_voice_disconnect

Disconnect a member from voice.

**Complexity:** moderate

**Tags:** `discord`, `admin`, `voice`, `disconnect`

**Use Cases:**
- Clear idle users
- End sessions

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `user_id` | string | Yes | ID of the user to disconnect |

### Example

```json
{
  "user_id": "example"
}
```

---

## aftrs_discord_admin_voice_move

Move a member to a different voice channel.

**Complexity:** moderate

**Tags:** `discord`, `admin`, `voice`, `move`

**Use Cases:**
- Redirect to meeting room
- Organize teams

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `channel_id` | string | Yes | ID of the destination voice channel |
| `user_id` | string | Yes | ID of the user to move |

### Example

```json
{
  "channel_id": "example",
  "user_id": "example"
}
```

---

## aftrs_discord_admin_voice_mute

Server mute/unmute a member in voice.

**Complexity:** moderate

**Tags:** `discord`, `admin`, `voice`, `mute`

**Use Cases:**
- Enforce silence during recording
- Mute during presentations

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `muted` | boolean | Yes | true to mute, false to unmute |
| `user_id` | string | Yes | ID of the user |

### Example

```json
{
  "muted": false,
  "user_id": "example"
}
```

---

## aftrs_discord_admin_voice_status

Get status of all voice channels (who is where).

**Complexity:** simple

**Tags:** `discord`, `admin`, `voice`, `status`

**Use Cases:**
- See who is in voice
- Monitor activity

---

## aftrs_discord_channels

List available Discord channels in the configured server.

**Complexity:** simple

**Tags:** `discord`, `channels`, `list`

**Use Cases:**
- Find channel IDs
- Browse server structure

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `type` | string |  | Filter by channel type: 'text', 'voice', 'category', or 'all' (default: all) |

### Example

```json
{
  "type": "example"
}
```

---

## aftrs_discord_delete

Delete a message.

**Complexity:** simple

**Tags:** `discord`, `delete`, `message`

**Use Cases:**
- Remove messages
- Clean up channels

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `channel_id` | string |  | Channel ID containing the message |
| `message_id` | string | Yes | ID of the message to delete |

### Example

```json
{
  "channel_id": "example",
  "message_id": "example"
}
```

---

## aftrs_discord_edit

Edit an existing message.

**Complexity:** simple

**Tags:** `discord`, `edit`, `message`

**Use Cases:**
- Update messages
- Fix typos

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `channel_id` | string |  | Channel ID containing the message |
| `content` | string | Yes | New message content |
| `message_id` | string | Yes | ID of the message to edit |

### Example

```json
{
  "channel_id": "example",
  "content": "example",
  "message_id": "example"
}
```

---

## aftrs_discord_history

Get message history from a Discord channel.

**Complexity:** simple

**Tags:** `discord`, `history`, `messages`

**Use Cases:**
- Read recent messages
- Check channel activity

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `channel_id` | string |  | Channel ID to get history from (uses default if not specified) |
| `limit` | number |  | Number of messages to retrieve (default: 25, max: 100) |

### Example

```json
{
  "channel_id": "example",
  "limit": 0
}
```

---

## aftrs_discord_notify

Send a formatted notification with title, message, and severity level.

**Complexity:** moderate

**Tags:** `discord`, `notify`, `alert`, `embed`

**Use Cases:**
- Send formatted alerts
- Studio notifications

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `channel_id` | string |  | Channel ID to send to (uses default if not specified) |
| `level` | string |  | Severity level: 'info', 'success', 'warning', 'error' (default: info) |
| `message` | string | Yes | Notification body text |
| `title` | string | Yes | Notification title |

### Example

```json
{
  "channel_id": "example",
  "level": "example",
  "message": "example",
  "title": "example"
}
```

---

## aftrs_discord_react

Add a reaction emoji to a message.

**Complexity:** simple

**Tags:** `discord`, `react`, `emoji`

**Use Cases:**
- React to messages
- Acknowledge notifications

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `channel_id` | string |  | Channel ID containing the message |
| `emoji` | string | Yes | Emoji to react with (e.g., '👍', ':thumbsup:', or custom emoji ID) |
| `message_id` | string | Yes | ID of the message to react to |

### Example

```json
{
  "channel_id": "example",
  "emoji": "example",
  "message_id": "example"
}
```

---

## aftrs_discord_send

Send a message to a Discord channel.

**Complexity:** simple

**Tags:** `discord`, `send`, `message`

**Use Cases:**
- Send notifications
- Post updates

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `channel_id` | string |  | Channel ID to send to (uses default if not specified) |
| `message` | string | Yes | The message content to send |

### Example

```json
{
  "channel_id": "example",
  "message": "example"
}
```

---

## aftrs_discord_status

Check Discord bot connection status and latency.

**Complexity:** simple

**Tags:** `discord`, `status`, `connection`

**Use Cases:**
- Check if bot is connected
- Monitor bot health

---

## aftrs_discord_studio_event

Send a studio event notification (stream, recording, TD error, session).

**Complexity:** moderate

**Tags:** `discord`, `studio`, `event`, `notification`

**Use Cases:**
- Stream alerts
- TD error notifications
- Session logging

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `channel_id` | string |  | Channel ID (uses default if not specified) |
| `details` | string |  | Event details/description |
| `event_type` | string | Yes | Event type: stream_start, stream_stop, record_start, record_stop, td_error, td_warning, health_warning, session_start, session_end |
| `title` | string | Yes | Event title |

### Example

```json
{
  "channel_id": "example",
  "details": "example",
  "event_type": "example",
  "title": "example"
}
```

---

## aftrs_discord_thread

Get messages from a Discord thread.

**Complexity:** simple

**Tags:** `discord`, `thread`, `messages`

**Use Cases:**
- Read thread discussions
- Follow conversations

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `limit` | number |  | Number of messages to retrieve (default: 25, max: 100) |
| `thread_id` | string | Yes | The thread ID to get messages from |

### Example

```json
{
  "limit": 0,
  "thread_id": "example"
}
```

---

## aftrs_discord_thread_create

Create a new thread in a channel.

**Complexity:** moderate

**Tags:** `discord`, `thread`, `create`

**Use Cases:**
- Start discussions
- Organize conversations

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `channel_id` | string |  | Channel ID to create thread in |
| `message_id` | string |  | Message ID to create thread from (optional - creates standalone thread if not provided) |
| `name` | string | Yes | Thread name |

### Example

```json
{
  "channel_id": "example",
  "message_id": "example",
  "name": "example"
}
```

---

## aftrs_discord_webhook

Send a message via Discord webhook URL.

**Complexity:** moderate

**Tags:** `discord`, `webhook`, `notification`

**Use Cases:**
- Send webhook notifications
- External integrations

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `color` | string |  | Embed color: 'red', 'green', 'blue', 'orange', or hex (default: blue) |
| `content` | string |  | Plain text content (optional if embed is provided) |
| `description` | string |  | Embed description/body |
| `title` | string |  | Embed title |
| `webhook_url` | string |  | Webhook URL (uses DISCORD_WEBHOOK_URL env if not specified) |

### Example

```json
{
  "color": "example",
  "content": "example",
  "description": "example",
  "title": "example",
  "webhook_url": "example"
}
```

---

