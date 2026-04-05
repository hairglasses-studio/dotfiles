// Package calendar provides MCP tools for Google Calendar integration.
package calendar

import (
	"context"
	"fmt"
	"time"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
	"github.com/mark3labs/mcp-go/mcp"
)

// Module implements the Calendar tools module
type Module struct{}

func init() {
	tools.GetRegistry().RegisterModule(&Module{})
}

// Name returns the module name
func (m *Module) Name() string {
	return "calendar"
}

// Description returns the module description
func (m *Module) Description() string {
	return "Google Calendar integration for show scheduling and event management"
}

// Tools returns the Calendar tool definitions
func (m *Module) Tools() []tools.ToolDefinition {
	return []tools.ToolDefinition{
		{
			Tool: mcp.Tool{
				Name:        "aftrs_calendar_list",
				Description: "List all accessible Google Calendars",
				InputSchema: mcp.ToolInputSchema{
					Type:       "object",
					Properties: map[string]interface{}{},
				},
			},
			Handler:             handleCalendarList,
			Category:            "calendar",
			Subcategory:         "management",
			Tags:                []string{"calendar", "google", "list"},
			UseCases:            []string{"List calendars", "Find calendar IDs", "View access roles"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "google",
		},
		{
			Tool: mcp.Tool{
				Name:        "aftrs_show_schedule",
				Description: "List upcoming shows and events from the calendar",
				InputSchema: mcp.ToolInputSchema{
					Type: "object",
					Properties: map[string]interface{}{
						"calendar_id": map[string]interface{}{
							"type":        "string",
							"description": "Calendar ID (default: primary)",
						},
						"days": map[string]interface{}{
							"type":        "integer",
							"description": "Number of days to look ahead (default: 7)",
							"minimum":     1,
							"maximum":     90,
						},
					},
				},
			},
			Handler:             handleShowSchedule,
			Category:            "calendar",
			Subcategory:         "shows",
			Tags:                []string{"calendar", "shows", "schedule", "upcoming"},
			UseCases:            []string{"View upcoming shows", "Plan show prep", "Check availability"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "google",
		},
		{
			Tool: mcp.Tool{
				Name:        "aftrs_schedule_show",
				Description: "Create a new show or event booking on the calendar",
				InputSchema: mcp.ToolInputSchema{
					Type: "object",
					Properties: map[string]interface{}{
						"title": map[string]interface{}{
							"type":        "string",
							"description": "Event title/summary",
						},
						"start": map[string]interface{}{
							"type":        "string",
							"description": "Start time (ISO 8601 format, e.g., 2024-03-15T20:00:00)",
						},
						"end": map[string]interface{}{
							"type":        "string",
							"description": "End time (ISO 8601 format)",
						},
						"description": map[string]interface{}{
							"type":        "string",
							"description": "Event description/notes",
						},
						"location": map[string]interface{}{
							"type":        "string",
							"description": "Event location",
						},
						"calendar_id": map[string]interface{}{
							"type":        "string",
							"description": "Calendar ID (default: primary)",
						},
						"all_day": map[string]interface{}{
							"type":        "boolean",
							"description": "Whether this is an all-day event",
						},
					},
					Required: []string{"title", "start", "end"},
				},
			},
			Handler:             handleScheduleShow,
			Category:            "calendar",
			Subcategory:         "shows",
			Tags:                []string{"calendar", "create", "booking", "schedule"},
			UseCases:            []string{"Book a show", "Create event", "Schedule gig"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "google",
		},
		{
			Tool: mcp.Tool{
				Name:        "aftrs_show_time_until",
				Description: "Get time remaining until the next scheduled show",
				InputSchema: mcp.ToolInputSchema{
					Type: "object",
					Properties: map[string]interface{}{
						"calendar_id": map[string]interface{}{
							"type":        "string",
							"description": "Calendar ID (default: primary)",
						},
					},
				},
			},
			Handler:             handleShowTimeUntil,
			Category:            "calendar",
			Subcategory:         "shows",
			Tags:                []string{"calendar", "countdown", "timer", "next"},
			UseCases:            []string{"Countdown to show", "Prep timing", "Show readiness"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "google",
		},
		{
			Tool: mcp.Tool{
				Name:        "aftrs_calendar_today",
				Description: "Get today's scheduled events",
				InputSchema: mcp.ToolInputSchema{
					Type: "object",
					Properties: map[string]interface{}{
						"calendar_id": map[string]interface{}{
							"type":        "string",
							"description": "Calendar ID (default: primary)",
						},
					},
				},
			},
			Handler:             handleCalendarToday,
			Category:            "calendar",
			Subcategory:         "events",
			Tags:                []string{"calendar", "today", "schedule", "daily"},
			UseCases:            []string{"Daily agenda", "Today's events", "Morning briefing"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "google",
		},
		{
			Tool: mcp.Tool{
				Name:        "aftrs_calendar_search",
				Description: "Search for events matching a query",
				InputSchema: mcp.ToolInputSchema{
					Type: "object",
					Properties: map[string]interface{}{
						"query": map[string]interface{}{
							"type":        "string",
							"description": "Search query (searches titles and descriptions)",
						},
						"calendar_id": map[string]interface{}{
							"type":        "string",
							"description": "Calendar ID (default: primary)",
						},
						"limit": map[string]interface{}{
							"type":        "integer",
							"description": "Maximum results (default: 25)",
							"minimum":     1,
							"maximum":     100,
						},
					},
					Required: []string{"query"},
				},
			},
			Handler:             handleCalendarSearch,
			Category:            "calendar",
			Subcategory:         "events",
			Tags:                []string{"calendar", "search", "find", "events"},
			UseCases:            []string{"Find past shows", "Search venues", "Find recurring events"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "google",
		},
		{
			Tool: mcp.Tool{
				Name:        "aftrs_calendar_event_update",
				Description: "Update an existing calendar event",
				InputSchema: mcp.ToolInputSchema{
					Type: "object",
					Properties: map[string]interface{}{
						"event_id": map[string]interface{}{
							"type":        "string",
							"description": "Event ID to update",
						},
						"calendar_id": map[string]interface{}{
							"type":        "string",
							"description": "Calendar ID (default: primary)",
						},
						"title": map[string]interface{}{
							"type":        "string",
							"description": "New event title",
						},
						"description": map[string]interface{}{
							"type":        "string",
							"description": "New event description",
						},
						"location": map[string]interface{}{
							"type":        "string",
							"description": "New event location",
						},
						"start": map[string]interface{}{
							"type":        "string",
							"description": "New start time (ISO 8601)",
						},
						"end": map[string]interface{}{
							"type":        "string",
							"description": "New end time (ISO 8601)",
						},
						"color_id": map[string]interface{}{
							"type":        "string",
							"description": "Color ID for the event",
						},
					},
					Required: []string{"event_id"},
				},
			},
			Handler:             handleEventUpdate,
			Category:            "calendar",
			Subcategory:         "events",
			Tags:                []string{"calendar", "update", "edit", "modify"},
			UseCases:            []string{"Reschedule event", "Update show details", "Change venue"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "google",
		},
		{
			Tool: mcp.Tool{
				Name:        "aftrs_calendar_event_delete",
				Description: "Delete a calendar event",
				InputSchema: mcp.ToolInputSchema{
					Type: "object",
					Properties: map[string]interface{}{
						"event_id": map[string]interface{}{
							"type":        "string",
							"description": "Event ID to delete",
						},
						"calendar_id": map[string]interface{}{
							"type":        "string",
							"description": "Calendar ID (default: primary)",
						},
						"notify_attendees": map[string]interface{}{
							"type":        "boolean",
							"description": "Send cancellation emails to attendees (default: true)",
						},
					},
					Required: []string{"event_id"},
				},
			},
			Handler:             handleEventDelete,
			Category:            "calendar",
			Subcategory:         "events",
			Tags:                []string{"calendar", "delete", "cancel", "remove"},
			UseCases:            []string{"Cancel show", "Remove event", "Clear booking"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "google",
		},
		{
			Tool: mcp.Tool{
				Name:        "aftrs_calendar_recurring_create",
				Description: "Create a recurring calendar event with RRULE pattern",
				InputSchema: mcp.ToolInputSchema{
					Type: "object",
					Properties: map[string]interface{}{
						"title": map[string]interface{}{
							"type":        "string",
							"description": "Event title",
						},
						"start": map[string]interface{}{
							"type":        "string",
							"description": "First occurrence start time (ISO 8601)",
						},
						"end": map[string]interface{}{
							"type":        "string",
							"description": "First occurrence end time (ISO 8601)",
						},
						"recurrence": map[string]interface{}{
							"type":        "string",
							"description": "RRULE pattern (e.g., FREQ=WEEKLY;BYDAY=FR or FREQ=MONTHLY;BYMONTHDAY=1)",
						},
						"description": map[string]interface{}{
							"type":        "string",
							"description": "Event description",
						},
						"location": map[string]interface{}{
							"type":        "string",
							"description": "Event location",
						},
						"calendar_id": map[string]interface{}{
							"type":        "string",
							"description": "Calendar ID (default: primary)",
						},
					},
					Required: []string{"title", "start", "end", "recurrence"},
				},
			},
			Handler:             handleRecurringCreate,
			Category:            "calendar",
			Subcategory:         "events",
			Tags:                []string{"calendar", "recurring", "repeat", "schedule"},
			UseCases:            []string{"Weekly show", "Monthly rehearsal", "Regular booking"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "google",
		},
		{
			Tool: mcp.Tool{
				Name:        "aftrs_calendar_quick_add",
				Description: "Create event using natural language (e.g., 'DJ Set at Club Friday 9pm-2am')",
				InputSchema: mcp.ToolInputSchema{
					Type: "object",
					Properties: map[string]interface{}{
						"text": map[string]interface{}{
							"type":        "string",
							"description": "Natural language event description",
						},
						"calendar_id": map[string]interface{}{
							"type":        "string",
							"description": "Calendar ID (default: primary)",
						},
					},
					Required: []string{"text"},
				},
			},
			Handler:             handleQuickAdd,
			Category:            "calendar",
			Subcategory:         "events",
			Tags:                []string{"calendar", "quick", "natural", "create"},
			UseCases:            []string{"Fast event creation", "Voice-style input", "Quick booking"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "google",
		},
		{
			Tool: mcp.Tool{
				Name:        "aftrs_calendar_freebusy",
				Description: "Check free/busy availability across calendars",
				InputSchema: mcp.ToolInputSchema{
					Type: "object",
					Properties: map[string]interface{}{
						"calendar_ids": map[string]interface{}{
							"type":        "array",
							"description": "Calendar IDs to check (default: primary)",
							"items":       map[string]interface{}{"type": "string"},
						},
						"start": map[string]interface{}{
							"type":        "string",
							"description": "Start of time range (ISO 8601)",
						},
						"end": map[string]interface{}{
							"type":        "string",
							"description": "End of time range (ISO 8601)",
						},
					},
					Required: []string{"start", "end"},
				},
			},
			Handler:             handleFreebusy,
			Category:            "calendar",
			Subcategory:         "availability",
			Tags:                []string{"calendar", "freebusy", "availability", "schedule"},
			UseCases:            []string{"Check availability", "Find open slots", "Schedule coordination"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "google",
		},
		{
			Tool: mcp.Tool{
				Name:        "aftrs_calendar_colors",
				Description: "Get available calendar event colors",
				InputSchema: mcp.ToolInputSchema{
					Type:       "object",
					Properties: map[string]interface{}{},
				},
			},
			Handler:             handleCalendarColors,
			Category:            "calendar",
			Subcategory:         "management",
			Tags:                []string{"calendar", "colors", "styling"},
			UseCases:            []string{"Color-code events", "Visual organization", "Event styling"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "google",
		},
	}
}

var getCalendarClient = tools.LazyClient(clients.GetCalendarClient)

func handleCalendarList(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getCalendarClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Calendar client: %w", err)), nil
	}

	calendars, err := clients.GetListCalendarsCached(ctx, client)
	if err != nil {
		return tools.CodedErrorResult(tools.ErrAPIError, fmt.Errorf("failed to list calendars: %w", err)), nil
	}

	result := map[string]interface{}{
		"calendars": calendars,
		"count":     len(calendars),
	}
	return tools.JSONResult(result), nil
}

func handleShowSchedule(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getCalendarClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Calendar client: %w", err)), nil
	}

	calendarID := tools.GetStringParam(req, "calendar_id")
	days := tools.GetIntParam(req, "days", 7)

	events, err := client.GetUpcomingEvents(ctx, calendarID, days)
	if err != nil {
		return tools.CodedErrorResult(tools.ErrAPIError, fmt.Errorf("failed to get events: %w", err)), nil
	}

	result := map[string]interface{}{
		"events":      events,
		"count":       len(events),
		"days_ahead":  days,
		"calendar_id": calendarID,
	}
	if calendarID == "" {
		result["calendar_id"] = "primary"
	}

	return tools.JSONResult(result), nil
}

func handleScheduleShow(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getCalendarClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Calendar client: %w", err)), nil
	}

	title, errResult := tools.RequireStringParam(req, "title")
	if errResult != nil {
		return errResult, nil
	}

	startStr := tools.GetStringParam(req, "start")
	endStr := tools.GetStringParam(req, "end")

	start, err := time.Parse(time.RFC3339, startStr)
	if err != nil {
		// Try parsing without timezone
		start, err = time.Parse("2006-01-02T15:04:05", startStr)
		if err != nil {
			return tools.CodedErrorResult(tools.ErrInvalidParam, fmt.Errorf("invalid start time format (use ISO 8601): %w", err)), nil
		}
	}

	end, err := time.Parse(time.RFC3339, endStr)
	if err != nil {
		end, err = time.Parse("2006-01-02T15:04:05", endStr)
		if err != nil {
			return tools.CodedErrorResult(tools.ErrInvalidParam, fmt.Errorf("invalid end time format (use ISO 8601): %w", err)), nil
		}
	}

	calendarID := tools.GetStringParam(req, "calendar_id")
	allDay := tools.GetBoolParam(req, "all_day", false)

	event := &clients.CalendarEvent{
		Summary:     title,
		Description: tools.GetStringParam(req, "description"),
		Location:    tools.GetStringParam(req, "location"),
		Start:       start,
		End:         end,
		AllDay:      allDay,
	}

	created, err := client.CreateEvent(ctx, calendarID, event)
	if err != nil {
		return tools.CodedErrorResult(tools.ErrAPIError, fmt.Errorf("failed to create event: %w", err)), nil
	}

	result := map[string]interface{}{
		"event":   created,
		"message": fmt.Sprintf("Successfully created event: %s", created.Summary),
	}
	return tools.JSONResult(result), nil
}

func handleShowTimeUntil(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getCalendarClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Calendar client: %w", err)), nil
	}

	calendarID := tools.GetStringParam(req, "calendar_id")

	timeUntil, err := client.TimeUntilNextShow(ctx, calendarID)
	if err != nil {
		return tools.CodedErrorResult(tools.ErrAPIError, fmt.Errorf("failed to get time until show: %w", err)), nil
	}

	return tools.JSONResult(timeUntil), nil
}

func handleCalendarToday(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getCalendarClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Calendar client: %w", err)), nil
	}

	calendarID := tools.GetStringParam(req, "calendar_id")

	events, err := client.GetTodayEvents(ctx, calendarID)
	if err != nil {
		return tools.CodedErrorResult(tools.ErrAPIError, fmt.Errorf("failed to get today's events: %w", err)), nil
	}

	result := map[string]interface{}{
		"events": events,
		"count":  len(events),
		"date":   time.Now().Format("2006-01-02"),
	}
	return tools.JSONResult(result), nil
}

func handleCalendarSearch(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getCalendarClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Calendar client: %w", err)), nil
	}

	query, errResult := tools.RequireStringParam(req, "query")
	if errResult != nil {
		return errResult, nil
	}

	calendarID := tools.GetStringParam(req, "calendar_id")
	limit := tools.GetIntParam(req, "limit", 25)

	events, err := client.SearchEvents(ctx, calendarID, query, int64(limit))
	if err != nil {
		return tools.CodedErrorResult(tools.ErrAPIError, fmt.Errorf("failed to search events: %w", err)), nil
	}

	result := map[string]interface{}{
		"events": events,
		"count":  len(events),
		"query":  query,
	}
	return tools.JSONResult(result), nil
}

func handleEventUpdate(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getCalendarClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Calendar client: %w", err)), nil
	}

	eventID, errResult := tools.RequireStringParam(req, "event_id")
	if errResult != nil {
		return errResult, nil
	}

	calendarID := tools.GetStringParam(req, "calendar_id")

	event := &clients.CalendarEvent{
		ID:          eventID,
		Summary:     tools.GetStringParam(req, "title"),
		Description: tools.GetStringParam(req, "description"),
		Location:    tools.GetStringParam(req, "location"),
		ColorID:     tools.GetStringParam(req, "color_id"),
	}

	if startStr := tools.GetStringParam(req, "start"); startStr != "" {
		if start, err := parseTime(startStr); err == nil {
			event.Start = start
		}
	}
	if endStr := tools.GetStringParam(req, "end"); endStr != "" {
		if end, err := parseTime(endStr); err == nil {
			event.End = end
		}
	}

	updated, err := client.UpdateEvent(ctx, calendarID, eventID, event)
	if err != nil {
		return tools.CodedErrorResult(tools.ErrAPIError, fmt.Errorf("failed to update event: %w", err)), nil
	}

	result := map[string]interface{}{
		"event":   updated,
		"message": fmt.Sprintf("Successfully updated event: %s", updated.Summary),
	}
	return tools.JSONResult(result), nil
}

func handleEventDelete(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getCalendarClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Calendar client: %w", err)), nil
	}

	eventID, errResult := tools.RequireStringParam(req, "event_id")
	if errResult != nil {
		return errResult, nil
	}

	calendarID := tools.GetStringParam(req, "calendar_id")
	notifyAttendees := tools.GetBoolParam(req, "notify_attendees", true)

	err = client.DeleteEvent(ctx, calendarID, eventID, notifyAttendees)
	if err != nil {
		return tools.CodedErrorResult(tools.ErrAPIError, fmt.Errorf("failed to delete event: %w", err)), nil
	}

	result := map[string]interface{}{
		"message":  "Event successfully deleted",
		"event_id": eventID,
	}
	return tools.JSONResult(result), nil
}

func handleRecurringCreate(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getCalendarClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Calendar client: %w", err)), nil
	}

	title, errResult := tools.RequireStringParam(req, "title")
	if errResult != nil {
		return errResult, nil
	}

	startStr := tools.GetStringParam(req, "start")
	endStr := tools.GetStringParam(req, "end")
	recurrence := tools.GetStringParam(req, "recurrence")

	start, err := parseTime(startStr)
	if err != nil {
		return tools.CodedErrorResult(tools.ErrInvalidParam, fmt.Errorf("invalid start time: %w", err)), nil
	}

	end, err := parseTime(endStr)
	if err != nil {
		return tools.CodedErrorResult(tools.ErrInvalidParam, fmt.Errorf("invalid end time: %w", err)), nil
	}

	calendarID := tools.GetStringParam(req, "calendar_id")

	event := &clients.CalendarEvent{
		Summary:     title,
		Description: tools.GetStringParam(req, "description"),
		Location:    tools.GetStringParam(req, "location"),
		Start:       start,
		End:         end,
	}

	created, err := client.CreateRecurringEvent(ctx, calendarID, event, recurrence)
	if err != nil {
		return tools.CodedErrorResult(tools.ErrAPIError, fmt.Errorf("failed to create recurring event: %w", err)), nil
	}

	result := map[string]interface{}{
		"event":      created,
		"recurrence": recurrence,
		"message":    fmt.Sprintf("Successfully created recurring event: %s", created.Summary),
	}
	return tools.JSONResult(result), nil
}

func handleQuickAdd(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getCalendarClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Calendar client: %w", err)), nil
	}

	text, errResult := tools.RequireStringParam(req, "text")
	if errResult != nil {
		return errResult, nil
	}

	calendarID := tools.GetStringParam(req, "calendar_id")

	created, err := client.QuickAddEvent(ctx, calendarID, text)
	if err != nil {
		return tools.CodedErrorResult(tools.ErrAPIError, fmt.Errorf("failed to quick add event: %w", err)), nil
	}

	result := map[string]interface{}{
		"event":   created,
		"input":   text,
		"message": fmt.Sprintf("Successfully created event: %s", created.Summary),
	}
	return tools.JSONResult(result), nil
}

func handleFreebusy(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getCalendarClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Calendar client: %w", err)), nil
	}

	startStr := tools.GetStringParam(req, "start")
	endStr := tools.GetStringParam(req, "end")

	start, err := parseTime(startStr)
	if err != nil {
		return tools.CodedErrorResult(tools.ErrInvalidParam, fmt.Errorf("invalid start time: %w", err)), nil
	}

	end, err := parseTime(endStr)
	if err != nil {
		return tools.CodedErrorResult(tools.ErrInvalidParam, fmt.Errorf("invalid end time: %w", err)), nil
	}

	calendarIDs := tools.GetStringArrayParam(req, "calendar_ids")
	if len(calendarIDs) == 0 {
		calendarIDs = []string{"primary"}
	}

	freebusy, err := client.GetFreebusy(ctx, calendarIDs, start, end)
	if err != nil {
		return tools.CodedErrorResult(tools.ErrAPIError, fmt.Errorf("failed to get freebusy: %w", err)), nil
	}

	result := map[string]interface{}{
		"freebusy":   freebusy,
		"time_range": map[string]string{"start": startStr, "end": endStr},
		"calendars":  calendarIDs,
	}
	return tools.JSONResult(result), nil
}

func handleCalendarColors(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getCalendarClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Calendar client: %w", err)), nil
	}

	colors, err := clients.GetCalendarColorsCached(ctx, client)
	if err != nil {
		return tools.CodedErrorResult(tools.ErrAPIError, fmt.Errorf("failed to get colors: %w", err)), nil
	}

	result := map[string]interface{}{
		"colors": colors,
		"count":  len(colors),
	}
	return tools.JSONResult(result), nil
}

func parseTime(s string) (time.Time, error) {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		t, err = time.Parse("2006-01-02T15:04:05", s)
	}
	return t, err
}
