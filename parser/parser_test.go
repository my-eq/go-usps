package parser

import (
	"testing"
)

func TestParse_SimpleAddress(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		wantStreet    string
		wantCity      string
		wantState     string
		wantZIP       string
		wantDiagCount int
	}{
		{
			name:          "simple residential address",
			input:         "123 Main St, New York, NY 10001",
			wantStreet:    "123 MAIN ST",
			wantCity:      "NEW YORK",
			wantState:     "NY",
			wantZIP:       "10001",
			wantDiagCount: 0,
		},
		{
			name:          "address with directional",
			input:         "456 North Oak Avenue, Los Angeles, CA 90001",
			wantStreet:    "456 N OAK AVE",
			wantCity:      "LOS ANGELES",
			wantState:     "CA",
			wantZIP:       "90001",
			wantDiagCount: 0,
		},
		{
			name:          "address with abbreviated street type",
			input:         "789 Elm Blvd, Chicago, IL 60601",
			wantStreet:    "789 ELM BLVD",
			wantCity:      "CHICAGO",
			wantState:     "IL",
			wantZIP:       "60601",
			wantDiagCount: 0,
		},
		{
			name:          "address with secondary unit",
			input:         "123 Main St Apt 4B, New York, NY 10001",
			wantStreet:    "123 MAIN ST",
			wantCity:      "NEW YORK",
			wantState:     "NY",
			wantZIP:       "10001",
			wantDiagCount: 0,
		},
		{
			name:          "address missing ZIP",
			input:         "123 Main St, New York, NY",
			wantStreet:    "123 MAIN ST",
			wantCity:      "NEW YORK",
			wantState:     "NY",
			wantZIP:       "",
			wantDiagCount: 1, // Warning about missing ZIP
		},
		{
			name:          "address missing state",
			input:         "123 Main St, New York",
			wantStreet:    "123 MAIN ST",
			wantCity:      "NEW YORK",
			wantState:     "",
			wantZIP:       "",
			wantDiagCount: 2, // Error for missing state, warning for missing ZIP
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed, diagnostics := Parse(tt.input)

			if parsed == nil {
				t.Fatal("Parse returned nil ParsedAddress")
			}

			// Build the street address for comparison
			req := parsed.ToAddressRequest()

			if req.StreetAddress != tt.wantStreet {
				t.Errorf("StreetAddress = %q, want %q", req.StreetAddress, tt.wantStreet)
			}

			if req.City != tt.wantCity {
				t.Errorf("City = %q, want %q", req.City, tt.wantCity)
			}

			if req.State != tt.wantState {
				t.Errorf("State = %q, want %q", req.State, tt.wantState)
			}

			if req.ZIPCode != tt.wantZIP {
				t.Errorf("ZIPCode = %q, want %q", req.ZIPCode, tt.wantZIP)
			}

			if len(diagnostics) != tt.wantDiagCount {
				t.Errorf("got %d diagnostics, want %d", len(diagnostics), tt.wantDiagCount)
				for _, d := range diagnostics {
					t.Logf("  %s: %s", d.Severity, d.Message)
				}
			}
		})
	}
}

func TestParse_WithSecondaryAddress(t *testing.T) {
	tests := []struct {
		name              string
		input             string
		wantSecondary     string
		wantSecondaryUnit string
		wantSecondaryNum  string
	}{
		{
			name:              "apartment",
			input:             "123 Main St Apt 4B, New York, NY 10001",
			wantSecondary:     "APT 4B",
			wantSecondaryUnit: "APT",
			wantSecondaryNum:  "4B",
		},
		{
			name:              "suite",
			input:             "456 Oak Ave Suite 200, Boston, MA 02101",
			wantSecondary:     "STE 200",
			wantSecondaryUnit: "STE",
			wantSecondaryNum:  "200",
		},
		{
			name:              "unit",
			input:             "789 Elm St Unit 12, Seattle, WA 98101",
			wantSecondary:     "UNIT 12",
			wantSecondaryUnit: "UNIT",
			wantSecondaryNum:  "12",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed, _ := Parse(tt.input)

			if parsed.SecondaryUnit != tt.wantSecondaryUnit {
				t.Errorf("SecondaryUnit = %q, want %q", parsed.SecondaryUnit, tt.wantSecondaryUnit)
			}

			if parsed.SecondaryNumber != tt.wantSecondaryNum {
				t.Errorf("SecondaryNumber = %q, want %q", parsed.SecondaryNumber, tt.wantSecondaryNum)
			}

			req := parsed.ToAddressRequest()
			if req.SecondaryAddress != tt.wantSecondary {
				t.Errorf("SecondaryAddress = %q, want %q", req.SecondaryAddress, tt.wantSecondary)
			}
		})
	}
}

