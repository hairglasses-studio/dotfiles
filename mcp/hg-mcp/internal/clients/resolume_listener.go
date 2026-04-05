// Package clients provides API clients for external services.
package clients

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/hypebeast/go-osc/osc"
)

// ResolumeOSCListener listens for OSC messages from Resolume
// for bidirectional communication and state tracking
type ResolumeOSCListener struct {
	port    int
	conn    *net.UDPConn
	running bool
	mu      sync.RWMutex

	// Callbacks for different message types
	onBPMChange      func(bpm float64)
	onClipTrigger    func(layer, column int)
	onLayerChange    func(layer int, opacity float64)
	onMasterChange   func(level float64)
	onTransportState func(playing bool)

	// State tracking
	lastBPM     float64
	lastMaster  float64
	activeClips map[string]bool // "layer:column" -> playing
	layerStates map[int]float64 // layer -> opacity

	// Message handler
	dispatcher *osc.StandardDispatcher
}

// NewResolumeOSCListener creates a new OSC listener for Resolume feedback
func NewResolumeOSCListener(port int) *ResolumeOSCListener {
	l := &ResolumeOSCListener{
		port:        port,
		activeClips: make(map[string]bool),
		layerStates: make(map[int]float64),
		dispatcher:  osc.NewStandardDispatcher(),
	}

	// Register default handlers
	l.setupDefaultHandlers()

	return l
}

// setupDefaultHandlers registers OSC address handlers
func (l *ResolumeOSCListener) setupDefaultHandlers() {
	// BPM feedback
	l.dispatcher.AddMsgHandler("/composition/tempocontroller/tempo", func(msg *osc.Message) {
		if len(msg.Arguments) > 0 {
			if bpm, ok := msg.Arguments[0].(float32); ok {
				l.mu.Lock()
				l.lastBPM = float64(bpm)
				l.mu.Unlock()

				if l.onBPMChange != nil {
					l.onBPMChange(float64(bpm))
				}
			}
		}
	})

	// Master level feedback
	l.dispatcher.AddMsgHandler("/composition/video/opacity/values", func(msg *osc.Message) {
		if len(msg.Arguments) > 0 {
			if level, ok := msg.Arguments[0].(float32); ok {
				l.mu.Lock()
				l.lastMaster = float64(level)
				l.mu.Unlock()

				if l.onMasterChange != nil {
					l.onMasterChange(float64(level))
				}
			}
		}
	})

	// Catch-all for clip triggers - pattern: /composition/layers/*/clips/*/connect
	l.dispatcher.AddMsgHandler("*", func(msg *osc.Message) {
		// Parse layer/clip triggers
		var layer, column int
		if n, _ := fmt.Sscanf(msg.Address, "/composition/layers/%d/clips/%d/connect", &layer, &column); n == 2 {
			if l.onClipTrigger != nil {
				l.onClipTrigger(layer, column)
			}

			// Track active state
			key := fmt.Sprintf("%d:%d", layer, column)
			l.mu.Lock()
			if len(msg.Arguments) > 0 {
				if val, ok := msg.Arguments[0].(int32); ok {
					l.activeClips[key] = val == 1
				}
			}
			l.mu.Unlock()
		}

		// Parse layer opacity changes
		if n, _ := fmt.Sscanf(msg.Address, "/composition/layers/%d/video/opacity/values", &layer); n == 1 {
			if len(msg.Arguments) > 0 {
				if opacity, ok := msg.Arguments[0].(float32); ok {
					l.mu.Lock()
					l.layerStates[layer] = float64(opacity)
					l.mu.Unlock()

					if l.onLayerChange != nil {
						l.onLayerChange(layer, float64(opacity))
					}
				}
			}
		}
	})
}

// Start begins listening for OSC messages
func (l *ResolumeOSCListener) Start(ctx context.Context) error {
	l.mu.Lock()
	if l.running {
		l.mu.Unlock()
		return fmt.Errorf("listener already running")
	}
	l.mu.Unlock()

	addr := fmt.Sprintf("0.0.0.0:%d", l.port)
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return fmt.Errorf("resolve address: %w", err)
	}

	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return fmt.Errorf("listen UDP: %w", err)
	}

	l.mu.Lock()
	l.conn = conn
	l.running = true
	l.mu.Unlock()

	// Start receive loop
	go l.receiveLoop(ctx)

	return nil
}

