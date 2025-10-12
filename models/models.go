package models

// Address represents the standard address fields common to all locations.
type Address struct {
	StreetAddress             string `json:"streetAddress,omitempty"`
	StreetAddressAbbreviation string `json:"streetAddressAbbreviation,omitempty"`
	SecondaryAddress          string `json:"secondaryAddress,omitempty"`
	CityAbbreviation          string `json:"cityAbbreviation,omitempty"`
}

// DomesticAddress represents address fields for US locations.
type DomesticAddress struct {
	Address
	City         string  `json:"city,omitempty"`
	State        string  `json:"state,omitempty"`
	ZIPCode      string  `json:"ZIPCode,omitempty"`
	ZIPPlus4     *string `json:"ZIPPlus4,omitempty"`
	Urbanization string  `json:"urbanization,omitempty"`
}

// AddressAdditionalInfo contains extra information about the address.
type AddressAdditionalInfo struct {
	DeliveryPoint        string `json:"deliveryPoint,omitempty"`
	CarrierRoute         string `json:"carrierRoute,omitempty"`
	DPVConfirmation      string `json:"DPVConfirmation,omitempty"`
	DPVCMRA              string `json:"DPVCMRA,omitempty"`
	Business             string `json:"business,omitempty"`
	CentralDeliveryPoint string `json:"centralDeliveryPoint,omitempty"`
	Vacant               string `json:"vacant,omitempty"`
}

// AddressCorrection represents a code indicating how to improve the address input.
type AddressCorrection struct {
	Code string `json:"code,omitempty"`
	Text string `json:"text,omitempty"`
}

// AddressMatch represents a code indicating if an address is an exact match.
type AddressMatch struct {
	Code string `json:"code,omitempty"`
	Text string `json:"text,omitempty"`
}

// AddressResponse represents the response from the address standardization endpoint.
type AddressResponse struct {
	Firm           string                 `json:"firm,omitempty"`
	Address        *DomesticAddress       `json:"address,omitempty"`
	AdditionalInfo *AddressAdditionalInfo `json:"additionalInfo,omitempty"`
	Corrections    []AddressCorrection    `json:"corrections,omitempty"`
	Matches        []AddressMatch         `json:"matches,omitempty"`
	Warnings       []string               `json:"warnings,omitempty"`
}

// CityStateResponse represents the response from the city-state lookup endpoint.
type CityStateResponse struct {
	City    string `json:"city,omitempty"`
	State   string `json:"state,omitempty"`
	ZIPCode string `json:"ZIPCode,omitempty"`
}

// ZIPCodeResponse represents the response from the ZIP code lookup endpoint.
type ZIPCodeResponse struct {
	Firm    string           `json:"firm,omitempty"`
	Address *DomesticAddress `json:"address,omitempty"`
}

// ErrorSource represents the element that is suspected of originating the error.
type ErrorSource struct {
	Parameter string `json:"parameter,omitempty"`
	Example   string `json:"example,omitempty"`
}

// ErrorDetail represents a detailed error within an error response.
type ErrorDetail struct {
	Status string       `json:"status,omitempty"`
	Code   string       `json:"code,omitempty"`
	Title  string       `json:"title,omitempty"`
	Detail string       `json:"detail,omitempty"`
	Source *ErrorSource `json:"source,omitempty"`
}

// ErrorInfo represents the high-level error information.
type ErrorInfo struct {
	Code    string        `json:"code,omitempty"`
	Message string        `json:"message,omitempty"`
	Errors  []ErrorDetail `json:"errors,omitempty"`
}

// ErrorMessage represents the standard USPS API error response.
type ErrorMessage struct {
	APIVersion string     `json:"apiVersion,omitempty"`
	Error      *ErrorInfo `json:"error,omitempty"`
}
