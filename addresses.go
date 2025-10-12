package usps

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
)

// AddressRequest represents an address validation request
type AddressRequest struct {
	StreetAddress    string `json:"streetAddress"`
	SecondaryAddress string `json:"secondaryAddress,omitempty"`
	City             string `json:"city,omitempty"`
	State            string `json:"state,omitempty"`
	ZIPCode          string `json:"ZIPCode,omitempty"`
	ZIPPlus4         string `json:"ZIPPlus4,omitempty"`
	Urbanization     string `json:"urbanization,omitempty"`
}

// AddressResponse represents the validated address response
type AddressResponse struct {
	Address *Address `json:"address,omitempty"`
	Error   *Error   `json:"error,omitempty"`
}

// Address represents a validated address
type Address struct {
	StreetAddress    string `json:"streetAddress"`
	SecondaryAddress string `json:"secondaryAddress,omitempty"`
	City             string `json:"city"`
	State            string `json:"state"`
	ZIPCode          string `json:"ZIPCode"`
	ZIPPlus4         string `json:"ZIPPlus4,omitempty"`
	Urbanization     string `json:"urbanization,omitempty"`
	DeliveryPoint    string `json:"deliveryPoint,omitempty"`
	CarrierRoute     string `json:"carrierRoute,omitempty"`
	ValidatedAddress bool   `json:"validatedAddress,omitempty"`
}

// Error represents an API error response
type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Detail  string `json:"detail,omitempty"`
}

// ValidateAddress validates a single address using the USPS Addresses API v3
func (c *Client) ValidateAddress(ctx context.Context, addr *AddressRequest) (*AddressResponse, error) {
	resp, err := c.do(ctx, "POST", "/addresses/v3/address", addr)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var result AddressResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &result, nil
}
