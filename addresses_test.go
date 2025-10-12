package usps

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"
)

func TestValidateAddress(t *testing.T) {
	apiKey := "test-api-key"
	responseJSON := `{
		"address": {
			"streetAddress": "123 MAIN ST",
			"city": "ANYTOWN",
			"state": "CA",
			"ZIPCode": "12345",
			"ZIPPlus4": "6789",
			"validatedAddress": true
		}
	}`

	mockClient := &MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(bytes.NewBufferString(responseJSON)),
			}, nil
		},
	}

	client := NewClient(apiKey, WithHTTPClient(mockClient))

	addr := &AddressRequest{
		StreetAddress: "123 Main St",
		City:          "Anytown",
		State:         "CA",
		ZIPCode:       "12345",
	}

	result, err := client.ValidateAddress(context.Background(), addr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Address == nil {
		t.Fatal("expected address in response")
	}

	if result.Address.StreetAddress != "123 MAIN ST" {
		t.Errorf("expected street address '123 MAIN ST', got '%s'", result.Address.StreetAddress)
	}

	if result.Address.City != "ANYTOWN" {
		t.Errorf("expected city 'ANYTOWN', got '%s'", result.Address.City)
	}

	if result.Address.State != "CA" {
		t.Errorf("expected state 'CA', got '%s'", result.Address.State)
	}

	if result.Address.ZIPCode != "12345" {
		t.Errorf("expected ZIP code '12345', got '%s'", result.Address.ZIPCode)
	}

	if result.Address.ZIPPlus4 != "6789" {
		t.Errorf("expected ZIP+4 '6789', got '%s'", result.Address.ZIPPlus4)
	}

	if !result.Address.ValidatedAddress {
		t.Error("expected validatedAddress to be true")
	}
}

func TestValidateAddressWithError(t *testing.T) {
	apiKey := "test-api-key"
	responseJSON := `{
		"error": {
			"code": "INVALID_ADDRESS",
			"message": "Address not found",
			"detail": "The address could not be validated"
		}
	}`

	mockClient := &MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(bytes.NewBufferString(responseJSON)),
			}, nil
		},
	}

	client := NewClient(apiKey, WithHTTPClient(mockClient))

	addr := &AddressRequest{
		StreetAddress: "Invalid Address",
	}

	result, err := client.ValidateAddress(context.Background(), addr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Error == nil {
		t.Fatal("expected error in response")
	}

	if result.Error.Code != "INVALID_ADDRESS" {
		t.Errorf("expected error code 'INVALID_ADDRESS', got '%s'", result.Error.Code)
	}

	if result.Error.Message != "Address not found" {
		t.Errorf("expected error message 'Address not found', got '%s'", result.Error.Message)
	}
}

func TestAddressRequest(t *testing.T) {
	addr := &AddressRequest{
		StreetAddress:    "123 Main St",
		SecondaryAddress: "Apt 4",
		City:             "Anytown",
		State:            "CA",
		ZIPCode:          "12345",
		ZIPPlus4:         "6789",
		Urbanization:     "URB LAS GLADIOLAS",
	}

	if addr.StreetAddress != "123 Main St" {
		t.Errorf("expected street address '123 Main St', got '%s'", addr.StreetAddress)
	}

	if addr.SecondaryAddress != "Apt 4" {
		t.Errorf("expected secondary address 'Apt 4', got '%s'", addr.SecondaryAddress)
	}

	if addr.City != "Anytown" {
		t.Errorf("expected city 'Anytown', got '%s'", addr.City)
	}

	if addr.State != "CA" {
		t.Errorf("expected state 'CA', got '%s'", addr.State)
	}

	if addr.ZIPCode != "12345" {
		t.Errorf("expected ZIP code '12345', got '%s'", addr.ZIPCode)
	}

	if addr.Urbanization != "URB LAS GLADIOLAS" {
		t.Errorf("expected urbanization 'URB LAS GLADIOLAS', got '%s'", addr.Urbanization)
	}
}
