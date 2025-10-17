package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/rand"
	"time"
)

// Monitor handles the stock monitoring loop
type Monitor struct {
	client    *HTTPClient
	state     *StateManager
	notifier  *TelegramNotifier
	config    *Config
	backoff   *ExponentialBackoff
}

// NewMonitor creates a new monitor
func NewMonitor(cfg *Config) (*Monitor, error) {
	client := NewHTTPClient(cfg.UserAgent)
	state := NewStateManager()

	var notifier *TelegramNotifier
	var err error
	if cfg.TelegramToken != "" && cfg.TelegramChatID != 0 {
		notifier, err = NewTelegramNotifier(cfg.TelegramToken, cfg.TelegramChatID, cfg.ShopURL)
		if err != nil {
			return nil, fmt.Errorf("creating telegram notifier: %w", err)
		}
	}

	return &Monitor{
		client:   client,
		state:    state,
		notifier: notifier,
		config:   cfg,
		backoff:  NewExponentialBackoff(cfg.PollInterval, 5*time.Minute),
	}, nil
}

// Start begins the monitoring loop
func (m *Monitor) Start(ctx context.Context) error {
	log.Printf("Starting monitor for %s%s", m.config.ShopURL, m.config.CollectionPath)
	log.Printf("Poll interval: %s", m.config.PollInterval)

	if m.notifier != nil {
		if err := m.notifier.SendMessage("🤖 Stock monitor started"); err != nil {
			log.Printf("Failed to send start notification: %v", err)
		}
	}

	ticker := time.NewTicker(m.config.PollInterval)
	defer ticker.Stop()

	// Do initial poll
	if err := m.poll(ctx); err != nil {
		log.Printf("Initial poll failed: %v", err)
	}

	for {
		select {
		case <-ctx.Done():
			log.Println("Monitor stopped")
			if m.notifier != nil {
				_ = m.notifier.SendMessage("🛑 Stock monitor stopped")
			}
			return ctx.Err()

		case <-ticker.C:
			if err := m.poll(ctx); err != nil {
				log.Printf("Poll failed: %v", err)
			}
		}
	}
}

// poll performs a single polling operation
func (m *Monitor) poll(ctx context.Context) error {
	url := m.config.ShopURL + m.config.CollectionPath

	result, err := m.client.Fetch(ctx, url)
	if err != nil {
		m.backoff.Failed()
		return fmt.Errorf("fetching data: %w", err)
	}

	// Handle different response types
	switch {
	case result.NotModified:
		log.Println("No changes (304 Not Modified)")
		m.backoff.Success()
		return nil

	case result.RateLimited:
		delay := m.backoff.Failed()
		log.Printf("Rate limited (429) - backing off for %s", delay)
		time.Sleep(delay)
		return fmt.Errorf("rate limited")

	case result.ServerError:
		delay := m.backoff.Failed()
		log.Printf("Server error (%d) - backing off for %s", result.StatusCode, delay)
		time.Sleep(delay)
		return fmt.Errorf("server error: %d", result.StatusCode)
	}

	// Parse response
	var response ShopifyResponse
	if err := json.Unmarshal(result.Body, &response); err != nil {
		return fmt.Errorf("parsing response: %w", err)
	}

	log.Printf("Fetched %d products", len(response.Products))

	// Check for changes
	changes := m.state.CheckAndUpdate(response.Products)
	if len(changes) > 0 {
		log.Printf("Detected %d stock changes", len(changes))

		// Send notifications
		if m.notifier != nil {
			if err := m.notifier.NotifyMultiple(changes); err != nil {
				log.Printf("Failed to send notifications: %v", err)
			}
		}

		// Log changes
		for _, change := range changes {
			if change.IsNewStock() {
				log.Printf("NEW STOCK: %s - %s ($%s)",
					change.ProductTitle,
					change.VariantTitle,
					change.VariantPrice,
				)
			}
		}
	}

	m.backoff.Success()
	return nil
}

// ExponentialBackoff implements exponential backoff with jitter
type ExponentialBackoff struct {
	baseDelay  time.Duration
	maxDelay   time.Duration
	multiplier float64
	failures   int
}

// NewExponentialBackoff creates a new backoff manager
func NewExponentialBackoff(baseDelay, maxDelay time.Duration) *ExponentialBackoff {
	return &ExponentialBackoff{
		baseDelay:  baseDelay,
		maxDelay:   maxDelay,
		multiplier: 2.0,
		failures:   0,
	}
}

// Failed records a failure and returns the backoff duration
func (eb *ExponentialBackoff) Failed() time.Duration {
	eb.failures++

	// Calculate exponential delay
	delay := float64(eb.baseDelay) * math.Pow(eb.multiplier, float64(eb.failures-1))

	// Cap at max delay
	if delay > float64(eb.maxDelay) {
		delay = float64(eb.maxDelay)
	}

	// Add jitter (±20%)
	jitter := delay * 0.2 * (rand.Float64()*2 - 1)
	delay += jitter

	return time.Duration(delay)
}

// Success resets the failure counter
func (eb *ExponentialBackoff) Success() {
	eb.failures = 0
}
