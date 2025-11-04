package parser

import "testing"

func TestTokenizer_Tokenize(t *testing.T) {
	tok := newTokenizer()

	tests := []struct {
		name       string
		input      string
		wantTokens int
		checkTypes []TokenType
	}{
		{
			name:       "simple address",
			input:      "123 Main St",
			wantTokens: 3,
			checkTypes: []TokenType{TokenHouseNumber, TokenStreetName, TokenStreetSuffix},
		},
		{
			name:       "address with directional",
			input:      "456 North Oak Ave",
			wantTokens: 4,
			checkTypes: []TokenType{TokenHouseNumber, TokenPreDirectional, TokenStreetName, TokenStreetSuffix},
		},
		{
			name:       "address with secondary",
			input:      "789 Elm St Apt 4B",
			wantTokens: 5,
			checkTypes: []TokenType{TokenHouseNumber, TokenStreetName, TokenStreetSuffix, TokenSecondaryDesignator, TokenSecondaryNumber},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens := tok.tokenize(tt.input)

			if len(tokens) != tt.wantTokens {
				t.Errorf("got %d tokens, want %d", len(tokens), tt.wantTokens)
				for i, tok := range tokens {
					t.Logf("  Token %d: Type=%v Value=%q", i, tok.Type, tok.Value)
				}
			}

			for i, wantType := range tt.checkTypes {
				if i >= len(tokens) {
					break
				}
				if tokens[i].Type != wantType {
					t.Errorf("token %d: got type %v, want %v", i, tokens[i].Type, wantType)
				}
			}
		})
	}
}

func TestNormalizeInput(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"123 Main St", "123 MAIN ST"},
		{"  123   Main   St  ", "123 MAIN ST"},
		{"123 Main St.", "123 MAIN ST"},
		{"123 Main St, New York", "123 MAIN ST NEW YORK"},
		{"123 main street", "123 MAIN STREET"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := normalizeInput(tt.input)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSplitAddressParts(t *testing.T) {
	tests := []struct {
		input     string
		wantParts int
	}{
		{"123 MAIN ST", 1},
		{"123 MAIN ST | NEW YORK", 2},
		{"123 MAIN ST | NEW YORK | NY", 3},
		{"123 MAIN ST | NEW YORK | NY | 10001", 4},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			parts := splitAddressParts(tt.input)
			if len(parts) != tt.wantParts {
				t.Errorf("got %d parts, want %d", len(parts), tt.wantParts)
				for i, part := range parts {
					t.Logf("  Part %d: %q", i, part)
				}
			}
		})
	}
}

func TestIsNumeric(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"123", true},
		{"456-789", true},
		{"abc", false},
		{"12a34", false},
		{"", false},
		{"0", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := isNumeric(tt.input)
			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsZIPCode(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"10001", true},
		{"10001-1234", true},
		{"100011234", true},
		{"1234", false},
		{"abcde", false},
		{"10001-", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := isZIPCode(tt.input)
			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsZIPPlus4(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"10001-1234", true},
		{"10001", false},
		{"100011234", false},
		{"10001-", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := isZIPPlus4(tt.input)
			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewTokenizer(t *testing.T) {
	tok := newTokenizer()

	if tok == nil {
		t.Fatal("newTokenizer() returned nil")
	}

	if tok.lexicon == nil {
		t.Error("lexicon is nil")
	}
}
