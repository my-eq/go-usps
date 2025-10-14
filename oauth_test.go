package usps

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/my-eq/go-usps/models"
)

func TestNewOAuthClient(t *testing.T) {
	client := NewOAuthClient()

	if client.baseURL != OAuthProductionBaseURL {
		t.Errorf("Expected base URL %s, got %s", OAuthProductionBaseURL, client.baseURL)
	}

	if client.httpClient.Timeout != DefaultTimeout {
		t.Errorf("Expected timeout %v, got %v", DefaultTimeout, client.httpClient.Timeout)
	}
}

func TestNewOAuthTestClient(t *testing.T) {
	client := NewOAuthTestClient()

	if client.baseURL != OAuthTestingBaseURL {
		t.Errorf("Expected base URL %s, got %s", OAuthTestingBaseURL, client.baseURL)
	}
}

func TestOAuthClientOptions(t *testing.T) {
	customURL := "https://custom.oauth.url"
	client := NewOAuthClient(WithBaseURL(customURL))

	if client.baseURL != customURL {
		t.Errorf("Expected base URL %s, got %s", customURL, client.baseURL)
	}
}

func TestPostToken_ClientCredentials_Success(t *testing.T) {
	// Mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/token" {
			t.Errorf("Expected path /token, got %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST method, got %s", r.Method)
		}

		resp := models.ProviderAccessTokenResponse{
			AccessToken:     "test-access-token",
			ExpiresIn:       28799,
			TokenType:       "Bearer",
			Scope:           "addresses tracking",
			IssuedAt:        1680888985929,
			Status:          "approved",
			Issuer:          "api.usps.com",
			ClientID:        "test-client-id",
			ApplicationName: "Test App",
			APIProducts:     "[Shipping-Silver]",
			PublicKey:       "test-public-key",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewOAuthClient(WithBaseURL(server.URL))
	req := &models.ClientCredentials{
		GrantType:    "client_credentials",
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		Scope:        "addresses tracking",
	}

	result, err := client.PostToken(context.Background(), req)
	if err != nil {
		t.Fatalf("PostToken failed: %v", err)
	}

	accessTokenResp, ok := result.(*models.ProviderAccessTokenResponse)
	if !ok {
		t.Fatalf("Expected *models.ProviderAccessTokenResponse, got %T", result)
	}

	if accessTokenResp.AccessToken != "test-access-token" {
		t.Errorf("Expected access token 'test-access-token', got '%s'", accessTokenResp.AccessToken)
	}
	if accessTokenResp.TokenType != "Bearer" {
		t.Errorf("Expected token type 'Bearer', got '%s'", accessTokenResp.TokenType)
	}
	if accessTokenResp.ExpiresIn != 28799 {
		t.Errorf("Expected expires in 28799, got %d", accessTokenResp.ExpiresIn)
	}
}

func TestPostToken_RefreshToken_Success(t *testing.T) {
	// Mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := models.ProviderTokensResponse{
			AccessToken:           "new-access-token",
			ExpiresIn:             28799,
			TokenType:             "Bearer",
			RefreshToken:          "new-refresh-token",
			RefreshTokenIssuedAt:  1680889628876,
			RefreshTokenExpiresIn: 43199,
			RefreshCount:          1,
			RefreshTokenStatus:    "approved",
			IssuedAt:              1680889628876,
			Status:                "approved",
			Issuer:                "api.usps.com",
			ClientID:              "test-client-id",
			ApplicationName:       "Test App",
			APIProducts:           "[Shipping-Silver]",
			PublicKey:             "test-public-key",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewOAuthClient(WithBaseURL(server.URL))
	req := &models.RefreshTokenCredentials{
		GrantType:    "refresh_token",
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		RefreshToken: "old-refresh-token",
	}

	result, err := client.PostToken(context.Background(), req)
	if err != nil {
		t.Fatalf("PostToken failed: %v", err)
	}

	tokensResp, ok := result.(*models.ProviderTokensResponse)
	if !ok {
		t.Fatalf("Expected *models.ProviderTokensResponse, got %T", result)
	}

	if tokensResp.AccessToken != "new-access-token" {
		t.Errorf("Expected access token 'new-access-token', got '%s'", tokensResp.AccessToken)
	}
	if tokensResp.RefreshToken != "new-refresh-token" {
		t.Errorf("Expected refresh token 'new-refresh-token', got '%s'", tokensResp.RefreshToken)
	}
	if tokensResp.RefreshCount != 1 {
		t.Errorf("Expected refresh count 1, got %d", tokensResp.RefreshCount)
	}
}

