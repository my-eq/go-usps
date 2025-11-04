package parser

import (
	"strings"
	
	"github.com/my-eq/go-usps/models"
)

// TokenType represents the classification of a parsed token.
type TokenType int

const (
	// TokenUnknown represents an unclassified token.
	TokenUnknown TokenType = iota
	// TokenHouseNumber represents a numeric street address component.
	TokenHouseNumber
	// TokenPreDirectional represents a directional prefix (N, S, E, W, etc.).
	TokenPreDirectional
	// TokenStreetName represents the street name.
	TokenStreetName
	// TokenStreetSuffix represents the street type (ST, AVE, BLVD, etc.).
	TokenStreetSuffix
	// TokenPostDirectional represents a directional suffix.
	TokenPostDirectional
	// TokenSecondaryDesignator represents apartment, suite, unit designators.
	TokenSecondaryDesignator
	// TokenSecondaryNumber represents the secondary unit number.
	TokenSecondaryNumber
	// TokenCity represents the city name.
	TokenCity
	// TokenState represents the state code.
	TokenState
	// TokenZIPCode represents a ZIP code.
	TokenZIPCode
	// TokenZIPPlus4 represents a ZIP+4 extension.
	TokenZIPPlus4
	// TokenFirm represents a firm or business name.
	TokenFirm
)

// Token represents a classified lexeme from the input.
type Token struct {
	Type     TokenType
	Value    string
	Original string // Original input before normalization
	Start    int    // Position in original input
	End      int    // End position in original input
}

// Diagnostic represents a parsing issue with severity and context.
type Diagnostic struct {
	Severity    DiagnosticSeverity
	Message     string
	Start       int    // Position in input
	End         int    // End position in input
	Remediation string // Suggested fix
	Code        string // Machine-readable code
}

// DiagnosticSeverity represents the severity level of a diagnostic.
type DiagnosticSeverity int

const (
	// SeverityInfo represents informational messages.
	SeverityInfo DiagnosticSeverity = iota
	// SeverityWarning represents potential issues.
	SeverityWarning
	// SeverityError represents errors that prevent successful parsing.
	SeverityError
)

// String returns the string representation of the severity.
func (s DiagnosticSeverity) String() string {
	switch s {
	case SeverityInfo:
		return "Info"
	case SeverityWarning:
		return "Warning"
	case SeverityError:
		return "Error"
	default:
		return "Unknown"
	}
}

// ParsedAddress represents the result of parsing a free-form address.
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

// ToAddressRequest converts a ParsedAddress to a models.AddressRequest.
// The method combines parsed components into the format required by the USPS API,
// including building the street address from house number, directionals, street name,
// and suffix, and the secondary address from unit designator and number.
func (p *ParsedAddress) ToAddressRequest() *models.AddressRequest {
	req := &models.AddressRequest{
		State: p.State,
	}

	// Build street address
	var streetParts []string
	if p.HouseNumber != "" {
		streetParts = append(streetParts, p.HouseNumber)
	}
	if p.PreDirectional != "" {
		streetParts = append(streetParts, p.PreDirectional)
	}
	if p.StreetName != "" {
		streetParts = append(streetParts, p.StreetName)
	}
	if p.StreetSuffix != "" {
		streetParts = append(streetParts, p.StreetSuffix)
	}
	if p.PostDirectional != "" {
		streetParts = append(streetParts, p.PostDirectional)
	}
	
	if len(streetParts) > 0 {
		req.StreetAddress = joinTokens(streetParts)
	}

	// Build secondary address
	if p.SecondaryUnit != "" || p.SecondaryNumber != "" {
		var secondaryParts []string
		if p.SecondaryUnit != "" {
			secondaryParts = append(secondaryParts, p.SecondaryUnit)
		}
		if p.SecondaryNumber != "" {
			secondaryParts = append(secondaryParts, p.SecondaryNumber)
		}
		req.SecondaryAddress = joinTokens(secondaryParts)
	}

	if p.Firm != "" {
		req.Firm = p.Firm
	}
	if p.City != "" {
		req.City = p.City
	}
	if p.ZIPCode != "" {
		req.ZIPCode = p.ZIPCode
	}
	if p.ZIPPlus4 != "" {
		req.ZIPPlus4 = p.ZIPPlus4
	}

	return req
}

// joinTokens joins string parts with a single space.
func joinTokens(parts []string) string {
	if len(parts) == 0 {
		return ""
	}
	if len(parts) == 1 {
		return parts[0]
	}
	
	// Calculate total length to pre-allocate
	totalLen := len(parts) - 1 // number of spaces
	for _, part := range parts {
		totalLen += len(part)
	}
	
	var b strings.Builder
	b.Grow(totalLen)
	
	for i, part := range parts {
		if i > 0 {
			b.WriteString(" ")
		}
		b.WriteString(part)
	}
	return b.String()
}
