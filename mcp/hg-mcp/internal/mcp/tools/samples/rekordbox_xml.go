package samples

import (
	"encoding/xml"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// RekordboxTrack holds metadata for a single track in the Rekordbox XML
type RekordboxTrack struct {
	ID         int
	Name       string
	Artist     string
	Album      string
	Genre      string
	FilePath   string
	Duration   float64
	FileSize   int64
	SampleRate int
	BitRate    int
	CuePoints  []CuePoint
}

// rekordbox XML structures

type djPlaylists struct {
	XMLName    xml.Name         `xml:"DJ_PLAYLISTS"`
	Version    string           `xml:"Version,attr"`
	Product    rekordboxProduct `xml:"PRODUCT"`
	Collection collection       `xml:"COLLECTION"`
	Playlists  playlists        `xml:"PLAYLISTS"`
}

type rekordboxProduct struct {
	Name    string `xml:"Name,attr"`
	Version string `xml:"Version,attr"`
	Company string `xml:"Company,attr"`
}

type collection struct {
	Entries int        `xml:"Entries,attr"`
	Tracks  []xmlTrack `xml:"TRACK"`
}

type xmlTrack struct {
	TrackID       int               `xml:"TrackID,attr"`
	Name          string            `xml:"Name,attr"`
	Artist        string            `xml:"Artist,attr"`
	Album         string            `xml:"Album,attr"`
	Genre         string            `xml:"Genre,attr"`
	Kind          string            `xml:"Kind,attr"`
	Size          int64             `xml:"Size,attr"`
	TotalTime     int               `xml:"TotalTime,attr"`
	SampleRate    int               `xml:"SampleRate,attr"`
	BitRate       int               `xml:"BitRate,attr"`
	Location      string            `xml:"Location,attr"`
	PositionMarks []xmlPositionMark `xml:"POSITION_MARK,omitempty"`
}

type xmlPositionMark struct {
	Name  string  `xml:"Name,attr"`
	Type  int     `xml:"Type,attr"`
	Start float64 `xml:"Start,attr"`
	Num   int     `xml:"Num,attr"`
	Red   int     `xml:"Red,attr,omitempty"`
	Green int     `xml:"Green,attr,omitempty"`
	Blue  int     `xml:"Blue,attr,omitempty"`
}

type playlists struct {
	Node playlistNode `xml:"NODE"`
}

type playlistNode struct {
	Type    int             `xml:"Type,attr"`
	Name    string          `xml:"Name,attr"`
	Count   int             `xml:"Count,attr,omitempty"`
	KeyType int             `xml:"KeyType,attr,omitempty"`
	Entries []playlistEntry `xml:"TRACK,omitempty"`
	Nodes   []playlistNode  `xml:"NODE,omitempty"`
}

type playlistEntry struct {
	Key int `xml:"Key,attr"`
}

// fileLocationURL converts an absolute file path to a Rekordbox-compatible file:// URL
func fileLocationURL(absPath string) string {
	// Rekordbox expects file://localhost/path/to/file with URL-encoded path segments
	parts := strings.Split(absPath, string(filepath.Separator))
	var encoded []string
	for _, p := range parts {
		if p == "" {
			continue
		}
		encoded = append(encoded, url.PathEscape(p))
	}
	return "file://localhost/" + strings.Join(encoded, "/")
}

// trackKind returns the Rekordbox Kind string based on file extension
func trackKind(filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".aiff", ".aif":
		return "AIFF File"
	case ".wav":
		return "WAV File"
	case ".mp3":
		return "MP3 File"
	case ".flac":
		return "FLAC File"
	case ".m4a":
		return "M4A File"
	default:
		return "Audio File"
	}
}

// GenerateRekordboxXML creates a Rekordbox-compatible XML playlist file
func GenerateRekordboxXML(tracks []RekordboxTrack, playlistName, outputPath string) error {
	// Build XML tracks
	xmlTracks := make([]xmlTrack, len(tracks))
	entries := make([]playlistEntry, len(tracks))

	for i, t := range tracks {
		absPath, _ := filepath.Abs(t.FilePath)
		xmlTracks[i] = xmlTrack{
			TrackID:    t.ID,
			Name:       t.Name,
			Artist:     t.Artist,
			Album:      t.Album,
			Genre:      t.Genre,
			Kind:       trackKind(t.FilePath),
			Size:       t.FileSize,
			TotalTime:  int(t.Duration),
			SampleRate: t.SampleRate,
			BitRate:    t.BitRate / 1000, // Rekordbox expects kbps
			Location:   fileLocationURL(absPath),
		}
		entries[i] = playlistEntry{Key: t.ID}
	}

	doc := djPlaylists{
		Version: "1.0.0",
		Product: rekordboxProduct{
			Name:    "hg-mcp",
			Version: "0.1.0",
			Company: "AFTRS Studio",
		},
		Collection: collection{
			Entries: len(xmlTracks),
			Tracks:  xmlTracks,
		},
		Playlists: playlists{
			Node: playlistNode{
				Type: 0,
				Name: "ROOT",
				Nodes: []playlistNode{
					{
						Type:    1,
						Name:    playlistName,
						KeyType: 0,
						Count:   len(entries),
						Entries: entries,
					},
				},
			},
		},
	}

	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("cannot create XML file: %w", err)
	}
	defer f.Close()

	// Write XML header
	if _, err := f.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n"); err != nil {
		return err
	}

	enc := xml.NewEncoder(f)
	enc.Indent("", "  ")
	if err := enc.Encode(doc); err != nil {
		return fmt.Errorf("failed to encode XML: %w", err)
	}

	return nil
}
