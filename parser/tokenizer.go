package parser

import (
	"regexp"
	"strings"
	"unicode"
)

// Tokenizer converts raw address input into classified tokens.
type Tokenizer struct {
	lexicon *Lexicon
}

// newTokenizer creates a new Tokenizer with initialized lexicon.
func newTokenizer() *Tokenizer {
	return &Tokenizer{
		lexicon: newLexicon(),
	}
}

// tokenize splits the input into tokens and classifies them.
func (t *Tokenizer) tokenize(input string) []Token {
	// Normalize input
	normalized := normalizeInput(input)

	// Split on common delimiters
	parts := splitAddressParts(normalized)

	var tokens []Token
	position := 0

	for _, part := range parts {
		if part == "" {
			continue
		}

		// Tokenize each part
		partTokens := t.tokenizePart(part, position)
		tokens = append(tokens, partTokens...)
		position += len(part) + 1 // +1 for delimiter
	}

	return tokens
}

// normalizeInput cleans and normalizes the input string.
func normalizeInput(input string) string {
	// Convert to uppercase for consistent matching
	s := strings.ToUpper(input)

	// Normalize whitespace
	s = strings.TrimSpace(s)
	s = regexp.MustCompile(`\s+`).ReplaceAllString(s, " ")

	// Remove excess punctuation but preserve necessary ones
	s = strings.ReplaceAll(s, ".", "")
	s = strings.ReplaceAll(s, ",", " ")
	s = strings.ReplaceAll(s, ";", " ")

	// Normalize whitespace again after punctuation removal
	s = regexp.MustCompile(`\s+`).ReplaceAllString(s, " ")
	s = strings.TrimSpace(s)

	return s
}

// splitAddressParts splits the address into logical parts using pipe delimiters.
// Note: The input has already been normalized by normalizeInput which converts commas to spaces.
func splitAddressParts(input string) []string {
	// Replace pipe delimiters with a special marker for splitting
	s := input
	s = strings.ReplaceAll(s, ",", " | ")

	// Split on the marker
	parts := strings.Split(s, "|")

	var result []string
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result
}

// tokenizePart tokenizes a single part of the address.
func (t *Tokenizer) tokenizePart(part string, basePosition int) []Token {
	words := strings.Fields(part)
	var tokens []Token
	position := basePosition

	for i := 0; i < len(words); i++ {
		word := words[i]
		original := word

		// Try to classify the token
		token := Token{
			Value:    word,
			Original: original,
			Start:    position,
			End:      position + len(word),
		}

		// Classification logic - check ZIP+4 first, then generic ZIP code, then numeric
		if isZIPPlus4(word) {
			// Split ZIP+4
			parts := strings.Split(word, "-")
			if len(parts) == 2 {
				// Add ZIP code token
				zipToken := Token{
					Type:     TokenZIPCode,
					Value:    parts[0],
					Original: parts[0],
					Start:    position,
					End:      position + len(parts[0]),
				}
				tokens = append(tokens, zipToken)

				// Add ZIP+4 token
				zip4Token := Token{
					Type:     TokenZIPPlus4,
					Value:    parts[1],
					Original: parts[1],
					Start:    position + len(parts[0]) + 1,
					End:      position + len(word),
				}
				tokens = append(tokens, zip4Token)
				position += len(word) + 1
				continue
			}
		} else if isZIPCode(word) {
			token.Type = TokenZIPCode
		} else if isNumeric(word) {
			// Check if previous token was a secondary designator
			if len(tokens) > 0 && tokens[len(tokens)-1].Type == TokenSecondaryDesignator {
				token.Type = TokenSecondaryNumber
			} else {
				token.Type = TokenHouseNumber
			}
		} else if normalized, ok := t.lexicon.NormalizeDirectional(word); ok {
			token.Type = TokenPreDirectional // May need to disambiguate later
			token.Value = normalized
		} else if normalized, ok := t.lexicon.NormalizeStreetSuffix(word); ok {
			token.Type = TokenStreetSuffix
			token.Value = normalized
		} else if normalized, ok := t.lexicon.NormalizeSecondaryDesignator(word); ok {
			token.Type = TokenSecondaryDesignator
			token.Value = normalized
		} else if normalized, ok := t.lexicon.NormalizeState(word); ok {
			token.Type = TokenState
			token.Value = normalized
		} else {
			// Check if it's alphanumeric (like "4B" for apartment)
			if len(tokens) > 0 && tokens[len(tokens)-1].Type == TokenSecondaryDesignator {
				token.Type = TokenSecondaryNumber
			} else {
				// Default to street name or city
				token.Type = TokenStreetName
			}
		}

		tokens = append(tokens, token)
		position += len(word) + 1 // +1 for space
	}

	return tokens
}

// isNumeric checks if a string is numeric.
func isNumeric(s string) bool {
	if len(s) == 0 {
		return false
	}
	for _, r := range s {
		if !unicode.IsDigit(r) && r != '-' {
			return false
		}
	}
	return true
}

// isZIPCode checks if a string looks like a ZIP code.
func isZIPCode(s string) bool {
	// 5-digit or 9-digit (with hyphen) or 10-digit (no hyphen)
	if len(s) == 5 {
		return isNumeric(s)
	}
	if len(s) == 10 && s[5] == '-' {
		return isNumeric(s[:5]) && isNumeric(s[6:])
	}
	if len(s) == 9 {
		return isNumeric(s)
	}
	return false
}

// isZIPPlus4 checks if a string is a ZIP+4 code.
func isZIPPlus4(s string) bool {
	return len(s) == 10 && s[5] == '-' && isNumeric(s[:5]) && isNumeric(s[6:])
}
