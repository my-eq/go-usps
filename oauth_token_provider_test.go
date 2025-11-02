package usps

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/my-eq/go-usps/models"
)

func TestNewOAuthTokenProvider(t *testing.T) {
	provider := NewOAuthTokenProvider("client-id", "client-secret")

	if provider.clientID != "client-id" {
		t.Errorf("Expected clientID 'client-id', got '%s'", provider.clientID)
	}
	if provider.clientSecret != "client-secret" {
		t.Errorf("Expected clientSecret 'client-secret', got '%s'", provider.clientSecret)
	}
	if provider.refreshBuffer != DefaultTokenRefreshBuffer {
		t.Errorf("Expected refreshBuffer %v, got %v", DefaultTokenRefreshBuffer, provider.refreshBuffer)
	}
	if provider.oauthClient == nil {
		t.Error("Expected oauthClient to be initialized")
	}
}

func TestNewOAuthTestTokenProvider(t *testing.T) {
	provider := NewOAuthTestTokenProvider("test-client-id", "test-client-secret")

	if provider.clientID != "test-client-id" {
		t.Errorf("Expected clientID 'test-client-id', got '%s'", provider.clientID)
	}
	if provider.clientSecret != "test-client-secret" {
		t.Errorf("Expected clientSecret 'test-client-secret', got '%s'", provider.clientSecret)
	}
	if provider.oauthClient.baseURL != OAuthTestingBaseURL {
		t.Errorf("Expected baseURL '%s', got '%s'", OAuthTestingBaseURL, provider.oauthClient.baseURL)
	}
}

func TestOAuthTokenProvider_WithOptions(t *testing.T) {
	provider := NewOAuthTokenProvider(
		"client-id",
		"client-secret",
		WithOAuthScopes("addresses tracking"),
		WithTokenRefreshBuffer(10*time.Minute),
	)

	if provider.scopes != "addresses tracking" {
		t.Errorf("Expected scopes 'addresses tracking', got '%s'", provider.scopes)
	}
	if provider.refreshBuffer != 10*time.Minute {
		t.Errorf("Expected refreshBuffer 10 minutes, got %v", provider.refreshBuffer)
	}
}

func TestOAuthTokenProvider_WithOAuthEnvironment(t *testing.T) {
	tests := []struct {
		name        string
		environment string
		expectedURL string
	}{
		{
			name:        "production environment",
			environment: "production",
			expectedURL: OAuthProductionBaseURL,
		},
		{
			name:        "testing environment",
			environment: "testing",
			expectedURL: OAuthTestingBaseURL,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := NewOAuthTokenProvider(
				"client-id",
				"client-secret",
				WithOAuthEnvironment(tt.environment),
			)

			if provider.oauthClient.baseURL != tt.expectedURL {
				t.Errorf("Expected baseURL '%s', got '%s'", tt.expectedURL, provider.oauthClient.baseURL)
			}
		})
	}
}

func TestOAuthTokenProvider_GetToken_Success(t *testing.T) {
	// Mock OAuth server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := models.ProviderAccessTokenResponse{
			AccessToken: "test-access-token",
			ExpiresIn:   28800, // 8 hours
			TokenType:   "Bearer",
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	provider := NewOAuthTokenProvider("client-id", "client-secret")
	provider.oauthClient = NewOAuthClient(WithBaseURL(server.URL))

	token, err := provider.GetToken(context.Background())
	if err != nil {
		t.Fatalf("GetToken failed: %v", err)
	}

	if token != "test-access-token" {
		t.Errorf("Expected token 'test-access-token', got '%s'", token)
	}

	// Verify token is cached
	if provider.cachedToken != "test-access-token" {
		t.Errorf("Expected cachedToken 'test-access-token', got '%s'", provider.cachedToken)
	}
}

