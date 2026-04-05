// Package clients provides client implementations for external services.
package clients

import (
	"context"
	"fmt"
	"sync"
	"time"

	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"

	"github.com/hairglasses-studio/hg-mcp/internal/config"
)

// CalendarClient provides Google Calendar operations
type CalendarClient struct {
	service *calendar.Service
	mu      sync.RWMutex
}

// CalendarEvent represents a calendar event
type CalendarEvent struct {
	ID           string          `json:"id"`
	Summary      string          `json:"summary"`
	Description  string          `json:"description,omitempty"`
	Location     string          `json:"location,omitempty"`
	Start        time.Time       `json:"start"`
	End          time.Time       `json:"end"`
	AllDay       bool            `json:"all_day"`
	Status       string          `json:"status"`
	HTMLLink     string          `json:"html_link,omitempty"`
	Attendees    []string        `json:"attendees,omitempty"`
	CalendarID   string          `json:"calendar_id"`
	Recurrence   []string        `json:"recurrence,omitempty"`
	RecurringID  string          `json:"recurring_event_id,omitempty"`
	Reminders    []EventReminder `json:"reminders,omitempty"`
	ColorID      string          `json:"color_id,omitempty"`
	Visibility   string          `json:"visibility,omitempty"`
	Transparency string          `json:"transparency,omitempty"`
}

// EventReminder represents a reminder for an event
type EventReminder struct {
	Method  string `json:"method"` // email or popup
	Minutes int    `json:"minutes"`
}

// CalendarInfo represents calendar metadata
type CalendarInfo struct {
	ID          string `json:"id"`
	Summary     string `json:"summary"`
	Description string `json:"description,omitempty"`
	TimeZone    string `json:"timezone"`
	Primary     bool   `json:"primary"`
	AccessRole  string `json:"access_role"`
}

// TimeUntilShow represents time until next show
type TimeUntilShow struct {
	Event     *CalendarEvent `json:"event,omitempty"`
	Days      int            `json:"days"`
	Hours     int            `json:"hours"`
	Minutes   int            `json:"minutes"`
	TotalMins int            `json:"total_minutes"`
	Message   string         `json:"message"`
}

var (
	calendarClient     *CalendarClient
	calendarClientOnce sync.Once
	calendarClientErr  error

	// TestOverrideCalendarClient, when non-nil, is returned by GetCalendarClient
	// instead of the real singleton. Used for testing without network access.
	TestOverrideCalendarClient *CalendarClient
)

// GetCalendarClient returns the singleton Calendar client
func GetCalendarClient() (*CalendarClient, error) {
	if TestOverrideCalendarClient != nil {
		return TestOverrideCalendarClient, nil
	}
	calendarClientOnce.Do(func() {
		calendarClient, calendarClientErr = NewCalendarClient()
	})
	return calendarClient, calendarClientErr
}

// NewTestCalendarClient creates an in-memory test client.
func NewTestCalendarClient() *CalendarClient {
	return &CalendarClient{}
}

// NewCalendarClient creates a new Google Calendar client
func NewCalendarClient() (*CalendarClient, error) {
	ctx := context.Background()

	var opts []option.ClientOption

	cfg := config.Get()
	if cfg.GoogleApplicationCredentials != "" {
		opts = append(opts, option.WithCredentialsFile(cfg.GoogleApplicationCredentials))
	} else if cfg.GoogleAPIKey != "" {
		opts = append(opts, option.WithAPIKey(cfg.GoogleAPIKey))
	} else {
		return nil, fmt.Errorf("no Google credentials configured (set GOOGLE_APPLICATION_CREDENTIALS or GOOGLE_API_KEY)")
	}

	service, err := calendar.NewService(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create Calendar service: %w", err)
	}

	return &CalendarClient{service: service}, nil
}

// IsConfigured returns true if the client is properly configured
func (c *CalendarClient) IsConfigured() bool {
	return c != nil && c.service != nil
}

// ListCalendars lists all accessible calendars
func (c *CalendarClient) ListCalendars(ctx context.Context) ([]CalendarInfo, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	list, err := c.service.CalendarList.List().Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to list calendars: %w", err)
	}

	calendars := make([]CalendarInfo, 0, len(list.Items))
	for _, cal := range list.Items {
		calendars = append(calendars, CalendarInfo{
			ID:          cal.Id,
			Summary:     cal.Summary,
			Description: cal.Description,
			TimeZone:    cal.TimeZone,
			Primary:     cal.Primary,
			AccessRole:  cal.AccessRole,
		})
	}

	return calendars, nil
}

