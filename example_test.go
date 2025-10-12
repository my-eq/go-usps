package usps_test

import (
	"context"
	"fmt"
	"log"

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
