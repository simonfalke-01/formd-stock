package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

// HTTPClient wraps http.Client with ETag support
type HTTPClient struct {
	client    *http.Client
	userAgent string
	etag      string
}

// NewHTTPClient creates an optimized HTTP client
func NewHTTPClient(userAgent string) *HTTPClient {
	return &HTTPClient{
		client: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 100,
				IdleConnTimeout:     90 * time.Second,
				DisableCompression:  false,
				ForceAttemptHTTP2:   true,
			},
		},
		userAgent: userAgent,
	}
}

// FetchResult contains the response data and metadata
type FetchResult struct {
	Body         []byte
	StatusCode   int
	NotModified  bool
	RateLimited  bool
	ServerError  bool
	NewETag      string
}

// Fetch makes an HTTP GET request with ETag support
func (c *HTTPClient) Fetch(ctx context.Context, url string) (*FetchResult, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	// Set headers
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "application/json")

	// Add ETag if we have one
	if c.etag != "" {
		req.Header.Set("If-None-Match", c.etag)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	result := &FetchResult{
		StatusCode: resp.StatusCode,
		NewETag:    resp.Header.Get("ETag"),
	}

	// Handle status codes
	switch {
	case resp.StatusCode == http.StatusNotModified:
		result.NotModified = true
		return result, nil

	case resp.StatusCode == http.StatusTooManyRequests:
		result.RateLimited = true
		return result, nil

	case resp.StatusCode >= 500:
		result.ServerError = true
		return result, nil

	case resp.StatusCode != http.StatusOK:
		return result, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Read body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	result.Body = body

	// Update stored ETag
	if result.NewETag != "" {
		c.etag = result.NewETag
	}

	return result, nil
}

// ResetETag clears the stored ETag (useful for testing)
func (c *HTTPClient) ResetETag() {
	c.etag = ""
}