func TestOAuthTokenProvider_GetToken_Cached(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		resp := models.ProviderAccessTokenResponse{
			AccessToken: "test-access-token",
			ExpiresIn:   28800,
			TokenType:   "Bearer",
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	provider := NewOAuthTokenProvider("client-id", "client-secret")
	provider.oauthClient = NewOAuthClient(WithBaseURL(server.URL))

	// First call - should hit the server
	token1, err := provider.GetToken(context.Background())
	if err != nil {
		t.Fatalf("First GetToken failed: %v", err)
	}

	// Second call - should use cached token
	token2, err := provider.GetToken(context.Background())
	if err != nil {
		t.Fatalf("Second GetToken failed: %v", err)
	}

	if token1 != token2 {
		t.Errorf("Expected same token, got '%s' and '%s'", token1, token2)
	}

	if callCount != 1 {
		t.Errorf("Expected 1 server call, got %d", callCount)
	}
}

func TestOAuthTokenProvider_GetToken_Refresh(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := models.ProviderAccessTokenResponse{
			AccessToken: "new-access-token",
			ExpiresIn:   28800,
			TokenType:   "Bearer",
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	provider := NewOAuthTokenProvider(
		"client-id",
		"client-secret",
		WithTokenRefreshBuffer(1*time.Hour), // Set high buffer to force immediate refresh
	)
	provider.oauthClient = NewOAuthClient(WithBaseURL(server.URL))

	// Set an expired token
	provider.cachedToken = "old-token"
	provider.tokenExpiration = time.Now().Add(-1 * time.Minute) // Expired

	token, err := provider.GetToken(context.Background())
	if err != nil {
		t.Fatalf("GetToken failed: %v", err)
	}

	if token != "new-access-token" {
		t.Errorf("Expected new token 'new-access-token', got '%s'", token)
	}
}

func TestOAuthTokenProvider_GetToken_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		resp := models.StandardErrorResponse{
			Error:            "invalid_client",
			ErrorDescription: "Client authentication failed",
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	provider := NewOAuthTokenProvider("invalid-client", "invalid-secret")
	provider.oauthClient = NewOAuthClient(WithBaseURL(server.URL))

	_, err := provider.GetToken(context.Background())
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

func TestOAuthTokenProvider_GetToken_WithScopes(t *testing.T) {
	var receivedScope string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		receivedScope = r.FormValue("scope")

		resp := models.ProviderAccessTokenResponse{
			AccessToken: "test-access-token",
			ExpiresIn:   28800,
			TokenType:   "Bearer",
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	provider := NewOAuthTokenProvider(
		"client-id",
		"client-secret",
		WithOAuthScopes("addresses tracking labels"),
	)
	provider.oauthClient = NewOAuthClient(WithBaseURL(server.URL))

	_, err := provider.GetToken(context.Background())
	if err != nil {
		t.Fatalf("GetToken failed: %v", err)
	}

	if receivedScope != "addresses tracking labels" {
		t.Errorf("Expected scope 'addresses tracking labels', got '%s'", receivedScope)
	}
}

func TestOAuthTokenProvider_ConcurrentAccess(t *testing.T) {
	callCount := 0
	var mu sync.Mutex

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		callCount++
		mu.Unlock()

		// Add a small delay to increase chance of race conditions
		time.Sleep(10 * time.Millisecond)

		resp := models.ProviderAccessTokenResponse{
			AccessToken: "test-access-token",
			ExpiresIn:   28800,
			TokenType:   "Bearer",
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	provider := NewOAuthTokenProvider("client-id", "client-secret")
	provider.oauthClient = NewOAuthClient(WithBaseURL(server.URL))

	// Launch multiple concurrent requests
	const numGoroutines = 10
	var wg sync.WaitGroup
	wg.Add(numGoroutines)
	errors := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			token, err := provider.GetToken(context.Background())
			if err != nil {
				errors <- err
				return
			}
			if token != "test-access-token" {
				errors <- fmt.Errorf("expected token 'test-access-token', got '%s'", token)
			}
		}()
	}

	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		t.Errorf("Concurrent GetToken failed: %v", err)
	}

	// Should only call the server once due to caching and locking
	mu.Lock()
	defer mu.Unlock()
	if callCount > 2 {
		t.Errorf("Expected at most 2 server calls due to race, got %d", callCount)
	}
}

func TestOAuthTokenProvider_WithRefreshTokens(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++

		if callCount == 1 {
			// First call - client credentials returns refresh token
			resp := models.ProviderTokensResponse{
				AccessToken:  "initial-access-token",
				RefreshToken: "refresh-token",
				ExpiresIn:    28800,
				TokenType:    "Bearer",
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		} else {
			// Second call - refresh token grant
			resp := models.ProviderTokensResponse{
				AccessToken:  "refreshed-access-token",
				RefreshToken: "new-refresh-token",
				ExpiresIn:    28800,
				TokenType:    "Bearer",
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}
	}))
	defer server.Close()

	provider := NewOAuthTokenProvider(
		"client-id",
		"client-secret",
		WithRefreshTokens(true),
	)
	provider.oauthClient = NewOAuthClient(WithBaseURL(server.URL))

	// First call - get initial token
	token1, err := provider.GetToken(context.Background())
	if err != nil {
		t.Fatalf("First GetToken failed: %v", err)
	}

	if token1 != "initial-access-token" {
		t.Errorf("Expected token 'initial-access-token', got '%s'", token1)
	}

	// Manually expire the token to force refresh
	provider.mutex.Lock()
	provider.tokenExpiration = time.Now().Add(-1 * time.Minute)
	provider.mutex.Unlock()

	// Second call - should use refresh token
	token2, err := provider.GetToken(context.Background())
	if err != nil {
		t.Fatalf("Second GetToken failed: %v", err)
	}

	if token2 != "refreshed-access-token" {
		t.Errorf("Expected token 'refreshed-access-token', got '%s'", token2)
	}

	if callCount != 2 {
		t.Errorf("Expected 2 server calls, got %d", callCount)
	}
}

func TestOAuthTokenProvider_RefreshTokenFallback(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++

		switch callCount {
		case 1:
			// First call - client credentials returns refresh token
			resp := models.ProviderTokensResponse{
				AccessToken:  "initial-access-token",
				RefreshToken: "refresh-token",
				ExpiresIn:    28800,
				TokenType:    "Bearer",
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		case 2:
			// Second call - refresh token fails
			w.WriteHeader(http.StatusBadRequest)
			resp := models.StandardErrorResponse{
				Error:            "invalid_grant",
				ErrorDescription: "Refresh token has expired",
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		default:
			// Third call - fallback to client credentials
			resp := models.ProviderAccessTokenResponse{
				AccessToken: "new-access-token",
				ExpiresIn:   28800,
				TokenType:   "Bearer",
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}
	}))
	defer server.Close()

	provider := NewOAuthTokenProvider(
		"client-id",
		"client-secret",
		WithRefreshTokens(true),
	)
	provider.oauthClient = NewOAuthClient(WithBaseURL(server.URL))

	// First call - get initial token with refresh token
	token1, err := provider.GetToken(context.Background())
	if err != nil {
		t.Fatalf("First GetToken failed: %v", err)
	}

	if token1 != "initial-access-token" {
		t.Errorf("Expected token 'initial-access-token', got '%s'", token1)
	}

	// Manually expire the token to force refresh
	provider.mutex.Lock()
	provider.tokenExpiration = time.Now().Add(-1 * time.Minute)
	provider.mutex.Unlock()

	// Second call - refresh token fails, should fallback to client credentials
	token2, err := provider.GetToken(context.Background())
	if err != nil {
		t.Fatalf("Second GetToken failed: %v", err)
	}

	if token2 != "new-access-token" {
		t.Errorf("Expected token 'new-access-token', got '%s'", token2)
	}

	if callCount != 3 {
		t.Errorf("Expected 3 server calls, got %d", callCount)
	}
}

func TestOAuthTokenProvider_TokenExpirationCalculation(t *testing.T) {
	now := time.Now()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := models.ProviderAccessTokenResponse{
			AccessToken: "test-access-token",
			ExpiresIn:   28800, // 8 hours
			TokenType:   "Bearer",
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	refreshBuffer := 10 * time.Minute
	provider := NewOAuthTokenProvider(
		"client-id",
		"client-secret",
		WithTokenRefreshBuffer(refreshBuffer),
	)
	provider.oauthClient = NewOAuthClient(WithBaseURL(server.URL))

	_, err := provider.GetToken(context.Background())
	if err != nil {
		t.Fatalf("GetToken failed: %v", err)
	}

	// Check token expiration is set correctly
	expectedExpiration := now.Add(28800*time.Second - refreshBuffer)
	// Allow 2 second tolerance for test execution time
	if provider.tokenExpiration.Before(expectedExpiration.Add(-2*time.Second)) ||
		provider.tokenExpiration.After(expectedExpiration.Add(2*time.Second)) {
		t.Errorf("Token expiration not set correctly. Expected around %v, got %v",
			expectedExpiration, provider.tokenExpiration)
	}
}

func TestOAuthTokenProvider_TokenExpirationShortLifespan(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := models.ProviderAccessTokenResponse{
			AccessToken: "short-lived-token",
			ExpiresIn:   30,
			TokenType:   "Bearer",
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Fatalf("failed to encode token response: %v", err)
		}
	}))
	defer server.Close()

	provider := NewOAuthTokenProvider(
		"client-id",
		"client-secret",
		WithTokenRefreshBuffer(10*time.Minute),
	)
	provider.oauthClient = NewOAuthClient(WithBaseURL(server.URL))

	// Capture time immediately before GetToken to match calculateExpiration's time.Now() call
	now := time.Now()
	if _, err := provider.GetToken(context.Background()); err != nil {
		t.Fatalf("GetToken failed: %v", err)
	}

	// The token expires in 30 seconds, but the refresh buffer is 10 minutes.
	// The provider should clamp the expiration to a minimum of 1 second in the future.
	// So, expected expiration is now + 1 second.
	// The token expires in 30 seconds, but the refresh buffer is 10 minutes.
	// The provider should clamp the expiration to a minimum of 1 second in the future.
	expiresIn := 30 * time.Second
	refreshBuffer := 10 * time.Minute
	minExpiration := time.Second
	expectedExpiration := now.Add(minExpiration)
	if expiresIn < refreshBuffer {
		expectedExpiration = now.Add(minExpiration)
	} else {
		expectedExpiration = now.Add(expiresIn - refreshBuffer)
	}
)
	if provider.tokenExpiration.Before(expectedExpiration.Add(-2*time.Second)) ||
		provider.tokenExpiration.After(expectedExpiration.Add(2*time.Second)) {
		t.Errorf("Token expiration should be approximately 1 second from GetToken call. Expected around %v, got %v",
			expectedExpiration, provider.tokenExpiration)
	}
}

