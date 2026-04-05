package daemon

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"path/filepath"
	"sync"
	"time"

	"github.com/hairglasses-studio/mapping"
	"github.com/hairglasses-studio/mapitall/internal/ipc"
	"github.com/hairglasses-studio/mapitall/internal/output"
	"github.com/hairglasses-studio/mapitall/internal/pipeline"
	"github.com/hairglasses-studio/mapitall/internal/reload"
	"github.com/hairglasses-studio/mapitall/internal/service"
	"github.com/hairglasses-studio/mcpkit/device"
)

// Daemon is the core runtime that ties everything together.
type Daemon struct {
	cfg       *Config
	providers []device.DeviceProvider
	registry  *output.Registry
	pipeline  *pipeline.Pipeline
	watcher   *reload.Watcher
	ipcServer *ipc.Server
	eventBus  *ipc.EventBus
	state     *mapping.EngineState
	startTime time.Time

	mu       sync.RWMutex
	profiles map[string]*profileState // device ID -> compiled profile
	devices  []device.Info            // currently connected devices
}

type profileState struct {
	profile *mapping.MappingProfile
	index   *mapping.RuleIndex
}

// New creates a daemon from config.
func New(cfg *Config) (*Daemon, error) {
	d := &Daemon{
		cfg:      cfg,
		state:    mapping.NewEngineState(),
		profiles: make(map[string]*profileState),
		registry: output.NewRegistry(),
		eventBus: ipc.NewEventBus(),
	}

	// Register all output targets.
	d.registry.Register(output.NewCommandTarget())
	d.registry.Register(output.NewOSCTarget())
	d.registry.Register(output.NewKeyTarget())
	d.registry.Register(output.NewMovementTarget())
	d.registry.Register(output.NewWebSocketTarget())
	d.registry.Register(output.NewMIDITarget())
	d.registry.Register(output.NewDBusTarget())
	d.registry.Register(output.NewVariableTarget(d.state))
	d.registry.Register(output.NewToggleVarTarget(d.state))
	d.registry.Register(output.NewLayerTarget(d.state))
	d.registry.Register(output.NewSequenceTarget(d.registry))

	return d, nil
}

// Run starts the daemon and blocks until ctx is cancelled.
func (d *Daemon) Run(ctx context.Context) error {
	// Load profiles from all configured directories.
	if err := d.loadProfiles(); err != nil {
		return fmt.Errorf("load profiles: %w", err)
	}

	// Initialize device providers.
	d.providers = device.PlatformProviders()
	slog.Info("initialized providers", "count", len(d.providers))

	// Start file watcher for hot reload.
	d.watcher = reload.New(d.cfg.ProfileDirs, func() {
		slog.Info("profile change detected, reloading")
		if err := d.loadProfiles(); err != nil {
			slog.Error("reload profiles", "error", err)
		}
	})
	if err := d.watcher.Start(ctx); err != nil {
		slog.Warn("file watcher failed to start", "error", err)
	}

	// Create pipeline with custom modifiers from loaded profiles.
	d.pipeline = pipeline.New(d.state, d.registry, d.resolveRule, d.collectCustomModifiers())
	d.pipeline.SetObserver(d.publishEvent)

	// Start IPC server.
	d.ipcServer = ipc.NewServer(d.cfg.SocketPath)
	d.registerIPCHandlers()
	go func() {
		if err := d.ipcServer.Start(ctx); err != nil {
			slog.Warn("IPC server error", "error", err)
		}
	}()

	// Open devices and feed events into the pipeline.
	var wg sync.WaitGroup
	for _, prov := range d.providers {
		devs, err := prov.Enumerate(ctx)
		if err != nil {
			slog.Warn("enumerate devices", "provider", prov.Name(), "error", err)
			continue
		}
		for _, info := range devs {
			if !d.deviceAllowed(info) {
				continue
			}
			conn, err := prov.Open(ctx, info.ID)
			if err != nil {
				slog.Warn("open device", "device", info.Name, "error", err)
				continue
			}
			slog.Info("opened device", "name", info.Name, "type", info.Type, "id", info.ID)

			if err := conn.Start(ctx); err != nil {
				slog.Warn("start device", "device", info.Name, "error", err)
				conn.Close()
				continue
			}

			// Grab device if profile requests it and connection supports it.
			if grabbable, ok := conn.(device.Grabbable); ok && d.shouldGrab(info) {
				if err := grabbable.Grab(); err != nil {
					slog.Warn("device grab failed", "device", info.Name, "error", err)
				} else {
					slog.Info("device grabbed", "device", info.Name)
				}
			}

			d.mu.Lock()
			d.devices = append(d.devices, info)
			d.mu.Unlock()

			wg.Add(1)
			go func(c device.DeviceConnection) {
				defer wg.Done()
				defer c.Close()
				d.pipeline.ProcessEvents(ctx, c.Events())
			}(conn)
		}
	}

	// Start hot-plug monitoring for device connect/disconnect.
	go d.watchHotPlug(ctx, &wg)

	// Start active-app tracking for app-specific overrides.
	d.startActiveAppTracker(ctx)

	d.startTime = time.Now()
	service.NotifyReady()
	slog.Info("mapitall running", "profiles", len(d.profiles), "profile_dirs", d.cfg.ProfileDirs)

	// Block until shutdown.
	<-ctx.Done()
	service.NotifyStopping()
	slog.Info("shutting down")

	// Shut down in order: event bus, IPC, watcher, devices, providers.
	if d.eventBus != nil {
		d.eventBus.Close()
	}
	if d.ipcServer != nil {
		d.ipcServer.Close()
	}
	if d.watcher != nil {
		d.watcher.Stop()
	}
	wg.Wait()
	d.registry.Close()
	for _, prov := range d.providers {
		prov.Close()
	}

	return nil
}

