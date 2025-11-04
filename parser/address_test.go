package parser

import (
	"testing"
)

// assertParsed is a helper to reduce repetitive field comparison noise in focused tests.
// It fails fast with clear labels for any mismatched component.
func assertParsed(t *testing.T, p ParsedAddress, street, secondary, city, state, zip, zip4 string) {
	if p.StreetAddress != street {
		t.Fatalf("street: want %q, got %q", street, p.StreetAddress)
	}
	if p.SecondaryAddress != secondary {
		t.Fatalf("secondary: want %q, got %q", secondary, p.SecondaryAddress)
	}
	if p.City != city {
		t.Fatalf("city: want %q, got %q", city, p.City)
	}
	if p.State != state {
		t.Fatalf("state: want %q, got %q", state, p.State)
	}
	if p.ZIPCode != zip {
		t.Fatalf("ZIP: want %q, got %q", zip, p.ZIPCode)
	}
	if p.ZIPPlus4 != zip4 {
		t.Fatalf("ZIP+4: want %q, got %q", zip4, p.ZIPPlus4)
	}
}

// TestParseAddress_TableDriven exercises broad end-to-end parsing behaviors. Each test row targets
// a unique normalization or diagnostic scenario (directionals, suffixes, secondary units, whitespace
// collapse, PO BOX, territories, malformed region data, etc.). Algorithmic edge cases are handled
// in focused tests below to avoid duplication here.
func TestParseAddress_TableDriven(t *testing.T) {
	type diagExpect struct {
		Code DiagnosticCode
	}
	tests := []struct {
		name            string
		input           string
		wantStreet      string
		wantSecondary   string
		wantCity        string
		wantState       string
		wantZIP         string
		wantZIPPlus4    string
		wantDiagnostics []diagExpect
	}{
		{
			name:            "Basic street address",
			input:           "123 Main Street, Springfield, IL 62704",
			wantStreet:      "123 MAIN ST",
			wantSecondary:   "",
			wantCity:        "SPRINGFIELD",
			wantState:       "IL",
			wantZIP:         "62704",
			wantZIPPlus4:    "",
			wantDiagnostics: nil,
		},
		{
			name:            "With secondary unit",
			input:           "456 Elm St Apt 5B, Chicago, IL 60614-1234",
			wantStreet:      "456 ELM ST",
			wantSecondary:   "APT 5B",
			wantCity:        "CHICAGO",
			wantState:       "IL",
			wantZIP:         "60614",
			wantZIPPlus4:    "1234",
			wantDiagnostics: nil,
		},
		{
			name:            "PO Box",
			input:           "PO Box 123, Anytown, NY 12345",
			wantStreet:      "PO BOX 123",
			wantSecondary:   "",
			wantCity:        "ANYTOWN",
			wantState:       "NY",
			wantZIP:         "12345",
			wantZIPPlus4:    "",
			wantDiagnostics: nil,
		},
		{
			name:            "Missing state produces diagnostic",
			input:           "123 Main Street, Springfield",
			wantStreet:      "123 MAIN ST",
			wantSecondary:   "",
			wantCity:        "SPRINGFIELD",
			wantState:       "",
			wantZIP:         "",
			wantZIPPlus4:    "",
			wantDiagnostics: []diagExpect{{Code: "insufficient_segments"}, {Code: "missing_state_zip"}},
		},
		// (Whitespace / empty input scenarios covered in TestEmptyInput)
		// (Invalid state code variants covered in TestInvalidStateCodes)
		{
			name:            "Lowercase state code",
			input:           "123 Main St, Springfield, il 62704",
			wantStreet:      "123 MAIN ST",
			wantSecondary:   "",
			wantCity:        "SPRINGFIELD",
			wantState:       "IL",
			wantZIP:         "62704",
			wantZIPPlus4:    "",
			wantDiagnostics: nil,
		},
		// (Malformed ZIP formats covered in TestMalformedZIPCodes)
		{
			name:            "ZIP+4 with dash",
			input:           "123 Main St, Springfield, IL 62704-1234",
			wantStreet:      "123 MAIN ST",
			wantSecondary:   "",
			wantCity:        "SPRINGFIELD",
			wantState:       "IL",
			wantZIP:         "62704",
			wantZIPPlus4:    "1234",
			wantDiagnostics: nil,
		},
		// ZIP+4 space variant exercised elsewhere
		// Directional normalization (single representative plus special mid-street case; other
		// directional variants appear later in table
		{
			name:            "Direction in middle of street",
			input:           "200 Oak East Trail, Madison, WI 53703",
			wantStreet:      "200 OAK E TRL",
			wantSecondary:   "",
			wantCity:        "MADISON",
			wantState:       "WI",
			wantZIP:         "53703",
			wantZIPPlus4:    "",
			wantDiagnostics: nil,
		},
		{
			name:            "Directional converts at street start",
			input:           "West End Ave, New York, NY 10023",
			wantStreet:      "W END AVE",
			wantSecondary:   "",
			wantCity:        "NEW YORK",
			wantState:       "NY",
			wantZIP:         "10023",
			wantZIPPlus4:    "",
			wantDiagnostics: nil,
		},
		{
			name:            "Directional with numeric ordinal",
			input:           "123 East 7th Street, St Louis, MO 63101",
			wantStreet:      "123 E 7TH ST",
			wantSecondary:   "",
			wantCity:        "ST LOUIS",
			wantState:       "MO",
			wantZIP:         "63101",
			wantZIPPlus4:    "",
			wantDiagnostics: nil,
		},
		// Street suffix abbreviations covered later
		// Rural route scenarios covered in focused TestRuralRoutes
		// Multiple chained secondary designators covered later in table
		{
			name:            "Secondary segment separated by comma",
			input:           "123 Main St, Apt 4B, Springfield, IL 62704",
			wantStreet:      "123 MAIN ST",
			wantSecondary:   "APT 4B",
			wantCity:        "SPRINGFIELD",
			wantState:       "IL",
			wantZIP:         "62704",
			wantZIPPlus4:    "",
			wantDiagnostics: nil,
		},
		{
			name:            "Unknown secondary designator segment",
			input:           "123 Main St, Wing 5, Springfield, IL 62704",
			wantStreet:      "123 MAIN ST",
			wantSecondary:   "WING 5",
			wantCity:        "SPRINGFIELD",
			wantState:       "IL",
			wantZIP:         "62704",
			wantZIPPlus4:    "",
			wantDiagnostics: []diagExpect{{Code: DiagnosticCodeUnknownSecondary}},
		},
		{
			name:            "Secondary spans street and segment",
			input:           "101 Main St Unit 1, Unit 2, Example City, CA 90210",
			wantStreet:      "101 MAIN ST",
			wantSecondary:   "UNIT 1 UNIT 2",
			wantCity:        "EXAMPLE CITY",
			wantState:       "CA",
			wantZIP:         "90210",
			wantZIPPlus4:    "",
			wantDiagnostics: nil,
		},
		{
			name:            "Secondary keyword appears earlier in street",
			input:           "101 Unit Hill Unit 5, Sampletown, TX 75001",
			wantStreet:      "101 UNIT HILL",
			wantSecondary:   "UNIT 5",
			wantCity:        "SAMPLETOWN",
			wantState:       "TX",
			wantZIP:         "75001",
			wantZIPPlus4:    "",
			wantDiagnostics: nil,
		},
		// Edge cases - Mixed case input
		{
			name:            "Mixed case address",
			input:           "123 MaIn StReEt, SpRiNgFiElD, iL 62704",
			wantStreet:      "123 MAIN ST",
			wantSecondary:   "",
			wantCity:        "SPRINGFIELD",
			wantState:       "IL",
			wantZIP:         "62704",
			wantZIPPlus4:    "",
			wantDiagnostics: nil,
		},
		// Edge cases - Extra whitespace
		{
			name:            "Extra whitespace",
			input:           "  123   Main   Street  ,  Springfield  ,  IL   62704  ",
			wantStreet:      "123 MAIN ST",
			wantSecondary:   "",
			wantCity:        "SPRINGFIELD",
			wantState:       "IL",
			wantZIP:         "62704",
			wantZIPPlus4:    "",
			wantDiagnostics: nil,
		},
		// Edge cases - PO Box variations
		{
			name:            "PO Box with space variations",
			input:           "P O Box 456, Anytown, NY 12345",
			wantStreet:      "PO BOX 456",
			wantSecondary:   "",
			wantCity:        "ANYTOWN",
			wantState:       "NY",
			wantZIP:         "12345",
			wantZIPPlus4:    "",
			wantDiagnostics: nil,
		},
		{
			name:            "PO Box with alphanumeric box number",
			input:           "PO Box 123A, Anytown, NY 12345",
			wantStreet:      "PO BOX 123A",
			wantSecondary:   "",
			wantCity:        "ANYTOWN",
			wantState:       "NY",
			wantZIP:         "12345",
			wantZIPPlus4:    "",
			wantDiagnostics: nil,
		},
		// Numeric street name scenario covered later
		// Building/Floor/Room designators covered later
		// Empty input handling
		// US Territories
		{
			name:            "Puerto Rico address",
			input:           "123 Calle Principal, San Juan, PR 00901",
			wantStreet:      "123 CALLE PRINCIPAL",
			wantSecondary:   "",
			wantCity:        "SAN JUAN",
			wantState:       "PR",
			wantZIP:         "00901",
			wantZIPPlus4:    "",
			wantDiagnostics: nil,
		},
		{
			name:            "Virgin Islands address",
			input:           "456 Main St, Charlotte Amalie, VI 00802",
			wantStreet:      "456 MAIN ST",
			wantSecondary:   "",
			wantCity:        "CHARLOTTE AMALIE",
			wantState:       "VI",
			wantZIP:         "00802",
			wantZIPPlus4:    "",
			wantDiagnostics: nil,
		},
		// (Additional malformed ZIP length variant retained in focused ZIP tests)
		{
			name:            "Directional prefix - North",
			input:           "123 North Main Street, Springfield, IL 62704",
			wantStreet:      "123 N MAIN ST",
			wantSecondary:   "",
			wantCity:        "SPRINGFIELD",
			wantState:       "IL",
			wantZIP:         "62704",
			wantZIPPlus4:    "",
			wantDiagnostics: nil,
		},
		{
			name:            "Directional prefix - South",
			input:           "456 South Elm Avenue, Chicago, IL 60614",
			wantStreet:      "456 S ELM AVE",
			wantSecondary:   "",
			wantCity:        "CHICAGO",
			wantState:       "IL",
			wantZIP:         "60614",
			wantZIPPlus4:    "",
			wantDiagnostics: nil,
		},
		{
			name:            "Directional prefix - East",
			input:           "789 East Oak Boulevard, Boston, MA 02101",
			wantStreet:      "789 E OAK BLVD",
			wantSecondary:   "",
			wantCity:        "BOSTON",
			wantState:       "MA",
			wantZIP:         "02101",
			wantZIPPlus4:    "",
			wantDiagnostics: nil,
		},
		{
			name:            "Directional prefix - West",
			input:           "321 West Pine Drive, Seattle, WA 98101",
			wantStreet:      "321 W PINE DR",
			wantSecondary:   "",
			wantCity:        "SEATTLE",
			wantState:       "WA",
			wantZIP:         "98101",
			wantZIPPlus4:    "",
			wantDiagnostics: nil,
		},
		{
			name:            "Directional prefix - Northeast",
			input:           "100 Northeast Broadway, Portland, OR 97201",
			wantStreet:      "100 NE BROADWAY",
			wantSecondary:   "",
			wantCity:        "PORTLAND",
			wantState:       "OR",
			wantZIP:         "97201",
			wantZIPPlus4:    "",
			wantDiagnostics: nil,
		},
		{
			name:            "Directional prefix - Northwest",
			input:           "200 Northwest Main Street, Portland, OR 97209",
			wantStreet:      "200 NW MAIN ST",
			wantSecondary:   "",
			wantCity:        "PORTLAND",
			wantState:       "OR",
			wantZIP:         "97209",
			wantZIPPlus4:    "",
			wantDiagnostics: nil,
		},
		{
			name:            "Directional prefix - Southeast",
			input:           "300 Southeast Park Place, Atlanta, GA 30303",
			wantStreet:      "300 SE PARK PL",
			wantSecondary:   "",
			wantCity:        "ATLANTA",
			wantState:       "GA",
			wantZIP:         "30303",
			wantZIPPlus4:    "",
			wantDiagnostics: nil,
		},
		{
			name:            "Directional prefix - Southwest",
			input:           "400 Southwest Terrace, Miami, FL 33101",
			wantStreet:      "400 SW TER",
			wantSecondary:   "",
			wantCity:        "MIAMI",
			wantState:       "FL",
			wantZIP:         "33101",
			wantZIPPlus4:    "",
			wantDiagnostics: nil,
		},
		{
			name:            "Street suffix - Alley",
			input:           "123 Main Alley, Springfield, IL 62704",
			wantStreet:      "123 MAIN ALY",
			wantSecondary:   "",
			wantCity:        "SPRINGFIELD",
			wantState:       "IL",
			wantZIP:         "62704",
			wantZIPPlus4:    "",
			wantDiagnostics: nil,
		},
		{
			name:            "Street suffix - Avenue",
			input:           "456 Elm Avenue, Chicago, IL 60614",
			wantStreet:      "456 ELM AVE",
			wantSecondary:   "",
			wantCity:        "CHICAGO",
			wantState:       "IL",
			wantZIP:         "60614",
			wantZIPPlus4:    "",
			wantDiagnostics: nil,
		},
		{
			name:            "Street suffix - Boulevard",
			input:           "789 Oak Boulevard, Boston, MA 02101",
			wantStreet:      "789 OAK BLVD",
			wantSecondary:   "",
			wantCity:        "BOSTON",
			wantState:       "MA",
			wantZIP:         "02101",
			wantZIPPlus4:    "",
			wantDiagnostics: nil,
		},
		{
			name:            "Street suffix - Circle",
			input:           "100 Park Circle, Denver, CO 80201",
			wantStreet:      "100 PARK CIR",
			wantSecondary:   "",
			wantCity:        "DENVER",
			wantState:       "CO",
			wantZIP:         "80201",
			wantZIPPlus4:    "",
			wantDiagnostics: nil,
		},
		{
			name:            "Street suffix - Court",
			input:           "200 Maple Court, Austin, TX 78701",
			wantStreet:      "200 MAPLE CT",
			wantSecondary:   "",
			wantCity:        "AUSTIN",
			wantState:       "TX",
			wantZIP:         "78701",
			wantZIPPlus4:    "",
			wantDiagnostics: nil,
		},
		{
			name:            "Street suffix - Drive",
			input:           "300 River Drive, Nashville, TN 37201",
			wantStreet:      "300 RIVER DR",
			wantSecondary:   "",
			wantCity:        "NASHVILLE",
			wantState:       "TN",
			wantZIP:         "37201",
			wantZIPPlus4:    "",
			wantDiagnostics: nil,
		},
		{
			name:            "Street suffix - Lane",
			input:           "400 Cherry Lane, Phoenix, AZ 85001",
			wantStreet:      "400 CHERRY LN",
			wantSecondary:   "",
			wantCity:        "PHOENIX",
			wantState:       "AZ",
			wantZIP:         "85001",
			wantZIPPlus4:    "",
			wantDiagnostics: nil,
		},
		{
			name:            "Street suffix - Parkway",
			input:           "500 Valley Parkway, San Diego, CA 92101",
			wantStreet:      "500 VALLEY PKWY",
			wantSecondary:   "",
			wantCity:        "SAN DIEGO",
			wantState:       "CA",
			wantZIP:         "92101",
			wantZIPPlus4:    "",
			wantDiagnostics: nil,
		},
		{
			name:            "Street suffix - Place",
			input:           "600 Garden Place, Portland, OR 97201",
			wantStreet:      "600 GARDEN PL",
			wantSecondary:   "",
			wantCity:        "PORTLAND",
			wantState:       "OR",
			wantZIP:         "97201",
			wantZIPPlus4:    "",
			wantDiagnostics: nil,
		},
		{
			name:            "Street suffix - Road",
			input:           "700 Mountain Road, Denver, CO 80202",
			wantStreet:      "700 MOUNTAIN RD",
			wantSecondary:   "",
			wantCity:        "DENVER",
			wantState:       "CO",
			wantZIP:         "80202",
			wantZIPPlus4:    "",
			wantDiagnostics: nil,
		},
		{
			name:            "Street suffix - Square",
			input:           "800 Market Square, Philadelphia, PA 19101",
			wantStreet:      "800 MARKET SQ",
			wantSecondary:   "",
			wantCity:        "PHILADELPHIA",
			wantState:       "PA",
			wantZIP:         "19101",
			wantZIPPlus4:    "",
			wantDiagnostics: nil,
		},
		{
			name:            "Street suffix - Terrace",
			input:           "900 Hill Terrace, Seattle, WA 98102",
			wantStreet:      "900 HILL TER",
			wantSecondary:   "",
			wantCity:        "SEATTLE",
			wantState:       "WA",
			wantZIP:         "98102",
			wantZIPPlus4:    "",
			wantDiagnostics: nil,
		},
		{
			name:            "Street suffix - Trail",
			input:           "1000 Forest Trail, Minneapolis, MN 55401",
			wantStreet:      "1000 FOREST TRL",
			wantSecondary:   "",
			wantCity:        "MINNEAPOLIS",
			wantState:       "MN",
			wantZIP:         "55401",
			wantZIPPlus4:    "",
			wantDiagnostics: nil,
		},
		{
			name:            "Street suffix - Way",
			input:           "1100 Ocean Way, Miami, FL 33101",
			wantStreet:      "1100 OCEAN WAY",
			wantSecondary:   "",
			wantCity:        "MIAMI",
			wantState:       "FL",
			wantZIP:         "33101",
			wantZIPPlus4:    "",
			wantDiagnostics: nil,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			parsed := Parse(tc.input)
			if parsed.StreetAddress != tc.wantStreet {
				t.Errorf("street: want %q, got %q", tc.wantStreet, parsed.StreetAddress)
			}
			if parsed.SecondaryAddress != tc.wantSecondary {
				t.Errorf("secondary: want %q, got %q", tc.wantSecondary, parsed.SecondaryAddress)
			}
			if parsed.City != tc.wantCity {
				t.Errorf("city: want %q, got %q", tc.wantCity, parsed.City)
			}
			if parsed.State != tc.wantState {
				t.Errorf("state: want %q, got %q", tc.wantState, parsed.State)
			}
			if parsed.ZIPCode != tc.wantZIP {
				t.Errorf("ZIP: want %q, got %q", tc.wantZIP, parsed.ZIPCode)
			}
			if parsed.ZIPPlus4 != tc.wantZIPPlus4 {
				t.Errorf("ZIP+4: want %q, got %q", tc.wantZIPPlus4, parsed.ZIPPlus4)
			}
			if tc.wantDiagnostics == nil {
				if len(parsed.Diagnostics) != 0 {
					t.Errorf("expected no diagnostics, got %v", parsed.Diagnostics)
				}
			} else {
				for _, wantDiag := range tc.wantDiagnostics {
					found := false
					for _, diag := range parsed.Diagnostics {
						if diag.Code == wantDiag.Code {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("expected diagnostic code %q, got %v", wantDiag.Code, parsed.Diagnostics)
					}
				}
			}
		})
	}
}