func TestOAuthTokenProvider_NoRefreshTokenWhenDisabled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return a response with refresh token
		resp := models.ProviderTokensResponse{
			AccessToken:  "access-token",
			RefreshToken: "refresh-token",
			ExpiresIn:    28800,
			TokenType:    "Bearer",
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// Create provider with refresh tokens disabled (default)
	provider := NewOAuthTokenProvider("client-id", "client-secret")
	provider.oauthClient = NewOAuthClient(WithBaseURL(server.URL))

	_, err := provider.GetToken(context.Background())
	if err != nil {
		t.Fatalf("GetToken failed: %v", err)
	}

	// Refresh token should not be stored when disabled
	if provider.refreshToken != "" {
		t.Errorf("Expected no refresh token when disabled, got '%s'", provider.refreshToken)
	}
}

func TestOAuthTokenProvider_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return invalid JSON
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`invalid json`))
	}))
	defer server.Close()

	provider := NewOAuthTokenProvider("client-id", "client-secret")
	provider.oauthClient = NewOAuthClient(WithBaseURL(server.URL))

	_, err := provider.GetToken(context.Background())
	if err == nil {
		t.Fatal("Expected error for invalid JSON, got nil")
	}
}

func TestOAuthTokenProvider_DoubleCheckLocking(t *testing.T) {
	callCount := 0
	var mu sync.Mutex
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		callCount++
		// Simulate slow response to increase chance of concurrent access
		time.Sleep(50 * time.Millisecond)
		mu.Unlock()

		resp := models.ProviderAccessTokenResponse{
			AccessToken: "test-access-token",
			ExpiresIn:   28800,
			TokenType:   "Bearer",
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	provider := NewOAuthTokenProvider("client-id", "client-secret")
	provider.oauthClient = NewOAuthClient(WithBaseURL(server.URL))

	// Set token to expired to force refresh
	provider.cachedToken = "old-token"
	provider.tokenExpiration = time.Now().Add(-1 * time.Minute)

	// Launch concurrent refresh attempts
	const numGoroutines = 5
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			_, _ = provider.GetToken(context.Background())
		}()
	}

	wg.Wait()

	// Due to double-check locking, should only call server once
	mu.Lock()
	defer mu.Unlock()
	if callCount > 2 {
		t.Errorf("Expected at most 2 server calls due to double-check locking, got %d", callCount)
	}
}
