package usps

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/my-eq/go-usps/models"
)

func TestDefaultBulkConfig(t *testing.T) {
	config := DefaultBulkConfig()

	if config.MaxConcurrency != 10 {
		t.Errorf("Expected MaxConcurrency=10, got %d", config.MaxConcurrency)
	}
	if config.RequestsPerSecond != 10 {
		t.Errorf("Expected RequestsPerSecond=10, got %d", config.RequestsPerSecond)
	}
	if config.MaxRetries != 3 {
		t.Errorf("Expected MaxRetries=3, got %d", config.MaxRetries)
	}
	if config.RetryBackoff != 1*time.Second {
		t.Errorf("Expected RetryBackoff=1s, got %v", config.RetryBackoff)
	}
}

func TestNewBulkProcessor(t *testing.T) {
	tokenProvider := NewStaticTokenProvider("test-token")
	client := NewClient(tokenProvider)

	t.Run("with custom config", func(t *testing.T) {
		config := &BulkConfig{
			MaxConcurrency:    5,
			RequestsPerSecond: 20,
			MaxRetries:        2,
			RetryBackoff:      500 * time.Millisecond,
		}
		processor := NewBulkProcessor(client, config)

		if processor.client != client {
			t.Error("Expected client to be set")
		}
		if processor.config != config {
			t.Error("Expected config to be set")
		}
	})

	t.Run("with nil config uses defaults", func(t *testing.T) {
		processor := NewBulkProcessor(client, nil)

		if processor.config == nil {
			t.Fatal("Expected config to be initialized")
		}
		if processor.config.MaxConcurrency != 10 {
			t.Errorf("Expected default MaxConcurrency=10, got %d", processor.config.MaxConcurrency)
		}
	})
}

func TestRateLimiter(t *testing.T) {
	t.Run("basic rate limiting", func(t *testing.T) {
		limiter := newRateLimiter(5) // 5 requests per second
		ctx := context.Background()

		start := time.Now()

		// First 5 requests should be immediate
		for i := 0; i < 5; i++ {
			if err := limiter.wait(ctx); err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
		}

		immediate := time.Since(start)
		if immediate > 100*time.Millisecond {
			t.Errorf("First 5 requests took too long: %v", immediate)
		}

		// Next request should wait
		if err := limiter.wait(ctx); err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		total := time.Since(start)
		if total < 200*time.Millisecond {
			t.Errorf("Rate limiter didn't wait long enough: %v", total)
		}
	})

	t.Run("context cancellation", func(t *testing.T) {
		limiter := newRateLimiter(1)
		ctx, cancel := context.WithCancel(context.Background())

		// Exhaust tokens
		_ = limiter.wait(ctx)

		// Cancel context
		cancel()

		// Should return context error
		err := limiter.wait(ctx)
		if err == nil {
			t.Error("Expected error from cancelled context")
		}
		if err != context.Canceled {
			t.Errorf("Expected context.Canceled, got %v", err)
		}
	})
}

func TestProcessAddresses_Success(t *testing.T) {
	var requestCount int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requestCount, 1)

		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Error("Missing or invalid Authorization header")
		}

		resp := models.AddressResponse{
			Address: &models.DomesticAddress{
				Address: models.Address{
					StreetAddress: "123 MAIN ST",
				},
				City:    "NEW YORK",
				State:   "NY",
				ZIPCode: "10001",
			},
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	tokenProvider := NewStaticTokenProvider("test-token")
	client := NewClient(tokenProvider, WithBaseURL(server.URL))

	config := &BulkConfig{
		MaxConcurrency:    3,
		RequestsPerSecond: 50, // High rate for faster testing
		MaxRetries:        1,
		RetryBackoff:      10 * time.Millisecond,
	}
	processor := NewBulkProcessor(client, config)

	requests := []*models.AddressRequest{
		{StreetAddress: "123 Main St", City: "New York", State: "NY"},
		{StreetAddress: "456 Oak Ave", City: "New York", State: "NY"},
		{StreetAddress: "789 Elm Blvd", City: "New York", State: "NY"},
	}

	results := processor.ProcessAddresses(context.Background(), requests)

	if len(results) != 3 {
		t.Fatalf("Expected 3 results, got %d", len(results))
	}

	for i, result := range results {
		if result.Index != i {
			t.Errorf("Result %d has wrong index: %d", i, result.Index)
		}
		if result.Request != requests[i] {
			t.Errorf("Result %d has wrong request", i)
		}
		if result.Error != nil {
			t.Errorf("Result %d has unexpected error: %v", i, result.Error)
		}
		if result.Response == nil {
			t.Errorf("Result %d has nil response", i)
		} else if result.Response.Address.StreetAddress != "123 MAIN ST" {
			t.Errorf("Result %d has wrong address: %s", i, result.Response.Address.StreetAddress)
		}
	}

	if atomic.LoadInt32(&requestCount) != 3 {
		t.Errorf("Expected 3 requests, got %d", requestCount)
	}
}

