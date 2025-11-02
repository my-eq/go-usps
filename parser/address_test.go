package parser

import "testing"

func TestParseAddress_TableDriven(t *testing.T) {
	type diagExpect struct {
		Code string
	}
	tests := []struct {
		name             string
		input            string
		wantStreet       string
		wantSecondary    string
		wantCity         string
		wantState        string
		wantZIP          string
		wantZIPPlus4     string
		wantDiagnostics  []diagExpect
	}{
		{
			name:          "Basic street address",
			input:         "123 Main Street, Springfield, IL 62704",
			wantStreet:    "123 MAIN ST",
			wantSecondary: "",
			wantCity:      "SPRINGFIELD",
			wantState:     "IL",
			wantZIP:       "62704",
			wantZIPPlus4:  "",
			wantDiagnostics: nil,
		},
		{
			name:          "With secondary unit",
			input:         "456 Elm St Apt 5B, Chicago, IL 60614-1234",
			wantStreet:    "456 ELM ST",
			wantSecondary: "APT 5B",
			wantCity:      "CHICAGO",
			wantState:     "IL",
			wantZIP:       "60614",
			wantZIPPlus4:  "1234",
			wantDiagnostics: nil,
		},
		{
			name:          "PO Box",
			input:         "PO Box 123, Anytown, NY 12345",
			wantStreet:    "PO BOX 123",
			wantSecondary: "",
			wantCity:      "ANYTOWN",
			wantState:     "NY",
			wantZIP:       "12345",
			wantZIPPlus4:  "",
			wantDiagnostics: nil,
		},
		{
			name:          "Missing state produces diagnostic",
			input:         "123 Main Street, Springfield",
			wantStreet:    "123 MAIN ST",
			wantSecondary: "",
			wantCity:      "SPRINGFIELD",
			wantState:     "",
			wantZIP:       "",
			wantZIPPlus4:  "",
			wantDiagnostics: []diagExpect{{Code: "missing_state_zip"}},
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
	// Test case from the issue: "123 Main St, Apt 4B, Springfield, IL 62704"
	parsed := Parse("123 Main St, Apt 4B, Springfield, IL 62704")

	if parsed.StreetAddress != "123 MAIN ST" {
		t.Fatalf("expected street %q, got %q", "123 MAIN ST", parsed.StreetAddress)
	}
	if parsed.SecondaryAddress != "APT 4B" {
		t.Fatalf("expected secondary address %q, got %q", "APT 4B", parsed.SecondaryAddress)
	}
	if parsed.City != "SPRINGFIELD" {
		t.Fatalf("expected city %q, got %q", "SPRINGFIELD", parsed.City)
	}
	if parsed.State != "IL" {
		t.Fatalf("expected state IL, got %s", parsed.State)
	}
	if parsed.ZIPCode != "62704" {
		t.Fatalf("expected ZIP 62704, got %s", parsed.ZIPCode)
	}
}

func TestParseFourSegmentAddressWithSuite(t *testing.T) {
	parsed := Parse("456 Oak Ave, Suite 200, Boston, MA 02101")

	if parsed.StreetAddress != "456 OAK AVE" {
		t.Fatalf("expected street %q, got %q", "456 OAK AVE", parsed.StreetAddress)
	}
	if parsed.SecondaryAddress != "STE 200" {
		t.Fatalf("expected secondary address %q, got %q", "STE 200", parsed.SecondaryAddress)
	}
	if parsed.City != "BOSTON" {
		t.Fatalf("expected city %q, got %q", "BOSTON", parsed.City)
	}
	if parsed.State != "MA" {
		t.Fatalf("expected state MA, got %s", parsed.State)
	}
}

func TestParseFourSegmentAddressWithUnit(t *testing.T) {
	parsed := Parse("789 Pine Rd, Unit 5, Seattle, WA 98101")

	if parsed.StreetAddress != "789 PINE RD" {
		t.Fatalf("expected street %q, got %q", "789 PINE RD", parsed.StreetAddress)
	}
	if parsed.SecondaryAddress != "UNIT 5" {
		t.Fatalf("expected secondary address %q, got %q", "UNIT 5", parsed.SecondaryAddress)
	}
	if parsed.City != "SEATTLE" {
		t.Fatalf("expected city %q, got %q", "SEATTLE", parsed.City)
	}
}

func TestParseFourSegmentAddressWithHashSign(t *testing.T) {
	parsed := Parse("321 Elm St, #12, Portland, OR 97201")

	if parsed.StreetAddress != "321 ELM ST" {
		t.Fatalf("expected street %q, got %q", "321 ELM ST", parsed.StreetAddress)
	}
	if parsed.SecondaryAddress != "#12" {
		t.Fatalf("expected secondary address %q, got %q", "#12", parsed.SecondaryAddress)
	}
	if parsed.City != "PORTLAND" {
		t.Fatalf("expected city %q, got %q", "PORTLAND", parsed.City)
	}
}

func TestParseFiveSegmentAddressWithMultipleCityParts(t *testing.T) {
	// Test with multiple city segments that are not secondary addresses
	parsed := Parse("100 Broadway, Floor 3, New York, NY, NY 10005")

	if parsed.StreetAddress != "100 BROADWAY" {
		t.Fatalf("expected street %q, got %q", "100 BROADWAY", parsed.StreetAddress)
	}
	if parsed.SecondaryAddress != "FL 3" {
		t.Fatalf("expected secondary address %q, got %q", "FL 3", parsed.SecondaryAddress)
	}
	// Note: This tests edge case handling, actual city will be "NEW YORK NY"
}

func TestParseThreeSegmentAddressStillWorks(t *testing.T) {
	// Ensure we don't break the existing 3-segment parsing
	parsed := Parse("123 Main Street, Springfield, IL 62704")

	if parsed.StreetAddress != "123 MAIN ST" {
		t.Fatalf("expected street %q, got %q", "123 MAIN ST", parsed.StreetAddress)
	}
	if parsed.SecondaryAddress != "" {
		t.Fatalf("expected no secondary address, got %q", parsed.SecondaryAddress)
	}
	if parsed.City != "SPRINGFIELD" {
		t.Fatalf("expected city %q, got %q", "SPRINGFIELD", parsed.City)
	}
	if parsed.State != "IL" {
		t.Fatalf("expected state IL, got %s", parsed.State)
	}
}

func TestParseStreetWithSecondaryPreserved(t *testing.T) {
	// Ensure existing functionality where secondary is in street segment still works
	parsed := Parse("456 Elm St Apt 5B, Chicago, IL 60614")

	if parsed.StreetAddress != "456 ELM ST" {
		t.Fatalf("expected street %q, got %q", "456 ELM ST", parsed.StreetAddress)
	}
	if parsed.SecondaryAddress != "APT 5B" {
		t.Fatalf("expected secondary address %q, got %q", "APT 5B", parsed.SecondaryAddress)
	}
	if parsed.City != "CHICAGO" {
		t.Fatalf("expected city %q, got %q", "CHICAGO", parsed.City)
	}
}
