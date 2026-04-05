package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/hairglasses-studio/mapitall/internal/daemon"
	"github.com/hairglasses-studio/mapitall/internal/ipc"
)

func main() {
	var (
		configPath string
		logLevel   string
	)
	flag.StringVar(&configPath, "config", "", "path to config file")
	flag.StringVar(&logLevel, "log-level", "info", "log level (debug, info, warn, error)")

	flag.Parse()

	level := slog.LevelInfo
	switch logLevel {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	}
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level}))
	slog.SetDefault(logger)

	subcmd := flag.Arg(0)
	switch subcmd {
	case "", "run":
		runDaemon(configPath)
	case "status":
		rpcCall(configPath, "status")
	case "reload":
		rpcCall(configPath, "reload")
	case "devices":
		rpcCall(configPath, "list_devices")
	case "profiles":
		rpcCall(configPath, "list_profiles")
	case "state":
		rpcCall(configPath, "get_state")
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\nusage: mapitall [run|status|reload|devices|profiles|state]\n", subcmd)
		os.Exit(1)
	}
}

func runDaemon(configPath string) {
	cfg, err := daemon.LoadConfig(configPath)
	if err != nil {
		slog.Error("load config", "error", err)
		os.Exit(1)
	}

	d, err := daemon.New(cfg)
	if err != nil {
		slog.Error("init daemon", "error", err)
		os.Exit(1)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// SIGHUP triggers profile reload without restart.
	sighup := make(chan os.Signal, 1)
	signal.Notify(sighup, syscall.SIGHUP)
	go func() {
		for range sighup {
			if err := d.Reload(); err != nil {
				slog.Error("SIGHUP reload failed", "error", err)
			}
		}
	}()

	if err := d.Run(ctx); err != nil {
		slog.Error("daemon error", "error", err)
		os.Exit(1)
	}
}

func rpcCall(configPath, method string) {
	cfg, err := daemon.LoadConfig(configPath)
	if err != nil {
		slog.Error("load config", "error", err)
		os.Exit(1)
	}

	result, err := ipc.Call(cfg.SocketPath, method, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	var pretty json.RawMessage
	if err := json.Unmarshal(result, &pretty); err != nil {
		fmt.Println(string(result))
		return
	}
	out, _ := json.MarshalIndent(pretty, "", "  ")
	fmt.Println(string(out))
}
