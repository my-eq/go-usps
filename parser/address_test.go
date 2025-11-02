package parser

import "testing"

func TestParseBasicStreetAddress(t *testing.T) {
	parsed := Parse("123 Main Street, Springfield, IL 62704")
	if parsed.StreetAddress != "123 MAIN ST" {
		t.Fatalf("expected street %q, got %q", "123 MAIN ST", parsed.StreetAddress)
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
	if len(parsed.Diagnostics) != 0 {
		t.Fatalf("expected no diagnostics, got %v", parsed.Diagnostics)
	}
}

func TestParseWithSecondaryUnit(t *testing.T) {
	parsed := Parse("456 Elm St Apt 5B, Chicago, IL 60614-1234")
	if parsed.SecondaryAddress != "APT 5B" {
		t.Fatalf("expected secondary address APT 5B, got %s", parsed.SecondaryAddress)
	}
	if parsed.ZIPCode != "60614" || parsed.ZIPPlus4 != "1234" {
		t.Fatalf("expected ZIP 60614-1234, got %s-%s", parsed.ZIPCode, parsed.ZIPPlus4)
	}
	if len(parsed.Diagnostics) != 0 {
		t.Fatalf("expected no diagnostics, got %v", parsed.Diagnostics)
	}
}

func TestParsePOBox(t *testing.T) {
	parsed := Parse("PO Box 123, Anytown, NY 12345")
	if parsed.StreetAddress != "PO BOX 123" {
		t.Fatalf("expected PO BOX 123, got %s", parsed.StreetAddress)
	}
	if parsed.SecondaryAddress != "" {
		t.Fatalf("expected empty secondary, got %s", parsed.SecondaryAddress)
	}
	if len(parsed.Diagnostics) != 0 {
		t.Fatalf("expected no diagnostics, got %v", parsed.Diagnostics)
	}
}

func TestParseMissingStateProducesDiagnostic(t *testing.T) {
	parsed := Parse("123 Main Street, Springfield")
	if len(parsed.Diagnostics) == 0 {
		t.Fatalf("expected diagnostics for missing state and ZIP")
	}
	found := false
	for _, diag := range parsed.Diagnostics {
		if diag.Code == "missing_state_zip" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected missing_state_zip diagnostic, got %v", parsed.Diagnostics)
	}
}