// receiveLoop reads OSC messages from the UDP connection
func (l *ResolumeOSCListener) receiveLoop(ctx context.Context) {
	buf := make([]byte, 4096)

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		// Set read deadline to allow context checking
		l.conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))

		n, _, err := l.conn.ReadFromUDP(buf)
		if err != nil {
			// Timeout is expected for context checking
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue
			}

			l.mu.RLock()
			running := l.running
			l.mu.RUnlock()

			if !running {
				return
			}
			continue
		}

		// Parse and dispatch OSC message
		packet, err := osc.ParsePacket(string(buf[:n]))
		if err != nil {
			continue
		}

		switch p := packet.(type) {
		case *osc.Message:
			l.dispatcher.Dispatch(p)
		case *osc.Bundle:
			for _, msg := range p.Messages {
				l.dispatcher.Dispatch(msg)
			}
		}
	}
}

// Stop stops the OSC listener
func (l *ResolumeOSCListener) Stop() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if !l.running {
		return nil
	}

	l.running = false
	if l.conn != nil {
		l.conn.Close()
		l.conn = nil
	}

	return nil
}

// IsRunning returns whether the listener is currently running
func (l *ResolumeOSCListener) IsRunning() bool {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.running
}

// Port returns the listening port
func (l *ResolumeOSCListener) Port() int {
	return l.port
}

// OnBPMChange sets a callback for BPM changes
func (l *ResolumeOSCListener) OnBPMChange(callback func(bpm float64)) {
	l.onBPMChange = callback
}

// OnClipTrigger sets a callback for clip triggers
func (l *ResolumeOSCListener) OnClipTrigger(callback func(layer, column int)) {
	l.onClipTrigger = callback
}

// OnLayerChange sets a callback for layer opacity changes
func (l *ResolumeOSCListener) OnLayerChange(callback func(layer int, opacity float64)) {
	l.onLayerChange = callback
}

// OnMasterChange sets a callback for master level changes
func (l *ResolumeOSCListener) OnMasterChange(callback func(level float64)) {
	l.onMasterChange = callback
}

// OnTransportState sets a callback for transport state changes
func (l *ResolumeOSCListener) OnTransportState(callback func(playing bool)) {
	l.onTransportState = callback
}

// GetLastBPM returns the last received BPM value
func (l *ResolumeOSCListener) GetLastBPM() float64 {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.lastBPM
}

// GetLastMasterLevel returns the last received master level
func (l *ResolumeOSCListener) GetLastMasterLevel() float64 {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.lastMaster
}

// GetActiveClips returns a copy of the active clips map
func (l *ResolumeOSCListener) GetActiveClips() map[string]bool {
	l.mu.RLock()
	defer l.mu.RUnlock()

	result := make(map[string]bool, len(l.activeClips))
	for k, v := range l.activeClips {
		result[k] = v
	}
	return result
}

// GetLayerStates returns a copy of the layer states map
func (l *ResolumeOSCListener) GetLayerStates() map[int]float64 {
	l.mu.RLock()
	defer l.mu.RUnlock()

	result := make(map[int]float64, len(l.layerStates))
	for k, v := range l.layerStates {
		result[k] = v
	}
	return result
}

// AddHandler adds a custom OSC message handler for a specific address pattern
func (l *ResolumeOSCListener) AddHandler(address string, handler func(*osc.Message)) {
	l.dispatcher.AddMsgHandler(address, handler)
}

// ListenerStatus represents the current state of the OSC listener
type ListenerStatus struct {
	Running     bool            `json:"running"`
	Port        int             `json:"port"`
	LastBPM     float64         `json:"last_bpm"`
	LastMaster  float64         `json:"last_master"`
	ActiveClips int             `json:"active_clips"`
	LayerStates map[int]float64 `json:"layer_states"`
}

// GetStatus returns the current listener status
func (l *ResolumeOSCListener) GetStatus() *ListenerStatus {
	l.mu.RLock()
	defer l.mu.RUnlock()

	activeCount := 0
	for _, active := range l.activeClips {
		if active {
			activeCount++
		}
	}

	states := make(map[int]float64, len(l.layerStates))
	for k, v := range l.layerStates {
		states[k] = v
	}

	return &ListenerStatus{
		Running:     l.running,
		Port:        l.port,
		LastBPM:     l.lastBPM,
		LastMaster:  l.lastMaster,
		ActiveClips: activeCount,
		LayerStates: states,
	}
}
