package usps_test

import (
	"context"
	"os"
	"testing"

	"github.com/my-eq/go-usps"
	"github.com/my-eq/go-usps/models"
)

// TestIntegration_GetAddress is an integration test that requires a real USPS OAuth token.
// Set the USPS_TOKEN environment variable to run this test.
func TestIntegration_GetAddress(t *testing.T) {
	token := os.Getenv("USPS_TOKEN")
	if token == "" {
		t.Skip("Skipping integration test: USPS_TOKEN not set")
	}

	tokenProvider := usps.NewStaticTokenProvider(token)
	client := usps.NewClient(tokenProvider)

	req := &models.AddressRequest{
		StreetAddress: "475 L'Enfant Plaza SW",
		City:          "Washington",
		State:         "DC",
	}

	resp, err := client.GetAddress(context.Background(), req)
	if err != nil {
		t.Fatalf("GetAddress failed: %v", err)
	}

	if resp.Address == nil {
		t.Fatal("Expected address in response")
	}

	t.Logf("Standardized address: %s, %s, %s %s",
		resp.Address.StreetAddress,
		resp.Address.City,
		resp.Address.State,
		resp.Address.ZIPCode)
}

// TestIntegration_GetCityState is an integration test that requires a real USPS OAuth token.
// Set the USPS_TOKEN environment variable to run this test.
func TestIntegration_GetCityState(t *testing.T) {
	token := os.Getenv("USPS_TOKEN")
	if token == "" {
		t.Skip("Skipping integration test: USPS_TOKEN not set")
	}

	tokenProvider := usps.NewStaticTokenProvider(token)
	client := usps.NewClient(tokenProvider)

	req := &models.CityStateRequest{
		ZIPCode: "20260",
	}

	resp, err := client.GetCityState(context.Background(), req)
	if err != nil {
		t.Fatalf("GetCityState failed: %v", err)
	}

	t.Logf("City: %s, State: %s", resp.City, resp.State)
}

// TestIntegration_GetZIPCode is an integration test that requires a real USPS OAuth token.
// Set the USPS_TOKEN environment variable to run this test.
func TestIntegration_GetZIPCode(t *testing.T) {
	token := os.Getenv("USPS_TOKEN")
	if token == "" {
		t.Skip("Skipping integration test: USPS_TOKEN not set")
	}

	tokenProvider := usps.NewStaticTokenProvider(token)
	client := usps.NewClient(tokenProvider)

	req := &models.ZIPCodeRequest{
		StreetAddress: "475 L'Enfant Plaza SW",
		City:          "Washington",
		State:         "DC",
	}

	resp, err := client.GetZIPCode(context.Background(), req)
	if err != nil {
		t.Fatalf("GetZIPCode failed: %v", err)
	}

	if resp.Address == nil {
		t.Fatal("Expected address in response")
	}

	t.Logf("ZIP Code: %s", resp.Address.ZIPCode)
	if resp.Address.ZIPPlus4 != nil {
		t.Logf("ZIP+4: %s", *resp.Address.ZIPPlus4)
	}
}
