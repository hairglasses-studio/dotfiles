// Package clients provides API clients for external services.
package clients

import (
	"context"
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"
)

// TraktorClient provides access to Traktor Pro library
type TraktorClient struct {
	collectionPath string
	collection     *TraktorCollection
}

// TraktorCollection represents the root of collection.nml
type TraktorCollection struct {
	XMLName      xml.Name         `xml:"NML"`
	Version      string           `xml:"VERSION,attr"`
	Head         TraktorHead      `xml:"HEAD"`
	MusicFolders []TraktorFolder  `xml:"MUSICFOLDERS>FOLDER"`
	Collection   TraktorEntries   `xml:"COLLECTION"`
	Playlists    TraktorPlaylists `xml:"PLAYLISTS"`
}

// TraktorHead contains header info
type TraktorHead struct {
	Company string `xml:"COMPANY,attr"`
	Program string `xml:"PROGRAM,attr"`
}

// TraktorFolder represents a music folder
type TraktorFolder struct {
	Path string `xml:"PATH,attr"`
}

// TraktorEntries contains all tracks
type TraktorEntries struct {
	Entries int            `xml:"ENTRIES,attr"`
	Tracks  []TraktorTrack `xml:"ENTRY"`
}

// TraktorTrack represents a track in the collection
type TraktorTrack struct {
	Artist       string          `xml:"ARTIST,attr"`
	Title        string          `xml:"TITLE,attr"`
	Album        TraktorAlbum    `xml:"ALBUM"`
	Location     TraktorLocation `xml:"LOCATION"`
	Info         TraktorInfo     `xml:"INFO"`
	Tempo        TraktorTempo    `xml:"TEMPO"`
	Loudness     TraktorLoudness `xml:"LOUDNESS"`
	MusicalKey   TraktorKey      `xml:"MUSICAL_KEY"`
	CuePoints    []TraktorCue    `xml:"CUE_V2"`
	Loops        []TraktorLoop   `xml:"LOOP"`
	ModifiedDate string          `xml:"MODIFIED_DATE,attr"`
	ModifiedTime string          `xml:"MODIFIED_TIME,attr"`
}

// TraktorAlbum contains album info
type TraktorAlbum struct {
	Title string `xml:"TITLE,attr"`
	Track int    `xml:"TRACK,attr"`
}

// TraktorLocation contains file location
type TraktorLocation struct {
	Dir      string `xml:"DIR,attr"`
	File     string `xml:"FILE,attr"`
	Volume   string `xml:"VOLUME,attr"`
	VolumeID string `xml:"VOLUMEID,attr"`
}

// TraktorInfo contains track metadata
type TraktorInfo struct {
	Bitrate       int     `xml:"BITRATE,attr"`
	Genre         string  `xml:"GENRE,attr"`
	Label         string  `xml:"LABEL,attr"`
	Comment       string  `xml:"COMMENT,attr"`
	Rating        int     `xml:"RANKING,attr"`
	PlayCount     int     `xml:"PLAYCOUNT,attr"`
	LastPlayed    string  `xml:"LAST_PLAYED,attr"`
	ImportDate    string  `xml:"IMPORT_DATE,attr"`
	ReleaseDate   string  `xml:"RELEASE_DATE,attr"`
	Playtime      int     `xml:"PLAYTIME,attr"`
	PlaytimeFloat float64 `xml:"PLAYTIME_FLOAT,attr"`
}

// TraktorTempo contains BPM info
type TraktorTempo struct {
	BPM        float64 `xml:"BPM,attr"`
	BPMQuality float64 `xml:"BPM_QUALITY,attr"`
}

// TraktorLoudness contains loudness info
type TraktorLoudness struct {
	PeakDB      float64 `xml:"PEAK_DB,attr"`
	PerceivedDB float64 `xml:"PERCEIVED_DB,attr"`
	AnalyzedDB  float64 `xml:"ANALYZED_DB,attr"`
}