func TestParseFourSegmentAddressWithApartment(t *testing.T) {
	parsed := Parse("123 Main St, Apt 4B, Springfield, IL 62704")
	assertParsed(t, parsed, "123 MAIN ST", "APT 4B", "SPRINGFIELD", "IL", "62704", "")
}

func TestParseFourSegmentAddressWithSuite(t *testing.T) {
	parsed := Parse("456 Oak Ave, Suite 200, Boston, MA 02101")
	assertParsed(t, parsed, "456 OAK AVE", "STE 200", "BOSTON", "MA", "02101", "")
}

func TestParseFourSegmentAddressWithUnit(t *testing.T) {
	parsed := Parse("789 Pine Rd, Unit 5, Seattle, WA 98101")
	assertParsed(t, parsed, "789 PINE RD", "UNIT 5", "SEATTLE", "WA", "98101", "")
}

func TestParseFourSegmentAddressWithHashSign(t *testing.T) {
	parsed := Parse("321 Elm St, #12, Portland, OR 97201")
	assertParsed(t, parsed, "321 ELM ST", "#12", "PORTLAND", "OR", "97201", "")
}

func TestParseFiveSegmentAddressWithMultipleCityParts(t *testing.T) {
	parsed := Parse("100 Broadway, Floor 3, New York, NY, NY 10005")
	if parsed.StreetAddress != "100 BROADWAY" || parsed.SecondaryAddress != "FL 3" {
		t.Fatalf("unexpected parsing for multi-city parts: %#v", parsed)
	}
	// City aggregation differs intentionally; only street + secondary validated here.
}

