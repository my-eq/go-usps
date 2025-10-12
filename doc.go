// Package usps provides a lightweight, production-grade Go client library
// for the USPS Addresses 3.0 REST API.
//
// The package implements all three USPS Addresses 3.0 API endpoints:
//   - Address Standardization (GetAddress)
//   - City/State Lookup (GetCityState)
//   - ZIP Code Lookup (GetZIPCode)
//
// # Quick Start
//
// Create a client with your OAuth token:
//
//	tokenProvider := usps.NewStaticTokenProvider("your-oauth-token")
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
// For more information, see https://developers.usps.com/addressesv3
package usps