func TestParse_ZIPPlus4(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantZIP     string
		wantZIPPlus4 string
	}{
		{
			name:        "ZIP+4 with hyphen",
			input:       "123 Main St, New York, NY 10001-1234",
			wantZIP:     "10001",
			wantZIPPlus4: "1234",
		},
		{
			name:        "5-digit ZIP only",
			input:       "123 Main St, New York, NY 10001",
			wantZIP:     "10001",
			wantZIPPlus4: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed, _ := Parse(tt.input)

			if parsed.ZIPCode != tt.wantZIP {
				t.Errorf("ZIPCode = %q, want %q", parsed.ZIPCode, tt.wantZIP)
			}

			if parsed.ZIPPlus4 != tt.wantZIPPlus4 {
				t.Errorf("ZIPPlus4 = %q, want %q", parsed.ZIPPlus4, tt.wantZIPPlus4)
			}

			req := parsed.ToAddressRequest()
			if req.ZIPCode != tt.wantZIP {
				t.Errorf("Request ZIPCode = %q, want %q", req.ZIPCode, tt.wantZIP)
			}

			if req.ZIPPlus4 != tt.wantZIPPlus4 {
				t.Errorf("Request ZIPPlus4 = %q, want %q", req.ZIPPlus4, tt.wantZIPPlus4)
			}
		})
	}
}

func TestParse_Directionals(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantPre string
		wantPost string
	}{
		{
			name:    "pre-directional",
			input:   "123 North Main St, New York, NY 10001",
			wantPre: "N",
			wantPost: "",
		},
		{
			name:    "abbreviated pre-directional",
			input:   "456 E Oak Ave, Boston, MA 02101",
			wantPre: "E",
			wantPost: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed, _ := Parse(tt.input)

			if parsed.PreDirectional != tt.wantPre {
				t.Errorf("PreDirectional = %q, want %q", parsed.PreDirectional, tt.wantPre)
			}

			if parsed.PostDirectional != tt.wantPost {
				t.Errorf("PostDirectional = %q, want %q", parsed.PostDirectional, tt.wantPost)
			}
		})
	}
}

func TestDiagnosticSeverity_String(t *testing.T) {
	tests := []struct {
		severity DiagnosticSeverity
		want     string
	}{
		{SeverityInfo, "Info"},
		{SeverityWarning, "Warning"},
		{SeverityError, "Error"},
		{DiagnosticSeverity(999), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.severity.String()
			if got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParsedAddress_ToAddressRequest(t *testing.T) {
	parsed := &ParsedAddress{
		HouseNumber:     "123",
		PreDirectional:  "N",
		StreetName:      "MAIN",
		StreetSuffix:    "ST",
		SecondaryUnit:   "APT",
		SecondaryNumber: "4B",
		City:            "NEW YORK",
		State:           "NY",
		ZIPCode:         "10001",
		ZIPPlus4:        "1234",
		Firm:            "ACME CORP",
	}

	req := parsed.ToAddressRequest()

	if req.StreetAddress != "123 N MAIN ST" {
		t.Errorf("StreetAddress = %q, want %q", req.StreetAddress, "123 N MAIN ST")
	}

	if req.SecondaryAddress != "APT 4B" {
		t.Errorf("SecondaryAddress = %q, want %q", req.SecondaryAddress, "APT 4B")
	}

	if req.City != "NEW YORK" {
		t.Errorf("City = %q, want %q", req.City, "NEW YORK")
	}

	if req.State != "NY" {
		t.Errorf("State = %q, want %q", req.State, "NY")
	}

	if req.ZIPCode != "10001" {
		t.Errorf("ZIPCode = %q, want %q", req.ZIPCode, "10001")
	}

	if req.ZIPPlus4 != "1234" {
		t.Errorf("ZIPPlus4 = %q, want %q", req.ZIPPlus4, "1234")
	}

	if req.Firm != "ACME CORP" {
		t.Errorf("Firm = %q, want %q", req.Firm, "ACME CORP")
	}
}

func TestParser_New(t *testing.T) {
	p := New()
	if p == nil {
		t.Fatal("New() returned nil")
	}

	if p.tokenizer == nil {
		t.Error("tokenizer is nil")
	}

	if p.normalizer == nil {
		t.Error("normalizer is nil")
	}

	if p.validator == nil {
		t.Error("validator is nil")
	}
}