func TestParseThreeSegmentAddressStillWorks(t *testing.T) {
	parsed := Parse("123 Main Street, Springfield, IL 62704")
	assertParsed(t, parsed, "123 MAIN ST", "", "SPRINGFIELD", "IL", "62704", "")
}

func TestParseStreetWithSecondaryPreserved(t *testing.T) {
	parsed := Parse("456 Elm St Apt 5B, Chicago, IL 60614")
	assertParsed(t, parsed, "456 ELM ST", "APT 5B", "CHICAGO", "IL", "60614", "")
}

// TestEmptyInput tests empty input handling
func TestEmptyInput(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"Empty string", ""},
		{"Whitespace only", "   "},
		{"Tabs and spaces", "\t  \n  "},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			parsed := Parse(tc.input)

			// Should have empty_input diagnostic
			foundDiag := false
			for _, diag := range parsed.Diagnostics {
				if diag.Code == "empty_input" {
					foundDiag = true
					break
				}
			}
			if !foundDiag {
				t.Errorf("expected empty_input diagnostic, got %v", parsed.Diagnostics)
			}

			// All fields should be empty
			if parsed.StreetAddress != "" || parsed.City != "" || parsed.State != "" {
				t.Errorf("expected all fields empty for empty input")
			}
		})
	}
}