// TraktorKey contains musical key info
type TraktorKey struct {
	Value int `xml:"VALUE,attr"`
}

// TraktorCue represents a cue point
type TraktorCue struct {
	Name    string  `xml:"NAME,attr"`
	Type    int     `xml:"TYPE,attr"`
	Start   float64 `xml:"START,attr"`
	Length  float64 `xml:"LEN,attr"`
	Repeats int     `xml:"REPEATS,attr"`
	Hotcue  int     `xml:"HOTCUE,attr"`
}

// TraktorLoop represents a saved loop
type TraktorLoop struct {
	Name   string  `xml:"NAME,attr"`
	Start  float64 `xml:"START,attr"`
	Length float64 `xml:"LEN,attr"`
}

// TraktorPlaylists contains playlist hierarchy
type TraktorPlaylists struct {
	Nodes []TraktorNode `xml:"NODE"`
}

// TraktorNode represents a playlist or folder
type TraktorNode struct {
	Type     string                 `xml:"TYPE,attr"`
	Name     string                 `xml:"NAME,attr"`
	Subnodes []TraktorNode          `xml:"SUBNODES>NODE"`
	Playlist TraktorPlaylistEntries `xml:"PLAYLIST"`
}

// TraktorPlaylistEntries contains playlist tracks
type TraktorPlaylistEntries struct {
	Entries int                    `xml:"ENTRIES,attr"`
	Type    string                 `xml:"TYPE,attr"`
	UUID    string                 `xml:"UUID,attr"`
	Tracks  []TraktorPlaylistTrack `xml:"ENTRY"`
}

// TraktorPlaylistTrack represents a track reference in a playlist
type TraktorPlaylistTrack struct {
	Key TraktorPrimaryKey `xml:"PRIMARYKEY"`
}

// TraktorPrimaryKey identifies a track
type TraktorPrimaryKey struct {
	Type string `xml:"TYPE,attr"`
	Key  string `xml:"KEY,attr"`
}

// Track represents a simplified track for API responses
type Track struct {
	Title      string            `json:"title"`
	Artist     string            `json:"artist"`
	Album      string            `json:"album,omitempty"`
	Genre      string            `json:"genre,omitempty"`
	Label      string            `json:"label,omitempty"`
	BPM        float64           `json:"bpm,omitempty"`
	Key        string            `json:"key,omitempty"`
	Duration   int               `json:"duration_seconds,omitempty"`
	Rating     int               `json:"rating,omitempty"`
	PlayCount  int               `json:"play_count,omitempty"`
	LastPlayed string            `json:"last_played,omitempty"`
	FilePath   string            `json:"file_path"`
	CuePoints  []TraktorCuePoint `json:"cue_points,omitempty"`
	Loops      []Loop            `json:"loops,omitempty"`
}

// TraktorCuePoint represents a cue point in Traktor
type TraktorCuePoint struct {
	Name     string  `json:"name"`
	Position float64 `json:"position_ms"`
	Type     string  `json:"type"`
	Hotcue   int     `json:"hotcue,omitempty"`
}

// Loop represents a saved loop
type Loop struct {
	Name   string  `json:"name"`
	Start  float64 `json:"start_ms"`
	Length float64 `json:"length_ms"`
}

// Playlist represents a playlist
type Playlist struct {
	Name       string  `json:"name"`
	TrackCount int     `json:"track_count"`
	Tracks     []Track `json:"tracks,omitempty"`
}

// TraktorStatus represents library status
type TraktorStatus struct {
	Connected      bool   `json:"connected"`
	CollectionPath string `json:"collection_path"`
	TrackCount     int    `json:"track_count"`
	PlaylistCount  int    `json:"playlist_count"`
	Version        string `json:"version"`
}

