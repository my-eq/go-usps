# Address Parser

Comprehensive free-form address parsing for USPS address standardization.

## Overview

The parser package provides intelligent parsing of free-form address strings into structured
`models.AddressRequest` objects suitable for use with the USPS Addresses API. It implements
a finite state machine (FSM) architecture based on USPS Publication 28 standards.

## Features

- **USPS Publication 28 Compliant** - Follows official USPS addressing standards
- **Intelligent Tokenization** - Recognizes address components using lookup tables
- **Standardization** - Automatically applies USPS abbreviations (Street → ST, Avenue → AVE)
- **Diagnostics** - Provides warnings and errors with remediation suggestions
- **Zero Dependencies** - Uses only Go standard library
- **High Coverage** - 94.8% test coverage

## Architecture

The parser follows a pipeline pattern:

```text
┌─────────────┐    ┌──────────────┐    ┌──────────────┐    ┌──────────────┐
│  Tokenizer  │ -> │ Normalizer   │ -> │ Validator    │ -> │ Formatter    │
└─────────────┘    └──────────────┘    └──────────────┘    └──────────────┘
```

1. **Tokenizer** - Classifies lexemes (house numbers, street names, states, ZIP codes)
2. **Normalizer** - Applies USPS standard abbreviations
3. **Validator** - Checks for required components and proper formatting
4. **Formatter** - Converts to `models.AddressRequest`

## Usage

### Basic Parsing

```go
import "github.com/my-eq/go-usps/parser"

input := "123 North Main Street Apartment 4B, New York, NY 10001-1234"
parsed, diagnostics := parser.Parse(input)

// Check for issues
for _, d := range diagnostics {
    fmt.Printf("%s: %s\n", d.Severity, d.Message)
}

// Convert to AddressRequest
req := parsed.ToAddressRequest()
fmt.Printf("Street: %s\n", req.StreetAddress)     // "123 N MAIN ST"
fmt.Printf("Secondary: %s\n", req.SecondaryAddress) // "APT 4B"
fmt.Printf("City: %s\n", req.City)                 // "NEW YORK"
fmt.Printf("State: %s\n", req.State)               // "NY"
fmt.Printf("ZIP: %s\n", req.ZIPCode)               // "10001"
fmt.Printf("ZIP+4: %s\n", req.ZIPPlus4)            // "1234"
```

### With USPS Validation

```go
import (
    "context"
    "github.com/my-eq/go-usps"
    "github.com/my-eq/go-usps/parser"
)

// Parse user input
input := "456 Oak Avenue, Los Angeles, CA 90001"
parsed, diagnostics := parser.Parse(input)

// Handle diagnostics
for _, d := range diagnostics {
    if d.Severity == parser.SeverityError {
        log.Printf("Error: %s - %s", d.Message, d.Remediation)
    }
}

// Validate with USPS API
client := usps.NewClientWithOAuth("client-id", "client-secret")
req := parsed.ToAddressRequest()
resp, err := client.GetAddress(context.Background(), req)
if err != nil {
    log.Fatalf("Validation failed: %v", err)
}

fmt.Printf("Validated: %s\n", resp.Address.StreetAddress)
```

## Supported Address Formats

The parser handles a wide variety of address formats:

### Residential Addresses

```go
parser.Parse("123 Main St, New York, NY 10001")
parser.Parse("456 North Oak Avenue, Los Angeles, CA 90001")
parser.Parse("789 Elm Blvd, Chicago, IL 60601")
```

### With Secondary Units

```go
parser.Parse("123 Main St Apt 4B, New York, NY 10001")
parser.Parse("456 Oak Ave Suite 200, Boston, MA 02101")
parser.Parse("789 Elm St Unit 12, Seattle, WA 98101")
```

### With ZIP+4

```go
parser.Parse("123 Main St, New York, NY 10001-1234")
parser.Parse("456 Oak Ave, Boston, MA 02101-5678")
```

### With Directionals

```go
parser.Parse("123 North Main Street, New York, NY 10001")
parser.Parse("456 E Oak Avenue, Boston, MA 02101")
```

## Standardization

The parser automatically applies USPS standard abbreviations:

| Input           | Standardized |
|----------------|--------------|
| Street         | ST           |
| Avenue         | AVE          |
| Boulevard      | BLVD         |
| Drive          | DR           |
| Lane           | LN           |
| North          | N            |
| South          | S            |
| East           | E            |
| West           | W            |
| Apartment      | APT          |
| Suite          | STE          |
| Unit           | UNIT         |

See [USPS Publication 28](https://pe.usps.com/archive/pdf/DMMArchive20050109/pub28.pdf)
for complete standards.

## Diagnostics

The parser provides rich diagnostics with severity levels:

```go
type DiagnosticSeverity int

const (
    SeverityInfo    // Informational messages
    SeverityWarning // Potential issues
    SeverityError   // Errors preventing parsing
)
```

Each diagnostic includes:

- **Severity** - Info, Warning, or Error
- **Message** - Human-readable description
- **Code** - Machine-readable identifier
- **Remediation** - Suggested fix
- **Start/End** - Position in original input

### Example Diagnostics

```go
parsed, diagnostics := parser.Parse("123 Main St, New York")

for _, d := range diagnostics {
    fmt.Printf("%s: %s\n", d.Severity, d.Message)
    fmt.Printf("Fix: %s\n", d.Remediation)
}

// Output:
// Error: Missing required state code
// Fix: Add a 2-letter state code (e.g., NY, CA, TX)
// Warning: Missing ZIP code
// Fix: Add a 5-digit ZIP code for better address validation
```

## API Reference

### Functions

#### Parse

```go
func Parse(input string) (*ParsedAddress, []Diagnostic)
```

Parses a free-form address string and returns the structured address and diagnostics.

### Types

#### ParsedAddress

```go
type ParsedAddress struct {
    Firm             string
    HouseNumber      string
    PreDirectional   string
    StreetName       string
    StreetSuffix     string
    PostDirectional  string
    SecondaryUnit    string
    SecondaryNumber  string
    City             string
    State            string
    ZIPCode          string
    ZIPPlus4         string
    Tokens           []Token
    OriginalInput    string
}
```

#### Methods

```go
func (p *ParsedAddress) ToAddressRequest() *models.AddressRequest
```

Converts the parsed address to a `models.AddressRequest` for use with the USPS API.

#### Diagnostic

```go
type Diagnostic struct {
    Severity    DiagnosticSeverity
    Message     string
    Start       int
    End         int
    Remediation string
    Code        string
}
```

## Design Principles

### 1. Simplicity

Following the Unix philosophy: "Perfection is attained not when there is nothing more to add,
but when there is none left to remove."

### 2. Strong Typing

No maps or `interface{}` - all structures are strongly typed for compile-time safety.

### 3. Low Coupling

Components are separated with clear interfaces and single responsibilities.

### 4. Idiomatic Go

Follows Go best practices and conventions throughout.

## Testing

The parser has comprehensive test coverage (94.8%):

```bash
# Run all tests
go test ./parser/...

# With coverage
go test -cover ./parser/...

# Verbose output
go test -v ./parser/...
```

## References

- [USPS Publication 28: Postal Addressing Standards](https://pe.usps.com/archive/pdf/DMMArchive20050109/pub28.pdf)
- [USPS Addressing Standards (DMM 602)](https://pe.usps.com/text/dmm300/602.htm)
- [USPS AIS Product Technical Guides](https://postalpro.usps.com/address-quality/ais-products)

## License

MIT License - See [LICENSE](../LICENSE) file for details.