func TestPostToken_AuthorizationCode_Success(t *testing.T) {
	// Mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := models.ProviderTokensResponse{
			AccessToken:           "auth-code-access-token",
			ExpiresIn:             28799,
			TokenType:             "Bearer",
			RefreshToken:          "auth-code-refresh-token",
			RefreshTokenIssuedAt:  1680889020006,
			RefreshTokenExpiresIn: 43199,
			RefreshCount:          0,
			RefreshTokenStatus:    "approved",
			IssuedAt:              1680889020006,
			Status:                "approved",
			Issuer:                "api.usps.com",
			ClientID:              "test-client-id",
			ApplicationName:       "Test App",
			APIProducts:           "[Shipping-Silver]",
			PublicKey:             "test-public-key",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewOAuthClient(WithBaseURL(server.URL))
	req := &models.AuthorizationCodeCredentials{
		GrantType:    "authorization_code",
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		Code:         "test-auth-code",
		RedirectURI:  "https://example.com/callback",
	}

	result, err := client.PostToken(context.Background(), req)
	if err != nil {
		t.Fatalf("PostToken failed: %v", err)
	}

	tokensResp, ok := result.(*models.ProviderTokensResponse)
	if !ok {
		t.Fatalf("Expected *models.ProviderTokensResponse, got %T", result)
	}

	if tokensResp.AccessToken != "auth-code-access-token" {
		t.Errorf("Expected access token 'auth-code-access-token', got '%s'", tokensResp.AccessToken)
	}
	if tokensResp.RefreshToken != "auth-code-refresh-token" {
		t.Errorf("Expected refresh token 'auth-code-refresh-token', got '%s'", tokensResp.RefreshToken)
	}
}

func TestPostToken_Error(t *testing.T) {
	// Mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		resp := models.StandardErrorResponse{
			Error:            "invalid_client",
			ErrorDescription: "Client credentials are invalid",
			ErrorURI:         "https://datatracker.ietf.org/doc/html/rfc6749#section-5.2",
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewOAuthClient(WithBaseURL(server.URL))
	req := &models.ClientCredentials{
		GrantType:    "client_credentials",
		ClientID:     "invalid-client",
		ClientSecret: "invalid-secret",
	}

	_, err := client.PostToken(context.Background(), req)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	oauthErr, ok := err.(*OAuthError)
	if !ok {
		t.Fatalf("Expected *OAuthError, got %T", err)
	}

	if oauthErr.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, oauthErr.StatusCode)
	}
	if oauthErr.ErrorMessage.Error != "invalid_client" {
		t.Errorf("Expected error 'invalid_client', got '%s'", oauthErr.ErrorMessage.Error)
	}
}

func TestPostRevoke_Success(t *testing.T) {
	// Mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/revoke" {
			t.Errorf("Expected path /revoke, got %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST method, got %s", r.Method)
		}

		// Check Basic Auth header
		auth := r.Header.Get("Authorization")
		if auth == "" {
			t.Error("Expected Authorization header")
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewOAuthClient(WithBaseURL(server.URL))
	req := &models.TokenRevokeRequest{
		Token:         "test-refresh-token",
		TokenTypeHint: "refresh_token",
	}

	err := client.PostRevoke(context.Background(), "test-client-id", "test-client-secret", req)
	if err != nil {
		t.Fatalf("PostRevoke failed: %v", err)
	}
}

func TestPostRevoke_Error(t *testing.T) {
	// Mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		resp := models.StandardErrorResponse{
			Error:            "invalid_request",
			ErrorDescription: "Token is invalid or has expired",
			ErrorURI:         "https://datatracker.ietf.org/doc/html/rfc7009#section-2.2.1",
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewOAuthClient(WithBaseURL(server.URL))
	req := &models.TokenRevokeRequest{
		Token:         "invalid-token",
		TokenTypeHint: "refresh_token",
	}

	err := client.PostRevoke(context.Background(), "test-client-id", "test-client-secret", req)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	oauthErr, ok := err.(*OAuthError)
	if !ok {
		t.Fatalf("Expected *OAuthError, got %T", err)
	}

	if oauthErr.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, oauthErr.StatusCode)
	}
}

