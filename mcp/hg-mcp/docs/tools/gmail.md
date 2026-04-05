# gmail

> Gmail integration for email management, show bookings, and communications

**14 tools**

## Tools

- [`aftrs_gmail_inbox`](#aftrs_gmail_inbox)
- [`aftrs_gmail_unread`](#aftrs_gmail_unread)
- [`aftrs_gmail_search`](#aftrs_gmail_search)
- [`aftrs_gmail_get`](#aftrs_gmail_get)
- [`aftrs_gmail_thread`](#aftrs_gmail_thread)
- [`aftrs_gmail_labels`](#aftrs_gmail_labels)
- [`aftrs_gmail_starred`](#aftrs_gmail_starred)
- [`aftrs_gmail_sent`](#aftrs_gmail_sent)
- [`aftrs_gmail_send`](#aftrs_gmail_send)
- [`aftrs_gmail_draft`](#aftrs_gmail_draft)
- [`aftrs_gmail_mark_read`](#aftrs_gmail_mark_read)
- [`aftrs_gmail_star`](#aftrs_gmail_star)
- [`aftrs_gmail_archive`](#aftrs_gmail_archive)
- [`aftrs_gmail_trash`](#aftrs_gmail_trash)

---

## aftrs_gmail_inbox

Get recent inbox messages

**Complexity:** simple

**Tags:** `gmail`, `inbox`, `email`, `messages`

**Use Cases:**
- Check inbox
- View recent emails
- Morning briefing

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `limit` | integer |  | Maximum messages to return (default: 25, max: 100) |

---

## aftrs_gmail_unread

Get unread messages

**Complexity:** simple

**Tags:** `gmail`, `unread`, `email`, `new`

**Use Cases:**
- Check new emails
- Unread count
- Pending messages

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `limit` | integer |  | Maximum messages to return (default: 25, max: 100) |

---

## aftrs_gmail_search

Search Gmail messages using Gmail search syntax

**Complexity:** simple

**Tags:** `gmail`, `search`, `find`, `query`

**Use Cases:**
- Find booking emails
- Search by sender
- Find show confirmations

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `query` | string | Yes | Gmail search query (e.g., 'from:booking@venue.com', 'subject:gig', 'after:2024/01/01') |
| `limit` | integer |  | Maximum messages to return (default: 25, max: 100) |

---

## aftrs_gmail_get

Get a specific email message by ID

**Complexity:** simple

**Tags:** `gmail`, `get`, `read`, `message`

**Use Cases:**
- Read email
- View full message
- Get details

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `message_id` | string | Yes | Gmail message ID |
| `include_body` | boolean |  | Include full message body (default: true) |

---

## aftrs_gmail_thread

Get an email thread with all messages

**Complexity:** simple

**Tags:** `gmail`, `thread`, `conversation`, `history`

**Use Cases:**
- View conversation
- Follow email thread
- Get context

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `thread_id` | string | Yes | Gmail thread ID |

---

## aftrs_gmail_labels

List all Gmail labels

**Complexity:** simple

**Tags:** `gmail`, `labels`, `folders`, `organize`

**Use Cases:**
- View labels
- Get label IDs
- Organize emails

---

## aftrs_gmail_starred

Get starred messages

**Complexity:** simple

**Tags:** `gmail`, `starred`, `important`, `flagged`

**Use Cases:**
- View important emails
- Check starred
- Priority inbox

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `limit` | integer |  | Maximum messages to return (default: 25, max: 100) |

---

## aftrs_gmail_sent

Get sent messages

**Complexity:** simple

**Tags:** `gmail`, `sent`, `outgoing`, `history`

**Use Cases:**
- View sent emails
- Check sent messages
- Outbox

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `limit` | integer |  | Maximum messages to return (default: 25, max: 100) |

---

## aftrs_gmail_send

Send an email

**Complexity:** moderate

**Tags:** `gmail`, `send`, `compose`, `email`

**Use Cases:**
- Send email
- Reply to booking
- Contact venue

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `to` | array | Yes | Recipient email addresses |
| `cc` | array |  | CC email addresses |
| `subject` | string | Yes | Email subject |
| `body` | string | Yes | Email body content |
| `is_html` | boolean |  | Whether body is HTML (default: false) |

---

## aftrs_gmail_draft

Create an email draft

**Complexity:** moderate

**Tags:** `gmail`, `draft`, `compose`, `save`

**Use Cases:**
- Create draft
- Save for later
- Prepare email

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `to` | array | Yes | Recipient email addresses |
| `cc` | array |  | CC email addresses |
| `subject` | string | Yes | Email subject |
| `body` | string | Yes | Email body content |
| `is_html` | boolean |  | Whether body is HTML (default: false) |

---

## aftrs_gmail_mark_read

Mark a message as read

**Complexity:** simple

**Tags:** `gmail`, `read`, `mark`, `status`

**Use Cases:**
- Mark as read
- Clear notification
- Update status

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `message_id` | string | Yes | Gmail message ID |

---

## aftrs_gmail_star

Star or unstar a message

**Complexity:** simple

**Tags:** `gmail`, `star`, `flag`, `important`

**Use Cases:**
- Star email
- Mark important
- Flag for follow-up

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `message_id` | string | Yes | Gmail message ID |
| `star` | boolean |  | True to star, false to unstar (default: true) |

---

## aftrs_gmail_archive

Archive a message (remove from inbox)

**Complexity:** simple

**Tags:** `gmail`, `archive`, `organize`, `cleanup`

**Use Cases:**
- Archive email
- Clean inbox
- Organize

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `message_id` | string | Yes | Gmail message ID |

---

## aftrs_gmail_trash

Move a message to trash

**Complexity:** simple

**Tags:** `gmail`, `trash`, `delete`, `remove`

**Use Cases:**
- Delete email
- Trash message
- Remove

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `message_id` | string | Yes | Gmail message ID |

---

## Authentication

Requires Google OAuth credentials via:
- `GOOGLE_APPLICATION_CREDENTIALS` environment variable pointing to a service account JSON file
- Or `GOOGLE_API_KEY` for API key authentication

## Gmail Search Syntax

Common search operators:
- `from:email@example.com` - Messages from specific sender
- `to:email@example.com` - Messages to specific recipient
- `subject:keyword` - Search in subject line
- `after:2024/01/01` - Messages after date
- `before:2024/12/31` - Messages before date
- `is:unread` - Unread messages
- `is:starred` - Starred messages
- `has:attachment` - Messages with attachments
- `label:labelname` - Messages with specific label
