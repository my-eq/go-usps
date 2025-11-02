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