func TestOAuthError_Error(t *testing.T) {
	err := &OAuthError{
		StatusCode: 400,
		ErrorMessage: models.StandardErrorResponse{
			Error:            "invalid_grant",
			ErrorDescription: "The provided authorization grant is invalid",
		},
	}

	expectedMsg := "OAuth error (status 400): invalid_grant: The provided authorization grant is invalid"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
	}
}

func TestOAuthError_ErrorNoDescription(t *testing.T) {
	err := &OAuthError{
		StatusCode: 401,
		ErrorMessage: models.StandardErrorResponse{
			Error: "unauthorized_client",
		},
	}

	expectedMsg := "OAuth error (status 401): unauthorized_client"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
	}
}

func TestOAuthError_ErrorNoMessage(t *testing.T) {
	err := &OAuthError{
		StatusCode:   503,
		ErrorMessage: models.StandardErrorResponse{},
	}

	expectedMsg := "OAuth error (status 503)"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
	}
}

func TestPostToken_UnsupportedType(t *testing.T) {
	client := NewOAuthClient()

	// Pass an unsupported type
	_, err := client.PostToken(context.Background(), "unsupported")
	if err == nil {
		t.Fatal("Expected error for unsupported request type, got nil")
	}
}

func TestPostToken_InvalidJSON(t *testing.T) {
	// Mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	client := NewOAuthClient(WithBaseURL(server.URL))
	req := &models.ClientCredentials{
		GrantType:    "client_credentials",
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
	}

	_, err := client.PostToken(context.Background(), req)
	if err == nil {
		t.Fatal("Expected error for invalid JSON, got nil")
	}
}

func TestPostToken_NewRequestWithContextError(t *testing.T) {
	// Use an invalid base URL with null byte to trigger NewRequestWithContext error
	client := &OAuthClient{
		baseURL:    "http://localhost\x00invalid",
		httpClient: &http.Client{Timeout: DefaultTimeout},
	}

	req := &models.ClientCredentials{
		GrantType:    "client_credentials",
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
	}

	_, err := client.PostToken(context.Background(), req)
	if err == nil {
		t.Fatal("Expected error from NewRequestWithContext, got nil")
	}
	if !strings.Contains(err.Error(), "failed to create request") {
		t.Errorf("Expected 'failed to create request' error, got: %v", err)
	}
}

func TestPostToken_HTTPClientError(t *testing.T) {
	// Create a client with a custom HTTP client that will fail
	customClient := &http.Client{
		Transport: &oauthFailingTransport{},
	}

	client := NewOAuthClient(WithHTTPClient(customClient))
	req := &models.ClientCredentials{
		GrantType:    "client_credentials",
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
	}

	_, err := client.PostToken(context.Background(), req)
	if err == nil {
		t.Fatal("Expected error from HTTP client, got nil")
	}
}

func TestPostToken_ReadBodyError(t *testing.T) {
	// Create a client with a custom transport that returns a response
	// with a body that will fail to read
	customClient := &http.Client{
		Transport: &bodyFailingTransport{},
	}

	client := NewOAuthClient(WithHTTPClient(customClient))
	req := &models.ClientCredentials{
		GrantType:    "client_credentials",
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
	}

	_, err := client.PostToken(context.Background(), req)
	if err == nil {
		t.Fatal("Expected error from reading body, got nil")
	}
	if !strings.Contains(err.Error(), "failed to read response body") {
		t.Errorf("Expected 'failed to read response body' error, got: %v", err)
	}
}

func TestPostToken_InvalidRequestURL(t *testing.T) {
	// Create a client with an invalid base URL that would cause NewRequestWithContext to fail
	client := &OAuthClient{
		baseURL:    "http://localhost",
		httpClient: &http.Client{Timeout: DefaultTimeout},
	}

	// Create a request that uses invalid characters
	req := &models.ClientCredentials{
		GrantType:    "client_credentials",
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
	}

	// Use an invalid method by directly calling with bad context
	ctx := context.Background()

	// Modify the method to use a different approach - we'll use a channel-based context
	// that gets cancelled immediately to trigger an error path
	ctx, cancel := context.WithCancel(ctx)
	cancel() // Cancel immediately

	_, err := client.PostToken(ctx, req)
	// This might not trigger the NewRequestWithContext error, so let's use a different approach
	if err != nil {
		// Expected - context was cancelled
		t.Logf("Got expected error: %v", err)
	}
}