// TraktorHealth represents health status
type TraktorHealth struct {
	Score           int      `json:"score"`
	Status          string   `json:"status"`
	Connected       bool     `json:"connected"`
	CollectionFound bool     `json:"collection_found"`
	Issues          []string `json:"issues,omitempty"`
	Recommendations []string `json:"recommendations,omitempty"`
}

// NewTraktorClient creates a new Traktor client
func NewTraktorClient() (*TraktorClient, error) {
	collectionPath := os.Getenv("TRAKTOR_COLLECTION_PATH")
	if collectionPath == "" {
		collectionPath = getDefaultTraktorPath()
	}

	return &TraktorClient{
		collectionPath: collectionPath,
	}, nil
}

// getDefaultTraktorPath returns the default Traktor collection path
func getDefaultTraktorPath() string {
	home, _ := os.UserHomeDir()

	// Find latest Traktor version
	var basePath string
	switch runtime.GOOS {
	case "darwin":
		basePath = filepath.Join(home, "Documents", "Native Instruments")
	case "windows":
		basePath = filepath.Join(home, "Documents", "Native Instruments")
	default:
		basePath = filepath.Join(home, "Documents", "Native Instruments")
	}

	// Look for Traktor directories
	entries, err := os.ReadDir(basePath)
	if err != nil {
		return filepath.Join(basePath, "Traktor Pro 3", "collection.nml")
	}

	var latestVersion string
	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), "Traktor") {
			if entry.Name() > latestVersion {
				latestVersion = entry.Name()
			}
		}
	}

	if latestVersion == "" {
		latestVersion = "Traktor Pro 3"
	}

	return filepath.Join(basePath, latestVersion, "collection.nml")
}

// loadCollection loads and parses the collection.nml file
func (c *TraktorClient) loadCollection() error {
	if c.collection != nil {
		return nil
	}

	data, err := os.ReadFile(c.collectionPath)
	if err != nil {
		return fmt.Errorf("failed to read collection: %w", err)
	}

	var collection TraktorCollection
	if err := xml.Unmarshal(data, &collection); err != nil {
		return fmt.Errorf("failed to parse collection: %w", err)
	}

	c.collection = &collection
	return nil
}

// GetStatus returns library status
func (c *TraktorClient) GetStatus(ctx context.Context) (*TraktorStatus, error) {
	status := &TraktorStatus{
		Connected:      false,
		CollectionPath: c.collectionPath,
	}

	if err := c.loadCollection(); err != nil {
		return status, nil
	}

	status.Connected = true
	status.TrackCount = c.collection.Collection.Entries
	status.PlaylistCount = countPlaylists(c.collection.Playlists.Nodes)
	status.Version = c.collection.Version

	return status, nil
}

// countPlaylists recursively counts playlists
func countPlaylists(nodes []TraktorNode) int {
	count := 0
	for _, node := range nodes {
		if node.Type == "PLAYLIST" {
			count++
		}
		count += countPlaylists(node.Subnodes)
	}
	return count
}

// SearchLibrary searches for tracks matching a query
func (c *TraktorClient) SearchLibrary(ctx context.Context, query string, limit int) ([]Track, error) {
	if err := c.loadCollection(); err != nil {
		return nil, err
	}

	query = strings.ToLower(query)
	var results []Track

	for _, t := range c.collection.Collection.Tracks {
		if strings.Contains(strings.ToLower(t.Title), query) ||
			strings.Contains(strings.ToLower(t.Artist), query) ||
			strings.Contains(strings.ToLower(t.Album.Title), query) ||
			strings.Contains(strings.ToLower(t.Info.Genre), query) {

			results = append(results, convertTrack(t))

			if limit > 0 && len(results) >= limit {
				break
			}
		}
	}

	return results, nil
}

