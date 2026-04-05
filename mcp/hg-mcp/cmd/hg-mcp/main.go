// Package main is the entry point for the hg-mcp MCP server.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/mark3labs/mcp-go/server"

	"github.com/hairglasses-studio/hg-mcp/internal/config"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
	"github.com/hairglasses-studio/hg-mcp/internal/observability"
	"github.com/hairglasses-studio/hg-mcp/internal/web"

	// Import modules to trigger registration via init()
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/ableton"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/analytics"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/archive"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/ardour"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/atem"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/avsync"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/backup"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/bandcamp"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/beatport"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/boomkat"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/bpmsync"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/calendar"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/chains"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/chataigne"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/companion"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/consolidated"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/cr8"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/dante"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/data_migration"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/dashboard"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/discogs"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/discord"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/discord_admin"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/discovery"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/federation"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/ffgl"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/fingerprint"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/gateway"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/gdrive"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/gmail"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/gpushare"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/grandma3"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/graph"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/gtasks"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/healing"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/homeassistant"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/hue"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/hwmonitor"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/inventory"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/juno"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/learning"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/ledfx"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/lighting"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/loader"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/linuxshowplayer"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/mapmap"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/maxforlive"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/memory"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/midi"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/mixcloud"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/mqtt"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/music_discovery"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/nanoleaf"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/ndicv"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/notion"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/obs"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/ola"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/ollama"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/opnsense"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/opc"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/ossia"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/pages"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/paramsync"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/plugins"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/prolink"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/ptz"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/ptztrack"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/puredata"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/qlcplus"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/rclone"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/rekordbox"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/resolume"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/resolume_plugins"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/retrogaming"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/router"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/sacn"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/samples"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/security"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/serato"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/setlist"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/showcontrol"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/showkontrol"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/slack"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/snapshots"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/soundcloud"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/spotify"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/stems"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/streamdeck"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/streaming"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/studio"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/supercollider"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/swarm"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/sync"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/system"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/tailscale"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/tasks"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/telegram"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/tidal"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/timecodesync"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/touchdesigner"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/traktor"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/traxsource"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/triggersync"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/twitch"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/unraid"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/usb"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/vault"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/video"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/videoai"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/videorouting"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/vimix"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/vj_clips"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/whisper"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/wled"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/workflow_automation"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/workflows"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/xlights"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/youtube_live"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/ytmusic"
)

func main() {
	startupStart := time.Now()

	// Load centralized config from environment (must happen before any config.Get() call)
	cfg := config.Load()

	// Configure structured logging
	if cfg.LogFormat == "json" {
		slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, nil)))
	} else {
		slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, nil)))
	}

	// Handle --generate-docs: output consolidated tool reference and exit
	for _, arg := range os.Args[1:] {
		if arg == "--generate-docs" {
			if err := generateToolDocs(); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			return
		}
	}

	// Start pprof profiling server if enabled
	if os.Getenv("PPROF_ENABLED") == "true" {
		pprofPort := os.Getenv("PPROF_PORT")
		if pprofPort == "" {
			pprofPort = "6060"
		}
		go func() {
			pprofAddr := "localhost:" + pprofPort
			slog.Info("pprof profiling enabled", "addr", "http://"+pprofAddr+"/debug/pprof/")
			if err := http.ListenAndServe(pprofAddr, nil); err != nil {
				slog.Error("pprof server error", "error", err)
			}
		}()
	}

	// Intercept SIGINT/SIGTERM for graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Initialize observability (OTEL + Prometheus)
	otelCfg := observability.DefaultConfig()
	if endpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"); endpoint != "" {
		otelCfg.OTLPEndpoint = endpoint
	}
	if port := os.Getenv("PROMETHEUS_PORT"); port != "" {
		otelCfg.PrometheusPort = port
	}
	if os.Getenv("DISABLE_TRACING") == "true" {
		otelCfg.EnableTracing = false
	}
	if os.Getenv("DISABLE_METRICS") == "true" {
		otelCfg.EnableMetrics = false
	}

	otelShutdown, err := observability.Init(ctx, otelCfg)
	if err != nil {
		slog.Warn("failed to initialize observability", "error", err)
	} else {
		defer func() {
			// Use background context — signal ctx is already cancelled at this point
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := otelShutdown(shutdownCtx); err != nil {
				slog.Error("failed to shut down observability", "error", err)
			}
		}()
	}

	// Config audit — log which services are configured
	config.AuditConfig().Log()

	// Create MCP server
	s := server.NewMCPServer("hg-mcp", "0.1.0",
		server.WithToolCapabilities(true),
		server.WithResourceCapabilities(true, true),
		server.WithPromptCapabilities(true),
	)

	// Set tool profile before registration (controls eager vs deferred loading)
	registry := tools.GetRegistry()
	registry.SetProfile(tools.HgToolProfileFromEnv())

	// Register all tools, resources, and prompts
	mcp.RegisterTools(s)
	mcp.RegisterResources(s)
	mcp.RegisterPrompts(s)

	// Log startup metrics
	slog.Info("hg-mcp server ready",
		"profile", registry.GetProfile(),
		"eager_tools", registry.ToolCount()-registry.DeferredToolCount(),
		"deferred_tools", registry.DeferredToolCount(),
		"modules", registry.ModuleCount(),
		"startup_ms", time.Since(startupStart).Milliseconds(),
	)

	switch cfg.MCPMode {
	case "web":
		port := cfg.Port

		devMode := os.Getenv("WEB_DEV") == "true"
		webHandler := web.NewServer(web.Config{
			Port:    port,
			DevMode: devMode,
		})

		httpSrv := &http.Server{
			Addr:    ":" + port,
			Handler: webHandler,
		}

		slog.Info("starting hg-mcp", "mode", "web", "port", port)
		slog.Info("web UI available", "url", "http://localhost:"+port)
		slog.Info("API available", "url", "http://localhost:"+port+"/api/v1/")
		if devMode {
			slog.Info("dev mode enabled", "vite_proxy", "http://localhost:5173")
		}

		go func() {
			if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				slog.Error("web server error", "error", err)
				os.Exit(1)
			}
		}()

		// Block until signal
		<-ctx.Done()
		stop() // Reset signal handling so second signal force-kills
		gracefulShutdown(httpSrv)

	case "sse":
		port := cfg.Port

		baseURL := os.Getenv("MCP_BASE_URL")
		if baseURL == "" {
			baseURL = fmt.Sprintf("http://localhost:%s", port)
		}

		sseServer := server.NewSSEServer(s,
			server.WithBaseURL(baseURL),
			server.WithKeepAlive(true),
		)

		// Create mux with health endpoint
		mux := http.NewServeMux()
		mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("ok"))
		})
		mux.Handle("/sse", sseServer.SSEHandler())
		mux.Handle("/message", sseServer.MessageHandler())

		httpSrv := &http.Server{
			Addr:    ":" + port,
			Handler: mux,
		}

		slog.Info("starting hg-mcp", "mode", "sse", "port", port)
		slog.Info("SSE endpoint available", "url", baseURL+"/sse")
		slog.Info("message endpoint available", "url", baseURL+"/message")
		slog.Info("health endpoint available", "url", baseURL+"/health")

		go func() {
			if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				slog.Error("SSE server error", "error", err)
				os.Exit(1)
			}
		}()

		// Block until signal
		<-ctx.Done()
		stop() // Reset signal handling so second signal force-kills
		gracefulShutdown(httpSrv)

	default:
		// Stdio mode for local Claude Code integration
		slog.Info("starting hg-mcp", "mode", "stdio")
		if err := server.ServeStdio(s); err != nil {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}

	slog.Info("hg-mcp server stopped")
}

