// Package models provides strongly-typed data structures for the USPS Addresses API and OAuth 2.0 API.
//
// All types are based on the USPS OpenAPI specifications and include
// proper JSON and URL encoding serialization tags.
//
// # Address API Request Types
//   - AddressRequest: Parameters for address standardization
//   - CityStateRequest: Parameters for city/state lookup by ZIP code
//   - ZIPCodeRequest: Parameters for ZIP code lookup by address
//
// # Address API Response Types
//   - AddressResponse: Standardized address with additional information
//   - CityStateResponse: City and state for a given ZIP code
//   - ZIPCodeResponse: ZIP code and ZIP+4 for a given address
//
// # OAuth 2.0 Request Types
//   - ClientCredentials: Client credentials grant request
//   - RefreshTokenCredentials: Refresh token grant request
//   - AuthorizationCodeCredentials: Authorization code grant request
//   - TokenRevokeRequest: Token revocation request
//
// # OAuth 2.0 Response Types
//   - ProviderAccessTokenResponse: Access token response (client credentials)
//   - ProviderTokensResponse: Access and refresh token response
//   - StandardErrorResponse: OAuth error response
//
// # Error Types
//   - ErrorMessage: Standard USPS API error response
//   - ErrorInfo: High-level error information
//   - ErrorDetail: Detailed error information
package models
