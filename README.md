# go-usps

A lightweight, production-grade Go client library for the USPS Addresses 3.0 REST API and OAuth 2.0 API.

[![Go Reference](https://pkg.go.dev/badge/github.com/my-eq/go-usps.svg)](https://pkg.go.dev/github.com/my-eq/go-usps)
[![Go Report Card](https://goreportcard.com/badge/github.com/my-eq/go-usps)](https://goreportcard.com/report/github.com/my-eq/go-usps)
[![Go Build, Lint and Test](https://github.com/my-eq/go-usps/actions/workflows/go.yml/badge.svg)](https://github.com/my-eq/go-usps/actions/workflows/go.yml)
[![Markdown Lint](https://github.com/my-eq/go-usps/actions/workflows/markdown-lint.yml/badge.svg)](https://github.com/my-eq/go-usps/actions/workflows/markdown-lint.yml)

## Features

- **🎯 Complete Coverage**: Implements all USPS Addresses 3.0 API endpoints and OAuth 2.0 endpoints
- **💪 Strongly Typed**: All request and response types are fully typed based on the OpenAPI specification
- **🔒 OAuth Support**: Built-in OAuth 2.0 with Client Credentials, Refresh Token, and Authorization Code grants
- **🧪 Testable**: Designed with dependency injection for easy testing
- **📦 Lightweight**: Zero external dependencies, uses Go standard library only
- **🏗️ Production Ready**: Used by millions of people as a critical part of their workflows

## Installation

```bash
go get github.com/my-eq/go-usps
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/my-eq/go-usps"
    "github.com/my-eq/go-usps/models"
)

func main() {
    // Create a token provider with your OAuth token
    tokenProvider := usps.NewStaticTokenProvider("your-oauth-token")

    // Create a new USPS client
    client := usps.NewClient(tokenProvider)

    // Standardize an address
    req := &models.AddressRequest{
        StreetAddress: "123 Main St",
        City:          "New York",
        State:         "NY",
    }

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
```

## API Endpoints

### Address Standardization

Standardizes street addresses including city and street abbreviations, and
provides missing information such as ZIP Code™ and ZIP + 4®.

```go
req := &models.AddressRequest{
    StreetAddress:    "123 Main St",
    SecondaryAddress: "Apt 4B",  // Optional
    City:             "New York",
    State:            "NY",
    ZIPCode:          "10001",   // Optional
}

resp, err := client.GetAddress(ctx, req)
```

### City/State Lookup

Returns the city and state corresponding to the given ZIP Code™.

```go
req := &models.CityStateRequest{
    ZIPCode: "10001",
}

resp, err := client.GetCityState(ctx, req)
```

### ZIP Code Lookup

Returns the ZIP Code™ and ZIP + 4® corresponding to the given address, city, and state.

```go
req := &models.ZIPCodeRequest{
    StreetAddress: "123 Main St",
    City:          "New York",
    State:         "NY",
}

resp, err := client.GetZIPCode(ctx, req)
```

## OAuth 2.0 Authentication

The library provides full support for USPS OAuth 2.0 API endpoints to obtain and manage access tokens.

### Obtaining an Access Token

Use the Client Credentials grant to obtain an access token:

```go
// Create an OAuth client
oauthClient := usps.NewOAuthClient()

// Request an access token using client credentials
req := &models.ClientCredentials{
    GrantType:    "client_credentials",
    ClientID:     "your-client-id",
    ClientSecret: "your-client-secret",
    Scope:        "addresses tracking labels",  // Optional
}

result, err := oauthClient.PostToken(context.Background(), req)
if err != nil {
    log.Fatalf("Error: %v", err)
}

// Type assert to get the response
accessTokenResp := result.(*models.ProviderAccessTokenResponse)
fmt.Printf("Access Token: %s\n", accessTokenResp.AccessToken)
fmt.Printf("Expires In: %d seconds\n", accessTokenResp.ExpiresIn)

// Use the access token with the addresses API
tokenProvider := usps.NewStaticTokenProvider(accessTokenResp.AccessToken)
client := usps.NewClient(tokenProvider)
```

**Note**: Access tokens are valid for 8 hours after issuance.

### Refreshing an Access Token

Use a refresh token to obtain a new access token:

```go
req := &models.RefreshTokenCredentials{
    GrantType:    "refresh_token",
    ClientID:     "your-client-id",
    ClientSecret: "your-client-secret",
    RefreshToken: "your-refresh-token",
}

result, err := oauthClient.PostToken(context.Background(), req)
if err != nil {
    log.Fatalf("Error: %v", err)
}

tokensResp := result.(*models.ProviderTokensResponse)
fmt.Printf("New Access Token: %s\n", tokensResp.AccessToken)
fmt.Printf("New Refresh Token: %s\n", tokensResp.RefreshToken)
```

**Note**: Refresh tokens are valid for 7 days after issuance.

### Authorization Code Grant

Use an authorization code to obtain access and refresh tokens:

```go
req := &models.AuthorizationCodeCredentials{
    GrantType:    "authorization_code",
    ClientID:     "your-client-id",
    ClientSecret: "your-client-secret",
    Code:         "authorization-code",
    RedirectURI:  "https://yourapp.com/callback",
}

result, err := oauthClient.PostToken(context.Background(), req)
if err != nil {
    log.Fatalf("Error: %v", err)
}

tokensResp := result.(*models.ProviderTokensResponse)
// Use tokensResp.AccessToken and tokensResp.RefreshToken
```

### Revoking a Token

Revoke a refresh token when it's no longer needed:

```go
req := &models.TokenRevokeRequest{
    Token:         "refresh-token-to-revoke",
    TokenTypeHint: "refresh_token",  // Optional: "access_token" or "refresh_token"
}

err := oauthClient.PostRevoke(
    context.Background(),
    "your-client-id",
    "your-client-secret",
    req,
)
if err != nil {
    log.Fatalf("Error: %v", err)
}

fmt.Println("Token revoked successfully")
```

### OAuth Error Handling

OAuth errors are returned as `*OAuthError`:

```go
result, err := oauthClient.PostToken(ctx, req)
if err != nil {
    if oauthErr, ok := err.(*usps.OAuthError); ok {
        fmt.Printf("OAuth Error (status %d): %s\n",
            oauthErr.StatusCode,
            oauthErr.ErrorMessage.Error)
        if oauthErr.ErrorMessage.ErrorDescription != "" {
            fmt.Printf("Description: %s\n", oauthErr.ErrorMessage.ErrorDescription)
        }
    } else {
        fmt.Printf("Error: %v\n", err)
    }
    return
}
```

## Configuration

### Production vs Testing Environment

```go
// Production environment (default)
client := usps.NewClient(tokenProvider)
oauthClient := usps.NewOAuthClient()

// Testing environment
client := usps.NewTestClient(tokenProvider)
oauthClient := usps.NewOAuthTestClient()
```

### Custom Options

```go
client := usps.NewClient(
    tokenProvider,
    usps.WithTimeout(60 * time.Second),           // Custom timeout
    usps.WithBaseURL("https://custom.url.com"),   // Custom base URL
    usps.WithHTTPClient(customHTTPClient),        // Custom HTTP client
)
```

### Custom Token Provider

Implement the `TokenProvider` interface for advanced authentication scenarios:

```go
type TokenProvider interface {
    GetToken(ctx context.Context) (string, error)
}

type MyTokenProvider struct {
    // Your implementation
}

func (p *MyTokenProvider) GetToken(ctx context.Context) (string, error) {
    // Fetch token from your auth service, refresh if needed, etc.
    return "token", nil
}

client := usps.NewClient(&MyTokenProvider{})
```

## Error Handling

The library provides structured error handling with detailed error information:

```go
resp, err := client.GetAddress(ctx, req)
if err != nil {
    // Check if it's an API error
    if apiErr, ok := err.(*usps.APIError); ok {
        fmt.Printf("API Error (status %d): %s\n", 
            apiErr.StatusCode, 
            apiErr.ErrorMessage.Error.Message)
        
        // Access detailed error information
        for _, detail := range apiErr.ErrorMessage.Error.Errors {
            fmt.Printf("  - %s: %s\n", detail.Title, detail.Detail)
        }
    } else {
        // Handle other errors (network, etc.)
        fmt.Printf("Error: %v\n", err)
    }
    return
}
```

## Response Fields

### AddressResponse

```go
type AddressResponse struct {
    Firm           string                 // Business name
    Address        *DomesticAddress       // Standardized address
    AdditionalInfo *AddressAdditionalInfo // Delivery point, carrier route, etc.
    Corrections    []AddressCorrection    // How to improve the address input
    Matches        []AddressMatch         // Exact match indicators
    Warnings       []string               // Any warnings
}
```

### Additional Address Information

The `AdditionalInfo` field provides valuable delivery information:

- `DeliveryPoint`: Unique identifier for every delivery address
- `CarrierRoute`: Carrier route code
- `DPVConfirmation`: Delivery Point Validation confirmation (Y, D, S, N)
- `DPVCMRA`: Commercial Mail Receiving Agency indicator
- `Business`: Business address indicator
- `CentralDeliveryPoint`: Central delivery point indicator
- `Vacant`: Vacancy indicator

## Testing

Run the test suite:

```bash
go test ./...
```

Run tests with coverage:

```bash
go test -cover ./...
```

## Requirements

- Go 1.19 or later
- USPS Client ID and Client Secret (obtain from [USPS Developer Portal](https://developers.usps.com))

## API Documentation

For complete API documentation, see:

- [USPS Addresses API Documentation](https://developers.usps.com/addressesv3)
- [USPS OAuth 2.0 API Documentation](https://developers.usps.com/oauth2v3)

## License

[MIT License](LICENSE)

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## Support

For USPS API support, contact [USPS API Support](https://emailus.usps.com/s/web-tools-inquiry).

For issues with this library, please [open an issue](https://github.com/my-eq/go-usps/issues) on GitHub