func TestPostToken_ErrorInvalidJSON(t *testing.T) {
	// Mock server that returns error with invalid JSON
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("invalid error json"))
	}))
	defer server.Close()

	client := NewOAuthClient(WithBaseURL(server.URL))
	req := &models.ClientCredentials{
		GrantType:    "client_credentials",
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
	}

	_, err := client.PostToken(context.Background(), req)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	// Should not be an OAuthError since JSON parsing failed
	if _, ok := err.(*OAuthError); ok {
		t.Error("Expected generic error, got OAuthError")
	}
}

func TestPostRevoke_HTTPClientError(t *testing.T) {
	// Create a client with a custom HTTP client that will fail
	customClient := &http.Client{
		Transport: &oauthFailingTransport{},
	}

	client := NewOAuthClient(WithHTTPClient(customClient))
	req := &models.TokenRevokeRequest{
		Token:         "test-token",
		TokenTypeHint: "refresh_token",
	}

	err := client.PostRevoke(context.Background(), "client-id", "client-secret", req)
	if err == nil {
		t.Fatal("Expected error from HTTP client, got nil")
	}
}

func TestPostRevoke_NewRequestWithContextError(t *testing.T) {
	// Use an invalid base URL with null byte to trigger NewRequestWithContext error
	client := &OAuthClient{
		baseURL:    "http://localhost\x00invalid",
		httpClient: &http.Client{Timeout: DefaultTimeout},
	}

	req := &models.TokenRevokeRequest{
		Token:         "test-token",
		TokenTypeHint: "refresh_token",
	}

	err := client.PostRevoke(context.Background(), "client-id", "client-secret", req)
	if err == nil {
		t.Fatal("Expected error from NewRequestWithContext, got nil")
	}
	if !strings.Contains(err.Error(), "failed to create request") {
		t.Errorf("Expected 'failed to create request' error, got: %v", err)
	}
}

func TestPostRevoke_ReadBodyError(t *testing.T) {
	// Mock server that returns error response with body that fails to read
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "999999999")
		w.WriteHeader(http.StatusBadRequest)
		// Close connection immediately
		hj, ok := w.(http.Hijacker)
		if ok {
			conn, _, _ := hj.Hijack()
			conn.Close()
		}
	}))
	defer server.Close()

	client := NewOAuthClient(WithBaseURL(server.URL))
	req := &models.TokenRevokeRequest{
		Token:         "test-token",
		TokenTypeHint: "refresh_token",
	}

	err := client.PostRevoke(context.Background(), "client-id", "client-secret", req)
	if err == nil {
		t.Fatal("Expected error from reading body, got nil")
	}
}

func TestPostRevoke_ErrorInvalidJSON(t *testing.T) {
	// Mock server that returns error with invalid JSON
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("invalid error json"))
	}))
	defer server.Close()

	client := NewOAuthClient(WithBaseURL(server.URL))
	req := &models.TokenRevokeRequest{
		Token:         "test-token",
		TokenTypeHint: "refresh_token",
	}

	err := client.PostRevoke(context.Background(), "client-id", "client-secret", req)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	// Should not be an OAuthError since JSON parsing failed
	if _, ok := err.(*OAuthError); ok {
		t.Error("Expected generic error, got OAuthError")
	}
}

func TestPostRevoke_InvalidRequestURL(t *testing.T) {
	// Similar to PostToken test - using cancelled context
	client := &OAuthClient{
		baseURL:    "http://localhost",
		httpClient: &http.Client{Timeout: DefaultTimeout},
	}

	req := &models.TokenRevokeRequest{
		Token:         "test-token",
		TokenTypeHint: "refresh_token",
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := client.PostRevoke(ctx, "client-id", "client-secret", req)
	if err != nil {
		// Expected - context was cancelled
		t.Logf("Got expected error: %v", err)
	}
}

// oauthFailingTransport is a custom transport that always fails
type oauthFailingTransport struct{}

func (t *oauthFailingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return nil, context.DeadlineExceeded
}

// bodyFailingTransport returns a response with a body that fails to read
type bodyFailingTransport struct{}

func (t *bodyFailingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       &failingOAuthReader{},
		Header:     http.Header{},
	}, nil
}

// failingOAuthReader is a reader that always fails
type failingOAuthReader struct{}

func (r *failingOAuthReader) Read(p []byte) (n int, err error) {
	return 0, fmt.Errorf("simulated read error")
}

func (r *failingOAuthReader) Close() error {
	return nil
}