// TestInvalidStateCodes tests invalid state code handling
func TestInvalidStateCodes(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedState string
	}{
		{"Invalid two-letter state", "123 Main St, Springfield, XX 12345", "XX"},
		{"Invalid state ZZ", "456 Oak Ave, Boston, ZZ 02101", "ZZ"},
		{"Three-letter state", "789 Pine Rd, Seattle, WAA 98101", ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			parsed := Parse(tc.input)

			if tc.expectedState != "" && parsed.State != tc.expectedState {
				t.Errorf("expected state %q, got %q", tc.expectedState, parsed.State)
			}

			// Should have unknown_state diagnostic for invalid two-letter codes
			if tc.expectedState != "" {
				foundDiag := false
				for _, diag := range parsed.Diagnostics {
					if diag.Code == "unknown_state" {
						foundDiag = true
						break
					}
				}
				if !foundDiag {
					t.Errorf("expected unknown_state diagnostic, got %v", parsed.Diagnostics)
				}
			}
		})
	}
}

// TestMalformedZIPCodes tests malformed ZIP code handling
func TestMalformedZIPCodes(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		wantZIP         string
		wantZIPPlus4    string
		wantDiagnostics bool
	}{
		{
			name:            "Four digit ZIP",
			input:           "123 Main St, Springfield, IL 1234",
			wantZIP:         "",
			wantDiagnostics: true,
		},
		{
			name:            "Six digit ZIP",
			input:           "123 Main St, Springfield, IL 123456",
			wantZIP:         "",
			wantDiagnostics: true,
		},
		{
			name:            "Valid ZIP",
			input:           "123 Main St, Springfield, IL 62704",
			wantZIP:         "62704",
			wantDiagnostics: false,
		},
		{
			name:            "Valid ZIP+4",
			input:           "123 Main St, Springfield, IL 62704-1234",
			wantZIP:         "62704",
			wantZIPPlus4:    "1234",
			wantDiagnostics: false,
		},
		{
			name:            "Invalid ZIP+4 (three digits)",
			input:           "123 Main St, Springfield, IL 62704-123",
			wantZIP:         "",
			wantDiagnostics: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			parsed := Parse(tc.input)

			if parsed.ZIPCode != tc.wantZIP {
				t.Errorf("ZIP: want %q, got %q", tc.wantZIP, parsed.ZIPCode)
			}
			if parsed.ZIPPlus4 != tc.wantZIPPlus4 {
				t.Errorf("ZIP+4: want %q, got %q", tc.wantZIPPlus4, parsed.ZIPPlus4)
			}

			if tc.wantDiagnostics && len(parsed.Diagnostics) == 0 {
				t.Errorf("expected diagnostics for malformed ZIP, got none")
			}
		})
	}
}

