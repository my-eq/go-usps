package parser

import (
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
	// Normalize input while tracking original positions
	normalized, positionMap := normalizeInputWithMapping(input)

	// Split on common delimiters
	parts := splitAddressParts(normalized)

	var tokens []Token
	position := 0

	for _, part := range parts {
		if part == "" {
			continue
		}

		// Tokenize each part
		partTokens := t.tokenizePart(part, position, positionMap)
		tokens = append(tokens, partTokens...)
		position += len(part) + 1 // +1 for delimiter
	}

	return tokens
}

// normalizeInputWithMapping cleans and normalizes the input string while maintaining
// a mapping from normalized positions back to original positions.
func normalizeInputWithMapping(input string) (string, []int) {
	var result strings.Builder
	positionMap := make([]int, 0, len(input))
	
	s := input
	origPos := 0
	
	// Convert to uppercase and build position map
	for i, r := range s {
		upper := unicode.ToUpper(r)
		
		// Skip punctuation that should be removed
		if r == '.' || r == ',' || r == ';' {
			continue
		}
		
		// Map whitespace to single space
		if unicode.IsSpace(r) {
			// Only add space if the last char wasn't a space
			if result.Len() == 0 || result.String()[result.Len()-1] != ' ' {
				result.WriteRune(' ')
				positionMap = append(positionMap, i)
			}
		} else {
			result.WriteRune(upper)
			positionMap = append(positionMap, i)
		}
		origPos = i
	}
	
	// Trim trailing spaces
	normalized := strings.TrimSpace(result.String())
	
	// Adjust position map for trimming
	if len(normalized) < result.Len() {
		positionMap = positionMap[:len(normalized)]
	}
	
	// Handle leading trim
	trimStart := len(result.String()) - len(strings.TrimLeft(result.String(), " "))
	if trimStart > 0 && trimStart < len(positionMap) {
		positionMap = positionMap[trimStart:]
	}
	
	return normalized, positionMap
}

// normalizeInput cleans and normalizes the input string.
// This is kept for backward compatibility but should use normalizeInputWithMapping.
func normalizeInput(input string) string {
	normalized, _ := normalizeInputWithMapping(input)
	return normalized
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
func (t *Tokenizer) tokenizePart(part string, basePosition int, positionMap []int) []Token {
	words := strings.Fields(part)
	var tokens []Token
	position := basePosition

	for i := 0; i < len(words); i++ {
		word := words[i]
		original := word
		
		// Calculate original positions using the position map
		var startPos, endPos int
		if position < len(positionMap) {
			startPos = positionMap[position]
			endIdx := position + len(word) - 1
			if endIdx < len(positionMap) {
				endPos = positionMap[endIdx] + 1 // +1 to make it exclusive
			} else if len(positionMap) > 0 {
				endPos = positionMap[len(positionMap)-1] + 1
			} else {
				endPos = startPos + len(word)
			}
		} else {
			// Fallback if position map is incomplete
			startPos = position
			endPos = position + len(word)
		}

		// Try to classify the token
		token := Token{
			Value:    word,
			Original: original,
			Start:    startPos,
			End:      endPos,
		}

		// Classification logic - check ZIP+4 first, then generic ZIP code, then numeric
		if isZIPPlus4(word) {
			// Split ZIP+4
			parts := strings.Split(word, "-")
			if len(parts) == 2 {
				// Calculate positions for ZIP code part
				zipStartPos := startPos
				zipLen := len(parts[0])
				var zipEndPos int
				if position+zipLen-1 < len(positionMap) {
					zipEndPos = positionMap[position+zipLen-1] + 1
				} else {
					zipEndPos = zipStartPos + zipLen
				}
				
				// Add ZIP code token
				zipToken := Token{
					Type:     TokenZIPCode,
					Value:    parts[0],
					Original: parts[0],
					Start:    zipStartPos,
					End:      zipEndPos,
				}
				tokens = append(tokens, zipToken)

				// Calculate positions for ZIP+4 part (after hyphen)
				zip4Start := position + zipLen + 1 // +1 for hyphen
				var zip4StartPos, zip4EndPos int
				if zip4Start < len(positionMap) {
					zip4StartPos = positionMap[zip4Start]
					zip4EndIdx := zip4Start + len(parts[1]) - 1
					if zip4EndIdx < len(positionMap) {
						zip4EndPos = positionMap[zip4EndIdx] + 1
					} else {
						zip4EndPos = zip4StartPos + len(parts[1])
					}
				} else {
					zip4StartPos = zipEndPos + 1
					zip4EndPos = zip4StartPos + len(parts[1])
				}
				
				// Add ZIP+4 token
				zip4Token := Token{
					Type:     TokenZIPPlus4,
					Value:    parts[1],
					Original: parts[1],
					Start:    zip4StartPos,
					End:      zip4EndPos,
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
