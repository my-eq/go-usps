# go-usps

A lightweight, concise Golang client library for the USPS REST API, starting with Addresses API v3.

## Features

- ✅ USPS Addresses API v3 support
- ✅ Dependency injection for custom HTTP clients
- ✅ Default HTTP client with retry and exponential backoff
- ✅ Clean, DRY code structure
- ✅ Comprehensive test coverage
- ✅ GitHub Actions for automated linting

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
)

func main() {
    // Create a new USPS client with your API key
    client := usps.NewClient("your-api-key-here")

    // Create an address validation request
    addr := &usps.AddressRequest{
        StreetAddress: "123 Main St",
        City:          "Anytown",
        State:         "CA",
        ZIPCode:       "12345",
    }

    // Validate the address
    result, err := client.ValidateAddress(context.Background(), addr)
    if err != nil {
        log.Fatalf("Error: %v", err)
    }

    if result.Address != nil {
        fmt.Printf("Validated: %s, %s, %s %s\n",
            result.Address.StreetAddress,
            result.Address.City,
            result.Address.State,
            result.Address.ZIPCode)
    }
}
```

## Advanced Usage

### Custom HTTP Client

You can provide your own HTTP client implementation:

```go
type MyHTTPClient struct {
    // your custom implementation
}

func (m *MyHTTPClient) Do(req *http.Request) (*http.Response, error) {
    // custom logic
}

client := usps.NewClient("api-key", usps.WithHTTPClient(&MyHTTPClient{}))
```

### Custom Retry Configuration

Configure retry behavior for the default HTTP client:

```go
retryConfig := &usps.RetryConfig{
    MaxRetries:     5,
    InitialBackoff: 1 * time.Second,
    MaxBackoff:     10 * time.Second,
    Multiplier:     2.0,
}

httpClient := usps.NewDefaultHTTPClient(retryConfig)
client := usps.NewClient("api-key", usps.WithHTTPClient(httpClient))
```

### Default Retry Configuration

The default HTTP client includes:
- **MaxRetries**: 3
- **InitialBackoff**: 500ms
- **MaxBackoff**: 5s
- **Multiplier**: 2.0 (exponential backoff)

## API Documentation

### Client

```go
// Create a new client
client := usps.NewClient(apiKey string, opts ...ClientOption)

// Available options
usps.WithHTTPClient(client HTTPClient)
usps.WithBaseURL(url string)
```

### Address Validation

```go
type AddressRequest struct {
    StreetAddress    string
    SecondaryAddress string
    City             string
    State            string
    ZIPCode          string
    ZIPPlus4         string
    Urbanization     string
}

result, err := client.ValidateAddress(ctx context.Context, addr *AddressRequest)
```

## Testing

Run tests with:

```bash
go test -v ./...
```

Run tests with coverage:

```bash
go test -v -race -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Contributing

Contributions are welcome! Please ensure:
- Code passes `golangci-lint`
- Tests are included for new features
- Code follows Go best practices

## License

MIT License - see LICENSE file for details

## Resources

- [USPS Addresses API v3 Documentation](https://developers.usps.com/addressesv3)