// ListEvents lists events from a calendar within a time range
func (c *CalendarClient) ListEvents(ctx context.Context, calendarID string, timeMin, timeMax time.Time, maxResults int64) ([]CalendarEvent, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if calendarID == "" {
		calendarID = "primary"
	}
	if maxResults <= 0 {
		maxResults = 50
	}

	call := c.service.Events.List(calendarID).
		Context(ctx).
		ShowDeleted(false).
		SingleEvents(true).
		OrderBy("startTime").
		MaxResults(maxResults)

	if !timeMin.IsZero() {
		call = call.TimeMin(timeMin.Format(time.RFC3339))
	}
	if !timeMax.IsZero() {
		call = call.TimeMax(timeMax.Format(time.RFC3339))
	}

	events, err := call.Do()
	if err != nil {
		return nil, fmt.Errorf("failed to list events: %w", err)
	}

	result := make([]CalendarEvent, 0, len(events.Items))
	for _, e := range events.Items {
		result = append(result, convertCalendarEvent(e, calendarID))
	}

	return result, nil
}

// GetUpcomingEvents returns events happening in the next N days
func (c *CalendarClient) GetUpcomingEvents(ctx context.Context, calendarID string, days int) ([]CalendarEvent, error) {
	if days <= 0 {
		days = 7
	}
	now := time.Now()
	timeMax := now.AddDate(0, 0, days)
	return c.ListEvents(ctx, calendarID, now, timeMax, 100)
}

// GetTodayEvents returns today's events
func (c *CalendarClient) GetTodayEvents(ctx context.Context, calendarID string) ([]CalendarEvent, error) {
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfDay := startOfDay.AddDate(0, 0, 1)
	return c.ListEvents(ctx, calendarID, startOfDay, endOfDay, 50)
}

// GetEvent gets a specific event by ID
func (c *CalendarClient) GetEvent(ctx context.Context, calendarID, eventID string) (*CalendarEvent, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if calendarID == "" {
		calendarID = "primary"
	}

	event, err := c.service.Events.Get(calendarID, eventID).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get event: %w", err)
	}

	result := convertCalendarEvent(event, calendarID)
	return &result, nil
}

// CreateEvent creates a new calendar event
func (c *CalendarClient) CreateEvent(ctx context.Context, calendarID string, event *CalendarEvent) (*CalendarEvent, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if calendarID == "" {
		calendarID = "primary"
	}

	newEvent := &calendar.Event{
		Summary:     event.Summary,
		Description: event.Description,
		Location:    event.Location,
	}

	if event.AllDay {
		newEvent.Start = &calendar.EventDateTime{Date: event.Start.Format("2006-01-02")}
		newEvent.End = &calendar.EventDateTime{Date: event.End.Format("2006-01-02")}
	} else {
		newEvent.Start = &calendar.EventDateTime{DateTime: event.Start.Format(time.RFC3339)}
		newEvent.End = &calendar.EventDateTime{DateTime: event.End.Format(time.RFC3339)}
	}

	// Add attendees
	for _, email := range event.Attendees {
		newEvent.Attendees = append(newEvent.Attendees, &calendar.EventAttendee{Email: email})
	}

	created, err := c.service.Events.Insert(calendarID, newEvent).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to create event: %w", err)
	}

	result := convertCalendarEvent(created, calendarID)
	return &result, nil
}

// TimeUntilNextShow calculates time until the next show
func (c *CalendarClient) TimeUntilNextShow(ctx context.Context, calendarID string) (*TimeUntilShow, error) {
	events, err := c.GetUpcomingEvents(ctx, calendarID, 30)
	if err != nil {
		return nil, err
	}

	if len(events) == 0 {
		return &TimeUntilShow{
			Message: "No upcoming shows scheduled in the next 30 days",
		}, nil
	}

	nextEvent := events[0]
	now := time.Now()
	duration := nextEvent.Start.Sub(now)

	totalMins := int(duration.Minutes())
	days := int(duration.Hours() / 24)
	hours := int(duration.Hours()) % 24
	minutes := int(duration.Minutes()) % 60

	var message string
	if days > 0 {
		message = fmt.Sprintf("%d days, %d hours, %d minutes until %s", days, hours, minutes, nextEvent.Summary)
	} else if hours > 0 {
		message = fmt.Sprintf("%d hours, %d minutes until %s", hours, minutes, nextEvent.Summary)
	} else {
		message = fmt.Sprintf("%d minutes until %s", minutes, nextEvent.Summary)
	}

	return &TimeUntilShow{
		Event:     &nextEvent,
		Days:      days,
		Hours:     hours,
		Minutes:   minutes,
		TotalMins: totalMins,
		Message:   message,
	}, nil
}

