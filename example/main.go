package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/my-eq/go-usps"
)

func main() {
	// Create a new USPS client with default HTTP client
	client := usps.NewClient("your-api-key-here")

	// Create an address validation request
	addr := &usps.AddressRequest{
		StreetAddress: "123 Main St",
		City:          "Anytown",
		State:         "CA",
		ZIPCode:       "12345",
	}

	// Validate the address
	ctx := context.Background()
	result, err := client.ValidateAddress(ctx, addr)
	if err != nil {
		log.Fatalf("Error validating address: %v", err)
	}

	if result.Error != nil {
		fmt.Printf("API Error: %s - %s\n", result.Error.Code, result.Error.Message)
		return
	}

	if result.Address != nil {
		fmt.Printf("Validated Address:\n")
		fmt.Printf("  Street: %s\n", result.Address.StreetAddress)
		fmt.Printf("  City: %s\n", result.Address.City)
		fmt.Printf("  State: %s\n", result.Address.State)
		fmt.Printf("  ZIP: %s-%s\n", result.Address.ZIPCode, result.Address.ZIPPlus4)
	}
}

func exampleWithCustomHTTPClient() {
	// Create a custom retry configuration
	retryConfig := &usps.RetryConfig{
		MaxRetries:     5,
		InitialBackoff: 1 * time.Second,
		MaxBackoff:     10 * time.Second,
		Multiplier:     2.0,
	}

	// Create a custom HTTP client with retry configuration
	httpClient := usps.NewDefaultHTTPClient(retryConfig)

	// Create USPS client with custom HTTP client
	client := usps.NewClient("your-api-key-here", usps.WithHTTPClient(httpClient))

	// Use the client
	addr := &usps.AddressRequest{
		StreetAddress: "123 Main St",
		City:          "Anytown",
		State:         "CA",
		ZIPCode:       "12345",
	}

	ctx := context.Background()
	result, err := client.ValidateAddress(ctx, addr)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	fmt.Printf("Result: %+v\n", result)
}