// loadProfiles scans all profile dirs and builds rule indices.
func (d *Daemon) loadProfiles() error {
	newProfiles := make(map[string]*profileState)

	for _, dir := range d.cfg.ProfileDirs {
		summaries, err := mapping.ListMappingProfiles(dir)
		if err != nil {
			slog.Warn("list profiles", "dir", dir, "error", err)
			continue
		}
		for _, s := range summaries {
			if s.Format == "error" {
				slog.Warn("skipping bad profile", "name", s.Name, "error", s.Error)
				continue
			}
			p, err := mapping.LoadMappingProfile(s.Path)
			if err != nil {
				slog.Warn("load profile", "path", s.Path, "error", err)
				continue
			}
			// Convert legacy to unified for the runtime.
			if p.IsLegacyFormat() {
				p = mapping.ConvertLegacyToUnified(p)
			}
			idx := mapping.BuildRuleIndex(p)
			deviceName := p.DeviceName()
			newProfiles[deviceName] = &profileState{profile: p, index: idx}
			slog.Debug("loaded profile", "name", s.Name, "device", deviceName, "mappings", p.MappingCount())
		}
	}

	d.mu.Lock()
	d.profiles = newProfiles
	d.mu.Unlock()

	slog.Info("profiles loaded", "count", len(newProfiles))
	return nil
}

// resolveRule finds the best matching rule for a device event.
func (d *Daemon) resolveRule(deviceID string, source string) *mapping.MappingRule {
	d.mu.RLock()
	defer d.mu.RUnlock()

	// Try exact device ID match first, then scan all profiles.
	if ps, ok := d.profiles[deviceID]; ok {
		if rule := ps.index.Resolve(source, d.state, deviceID); rule != nil {
			return rule
		}
	}

	// Fallback: try all profiles (for when device names don't match exactly).
	for _, ps := range d.profiles {
		if rule := ps.index.Resolve(source, d.state, deviceID); rule != nil {
			return rule
		}
	}

	return nil
}

// registerIPCHandlers wires IPC methods to daemon state.
func (d *Daemon) registerIPCHandlers() {
	d.ipcServer.Handle("status", func(_ json.RawMessage) (any, error) {
		d.mu.RLock()
		defer d.mu.RUnlock()
		return map[string]any{
			"uptime_seconds": time.Since(d.startTime).Seconds(),
			"profiles":       len(d.profiles),
			"devices":        len(d.devices),
			"profile_dirs":   d.cfg.ProfileDirs,
		}, nil
	})

	d.ipcServer.Handle("reload", func(_ json.RawMessage) (any, error) {
		if err := d.loadProfiles(); err != nil {
			return nil, err
		}
		d.mu.RLock()
		count := len(d.profiles)
		d.mu.RUnlock()
		return map[string]any{"reloaded": true, "profiles": count}, nil
	})

	d.ipcServer.Handle("list_devices", func(_ json.RawMessage) (any, error) {
		d.mu.RLock()
		defer d.mu.RUnlock()
		devList := make([]map[string]any, len(d.devices))
		for i, info := range d.devices {
			devList[i] = map[string]any{
				"id":   info.ID,
				"name": info.Name,
				"type": info.Type.String(),
			}
		}
		return devList, nil
	})

	d.ipcServer.Handle("list_profiles", func(_ json.RawMessage) (any, error) {
		d.mu.RLock()
		defer d.mu.RUnlock()
		names := make([]string, 0, len(d.profiles))
		for name := range d.profiles {
			names = append(names, name)
		}
		return names, nil
	})

	d.ipcServer.Handle("get_state", func(_ json.RawMessage) (any, error) {
		vars := make(map[string]any)
		d.state.RangeVariables(func(k string, v any) bool {
			vars[k] = v
			return true
		})
		return map[string]any{
			"active_app": d.state.GetActiveApp(),
			"variables":  vars,
		}, nil
	})

	d.ipcServer.Handle("set_variable", func(params json.RawMessage) (any, error) {
		var req struct {
			Name  string `json:"name"`
			Value any    `json:"value"`
		}
		if err := json.Unmarshal(params, &req); err != nil {
			return nil, err
		}
		d.state.SetVariable(req.Name, req.Value)
		return map[string]any{"set": true}, nil
	})

	// subscribe_events streams device events as JSON-RPC notifications until disconnect.
	d.ipcServer.HandleStream("subscribe_events", func(params json.RawMessage, enc *json.Encoder, done <-chan struct{}) error {
		var filter ipc.EventFilter
		if len(params) > 0 {
			json.Unmarshal(params, &filter)
		}

		events, unsub := d.eventBus.Subscribe(filter)
		defer unsub()

		for {
			select {
			case <-done:
				return nil
			case ev, ok := <-events:
				if !ok {
					return nil
				}
				if err := enc.Encode(ipc.Notification{
					JSONRPC: "2.0",
					Method:  "event",
					Params:  ev,
				}); err != nil {
					return err
				}
			}
		}
	})
}