// GetTrack returns a specific track with full details
func (c *TraktorClient) GetTrack(ctx context.Context, filePath string) (*Track, error) {
	if err := c.loadCollection(); err != nil {
		return nil, err
	}

	for _, t := range c.collection.Collection.Tracks {
		trackPath := buildFilePath(t.Location)
		if trackPath == filePath || strings.HasSuffix(trackPath, filePath) {
			track := convertTrack(t)
			return &track, nil
		}
	}

	return nil, fmt.Errorf("track not found: %s", filePath)
}

// GetPlaylists returns all playlists
func (c *TraktorClient) GetPlaylists(ctx context.Context) ([]Playlist, error) {
	if err := c.loadCollection(); err != nil {
		return nil, err
	}

	var playlists []Playlist
	collectPlaylists(c.collection.Playlists.Nodes, "", &playlists)
	return playlists, nil
}

// collectPlaylists recursively collects playlists
func collectPlaylists(nodes []TraktorNode, prefix string, playlists *[]Playlist) {
	for _, node := range nodes {
		name := node.Name
		if prefix != "" {
			name = prefix + "/" + name
		}

		if node.Type == "PLAYLIST" {
			*playlists = append(*playlists, Playlist{
				Name:       name,
				TrackCount: node.Playlist.Entries,
			})
		}

		collectPlaylists(node.Subnodes, name, playlists)
	}
}

// GetPlaylist returns a specific playlist with tracks
func (c *TraktorClient) GetPlaylist(ctx context.Context, name string) (*Playlist, error) {
	if err := c.loadCollection(); err != nil {
		return nil, err
	}

	node := findPlaylistNode(c.collection.Playlists.Nodes, name, "")
	if node == nil {
		return nil, fmt.Errorf("playlist not found: %s", name)
	}

	playlist := &Playlist{
		Name:       name,
		TrackCount: node.Playlist.Entries,
		Tracks:     make([]Track, 0),
	}

	// Build track lookup
	trackMap := make(map[string]TraktorTrack)
	for _, t := range c.collection.Collection.Tracks {
		key := buildFilePath(t.Location)
		trackMap[key] = t
	}

	// Resolve playlist tracks
	for _, pt := range node.Playlist.Tracks {
		if track, ok := trackMap[pt.Key.Key]; ok {
			playlist.Tracks = append(playlist.Tracks, convertTrack(track))
		}
	}

	return playlist, nil
}

// findPlaylistNode finds a playlist by name
func findPlaylistNode(nodes []TraktorNode, name, prefix string) *TraktorNode {
	for i, node := range nodes {
		fullName := node.Name
		if prefix != "" {
			fullName = prefix + "/" + node.Name
		}

		if node.Type == "PLAYLIST" && (fullName == name || node.Name == name) {
			return &nodes[i]
		}

		if found := findPlaylistNode(node.Subnodes, name, fullName); found != nil {
			return found
		}
	}
	return nil
}

// GetHistory returns recently played tracks
func (c *TraktorClient) GetHistory(ctx context.Context, limit int) ([]Track, error) {
	if err := c.loadCollection(); err != nil {
		return nil, err
	}

	// Collect tracks with play history
	var played []TraktorTrack
	for _, t := range c.collection.Collection.Tracks {
		if t.Info.LastPlayed != "" || t.Info.PlayCount > 0 {
			played = append(played, t)
		}
	}

	// Sort by last played date
	sort.Slice(played, func(i, j int) bool {
		return played[i].Info.LastPlayed > played[j].Info.LastPlayed
	})

	// Apply limit
	if limit > 0 && len(played) > limit {
		played = played[:limit]
	}

	// Convert to Track
	var tracks []Track
	for _, t := range played {
		tracks = append(tracks, convertTrack(t))
	}

	return tracks, nil
}

// GetCuePoints returns cue points for a track
func (c *TraktorClient) GetCuePoints(ctx context.Context, filePath string) ([]TraktorCuePoint, error) {
	track, err := c.GetTrack(ctx, filePath)
	if err != nil {
		return nil, err
	}
	return track.CuePoints, nil
}

