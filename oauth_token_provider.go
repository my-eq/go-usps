package usps

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/my-eq/go-usps/models"
)

const (
	// DefaultTokenRefreshBuffer is the default time before token expiration to refresh.
	// Tokens are refreshed 5 minutes before they expire by default.
	DefaultTokenRefreshBuffer = 5 * time.Minute
)

// OAuthTokenProvider is a TokenProvider that automatically manages OAuth 2.0 tokens.
// It handles token acquisition, caching, and automatic refresh before expiration.
// This provider is thread-safe and suitable for concurrent use in production environments.
type OAuthTokenProvider struct {
	clientID         string
	clientSecret     string
	scopes           string
	refreshBuffer    time.Duration
	oauthClient      *OAuthClient
	mutex            sync.RWMutex
	cachedToken      string
	tokenExpiration  time.Time
	refreshToken     string
	useRefreshTokens bool
}

// OAuthTokenOption is a functional option for configuring OAuthTokenProvider.
type OAuthTokenOption func(*OAuthTokenProvider)

// WithOAuthScopes sets the OAuth scopes for token requests.
// Multiple scopes should be space-separated (e.g., "addresses tracking labels").
func WithOAuthScopes(scopes string) OAuthTokenOption {
	return func(p *OAuthTokenProvider) {
		p.scopes = scopes
	}
}

// WithTokenRefreshBuffer sets how early before expiration to refresh the token.
// Default is DefaultTokenRefreshBuffer (5 minutes) before the token expires.
// This ensures tokens are refreshed proactively before they expire.
func WithTokenRefreshBuffer(duration time.Duration) OAuthTokenOption {
	return func(p *OAuthTokenProvider) {
		p.refreshBuffer = duration
	}
}

// WithOAuthEnvironment configures the OAuth environment.
// Use "production" (default) or "testing" to set the OAuth base URL.
func WithOAuthEnvironment(env string) OAuthTokenOption {
	return func(p *OAuthTokenProvider) {
		if env == "testing" {
			p.oauthClient = NewOAuthTestClient()
		} else {
			p.oauthClient = NewOAuthClient()
		}
	}
}

// WithRefreshTokens enables the use of refresh tokens when available.
// When enabled, the provider will use refresh tokens to obtain new access tokens
// instead of always using client credentials. This is more efficient and allows
// for longer-lived sessions.
// Default is false (always use client credentials).
func WithRefreshTokens(enabled bool) OAuthTokenOption {
	return func(p *OAuthTokenProvider) {
		p.useRefreshTokens = enabled
	}
}

// NewOAuthTokenProvider creates a new OAuthTokenProvider that automatically manages
// OAuth 2.0 tokens using the client credentials flow.
//
// The provider handles:
//   - Initial token acquisition
//   - Token caching
//   - Automatic refresh before expiration (default: 5 minutes before expiry)
//   - Thread-safe concurrent access
//
// Access tokens from USPS are valid for 8 hours. The provider will automatically
// refresh the token 5 minutes before expiration (configurable via WithTokenRefreshBuffer).
//
// Example:
//
//	provider := usps.NewOAuthTokenProvider("client-id", "client-secret")
//	client := usps.NewClient(provider)
//
// Example with options:
//
//	provider := usps.NewOAuthTokenProvider(
//	    "client-id",
//	    "client-secret",
//	    usps.WithOAuthScopes("addresses tracking"),
//	    usps.WithTokenRefreshBuffer(10 * time.Minute),
//	    usps.WithOAuthEnvironment("testing"),
//	)
func NewOAuthTokenProvider(clientID, clientSecret string, opts ...OAuthTokenOption) *OAuthTokenProvider {
	p := &OAuthTokenProvider{
		clientID:      clientID,
		clientSecret:  clientSecret,
		refreshBuffer: DefaultTokenRefreshBuffer,
		oauthClient:   NewOAuthClient(),
	}

	for _, opt := range opts {
		opt(p)
	}

	return p
}

// NewOAuthTestTokenProvider creates a new OAuthTokenProvider configured for the testing environment.
// This is equivalent to calling NewOAuthTokenProvider with WithOAuthEnvironment("testing").
//
// Example:
//
//	provider := usps.NewOAuthTestTokenProvider("test-client-id", "test-client-secret")
//	client := usps.NewTestClient(provider)
func NewOAuthTestTokenProvider(clientID, clientSecret string, opts ...OAuthTokenOption) *OAuthTokenProvider {
	opts = append([]OAuthTokenOption{WithOAuthEnvironment("testing")}, opts...)
	return NewOAuthTokenProvider(clientID, clientSecret, opts...)
}

