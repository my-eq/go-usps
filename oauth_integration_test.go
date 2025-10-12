package usps_test

import (
	"context"
	"os"
	"testing"

	"github.com/my-eq/go-usps"
	"github.com/my-eq/go-usps/models"
)

// TestIntegration_OAuth_PostToken is an integration test that requires real USPS client credentials.
// Set the USPS_CLIENT_ID and USPS_CLIENT_SECRET environment variables to run this test.
func TestIntegration_OAuth_PostToken(t *testing.T) {
	clientID := os.Getenv("USPS_CLIENT_ID")
	clientSecret := os.Getenv("USPS_CLIENT_SECRET")
	if clientID == "" || clientSecret == "" {
		t.Skip("Skipping integration test: USPS_CLIENT_ID and USPS_CLIENT_SECRET not set")
	}

	client := usps.NewOAuthClient()

	req := &models.ClientCredentials{
		GrantType:    "client_credentials",
		ClientID:     clientID,
		ClientSecret: clientSecret,
	}

	result, err := client.PostToken(context.Background(), req)
	if err != nil {
		t.Fatalf("PostToken failed: %v", err)
	}

	accessTokenResp, ok := result.(*models.ProviderAccessTokenResponse)
	if !ok {
		t.Fatalf("Expected *models.ProviderAccessTokenResponse, got %T", result)
	}

	if accessTokenResp.AccessToken == "" {
		t.Error("Expected non-empty access token")
	}
	if accessTokenResp.TokenType != "Bearer" {
		t.Errorf("Expected token type 'Bearer', got '%s'", accessTokenResp.TokenType)
	}

	t.Logf("Access Token obtained successfully (expires in %d seconds)", accessTokenResp.ExpiresIn)
}

// TestIntegration_OAuth_PostRevoke is an integration test that requires real USPS client credentials and a refresh token.
// Set the USPS_CLIENT_ID, USPS_CLIENT_SECRET, and USPS_REFRESH_TOKEN environment variables to run this test.
func TestIntegration_OAuth_PostRevoke(t *testing.T) {
	clientID := os.Getenv("USPS_CLIENT_ID")
	clientSecret := os.Getenv("USPS_CLIENT_SECRET")
	refreshToken := os.Getenv("USPS_REFRESH_TOKEN")
	if clientID == "" || clientSecret == "" || refreshToken == "" {
		t.Skip("Skipping integration test: USPS_CLIENT_ID, USPS_CLIENT_SECRET, and USPS_REFRESH_TOKEN not set")
	}

	client := usps.NewOAuthClient()

	req := &models.TokenRevokeRequest{
		Token:         refreshToken,
		TokenTypeHint: "refresh_token",
	}

	err := client.PostRevoke(context.Background(), clientID, clientSecret, req)
	if err != nil {
		t.Fatalf("PostRevoke failed: %v", err)
	}

	t.Log("Token revoked successfully")
}
