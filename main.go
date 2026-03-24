package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	telegram "github.com/leonletto/gt-telegram/internal/telegram"
)

// Version and Build are set by ldflags at build time.
var (
	Version = "dev"
	Build   = "unknown"
)

const usage = `gt-telegram — Telegram bridge for Gas Town overseer communication

Usage:
  gt-telegram <command> [flags]

Commands:
  configure   Configure the Telegram bridge
  status      Show bridge configuration status
  run         Run the bridge in the foreground
  version     Show version information

Environment:
  GT_TOWN     Gas Town root directory (default: current directory)
`

func main() {
	if len(os.Args) < 2 {
		fmt.Fprint(os.Stderr, usage)
		os.Exit(1)
	}

	cmd := os.Args[1]
	os.Args = append(os.Args[:1], os.Args[2:]...) // shift for flag parsing

	var err error
	switch cmd {
	case "configure":
		err = runConfigure()
	case "status":
		err = runStatus()
	case "run":
		err = runBridge()
	case "version", "-v", "--version":
		fmt.Printf("gt-telegram %s (build %s)\n", Version, Build)
		return
	case "-h", "--help", "help":
		fmt.Print(usage)
		return
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n%s", cmd, usage)
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// townRoot returns the Gas Town root directory from GT_TOWN env var
// or falls back to the current working directory.
func townRoot() string {
	if root := os.Getenv("GT_TOWN"); root != "" {
		return root
	}
	if root := os.Getenv("GT_TOWN_ROOT"); root != "" {
		return root
	}
	dir, _ := os.Getwd()
	return dir
}

// --- configure ---

func runConfigure() error {
	fs := flag.NewFlagSet("configure", flag.ExitOnError)
	token := fs.String("token", "", "Telegram bot token (from @BotFather)")
	chatID := fs.Int64("chat-id", 0, "Telegram chat ID to send messages to")
	allowFromStr := fs.String("allow-from", "", "Allowed sender user IDs (comma-separated)")
	notifyStr := fs.String("notify", "", "Notification categories (comma-separated)")
	yes := fs.Bool("yes", false, "Skip confirmation prompts")
	fs.Parse(os.Args[1:]) //nolint:errcheck

	root := townRoot()
	configPath := telegram.ConfigPath(root)

	// Load existing config if present.
	var cfg telegram.Config
	existing, loadErr := telegram.LoadConfig(configPath)
	if loadErr == nil {
		cfg = existing
	} else if !errors.Is(loadErr, os.ErrNotExist) {
		return fmt.Errorf("loading existing config: %w", loadErr)
	}

	// Apply flags that were provided.
	if *token != "" {
		if cfg.Token != "" && cfg.Token != *token && !*yes {
			fmt.Printf("Replacing existing token (%s).\n", cfg.MaskedToken())
			fmt.Print("Continue? [y/N] ")
			var answer string
			fmt.Scanln(&answer) //nolint:errcheck
			if answer != "y" && answer != "Y" {
				fmt.Println("Aborted.")
				return nil
			}
		}
		cfg.Token = *token
	}
	if *chatID != 0 {
		cfg.ChatID = *chatID
	}
	if *allowFromStr != "" {
		ids, err := parseIntList(*allowFromStr)
		if err != nil {
			return fmt.Errorf("invalid --allow-from: %w", err)
		}
		cfg.AllowFrom = ids
	}
	if *notifyStr != "" {
		cfg.Notify = strings.Split(*notifyStr, ",")
	}

	cfg.Enabled = true
	cfg.ApplyDefaults()

	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	if err := telegram.SaveConfig(configPath, cfg); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	fmt.Printf("Telegram bridge configured (%s).\n", configPath)
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  gt-telegram status    # verify configuration")
	fmt.Println("  gt-telegram run       # start the bridge")
	return nil
}

func parseIntList(s string) ([]int64, error) {
	parts := strings.Split(s, ",")
	var ids []int64
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		id, err := strconv.ParseInt(p, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid ID %q: %w", p, err)
		}
		ids = append(ids, id)
	}
	return ids, nil
}

// --- status ---

func runStatus() error {
	fs := flag.NewFlagSet("status", flag.ExitOnError)
	jsonOutput := fs.Bool("json", false, "Output as JSON")
	fs.Parse(os.Args[1:]) //nolint:errcheck

	root := townRoot()
	configPath := telegram.ConfigPath(root)
	cfg, err := telegram.LoadConfig(configPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			fmt.Println("Telegram bridge: not configured")
			fmt.Println()
			fmt.Println("Run 'gt-telegram configure --help' to get started.")
			return nil
		}
		return fmt.Errorf("loading config: %w", err)
	}

	if *jsonOutput {
		out := struct {
			Token     string   `json:"token"`
			ChatID    int64    `json:"chat_id"`
			AllowFrom []int64  `json:"allow_from,omitempty"`
			Target    string   `json:"target,omitempty"`
			Enabled   bool     `json:"enabled"`
			Notify    []string `json:"notify,omitempty"`
			RateLimit int      `json:"rate_limit,omitempty"`
		}{
			Token:     cfg.MaskedToken(),
			ChatID:    cfg.ChatID,
			AllowFrom: cfg.AllowFrom,
			Target:    cfg.Target,
			Enabled:   cfg.Enabled,
			Notify:    cfg.Notify,
			RateLimit: cfg.RateLimit,
		}
		data, err := json.MarshalIndent(out, "", "  ")
		if err != nil {
			return fmt.Errorf("encoding JSON: %w", err)
		}
		fmt.Println(string(data))
		return nil
	}

	enabledStr := "no"
	if cfg.IsEnabled() {
		enabledStr = "yes"
	}
	fmt.Printf("Telegram bridge status\n")
	fmt.Printf("  Config:    %s\n", configPath)
	fmt.Printf("  Enabled:   %s\n", enabledStr)
	fmt.Printf("  Token:     %s\n", cfg.MaskedToken())
	fmt.Printf("  Chat ID:   %d\n", cfg.ChatID)
	fmt.Printf("  Target:    %s\n", cfg.Target)
	if len(cfg.AllowFrom) > 0 {
		fmt.Printf("  Allow from: %v\n", cfg.AllowFrom)
	} else {
		fmt.Printf("  Allow from: (none — all users blocked)\n")
	}
	if len(cfg.Notify) > 0 {
		fmt.Printf("  Notify:    %v\n", cfg.Notify)
	}
	if cfg.RateLimit > 0 {
		fmt.Printf("  Rate limit: %d msg/min\n", cfg.RateLimit)
	}
	return nil
}

// --- run ---

func runBridge() error {
	root := townRoot()
	configPath := telegram.ConfigPath(root)
	cfg, err := telegram.LoadConfig(configPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("telegram bridge is not configured — run 'gt-telegram configure' first")
		}
		return fmt.Errorf("loading config: %w", err)
	}

	if !cfg.IsEnabled() {
		return fmt.Errorf("telegram bridge is disabled — set enabled=true in %s", configPath)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-sigCh
		signal.Reset(syscall.SIGTERM, syscall.SIGINT)
		fmt.Println("\nShutting down Telegram bridge...")
		cancel()
	}()

	sender := telegram.NewCLISender(root)
	bridge := telegram.NewBridge(cfg, sender, root)

	fmt.Printf("Starting Telegram bridge (token: %s, chat: %d)...\n", cfg.MaskedToken(), cfg.ChatID)

	if err := bridge.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
		return fmt.Errorf("bridge exited: %w", err)
	}
	return nil
}
