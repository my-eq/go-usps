package usps

import (
	"math"
	"net/http"
	"time"
)

// DefaultHTTPClient wraps http.Client with retry and backoff logic
type DefaultHTTPClient struct {
	client *http.Client
	config *RetryConfig
}

// NewDefaultHTTPClient creates a new default HTTP client with retry/backoff
func NewDefaultHTTPClient(config *RetryConfig) *DefaultHTTPClient {
	if config == nil {
		config = DefaultRetryConfig()
	}

	return &DefaultHTTPClient{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		config: config,
	}
}

// Do executes the HTTP request with retry and exponential backoff
func (d *DefaultHTTPClient) Do(req *http.Request) (*http.Response, error) {
	var lastErr error
	backoff := d.config.InitialBackoff

	for attempt := 0; attempt <= d.config.MaxRetries; attempt++ {
		resp, err := d.client.Do(req)

		// Success - return immediately
		if err == nil && resp.StatusCode < 500 {
			return resp, nil
		}

		// Store the error
		lastErr = err
		if err == nil {
			resp.Body.Close()
		}

		// Don't retry on last attempt
		if attempt == d.config.MaxRetries {
			break
		}

		// Wait before retry with exponential backoff
		time.Sleep(backoff)

		// Calculate next backoff duration
		backoff = time.Duration(float64(backoff) * d.config.Multiplier)
		if backoff > d.config.MaxBackoff {
			backoff = d.config.MaxBackoff
		}
	}

	if lastErr != nil {
		return nil, lastErr
	}

	// Return the last response if no error but status >= 500
	return d.client.Do(req)
}

// calculateBackoff computes exponential backoff duration
func calculateBackoff(initialBackoff time.Duration, attempt int, multiplier float64, maxBackoff time.Duration) time.Duration {
	backoff := time.Duration(float64(initialBackoff) * math.Pow(multiplier, float64(attempt)))
	if backoff > maxBackoff {
		return maxBackoff
	}
	return backoff
}
