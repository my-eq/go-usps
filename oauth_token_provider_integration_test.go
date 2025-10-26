package usps

import (
	"context"
	"os"
	"testing"

	"github.com/my-eq/go-usps/models"
)

func TestIntegration_OAuthTokenProvider(t *testing.T) {
	clientID := os.Getenv("USPS_CLIENT_ID")
	clientSecret := os.Getenv("USPS_CLIENT_SECRET")

	if clientID == "" || clientSecret == "" {
		t.Skip("Skipping integration test: USPS_CLIENT_ID and USPS_CLIENT_SECRET not set")
	}

	// Create OAuth token provider
	provider := NewOAuthTokenProvider(
		clientID,
		clientSecret,
		WithOAuthScopes("addresses"),
	)

	// Create client using the OAuth token provider
	client := NewClient(provider)

	// Test getting an address - this will trigger token acquisition
	req := &models.AddressRequest{
		StreetAddress: "475 L ENFANT PLZ SW RM 10541",
		City:          "WASHINGTON",
		State:         "DC",
		ZIPCode:       "20260",
	}

	resp, err := client.GetAddress(context.Background(), req)
	if err != nil {
		t.Fatalf("GetAddress failed: %v", err)
	}

	if resp.Address == nil {
		t.Fatal("Expected address in response, got nil")
	}

	t.Logf("Successfully got address: %s, %s, %s %s",
		resp.Address.StreetAddress,
		resp.Address.City,
		resp.Address.State,
		resp.Address.ZIPCode)

	// Make another request - should use cached token
	resp2, err := client.GetAddress(context.Background(), req)
	if err != nil {
		t.Fatalf("Second GetAddress failed: %v", err)
	}

	if resp2.Address == nil {
		t.Fatal("Expected address in second response, got nil")
	}

	t.Log("Successfully used cached token for second request")
}

func TestIntegration_OAuthTokenProvider_TestingEnvironment(t *testing.T) {
	clientID := os.Getenv("USPS_TEST_CLIENT_ID")
	clientSecret := os.Getenv("USPS_TEST_CLIENT_SECRET")

	if clientID == "" || clientSecret == "" {
		t.Skip("Skipping integration test: USPS_TEST_CLIENT_ID and USPS_TEST_CLIENT_SECRET not set")
	}

	// Create OAuth token provider for testing environment
	provider := NewOAuthTokenProvider(
		clientID,
		clientSecret,
		WithOAuthEnvironment("testing"),
		WithOAuthScopes("addresses"),
	)

	// Create test client using the OAuth token provider
	client := NewTestClient(provider)

	// Test getting an address
	req := &models.AddressRequest{
		StreetAddress: "123 Main St",
		City:          "New York",
		State:         "NY",
	}

	resp, err := client.GetAddress(context.Background(), req)
	if err != nil {
		t.Fatalf("GetAddress failed: %v", err)
	}

	if resp.Address == nil {
		t.Fatal("Expected address in response, got nil")
	}

	t.Logf("Successfully got test address: %s", resp.Address.StreetAddress)
}
