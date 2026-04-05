// Package clients provides API clients for external services.
package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// MQTTClient provides MQTT messaging capabilities
type MQTTClient struct {
	client        mqtt.Client
	broker        string
	clientID      string
	mu            sync.RWMutex
	subscriptions map[string][]MQTTMessage
	connected     bool
}

// MQTTConfig holds MQTT connection configuration
type MQTTConfig struct {
	Broker   string `json:"broker"`
	Port     int    `json:"port"`
	ClientID string `json:"client_id"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	UseTLS   bool   `json:"use_tls"`
}

// MQTTMessage represents a received MQTT message
type MQTTMessage struct {
	Topic     string    `json:"topic"`
	Payload   string    `json:"payload"`
	QoS       byte      `json:"qos"`
	Retained  bool      `json:"retained"`
	Timestamp time.Time `json:"timestamp"`
}

// MQTTStatus represents MQTT connection status
type MQTTStatus struct {
	Connected     bool     `json:"connected"`
	Broker        string   `json:"broker"`
	ClientID      string   `json:"client_id"`
	Subscriptions []string `json:"subscriptions"`
}

// MQTTPublishResult represents the result of a publish operation
type MQTTPublishResult struct {
	Success bool   `json:"success"`
	Topic   string `json:"topic"`
	QoS     byte   `json:"qos"`
	Error   string `json:"error,omitempty"`
}

// NewMQTTClient creates a new MQTT client
func NewMQTTClient() (*MQTTClient, error) {
	broker := os.Getenv("MQTT_BROKER")
	if broker == "" {
		broker = "tcp://localhost:1883"
	}

	clientID := os.Getenv("MQTT_CLIENT_ID")
	if clientID == "" {
		clientID = fmt.Sprintf("hg-mcp-%d", time.Now().UnixNano())
	}

	return &MQTTClient{
		broker:        broker,
		clientID:      clientID,
		subscriptions: make(map[string][]MQTTMessage),
	}, nil
}

// Connect establishes connection to the MQTT broker
func (c *MQTTClient) Connect(ctx context.Context) error {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(c.broker)
	opts.SetClientID(c.clientID)

	// Add credentials if available
	if username := os.Getenv("MQTT_USERNAME"); username != "" {
		opts.SetUsername(username)
		if password := os.Getenv("MQTT_PASSWORD"); password != "" {
			opts.SetPassword(password)
		}
	}

	opts.SetAutoReconnect(true)
	opts.SetConnectRetry(true)
	opts.SetConnectRetryInterval(5 * time.Second)
	opts.SetKeepAlive(30 * time.Second)

	opts.SetConnectionLostHandler(func(client mqtt.Client, err error) {
		c.mu.Lock()
		c.connected = false
		c.mu.Unlock()
	})

	opts.SetOnConnectHandler(func(client mqtt.Client) {
		c.mu.Lock()
		c.connected = true
		c.mu.Unlock()
	})

	c.client = mqtt.NewClient(opts)

	token := c.client.Connect()
	if !token.WaitTimeout(10 * time.Second) {
		return fmt.Errorf("connection timeout")
	}
	if err := token.Error(); err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}

	c.mu.Lock()
	c.connected = true
	c.mu.Unlock()

	return nil
}

// Disconnect closes the MQTT connection
func (c *MQTTClient) Disconnect() {
	if c.client != nil && c.client.IsConnected() {
		c.client.Disconnect(1000)
	}
	c.mu.Lock()
	c.connected = false
	c.mu.Unlock()
}

// IsConnected checks if connected to broker
func (c *MQTTClient) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.connected && c.client != nil && c.client.IsConnected()
}

// GetStatus returns MQTT connection status
func (c *MQTTClient) GetStatus(ctx context.Context) (*MQTTStatus, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	subs := []string{}
	for topic := range c.subscriptions {
		subs = append(subs, topic)
	}

	return &MQTTStatus{
		Connected:     c.connected && c.client != nil && c.client.IsConnected(),
		Broker:        c.broker,
		ClientID:      c.clientID,
		Subscriptions: subs,
	}, nil
}

// Publish sends a message to a topic
func (c *MQTTClient) Publish(ctx context.Context, topic string, payload interface{}, qos byte, retained bool) (*MQTTPublishResult, error) {
	result := &MQTTPublishResult{
		Topic: topic,
		QoS:   qos,
	}

	if !c.IsConnected() {
		result.Error = "not connected to broker"
		return result, nil
	}

	// Convert payload to bytes
	var payloadBytes []byte
	switch p := payload.(type) {
	case string:
		payloadBytes = []byte(p)
	case []byte:
		payloadBytes = p
	default:
		var err error
		payloadBytes, err = json.Marshal(p)
		if err != nil {
			result.Error = fmt.Sprintf("failed to marshal payload: %v", err)
			return result, nil
		}
	}

	token := c.client.Publish(topic, qos, retained, payloadBytes)
	if !token.WaitTimeout(5 * time.Second) {
		result.Error = "publish timeout"
		return result, nil
	}
	if err := token.Error(); err != nil {
		result.Error = err.Error()
		return result, nil
	}

	result.Success = true
	return result, nil
}

// Subscribe subscribes to a topic
func (c *MQTTClient) Subscribe(ctx context.Context, topic string, qos byte) error {
	if !c.IsConnected() {
		return fmt.Errorf("not connected to broker")
	}

	callback := func(client mqtt.Client, msg mqtt.Message) {
		c.mu.Lock()
		defer c.mu.Unlock()

		message := MQTTMessage{
			Topic:     msg.Topic(),
			Payload:   string(msg.Payload()),
			QoS:       msg.Qos(),
			Retained:  msg.Retained(),
			Timestamp: time.Now(),
		}

		// Keep last 100 messages per topic
		if len(c.subscriptions[msg.Topic()]) >= 100 {
			c.subscriptions[msg.Topic()] = c.subscriptions[msg.Topic()][1:]
		}
		c.subscriptions[msg.Topic()] = append(c.subscriptions[msg.Topic()], message)
	}

	token := c.client.Subscribe(topic, qos, callback)
	if !token.WaitTimeout(5 * time.Second) {
		return fmt.Errorf("subscribe timeout")
	}
	if err := token.Error(); err != nil {
		return fmt.Errorf("subscribe failed: %w", err)
	}

	c.mu.Lock()
	if _, exists := c.subscriptions[topic]; !exists {
		c.subscriptions[topic] = []MQTTMessage{}
	}
	c.mu.Unlock()

	return nil
}

// Unsubscribe removes a topic subscription
func (c *MQTTClient) Unsubscribe(ctx context.Context, topic string) error {
	if !c.IsConnected() {
		return fmt.Errorf("not connected to broker")
	}

	token := c.client.Unsubscribe(topic)
	if !token.WaitTimeout(5 * time.Second) {
		return fmt.Errorf("unsubscribe timeout")
	}
	if err := token.Error(); err != nil {
		return fmt.Errorf("unsubscribe failed: %w", err)
	}

	c.mu.Lock()
	delete(c.subscriptions, topic)
	c.mu.Unlock()

	return nil
}

// GetMessages returns received messages for a topic
func (c *MQTTClient) GetMessages(topic string, limit int) []MQTTMessage {
	c.mu.RLock()
	defer c.mu.RUnlock()

	messages := []MQTTMessage{}

	// Exact match
	if msgs, ok := c.subscriptions[topic]; ok {
		messages = append(messages, msgs...)
	}

	// Wildcard match
	for t, msgs := range c.subscriptions {
		if matchTopic(topic, t) {
			messages = append(messages, msgs...)
		}
	}

	// Apply limit
	if limit > 0 && len(messages) > limit {
		messages = messages[len(messages)-limit:]
	}

	return messages
}

// ListTopics returns all subscribed topics
func (c *MQTTClient) ListTopics() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	topics := make([]string, 0, len(c.subscriptions))
	for topic := range c.subscriptions {
		topics = append(topics, topic)
	}
	return topics
}

// matchTopic checks if a topic matches a subscription pattern
func matchTopic(pattern, topic string) bool {
	// Handle # wildcard (multi-level)
	if strings.HasSuffix(pattern, "/#") {
		prefix := strings.TrimSuffix(pattern, "/#")
		return strings.HasPrefix(topic, prefix)
	}

	// Handle + wildcard (single-level)
	patternParts := strings.Split(pattern, "/")
	topicParts := strings.Split(topic, "/")

	if len(patternParts) != len(topicParts) {
		return false
	}

	for i := range patternParts {
		if patternParts[i] == "+" {
			continue
		}
		if patternParts[i] != topicParts[i] {
			return false
		}
	}

	return true
}

// IoTDevice represents a discovered IoT device
type IoTDevice struct {
	ID         string            `json:"id"`
	Name       string            `json:"name"`
	Type       string            `json:"type"`
	Topics     []string          `json:"topics"`
	LastSeen   time.Time         `json:"last_seen"`
	Attributes map[string]string `json:"attributes,omitempty"`
}

// DiscoverDevices attempts to discover IoT devices on common topics
func (c *MQTTClient) DiscoverDevices(ctx context.Context) ([]IoTDevice, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("not connected to broker")
	}

	devices := []IoTDevice{}

	// Common discovery topics
	discoveryTopics := []string{
		"homeassistant/+/+/config",   // Home Assistant discovery
		"tasmota/discovery/+/config", // Tasmota devices
		"zigbee2mqtt/bridge/devices", // Zigbee2MQTT
		"wled/+",                     // WLED devices
	}

	// Subscribe to discovery topics briefly
	for _, topic := range discoveryTopics {
		c.Subscribe(ctx, topic, 0)
	}

	// Wait briefly for messages
	time.Sleep(2 * time.Second)

	// Process received messages
	c.mu.RLock()
	for topic, messages := range c.subscriptions {
		for _, msg := range messages {
			// Try to parse as device info
			var data map[string]interface{}
			if err := json.Unmarshal([]byte(msg.Payload), &data); err != nil {
				continue
			}

			device := IoTDevice{
				Topics:   []string{topic},
				LastSeen: msg.Timestamp,
			}

			// Extract common fields
			if id, ok := data["unique_id"].(string); ok {
				device.ID = id
			} else if id, ok := data["mac"].(string); ok {
				device.ID = id
			}

			if name, ok := data["name"].(string); ok {
				device.Name = name
			} else if name, ok := data["device_name"].(string); ok {
				device.Name = name
			}

			if deviceType, ok := data["device_class"].(string); ok {
				device.Type = deviceType
			} else if deviceType, ok := data["type"].(string); ok {
				device.Type = deviceType
			}

			if device.ID != "" || device.Name != "" {
				devices = append(devices, device)
			}
		}
	}
	c.mu.RUnlock()

	// Unsubscribe from discovery topics
	for _, topic := range discoveryTopics {
		c.Unsubscribe(ctx, topic)
	}

	return devices, nil
}
