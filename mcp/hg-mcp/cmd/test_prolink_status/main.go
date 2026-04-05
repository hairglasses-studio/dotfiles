package main

import (
	"encoding/json"
	"fmt"
	"time"

	"go.evanpurkhiser.com/prolink"
)

func main() {
	fmt.Println("========================================")
	fmt.Println("  Pro DJ Link - Full Data Monitor")
	fmt.Println("========================================")
	fmt.Println("")

	network, err := prolink.Connect()
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
		return
	}

	if err := network.AutoConfigure(5 * time.Second); err != nil {
		fmt.Printf("WARNING: Auto-configure failed: %v\n", err)
	}

	dm := network.DeviceManager()
	st := network.CDJStatusMonitor()
	rdb := network.RemoteDB()

	fmt.Println("Listening for status updates (10 seconds)...")
	fmt.Println("")

	// Track last status to avoid spam
	var lastTrackID uint32
	var lastStatus *prolink.CDJStatus
	var lastTrack *prolink.Track

	st.AddStatusHandler(prolink.StatusHandlerFunc(func(status *prolink.CDJStatus) {
		lastStatus = status

		// Only fetch track when it changes
		if status.TrackID != lastTrackID || lastTrackID == 0 {
			lastTrackID = status.TrackID

			if trackKey := status.TrackKey(); trackKey != nil {
				if track, err := rdb.GetTrack(trackKey); err == nil {
					lastTrack = track
				}
			}
		}
	}))

	// Wait for devices to be discovered and status to be received
	time.Sleep(10 * time.Second)

	// Print comprehensive data
	devices := dm.ActiveDevices()

	fmt.Println("========================================")
	fmt.Println("  FULL PROLINK DATA DUMP")
	fmt.Println("========================================")
	fmt.Println("")

	// Devices
	fmt.Printf("DEVICES (%d found):\n", len(devices))
	for _, dev := range devices {
		fmt.Printf("  [%d] %s\n", dev.ID, dev.Name)
		fmt.Printf("      Type: %s\n", dev.Type.String())
		fmt.Printf("      IP:   %s\n", dev.IP.String())
		fmt.Printf("      MAC:  %s\n", dev.MacAddr.String())
		fmt.Printf("      Last Active: %s\n", dev.LastActive.Format(time.RFC3339))
	}
	fmt.Println("")

	// Status
	if lastStatus != nil {
		effectiveBPM := lastStatus.TrackBPM * (1 + lastStatus.SliderPitch/100)
		msPerBeat := 60000.0 / float64(effectiveBPM)

		fmt.Println("PLAYER STATUS:")
		fmt.Printf("  Player ID:       %d\n", lastStatus.PlayerID)
		fmt.Printf("  Play State:      %s\n", lastStatus.PlayState.String())
		fmt.Printf("  Track ID:        %d\n", lastStatus.TrackID)
		fmt.Printf("  Track Device:    %d\n", lastStatus.TrackDevice)
		fmt.Printf("  Track Slot:      %s\n", lastStatus.TrackSlot.String())
		fmt.Printf("  Track Type:      %s\n", lastStatus.TrackType.String())
		fmt.Printf("  BPM:             %.2f\n", lastStatus.TrackBPM)
		fmt.Printf("  Effective BPM:   %.2f\n", effectiveBPM)
		fmt.Printf("  Slider Pitch:    %.2f%%\n", lastStatus.SliderPitch)
		fmt.Printf("  Effective Pitch: %.2f%%\n", lastStatus.EffectivePitch)
		fmt.Printf("  Beat in Measure: %d/4\n", lastStatus.BeatInMeasure)
		fmt.Printf("  Beat (total):    %d\n", lastStatus.Beat)
		fmt.Printf("  Beats Until Cue: %d\n", lastStatus.BeatsUntilCue)
		fmt.Printf("  Packet Num:      %d\n", lastStatus.PacketNum)
		fmt.Printf("  ms/beat:         %.1f\n", msPerBeat)
		fmt.Printf("  Is Master:       %v\n", lastStatus.IsMaster)
		fmt.Printf("  Is Sync:         %v\n", lastStatus.IsSync)
		fmt.Printf("  Is On Air:       %v\n", lastStatus.IsOnAir)
		fmt.Println("")
	}

	// Track metadata
	if lastTrack != nil {
		fmt.Println("TRACK METADATA:")
		fmt.Printf("  ID:         %d\n", lastTrack.ID)
		fmt.Printf("  Title:      %s\n", lastTrack.Title)
		fmt.Printf("  Artist:     %s\n", lastTrack.Artist)
		fmt.Printf("  Album:      %s\n", lastTrack.Album)
		fmt.Printf("  Genre:      %s\n", lastTrack.Genre)
		fmt.Printf("  Label:      %s\n", lastTrack.Label)
		fmt.Printf("  Key:        %s\n", lastTrack.Key)
		fmt.Printf("  Length:     %s\n", lastTrack.Length.String())
		fmt.Printf("  Comment:    %s\n", lastTrack.Comment)
		fmt.Printf("  Path:       %s\n", lastTrack.Path)
		fmt.Printf("  Date Added: %s\n", lastTrack.DateAdded.Format(time.RFC3339))
		fmt.Printf("  Has Artwork: %v (%d bytes)\n", len(lastTrack.Artwork) > 0, len(lastTrack.Artwork))
		fmt.Println("")
	}

	// JSON output
	fmt.Println("JSON OUTPUT:")
	data := map[string]interface{}{
		"timestamp": time.Now().Format(time.RFC3339Nano),
		"devices":   formatDevices(devices),
		"status":    formatStatus(lastStatus),
		"track":     formatTrack(lastTrack),
	}
	jsonBytes, _ := json.MarshalIndent(data, "", "  ")
	fmt.Println(string(jsonBytes))

	fmt.Println("")
	fmt.Println("Done!")
}

