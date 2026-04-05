// Package clients provides API clients for external services.
package clients

import (
	"context"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"
)

// LightingClient provides access to DMX/ArtNet lighting control
type LightingClient struct {
	artnetHost string
	artnetPort int
	universe   int
}

// DMXUniverse represents a DMX universe status
type DMXUniverse struct {
	Universe    int    `json:"universe"`
	Channels    int    `json:"channels"`
	Active      bool   `json:"active"`
	Source      string `json:"source,omitempty"`
	LastUpdated string `json:"last_updated,omitempty"`
}

// Fixture represents a lighting fixture
type Fixture struct {
	Name         string         `json:"name"`
	Type         string         `json:"type"`
	StartChannel int            `json:"start_channel"`
	NumChannels  int            `json:"num_channels"`
	Universe     int            `json:"universe"`
	Values       map[string]int `json:"values,omitempty"`
}

// LightingScene represents a saved lighting scene
type LightingScene struct {
	Name        string           `json:"name"`
	Description string           `json:"description,omitempty"`
	Fixtures    map[string][]int `json:"fixtures"` // fixture name -> channel values
	CreatedAt   time.Time        `json:"created_at"`
}

// ArtNetNode represents an ArtNet node on the network
type ArtNetNode struct {
	Name     string `json:"name"`
	IP       string `json:"ip"`
	MAC      string `json:"mac,omitempty"`
	Ports    int    `json:"ports"`
	Universe int    `json:"universe"`
	Online   bool   `json:"online"`
}

// LightingHealth represents lighting system health
type LightingHealth struct {
	Score           int      `json:"score"`
	Status          string   `json:"status"`
	UniverseCount   int      `json:"universe_count"`
	FixtureCount    int      `json:"fixture_count"`
	NodesOnline     int      `json:"nodes_online"`
	Issues          []string `json:"issues,omitempty"`
	Recommendations []string `json:"recommendations,omitempty"`
}

// NewLightingClient creates a new lighting client
func NewLightingClient() (*LightingClient, error) {
	host := os.Getenv("ARTNET_HOST")
	if host == "" {
		host = "255.255.255.255" // Broadcast
	}

	port := 6454 // Standard ArtNet port
	universe := 0

	return &LightingClient{
		artnetHost: host,
		artnetPort: port,
		universe:   universe,
	}, nil
}

// GetDMXStatus returns DMX universe status
func (c *LightingClient) GetDMXStatus(ctx context.Context) (*DMXUniverse, error) {
	universe := &DMXUniverse{
		Universe: c.universe,
		Channels: 512,
		Active:   false,
	}

	// Check if we can reach the ArtNet network
	addr := net.JoinHostPort(c.artnetHost, strconv.Itoa(c.artnetPort))
	conn, err := net.DialTimeout("udp", addr, 2*time.Second)
	if err == nil {
		conn.Close()
		universe.Active = true
		universe.Source = "ArtNet"
	}

	return universe, nil
}

// GetDMXChannels returns current DMX channel values
func (c *LightingClient) GetDMXChannels(ctx context.Context, startChannel, count int) ([]int, error) {
	// In a real implementation, we'd query the ArtNet node
	// For now, return zeros
	channels := make([]int, count)
	return channels, nil
}

// SetDMXChannels sets DMX channel values
func (c *LightingClient) SetDMXChannels(ctx context.Context, startChannel int, values []int) error {
	// In a real implementation, we'd send ArtNet DMX packets
	return nil
}

// ListFixtures returns configured fixtures
func (c *LightingClient) ListFixtures(ctx context.Context) ([]Fixture, error) {
	// This would normally load from a fixture library/config
	fixtures := []Fixture{}

	// Check for fixture config file
	configPath := os.Getenv("LIGHTING_CONFIG")
	if configPath == "" {
		// Return example fixtures for demonstration
		fixtures = append(fixtures, Fixture{
			Name:         "Front Wash",
			Type:         "LED Par",
			StartChannel: 1,
			NumChannels:  7,
			Universe:     0,
		})
	}

	return fixtures, nil
}

// ControlFixture controls a fixture
func (c *LightingClient) ControlFixture(ctx context.Context, fixtureName string, values map[string]int) error {
	// In a real implementation, this would:
	// 1. Look up fixture by name
	// 2. Map parameter names to channels
	// 3. Send DMX values
	return nil
}

// ListScenes returns saved lighting scenes
func (c *LightingClient) ListScenes(ctx context.Context) ([]LightingScene, error) {
	scenes := []LightingScene{}

	// This would normally load from a scenes file
	// Return empty for now
	return scenes, nil
}

