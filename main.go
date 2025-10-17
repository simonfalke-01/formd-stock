package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// Parse flags
	configPath := flag.String("config", "", "Path to config file (optional, uses env vars if not provided)")
	flag.Parse()

	// Load configuration
	var cfg *Config
	var err error

	if *configPath != "" {
		cfg, err = LoadConfig(*configPath)
		if err != nil {
			log.Fatalf("Failed to load config from file: %v", err)
		}
		log.Printf("Loaded config from %s", *configPath)
	} else {
		cfg = LoadConfigFromEnv()
		log.Println("Loaded config from environment variables")
	}

	// Validate configuration
	if err := validateConfig(cfg); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	// Create monitor
	monitor, err := NewMonitor(cfg)
	if err != nil {
		log.Fatalf("Failed to create monitor: %v", err)
	}

	// Setup context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		log.Printf("Received signal %v, shutting down gracefully...", sig)
		cancel()
	}()

	// Start monitoring
	if err := monitor.Start(ctx); err != nil {
		if err != context.Canceled {
			log.Fatalf("Monitor error: %v", err)
		}
	}

	log.Println("Shutdown complete")
}

func validateConfig(cfg *Config) error {
	if cfg.ShopURL == "" {
		return ErrMissingShopURL
	}
	if cfg.TelegramToken == "" {
		log.Println("Warning: No Telegram token provided, notifications will be disabled")
	}
	if cfg.TelegramChatID == 0 {
		log.Println("Warning: No Telegram chat ID provided, notifications will be disabled")
	}
	return nil
}

// Errors
var (
	ErrMissingShopURL = &ConfigError{"SHOP_URL is required"}
)

type ConfigError struct {
	Message string
}

func (e *ConfigError) Error() string {
	return e.Message
}
