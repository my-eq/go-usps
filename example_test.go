package usps_test

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/my-eq/go-usps"
	"github.com/my-eq/go-usps/models"
)

func ExampleClient_GetAddress() {
	// Create a token provider with your OAuth token
	tokenProvider := usps.NewStaticTokenProvider("your-oauth-token")

	// Create a new client
	client := usps.NewClient(tokenProvider)

	// Prepare the address request
	req := &models.AddressRequest{
		StreetAddress: "123 Main St",
		City:          "New York",
		State:         "NY",
	}

	// Get the standardized address
	resp, err := client.GetAddress(context.Background(), req)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	fmt.Printf("Standardized Address: %s, %s, %s %s\n",
		resp.Address.StreetAddress,
		resp.Address.City,
		resp.Address.State,
		resp.Address.ZIPCode)
}

func ExampleClient_GetCityState() {
	// Create a token provider with your OAuth token
	tokenProvider := usps.NewStaticTokenProvider("your-oauth-token")

	// Create a new client
	client := usps.NewClient(tokenProvider)

	// Prepare the city-state request
	req := &models.CityStateRequest{
		ZIPCode: "10001",
	}

	// Get the city and state
	resp, err := client.GetCityState(context.Background(), req)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	fmt.Printf("City: %s, State: %s\n", resp.City, resp.State)
}

func ExampleClient_GetZIPCode() {
	// Create a token provider with your OAuth token
	tokenProvider := usps.NewStaticTokenProvider("your-oauth-token")

	// Create a new client
	client := usps.NewClient(tokenProvider)

	// Prepare the ZIP code request
	req := &models.ZIPCodeRequest{
		StreetAddress: "123 Main St",
		City:          "New York",
		State:         "NY",
	}

	// Get the ZIP code
	resp, err := client.GetZIPCode(context.Background(), req)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	fmt.Printf("ZIP Code: %s", resp.Address.ZIPCode)
	if resp.Address.ZIPPlus4 != nil {
		fmt.Printf("-%s", *resp.Address.ZIPPlus4)
	}
	fmt.Println()
}

func ExampleNewTestClient() {
	// Create a token provider with your test OAuth token
	tokenProvider := usps.NewStaticTokenProvider("your-test-oauth-token")

	// Create a client configured for the testing environment
	client := usps.NewTestClient(tokenProvider)

	req := &models.AddressRequest{
		StreetAddress: "123 Test St",
		City:          "Test City",
		State:         "NY",
	}

	resp, err := client.GetAddress(context.Background(), req)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	fmt.Printf("Test Address: %s\n", resp.Address.StreetAddress)
}

func ExampleOAuthClient_PostToken_clientCredentials() {
	// Create an OAuth client
	client := usps.NewOAuthClient()

	// Prepare the client credentials request
	req := &models.ClientCredentials{
		GrantType:    "client_credentials",
		ClientID:     "your-client-id",
		ClientSecret: "your-client-secret",
		Scope:        "addresses tracking labels",
	}

	// Get an access token
	result, err := client.PostToken(context.Background(), req)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	accessTokenResp := result.(*models.ProviderAccessTokenResponse)
	fmt.Printf("Access Token: %s\n", accessTokenResp.AccessToken)
	fmt.Printf("Expires In: %d seconds\n", accessTokenResp.ExpiresIn)
}

func ExampleOAuthClient_PostToken_refreshToken() {
	// Create an OAuth client
	client := usps.NewOAuthClient()

	// Prepare the refresh token request
	req := &models.RefreshTokenCredentials{
		GrantType:    "refresh_token",
		ClientID:     "your-client-id",
		ClientSecret: "your-client-secret",
		RefreshToken: "your-refresh-token",
	}

	// Get a new access token and refresh token
	result, err := client.PostToken(context.Background(), req)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	tokensResp := result.(*models.ProviderTokensResponse)
	fmt.Printf("Access Token: %s\n", tokensResp.AccessToken)
	fmt.Printf("New Refresh Token: %s\n", tokensResp.RefreshToken)
}

func ExampleOAuthClient_PostRevoke() {
	// Create an OAuth client
	client := usps.NewOAuthClient()

	// Prepare the revoke request
	req := &models.TokenRevokeRequest{
		Token:         "refresh-token-to-revoke",
		TokenTypeHint: "refresh_token",
	}

	// Revoke the token
	err := client.PostRevoke(context.Background(), "your-client-id", "your-client-secret", req)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	fmt.Println("Token revoked successfully")
}

func ExampleNewOAuthTestClient() {
	// Create an OAuth client configured for the testing environment
	client := usps.NewOAuthTestClient()

	req := &models.ClientCredentials{
		GrantType:    "client_credentials",
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
	}

	result, err := client.PostToken(context.Background(), req)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	accessTokenResp := result.(*models.ProviderAccessTokenResponse)
	fmt.Printf("Test Access Token obtained (expires in %d seconds)\n", accessTokenResp.ExpiresIn)
}

func ExampleNewOAuthTokenProvider() {
	// Create an OAuth token provider with your client credentials
	// This will automatically handle token acquisition and refresh
	tokenProvider := usps.NewOAuthTokenProvider("your-client-id", "your-client-secret")

	// Create a client using the OAuth token provider
	client := usps.NewClient(tokenProvider)

	// The token provider automatically manages tokens
	req := &models.AddressRequest{
		StreetAddress: "123 Main St",
		City:          "New York",
		State:         "NY",
	}

	resp, err := client.GetAddress(context.Background(), req)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	fmt.Printf("Address: %s\n", resp.Address.StreetAddress)
}

func ExampleNewOAuthTokenProvider_withOptions() {
	// Create an OAuth token provider with custom options
	tokenProvider := usps.NewOAuthTokenProvider(
		"your-client-id",
		"your-client-secret",
		usps.WithOAuthScopes("addresses tracking labels"),
		usps.WithTokenRefreshBuffer(10*time.Minute),
		usps.WithOAuthEnvironment("testing"),
		usps.WithRefreshTokens(true),
	)

	// Use with the USPS client
	client := usps.NewClient(tokenProvider)

	req := &models.CityStateRequest{
		ZIPCode: "10001",
	}

	resp, err := client.GetCityState(context.Background(), req)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	fmt.Printf("City: %s, State: %s\n", resp.City, resp.State)
}

func ExampleNewClientWithOAuth() {
	// Create a client with automatic OAuth token management in one step
	client := usps.NewClientWithOAuth("your-client-id", "your-client-secret")

	// Use the client - tokens are managed automatically
	req := &models.AddressRequest{
		StreetAddress: "123 Main St",
		City:          "New York",
		State:         "NY",
	}

	resp, err := client.GetAddress(context.Background(), req)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	fmt.Printf("Address: %s\n", resp.Address.StreetAddress)
}

func ExampleNewTestClientWithOAuth() {
	// Create a test client with automatic OAuth token management
	client := usps.NewTestClientWithOAuth("test-client-id", "test-client-secret")

	req := &models.CityStateRequest{
		ZIPCode: "10001",
	}

	resp, err := client.GetCityState(context.Background(), req)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	fmt.Printf("City: %s, State: %s\n", resp.City, resp.State)
}