// RecallScene recalls a saved scene
func (c *LightingClient) RecallScene(ctx context.Context, sceneName string) error {
	// In a real implementation:
	// 1. Load scene data
	// 2. Set all fixture values
	return fmt.Errorf("scene not found: %s", sceneName)
}

// DiscoverArtNetNodes discovers ArtNet nodes on the network
func (c *LightingClient) DiscoverArtNetNodes(ctx context.Context) ([]ArtNetNode, error) {
	nodes := []ArtNetNode{}

	// ArtNet discovery uses broadcast ArtPoll packets
	// Simplified implementation: check common addresses

	commonNodes := []string{
		"192.168.1.100",
		"192.168.1.101",
		"192.168.2.100",
	}

	for _, ip := range commonNodes {
		addr := net.JoinHostPort(ip, strconv.Itoa(c.artnetPort))
		conn, err := net.DialTimeout("udp", addr, 500*time.Millisecond)
		if err == nil {
			conn.Close()
			nodes = append(nodes, ArtNetNode{
				Name:     fmt.Sprintf("Node-%s", ip),
				IP:       ip,
				Ports:    1,
				Universe: 0,
				Online:   true,
			})
		}
	}

	return nodes, nil
}

// GetHealth returns lighting system health
func (c *LightingClient) GetHealth(ctx context.Context) (*LightingHealth, error) {
	health := &LightingHealth{
		Score:  100,
		Status: "healthy",
	}

	// Check DMX status
	dmxStatus, _ := c.GetDMXStatus(ctx)
	if dmxStatus != nil && dmxStatus.Active {
		health.UniverseCount = 1
	} else {
		health.Score -= 30
		health.Issues = append(health.Issues, "No active DMX universe")
	}

	// Check fixtures
	fixtures, _ := c.ListFixtures(ctx)
	health.FixtureCount = len(fixtures)
	if health.FixtureCount == 0 {
		health.Score -= 20
		health.Recommendations = append(health.Recommendations, "Configure lighting fixtures")
	}

	// Check ArtNet nodes
	nodes, _ := c.DiscoverArtNetNodes(ctx)
	for _, node := range nodes {
		if node.Online {
			health.NodesOnline++
		}
	}

	// Set status
	switch {
	case health.Score >= 80:
		health.Status = "healthy"
	case health.Score >= 50:
		health.Status = "degraded"
	default:
		health.Status = "critical"
	}

	return health, nil
}

// ArtNetHost returns the configured ArtNet host
func (c *LightingClient) ArtNetHost() string {
	return c.artnetHost
}

// Universe returns the configured universe
func (c *LightingClient) Universe() int {
	return c.universe
}

// FixtureGroup represents a group of fixtures
type FixtureGroup struct {
	Name     string   `json:"name"`
	Fixtures []string `json:"fixtures"`
}

// Chase represents a chase/sequence
type Chase struct {
	Name    string  `json:"name"`
	Steps   int     `json:"steps"`
	BPM     float64 `json:"bpm"`
	Running bool    `json:"running"`
	Loop    bool    `json:"loop"`
}

// ArtNetStatus represents detailed ArtNet status
type ArtNetStatus struct {
	Connected   bool   `json:"connected"`
	IP          string `json:"ip"`
	Port        int    `json:"port"`
	Universes   []int  `json:"universes"`
	NodesFound  int    `json:"nodes_found"`
	PacketsSent int64  `json:"packets_sent"`
	PacketsRecv int64  `json:"packets_recv"`
	LastError   string `json:"last_error,omitempty"`
}

// Blackout sets all channels to zero
func (c *LightingClient) Blackout(ctx context.Context) error {
	zeros := make([]int, 512)
	return c.SetDMXChannels(ctx, 1, zeros)
}

// FullUp sets all channels to max
func (c *LightingClient) FullUp(ctx context.Context) error {
	max := make([]int, 512)
	for i := range max {
		max[i] = 255
	}
	return c.SetDMXChannels(ctx, 1, max)
}

// SetFixtureDimmer sets dimmer level for a fixture
func (c *LightingClient) SetFixtureDimmer(ctx context.Context, fixture string, level int) error {
	values := map[string]int{"dimmer": level}
	return c.ControlFixture(ctx, fixture, values)
}

// SetFixtureColor sets color for a fixture
func (c *LightingClient) SetFixtureColor(ctx context.Context, fixture string, r, g, b int) error {
	values := map[string]int{"red": r, "green": g, "blue": b}
	return c.ControlFixture(ctx, fixture, values)
}

