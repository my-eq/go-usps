// Package models provides strongly-typed data structures for the USPS Addresses API.
//
// All types are based on the USPS Addresses 3.0 OpenAPI specification and include
// proper JSON serialization tags.
//
// Request Types:
//   - AddressRequest: Parameters for address standardization
//   - CityStateRequest: Parameters for city/state lookup by ZIP code
//   - ZIPCodeRequest: Parameters for ZIP code lookup by address
//
// Response Types:
//   - AddressResponse: Standardized address with additional information
//   - CityStateResponse: City and state for a given ZIP code
//   - ZIPCodeResponse: ZIP code and ZIP+4 for a given address
//
// Error Types:
//   - ErrorMessage: Standard USPS API error response
//   - ErrorInfo: High-level error information
//   - ErrorDetail: Detailed error information
package models
