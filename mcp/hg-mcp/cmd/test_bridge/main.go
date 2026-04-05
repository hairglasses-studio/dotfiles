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
	fmt.Println("  Pro DJ Link → Showkontrol Bridge Test")
	fmt.Println("========================================")
	fmt.Println("")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// First connect to Pro DJ Link
	fmt.Println("[1/3] Connecting to Pro DJ Link network...")
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
	fmt.Println("      Connected!")

	// Wait for devices to be discovered
	fmt.Println("[2/3] Waiting for device discovery (5 seconds)...")
	time.Sleep(5 * time.Second)

	devices, _ := prolinkClient.GetDevices(ctx)
	fmt.Printf("      Found %d device(s)\n", len(devices))
	for _, dev := range devices {
		fmt.Printf("        - %s (Player %d)\n", dev.Name, dev.ID)
	}

	// Initialize bridge
	fmt.Println("")
	fmt.Println("[3/3] Starting bridge...")

	b, err := bridge.GetBridge()
	if err != nil {
		fmt.Printf("ERROR: Failed to get bridge: %v\n", err)
		return
	}

	// Configure bridge
	config := &bridge.BridgeConfig{
		Mode:           bridge.BridgeModeMaster,
		PollIntervalMs: 200,
		// BeatCue:        "beat_cue_1",     // Set these if you have Showkontrol running
		// TrackChangeCue: "track_change_1",
	}
	b.Configure(config)

	// Set up beat callback to show activity
	b.SetOnBeat(func(event bridge.BeatEvent) {
		if event.IsDownbeat {
			fmt.Printf("\r🥁 Beat 1 | BPM: %.1f | Player: %d                    ", event.EffectiveBPM, event.PlayerID)
		}
	})

	// Set up track change callback
	b.SetOnTrackChange(func(event bridge.TrackChangeEvent) {
		fmt.Printf("\n\n🎵 Track Changed on Player %d:\n", event.PlayerID)
		if event.NewTrack != nil {
			fmt.Printf("   %s - %s\n", event.NewTrack.Artist, event.NewTrack.Title)
			if event.NewTrack.Key != "" {
				color := bridge.KeyToColor(event.NewTrack.Key)
				fmt.Printf("   Key: %s → Color: %s\n", event.NewTrack.Key, color)
			}
			if event.NewTrack.Genre != "" {
				fmt.Printf("   Genre: %s\n", event.NewTrack.Genre)
			}
		}
		fmt.Println("")
	})

	// Start the bridge
	if err := b.Start(ctx); err != nil {
		fmt.Printf("ERROR: Failed to start bridge: %v\n", err)
		return
	}

	fmt.Println("========================================")
	fmt.Println("  Bridge Running - Press Ctrl+C to stop")
	fmt.Println("========================================")
	fmt.Println("")
	fmt.Println("Watching for beats and track changes...")
	fmt.Println("")

	// Wait for interrupt
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	fmt.Println("\n\nStopping bridge...")
	b.Stop()

	// Print final stats
	status := b.GetStatus()
	fmt.Println("")
	fmt.Println("========================================")
	fmt.Println("  Session Stats")
	fmt.Println("========================================")
	fmt.Printf("  Beats Synced:   %d\n", status.BeatsSynced)
	fmt.Printf("  Track Changes:  %d\n", status.TrackChanges)
	if status.LastCueFired != "" {
		fmt.Printf("  Last Cue Fired: %s\n", status.LastCueFired)
	}
	fmt.Println("")
	fmt.Println("Done!")
}
