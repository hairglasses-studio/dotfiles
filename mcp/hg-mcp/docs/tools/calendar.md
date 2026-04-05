# calendar

> Google Calendar integration for show scheduling and event management

**12 tools**

## Tools

- [`aftrs_calendar_list`](#aftrs-calendar-list)
- [`aftrs_calendar_search`](#aftrs-calendar-search)
- [`aftrs_calendar_today`](#aftrs-calendar-today)
- [`aftrs_schedule_show`](#aftrs-schedule-show)
- [`aftrs_show_schedule`](#aftrs-show-schedule)
- [`aftrs_show_time_until`](#aftrs-show-time-until)
- [`aftrs_calendar_event_update`](#aftrs-calendar-event-update)
- [`aftrs_calendar_event_delete`](#aftrs-calendar-event-delete)
- [`aftrs_calendar_recurring_create`](#aftrs-calendar-recurring-create)
- [`aftrs_calendar_quick_add`](#aftrs-calendar-quick-add)
- [`aftrs_calendar_freebusy`](#aftrs-calendar-freebusy)
- [`aftrs_calendar_colors`](#aftrs-calendar-colors)

---

## aftrs_calendar_list

List all accessible Google Calendars

**Complexity:** simple

**Tags:** `calendar`, `google`, `list`

**Use Cases:**
- List calendars
- Find calendar IDs
- View access roles

---

## aftrs_calendar_search

Search for events matching a query

**Complexity:** simple

**Tags:** `calendar`, `search`, `find`, `events`

**Use Cases:**
- Find past shows
- Search venues
- Find recurring events

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `calendar_id` | string |  | Calendar ID (default: primary) |
| `limit` | integer |  | Maximum results (default: 25) |
| `query` | string | Yes | Search query (searches titles and descriptions) |

---

## aftrs_calendar_today

Get today's scheduled events

**Complexity:** simple

**Tags:** `calendar`, `today`, `schedule`, `daily`

**Use Cases:**
- Daily agenda
- Today's events
- Morning briefing

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `calendar_id` | string |  | Calendar ID (default: primary) |

---

## aftrs_schedule_show

Create a new show or event booking on the calendar

**Complexity:** moderate

**Tags:** `calendar`, `create`, `booking`, `schedule`

**Use Cases:**
- Book a show
- Create event
- Schedule gig

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `all_day` | boolean |  | Whether this is an all-day event |
| `calendar_id` | string |  | Calendar ID (default: primary) |
| `description` | string |  | Event description/notes |
| `end` | string | Yes | End time (ISO 8601 format) |
| `location` | string |  | Event location |
| `start` | string | Yes | Start time (ISO 8601 format, e.g., 2024-03-15T20:00:00) |
| `title` | string | Yes | Event title/summary |

---

## aftrs_show_schedule

List upcoming shows and events from the calendar

**Complexity:** simple

**Tags:** `calendar`, `shows`, `schedule`, `upcoming`

**Use Cases:**
- View upcoming shows
- Plan show prep
- Check availability

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `calendar_id` | string |  | Calendar ID (default: primary) |
| `days` | integer |  | Number of days to look ahead (default: 7) |

---

## aftrs_show_time_until

Get time remaining until the next scheduled show

**Complexity:** simple

**Tags:** `calendar`, `countdown`, `timer`, `next`

**Use Cases:**
- Countdown to show
- Prep timing
- Show readiness

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `calendar_id` | string |  | Calendar ID (default: primary) |

---

## aftrs_calendar_event_update

Update an existing calendar event

**Complexity:** moderate

**Tags:** `calendar`, `update`, `edit`, `modify`

**Use Cases:**
- Reschedule event
- Update show details
- Change venue

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `event_id` | string | Yes | Event ID to update |
| `calendar_id` | string |  | Calendar ID (default: primary) |
| `title` | string |  | New event title |
| `description` | string |  | New event description |
| `location` | string |  | New event location |
| `start` | string |  | New start time (ISO 8601) |
| `end` | string |  | New end time (ISO 8601) |
| `color_id` | string |  | Color ID for the event |

---

## aftrs_calendar_event_delete

Delete a calendar event

**Complexity:** moderate

**Tags:** `calendar`, `delete`, `cancel`, `remove`

**Use Cases:**
- Cancel show
- Remove event
- Clear booking

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `event_id` | string | Yes | Event ID to delete |
| `calendar_id` | string |  | Calendar ID (default: primary) |
| `notify_attendees` | boolean |  | Send cancellation emails to attendees (default: true) |

---

## aftrs_calendar_recurring_create

Create a recurring calendar event with RRULE pattern

**Complexity:** moderate

**Tags:** `calendar`, `recurring`, `repeat`, `schedule`

**Use Cases:**
- Weekly show
- Monthly rehearsal
- Regular booking

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `title` | string | Yes | Event title |
| `start` | string | Yes | First occurrence start time (ISO 8601) |
| `end` | string | Yes | First occurrence end time (ISO 8601) |
| `recurrence` | string | Yes | RRULE pattern (e.g., FREQ=WEEKLY;BYDAY=FR or FREQ=MONTHLY;BYMONTHDAY=1) |
| `description` | string |  | Event description |
| `location` | string |  | Event location |
| `calendar_id` | string |  | Calendar ID (default: primary) |

---

## aftrs_calendar_quick_add

Create event using natural language (e.g., 'DJ Set at Club Friday 9pm-2am')

**Complexity:** simple

**Tags:** `calendar`, `quick`, `natural`, `create`

**Use Cases:**
- Fast event creation
- Voice-style input
- Quick booking

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `text` | string | Yes | Natural language event description |
| `calendar_id` | string |  | Calendar ID (default: primary) |

---

## aftrs_calendar_freebusy

Check free/busy availability across calendars

**Complexity:** simple

**Tags:** `calendar`, `freebusy`, `availability`, `schedule`

**Use Cases:**
- Check availability
- Find open slots
- Schedule coordination

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `calendar_ids` | array |  | Calendar IDs to check (default: primary) |
| `start` | string | Yes | Start of time range (ISO 8601) |
| `end` | string | Yes | End of time range (ISO 8601) |

---

## aftrs_calendar_colors

Get available calendar event colors

**Complexity:** simple

**Tags:** `calendar`, `colors`, `styling`

**Use Cases:**
- Color-code events
- Visual organization
- Event styling

---
