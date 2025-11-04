package usps

import (
	"context"
	"sync"
	"time"

	"github.com/my-eq/go-usps/models"
)

// maxBackoffAttempt is the maximum allowed backoff exponent (2^5 = 32) to prevent overflow.
const maxBackoffAttempt = 5

// calculateBackoff calculates exponential backoff duration safely without overflow
func calculateBackoff(base time.Duration, attempt int) time.Duration {
	// Cap at maxBackoffAttempt (2^5 = 32) to prevent overflow
	if attempt > maxBackoffAttempt {
		attempt = maxBackoffAttempt
	}
	multiplier := 1 << uint(attempt)
	return base * time.Duration(multiplier)
}

// BulkConfig contains configuration options for bulk operations
type BulkConfig struct {
	// MaxConcurrency is the maximum number of concurrent requests (default: 10)
	MaxConcurrency int
	// RequestsPerSecond is the rate limit for API requests (default: 10)
	RequestsPerSecond int
	// MaxRetries is the maximum number of retry attempts for failed requests (default: 3)
	MaxRetries int
	// RetryBackoff is the base duration for exponential backoff (default: 1 second)
	RetryBackoff time.Duration
	// ProgressCallback is called after each request completes (optional)
	ProgressCallback func(completed, total int, err error)
}

// DefaultBulkConfig returns a BulkConfig with sensible defaults
func DefaultBulkConfig() *BulkConfig {
	return &BulkConfig{
		MaxConcurrency:    10,
		RequestsPerSecond: 10,
		MaxRetries:        3,
		RetryBackoff:      1 * time.Second,
	}
}

// AddressResult represents the result of a bulk address validation
type AddressResult struct {
	Index    int
	Request  *models.AddressRequest
	Response *models.AddressResponse
	Error    error
}

// CityStateResult represents the result of a bulk city/state lookup
type CityStateResult struct {
	Index    int
	Request  *models.CityStateRequest
	Response *models.CityStateResponse
	Error    error
}

// ZIPCodeResult represents the result of a bulk ZIP code lookup
type ZIPCodeResult struct {
	Index    int
	Request  *models.ZIPCodeRequest
	Response *models.ZIPCodeResponse
	Error    error
}

// BulkProcessor handles bulk operations with rate limiting and retries
type BulkProcessor struct {
	client *Client
	config *BulkConfig
}

// NewBulkProcessor creates a new BulkProcessor with the given client and config
func NewBulkProcessor(client *Client, config *BulkConfig) *BulkProcessor {
	if config == nil {
		config = DefaultBulkConfig()
	}
	return &BulkProcessor{
		client: client,
		config: config,
	}
}

// rateLimiter implements a simple token bucket rate limiter using only stdlib
type rateLimiter struct {
	tokens     int
	maxTokens  int
	refillRate time.Duration
	lastRefill time.Time
	mu         sync.Mutex
}

// newRateLimiter creates a new rate limiter
func newRateLimiter(requestsPerSecond int) *rateLimiter {
	return &rateLimiter{
		tokens:     requestsPerSecond,
		maxTokens:  requestsPerSecond,
		refillRate: time.Second / time.Duration(requestsPerSecond),
		lastRefill: time.Now(),
	}
}

// wait blocks until a token is available, respecting context cancellation
func (rl *rateLimiter) wait(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		rl.mu.Lock()
		select {
		case <-ctx.Done():
			rl.mu.Unlock()
			return ctx.Err()
		default:
		}

		// Refill tokens based on time elapsed
		now := time.Now()
		elapsed := now.Sub(rl.lastRefill)
		tokensToAdd := int(elapsed / rl.refillRate)

		if tokensToAdd > 0 {
			rl.tokens += tokensToAdd
			if rl.tokens > rl.maxTokens {
				rl.tokens = rl.maxTokens
			}
			rl.lastRefill = rl.lastRefill.Add(time.Duration(tokensToAdd) * rl.refillRate)
		}

		// Try to acquire a token
		if rl.tokens > 0 {
			rl.tokens--
			rl.mu.Unlock()
			return nil
		}

		rl.mu.Unlock()

		// Sleep briefly before retrying (half the refill rate to poll efficiently
		// without busy-waiting, ensuring we check for new tokens roughly twice
		// per token availability period)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(rl.refillRate / 2):
		}
	}
}