// GetLoops returns loops for a track
func (c *TraktorClient) GetLoops(ctx context.Context, filePath string) ([]Loop, error) {
	track, err := c.GetTrack(ctx, filePath)
	if err != nil {
		return nil, err
	}
	return track.Loops, nil
}

// GetHealth returns health status
func (c *TraktorClient) GetHealth(ctx context.Context) (*TraktorHealth, error) {
	health := &TraktorHealth{
		Score:  100,
		Status: "healthy",
	}

	// Check if collection exists
	if _, err := os.Stat(c.collectionPath); err != nil {
		health.Score = 0
		health.Status = "critical"
		health.CollectionFound = false
		health.Issues = append(health.Issues, fmt.Sprintf("Collection not found at: %s", c.collectionPath))
		health.Recommendations = append(health.Recommendations,
			"Set TRAKTOR_COLLECTION_PATH environment variable")
		health.Recommendations = append(health.Recommendations,
			"Ensure Traktor Pro has been run at least once")
		return health, nil
	}

	health.CollectionFound = true

	// Try to load collection
	if err := c.loadCollection(); err != nil {
		health.Score = 30
		health.Status = "degraded"
		health.Issues = append(health.Issues, fmt.Sprintf("Failed to parse collection: %v", err))
		return health, nil
	}

	health.Connected = true
	return health, nil
}

// ExportToRekordboxXML exports the collection to Rekordbox XML format
func (c *TraktorClient) ExportToRekordboxXML(ctx context.Context, outputPath string) error {
	if err := c.loadCollection(); err != nil {
		return err
	}

	// Create Rekordbox XML structure
	xml := `<?xml version="1.0" encoding="UTF-8"?>
<DJ_PLAYLISTS Version="1.0.0">
  <PRODUCT Name="Traktor Export" Version="1.0" Company="AFTRS"/>
  <COLLECTION Entries="%d">
%s  </COLLECTION>
  <PLAYLISTS>
%s  </PLAYLISTS>
</DJ_PLAYLISTS>`

	// Build track entries
	var trackEntries strings.Builder
	for i, t := range c.collection.Collection.Tracks {
		trackEntries.WriteString(fmt.Sprintf(`    <TRACK TrackID="%d" Name="%s" Artist="%s" Album="%s" Genre="%s" Kind="MP3 File" Size="0" TotalTime="%d" AverageBpm="%.2f" DateAdded="%s" Location="file://localhost%s"/>
`,
			i+1,
			escapeXML(t.Title),
			escapeXML(t.Artist),
			escapeXML(t.Album.Title),
			escapeXML(t.Info.Genre),
			t.Info.Playtime,
			t.Tempo.BPM,
			t.Info.ImportDate,
			escapeXML(buildFilePath(t.Location)),
		))
	}

	// Build playlist entries
	var playlistEntries strings.Builder
	playlistEntries.WriteString(`    <NODE Type="0" Name="ROOT" Count="1">
`)
	buildRekordboxPlaylists(c.collection.Playlists.Nodes, &playlistEntries, 3)
	playlistEntries.WriteString(`    </NODE>
`)

	output := fmt.Sprintf(xml,
		c.collection.Collection.Entries,
		trackEntries.String(),
		playlistEntries.String(),
	)

	return os.WriteFile(outputPath, []byte(output), 0644)
}

// buildRekordboxPlaylists recursively builds playlist XML
func buildRekordboxPlaylists(nodes []TraktorNode, sb *strings.Builder, indent int) {
	prefix := strings.Repeat("  ", indent)
	for _, node := range nodes {
		if node.Type == "FOLDER" {
			sb.WriteString(fmt.Sprintf(`%s<NODE Type="0" Name="%s" Count="%d">
`, prefix, escapeXML(node.Name), len(node.Subnodes)))
			buildRekordboxPlaylists(node.Subnodes, sb, indent+1)
			sb.WriteString(fmt.Sprintf(`%s</NODE>
`, prefix))
		} else if node.Type == "PLAYLIST" {
			sb.WriteString(fmt.Sprintf(`%s<NODE Type="1" Name="%s" KeyType="0" Entries="%d">
`, prefix, escapeXML(node.Name), node.Playlist.Entries))
			for i, t := range node.Playlist.Tracks {
				sb.WriteString(fmt.Sprintf(`%s  <TRACK Key="%d"/>
`, prefix, i+1))
				_ = t // Suppress unused warning
			}
			sb.WriteString(fmt.Sprintf(`%s</NODE>
`, prefix))
		}
	}
}

