package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/samples"
)

var defaultVideoPath = func() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "American.Psycho.2000.1080p.BrRip.x264.YIFY.mp4"
	}
	return filepath.Join(home, "Downloads", "American.Psycho.2000.1080p.BrRip.x264.YIFY.mp4")
}()

const (
	packName = "american-psycho"
	artist           = "American Psycho"

	batchTimeout   = 5 * time.Minute
	convertTimeout = 30 * time.Second
	xmlTimeout     = 15 * time.Second
)

type sample struct {
	Name        string
	Start       string
	End         string
	Description string
}

var americanPsychoSamples = []sample{
	{"balanced-diet", "2:50", "2:57", "I believe in taking care of myself"},
	{"idea-of-patrick-bateman", "4:20", "4:30", "There is an idea of a Patrick Bateman"},
	{"simply-am-not-there", "4:58", "5:01", "I simply am not there"},
	{"silian-rail", "15:50", "15:55", "Silian Rail"},
	{"subtle-off-white", "16:50", "16:56", "Subtle off-white coloring"},
	{"watermark", "17:20", "17:24", "It even has a watermark"},
	{"paul-allens-card", "17:50", "17:53", "Let's see Paul Allen's card"},
	{"not-going-to-dorsia", "11:50", "11:54", "Not going to Dorsia"},
	{"reservation-at-dorsia", "34:50", "34:55", "Reservation at Dorsia now"},
	{"too-new-wave", "24:50", "24:55", "Too New Wave"},
	{"sports-came-out", "25:20", "25:27", "When Sports came out in 83"},
	{"huey-lewis", "25:50", "25:54", "Do you like Huey Lewis"},
	{"hey-paul", "26:50", "26:53", "Hey Paul"},
	{"hip-to-be-square", "26:20", "26:24", "Hip to be square"},
	{"phil-collins", "41:50", "41:53", "Do you like Phil Collins"},
	{"genesis-fan", "42:50", "42:57", "Big Genesis fan"},
	{"in-too-deep", "43:50", "43:55", "In Too Deep"},
	{"whitney-houston", "1:11:50", "1:11:53", "Do you like Whitney Houston"},
	{"greatest-love", "1:12:50", "1:12:57", "The greatest love of all"},
	{"videotapes", "9:30", "9:33", "I have to return some videotapes"},
	{"utterly-insane", "9:50", "9:55", "I like to dissect girls"},
	{"murders-and-executions", "7:50", "7:55", "Murders and executions"},
	{"killed-paul-allen", "1:26:50", "1:26:54", "I killed Paul Allen"},
	{"pain-is-constant", "1:27:50", "1:27:58", "My pain is constant and sharp"},
	{"meant-nothing", "1:39:50", "1:39:54", "This confession has meant nothing"},
}

func callTool(ctx context.Context, name string, args map[string]interface{}) (string, error) {
	td, ok := tools.GetRegistry().GetTool(name)
	if !ok {
		return "", fmt.Errorf("tool not found: %s", name)
	}

	req := mcp.CallToolRequest{}
	req.Params.Name = name
	req.Params.Arguments = args

	result, err := td.Handler(ctx, req)
	if err != nil {
		return "", fmt.Errorf("%s returned error: %w", name, err)
	}
	if result.IsError {
		for _, c := range result.Content {
			if tc, ok := c.(mcp.TextContent); ok {
				return "", fmt.Errorf("%s failed: %s", name, tc.Text)
			}
		}
		return "", fmt.Errorf("%s failed with unknown error", name)
	}

	for _, c := range result.Content {
		if tc, ok := c.(mcp.TextContent); ok {
			return tc.Text, nil
		}
	}
	return "", fmt.Errorf("%s returned no text content", name)
}

func main() {
	fmt.Println("=== American Psycho DJ Sample Extraction ===")
	fmt.Printf("Source: %s\n", defaultVideoPath)
	fmt.Printf("Pack: %s (%d samples)\n\n", packName, len(americanPsychoSamples))

	start := time.Now()

	// Step 1: Batch extract and process all samples
	fmt.Println("[Step 1/3] Batch extracting and processing samples...")

	samplesJSON, err := json.Marshal(americanPsychoSamples)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to marshal samples: %v\n", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), batchTimeout)
	defer cancel()

	output, err := callTool(ctx, "aftrs_samples_batch", map[string]interface{}{
		"source_path":  defaultVideoPath,
		"pack_name":    packName,
		"samples":      string(samplesJSON),
		"artist":       artist,
		"generate_xml": false,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Batch extraction failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(output)
	fmt.Printf("[Step 1/3] Done (%.1fs)\n\n", time.Since(start).Seconds())

	// Step 2: Convert each AIFF to WAV
	fmt.Println("[Step 2/3] Converting AIFF -> WAV...")

	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot determine home directory: %v\n", err)
		os.Exit(1)
	}
	packDir := filepath.Join(home, "Music", "Samples", packName)

	convertStart := time.Now()
	for i, s := range americanPsychoSamples {
		aiffPath := filepath.Join(packDir, s.Name+".aiff")
		if _, err := os.Stat(aiffPath); os.IsNotExist(err) {
			fmt.Printf("  [%d/%d] %s — skipped (AIFF not found)\n", i+1, len(americanPsychoSamples), s.Name)
			continue
		}

		ctx, cancel := context.WithTimeout(context.Background(), convertTimeout)
		output, err := callTool(ctx, "aftrs_samples_convert", map[string]interface{}{
			"audio_path": aiffPath,
			"format":     "wav",
		})
		cancel()

		if err != nil {
			fmt.Printf("  [%d/%d] %s — FAILED: %v\n", i+1, len(americanPsychoSamples), s.Name, err)
			continue
		}

		// Delete source AIFF after successful conversion
		os.Remove(aiffPath)
		fmt.Printf("  [%d/%d] %s — OK\n", i+1, len(americanPsychoSamples), s.Name)
		_ = output
	}
	fmt.Printf("[Step 2/3] Done (%.1fs)\n\n", time.Since(convertStart).Seconds())

	// Step 3: Generate Rekordbox XML over the final WAV files
	fmt.Println("[Step 3/3] Generating Rekordbox XML playlist...")

	ctx, cancel = context.WithTimeout(context.Background(), xmlTimeout)
	defer cancel()

	output, err = callTool(ctx, "aftrs_samples_rekordbox_xml", map[string]interface{}{
		"directory":     packDir,
		"playlist_name": "American Psycho",
		"artist":        artist,
		"genre":         "DJ Sample",
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Rekordbox XML generation failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(output)

	fmt.Printf("\n=== Complete in %.1fs ===\n", time.Since(start).Seconds())
	fmt.Printf("Output: %s\n", packDir)
}
