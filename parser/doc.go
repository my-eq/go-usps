// Package parser provides comprehensive parsing for free-form address entry.
//
// It normalizes free-form addresses into structured fields required by models.AddressRequest
// while preserving context for corrective feedback. The package follows a finite state machine
// approach with a pipeline architecture consisting of:
//
//   - Tokenizer: Classifies lexemes using USPS Publication 28 lookup tables
//   - Normalizer: Applies USPS abbreviation rules and standardization
//   - Validator: Enforces component ordering and presence requirements
//   - Formatter: Converts to models.AddressRequest format
//
// Example usage:
//
//	input := "123 North Main Street Apartment 4B, New York, NY 10001"
//	result, diagnostics := parser.Parse(input)
//	if len(diagnostics) > 0 {
//	    // Handle warnings or errors
//	    for _, d := range diagnostics {
//	        fmt.Printf("%s: %s\n", d.Severity, d.Message)
//	    }
//	}
//	req := result.ToAddressRequest()
//
// The parser is designed to be extensible and follows idiomatic Go patterns with
// strong typing and zero dependencies beyond the Go standard library.
package parser