// escapeXML escapes special XML characters
func escapeXML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&apos;")
	return s
}

// convertTrack converts a TraktorTrack to Track
func convertTrack(t TraktorTrack) Track {
	track := Track{
		Title:      t.Title,
		Artist:     t.Artist,
		Album:      t.Album.Title,
		Genre:      t.Info.Genre,
		Label:      t.Info.Label,
		BPM:        t.Tempo.BPM,
		Key:        traktorKeyToString(t.MusicalKey.Value),
		Duration:   t.Info.Playtime,
		Rating:     t.Info.Rating,
		PlayCount:  t.Info.PlayCount,
		LastPlayed: t.Info.LastPlayed,
		FilePath:   buildFilePath(t.Location),
	}

	// Convert cue points
	for _, cue := range t.CuePoints {
		track.CuePoints = append(track.CuePoints, TraktorCuePoint{
			Name:     cue.Name,
			Position: cue.Start,
			Type:     cueTypeToString(cue.Type),
			Hotcue:   cue.Hotcue,
		})
	}

	// Convert loops
	for _, loop := range t.Loops {
		track.Loops = append(track.Loops, Loop{
			Name:   loop.Name,
			Start:  loop.Start,
			Length: loop.Length,
		})
	}

	return track
}

// buildFilePath constructs file path from location
func buildFilePath(loc TraktorLocation) string {
	// Traktor uses /:/ as path separator on macOS
	dir := strings.ReplaceAll(loc.Dir, "/:/", "/")
	if runtime.GOOS == "windows" {
		dir = strings.ReplaceAll(dir, "/:/", "\\")
	}
	return filepath.Join(loc.Volume, dir, loc.File)
}

// traktorKeyToString converts Traktor key value to string
func traktorKeyToString(value int) string {
	// Traktor uses 0-23 for keys (0=C, 1=Db, etc. with major/minor)
	keys := []string{
		"C", "Db", "D", "Eb", "E", "F",
		"Gb", "G", "Ab", "A", "Bb", "B",
	}
	if value < 0 || value > 23 {
		return ""
	}
	keyIndex := value % 12
	isMinor := value >= 12
	suffix := ""
	if isMinor {
		suffix = "m"
	}
	return keys[keyIndex] + suffix
}

// cueTypeToString converts cue type to string
func cueTypeToString(cueType int) string {
	switch cueType {
	case 0:
		return "cue"
	case 1:
		return "fade-in"
	case 2:
		return "fade-out"
	case 3:
		return "load"
	case 4:
		return "grid"
	case 5:
		return "loop"
	default:
		return "unknown"
	}
}

// CollectionPath returns the collection path
func (c *TraktorClient) CollectionPath() string {
	return c.collectionPath
}

// Refresh reloads the collection from disk
func (c *TraktorClient) Refresh(ctx context.Context) error {
	c.collection = nil
	return c.loadCollection()
}

// parseTraktorDate parses Traktor date format
func parseTraktorDate(date string) time.Time {
	// Traktor uses format like "2024/1/15"
	parts := strings.Split(date, "/")
	if len(parts) != 3 {
		return time.Time{}
	}
	year, _ := strconv.Atoi(parts[0])
	month, _ := strconv.Atoi(parts[1])
	day, _ := strconv.Atoi(parts[2])
	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
}
