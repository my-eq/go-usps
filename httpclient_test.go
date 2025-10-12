package usps

import (
	"errors"
	"net/http"
	"testing"
	"time"
)

func TestNewDefaultHTTPClient(t *testing.T) {
	client := NewDefaultHTTPClient(nil)

	if client.client == nil {
		t.Error("expected client to be initialized")
	}

	if client.config == nil {
		t.Error("expected config to be initialized")
	}

	// Verify default config values
	if client.config.MaxRetries != 3 {
		t.Errorf("expected MaxRetries 3, got %d", client.config.MaxRetries)
	}
}

func TestDefaultHTTPClientWithCustomConfig(t *testing.T) {
	config := &RetryConfig{
		MaxRetries:     5,
		InitialBackoff: 1 * time.Second,
		MaxBackoff:     10 * time.Second,
		Multiplier:     3.0,
	}

	client := NewDefaultHTTPClient(config)

	if client.config.MaxRetries != 5 {
		t.Errorf("expected MaxRetries 5, got %d", client.config.MaxRetries)
	}

	if client.config.InitialBackoff != 1*time.Second {
		t.Errorf("expected InitialBackoff 1s, got %v", client.config.InitialBackoff)
	}

	if client.config.MaxBackoff != 10*time.Second {
		t.Errorf("expected MaxBackoff 10s, got %v", client.config.MaxBackoff)
	}

	if client.config.Multiplier != 3.0 {
		t.Errorf("expected Multiplier 3.0, got %f", client.config.Multiplier)
	}
}

func TestCalculateBackoff(t *testing.T) {
	tests := []struct {
		name           string
		initialBackoff time.Duration
		attempt        int
		multiplier     float64
		maxBackoff     time.Duration
		expected       time.Duration
	}{
		{
			name:           "first backoff",
			initialBackoff: 100 * time.Millisecond,
			attempt:        0,
			multiplier:     2.0,
			maxBackoff:     5 * time.Second,
			expected:       100 * time.Millisecond,
		},
		{
			name:           "second backoff",
			initialBackoff: 100 * time.Millisecond,
			attempt:        1,
			multiplier:     2.0,
			maxBackoff:     5 * time.Second,
			expected:       200 * time.Millisecond,
		},
		{
			name:           "max backoff reached",
			initialBackoff: 100 * time.Millisecond,
			attempt:        10,
			multiplier:     2.0,
			maxBackoff:     1 * time.Second,
			expected:       1 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateBackoff(tt.initialBackoff, tt.attempt, tt.multiplier, tt.maxBackoff)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestDefaultHTTPClientRetry(t *testing.T) {
	attempts := 0
	config := &RetryConfig{
		MaxRetries:     2,
		InitialBackoff: 10 * time.Millisecond,
		MaxBackoff:     100 * time.Millisecond,
		Multiplier:     2.0,
	}

	client := NewDefaultHTTPClient(config)

	// Create a mock roundtripper that fails twice then succeeds
	client.client.Transport = &mockRoundTripper{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			attempts++
			if attempts < 3 {
				return nil, errors.New("temporary error")
			}
			return &http.Response{
				StatusCode: 200,
				Body:       http.NoBody,
			}, nil
		},
	}

	req, _ := http.NewRequest("GET", "http://example.com", nil)
	resp, err := client.Do(req)

	if err != nil {
		t.Errorf("expected success after retries, got error: %v", err)
	}

	if resp.StatusCode != 200 {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	if attempts != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts)
	}
}

type mockRoundTripper struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.DoFunc(req)
}
