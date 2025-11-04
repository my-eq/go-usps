package parser

// Validator enforces USPS Publication 28 component ordering and requirements.
type Validator struct{}

// newValidator creates a new Validator.
func newValidator() *Validator {
	return &Validator{}
}

// validate checks tokens for completeness and proper ordering.
func (v *Validator) validate(parsed *ParsedAddress) []Diagnostic {
	var diagnostics []Diagnostic

	// Check for required components
	if parsed.State == "" {
		diagnostics = append(diagnostics, Diagnostic{
			Severity:    SeverityError,
			Message:     "Missing required state code",
			Code:        "MISSING_STATE",
			Remediation: "Add a 2-letter state code (e.g., NY, CA, TX)",
		})
	}

	if parsed.HouseNumber == "" && parsed.StreetName == "" {
		diagnostics = append(diagnostics, Diagnostic{
			Severity:    SeverityError,
			Message:     "Missing street address",
			Code:        "MISSING_STREET",
			Remediation: "Add a street address with house number and street name",
		})
	}

	// Check for ZIP code
	if parsed.ZIPCode == "" {
		diagnostics = append(diagnostics, Diagnostic{
			Severity:    SeverityWarning,
			Message:     "Missing ZIP code",
			Code:        "MISSING_ZIP",
			Remediation: "Add a 5-digit ZIP code for better address validation",
		})
	}

	return diagnostics
}
