package usps_test

import (
	"context"
	"fmt"
	"log"

	"github.com/my-eq/go-usps"
	"github.com/my-eq/go-usps/models"
)

// ExampleBulkProcessor_ProcessAddresses demonstrates bulk address validation
func ExampleBulkProcessor_ProcessAddresses() {
	// Create client (in production, use real credentials)
	client := usps.NewClientWithOAuth("client-id", "client-secret")

	// Configure bulk processor
	config := &usps.BulkConfig{
		MaxConcurrency:    10, // Process 10 addresses concurrently
		RequestsPerSecond: 10, // Rate limit to 10 requests/second
		MaxRetries:        3,  // Retry failed requests up to 3 times
		ProgressCallback: func(completed, total int, err error) {
			fmt.Printf("Progress: %d/%d\n", completed, total)
		},
	}

	processor := usps.NewBulkProcessor(client, config)

	// Prepare requests
	requests := []*models.AddressRequest{
		{StreetAddress: "123 Main St", City: "New York", State: "NY"},
		{StreetAddress: "456 Oak Ave", City: "Los Angeles", State: "CA"},
		{StreetAddress: "789 Elm Blvd", City: "Chicago", State: "IL"},
	}

	// Process bulk addresses
	results := processor.ProcessAddresses(context.Background(), requests)

	// Handle results
	for _, result := range results {
		if result.Error != nil {
			log.Printf("Address %d failed: %v", result.Index, result.Error)
			continue
		}

		fmt.Printf("Standardized: %s, %s, %s %s\n",
			result.Response.Address.StreetAddress,
			result.Response.Address.City,
			result.Response.Address.State,
			result.Response.Address.ZIPCode)
	}
}

// ExampleBulkProcessor_ProcessCityStates demonstrates bulk city/state lookup
func ExampleBulkProcessor_ProcessCityStates() {
	client := usps.NewClientWithOAuth("client-id", "client-secret")
	processor := usps.NewBulkProcessor(client, usps.DefaultBulkConfig())

	requests := []*models.CityStateRequest{
		{ZIPCode: "10001"},
		{ZIPCode: "90210"},
		{ZIPCode: "60601"},
	}

	results := processor.ProcessCityStates(context.Background(), requests)

	for _, result := range results {
		if result.Error != nil {
			log.Printf("ZIP %s failed: %v", result.Request.ZIPCode, result.Error)
			continue
		}

		fmt.Printf("%s: %s, %s\n",
			result.Request.ZIPCode,
			result.Response.City,
			result.Response.State)
	}
}

// ExampleBulkProcessor_ProcessZIPCodes demonstrates bulk ZIP code lookup
func ExampleBulkProcessor_ProcessZIPCodes() {
	client := usps.NewClientWithOAuth("client-id", "client-secret")

	// Custom configuration for high-volume processing
	config := &usps.BulkConfig{
		MaxConcurrency:    20, // Higher concurrency
		RequestsPerSecond: 50, // Higher rate limit (if your API plan allows)
		MaxRetries:        5,  // More retries for critical operations
	}

	processor := usps.NewBulkProcessor(client, config)

	requests := []*models.ZIPCodeRequest{
		{StreetAddress: "123 Main St", City: "New York", State: "NY"},
		{StreetAddress: "456 Oak Ave", City: "Los Angeles", State: "CA"},
	}

	results := processor.ProcessZIPCodes(context.Background(), requests)

	for _, result := range results {
		if result.Error != nil {
			log.Printf("Request %d failed: %v", result.Index, result.Error)
			continue
		}

		zip := result.Response.Address.ZIPCode
		if result.Response.Address.ZIPPlus4 != nil && *result.Response.Address.ZIPPlus4 != "" {
			zip = fmt.Sprintf("%s-%s", zip, *result.Response.Address.ZIPPlus4)
		}

		fmt.Printf("ZIP Code: %s\n", zip)
	}
}

// ExampleDefaultBulkConfig demonstrates using the default bulk configuration
func ExampleDefaultBulkConfig() {
	client := usps.NewClientWithOAuth("client-id", "client-secret")

	// Use default configuration
	processor := usps.NewBulkProcessor(client, usps.DefaultBulkConfig())

	requests := []*models.AddressRequest{
		{StreetAddress: "475 L'Enfant Plaza SW", City: "Washington", State: "DC"},
	}

	results := processor.ProcessAddresses(context.Background(), requests)

	fmt.Printf("Processed %d addresses\n", len(results))
	// Output: Processed 1 addresses
}
