// Package twitch provides MCP tools for Twitch streaming integration.
package twitch

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

var getClient = tools.LazyClient(clients.NewTwitchClient)

// Module implements the ToolModule interface for Twitch.
type Module struct{}

func init() {
	tools.GetRegistry().RegisterModule(&Module{})
}

func (m *Module) Name() string        { return "twitch" }
func (m *Module) Description() string { return "Twitch streaming platform integration" }

func (m *Module) Tools() []tools.ToolDefinition {
	allTools := []tools.ToolDefinition{
		{
			Tool:        mcp.NewTool("aftrs_twitch_status", mcp.WithDescription("Get Twitch stream status including live state, viewer count, and stream info.")),
			Handler:     handleStatus,
			Category:    "twitch",
			Subcategory: "status",
			Tags:        []string{"twitch", "streaming", "status"},
			UseCases:    []string{"Check if stream is live", "Get viewer count"},
			Complexity:  tools.ComplexitySimple,
		},
		{
			Tool: mcp.NewTool("aftrs_twitch_stream", mcp.WithDescription("Control and update stream information."),
				mcp.WithString("action", mcp.Description("Action: info, update_title, update_game")),
				mcp.WithString("title", mcp.Description("New stream title")),
				mcp.WithString("game", mcp.Description("Game name to search/set"))),
			Handler:     handleStream,
			Category:    "twitch",
			Subcategory: "stream",
			Tags:        []string{"twitch", "stream", "title", "game"},
			UseCases:    []string{"Update stream title", "Change game category"},
			Complexity:  tools.ComplexitySimple,
			IsWrite:     true,
		},
		{
			Tool: mcp.NewTool("aftrs_twitch_chat", mcp.WithDescription("Send messages to Twitch chat."),
				mcp.WithString("message", mcp.Required(), mcp.Description("Message to send"))),
			Handler:     handleChat,
			Category:    "twitch",
			Subcategory: "chat",
			Tags:        []string{"twitch", "chat", "message"},
			UseCases:    []string{"Send announcement", "Interact with viewers"},
			Complexity:  tools.ComplexitySimple,
			IsWrite:     true,
		},
		{
			Tool: mcp.NewTool("aftrs_twitch_mod", mcp.WithDescription("Moderate chat users."),
				mcp.WithString("action", mcp.Required(), mcp.Description("Action: timeout, ban, unban")),
				mcp.WithString("user", mcp.Required(), mcp.Description("Username")),
				mcp.WithNumber("duration", mcp.Description("Timeout seconds (default 600)")),
				mcp.WithString("reason", mcp.Description("Reason"))),
			Handler:     handleMod,
			Category:    "twitch",
			Subcategory: "moderation",
			Tags:        []string{"twitch", "moderation", "ban", "timeout"},
			UseCases:    []string{"Timeout user", "Ban user", "Unban user"},
			Complexity:  tools.ComplexitySimple,
			IsWrite:     true,
		},
		{
			Tool: mcp.NewTool("aftrs_twitch_poll", mcp.WithDescription("Create and manage polls."),
				mcp.WithString("action", mcp.Required(), mcp.Description("Action: create, end, list")),
				mcp.WithString("title", mcp.Description("Poll title")),
				mcp.WithString("choices", mcp.Description("Comma-separated choices")),
				mcp.WithNumber("duration", mcp.Description("Duration seconds")),
				mcp.WithString("poll_id", mcp.Description("Poll ID for end"))),
			Handler:     handlePoll,
			Category:    "twitch",
			Subcategory: "engagement",
			Tags:        []string{"twitch", "poll", "engagement"},
			UseCases:    []string{"Create viewer poll", "End poll", "List polls"},
			Complexity:  tools.ComplexityModerate,
			IsWrite:     true,
		},
		{
			Tool: mcp.NewTool("aftrs_twitch_prediction", mcp.WithDescription("Create and manage predictions."),
				mcp.WithString("action", mcp.Required(), mcp.Description("Action: create, resolve, cancel, list")),
				mcp.WithString("title", mcp.Description("Prediction title")),
				mcp.WithString("outcomes", mcp.Description("Comma-separated outcomes")),
				mcp.WithNumber("duration", mcp.Description("Duration seconds")),
				mcp.WithString("prediction_id", mcp.Description("Prediction ID")),
				mcp.WithString("winning_outcome_id", mcp.Description("Winning outcome ID"))),
			Handler:     handlePrediction,
			Category:    "twitch",
			Subcategory: "engagement",
			Tags:        []string{"twitch", "prediction", "channel-points"},
			UseCases:    []string{"Create prediction", "Resolve prediction", "Cancel prediction"},
			Complexity:  tools.ComplexityModerate,
			IsWrite:     true,
		},
		{
			Tool:        mcp.NewTool("aftrs_twitch_clip", mcp.WithDescription("Create clips from current stream.")),
			Handler:     handleClip,
			Category:    "twitch",
			Subcategory: "content",
			Tags:        []string{"twitch", "clip", "highlight"},
			UseCases:    []string{"Create clip from stream"},
			Complexity:  tools.ComplexitySimple,
			IsWrite:     true,
		},
		{
			Tool: mcp.NewTool("aftrs_twitch_markers", mcp.WithDescription("Add stream markers."),
				mcp.WithString("description", mcp.Description("Marker description"))),
			Handler:     handleMarkers,
			Category:    "twitch",
			Subcategory: "content",
			Tags:        []string{"twitch", "marker", "highlight"},
			UseCases:    []string{"Mark highlight moment"},
			Complexity:  tools.ComplexitySimple,
			IsWrite:     true,
		},
		{
			Tool: mcp.NewTool("aftrs_twitch_raid", mcp.WithDescription("Start or cancel a raid."),
				mcp.WithString("action", mcp.Required(), mcp.Description("Action: start, cancel")),
				mcp.WithString("channel", mcp.Description("Target channel"))),
			Handler:     handleRaid,
			Category:    "twitch",
			Subcategory: "stream",
			Tags:        []string{"twitch", "raid", "host"},
			UseCases:    []string{"Start raid to another channel", "Cancel raid"},
			Complexity:  tools.ComplexitySimple,
			IsWrite:     true,
		},
		{
			Tool:        mcp.NewTool("aftrs_twitch_health", mcp.WithDescription("Check Twitch API health.")),
			Handler:     handleHealth,
			Category:    "twitch",
			Subcategory: "health",
			Tags:        []string{"twitch", "health", "status"},
			UseCases:    []string{"Check API connection", "Diagnose issues"},
			Complexity:  tools.ComplexitySimple,
		},
	}

	// Apply circuit breaker to all tools — network-dependent (Twitch API)
	for i := range allTools {
		allTools[i].CircuitBreakerGroup = "twitch"
	}

	return allTools
}

