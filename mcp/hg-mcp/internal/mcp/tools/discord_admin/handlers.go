package discordadmin

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

// =============================================================================
// Channel Management Handlers
// =============================================================================

func handleCreateChannel(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	name, errResult := tools.RequireStringParam(req, "name")
	if errResult != nil {
		return errResult, nil
	}

	channelType := tools.GetStringParam(req, "type")
	categoryID := tools.GetStringParam(req, "category_id")
	topic := tools.GetStringParam(req, "topic")

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	defer client.Close()

	channel, err := client.CreateChannel(ctx, clients.ChannelCreateOptions{
		Name:     name,
		Type:     channelTypeFromString(channelType),
		Topic:    topic,
		ParentID: categoryID,
	})
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	return tools.TextResult(fmt.Sprintf("Channel created!\n\n**Name:** %s\n**ID:** %s\n**Type:** %s",
		channel.Name, channel.ID, channel.Type)), nil
}

func handleDeleteChannel(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	channelID, errResult := tools.RequireStringParam(req, "channel_id")
	if errResult != nil {
		return errResult, nil
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	defer client.Close()

	if err := client.DeleteChannel(ctx, channelID); err != nil {
		return tools.ErrorResult(err), nil
	}

	return tools.TextResult(fmt.Sprintf("Channel %s deleted successfully", channelID)), nil
}

func handleCreateCategory(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	name, errResult := tools.RequireStringParam(req, "name")
	if errResult != nil {
		return errResult, nil
	}

	position := tools.GetIntParam(req, "position", 0)

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	defer client.Close()

	category, err := client.CreateCategory(ctx, name, position)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	return tools.TextResult(fmt.Sprintf("Category created!\n\n**Name:** %s\n**ID:** %s",
		category.Name, category.ID)), nil
}

func handleEditChannel(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	channelID, errResult := tools.RequireStringParam(req, "channel_id")
	if errResult != nil {
		return errResult, nil
	}

	name := tools.GetStringParam(req, "name")
	topic := tools.GetStringParam(req, "topic")
	slowmode := tools.GetIntParam(req, "slowmode", 0)

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	defer client.Close()

	channel, err := client.EditChannel(ctx, channelID, clients.ChannelEditOptions{
		Name:     name,
		Topic:    topic,
		Slowmode: slowmode,
	})
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	return tools.TextResult(fmt.Sprintf("Channel updated!\n\n**Name:** %s\n**ID:** %s",
		channel.Name, channel.ID)), nil
}

func handleLockChannel(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	channelID, errResult := tools.RequireStringParam(req, "channel_id")
	if errResult != nil {
		return errResult, nil
	}

	locked := tools.GetBoolParam(req, "locked", false)

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	defer client.Close()

	if locked {
		if err := client.LockChannel(ctx, channelID); err != nil {
			return tools.ErrorResult(err), nil
		}
		return tools.TextResult(fmt.Sprintf("Channel %s locked", channelID)), nil
	} else {
		if err := client.UnlockChannel(ctx, channelID); err != nil {
			return tools.ErrorResult(err), nil
		}
		return tools.TextResult(fmt.Sprintf("Channel %s unlocked", channelID)), nil
	}
}

// =============================================================================
// Role Management Handlers
// =============================================================================

func handleCreateRole(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	name, errResult := tools.RequireStringParam(req, "name")
	if errResult != nil {
		return errResult, nil
	}

	colorStr := tools.GetStringParam(req, "color")
	hoist := tools.GetBoolParam(req, "hoist", false)
	mentionable := tools.GetBoolParam(req, "mentionable", false)

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	defer client.Close()

	role, err := client.CreateRole(ctx, clients.RoleCreateOptions{
		Name:        name,
		Color:       parseColor(colorStr),
		Hoist:       hoist,
		Mentionable: mentionable,
	})
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	return tools.TextResult(fmt.Sprintf("Role created!\n\n**Name:** %s\n**ID:** %s\n**Color:** #%06X",
		role.Name, role.ID, role.Color)), nil
}

func handleDeleteRole(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	roleID, errResult := tools.RequireStringParam(req, "role_id")
	if errResult != nil {
		return errResult, nil
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	defer client.Close()

	if err := client.DeleteRole(ctx, roleID); err != nil {
		return tools.ErrorResult(err), nil
	}

	return tools.TextResult(fmt.Sprintf("Role %s deleted successfully", roleID)), nil
}

func handleAssignRole(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	userID, errResult := tools.RequireStringParam(req, "user_id")
	if errResult != nil {
		return errResult, nil
	}
	roleID, errResult := tools.RequireStringParam(req, "role_id")
	if errResult != nil {
		return errResult, nil
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	defer client.Close()

	if err := client.AssignRole(ctx, userID, roleID); err != nil {
		return tools.ErrorResult(err), nil
	}

	return tools.TextResult(fmt.Sprintf("Role %s assigned to user %s", roleID, userID)), nil
}

func handleRevokeRole(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	userID, errResult := tools.RequireStringParam(req, "user_id")
	if errResult != nil {
		return errResult, nil
	}
	roleID, errResult := tools.RequireStringParam(req, "role_id")
	if errResult != nil {
		return errResult, nil
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	defer client.Close()

	if err := client.RevokeRole(ctx, userID, roleID); err != nil {
		return tools.ErrorResult(err), nil
	}

	return tools.TextResult(fmt.Sprintf("Role %s revoked from user %s", roleID, userID)), nil
}

func handleListRoles(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	defer client.Close()

	roles, err := client.ListRoles(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# Server Roles\n\n")
	sb.WriteString("| Position | Name | ID | Color | Mentionable |\n")
	sb.WriteString("|----------|------|-----|-------|-------------|\n")

	for _, role := range roles {
		mentionable := "No"
		if role.Mentionable {
			mentionable = "Yes"
		}
		sb.WriteString(fmt.Sprintf("| %d | %s | `%s` | #%06X | %s |\n",
			role.Position, role.Name, role.ID, role.Color, mentionable))
	}

	return tools.TextResult(sb.String()), nil
}

func handleEditRole(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	roleID, errResult := tools.RequireStringParam(req, "role_id")
	if errResult != nil {
		return errResult, nil
	}

	name := tools.GetStringParam(req, "name")
	colorStr := tools.GetStringParam(req, "color")

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	defer client.Close()

	role, err := client.EditRole(ctx, roleID, clients.RoleCreateOptions{
		Name:  name,
		Color: parseColor(colorStr),
	})
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	return tools.TextResult(fmt.Sprintf("Role updated!\n\n**Name:** %s\n**ID:** %s\n**Color:** #%06X",
		role.Name, role.ID, role.Color)), nil
}

// =============================================================================
// Member Management Handlers
// =============================================================================

func handleMemberInfo(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	userID, errResult := tools.RequireStringParam(req, "user_id")
	if errResult != nil {
		return errResult, nil
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	defer client.Close()

	member, err := client.GetMemberInfo(ctx, userID)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# Member Info\n\n")
	sb.WriteString(fmt.Sprintf("**Username:** %s\n", member.Username))
	sb.WriteString(fmt.Sprintf("**Display Name:** %s\n", member.DisplayName))
	if member.Nickname != "" {
		sb.WriteString(fmt.Sprintf("**Nickname:** %s\n", member.Nickname))
	}
	sb.WriteString(fmt.Sprintf("**User ID:** `%s`\n", member.UserID))
	sb.WriteString(fmt.Sprintf("**Joined:** %s\n", member.JoinedAt.Format("Jan 2, 2006")))
	sb.WriteString(fmt.Sprintf("**Roles:** %d\n", len(member.Roles)))
	if len(member.Roles) > 0 {
		sb.WriteString(fmt.Sprintf("**Role IDs:** %s\n", strings.Join(member.Roles, ", ")))
	}
	if member.IsMuted {
		sb.WriteString("**Server Muted:** Yes\n")
	}
	if member.IsDeafened {
		sb.WriteString("**Server Deafened:** Yes\n")
	}

	return tools.TextResult(sb.String()), nil
}

func handleListMembers(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	limit := tools.GetIntParam(req, "limit", 100)

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	defer client.Close()

	members, err := client.ListMembers(ctx, limit, "")
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Server Members (%d)\n\n", len(members)))
	sb.WriteString("| Username | Display Name | Roles | Joined |\n")
	sb.WriteString("|----------|--------------|-------|--------|\n")

	for _, m := range members {
		sb.WriteString(fmt.Sprintf("| %s | %s | %d | %s |\n",
			m.Username, m.DisplayName, len(m.Roles), m.JoinedAt.Format("Jan 2, 2006")))
	}

	return tools.TextResult(sb.String()), nil
}

func handleKick(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	userID, errResult := tools.RequireStringParam(req, "user_id")
	if errResult != nil {
		return errResult, nil
	}

	reason := tools.GetStringParam(req, "reason")

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	defer client.Close()

	if err := client.KickMember(ctx, userID, reason); err != nil {
		return tools.ErrorResult(err), nil
	}

	result := fmt.Sprintf("User %s kicked from the server", userID)
	if reason != "" {
		result += fmt.Sprintf("\n**Reason:** %s", reason)
	}
	return tools.TextResult(result), nil
}

func handleBan(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	userID, errResult := tools.RequireStringParam(req, "user_id")
	if errResult != nil {
		return errResult, nil
	}

	reason := tools.GetStringParam(req, "reason")
	deleteDays := tools.GetIntParam(req, "delete_days", 0)

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	defer client.Close()

	if err := client.BanMember(ctx, userID, reason, deleteDays); err != nil {
		return tools.ErrorResult(err), nil
	}

	result := fmt.Sprintf("User %s banned from the server", userID)
	if reason != "" {
		result += fmt.Sprintf("\n**Reason:** %s", reason)
	}
	if deleteDays > 0 {
		result += fmt.Sprintf("\n**Messages deleted:** %d days", deleteDays)
	}
	return tools.TextResult(result), nil
}

func handleUnban(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	userID, errResult := tools.RequireStringParam(req, "user_id")
	if errResult != nil {
		return errResult, nil
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	defer client.Close()

	if err := client.UnbanMember(ctx, userID); err != nil {
		return tools.ErrorResult(err), nil
	}

	return tools.TextResult(fmt.Sprintf("User %s unbanned", userID)), nil
}

func handleTimeout(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	userID, errResult := tools.RequireStringParam(req, "user_id")
	if errResult != nil {
		return errResult, nil
	}

	durationMinutes, errResult := tools.RequireIntParam(req, "duration_minutes")
	if errResult != nil {
		return errResult, nil
	}

	reason := tools.GetStringParam(req, "reason")

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	defer client.Close()

	duration := time.Duration(durationMinutes) * time.Minute
	if err := client.TimeoutMember(ctx, userID, duration, reason); err != nil {
		return tools.ErrorResult(err), nil
	}

	result := fmt.Sprintf("User %s timed out for %d minutes", userID, durationMinutes)
	if reason != "" {
		result += fmt.Sprintf("\n**Reason:** %s", reason)
	}
	return tools.TextResult(result), nil
}

func handleSetNickname(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	userID, errResult := tools.RequireStringParam(req, "user_id")
	if errResult != nil {
		return errResult, nil
	}
	nickname := tools.GetStringParam(req, "nickname")

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	defer client.Close()

	if err := client.SetNickname(ctx, userID, nickname); err != nil {
		return tools.ErrorResult(err), nil
	}

	if nickname == "" {
		return tools.TextResult(fmt.Sprintf("Nickname cleared for user %s", userID)), nil
	}
	return tools.TextResult(fmt.Sprintf("Nickname set to '%s' for user %s", nickname, userID)), nil
}

// =============================================================================
// Voice Management Handlers
// =============================================================================

func handleVoiceStatus(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	defer client.Close()

	status, err := client.GetVoiceStatus(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# Voice Channel Status\n\n")

	for _, ch := range status {
		sb.WriteString(fmt.Sprintf("## %s\n", ch.ChannelName))
		sb.WriteString(fmt.Sprintf("**ID:** `%s`\n", ch.ChannelID))
		sb.WriteString(fmt.Sprintf("**Members:** %d\n", ch.MemberCount))
		if len(ch.Members) > 0 {
			sb.WriteString(fmt.Sprintf("**User IDs:** %s\n", strings.Join(ch.Members, ", ")))
		}
		sb.WriteString("\n")
	}

	if len(status) == 0 {
		sb.WriteString("No voice channels found.\n")
	}

	return tools.TextResult(sb.String()), nil
}

func handleVoiceMove(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	userID, errResult := tools.RequireStringParam(req, "user_id")
	if errResult != nil {
		return errResult, nil
	}
	channelID, errResult := tools.RequireStringParam(req, "channel_id")
	if errResult != nil {
		return errResult, nil
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	defer client.Close()

	if err := client.VoiceMove(ctx, userID, channelID); err != nil {
		return tools.ErrorResult(err), nil
	}

	return tools.TextResult(fmt.Sprintf("User %s moved to channel %s", userID, channelID)), nil
}

func handleVoiceDisconnect(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	userID, errResult := tools.RequireStringParam(req, "user_id")
	if errResult != nil {
		return errResult, nil
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	defer client.Close()

	if err := client.VoiceDisconnect(ctx, userID); err != nil {
		return tools.ErrorResult(err), nil
	}

	return tools.TextResult(fmt.Sprintf("User %s disconnected from voice", userID)), nil
}

func handleVoiceMute(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	userID, errResult := tools.RequireStringParam(req, "user_id")
	if errResult != nil {
		return errResult, nil
	}

	muted := tools.GetBoolParam(req, "muted", false)

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	defer client.Close()

	if err := client.VoiceMute(ctx, userID, muted); err != nil {
		return tools.ErrorResult(err), nil
	}

	action := "unmuted"
	if muted {
		action = "muted"
	}
	return tools.TextResult(fmt.Sprintf("User %s %s", userID, action)), nil
}

func handleVoiceDeafen(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	userID, errResult := tools.RequireStringParam(req, "user_id")
	if errResult != nil {
		return errResult, nil
	}

	deafened := tools.GetBoolParam(req, "deafened", false)

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	defer client.Close()

	if err := client.VoiceDeafen(ctx, userID, deafened); err != nil {
		return tools.ErrorResult(err), nil
	}

	action := "undeafened"
	if deafened {
		action = "deafened"
	}
	return tools.TextResult(fmt.Sprintf("User %s %s", userID, action)), nil
}

// =============================================================================
// Moderation Handlers
// =============================================================================

func handlePurge(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	channelID, errResult := tools.RequireStringParam(req, "channel_id")
	if errResult != nil {
		return errResult, nil
	}
	count, errResult := tools.RequireIntParam(req, "count")
	if errResult != nil {
		return errResult, nil
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	defer client.Close()

	deleted, err := client.PurgeMessages(ctx, channelID, count)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	return tools.TextResult(fmt.Sprintf("Deleted %d messages from channel %s", deleted, channelID)), nil
}

func handlePin(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	channelID, errResult := tools.RequireStringParam(req, "channel_id")
	if errResult != nil {
		return errResult, nil
	}
	messageID, errResult := tools.RequireStringParam(req, "message_id")
	if errResult != nil {
		return errResult, nil
	}

	pinned := tools.GetBoolParam(req, "pinned", false)

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	defer client.Close()

	if pinned {
		if err := client.PinMessage(ctx, channelID, messageID); err != nil {
			return tools.ErrorResult(err), nil
		}
		return tools.TextResult(fmt.Sprintf("Message %s pinned", messageID)), nil
	} else {
		if err := client.UnpinMessage(ctx, channelID, messageID); err != nil {
			return tools.ErrorResult(err), nil
		}
		return tools.TextResult(fmt.Sprintf("Message %s unpinned", messageID)), nil
	}
}

func handleAuditLog(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	limit := tools.GetIntParam(req, "limit", 50)
	userID := tools.GetStringParam(req, "user_id")
	actionType := tools.GetIntParam(req, "action_type", 0)

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	defer client.Close()

	entries, err := client.GetAuditLog(ctx, actionType, userID, limit)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Audit Log (%d entries)\n\n", len(entries)))

	for _, entry := range entries {
		sb.WriteString(fmt.Sprintf("## %s\n", entry.CreatedAt.Format("Jan 2, 2006 15:04:05")))
		sb.WriteString(fmt.Sprintf("**Action Type:** %d\n", entry.ActionType))
		sb.WriteString(fmt.Sprintf("**User ID:** %s\n", entry.UserID))
		if entry.TargetID != "" {
			sb.WriteString(fmt.Sprintf("**Target ID:** %s\n", entry.TargetID))
		}
		if entry.Reason != "" {
			sb.WriteString(fmt.Sprintf("**Reason:** %s\n", entry.Reason))
		}
		sb.WriteString("\n")
	}

	if len(entries) == 0 {
		sb.WriteString("No audit log entries found.\n")
	}

	return tools.TextResult(sb.String()), nil
}

func handleSlowmode(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	channelID, errResult := tools.RequireStringParam(req, "channel_id")
	if errResult != nil {
		return errResult, nil
	}
	seconds := tools.GetIntParam(req, "seconds", 0)

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	defer client.Close()

	if err := client.SetSlowmode(ctx, channelID, seconds); err != nil {
		return tools.ErrorResult(err), nil
	}

	if seconds == 0 {
		return tools.TextResult(fmt.Sprintf("Slowmode disabled for channel %s", channelID)), nil
	}
	return tools.TextResult(fmt.Sprintf("Slowmode set to %d seconds for channel %s", seconds, channelID)), nil
}

func handleListBans(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	limit := tools.GetIntParam(req, "limit", 100)

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	defer client.Close()

	bans, err := client.ListBans(ctx, limit)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Banned Users (%d)\n\n", len(bans)))

	if len(bans) == 0 {
		sb.WriteString("No banned users.\n")
	} else {
		sb.WriteString("| User ID | Username | Reason |\n")
		sb.WriteString("|---------|----------|--------|\n")
		for _, ban := range bans {
			reason := ""
			if r, ok := ban["reason"].(string); ok && r != "" {
				reason = r
			}
			sb.WriteString(fmt.Sprintf("| `%s` | %s | %s |\n",
				ban["user_id"], ban["username"], reason))
		}
	}

	return tools.TextResult(sb.String()), nil
}

// =============================================================================
// Scheduled Events Handlers
// =============================================================================

func handleCreateEvent(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	name, errResult := tools.RequireStringParam(req, "name")
	if errResult != nil {
		return errResult, nil
	}

	startTimeStr, errResult := tools.RequireStringParam(req, "start_time")
	if errResult != nil {
		return errResult, nil
	}

	startTime, err := time.Parse(time.RFC3339, startTimeStr)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("invalid start_time format (use RFC3339): %w", err)), nil
	}

	description := tools.GetStringParam(req, "description")
	endTimeStr := tools.GetStringParam(req, "end_time")
	location := tools.GetStringParam(req, "location")
	channelID := tools.GetStringParam(req, "channel_id")

	var endTime time.Time
	if endTimeStr != "" {
		endTime, err = time.Parse(time.RFC3339, endTimeStr)
		if err != nil {
			return tools.ErrorResult(fmt.Errorf("invalid end_time format: %w", err)), nil
		}
	}

	// Determine event type
	eventType := clients.ScheduledEventVoice
	if location != "" {
		eventType = clients.ScheduledEventExternal
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	defer client.Close()

	event, err := client.CreateScheduledEvent(ctx, clients.ScheduledEventCreateOptions{
		Name:        name,
		Description: description,
		ChannelID:   channelID,
		Location:    location,
		StartTime:   startTime,
		EndTime:     endTime,
		EventType:   eventType,
	})
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	return tools.TextResult(fmt.Sprintf("Event created!\n\n**Name:** %s\n**ID:** %s\n**Start:** %s",
		event.Name, event.ID, event.StartTime.Format("Jan 2, 2006 15:04"))), nil
}

func handleUpdateEvent(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	eventID, errResult := tools.RequireStringParam(req, "event_id")
	if errResult != nil {
		return errResult, nil
	}

	name := tools.GetStringParam(req, "name")
	description := tools.GetStringParam(req, "description")
	startTimeStr := tools.GetStringParam(req, "start_time")

	var startTime time.Time
	if startTimeStr != "" {
		var err error
		startTime, err = time.Parse(time.RFC3339, startTimeStr)
		if err != nil {
			return tools.ErrorResult(fmt.Errorf("invalid start_time format: %w", err)), nil
		}
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	defer client.Close()

	event, err := client.UpdateScheduledEvent(ctx, eventID, clients.ScheduledEventCreateOptions{
		Name:        name,
		Description: description,
		StartTime:   startTime,
	})
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	return tools.TextResult(fmt.Sprintf("Event updated!\n\n**Name:** %s\n**ID:** %s",
		event.Name, event.ID)), nil
}

func handleDeleteEvent(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	eventID, errResult := tools.RequireStringParam(req, "event_id")
	if errResult != nil {
		return errResult, nil
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	defer client.Close()

	if err := client.DeleteScheduledEvent(ctx, eventID); err != nil {
		return tools.ErrorResult(err), nil
	}

	return tools.TextResult(fmt.Sprintf("Event %s deleted", eventID)), nil
}

func handleListEvents(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	defer client.Close()

	events, err := client.ListScheduledEvents(ctx, true)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Scheduled Events (%d)\n\n", len(events)))

	for _, event := range events {
		sb.WriteString(fmt.Sprintf("## %s\n", event.Name))
		sb.WriteString(fmt.Sprintf("**ID:** `%s`\n", event.ID))
		sb.WriteString(fmt.Sprintf("**Start:** %s\n", event.StartTime.Format("Jan 2, 2006 15:04")))
		if !event.EndTime.IsZero() {
			sb.WriteString(fmt.Sprintf("**End:** %s\n", event.EndTime.Format("Jan 2, 2006 15:04")))
		}
		if event.Description != "" {
			sb.WriteString(fmt.Sprintf("**Description:** %s\n", event.Description))
		}
		if event.Location != "" {
			sb.WriteString(fmt.Sprintf("**Location:** %s\n", event.Location))
		}
		sb.WriteString(fmt.Sprintf("**Interested:** %d\n\n", event.UserCount))
	}

	if len(events) == 0 {
		sb.WriteString("No scheduled events.\n")
	}

	return tools.TextResult(sb.String()), nil
}

// =============================================================================
// Interactive Components Handlers
// =============================================================================

type buttonConfig struct {
	Label    string `json:"label"`
	CustomID string `json:"custom_id"`
	Style    string `json:"style"`
}

func handleSendButtons(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	channelID := tools.GetStringParam(req, "channel_id")
	content, errResult := tools.RequireStringParam(req, "content")
	if errResult != nil {
		return errResult, nil
	}
	buttonsJSON, errResult := tools.RequireStringParam(req, "buttons")
	if errResult != nil {
		return errResult, nil
	}

	var buttonConfigs []buttonConfig
	if err := json.Unmarshal([]byte(buttonsJSON), &buttonConfigs); err != nil {
		return tools.ErrorResult(fmt.Errorf("invalid buttons JSON: %w", err)), nil
	}

	buttons := make([]clients.Button, len(buttonConfigs))
	for i, bc := range buttonConfigs {
		buttons[i] = clients.Button{
			Label:    bc.Label,
			CustomID: bc.CustomID,
			Style:    buttonStyleFromString(bc.Style),
		}
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	defer client.Close()

	msg, err := client.SendMessageWithButtons(ctx, channelID, content, buttons)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	return tools.TextResult(fmt.Sprintf("Message with buttons sent!\n\n**Message ID:** %s\n**Buttons:** %d",
		msg.ID, len(buttons))), nil
}

type selectOptionConfig struct {
	Label       string `json:"label"`
	Value       string `json:"value"`
	Description string `json:"description"`
}

func handleSendSelect(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	channelID := tools.GetStringParam(req, "channel_id")
	content, errResult := tools.RequireStringParam(req, "content")
	if errResult != nil {
		return errResult, nil
	}
	placeholder := tools.GetStringParam(req, "placeholder")
	customID, errResult := tools.RequireStringParam(req, "custom_id")
	if errResult != nil {
		return errResult, nil
	}
	optionsJSON, errResult := tools.RequireStringParam(req, "options")
	if errResult != nil {
		return errResult, nil
	}

	var optionConfigs []selectOptionConfig
	if err := json.Unmarshal([]byte(optionsJSON), &optionConfigs); err != nil {
		return tools.ErrorResult(fmt.Errorf("invalid options JSON: %w", err)), nil
	}

	options := make([]clients.SelectOption, len(optionConfigs))
	for i, oc := range optionConfigs {
		options[i] = clients.SelectOption{
			Label:       oc.Label,
			Value:       oc.Value,
			Description: oc.Description,
		}
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	defer client.Close()

	msg, err := client.SendMessageWithSelect(ctx, channelID, content, placeholder, customID, options)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	return tools.TextResult(fmt.Sprintf("Message with select menu sent!\n\n**Message ID:** %s\n**Options:** %d",
		msg.ID, len(options))), nil
}

// =============================================================================
// Server Info Handler
// =============================================================================

func handleServerInfo(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	defer client.Close()

	info, err := client.GetServerInfo(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# Server Information\n\n")
	sb.WriteString(fmt.Sprintf("**Name:** %s\n", info.Name))
	sb.WriteString(fmt.Sprintf("**ID:** `%s`\n", info.ID))
	sb.WriteString(fmt.Sprintf("**Owner ID:** `%s`\n", info.OwnerID))
	sb.WriteString(fmt.Sprintf("**Created:** %s\n\n", info.CreatedAt.Format("Jan 2, 2006")))

	sb.WriteString("## Statistics\n\n")
	sb.WriteString(fmt.Sprintf("- **Members:** %d\n", info.MemberCount))
	sb.WriteString(fmt.Sprintf("- **Channels:** %d\n", info.ChannelCount))
	sb.WriteString(fmt.Sprintf("- **Roles:** %d\n", info.RoleCount))
	sb.WriteString(fmt.Sprintf("- **Emojis:** %d\n", info.EmojiCount))

	sb.WriteString("\n## Boost Status\n\n")
	sb.WriteString(fmt.Sprintf("- **Boost Level:** %d\n", info.BoostLevel))
	sb.WriteString(fmt.Sprintf("- **Boost Count:** %d\n", info.BoostCount))

	if info.Description != "" {
		sb.WriteString(fmt.Sprintf("\n## Description\n\n%s\n", info.Description))
	}

	return tools.TextResult(sb.String()), nil
}