func TestProcessAddresses_WithErrors(t *testing.T) {
	const expectedFailuresBeforeSuccess = 2 // Number of initial failures to test retry behavior
	callCount := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++

		// First requests fail with 500, should retry
		if callCount <= expectedFailuresBeforeSuccess {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(models.ErrorMessage{
				Error: &models.ErrorInfo{
					Message: "Internal server error",
				},
			})
			return
		}

		// Subsequent requests succeed
		resp := models.AddressResponse{
			Address: &models.DomesticAddress{
				Address: models.Address{
					StreetAddress: "123 MAIN ST",
				},
				City:    "NEW YORK",
				State:   "NY",
				ZIPCode: "10001",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	tokenProvider := NewStaticTokenProvider("test-token")
	client := NewClient(tokenProvider, WithBaseURL(server.URL))

	config := &BulkConfig{
		MaxConcurrency:    1, // Sequential for predictable testing
		RequestsPerSecond: 100,
		MaxRetries:        2,
		RetryBackoff:      10 * time.Millisecond,
	}
	processor := NewBulkProcessor(client, config)

	requests := []*models.AddressRequest{
		{StreetAddress: "123 Main St", City: "New York", State: "NY"},
	}

	results := processor.ProcessAddresses(context.Background(), requests)

	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}

	// Should succeed after retries
	if results[0].Error != nil {
		t.Errorf("Expected success after retries, got error: %v", results[0].Error)
	}
	if results[0].Response == nil {
		t.Error("Expected response after retries")
	}

	// Should have made 1 initial call + config.MaxRetries retries
	if callCount != 1+config.MaxRetries {
		t.Errorf("Expected %d calls (with retries), got %d", 1+config.MaxRetries, callCount)
	}
}

func TestProcessAddresses_NonRetryableError(t *testing.T) {
	callCount := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++

		// 400 errors should not be retried
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(models.ErrorMessage{
			Error: &models.ErrorInfo{
				Message: "Bad request",
			},
		})
	}))
	defer server.Close()

	tokenProvider := NewStaticTokenProvider("test-token")
	client := NewClient(tokenProvider, WithBaseURL(server.URL))

	config := &BulkConfig{
		MaxConcurrency:    1,
		RequestsPerSecond: 100,
		MaxRetries:        2,
		RetryBackoff:      10 * time.Millisecond,
	}
	processor := NewBulkProcessor(client, config)

	requests := []*models.AddressRequest{
		{StreetAddress: "123 Main St", City: "New York", State: "NY"},
	}

	results := processor.ProcessAddresses(context.Background(), requests)

	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}

	// Should fail without retrying
	if results[0].Error == nil {
		t.Error("Expected error for bad request")
	}

	// Should have made only 1 call (no retries for 400)
	if callCount != 1 {
		t.Errorf("Expected 1 call (no retries for 400), got %d", callCount)
	}
}