// watchHotPlug monitors for device connect/disconnect and starts/stops pipelines.
func (d *Daemon) watchHotPlug(ctx context.Context, wg *sync.WaitGroup) {
	// Poll all providers every 3 seconds for new devices.
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			d.checkForNewDevices(ctx, wg)
		}
	}
}

// checkForNewDevices enumerates all providers and starts pipelines for new devices.
func (d *Daemon) checkForNewDevices(ctx context.Context, wg *sync.WaitGroup) {
	d.mu.RLock()
	known := make(map[device.DeviceID]bool, len(d.devices))
	for _, info := range d.devices {
		known[info.ID] = true
	}
	d.mu.RUnlock()

	for _, prov := range d.providers {
		devs, err := prov.Enumerate(ctx)
		if err != nil {
			continue
		}
		for _, info := range devs {
			if known[info.ID] || !d.deviceAllowed(info) {
				continue
			}

			conn, err := prov.Open(ctx, info.ID)
			if err != nil {
				slog.Warn("hot-plug open", "device", info.Name, "error", err)
				continue
			}
			if err := conn.Start(ctx); err != nil {
				slog.Warn("hot-plug start", "device", info.Name, "error", err)
				conn.Close()
				continue
			}

			if grabbable, ok := conn.(device.Grabbable); ok && d.shouldGrab(info) {
				if err := grabbable.Grab(); err != nil {
					slog.Warn("hot-plug grab", "device", info.Name, "error", err)
				}
			}

			d.mu.Lock()
			d.devices = append(d.devices, info)
			d.mu.Unlock()

			slog.Info("hot-plug connected", "name", info.Name, "type", info.Type, "id", info.ID)

			wg.Add(1)
			go func(c device.DeviceConnection) {
				defer wg.Done()
				defer c.Close()
				d.pipeline.ProcessEvents(ctx, c.Events())
			}(conn)
		}
	}
}

// shouldGrab checks if any loaded profile wants to grab this device.
func (d *Daemon) shouldGrab(info device.Info) bool {
	d.mu.RLock()
	defer d.mu.RUnlock()

	deviceName := string(info.ID)
	if ps, ok := d.profiles[deviceName]; ok {
		return ps.profile.Settings != nil && ps.profile.Settings.GrabDevice
	}
	// Check by device name too.
	if ps, ok := d.profiles[info.Name]; ok {
		return ps.profile.Settings != nil && ps.profile.Settings.GrabDevice
	}
	return false
}

// Reload triggers a profile reload. Exported for use by signal handlers.
func (d *Daemon) Reload() error {
	slog.Info("profile reload triggered")
	if err := d.loadProfiles(); err != nil {
		return err
	}
	// Rebuild pipeline with potentially updated custom modifiers.
	d.pipeline = pipeline.New(d.state, d.registry, d.resolveRule, d.collectCustomModifiers())
	d.pipeline.SetObserver(d.publishEvent)
	return nil
}

// publishEvent converts a device.Event to an IPC DeviceEvent and publishes it.
func (d *Daemon) publishEvent(ev device.Event) {
	if d.eventBus.SubscriberCount() == 0 {
		return // fast path: no subscribers
	}
	d.eventBus.Publish(ipc.DeviceEvent{
		DeviceID:  string(ev.DeviceID),
		Type:      ev.Type.String(),
		Source:    ev.Source,
		Timestamp: ev.Timestamp,
		Value:     ev.Value,
		Pressed:   ev.Pressed,
		Code:      ev.Code,
		Channel:   ev.Channel,
		MIDINote:  ev.MIDINote,
		Velocity:  ev.Velocity,
		MIDIValue: ev.MIDIValue,
	})
}

// collectCustomModifiers gathers all custom modifier inputs from loaded profiles.
func (d *Daemon) collectCustomModifiers() []string {
	d.mu.RLock()
	defer d.mu.RUnlock()

	seen := make(map[string]bool)
	for _, ps := range d.profiles {
		if ps.profile.Settings != nil {
			for _, mod := range ps.profile.Settings.CustomModifiers {
				seen[mod] = true
			}
		}
	}

	mods := make([]string, 0, len(seen))
	for mod := range seen {
		mods = append(mods, mod)
	}
	return mods
}

// deviceAllowed checks if a device passes the filter.
func (d *Daemon) deviceAllowed(info device.Info) bool {
	if len(d.cfg.DeviceFilter) == 0 {
		return true // no filter = accept all
	}
	for _, pattern := range d.cfg.DeviceFilter {
		matched, _ := filepath.Match(pattern, info.Name)
		if matched {
			return true
		}
	}
	return false
}