// ProcessAddresses validates multiple addresses concurrently with rate limiting
func (bp *BulkProcessor) ProcessAddresses(ctx context.Context, requests []*models.AddressRequest) []*AddressResult {
	results := make([]*AddressResult, len(requests))
	for i := range results {
		results[i] = &AddressResult{Index: i, Request: requests[i]}
	}

	bp.processBulk(ctx, len(requests), func(idx int, limiter *rateLimiter) error {
		resp, err := bp.processWithRetry(ctx, limiter, func() (interface{}, error) {
			return bp.client.GetAddress(ctx, requests[idx])
		})

		if err != nil {
			results[idx].Error = err
		} else {
			results[idx].Response = resp.(*models.AddressResponse)
		}
		return err
	}, func(idx int, err error) {
		if bp.config.ProgressCallback != nil {
			bp.config.ProgressCallback(idx+1, len(requests), err)
		}
	})

	return results
}

// ProcessCityStates looks up city/state for multiple ZIP codes concurrently with rate limiting
func (bp *BulkProcessor) ProcessCityStates(ctx context.Context, requests []*models.CityStateRequest) []*CityStateResult {
	results := make([]*CityStateResult, len(requests))
	for i := range results {
		results[i] = &CityStateResult{Index: i, Request: requests[i]}
	}

	bp.processBulk(ctx, len(requests), func(idx int, limiter *rateLimiter) error {
		resp, err := bp.processWithRetry(ctx, limiter, func() (interface{}, error) {
			return bp.client.GetCityState(ctx, requests[idx])
		})

		if err != nil {
			results[idx].Error = err
		} else {
			results[idx].Response = resp.(*models.CityStateResponse)
		}
		return err
	}, func(idx int, err error) {
		if bp.config.ProgressCallback != nil {
			bp.config.ProgressCallback(idx+1, len(requests), err)
		}
	})

	return results
}

// ProcessZIPCodes looks up ZIP codes for multiple addresses concurrently with rate limiting
func (bp *BulkProcessor) ProcessZIPCodes(ctx context.Context, requests []*models.ZIPCodeRequest) []*ZIPCodeResult {
	results := make([]*ZIPCodeResult, len(requests))
	for i := range results {
		results[i] = &ZIPCodeResult{Index: i, Request: requests[i]}
	}

	bp.processBulk(ctx, len(requests), func(idx int, limiter *rateLimiter) error {
		resp, err := bp.processWithRetry(ctx, limiter, func() (interface{}, error) {
			return bp.client.GetZIPCode(ctx, requests[idx])
		})

		if err != nil {
			results[idx].Error = err
		} else {
			results[idx].Response = resp.(*models.ZIPCodeResponse)
		}
		return err
	}, func(idx int, err error) {
		if bp.config.ProgressCallback != nil {
			bp.config.ProgressCallback(idx+1, len(requests), err)
		}
	})

	return results
}

// processBulk is a generic helper that handles the concurrent processing logic
// with semaphore-based concurrency control
func (bp *BulkProcessor) processBulk(
	ctx context.Context,
	count int,
	processFunc func(idx int, limiter *rateLimiter) error,
	progressFunc func(idx int, err error),
) {
	limiter := newRateLimiter(bp.config.RequestsPerSecond)
	sem := make(chan struct{}, bp.config.MaxConcurrency)
	var wg sync.WaitGroup

	for i := 0; i < count; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			// Acquire semaphore slot
			select {
			case sem <- struct{}{}:
				defer func() { <-sem }()
			case <-ctx.Done():
				progressFunc(idx, ctx.Err())
				return
			}

			// Process the request
			err := processFunc(idx, limiter)

			// Report progress
			progressFunc(idx, err)
		}(i)
	}

	wg.Wait()
}

// processWithRetry handles the retry logic with exponential backoff and rate limiting
func (bp *BulkProcessor) processWithRetry(
	ctx context.Context,
	limiter *rateLimiter,
	apiCall func() (interface{}, error),
) (interface{}, error) {
	var resp interface{}
	var err error

	for attempt := 0; attempt <= bp.config.MaxRetries; attempt++ {
		// Wait for rate limiter
		if err := limiter.wait(ctx); err != nil {
			return nil, err
		}

		resp, err = apiCall()
		if err == nil {
			return resp, nil
		}

		// Check if error is retryable
		if !isRetryableError(err) {
			return nil, err
		}

		// Exponential backoff
		if attempt < bp.config.MaxRetries {
			backoff := calculateBackoff(bp.config.RetryBackoff, attempt)
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff):
			}
		}
	}

	return nil, err
}

// isRetryableError determines if an error should trigger a retry
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// Check for API errors
	if apiErr, ok := err.(*APIError); ok {
		// Retry on 429 (rate limit), 500, 503 (service unavailable)
		return apiErr.StatusCode == 429 || apiErr.StatusCode >= 500
	}

	// Don't retry on context errors
	if err == context.Canceled || err == context.DeadlineExceeded {
		return false
	}

	// Retry on other network errors
	return true
}
