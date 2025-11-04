package parser

// Normalizer applies USPS standardization rules to tokens.
type Normalizer struct {
	lexicon *Lexicon
}

// newNormalizer creates a new Normalizer.
func newNormalizer() *Normalizer {
	return &Normalizer{
		lexicon: newLexicon(),
	}
}

// normalize processes tokens and applies standardization rules.
func (n *Normalizer) normalize(tokens []Token) ([]Token, []Diagnostic) {
	var normalized []Token
	var diagnostics []Diagnostic

	for i := 0; i < len(tokens); i++ {
		token := tokens[i]

		// Disambiguate directionals (pre vs post)
		if token.Type == TokenPreDirectional {
			// If followed by street suffix or another directional, it's a pre-directional
			// Otherwise, might be part of street name or post-directional
			if i+1 < len(tokens) {
				next := tokens[i+1]
				if next.Type == TokenStreetSuffix || next.Type == TokenPreDirectional {
					// It's a pre-directional, keep as is
				} else if next.Type == TokenStreetName {
					// Could be pre-directional before street name
				} else {
					// Might be post-directional
					token.Type = TokenPostDirectional
				}
			}
		}

		// Reclassify street name tokens in city/state context
		// This is a simple heuristic: tokens after street address are likely city
		if token.Type == TokenStreetName && i > 0 {
			// Check if we've seen a street suffix already
			hasStreetSuffix := false
			for j := 0; j < i; j++ {
				if tokens[j].Type == TokenStreetSuffix {
					hasStreetSuffix = true
					break
				}
			}

			// If we have a suffix and see more street names, they might be city
			if hasStreetSuffix {
				// Check if next token is a state
				if i+1 < len(tokens) && tokens[i+1].Type == TokenState {
					token.Type = TokenCity
				}
			}
		}

		normalized = append(normalized, token)
	}

	return normalized, diagnostics
}
