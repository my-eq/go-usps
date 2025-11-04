package parser

// Parser coordinates the tokenization, normalization, validation, and formatting pipeline.
type Parser struct {
	tokenizer  *Tokenizer
	normalizer *Normalizer
	validator  *Validator
}

// New creates a new Parser with default configuration.
func New() *Parser {
	return &Parser{
		tokenizer:  newTokenizer(),
		normalizer: newNormalizer(),
		validator:  newValidator(),
	}
}

// Parse parses a free-form address string into a structured ParsedAddress.
// It tokenizes the input, applies USPS standardization rules, and validates
// the address components. Returns the parsed address and any diagnostics
// (warnings or errors) encountered during parsing.
func Parse(input string) (*ParsedAddress, []Diagnostic) {
	p := New()
	return p.Parse(input)
}

// Parse parses a free-form address string using this parser instance.
func (p *Parser) Parse(input string) (*ParsedAddress, []Diagnostic) {
	// Tokenize
	tokens := p.tokenizer.tokenize(input)

	// Normalize
	normalizedTokens, normDiagnostics := p.normalizer.normalize(tokens)

	// Build ParsedAddress
	parsed := p.buildParsedAddress(normalizedTokens, input)

	// Validate
	valDiagnostics := p.validator.validate(parsed)

	// Combine diagnostics
	diagnostics := append(normDiagnostics, valDiagnostics...)

	return parsed, diagnostics
}

// buildParsedAddress constructs a ParsedAddress from normalized tokens.
func (p *Parser) buildParsedAddress(tokens []Token, originalInput string) *ParsedAddress {
	addr := &ParsedAddress{
		Tokens:        tokens,
		OriginalInput: originalInput,
	}

	// Track what we've seen to handle ordering
	var streetNameParts []string
	var cityParts []string
	seenStreetSuffix := false
	seenSecondaryDesignator := false
	seenState := false

	// Find state index to help identify city
	stateIndex := -1
	for i, token := range tokens {
		if token.Type == TokenState {
			stateIndex = i
			break
		}
	}

	for i, token := range tokens {
		switch token.Type {
		case TokenHouseNumber:
			// If we've seen a state, this is probably a ZIP code
			if seenState && addr.HouseNumber != "" {
				// Treat as ZIP code if it's 5 or 9 digits
				if len(token.Value) == 5 || len(token.Value) == 9 {
					if addr.ZIPCode == "" {
						addr.ZIPCode = token.Value
					}
				}
			} else if addr.HouseNumber == "" {
				addr.HouseNumber = token.Value
			}
		case TokenPreDirectional:
			// If we haven't seen the street suffix yet, this is a pre-directional
			if !seenStreetSuffix && addr.PreDirectional == "" {
				addr.PreDirectional = token.Value
			} else if seenStreetSuffix && addr.PostDirectional == "" {
				// After street suffix, directionals become post-directionals
				addr.PostDirectional = token.Value
			}
		case TokenStreetName:
			// If we have a state and this token is right before it, it's city
			if stateIndex >= 0 && i == stateIndex-1 {
				cityParts = append(cityParts, token.Value)
			} else if !seenStreetSuffix && !seenSecondaryDesignator {
				// Before street suffix or secondary designator = street name
				streetNameParts = append(streetNameParts, token.Value)
			} else if seenStreetSuffix || seenSecondaryDesignator {
				// After street components = city
				cityParts = append(cityParts, token.Value)
			}
		case TokenStreetSuffix:
			addr.StreetSuffix = token.Value
			seenStreetSuffix = true
		case TokenPostDirectional:
			if addr.PostDirectional == "" {
				addr.PostDirectional = token.Value
			}
		case TokenSecondaryDesignator:
			if addr.SecondaryUnit == "" {
				addr.SecondaryUnit = token.Value
			}
			seenSecondaryDesignator = true
		case TokenSecondaryNumber:
			if addr.SecondaryNumber == "" {
				addr.SecondaryNumber = token.Value
			}
		case TokenCity:
			cityParts = append(cityParts, token.Value)
		case TokenState:
			if addr.State == "" {
				addr.State = token.Value
			}
			seenState = true
		case TokenZIPCode:
			if addr.ZIPCode == "" {
				addr.ZIPCode = token.Value
			}
		case TokenZIPPlus4:
			if addr.ZIPPlus4 == "" {
				addr.ZIPPlus4 = token.Value
			}
		case TokenFirm:
			if addr.Firm == "" {
				addr.Firm = token.Value
			}
		}
	}

	// Join street name parts
	if len(streetNameParts) > 0 {
		addr.StreetName = joinTokens(streetNameParts)
	}

	// Join city parts
	if len(cityParts) > 0 {
		addr.City = joinTokens(cityParts)
	}

	return addr
}