// TestDirectionalNormalization tests directional prefix/suffix normalization
func TestDirectionalNormalization(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantStreet string
	}{
		{
			name:       "North prefix",
			input:      "123 North Main St, Springfield, IL 62704",
			wantStreet: "123 N MAIN ST",
		},
		{
			name:       "South prefix",
			input:      "456 South Oak Ave, Chicago, IL 60614",
			wantStreet: "456 S OAK AVE",
		},
		{
			name:       "East prefix",
			input:      "789 East Elm St, Boston, MA 02101",
			wantStreet: "789 E ELM ST",
		},
		{
			name:       "West prefix",
			input:      "321 West Pine Rd, Seattle, WA 98101",
			wantStreet: "321 W PINE RD",
		},
		{
			name:       "Northeast prefix",
			input:      "100 Northeast Broadway, Portland, OR 97201",
			wantStreet: "100 NE BROADWAY",
		},
		{
			name:       "Northwest prefix",
			input:      "200 Northwest Park Ave, Miami, FL 33101",
			wantStreet: "200 NW PARK AVE",
		},
		{
			name:       "Southeast prefix",
			input:      "300 Southeast Ocean Dr, Key West, FL 33040",
			wantStreet: "300 SE OCEAN DR",
		},
		{
			name:       "Southwest prefix",
			input:      "400 Southwest Sunset Blvd, Los Angeles, CA 90001",
			wantStreet: "400 SW SUNSET BLVD",
		},
		{
			name:       "Abbreviated directional N",
			input:      "500 N Michigan Ave, Chicago, IL 60611",
			wantStreet: "500 N MICHIGAN AVE",
		},
		{
			name:       "Abbreviated directional S",
			input:      "600 S State St, Salt Lake City, UT 84101",
			wantStreet: "600 S STATE ST",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			parsed := Parse(tc.input)
			if parsed.StreetAddress != tc.wantStreet {
				t.Errorf("street: want %q, got %q", tc.wantStreet, parsed.StreetAddress)
			}
		})
	}
}

