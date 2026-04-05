// Package mqtt provides MQTT/IoT messaging tools for hg-mcp.
package mqtt

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

// Module implements the ToolModule interface for MQTT/IoT
type Module struct{}

// getClient returns the singleton MQTT client (thread-safe via LazyClient)
var getClient = tools.LazyClient(clients.NewMQTTClient)

func (m *Module) Name() string {
	return "mqtt"
}

func (m *Module) Description() string {
	return "MQTT messaging and IoT device integration"
}

func (m *Module) Tools() []tools.ToolDefinition {
	return []tools.ToolDefinition{
		{
			Tool: mcp.NewTool("aftrs_mqtt_status",
				mcp.WithDescription("Get MQTT broker connection status."),
			),
			Handler:             handleStatus,
			Category:            "mqtt",
			Subcategory:         "status",
			Tags:                []string{"mqtt", "iot", "status", "broker"},
			UseCases:            []string{"Check MQTT connection", "View subscriptions"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "mqtt",
		},
		{
			Tool: mcp.NewTool("aftrs_mqtt_connect",
				mcp.WithDescription("Connect to MQTT broker."),
				mcp.WithString("broker", mcp.Description("Broker URL (default: tcp://localhost:1883)")),
			),
			Handler:             handleConnect,
			Category:            "mqtt",
			Subcategory:         "connection",
			Tags:                []string{"mqtt", "connect", "broker"},
			UseCases:            []string{"Connect to MQTT broker", "Establish connection"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "mqtt",
		},
		{
			Tool: mcp.NewTool("aftrs_mqtt_publish",
				mcp.WithDescription("Publish a message to an MQTT topic."),
				mcp.WithString("topic", mcp.Required(), mcp.Description("Topic to publish to")),
				mcp.WithString("message", mcp.Required(), mcp.Description("Message payload")),
				mcp.WithBoolean("retain", mcp.Description("Retain message on broker")),
			),
			Handler:             handlePublish,
			Category:            "mqtt",
			Subcategory:         "publish",
			Tags:                []string{"mqtt", "publish", "message", "iot"},
			UseCases:            []string{"Send MQTT message", "Control IoT device"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "mqtt",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_mqtt_subscribe",
				mcp.WithDescription("Subscribe to an MQTT topic."),
				mcp.WithString("topic", mcp.Required(), mcp.Description("Topic to subscribe to (supports + and # wildcards)")),
			),
			Handler:             handleSubscribe,
			Category:            "mqtt",
			Subcategory:         "subscribe",
			Tags:                []string{"mqtt", "subscribe", "topic", "listen"},
			UseCases:            []string{"Listen to MQTT topic", "Monitor IoT devices"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "mqtt",
		},
		{
			Tool: mcp.NewTool("aftrs_mqtt_messages",
				mcp.WithDescription("Get received messages from subscribed topics."),
				mcp.WithString("topic", mcp.Description("Filter by topic (optional)")),
				mcp.WithNumber("limit", mcp.Description("Maximum messages to return (default: 20)")),
			),
			Handler:             handleMessages,
			Category:            "mqtt",
			Subcategory:         "messages",
			Tags:                []string{"mqtt", "messages", "receive", "read"},
			UseCases:            []string{"Read MQTT messages", "View received data"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "mqtt",
		},
	}
}

// handleStatus handles the aftrs_mqtt_status tool
func handleStatus(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	status, err := client.GetStatus(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# MQTT Status\n\n")

	if status.Connected {
		sb.WriteString("**Status:** Connected\n")
	} else {
		sb.WriteString("**Status:** Disconnected\n")
		sb.WriteString("\nUse `aftrs_mqtt_connect` to connect to a broker.\n")
		return tools.TextResult(sb.String()), nil
	}

	sb.WriteString(fmt.Sprintf("**Broker:** %s\n", status.Broker))
	sb.WriteString(fmt.Sprintf("**Client ID:** %s\n", status.ClientID))

	if len(status.Subscriptions) > 0 {
		sb.WriteString("\n## Subscriptions\n\n")
		for _, topic := range status.Subscriptions {
			sb.WriteString(fmt.Sprintf("- `%s`\n", topic))
		}
	} else {
		sb.WriteString("\n*No active subscriptions*\n")
	}

	return tools.TextResult(sb.String()), nil
}

// handleConnect handles the aftrs_mqtt_connect tool
func handleConnect(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	// Check if already connected
	if client.IsConnected() {
		return tools.TextResult("Already connected to MQTT broker."), nil
	}

	if err := client.Connect(ctx); err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to connect: %v", err)), nil
	}

	status, _ := client.GetStatus(ctx)

	var sb strings.Builder
	sb.WriteString("# Connected to MQTT Broker\n\n")
	sb.WriteString(fmt.Sprintf("**Broker:** %s\n", status.Broker))
	sb.WriteString(fmt.Sprintf("**Client ID:** %s\n", status.ClientID))

	return tools.TextResult(sb.String()), nil
}

// handlePublish handles the aftrs_mqtt_publish tool
func handlePublish(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	topic, errResult := tools.RequireStringParam(req, "topic")
	if errResult != nil {
		return errResult, nil
	}

	message, errResult := tools.RequireStringParam(req, "message")
	if errResult != nil {
		return errResult, nil
	}

	retain := tools.GetBoolParam(req, "retain", false)

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	if !client.IsConnected() {
		return tools.ErrorResult(fmt.Errorf("not connected to MQTT broker - use aftrs_mqtt_connect first")), nil
	}

	result, err := client.Publish(ctx, topic, message, 0, retain)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# MQTT Publish\n\n")

	if result.Success {
		sb.WriteString("**Status:** Published\n")
		sb.WriteString(fmt.Sprintf("**Topic:** `%s`\n", result.Topic))
		sb.WriteString(fmt.Sprintf("**Message:** %s\n", message))
		if retain {
			sb.WriteString("**Retained:** Yes\n")
		}
	} else {
		sb.WriteString("**Status:** Failed\n")
		sb.WriteString(fmt.Sprintf("**Error:** %s\n", result.Error))
	}

	return tools.TextResult(sb.String()), nil
}

// handleSubscribe handles the aftrs_mqtt_subscribe tool
func handleSubscribe(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	topic, errResult := tools.RequireStringParam(req, "topic")
	if errResult != nil {
		return errResult, nil
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	if !client.IsConnected() {
		return tools.ErrorResult(fmt.Errorf("not connected to MQTT broker - use aftrs_mqtt_connect first")), nil
	}

	if err := client.Subscribe(ctx, topic, 0); err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# Subscribed to Topic\n\n")
	sb.WriteString(fmt.Sprintf("**Topic:** `%s`\n", topic))
	sb.WriteString("\nMessages will be collected. Use `aftrs_mqtt_messages` to view received messages.\n")

	return tools.TextResult(sb.String()), nil
}

// handleMessages handles the aftrs_mqtt_messages tool
func handleMessages(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	topic := tools.GetStringParam(req, "topic")
	limit := tools.GetIntParam(req, "limit", 20)

	var messages []clients.MQTTMessage
	if topic != "" {
		messages = client.GetMessages(topic, limit)
	} else {
		// Get messages from all topics
		for _, t := range client.ListTopics() {
			messages = append(messages, client.GetMessages(t, 0)...)
		}
		// Apply limit
		if len(messages) > limit {
			messages = messages[len(messages)-limit:]
		}
	}

	var sb strings.Builder
	sb.WriteString("# MQTT Messages\n\n")

	if len(messages) == 0 {
		sb.WriteString("No messages received.\n")
		if !client.IsConnected() {
			sb.WriteString("\n*Not connected to broker.*\n")
		}
		return tools.TextResult(sb.String()), nil
	}

	sb.WriteString(fmt.Sprintf("Showing **%d** messages:\n\n", len(messages)))

	for i, msg := range messages {
		sb.WriteString(fmt.Sprintf("### Message %d\n", i+1))
		sb.WriteString(fmt.Sprintf("**Topic:** `%s`\n", msg.Topic))
		sb.WriteString(fmt.Sprintf("**Time:** %s\n", msg.Timestamp.Format("15:04:05")))
		sb.WriteString(fmt.Sprintf("**Payload:**\n```\n%s\n```\n\n", msg.Payload))
	}

	return tools.TextResult(sb.String()), nil
}

// init registers this module with the global registry
func init() {
	tools.GetRegistry().RegisterModule(&Module{})
}
