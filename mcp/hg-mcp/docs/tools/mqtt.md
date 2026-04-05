# mqtt

> MQTT messaging and IoT device integration

**5 tools**

## Tools

- [`aftrs_mqtt_connect`](#aftrs-mqtt-connect)
- [`aftrs_mqtt_messages`](#aftrs-mqtt-messages)
- [`aftrs_mqtt_publish`](#aftrs-mqtt-publish)
- [`aftrs_mqtt_status`](#aftrs-mqtt-status)
- [`aftrs_mqtt_subscribe`](#aftrs-mqtt-subscribe)

---

## aftrs_mqtt_connect

Connect to MQTT broker.

**Complexity:** simple

**Tags:** `mqtt`, `connect`, `broker`

**Use Cases:**
- Connect to MQTT broker
- Establish connection

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `broker` | string |  | Broker URL (default: tcp://localhost:1883) |

### Example

```json
{
  "broker": "example"
}
```

---

## aftrs_mqtt_messages

Get received messages from subscribed topics.

**Complexity:** simple

**Tags:** `mqtt`, `messages`, `receive`, `read`

**Use Cases:**
- Read MQTT messages
- View received data

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `limit` | number |  | Maximum messages to return (default: 20) |
| `topic` | string |  | Filter by topic (optional) |

### Example

```json
{
  "limit": 0,
  "topic": "example"
}
```

---

## aftrs_mqtt_publish

Publish a message to an MQTT topic.

**Complexity:** simple

**Tags:** `mqtt`, `publish`, `message`, `iot`

**Use Cases:**
- Send MQTT message
- Control IoT device

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `message` | string | Yes | Message payload |
| `retain` | boolean |  | Retain message on broker |
| `topic` | string | Yes | Topic to publish to |

### Example

```json
{
  "message": "example",
  "retain": false,
  "topic": "example"
}
```

---

## aftrs_mqtt_status

Get MQTT broker connection status.

**Complexity:** simple

**Tags:** `mqtt`, `iot`, `status`, `broker`

**Use Cases:**
- Check MQTT connection
- View subscriptions

---

## aftrs_mqtt_subscribe

Subscribe to an MQTT topic.

**Complexity:** simple

**Tags:** `mqtt`, `subscribe`, `topic`, `listen`

**Use Cases:**
- Listen to MQTT topic
- Monitor IoT devices

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `topic` | string | Yes | Topic to subscribe to (supports + and # wildcards) |

### Example

```json
{
  "topic": "example"
}
```

---