// TestStreetSuffixAbbreviations tests street suffix normalization
func TestStreetSuffixAbbreviations(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantStreet string
	}{
		{
			name:       "Street to ST",
			input:      "123 Main Street, Springfield, IL 62704",
			wantStreet: "123 MAIN ST",
		},
		{
			name:       "Avenue to AVE",
			input:      "456 Oak Avenue, Chicago, IL 60614",
			wantStreet: "456 OAK AVE",
		},
		{
			name:       "Boulevard to BLVD",
			input:      "789 Sunset Boulevard, Los Angeles, CA 90001",
			wantStreet: "789 SUNSET BLVD",
		},
		{
			name:       "Drive to DR",
			input:      "321 Park Drive, Boston, MA 02101",
			wantStreet: "321 PARK DR",
		},
		{
			name:       "Lane to LN",
			input:      "100 Memory Lane, Seattle, WA 98101",
			wantStreet: "100 MEMORY LN",
		},
		{
			name:       "Road to RD",
			input:      "200 Country Road, Portland, OR 97201",
			wantStreet: "200 COUNTRY RD",
		},
		{
			name:       "Court to CT",
			input:      "300 Kings Court, Miami, FL 33101",
			wantStreet: "300 KINGS CT",
		},
		{
			name:       "Circle to CIR",
			input:      "400 Winners Circle, Louisville, KY 40201",
			wantStreet: "400 WINNERS CIR",
		},
		{
			name:       "Place to PL",
			input:      "500 Market Place, New York, NY 10001",
			wantStreet: "500 MARKET PL",
		},
		{
			name:       "Terrace to TER",
			input:      "600 Ocean Terrace, Key West, FL 33040",
			wantStreet: "600 OCEAN TER",
		},
		{
			name:       "Trail to TRL",
			input:      "700 Mountain Trail, Denver, CO 80201",
			wantStreet: "700 MOUNTAIN TRL",
		},
		{
			name:       "Parkway to PKWY",
			input:      "800 Lake Parkway, Minneapolis, MN 55401",
			wantStreet: "800 LAKE PKWY",
		},
		{
			name:       "Alley to ALY",
			input:      "900 Back Alley, San Francisco, CA 94101",
			wantStreet: "900 BACK ALY",
		},
		{
			name:       "Square to SQ",
			input:      "1000 Town Square, Boston, MA 02101",
			wantStreet: "1000 TOWN SQ",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			parsed := Parse(tc.input)
			if parsed.StreetAddress != tc.wantStreet {
				t.Errorf("street: want %q, got %q", tc.wantStreet, parsed.StreetAddress)
			}
		})
	}
}

// TestRuralRoutes tests rural route address handling
func TestRuralRoutes(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantStreet string
	}{
		{
			name:       "Rural Route abbreviated",
			input:      "RR 1 Box 123, Springfield, IL 62704",
			wantStreet: "RR 1 BOX 123",
		},
		{
			name:       "Rural Route full",
			input:      "Rural Route 2 Box 456, Anytown, NY 12345",
			wantStreet: "RURAL ROUTE 2 BOX 456",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			parsed := Parse(tc.input)
			if parsed.StreetAddress != tc.wantStreet {
				t.Errorf("street: want %q, got %q", tc.wantStreet, parsed.StreetAddress)
			}
		})
	}
}