func handleStatus(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	status, err := client.GetStatus(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	return tools.JSONResult(status), nil
}

func handleStream(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	action := tools.OptionalStringParam(req, "action", "info")
	switch action {
	case "info":
		stream, err := client.GetStream(ctx, "")
		if err != nil {
			return tools.ErrorResult(err), nil
		}
		if stream == nil {
			return tools.TextResult("Stream is offline"), nil
		}
		return tools.JSONResult(stream), nil
	case "update_title":
		title, errResult := tools.RequireStringParam(req, "title")
		if errResult != nil {
			return errResult, nil
		}
		if err := client.UpdateStreamInfo(ctx, title, ""); err != nil {
			return tools.ErrorResult(err), nil
		}
		return tools.TextResult(fmt.Sprintf("Title updated to: %s", title)), nil
	case "update_game":
		game, errResult := tools.RequireStringParam(req, "game")
		if errResult != nil {
			return errResult, nil
		}
		games, err := client.SearchGames(ctx, game)
		if err != nil || len(games) == 0 {
			return tools.ErrorResult(fmt.Errorf("game not found: %s", game)), nil
		}
		gameID, _ := games[0]["id"].(string)
		gameName, _ := games[0]["name"].(string)
		if err := client.UpdateStreamInfo(ctx, "", gameID); err != nil {
			return tools.ErrorResult(err), nil
		}
		return tools.TextResult(fmt.Sprintf("Game updated to: %s", gameName)), nil
	}
	return tools.ErrorResult(fmt.Errorf("unknown action: %s", action)), nil
}

func handleChat(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	message, errResult := tools.RequireStringParam(req, "message")
	if errResult != nil {
		return errResult, nil
	}
	if err := client.SendChatMessage(ctx, message); err != nil {
		return tools.ErrorResult(err), nil
	}
	return tools.TextResult("Message sent"), nil
}

func handleMod(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	action := tools.GetStringParam(req, "action")
	username, errResult := tools.RequireStringParam(req, "user")
	if errResult != nil {
		return errResult, nil
	}
	reason := tools.GetStringParam(req, "reason")
	user, err := client.GetUserByLogin(ctx, username)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("user not found: %v", err)), nil
	}
	switch action {
	case "timeout":
		duration := tools.GetIntParam(req, "duration", 600)
		if err := client.TimeoutUser(ctx, user.ID, duration, reason); err != nil {
			return tools.ErrorResult(err), nil
		}
		return tools.TextResult(fmt.Sprintf("User %s timed out for %ds", username, duration)), nil
	case "ban":
		if err := client.BanUser(ctx, user.ID, reason); err != nil {
			return tools.ErrorResult(err), nil
		}
		return tools.TextResult(fmt.Sprintf("User %s banned", username)), nil
	case "unban":
		if err := client.UnbanUser(ctx, user.ID); err != nil {
			return tools.ErrorResult(err), nil
		}
		return tools.TextResult(fmt.Sprintf("User %s unbanned", username)), nil
	}
	return tools.ErrorResult(fmt.Errorf("unknown action: %s", action)), nil
}