// GetToken returns a valid OAuth token, refreshing it if necessary.
// This method is thread-safe and implements the TokenProvider interface.
func (p *OAuthTokenProvider) GetToken(ctx context.Context) (string, error) {
	// Check if we have a valid cached token
	p.mutex.RLock()
	if p.cachedToken != "" && time.Now().Before(p.tokenExpiration) {
		token := p.cachedToken
		p.mutex.RUnlock()
		return token, nil
	}
	refreshToken := p.refreshToken
	useRefresh := p.useRefreshTokens && refreshToken != ""
	p.mutex.RUnlock()

	// Need to acquire or refresh token
	p.mutex.Lock()
	defer p.mutex.Unlock()

	// Double-check after acquiring write lock (another goroutine may have refreshed)
	if p.cachedToken != "" && time.Now().Before(p.tokenExpiration) {
		return p.cachedToken, nil
	}

	// Refresh token if we have one and refresh tokens are enabled
	if useRefresh {
		if err := p.refreshTokenLocked(ctx); err != nil {
			// If refresh fails, fall back to client credentials
			if err := p.acquireTokenLocked(ctx); err != nil {
				return "", err
			}
		}
	} else {
		// Acquire new token using client credentials
		if err := p.acquireTokenLocked(ctx); err != nil {
			return "", err
		}
	}

	return p.cachedToken, nil
}

// calculateExpiration calculates the token expiration time with the configured refresh buffer.
func (p *OAuthTokenProvider) calculateExpiration(expiresIn int) time.Time {
	if expiresIn <= 0 {
		// If the server does not provide a valid expiration, force an immediate refresh.
		return time.Now()
	}

	expiresInDuration := time.Duration(expiresIn) * time.Second

	buffer := p.refreshBuffer
	if buffer >= expiresInDuration {
		// Refresh at least one second before expiration when the buffer exceeds the token lifetime.
		buffer = expiresInDuration - time.Second
		if buffer < 0 {
			buffer = 0
		}
	}

	return time.Now().Add(expiresInDuration - buffer)
}

// acquireTokenLocked acquires a new token using client credentials.
// Caller must hold the write lock.
func (p *OAuthTokenProvider) acquireTokenLocked(ctx context.Context) error {
	req := &models.ClientCredentials{
		GrantType:    "client_credentials",
		ClientID:     p.clientID,
		ClientSecret: p.clientSecret,
		Scope:        p.scopes,
	}

	result, err := p.oauthClient.PostToken(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to acquire OAuth token: %w", err)
	}

	// Handle both response types
	switch resp := result.(type) {
	case *models.ProviderAccessTokenResponse:
		p.cachedToken = resp.AccessToken
		p.tokenExpiration = p.calculateExpiration(resp.ExpiresIn)
		// Clear refresh token since client credentials don't return one
		p.refreshToken = ""
	case *models.ProviderTokensResponse:
		p.cachedToken = resp.AccessToken
		p.tokenExpiration = p.calculateExpiration(resp.ExpiresIn)
		// Store refresh token if refresh tokens are enabled
		if p.useRefreshTokens {
			p.refreshToken = resp.RefreshToken
		}
	default:
		return fmt.Errorf("unexpected token response type: %T", result)
	}

	return nil
}

// refreshTokenLocked refreshes the token using a refresh token.
// Caller must hold the write lock.
func (p *OAuthTokenProvider) refreshTokenLocked(ctx context.Context) error {
	if p.refreshToken == "" {
		return fmt.Errorf("no refresh token available")
	}

	req := &models.RefreshTokenCredentials{
		GrantType:    "refresh_token",
		ClientID:     p.clientID,
		ClientSecret: p.clientSecret,
		RefreshToken: p.refreshToken,
		Scope:        p.scopes,
	}

	result, err := p.oauthClient.PostToken(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to refresh OAuth token: %w", err)
	}

	// Refresh token always returns ProviderTokensResponse
	tokensResp, ok := result.(*models.ProviderTokensResponse)
	if !ok {
		return fmt.Errorf("unexpected token response type: %T", result)
	}

	p.cachedToken = tokensResp.AccessToken
	p.tokenExpiration = p.calculateExpiration(tokensResp.ExpiresIn)
	// Update refresh token
	p.refreshToken = tokensResp.RefreshToken

	return nil
}