// TestMilitaryMail tests APO, FPO, DPO addresses
func TestMilitaryMail(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		wantStreet    string
		wantSecondary string
		wantCity      string
		wantState     string
		wantZIP       string
	}{
		{
			name:          "APO address",
			input:         "PSC 1234, APO, AE 09123",
			wantStreet:    "PSC 1234",
			wantSecondary: "",
			wantCity:      "APO",
			wantState:     "AE",
			wantZIP:       "09123",
		},
		{
			name:          "FPO address - Box is part of street",
			input:         "PSC 1234 Box 5678, FPO, AP 96543",
			wantStreet:    "PSC 1234 BOX 5678",
			wantSecondary: "",
			wantCity:      "FPO",
			wantState:     "AP",
			wantZIP:       "96543",
		},
		{
			name:          "DPO address - Box is part of street",
			input:         "CMR 456 Box 789, DPO, AE 09876",
			wantStreet:    "CMR 456 BOX 789",
			wantSecondary: "",
			wantCity:      "DPO",
			wantState:     "AE",
			wantZIP:       "09876",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			parsed := Parse(tc.input)
			if parsed.StreetAddress != tc.wantStreet {
				t.Errorf("street: want %q, got %q", tc.wantStreet, parsed.StreetAddress)
			}
			if parsed.SecondaryAddress != tc.wantSecondary {
				t.Errorf("secondary: want %q, got %q", tc.wantSecondary, parsed.SecondaryAddress)
			}
			if parsed.City != tc.wantCity {
				t.Errorf("city: want %q, got %q", tc.wantCity, parsed.City)
			}
			if parsed.State != tc.wantState {
				t.Errorf("state: want %q, got %q", tc.wantState, parsed.State)
			}
			if parsed.ZIPCode != tc.wantZIP {
				t.Errorf("ZIP: want %q, got %q", tc.wantZIP, parsed.ZIPCode)
			}
		})
	}
}

// TestIntersections tests intersection address handling
func TestIntersections(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantStreet string
	}{
		{
			name:       "Intersection with &",
			input:      "Main St & Oak Ave, Springfield, IL 62704",
			wantStreet: "MAIN ST & OAK AVE",
		},
		{
			name:       "Intersection with AND",
			input:      "Elm St and Pine Rd, Chicago, IL 60614",
			wantStreet: "ELM ST AND PINE RD",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			parsed := Parse(tc.input)
			if parsed.StreetAddress != tc.wantStreet {
				t.Errorf("street: want %q, got %q", tc.wantStreet, parsed.StreetAddress)
			}
		})
	}
}

// TestAdditionalSecondaryDesignators tests various secondary address formats
func TestAdditionalSecondaryDesignators(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		wantSecondary string
	}{
		{
			name:          "Room number",
			input:         "123 Main St, Room 101, Springfield, IL 62704",
			wantSecondary: "RM 101",
		},
		{
			name:          "Room abbreviated",
			input:         "456 Oak Ave, Rm 202, Chicago, IL 60614",
			wantSecondary: "RM 202",
		},
		{
			name:          "Building",
			input:         "789 Elm St, Building A, Boston, MA 02101",
			wantSecondary: "BLDG A",
		},
		{
			name:          "Building abbreviated",
			input:         "321 Pine Rd, Bldg B, Seattle, WA 98101",
			wantSecondary: "BLDG B",
		},
		{
			name:          "Lot number",
			input:         "100 Park Dr, Lot 5, Portland, OR 97201",
			wantSecondary: "LOT 5",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			parsed := Parse(tc.input)
			if parsed.SecondaryAddress != tc.wantSecondary {
				t.Errorf("secondary: want %q, got %q", tc.wantSecondary, parsed.SecondaryAddress)
			}
		})
	}
}

// TestEdgeCasesAndDiagnostics tests additional edge cases
func TestEdgeCasesAndDiagnostics(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		wantDiagCode    DiagnosticCode
		wantDiagPresent bool
	}{
		{
			name:            "Single segment only",
			input:           "123 Main Street",
			wantDiagCode:    "insufficient_segments",
			wantDiagPresent: true,
		},
		{
			name:            "No street number",
			input:           "Main Street, Springfield, IL 62704",
			wantDiagCode:    "",
			wantDiagPresent: false,
		},
		{
			name:            "State without ZIP",
			input:           "123 Main St, Springfield, IL",
			wantDiagCode:    "missing_state_zip",
			wantDiagPresent: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			parsed := Parse(tc.input)

			if tc.wantDiagPresent {
				foundDiag := false
				for _, diag := range parsed.Diagnostics {
					if diag.Code == tc.wantDiagCode {
						foundDiag = true
						break
					}
				}
				if !foundDiag {
					t.Errorf("expected diagnostic code %q, got %v", tc.wantDiagCode, parsed.Diagnostics)
				}
			}
		})
	}
}

// TestToAddressRequest tests the conversion to AddressRequest
func TestToAddressRequest(t *testing.T) {
	parsed := Parse("123 Main St Apt 5B, Springfield, IL 62704-1234")
	req := parsed.ToAddressRequest()

	if req.StreetAddress != "123 MAIN ST" {
		t.Errorf("street: want %q, got %q", "123 MAIN ST", req.StreetAddress)
	}
	if req.SecondaryAddress != "APT 5B" {
		t.Errorf("secondary: want %q, got %q", "APT 5B", req.SecondaryAddress)
	}
	if req.City != "SPRINGFIELD" {
		t.Errorf("city: want %q, got %q", "SPRINGFIELD", req.City)
	}
	if req.State != "IL" {
		t.Errorf("state: want %q, got %q", "IL", req.State)
	}
	if req.ZIPCode != "62704" {
		t.Errorf("ZIP: want %q, got %q", "62704", req.ZIPCode)
	}
	if req.ZIPPlus4 != "1234" {
		t.Errorf("ZIP+4: want %q, got %q", "1234", req.ZIPPlus4)
	}
}