// SearchEvents searches for events matching a query
func (c *CalendarClient) SearchEvents(ctx context.Context, calendarID, query string, maxResults int64) ([]CalendarEvent, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if calendarID == "" {
		calendarID = "primary"
	}
	if maxResults <= 0 {
		maxResults = 25
	}

	// Search within next year
	now := time.Now()
	timeMax := now.AddDate(1, 0, 0)

	events, err := c.service.Events.List(calendarID).
		Context(ctx).
		Q(query).
		ShowDeleted(false).
		SingleEvents(true).
		TimeMin(now.Format(time.RFC3339)).
		TimeMax(timeMax.Format(time.RFC3339)).
		OrderBy("startTime").
		MaxResults(maxResults).
		Do()
	if err != nil {
		return nil, fmt.Errorf("failed to search events: %w", err)
	}

	result := make([]CalendarEvent, 0, len(events.Items))
	for _, e := range events.Items {
		result = append(result, convertCalendarEvent(e, calendarID))
	}

	return result, nil
}

// UpdateEvent updates an existing calendar event
func (c *CalendarClient) UpdateEvent(ctx context.Context, calendarID, eventID string, event *CalendarEvent) (*CalendarEvent, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if calendarID == "" {
		calendarID = "primary"
	}

	// Get existing event first
	existing, err := c.service.Events.Get(calendarID, eventID).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get event: %w", err)
	}

	// Update fields if provided
	if event.Summary != "" {
		existing.Summary = event.Summary
	}
	if event.Description != "" {
		existing.Description = event.Description
	}
	if event.Location != "" {
		existing.Location = event.Location
	}
	if !event.Start.IsZero() {
		if event.AllDay {
			existing.Start = &calendar.EventDateTime{Date: event.Start.Format("2006-01-02")}
		} else {
			existing.Start = &calendar.EventDateTime{DateTime: event.Start.Format(time.RFC3339)}
		}
	}
	if !event.End.IsZero() {
		if event.AllDay {
			existing.End = &calendar.EventDateTime{Date: event.End.Format("2006-01-02")}
		} else {
			existing.End = &calendar.EventDateTime{DateTime: event.End.Format(time.RFC3339)}
		}
	}
	if event.ColorID != "" {
		existing.ColorId = event.ColorID
	}
	if event.Visibility != "" {
		existing.Visibility = event.Visibility
	}

	// Update attendees if provided
	if len(event.Attendees) > 0 {
		existing.Attendees = nil
		for _, email := range event.Attendees {
			existing.Attendees = append(existing.Attendees, &calendar.EventAttendee{Email: email})
		}
	}

	updated, err := c.service.Events.Update(calendarID, eventID, existing).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to update event: %w", err)
	}

	result := convertCalendarEvent(updated, calendarID)
	return &result, nil
}

// DeleteEvent deletes a calendar event
func (c *CalendarClient) DeleteEvent(ctx context.Context, calendarID, eventID string, sendUpdates bool) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if calendarID == "" {
		calendarID = "primary"
	}

	call := c.service.Events.Delete(calendarID, eventID).Context(ctx)
	if sendUpdates {
		call = call.SendUpdates("all")
	}

	if err := call.Do(); err != nil {
		return fmt.Errorf("failed to delete event: %w", err)
	}
	return nil
}

