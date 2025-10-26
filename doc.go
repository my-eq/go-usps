// Package usps provides a lightweight, production-grade Go client library
// for the USPS Addresses 3.0 REST API and OAuth 2.0 API.
//
// The package implements all three USPS Addresses 3.0 API endpoints:
//   - Address Standardization (GetAddress)
//   - City/State Lookup (GetCityState)
//   - ZIP Code Lookup (GetZIPCode)
//
// And the USPS OAuth 2.0 API endpoints:
//   - Token Generation (PostToken) - supports Client Credentials, Refresh Token, and Authorization Code grants
//   - Token Revocation (PostRevoke)
//
// # Quick Start
//
// The easiest way to use the library is with automatic OAuth token management:
//
//	tokenProvider := usps.NewOAuthTokenProvider("client-id", "client-secret")
//	client := usps.NewClient(tokenProvider)
//
// Standardize an address:
//
//	req := &models.AddressRequest{
//	    StreetAddress: "123 Main St",
//	    City:          "New York",
//	    State:         "NY",
//	}
//	resp, err := client.GetAddress(context.Background(), req)
//
// Look up city and state by ZIP code:
//
//	req := &models.CityStateRequest{ZIPCode: "10001"}
//	resp, err := client.GetCityState(context.Background(), req)
//
// Look up ZIP code by address:
//
//	req := &models.ZIPCodeRequest{
//	    StreetAddress: "123 Main St",
//	    City:          "New York",
//	    State:         "NY",
//	}
//	resp, err := client.GetZIPCode(context.Background(), req)
//
// # OAuth Authentication
//
// The library provides automatic OAuth token management via OAuthTokenProvider (recommended):
//
//	tokenProvider := usps.NewOAuthTokenProvider(
//	    "your-client-id",
//	    "your-client-secret",
//	    usps.WithOAuthScopes("addresses tracking"),
//	    usps.WithTokenRefreshBuffer(10 * time.Minute),
//	)
//	client := usps.NewClient(tokenProvider)
//
// The OAuthTokenProvider automatically:
//   - Acquires tokens using client credentials flow
//   - Caches tokens to minimize API calls
//   - Refreshes tokens before expiration (default: 5 minutes before)
//   - Handles concurrent access safely
//
// For manual token management, use OAuthClient:
//
//	oauthClient := usps.NewOAuthClient()
//	req := &models.ClientCredentials{
//	    GrantType:    "client_credentials",
//	    ClientID:     "your-client-id",
//	    ClientSecret: "your-client-secret",
//	    Scope:        "addresses tracking labels",
//	}
//	result, err := oauthClient.PostToken(context.Background(), req)
//
// Access tokens expire after 8 hours. Refresh tokens can be used to obtain new access tokens:
//
//	req := &models.RefreshTokenCredentials{
//	    GrantType:    "refresh_token",
//	    ClientID:     "your-client-id",
//	    ClientSecret: "your-client-secret",
//	    RefreshToken: "your-refresh-token",
//	}
//	result, err := oauthClient.PostToken(context.Background(), req)
//
// Revoke a refresh token when no longer needed:
//
//	req := &models.TokenRevokeRequest{
//	    Token:         "refresh-token-to-revoke",
//	    TokenTypeHint: "refresh_token",
//	}
//	err := oauthClient.PostRevoke(context.Background(), "client-id", "client-secret", req)
//
// For static tokens (not recommended for production), use StaticTokenProvider:
//
//	tokenProvider := usps.NewStaticTokenProvider("your-oauth-token")
//	client := usps.NewClient(tokenProvider)
//
// # Configuration
//
// The client can be configured with various options:
//
//	client := usps.NewClient(
//	    tokenProvider,
//	    usps.WithTimeout(60 * time.Second),
//	    usps.WithBaseURL("https://custom.url.com"),
//	)
//
// For testing, use the test environment:
//
//	client := usps.NewTestClient(tokenProvider)
//	oauthClient := usps.NewOAuthTestClient()
//
// # Error Handling
//
// API errors are returned as *APIError with detailed error information:
//
//	resp, err := client.GetAddress(ctx, req)
//	if err != nil {
//	    if apiErr, ok := err.(*usps.APIError); ok {
//	        fmt.Printf("API Error: %s\n", apiErr.ErrorMessage.Error.Message)
//	    }
//	    return
//	}
//
// OAuth errors are returned as *OAuthError:
//
//	result, err := oauthClient.PostToken(ctx, req)
//	if err != nil {
//	    if oauthErr, ok := err.(*usps.OAuthError); ok {
//	        fmt.Printf("OAuth Error: %s\n", oauthErr.ErrorMessage.Error)
//	    }
//	    return
//	}
//
// For more information, see:
//   - Addresses API: https://developers.usps.com/addressesv3
//   - OAuth API: https://developers.usps.com/oauth2v3
package usps
