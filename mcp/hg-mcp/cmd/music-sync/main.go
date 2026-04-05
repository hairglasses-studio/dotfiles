package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/hairglasses-studio/hg-mcp/internal/sync"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	// Flags
	allFlag := flag.Bool("all", false, "Sync all services (SoundCloud + Beatport + Rekordbox)")
	soundcloudFlag := flag.Bool("soundcloud", false, "Sync SoundCloud only")
	beatportFlag := flag.Bool("beatport", false, "Sync Beatport only")
	rekordboxFlag := flag.Bool("rekordbox", false, "Import to Rekordbox only")
	userFlag := flag.String("user", "", "Sync specific user only")
	dryRunFlag := flag.Bool("dry-run", false, "Preview without making changes")
	statusFlag := flag.Bool("status", false, "Show sync status")
	healthFlag := flag.Bool("health", false, "Run health checks")
	jsonFlag := flag.Bool("json", false, "Output as JSON")
	serveFlag := flag.Bool("serve", false, "Keep running and expose metrics (for Prometheus scraping)")
	flag.Parse()

	// Setup context with signal handling
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Println("Received shutdown signal...")
		cancel()
	}()

	// Start metrics server with health endpoint
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			config := sync.DefaultConfig()
			checker := sync.NewHealthChecker(config)
			result := checker.Check(r.Context())
			w.Header().Set("Content-Type", "application/json")
			if result.Overall != "healthy" {
				w.WriteHeader(http.StatusServiceUnavailable)
			}
			json.NewEncoder(w).Encode(result)
		})
		log.Println("Metrics available at http://localhost:9091/metrics")
		log.Println("Health check at http://localhost:9091/health")
		if err := http.ListenAndServe(":9091", nil); err != nil && err != http.ErrServerClosed {
			log.Printf("Metrics server error: %v", err)
		}
	}()

	// Create config
	config := sync.DefaultConfig()
	config.DryRun = *dryRunFlag

	// Filter users if specified
	if *userFlag != "" {
		var filteredUsers []sync.UserConfig
		for _, u := range config.Users {
			if u.Username == *userFlag {
				filteredUsers = append(filteredUsers, u)
				break
			}
		}
		if len(filteredUsers) == 0 {
			log.Fatalf("User not found: %s", *userFlag)
		}
		config.Users = filteredUsers
	}

	// Create manager
	manager, err := sync.NewManager(config)
	if err != nil {
		log.Fatalf("Failed to create manager: %v", err)
	}

	// Handle health check command
	if *healthFlag {
		checker := sync.NewHealthChecker(config)
		result := checker.Check(ctx)
		if *jsonFlag {
			data, _ := json.MarshalIndent(result, "", "  ")
			fmt.Println(string(data))
		} else {
			printHealth(result)
		}
		if result.Overall != "healthy" {
			os.Exit(1)
		}
		return
	}

	// Handle status command
	if *statusFlag {
		state, err := manager.Status()
		if err != nil {
			log.Fatalf("Failed to get status: %v", err)
		}
		if *jsonFlag {
			data, _ := json.MarshalIndent(state, "", "  ")
			fmt.Println(string(data))
		} else {
			printStatus(state)
		}
		return
	}

	// Determine what to sync
	syncAll := *allFlag || (!*soundcloudFlag && !*beatportFlag && !*rekordboxFlag)

	var results []sync.SyncResult

	if syncAll {
		results, err = manager.SyncAll(ctx)
	} else {
		if *soundcloudFlag {
			for _, user := range config.Users {
				if user.SoundCloud {
					r, e := manager.SyncSoundCloud(ctx, user.Username)
					if e != nil {
						log.Printf("SoundCloud sync error: %v", e)
					}
					results = append(results, r...)
				}
			}
		}

		if *beatportFlag {
			for _, user := range config.Users {
				if user.Beatport {
					r, e := manager.SyncBeatport(ctx, user.Username)
					if e != nil {
						log.Printf("Beatport sync error: %v", e)
					}
					results = append(results, r...)
				}
			}
		}

		if *rekordboxFlag {
			r, e := manager.SyncRekordbox(ctx)
			if e != nil {
				log.Printf("Rekordbox import error: %v", e)
			}
			results = append(results, r)
		}
	}

	// Output results
	if *jsonFlag {
		data, _ := json.MarshalIndent(results, "", "  ")
		fmt.Println(string(data))
	} else {
		printResults(results)
	}

	if err != nil {
		os.Exit(1)
	}

	// Serve mode - keep running for metrics scraping
	if *serveFlag {
		log.Println("Serve mode: keeping metrics server running. Press Ctrl+C to stop.")
		<-ctx.Done()
	}
}

func printStatus(state *sync.State) {
	fmt.Println("=== Music Sync Status ===")
	fmt.Println()

	fmt.Println("SoundCloud:")
	for user, userState := range state.SoundCloud {
		for playlist, ps := range userState.Playlists {
			fmt.Printf("  %s/%s: %d tracks (last sync: %s)\n",
				user, playlist, ps.Synced, ps.LastSync.Format("2006-01-02 15:04"))
		}
	}

	fmt.Println()
	fmt.Println("Beatport:")
	for user, userState := range state.Beatport {
		for playlist, ps := range userState.Playlists {
			fmt.Printf("  %s/%s: %d tracks (last sync: %s)\n",
				user, playlist, ps.Synced, ps.LastSync.Format("2006-01-02 15:04"))
		}
	}

	fmt.Println()
	fmt.Println("Rekordbox:")
	if state.Rekordbox != nil {
		fmt.Printf("  Last import: %s\n", state.Rekordbox.LastImport.Format("2006-01-02 15:04"))
		if len(state.Rekordbox.PendingFiles) > 0 {
			fmt.Printf("  Pending files: %d\n", len(state.Rekordbox.PendingFiles))
		}
		if state.Rekordbox.LastError != "" {
			fmt.Printf("  Last error: %s\n", state.Rekordbox.LastError)
		}
	}
}

func printResults(results []sync.SyncResult) {
	fmt.Println("=== Sync Results ===")
	fmt.Println()

	for _, r := range results {
		status := "OK"
		if len(r.Errors) > 0 {
			status = "ERROR"
		}
		if r.DryRun {
			status = "DRY-RUN"
		}

		fmt.Printf("[%s] %s/%s/%s\n", status, r.Service, r.User, r.Playlist)
		fmt.Printf("  Total: %d | Synced: %d | Skipped: %d | Failed: %d\n",
			r.Total, r.Synced, r.Skipped, r.Failed)
		fmt.Printf("  Duration: %s\n", r.EndTime.Sub(r.StartTime).Round(1e9))

		if len(r.Errors) > 0 {
			fmt.Println("  Errors:")
			for _, e := range r.Errors {
				fmt.Printf("    - %s\n", e)
			}
		}
		fmt.Println()
	}
}

func printHealth(result sync.HealthCheckResult) {
	fmt.Println("=== Health Check ===")
	fmt.Println()

	statusIcon := map[string]string{
		"healthy":   "OK",
		"degraded":  "WARN",
		"unhealthy": "FAIL",
	}

	fmt.Printf("Overall: [%s] %s\n\n", statusIcon[result.Overall], result.Overall)

	for _, svc := range result.Services {
		fmt.Printf("[%s] %s\n", statusIcon[svc.Status], svc.Name)
		if svc.Message != "" {
			fmt.Printf("  Message: %s\n", svc.Message)
		}
		if svc.Latency != "" {
			fmt.Printf("  Latency: %s\n", svc.Latency)
		}
	}
}