// CreateRecurringEvent creates a recurring calendar event
func (c *CalendarClient) CreateRecurringEvent(ctx context.Context, calendarID string, event *CalendarEvent, recurrence string) (*CalendarEvent, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if calendarID == "" {
		calendarID = "primary"
	}

	newEvent := &calendar.Event{
		Summary:     event.Summary,
		Description: event.Description,
		Location:    event.Location,
		Recurrence:  []string{recurrence},
	}

	if event.AllDay {
		newEvent.Start = &calendar.EventDateTime{Date: event.Start.Format("2006-01-02")}
		newEvent.End = &calendar.EventDateTime{Date: event.End.Format("2006-01-02")}
	} else {
		newEvent.Start = &calendar.EventDateTime{DateTime: event.Start.Format(time.RFC3339)}
		newEvent.End = &calendar.EventDateTime{DateTime: event.End.Format(time.RFC3339)}
	}

	// Add attendees
	for _, email := range event.Attendees {
		newEvent.Attendees = append(newEvent.Attendees, &calendar.EventAttendee{Email: email})
	}

	// Add reminders if provided
	if len(event.Reminders) > 0 {
		reminders := make([]*calendar.EventReminder, 0, len(event.Reminders))
		for _, r := range event.Reminders {
			reminders = append(reminders, &calendar.EventReminder{
				Method:  r.Method,
				Minutes: int64(r.Minutes),
			})
		}
		newEvent.Reminders = &calendar.EventReminders{
			UseDefault: false,
			Overrides:  reminders,
		}
	}

	if event.ColorID != "" {
		newEvent.ColorId = event.ColorID
	}
	if event.Visibility != "" {
		newEvent.Visibility = event.Visibility
	}

	created, err := c.service.Events.Insert(calendarID, newEvent).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to create recurring event: %w", err)
	}

	result := convertCalendarEvent(created, calendarID)
	return &result, nil
}

// QuickAddEvent creates an event from a natural language text
func (c *CalendarClient) QuickAddEvent(ctx context.Context, calendarID, text string) (*CalendarEvent, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if calendarID == "" {
		calendarID = "primary"
	}

	event, err := c.service.Events.QuickAdd(calendarID, text).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to quick add event: %w", err)
	}

	result := convertCalendarEvent(event, calendarID)
	return &result, nil
}

// GetFreebusy returns free/busy information for calendars
func (c *CalendarClient) GetFreebusy(ctx context.Context, calendarIDs []string, timeMin, timeMax time.Time) (map[string][]FreebusyPeriod, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	items := make([]*calendar.FreeBusyRequestItem, 0, len(calendarIDs))
	for _, id := range calendarIDs {
		items = append(items, &calendar.FreeBusyRequestItem{Id: id})
	}

	resp, err := c.service.Freebusy.Query(&calendar.FreeBusyRequest{
		TimeMin: timeMin.Format(time.RFC3339),
		TimeMax: timeMax.Format(time.RFC3339),
		Items:   items,
	}).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get freebusy: %w", err)
	}

	result := make(map[string][]FreebusyPeriod)
	for calID, fb := range resp.Calendars {
		periods := make([]FreebusyPeriod, 0, len(fb.Busy))
		for _, b := range fb.Busy {
			start, _ := time.Parse(time.RFC3339, b.Start)
			end, _ := time.Parse(time.RFC3339, b.End)
			periods = append(periods, FreebusyPeriod{Start: start, End: end})
		}
		result[calID] = periods
	}

	return result, nil
}

// FreebusyPeriod represents a busy period
type FreebusyPeriod struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// GetColors returns available calendar colors
func (c *CalendarClient) GetColors(ctx context.Context) (map[string]string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	colors, err := c.service.Colors.Get().Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get colors: %w", err)
	}

	result := make(map[string]string)
	for id, color := range colors.Event {
		result[id] = color.Background
	}

	return result, nil
}

func convertCalendarEvent(e *calendar.Event, calendarID string) CalendarEvent {
	event := CalendarEvent{
		ID:           e.Id,
		Summary:      e.Summary,
		Description:  e.Description,
		Location:     e.Location,
		Status:       e.Status,
		HTMLLink:     e.HtmlLink,
		CalendarID:   calendarID,
		Recurrence:   e.Recurrence,
		RecurringID:  e.RecurringEventId,
		ColorID:      e.ColorId,
		Visibility:   e.Visibility,
		Transparency: e.Transparency,
	}

	// Parse start time
	if e.Start != nil {
		if e.Start.DateTime != "" {
			if t, err := time.Parse(time.RFC3339, e.Start.DateTime); err == nil {
				event.Start = t
			}
		} else if e.Start.Date != "" {
			if t, err := time.Parse("2006-01-02", e.Start.Date); err == nil {
				event.Start = t
				event.AllDay = true
			}
		}
	}

	// Parse end time
	if e.End != nil {
		if e.End.DateTime != "" {
			if t, err := time.Parse(time.RFC3339, e.End.DateTime); err == nil {
				event.End = t
			}
		} else if e.End.Date != "" {
			if t, err := time.Parse("2006-01-02", e.End.Date); err == nil {
				event.End = t
			}
		}
	}

	// Collect attendees
	for _, a := range e.Attendees {
		event.Attendees = append(event.Attendees, a.Email)
	}

	return event
}
