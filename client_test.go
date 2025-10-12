package usps

import (
	"context"
	"net/http"
	"testing"
	"time"
)

// MockHTTPClient is a mock implementation of HTTPClient for testing
type MockHTTPClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return m.DoFunc(req)
}

func TestNewClient(t *testing.T) {
	apiKey := "test-api-key"
	client := NewClient(apiKey)

	if client.apiKey != apiKey {
		t.Errorf("expected apiKey %s, got %s", apiKey, client.apiKey)
	}

	if client.baseURL != baseURL {
		t.Errorf("expected baseURL %s, got %s", baseURL, client.baseURL)
	}

	if client.httpClient == nil {
		t.Error("expected httpClient to be initialized")
	}
}

func TestWithHTTPClient(t *testing.T) {
	apiKey := "test-api-key"
	mockClient := &MockHTTPClient{}

	client := NewClient(apiKey, WithHTTPClient(mockClient))

	if client.httpClient != mockClient {
		t.Error("expected custom HTTP client to be set")
	}
}

func TestWithBaseURL(t *testing.T) {
	apiKey := "test-api-key"
	customURL := "https://custom.api.com"

	client := NewClient(apiKey, WithBaseURL(customURL))

	if client.baseURL != customURL {
		t.Errorf("expected baseURL %s, got %s", customURL, client.baseURL)
	}
}

func TestDefaultRetryConfig(t *testing.T) {
	config := DefaultRetryConfig()

	if config.MaxRetries != 3 {
		t.Errorf("expected MaxRetries 3, got %d", config.MaxRetries)
	}

	if config.InitialBackoff != 500*time.Millisecond {
		t.Errorf("expected InitialBackoff 500ms, got %v", config.InitialBackoff)
	}

	if config.MaxBackoff != 5*time.Second {
		t.Errorf("expected MaxBackoff 5s, got %v", config.MaxBackoff)
	}

	if config.Multiplier != 2.0 {
		t.Errorf("expected Multiplier 2.0, got %f", config.Multiplier)
	}
}

func TestClientOptions(t *testing.T) {
	apiKey := "test-api-key"
	customURL := "https://custom.api.com"
	mockClient := &MockHTTPClient{}

	client := NewClient(
		apiKey,
		WithHTTPClient(mockClient),
		WithBaseURL(customURL),
	)

	if client.apiKey != apiKey {
		t.Errorf("expected apiKey %s, got %s", apiKey, client.apiKey)
	}

	if client.baseURL != customURL {
		t.Errorf("expected baseURL %s, got %s", customURL, client.baseURL)
	}

	if client.httpClient != mockClient {
		t.Error("expected custom HTTP client to be set")
	}
}

func TestDoRequest(t *testing.T) {
	apiKey := "test-api-key"
	callCount := 0

	mockClient := &MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			callCount++

			// Verify headers
			if req.Header.Get("Content-Type") != "application/json" {
				t.Error("expected Content-Type header to be application/json")
			}

			expectedAuth := "Bearer " + apiKey
			if req.Header.Get("Authorization") != expectedAuth {
				t.Errorf("expected Authorization header %s, got %s", expectedAuth, req.Header.Get("Authorization"))
			}

			return &http.Response{
				StatusCode: 200,
				Body:       http.NoBody,
			}, nil
		},
	}

	client := NewClient(apiKey, WithHTTPClient(mockClient))

	_, err := client.do(context.Background(), "GET", "/test", nil)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if callCount != 1 {
		t.Errorf("expected 1 call, got %d", callCount)
	}
}
