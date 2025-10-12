package usps

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/my-eq/go-usps/models"
)

const (
	// OAuthProductionBaseURL is the base URL for the USPS OAuth production API
	OAuthProductionBaseURL = "https://apis.usps.com/oauth2/v3"
	// OAuthTestingBaseURL is the base URL for the USPS OAuth testing API
	OAuthTestingBaseURL = "https://apis-tem.usps.com/oauth2/v3"
)

// OAuthClient is the USPS OAuth API client
type OAuthClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewOAuthClient creates a new USPS OAuth API client
func NewOAuthClient(opts ...Option) *OAuthClient {
	c := &OAuthClient{
		baseURL:    OAuthProductionBaseURL,
		httpClient: &http.Client{Timeout: DefaultTimeout},
	}

	// Apply options using a temporary regular client
	tempClient := &Client{
		baseURL:    c.baseURL,
		httpClient: c.httpClient,
	}
	for _, opt := range opts {
		opt(tempClient)
	}
	c.baseURL = tempClient.baseURL
	c.httpClient = tempClient.httpClient

	return c
}

// NewOAuthTestClient creates a new USPS OAuth API client configured for the testing environment
func NewOAuthTestClient(opts ...Option) *OAuthClient {
	opts = append([]Option{WithBaseURL(OAuthTestingBaseURL)}, opts...)
	return NewOAuthClient(opts...)
}

// PostToken generates OAuth tokens based on the grant type
func (c *OAuthClient) PostToken(ctx context.Context, req interface{}) (interface{}, error) {
	var contentType string
	var body io.Reader

	// Determine content type and encode body
	switch r := req.(type) {
	case *models.ClientCredentials:
		contentType = "application/x-www-form-urlencoded"
		values := url.Values{}
		values.Set("grant_type", r.GrantType)
		values.Set("client_id", r.ClientID)
		values.Set("client_secret", r.ClientSecret)
		if r.Scope != "" {
			values.Set("scope", r.Scope)
		}
		body = strings.NewReader(values.Encode())
	case *models.RefreshTokenCredentials:
		contentType = "application/json"
		jsonData, err := json.Marshal(r)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request: %w", err)
		}
		body = bytes.NewReader(jsonData)
	case *models.AuthorizationCodeCredentials:
		contentType = "application/json"
		jsonData, err := json.Marshal(r)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request: %w", err)
		}
		body = bytes.NewReader(jsonData)
	default:
		return nil, fmt.Errorf("unsupported request type")
	}

	// Create request
	fullURL := c.baseURL + "/token"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, fullURL, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", contentType)
	httpReq.Header.Set("Accept", "application/json")

	// Execute request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Handle error responses
	if resp.StatusCode >= 400 {
		var errResp models.StandardErrorResponse
		if err := json.Unmarshal(respBody, &errResp); err != nil {
			return nil, fmt.Errorf("OAuth error (status %d): %s", resp.StatusCode, string(respBody))
		}
		return nil, &OAuthError{
			StatusCode:   resp.StatusCode,
			ErrorMessage: errResp,
		}
	}

	// Try to unmarshal as ProviderTokensResponse first (has refresh_token)
	var tokensResp models.ProviderTokensResponse
	if err := json.Unmarshal(respBody, &tokensResp); err == nil && tokensResp.RefreshToken != "" {
		return &tokensResp, nil
	}

	// Otherwise unmarshal as ProviderAccessTokenResponse
	var accessTokenResp models.ProviderAccessTokenResponse
	if err := json.Unmarshal(respBody, &accessTokenResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &accessTokenResp, nil
}

// PostRevoke revokes an OAuth token
func (c *OAuthClient) PostRevoke(ctx context.Context, clientID, clientSecret string, req *models.TokenRevokeRequest) error {
	// Encode request body
	values := url.Values{}
	values.Set("token", req.Token)
	if req.TokenTypeHint != "" {
		values.Set("token_type_hint", req.TokenTypeHint)
	}

	// Create request
	fullURL := c.baseURL + "/revoke"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, fullURL, strings.NewReader(values.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set Basic Authentication
	auth := base64.StdEncoding.EncodeToString([]byte(clientID + ":" + clientSecret))
	httpReq.Header.Set("Authorization", "Basic "+auth)
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	httpReq.Header.Set("Accept", "application/json")

	// Execute request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Handle error responses
	if resp.StatusCode >= 400 {
		// Read response body
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("OAuth error (status %d): failed to read response", resp.StatusCode)
		}

		var errResp models.StandardErrorResponse
		if err := json.Unmarshal(respBody, &errResp); err != nil {
			return fmt.Errorf("OAuth error (status %d): %s", resp.StatusCode, string(respBody))
		}
		return &OAuthError{
			StatusCode:   resp.StatusCode,
			ErrorMessage: errResp,
		}
	}

	return nil
}

// OAuthError represents an error returned by the USPS OAuth API
type OAuthError struct {
	StatusCode   int
	ErrorMessage models.StandardErrorResponse
}

// Error implements the error interface
func (e *OAuthError) Error() string {
	if e.ErrorMessage.Error != "" {
		msg := e.ErrorMessage.Error
		if e.ErrorMessage.ErrorDescription != "" {
			msg += ": " + e.ErrorMessage.ErrorDescription
		}
		return fmt.Sprintf("OAuth error (status %d): %s", e.StatusCode, msg)
	}
	return fmt.Sprintf("OAuth error (status %d)", e.StatusCode)
}
