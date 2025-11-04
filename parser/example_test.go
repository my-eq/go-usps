package parser_test

import (
	"context"
	"fmt"
	"log"

	"github.com/my-eq/go-usps"
	"github.com/my-eq/go-usps/parser"
)

func ExampleParse() {
	// Parse a free-form address string
	input := "123 North Main Street Apartment 4B, New York, NY 10001-1234"
	parsed, diagnostics := parser.Parse(input)

	// Check for diagnostics
	if len(diagnostics) > 0 {
		for _, d := range diagnostics {
			fmt.Printf("%s: %s\n", d.Severity, d.Message)
		}
	}

	// Convert to AddressRequest for USPS API
	req := parsed.ToAddressRequest()

	fmt.Printf("Street: %s\n", req.StreetAddress)
	fmt.Printf("Secondary: %s\n", req.SecondaryAddress)
	fmt.Printf("City: %s\n", req.City)
	fmt.Printf("State: %s\n", req.State)
	fmt.Printf("ZIP: %s\n", req.ZIPCode)
	fmt.Printf("ZIP+4: %s\n", req.ZIPPlus4)

	// Output:
	// Street: 123 N MAIN ST
	// Secondary: APT 4B
	// City: NEW YORK
	// State: NY
	// ZIP: 10001
	// ZIP+4: 1234
}

func ExampleParse_withUSPSValidation() {
	// Parse a free-form address
	input := "456 Oak Avenue, Los Angeles, CA 90001"
	parsed, diagnostics := parser.Parse(input)

	// Check for warnings or errors
	for _, d := range diagnostics {
		log.Printf("%s: %s - %s\n", d.Severity, d.Message, d.Remediation)
	}

	// Convert to AddressRequest
	req := parsed.ToAddressRequest()

	// Use with USPS client to validate
	client := usps.NewClientWithOAuth("client-id", "client-secret")
	resp, err := client.GetAddress(context.Background(), req)
	if err != nil {
		log.Fatalf("Validation failed: %v", err)
	}

	fmt.Printf("Validated: %s, %s, %s %s\n",
		resp.Address.StreetAddress,
		resp.Address.City,
		resp.Address.State,
		resp.Address.ZIPCode)
}

func ExampleParsedAddress_ToAddressRequest() {
	// Parse an address
	parsed, _ := parser.Parse("789 Elm Blvd Suite 200, Chicago, IL 60601")

	// Convert to AddressRequest for API calls
	req := parsed.ToAddressRequest()

	fmt.Printf("Ready for USPS API:\n")
	fmt.Printf("  Street: %s\n", req.StreetAddress)
	fmt.Printf("  Secondary: %s\n", req.SecondaryAddress)
	fmt.Printf("  City: %s\n", req.City)
	fmt.Printf("  State: %s\n", req.State)
	fmt.Printf("  ZIP: %s\n", req.ZIPCode)

	// Output:
	// Ready for USPS API:
	//   Street: 789 ELM BLVD
	//   Secondary: STE 200
	//   City: CHICAGO
	//   State: IL
	//   ZIP: 60601
}