func formatDevices(devices []*prolink.Device) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(devices))
	for _, dev := range devices {
		result = append(result, map[string]interface{}{
			"id":          dev.ID,
			"name":        dev.Name,
			"type":        dev.Type.String(),
			"ip":          dev.IP.String(),
			"mac":         dev.MacAddr.String(),
			"last_active": dev.LastActive.Format(time.RFC3339),
		})
	}
	return result
}

func formatStatus(status *prolink.CDJStatus) map[string]interface{} {
	if status == nil {
		return nil
	}
	effectiveBPM := status.TrackBPM * (1 + status.SliderPitch/100)
	msPerBeat := float64(0)
	if effectiveBPM > 0 {
		msPerBeat = 60000.0 / float64(effectiveBPM)
	}
	return map[string]interface{}{
		"player_id":       status.PlayerID,
		"play_state":      status.PlayState.String(),
		"track_id":        status.TrackID,
		"track_device":    status.TrackDevice,
		"track_slot":      status.TrackSlot.String(),
		"track_type":      status.TrackType.String(),
		"bpm":             status.TrackBPM,
		"effective_bpm":   effectiveBPM,
		"slider_pitch":    status.SliderPitch,
		"effective_pitch": status.EffectivePitch,
		"beat_in_measure": status.BeatInMeasure,
		"beat":            status.Beat,
		"beats_until_cue": status.BeatsUntilCue,
		"packet_num":      status.PacketNum,
		"ms_per_beat":     msPerBeat,
		"is_master":       status.IsMaster,
		"is_sync":         status.IsSync,
		"is_on_air":       status.IsOnAir,
	}
}

func formatTrack(track *prolink.Track) map[string]interface{} {
	if track == nil {
		return nil
	}
	return map[string]interface{}{
		"id":          track.ID,
		"title":       track.Title,
		"artist":      track.Artist,
		"album":       track.Album,
		"genre":       track.Genre,
		"label":       track.Label,
		"key":         track.Key,
		"length_sec":  track.Length.Seconds(),
		"comment":     track.Comment,
		"path":        track.Path,
		"date_added":  track.DateAdded.Format(time.RFC3339),
		"has_artwork": len(track.Artwork) > 0,
	}
}