func handlePoll(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	action := tools.GetStringParam(req, "action")
	switch action {
	case "create":
		title, errResult := tools.RequireStringParam(req, "title")
		if errResult != nil {
			return errResult, nil
		}
		choicesStr, errResult := tools.RequireStringParam(req, "choices")
		if errResult != nil {
			return errResult, nil
		}
		choices := strings.Split(choicesStr, ",")
		for i := range choices {
			choices[i] = strings.TrimSpace(choices[i])
		}
		duration := tools.GetIntParam(req, "duration", 60)
		poll, err := client.CreatePoll(ctx, title, choices, duration)
		if err != nil {
			return tools.ErrorResult(err), nil
		}
		return tools.JSONResult(poll), nil
	case "end":
		pollID, errResult := tools.RequireStringParam(req, "poll_id")
		if errResult != nil {
			return errResult, nil
		}
		poll, err := client.EndPoll(ctx, pollID, true)
		if err != nil {
			return tools.ErrorResult(err), nil
		}
		return tools.JSONResult(poll), nil
	case "list":
		polls, err := client.GetPolls(ctx)
		if err != nil {
			return tools.ErrorResult(err), nil
		}
		return tools.JSONResult(polls), nil
	}
	return tools.ErrorResult(fmt.Errorf("unknown action: %s", action)), nil
}

func handlePrediction(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	action := tools.GetStringParam(req, "action")
	switch action {
	case "create":
		title, errResult := tools.RequireStringParam(req, "title")
		if errResult != nil {
			return errResult, nil
		}
		outcomesStr, errResult := tools.RequireStringParam(req, "outcomes")
		if errResult != nil {
			return errResult, nil
		}
		outcomes := strings.Split(outcomesStr, ",")
		for i := range outcomes {
			outcomes[i] = strings.TrimSpace(outcomes[i])
		}
		duration := tools.GetIntParam(req, "duration", 120)
		pred, err := client.CreatePrediction(ctx, title, outcomes, duration)
		if err != nil {
			return tools.ErrorResult(err), nil
		}
		return tools.JSONResult(pred), nil
	case "resolve":
		predID, errResult := tools.RequireStringParam(req, "prediction_id")
		if errResult != nil {
			return errResult, nil
		}
		winID, errResult := tools.RequireStringParam(req, "winning_outcome_id")
		if errResult != nil {
			return errResult, nil
		}
		pred, err := client.ResolvePrediction(ctx, predID, winID)
		if err != nil {
			return tools.ErrorResult(err), nil
		}
		return tools.JSONResult(pred), nil
	case "cancel":
		predID, errResult := tools.RequireStringParam(req, "prediction_id")
		if errResult != nil {
			return errResult, nil
		}
		pred, err := client.CancelPrediction(ctx, predID)
		if err != nil {
			return tools.ErrorResult(err), nil
		}
		return tools.JSONResult(pred), nil
	case "list":
		preds, err := client.GetPredictions(ctx)
		if err != nil {
			return tools.ErrorResult(err), nil
		}
		return tools.JSONResult(preds), nil
	}
	return tools.ErrorResult(fmt.Errorf("unknown action: %s", action)), nil
}

func handleClip(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	clip, err := client.CreateClip(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	return tools.JSONResult(clip), nil
}

func handleMarkers(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	desc := tools.OptionalStringParam(req, "description", "Highlight")
	marker, err := client.CreateStreamMarker(ctx, desc)
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	return tools.JSONResult(marker), nil
}

func handleRaid(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	action := tools.GetStringParam(req, "action")
	switch action {
	case "start":
		channel, errResult := tools.RequireStringParam(req, "channel")
		if errResult != nil {
			return errResult, nil
		}
		user, err := client.GetUserByLogin(ctx, channel)
		if err != nil {
			return tools.ErrorResult(fmt.Errorf("channel not found: %v", err)), nil
		}
		if err := client.StartRaid(ctx, user.ID); err != nil {
			return tools.ErrorResult(err), nil
		}
		return tools.TextResult(fmt.Sprintf("Raid started to %s", channel)), nil
	case "cancel":
		if err := client.CancelRaid(ctx); err != nil {
			return tools.ErrorResult(err), nil
		}
		return tools.TextResult("Raid cancelled"), nil
	}
	return tools.ErrorResult(fmt.Errorf("unknown action: %s", action)), nil
}

func handleHealth(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	health, err := client.GetHealth(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	return tools.JSONResult(health), nil
}