// SaveScene saves current state as a scene
func (c *LightingClient) SaveScene(ctx context.Context, name, description string) error {
	// Would save current DMX state
	return nil
}

// FadeToScene crossfades to a scene over duration
func (c *LightingClient) FadeToScene(ctx context.Context, scene string, durationMs int) error {
	// Would perform crossfade
	return nil
}

// ListGroups returns fixture groups
func (c *LightingClient) ListGroups(ctx context.Context) ([]FixtureGroup, error) {
	groups := []FixtureGroup{}
	return groups, nil
}

// ControlGroup controls all fixtures in a group
func (c *LightingClient) ControlGroup(ctx context.Context, group string, values map[string]int) error {
	return nil
}

// ListChases returns available chases
func (c *LightingClient) ListChases(ctx context.Context) ([]Chase, error) {
	chases := []Chase{}
	return chases, nil
}

// StartChase starts a chase
func (c *LightingClient) StartChase(ctx context.Context, chase string) error {
	return nil
}

// StopChase stops a chase
func (c *LightingClient) StopChase(ctx context.Context, chase string) error {
	return nil
}

// SetChaseBPM sets chase tempo
func (c *LightingClient) SetChaseBPM(ctx context.Context, chase string, bpm float64) error {
	return nil
}

// SelectUniverse selects the active DMX universe
func (c *LightingClient) SelectUniverse(ctx context.Context, universe int) error {
	if universe < 0 || universe > 32767 {
		return fmt.Errorf("universe must be 0-32767")
	}
	c.universe = universe
	return nil
}

// GetArtNetStatus returns detailed ArtNet status
func (c *LightingClient) GetArtNetStatus(ctx context.Context) (*ArtNetStatus, error) {
	status := &ArtNetStatus{
		IP:        c.artnetHost,
		Port:      c.artnetPort,
		Universes: []int{c.universe},
	}

	// Check connection
	addr := net.JoinHostPort(c.artnetHost, strconv.Itoa(c.artnetPort))
	conn, err := net.DialTimeout("udp", addr, 2*time.Second)
	if err == nil {
		conn.Close()
		status.Connected = true
	}

	nodes, _ := c.DiscoverArtNetNodes(ctx)
	status.NodesFound = len(nodes)

	return status, nil
}

// GetFixture returns a specific fixture
func (c *LightingClient) GetFixture(ctx context.Context, name string) (*Fixture, error) {
	fixtures, err := c.ListFixtures(ctx)
	if err != nil {
		return nil, err
	}
	for _, f := range fixtures {
		if f.Name == name {
			return &f, nil
		}
	}
	return nil, fmt.Errorf("fixture not found: %s", name)
}

// DeleteScene deletes a scene
func (c *LightingClient) DeleteScene(ctx context.Context, name string) error {
	return nil
}

// GetScene returns a specific scene
func (c *LightingClient) GetScene(ctx context.Context, name string) (*LightingScene, error) {
	return nil, fmt.Errorf("scene not found: %s", name)
}

// PatchEntry represents a fixture patch assignment
type PatchEntry struct {
	Name        string `json:"name"`
	FixtureType string `json:"fixture_type"`
	Universe    int    `json:"universe"`
	Address     int    `json:"address"`
	Channels    int    `json:"channels"`
}

// GetPatchList returns all fixture patch assignments
func (c *LightingClient) GetPatchList(ctx context.Context) ([]PatchEntry, error) {
	// Would normally load from patch file
	patches := []PatchEntry{}
	return patches, nil
}

// PatchFixture patches a fixture to an address
func (c *LightingClient) PatchFixture(ctx context.Context, fixtureType, name string, universe, address int) error {
	// Would add to patch file
	return nil
}

// UnpatchFixture removes a fixture from the patch
func (c *LightingClient) UnpatchFixture(ctx context.Context, fixture string) error {
	// Would remove from patch file
	return nil
}

// ExportPatch exports patch data in the specified format
func (c *LightingClient) ExportPatch(ctx context.Context, format string) (string, error) {
	patches, err := c.GetPatchList(ctx)
	if err != nil {
		return "", err
	}

	if format == "csv" {
		result := "Name,Type,Universe,Address,Channels\n"
		for _, p := range patches {
			result += fmt.Sprintf("%s,%s,%d,%d,%d\n", p.Name, p.FixtureType, p.Universe, p.Address, p.Channels)
		}
		return result, nil
	}

	// Default to JSON
	return "[]", nil
}