func TestProcessAddresses_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Slow response
		time.Sleep(100 * time.Millisecond)
		resp := models.AddressResponse{
			Address: &models.DomesticAddress{
				Address: models.Address{
					StreetAddress: "123 MAIN ST",
				},
				City:    "NEW YORK",
				State:   "NY",
				ZIPCode: "10001",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	tokenProvider := NewStaticTokenProvider("test-token")
	client := NewClient(tokenProvider, WithBaseURL(server.URL))

	config := &BulkConfig{
		MaxConcurrency:    5,
		RequestsPerSecond: 100,
		MaxRetries:        0,
		RetryBackoff:      10 * time.Millisecond,
	}
	processor := NewBulkProcessor(client, config)

	requests := []*models.AddressRequest{
		{StreetAddress: "123 Main St", City: "New York", State: "NY"},
		{StreetAddress: "456 Oak Ave", City: "New York", State: "NY"},
		{StreetAddress: "789 Elm Blvd", City: "New York", State: "NY"},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	results := processor.ProcessAddresses(ctx, requests)

	// At least some should have context errors
	errorCount := 0
	for _, result := range results {
		if result.Error != nil {
			errorCount++
			if !strings.Contains(result.Error.Error(), "context") {
				t.Logf("Warning: Got error but not context-related: %v", result.Error)
			}
		}
	}

	if errorCount == 0 {
		t.Error("Expected at least some context errors due to timeout")
	}
}

func TestProcessAddresses_ProgressCallback(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := models.AddressResponse{
			Address: &models.DomesticAddress{
				Address: models.Address{
					StreetAddress: "123 MAIN ST",
				},
				City:    "NEW YORK",
				State:   "NY",
				ZIPCode: "10001",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	tokenProvider := NewStaticTokenProvider("test-token")
	client := NewClient(tokenProvider, WithBaseURL(server.URL))

	var callbackCount int32
	config := &BulkConfig{
		MaxConcurrency:    2,
		RequestsPerSecond: 100,
		MaxRetries:        1,
		RetryBackoff:      10 * time.Millisecond,
		ProgressCallback: func(completed, total int, err error) {
			atomic.AddInt32(&callbackCount, 1)
			if completed < 1 || completed > total {
				t.Errorf("Invalid progress: completed=%d, total=%d", completed, total)
			}
			if total != 3 {
				t.Errorf("Expected total=3, got %d", total)
			}
		},
	}
	processor := NewBulkProcessor(client, config)

	requests := []*models.AddressRequest{
		{StreetAddress: "123 Main St", City: "New York", State: "NY"},
		{StreetAddress: "456 Oak Ave", City: "New York", State: "NY"},
		{StreetAddress: "789 Elm Blvd", City: "New York", State: "NY"},
	}

	processor.ProcessAddresses(context.Background(), requests)

	if atomic.LoadInt32(&callbackCount) != 3 {
		t.Errorf("Expected 3 callbacks, got %d", callbackCount)
	}
}

func TestProcessCityStates_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := models.CityStateResponse{
			City:    "NEW YORK",
			State:   "NY",
			ZIPCode: "10001",
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	tokenProvider := NewStaticTokenProvider("test-token")
	client := NewClient(tokenProvider, WithBaseURL(server.URL))

	config := &BulkConfig{
		MaxConcurrency:    2,
		RequestsPerSecond: 100,
		MaxRetries:        1,
		RetryBackoff:      10 * time.Millisecond,
	}
	processor := NewBulkProcessor(client, config)

	requests := []*models.CityStateRequest{
		{ZIPCode: "10001"},
		{ZIPCode: "90210"},
	}

	results := processor.ProcessCityStates(context.Background(), requests)

	if len(results) != 2 {
		t.Fatalf("Expected 2 results, got %d", len(results))
	}

	for i, result := range results {
		if result.Index != i {
			t.Errorf("Result %d has wrong index: %d", i, result.Index)
		}
		if result.Error != nil {
			t.Errorf("Result %d has unexpected error: %v", i, result.Error)
		}
		if result.Response == nil {
			t.Errorf("Result %d has nil response", i)
		}
	}
}

func TestProcessZIPCodes_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := models.ZIPCodeResponse{
			Address: &models.DomesticAddress{
				Address: models.Address{
					StreetAddress: "123 MAIN ST",
				},
				City:    "NEW YORK",
				State:   "NY",
				ZIPCode: "10001",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	tokenProvider := NewStaticTokenProvider("test-token")
	client := NewClient(tokenProvider, WithBaseURL(server.URL))

	config := &BulkConfig{
		MaxConcurrency:    2,
		RequestsPerSecond: 100,
		MaxRetries:        1,
		RetryBackoff:      10 * time.Millisecond,
	}
	processor := NewBulkProcessor(client, config)

	requests := []*models.ZIPCodeRequest{
		{StreetAddress: "123 Main St", City: "New York", State: "NY"},
		{StreetAddress: "456 Oak Ave", City: "Los Angeles", State: "CA"},
	}

	results := processor.ProcessZIPCodes(context.Background(), requests)

	if len(results) != 2 {
		t.Fatalf("Expected 2 results, got %d", len(results))
	}

	for i, result := range results {
		if result.Index != i {
			t.Errorf("Result %d has wrong index: %d", i, result.Index)
		}
		if result.Error != nil {
			t.Errorf("Result %d has unexpected error: %v", i, result.Error)
		}
		if result.Response == nil {
			t.Errorf("Result %d has nil response", i)
		}
	}
}

func TestIsRetryableError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name: "429 rate limit error",
			err: &APIError{
				StatusCode: 429,
				ErrorMessage: models.ErrorMessage{
					Error: &models.ErrorInfo{Message: "Rate limit exceeded"},
				},
			},
			expected: true,
		},
		{
			name: "500 server error",
			err: &APIError{
				StatusCode: 500,
				ErrorMessage: models.ErrorMessage{
					Error: &models.ErrorInfo{Message: "Internal server error"},
				},
			},
			expected: true,
		},
		{
			name: "503 service unavailable",
			err: &APIError{
				StatusCode: 503,
				ErrorMessage: models.ErrorMessage{
					Error: &models.ErrorInfo{Message: "Service unavailable"},
				},
			},
			expected: true,
		},
		{
			name: "400 bad request",
			err: &APIError{
				StatusCode: 400,
				ErrorMessage: models.ErrorMessage{
					Error: &models.ErrorInfo{Message: "Bad request"},
				},
			},
			expected: false,
		},
		{
			name: "404 not found",
			err: &APIError{
				StatusCode: 404,
				ErrorMessage: models.ErrorMessage{
					Error: &models.ErrorInfo{Message: "Not found"},
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isRetryableError(tt.err)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v for error: %v", tt.expected, result, tt.err)
			}
		})
	}
}

func TestBulkProcessor_RateLimiting(t *testing.T) {
	var requestCount int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requestCount, 1)

		resp := models.AddressResponse{
			Address: &models.DomesticAddress{
				Address: models.Address{
					StreetAddress: "123 MAIN ST",
				},
				City:    "NEW YORK",
				State:   "NY",
				ZIPCode: "10001",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	tokenProvider := NewStaticTokenProvider("test-token")
	client := NewClient(tokenProvider, WithBaseURL(server.URL))

	// Configure for 5 requests per second
	config := &BulkConfig{
		MaxConcurrency:    10, // High concurrency to test rate limiting
		RequestsPerSecond: 5,
		MaxRetries:        0,
		RetryBackoff:      10 * time.Millisecond,
	}
	processor := NewBulkProcessor(client, config)

	// Send 10 requests - should take at least 2 seconds at 5 req/sec
	requests := make([]*models.AddressRequest, 10)
	for i := range requests {
		requests[i] = &models.AddressRequest{
			StreetAddress: "123 Main St",
			City:          "New York",
			State:         "NY",
		}
	}

	start := time.Now()
	results := processor.ProcessAddresses(context.Background(), requests)
	duration := time.Since(start)

	// Verify all succeeded
	for i, result := range results {
		if result.Error != nil {
			t.Errorf("Result %d has error: %v", i, result.Error)
		}
	}

	// The token bucket starts with maxTokens (5). First 5 requests go immediately,
	// then we need to wait for 5 more tokens at 1 token per (1/5) second = 1 second total.
	// With some timing variance, we expect at least 0.9 seconds.
	minExpectedDuration := 900 * time.Millisecond
	if duration < minExpectedDuration {
		t.Errorf("Rate limiting not working: 10 requests at 5/sec took only %v (expected at least %v)", duration, minExpectedDuration)
	}

	if atomic.LoadInt32(&requestCount) != 10 {
		t.Errorf("Expected 10 requests, got %d", requestCount)
	}
}