// gracefulShutdown drains in-flight requests with a 15-second deadline.
func gracefulShutdown(srv *http.Server) {
	slog.Info("shutting down server", "drain_timeout", "15s")
	drainCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := srv.Shutdown(drainCtx); err != nil {
		slog.Error("forced shutdown", "error", err)
	}
}

// generateToolDocs writes a consolidated tools-reference.md to docs/
func generateToolDocs() error {
	registry := tools.GetRegistry()
	stats := registry.GetToolStats()
	allTools := registry.GetAllToolDefinitions()

	// Group by category
	byCategory := make(map[string][]tools.ToolDefinition)
	for _, td := range allTools {
		cat := td.Category
		if cat == "" {
			cat = "uncategorized"
		}
		byCategory[cat] = append(byCategory[cat], td)
	}

	// Sort categories
	cats := make([]string, 0, len(byCategory))
	for c := range byCategory {
		cats = append(cats, c)
	}
	sort.Strings(cats)

	var sb strings.Builder
	sb.WriteString("# hg-mcp Tools Reference\n\n")
	sb.WriteString(fmt.Sprintf("> Auto-generated on %s\n\n", time.Now().Format("2006-01-02")))
	sb.WriteString(fmt.Sprintf("**%d tools** across **%d modules**\n\n", stats.TotalTools, stats.ModuleCount))

	// Table of contents
	sb.WriteString("## Categories\n\n")
	for _, cat := range cats {
		anchor := strings.ReplaceAll(cat, " ", "-")
		sb.WriteString(fmt.Sprintf("- [%s](#%s) (%d tools)\n", cat, anchor, len(byCategory[cat])))
	}
	sb.WriteString("\n---\n\n")

	// Per-category tool listing
	for _, cat := range cats {
		toolList := byCategory[cat]
		sort.Slice(toolList, func(i, j int) bool {
			return toolList[i].Tool.Name < toolList[j].Tool.Name
		})

		sb.WriteString(fmt.Sprintf("## %s\n\n", cat))

		for _, td := range toolList {
			sb.WriteString(fmt.Sprintf("### `%s`\n\n", td.Tool.Name))
			sb.WriteString(fmt.Sprintf("%s\n\n", td.Tool.Description))

			// Parameters table
			if td.Tool.InputSchema.Properties != nil && len(td.Tool.InputSchema.Properties) > 0 {
				required := make(map[string]bool)
				for _, r := range td.Tool.InputSchema.Required {
					required[r] = true
				}

				params := make([]string, 0, len(td.Tool.InputSchema.Properties))
				for p := range td.Tool.InputSchema.Properties {
					params = append(params, p)
				}
				sort.Strings(params)

				sb.WriteString("| Parameter | Type | Required | Description |\n")
				sb.WriteString("|-----------|------|----------|-------------|\n")

				for _, p := range params {
					prop := td.Tool.InputSchema.Properties[p]
					propMap, ok := prop.(map[string]interface{})
					if !ok {
						continue
					}
					pType := "any"
					if t, ok := propMap["type"].(string); ok {
						pType = t
					}
					pDesc := ""
					if d, ok := propMap["description"].(string); ok {
						pDesc = d
					}
					req := ""
					if required[p] {
						req = "Yes"
					}
					sb.WriteString(fmt.Sprintf("| `%s` | %s | %s | %s |\n", p, pType, req, pDesc))
				}
				sb.WriteString("\n")
			}
		}
	}

	// Write file
	if err := os.MkdirAll("docs", 0755); err != nil {
		return fmt.Errorf("failed to create docs dir: %w", err)
	}

	path := "docs/tools-reference.md"
	if err := os.WriteFile(path, []byte(sb.String()), 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", path, err)
	}

	fmt.Printf("Generated %s (%d tools, %d categories)\n", path, stats.TotalTools, len(cats))
	return nil
}