// TestSortDiagnostics tests the diagnostic sorting behavior
func TestSortDiagnostics(t *testing.T) {
	// Test case with multiple diagnostics that should be sorted
	parsed := Parse("123")

	// Should have multiple diagnostics in sorted order
	if len(parsed.Diagnostics) == 0 {
		t.Fatal("expected diagnostics, got none")
	}

	// Errors should come before warnings
	for i := 0; i < len(parsed.Diagnostics)-1; i++ {
		curr := parsed.Diagnostics[i]
		next := parsed.Diagnostics[i+1]

		// If current is warning, next should not be error
		if curr.Severity == SeverityWarning && next.Severity == SeverityError {
			t.Errorf("diagnostics not sorted: warning before error at index %d", i)
		}

		// Within same severity, codes should be sorted
		if curr.Severity == next.Severity {
			if curr.Code > next.Code {
				t.Errorf("diagnostics not sorted: code %q before %q at index %d", curr.Code, next.Code, i)
			}
		}
	}
}

// TestSecondaryDesignatorVariations tests various secondary designator normalizations
func TestSecondaryDesignatorVariations(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		wantSecondary string
	}{
		{
			name:          "Apartment abbreviation with period",
			input:         "123 Main St Apt. 5B, Chicago, IL 60614",
			wantSecondary: "APT 5B",
		},
		{
			name:          "Suite abbreviation",
			input:         "456 Oak Ave Ste 200, Boston, MA 02101",
			wantSecondary: "STE 200",
		},
		{
			name:          "Room abbreviation",
			input:         "789 Pine Rd Rm 10, Seattle, WA 98101",
			wantSecondary: "RM 10",
		},
		{
			name:          "Lot designator",
			input:         "321 Oak St Lot 45, Portland, OR 97201",
			wantSecondary: "LOT 45",
		},
		{
			name:          "Building full word",
			input:         "100 Main St Building C, Austin, TX 78701",
			wantSecondary: "BLDG C",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			parsed := Parse(tc.input)
			if parsed.SecondaryAddress != tc.wantSecondary {
				t.Errorf("secondary: want %q, got %q", tc.wantSecondary, parsed.SecondaryAddress)
			}
		})
	}
}

// TestStandaloneSecondarySegments tests secondary addresses in separate segments
func TestStandaloneSecondarySegments(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		wantSecondary string
	}{
		{
			name:          "Apartment in separate segment",
			input:         "123 Main St, Apartment 5B, Chicago, IL 60614",
			wantSecondary: "APT 5B",
		},
		{
			name:          "Suite in separate segment",
			input:         "456 Oak Ave, Suite 200, Boston, MA 02101",
			wantSecondary: "STE 200",
		},
		{
			name:          "Unit in separate segment with dash",
			input:         "789 Pine Rd, Unit-3, Seattle, WA 98101",
			wantSecondary: "UNIT 3",
		},
		{
			name:          "Floor in separate segment",
			input:         "100 Broadway, Floor 5, New York, NY 10001",
			wantSecondary: "FL 5",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			parsed := Parse(tc.input)
			if parsed.SecondaryAddress != tc.wantSecondary {
				t.Errorf("secondary: want %q, got %q", tc.wantSecondary, parsed.SecondaryAddress)
			}
		})
	}
}

// TestSplitSegments tests edge cases in segment splitting
func TestSplitSegments(t *testing.T) {
	// Test with trailing commas
	parsed := Parse("123 Main St,, Springfield,, IL 62704")
	if parsed.StreetAddress != "123 MAIN ST" {
		t.Errorf("street: want %q, got %q", "123 MAIN ST", parsed.StreetAddress)
	}
	if parsed.City != "SPRINGFIELD" {
		t.Errorf("city: want %q, got %q", "SPRINGFIELD", parsed.City)
	}
}

func TestPreprocessInput(t *testing.T) {
	tests := []struct {
		name string
		in   string
		out  string
	}{
		{
			name: "Empty string",
			in:   "",
			out:  "",
		},
		{
			name: "Collapse multiple spaces",
			in:   "123   Main   St",
			out:  "123 Main St",
		},
		{
			name: "Trim leading and trailing",
			in:   "   456 Oak Ave   ",
			out:  "456 Oak Ave",
		},
		{
			name: "Normalize mixed whitespace",
			in:   "789\tPine\nRd",
			out:  "789 Pine Rd",
		},
		{
			name: "Preserve punctuation spacing",
			in:   "101 Maple St , Springfield",
			out:  "101 Maple St , Springfield",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := preprocessInput(tc.in)
			if got != tc.out {
				t.Errorf("preprocessInput(%q) = %q, want %q", tc.in, got, tc.out)
			}
		})
	}
}
