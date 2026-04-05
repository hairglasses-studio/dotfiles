package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hairglasses-studio/hg-mcp/internal/bridge"
	"github.com/hairglasses-studio/hg-mcp/internal/clients"
)

func main() {
	fmt.Println("========================================")
	fmt.Println("  XDJ → Resolume Track Display Test")
	fmt.Println("========================================")
	fmt.Println("")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Test Resolume connection first
	fmt.Println("[1/4] Testing Resolume connection...")
	resolume, err := clients.NewResolumeClient()
	if err != nil {
		fmt.Printf("ERROR: Failed to create Resolume client: %v\n", err)
		fmt.Println("")
		fmt.Println("Make sure:")
		fmt.Println("  - Resolume Arena/Avenue is running")
		fmt.Println("  - OSC is enabled in Preferences → OSC (port 7000)")
		return
	}

	// Send a test message
	err = resolume.SetFormattedNowPlaying("Test", "Connection OK")
	if err != nil {
		fmt.Printf("WARNING: OSC send may have failed: %v\n", err)
	} else {
		fmt.Println("      Resolume OSC connected!")
	}

	// Connect to Pro DJ Link
	fmt.Println("[2/4] Connecting to Pro DJ Link network...")
	prolinkClient, err := clients.GetProlinkClient()
	if err != nil {
		fmt.Printf("ERROR: Failed to get prolink client: %v\n", err)
		return
	}

	if err := prolinkClient.Connect(ctx); err != nil {
		fmt.Printf("ERROR: Failed to connect to Pro DJ Link: %v\n", err)
		fmt.Println("")
		fmt.Println("Make sure:")
		fmt.Println("  - Rekordbox is closed (blocks UDP port 50000)")
		fmt.Println("  - XDJ/CDJ is on the same network")
		fmt.Println("  - Windows Firewall allows UDP 50000-50002")
		return
	}
	fmt.Println("      Pro DJ Link connected!")

	// Wait for devices
	fmt.Println("[3/4] Waiting for device discovery (5 seconds)...")
	time.Sleep(5 * time.Second)

	devices, _ := prolinkClient.GetDevices(ctx)
	fmt.Printf("      Found %d device(s)\n", len(devices))
	for _, dev := range devices {
		fmt.Printf("        - %s (Player %d)\n", dev.Name, dev.ID)
	}

	// Start the display bridge
	fmt.Println("")
	fmt.Println("[4/4] Starting Resolume display bridge...")

	displayBridge, err := bridge.GetResolumeDisplayBridge()
	if err != nil {
		fmt.Printf("ERROR: Failed to get display bridge: %v\n", err)
		return
	}

	// Configure for combined display mode
	displayBridge.Configure(&bridge.ResolumeConfig{
		DisplayMode:    bridge.ResolumeDisplayFull,
		UpdateOnLoad:   true,
		ClearOnStop:    true,
		PollIntervalMs: 500,
	})

	// Set up track change callback
	displayBridge.SetOnTrackChange(func(event bridge.ResolumeTrackEvent) {
		fmt.Printf("\n🎵 Track Changed → Resolume:\n")
		fmt.Printf("   Artist: %s\n", event.Artist)
		fmt.Printf("   Title:  %s\n", event.Title)
		if event.Key != "" {
			fmt.Printf("   Key:    %s\n", event.Key)
		}
		if event.BPM > 0 {
			fmt.Printf("   BPM:    %.1f\n", event.BPM)
		}
		if event.Genre != "" {
			fmt.Printf("   Genre:  %s\n", event.Genre)
		}
		fmt.Println("")
	})

	if err := displayBridge.Start(ctx); err != nil {
		fmt.Printf("ERROR: Failed to start display bridge: %v\n", err)
		return
	}

	fmt.Println("========================================")
	fmt.Println("  Display Bridge Running!")
	fmt.Println("  Press Ctrl+C to stop")
	fmt.Println("========================================")
	fmt.Println("")
	fmt.Println("In Resolume:")
	fmt.Println("  1. Add a Text source to a clip")
	fmt.Println("  2. Go to the source's String parameter")
	fmt.Println("  3. Right-click → Link to Dashboard String 1")
	fmt.Println("     (for artist)")
	fmt.Println("  4. Link another text to Dashboard String 2")
	fmt.Println("     (for title)")
	fmt.Println("")
	fmt.Println("Dashboard String mapping:")
	fmt.Println("  1 = Artist")
	fmt.Println("  2 = Title")
	fmt.Println("  3 = Key")
	fmt.Println("  4 = BPM")
	fmt.Println("  5 = Genre")
	fmt.Println("")
	fmt.Println("Waiting for track changes...")
	fmt.Println("")

	// Wait for interrupt
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	fmt.Println("\n\nStopping bridge...")
	displayBridge.Stop()

	// Print final stats
	status := displayBridge.GetStatus()
	fmt.Println("")
	fmt.Println("========================================")
	fmt.Println("  Session Stats")
	fmt.Println("========================================")
	fmt.Printf("  Track Changes: %d\n", status.TrackChanges)
	if status.CurrentArtist != "" {
		fmt.Printf("  Last Track:    %s - %s\n", status.CurrentArtist, status.CurrentTitle)
	}
	fmt.Println("")
	fmt.Println("Done!")
}
