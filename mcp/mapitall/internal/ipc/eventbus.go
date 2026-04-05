package ipc

import (
	"sync"
	"time"
)

// DeviceEvent is a JSON-serializable event sent to IPC subscribers.
// It mirrors the fields of device.Event that are useful for monitoring.
type DeviceEvent struct {
	DeviceID  string    `json:"device_id"`
	Type      string    `json:"type"`       // "button", "axis", "hat", "midi_note", etc.
	Source    string    `json:"source"`      // Canonical ID: "BTN_SOUTH", "ABS_X", "midi:cc:1"
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value,omitempty"`
	Pressed   bool      `json:"pressed,omitempty"`
	Code      uint16    `json:"code,omitempty"`
	Channel   uint8     `json:"channel,omitempty"`
	MIDINote  uint8     `json:"midi_note,omitempty"`
	Velocity  uint8     `json:"velocity,omitempty"`
	MIDIValue uint8     `json:"midi_value,omitempty"`
}

// EventFilter controls which events a subscriber receives.
type EventFilter struct {
	DeviceID  string `json:"device_id,omitempty"`
	EventType string `json:"event_type,omitempty"` // "button", "axis", etc.
}

// subscriber is a registered event listener.
type subscriber struct {
	ch     chan DeviceEvent
	filter EventFilter
}

// EventBus provides a fan-out mechanism for distributing device events
// to multiple IPC subscribers. It is goroutine-safe.
type EventBus struct {
	mu          sync.RWMutex
	subscribers map[uint64]*subscriber
	nextID      uint64
}

// NewEventBus creates a new event bus.
func NewEventBus() *EventBus {
	return &EventBus{
		subscribers: make(map[uint64]*subscriber),
	}
}

// Subscribe registers a new subscriber with an optional filter.
// Returns a channel that receives matching events and an unsubscribe function.
// The channel has a buffer of 64 events; slow consumers will have events dropped.
func (b *EventBus) Subscribe(filter EventFilter) (<-chan DeviceEvent, func()) {
	ch := make(chan DeviceEvent, 64)
	b.mu.Lock()
	id := b.nextID
	b.nextID++
	b.subscribers[id] = &subscriber{ch: ch, filter: filter}
	b.mu.Unlock()

	unsub := func() {
		b.mu.Lock()
		delete(b.subscribers, id)
		b.mu.Unlock()
		// Drain remaining events so senders don't block.
		for range ch {
		}
	}
	return ch, unsub
}

// Publish sends an event to all matching subscribers.
// Non-blocking: if a subscriber's channel is full, the event is dropped for that subscriber.
func (b *EventBus) Publish(ev DeviceEvent) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	for _, sub := range b.subscribers {
		if !matchesFilter(ev, sub.filter) {
			continue
		}
		select {
		case sub.ch <- ev:
		default:
			// Slow consumer, drop event.
		}
	}
}

// SubscriberCount returns the number of active subscribers.
func (b *EventBus) SubscriberCount() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.subscribers)
}

// Close closes all subscriber channels. Call during shutdown.
func (b *EventBus) Close() {
	b.mu.Lock()
	defer b.mu.Unlock()
	for id, sub := range b.subscribers {
		close(sub.ch)
		delete(b.subscribers, id)
	}
}

func matchesFilter(ev DeviceEvent, f EventFilter) bool {
	if f.DeviceID != "" && ev.DeviceID != f.DeviceID {
		return false
	}
	if f.EventType != "" && ev.Type != f.EventType {
		return false
	}
	return true
}
