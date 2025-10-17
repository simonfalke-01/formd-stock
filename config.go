package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type Config struct {
	ShopURL        string        `json:"shop_url"`
	CollectionPath string        `json:"collection_path"`
	PollInterval   time.Duration `json:"poll_interval"`
	TelegramToken  string        `json:"telegram_token"`
	TelegramChatID int64         `json:"telegram_chat_id"`
	UserAgent      string        `json:"user_agent"`
}

func LoadConfig(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var cfg Config
	if err := json.NewDecoder(file).Decode(&cfg); err != nil {
		return nil, err
	}

	// Set defaults
	if cfg.PollInterval == 0 {
		cfg.PollInterval = 15 * time.Second
	}
	if cfg.UserAgent == "" {
		cfg.UserAgent = "FormD-Stock-Monitor/1.0"
	}
	if cfg.CollectionPath == "" {
		cfg.CollectionPath = "/collections/all/products.json?limit=250"
	}

	return &cfg, nil
}

func LoadConfigFromEnv() *Config {
	return &Config{
		ShopURL:        getEnv("SHOP_URL", "https://formdt1.com"),
		CollectionPath: getEnv("COLLECTION_PATH", "/collections/all/products.json?limit=250"),
		PollInterval:   getDurationEnv("POLL_INTERVAL", 15*time.Second),
		TelegramToken:  os.Getenv("TELEGRAM_TOKEN"),
		TelegramChatID: getInt64Env("TELEGRAM_CHAT_ID", 0),
		UserAgent:      getEnv("USER_AGENT", "FormD-Stock-Monitor/1.0"),
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getDurationEnv(key string, fallback time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if d, err := time.ParseDuration(value); err == nil {
			return d
		}
	}
	return fallback
}

func getInt64Env(key string, fallback int64) int64 {
	if value := os.Getenv(key); value != "" {
		var i int64
		if _, err := fmt.Sscan(value, &i); err == nil {
			return i
		}
	}
	return fallback
}
