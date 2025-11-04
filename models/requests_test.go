package models

import (
	"fmt"
	"reflect"
	"testing"
)

func TestAddressRequest_DeliveryLine(t *testing.T) {
	tests := []struct {
		name string
		addr *AddressRequest
		want string
	}{
		{
			name: "street address only",
			addr: &AddressRequest{
				StreetAddress: "123 MAIN ST",
			},
			want: "123 MAIN ST",
		},
		{
			name: "street with secondary",
			addr: &AddressRequest{
				StreetAddress:    "123 MAIN ST",
				SecondaryAddress: "APT 4B",
			},
			want: "123 MAIN ST APT 4B",
		},
		{
			name: "street with suite",
			addr: &AddressRequest{
				StreetAddress:    "456 ELM AVE",
				SecondaryAddress: "STE 200",
			},
			want: "456 ELM AVE STE 200",
		},
		{
			name: "street with unit",
			addr: &AddressRequest{
				StreetAddress:    "789 OAK BLVD",
				SecondaryAddress: "UNIT 3",
			},
			want: "789 OAK BLVD UNIT 3",
		},
		{
			name: "firm only (no street)",
			addr: &AddressRequest{
				Firm: "ACME CORPORATION",
			},
			want: "ACME CORPORATION",
		},
		{
			name: "PO BOX address",
			addr: &AddressRequest{
				StreetAddress: "PO BOX 12345",
			},
			want: "PO BOX 12345",
		},
		{
			name: "excess whitespace in street",
			addr: &AddressRequest{
				StreetAddress: "  123 MAIN ST  ",
			},
			want: "123 MAIN ST",
		},
		{
			name: "excess whitespace in secondary",
			addr: &AddressRequest{
				StreetAddress:    "123 MAIN ST",
				SecondaryAddress: "  APT 4B  ",
			},
			want: "123 MAIN ST APT 4B",
		},
		{
			name: "excess whitespace in both",
			addr: &AddressRequest{
				StreetAddress:    "  123 MAIN ST  ",
				SecondaryAddress: "  APT 4B  ",
			},
			want: "123 MAIN ST APT 4B",
		},
		{
			name: "very long secondary token",
			addr: &AddressRequest{
				StreetAddress:    "123 MAIN ST",
				SecondaryAddress: "SUITE 200 BUILDING A FLOOR 3",
			},
			want: "123 MAIN ST SUITE 200 BUILDING A FLOOR 3",
		},
		{
			name: "firm with whitespace",
			addr: &AddressRequest{
				Firm: "  ACME CORPORATION  ",
			},
			want: "ACME CORPORATION",
		},
		{
			name: "empty address",
			addr: &AddressRequest{},
			want: "",
		},
		{
			name: "nil receiver",
			addr: nil,
			want: "",
		},
		{
			name: "empty strings only",
			addr: &AddressRequest{
				StreetAddress: "",
				Firm:          "",
			},
			want: "",
		},
		{
			name: "street overrides firm",
			addr: &AddressRequest{
				Firm:          "ACME CORPORATION",
				StreetAddress: "123 MAIN ST",
			},
			want: "123 MAIN ST",
		},
		{
			name: "lowercase input (passthrough - no re-casing)",
			addr: &AddressRequest{
				StreetAddress: "123 main st",
			},
			want: "123 main st",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.addr.DeliveryLine()
			if got != tt.want {
				t.Errorf("DeliveryLine() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestAddressRequest_LastLine(t *testing.T) {
	tests := []struct {
		name string
		addr *AddressRequest
		want string
	}{
		{
			name: "city, state, ZIP",
			addr: &AddressRequest{
				City:    "NEW YORK",
				State:   "NY",
				ZIPCode: "10001",
			},
			want: "NEW YORK, NY 10001",
		},
		{
			name: "city, state, ZIP+4",
			addr: &AddressRequest{
				City:     "NEW YORK",
				State:    "NY",
				ZIPCode:  "10001",
				ZIPPlus4: "1234",
			},
			want: "NEW YORK, NY 10001-1234",
		},
		{
			name: "multi-word city name",
			addr: &AddressRequest{
				City:    "LOS ANGELES",
				State:   "CA",
				ZIPCode: "90001",
			},
			want: "LOS ANGELES, CA 90001",
		},
		{
			name: "city and state only",
			addr: &AddressRequest{
				City:  "NEW YORK",
				State: "NY",
			},
			want: "NEW YORK, NY",
		},
		{
			name: "state and ZIP only",
			addr: &AddressRequest{
				State:   "NY",
				ZIPCode: "10001",
			},
			want: "NY 10001",
		},
		{
			name: "city only",
			addr: &AddressRequest{
				City: "NEW YORK",
			},
			want: "NEW YORK",
		},
		{
			name: "ZIP only",
			addr: &AddressRequest{
				ZIPCode: "10001",
			},
			want: "10001",
		},
		{
			name: "ZIP+4 only (with base ZIP)",
			addr: &AddressRequest{
				ZIPCode:  "10001",
				ZIPPlus4: "1234",
			},
			want: "10001-1234",
		},
		{
			name: "excess whitespace in city",
			addr: &AddressRequest{
				City:    "  NEW YORK  ",
				State:   "NY",
				ZIPCode: "10001",
			},
			want: "NEW YORK, NY 10001",
		},
		{
			name: "excess whitespace in state",
			addr: &AddressRequest{
				City:    "NEW YORK",
				State:   "  NY  ",
				ZIPCode: "10001",
			},
			want: "NEW YORK, NY 10001",
		},
		{
			name: "excess whitespace in ZIP",
			addr: &AddressRequest{
				City:    "NEW YORK",
				State:   "NY",
				ZIPCode: "  10001  ",
			},
			want: "NEW YORK, NY 10001",
		},
		{
			name: "excess whitespace in ZIP+4",
			addr: &AddressRequest{
				City:     "NEW YORK",
				State:    "NY",
				ZIPCode:  "10001",
				ZIPPlus4: "  1234  ",
			},
			want: "NEW YORK, NY 10001-1234",
		},
		{
			name: "military APO address",
			addr: &AddressRequest{
				City:    "APO",
				State:   "AE",
				ZIPCode: "09012",
			},
			want: "APO, AE 09012",
		},
		{
			name: "military FPO address",
			addr: &AddressRequest{
				City:    "FPO",
				State:   "AP",
				ZIPCode: "96349",
			},
			want: "FPO, AP 96349",
		},
		{
			name: "empty address",
			addr: &AddressRequest{},
			want: "",
		},
		{
			name: "nil receiver",
			addr: nil,
			want: "",
		},
		{
			name: "city without state (graceful degradation)",
			addr: &AddressRequest{
				City:    "NEW YORK",
				ZIPCode: "10001",
			},
			want: "NEW YORK 10001",
		},
		{
			name: "lowercase input (passthrough - no re-casing)",
			addr: &AddressRequest{
				City:    "new york",
				State:   "ny",
				ZIPCode: "10001",
			},
			want: "new york, ny 10001",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.addr.LastLine()
			if got != tt.want {
				t.Errorf("LastLine() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestAddressRequest_String(t *testing.T) {
	tests := []struct {
		name string
		addr *AddressRequest
		want string
	}{
		{
			name: "complete address",
			addr: &AddressRequest{
				StreetAddress: "123 MAIN ST",
				City:          "NEW YORK",
				State:         "NY",
				ZIPCode:       "10001",
			},
			want: "123 MAIN ST, NEW YORK, NY 10001",
		},
		{
			name: "complete address with secondary",
			addr: &AddressRequest{
				StreetAddress:    "123 MAIN ST",
				SecondaryAddress: "APT 4B",
				City:             "NEW YORK",
				State:            "NY",
				ZIPCode:          "10001",
			},
			want: "123 MAIN ST APT 4B, NEW YORK, NY 10001",
		},
		{
			name: "complete address with ZIP+4",
			addr: &AddressRequest{
				StreetAddress: "123 MAIN ST",
				City:          "NEW YORK",
				State:         "NY",
				ZIPCode:       "10001",
				ZIPPlus4:      "1234",
			},
			want: "123 MAIN ST, NEW YORK, NY 10001-1234",
		},
		{
			name: "firm only address",
			addr: &AddressRequest{
				Firm:    "ACME CORPORATION",
				City:    "NEW YORK",
				State:   "NY",
				ZIPCode: "10001",
			},
			want: "ACME CORPORATION, NEW YORK, NY 10001",
		},
		{
			name: "PO BOX address",
			addr: &AddressRequest{
				StreetAddress: "PO BOX 12345",
				City:          "NEW YORK",
				State:         "NY",
				ZIPCode:       "10001",
			},
			want: "PO BOX 12345, NEW YORK, NY 10001",
		},
		{
			name: "multi-word city",
			addr: &AddressRequest{
				StreetAddress: "456 ELM AVE",
				City:          "LOS ANGELES",
				State:         "CA",
				ZIPCode:       "90001",
			},
			want: "456 ELM AVE, LOS ANGELES, CA 90001",
		},
		{
			name: "military APO address",
			addr: &AddressRequest{
				StreetAddress: "UNIT 12345",
				City:          "APO",
				State:         "AE",
				ZIPCode:       "09012",
			},
			want: "UNIT 12345, APO, AE 09012",
		},
		{
			name: "delivery line only",
			addr: &AddressRequest{
				StreetAddress: "123 MAIN ST",
			},
			want: "123 MAIN ST",
		},
		{
			name: "last line only",
			addr: &AddressRequest{
				City:    "NEW YORK",
				State:   "NY",
				ZIPCode: "10001",
			},
			want: "NEW YORK, NY 10001",
		},
		{
			name: "empty address",
			addr: &AddressRequest{},
			want: "",
		},
		{
			name: "nil receiver",
			addr: nil,
			want: "",
		},
		{
			name: "excess whitespace",
			addr: &AddressRequest{
				StreetAddress: "  123 MAIN ST  ",
				City:          "  NEW YORK  ",
				State:         "  NY  ",
				ZIPCode:       "  10001  ",
			},
			want: "123 MAIN ST, NEW YORK, NY 10001",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.addr.String()
			if got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestAddressRequest_Lines(t *testing.T) {
	tests := []struct {
		name string
		addr *AddressRequest
		want []string
	}{
		{
			name: "complete address",
			addr: &AddressRequest{
				StreetAddress: "123 MAIN ST",
				City:          "NEW YORK",
				State:         "NY",
				ZIPCode:       "10001",
			},
			want: []string{"123 MAIN ST", "NEW YORK, NY 10001"},
		},
		{
			name: "complete address with secondary",
			addr: &AddressRequest{
				StreetAddress:    "123 MAIN ST",
				SecondaryAddress: "APT 4B",
				City:             "NEW YORK",
				State:            "NY",
				ZIPCode:          "10001",
			},
			want: []string{"123 MAIN ST APT 4B", "NEW YORK, NY 10001"},
		},
		{
			name: "complete address with ZIP+4",
			addr: &AddressRequest{
				StreetAddress: "123 MAIN ST",
				City:          "NEW YORK",
				State:         "NY",
				ZIPCode:       "10001",
				ZIPPlus4:      "1234",
			},
			want: []string{"123 MAIN ST", "NEW YORK, NY 10001-1234"},
		},
		{
			name: "firm only address",
			addr: &AddressRequest{
				Firm:    "ACME CORPORATION",
				City:    "NEW YORK",
				State:   "NY",
				ZIPCode: "10001",
			},
			want: []string{"ACME CORPORATION", "NEW YORK, NY 10001"},
		},
		{
			name: "PO BOX address",
			addr: &AddressRequest{
				StreetAddress: "PO BOX 12345",
				City:          "NEW YORK",
				State:         "NY",
				ZIPCode:       "10001",
			},
			want: []string{"PO BOX 12345", "NEW YORK, NY 10001"},
		},
		{
			name: "delivery line only",
			addr: &AddressRequest{
				StreetAddress: "123 MAIN ST",
			},
			want: []string{"123 MAIN ST"},
		},
		{
			name: "last line only",
			addr: &AddressRequest{
				City:    "NEW YORK",
				State:   "NY",
				ZIPCode: "10001",
			},
			want: []string{"NEW YORK, NY 10001"},
		},
		{
			name: "empty address",
			addr: &AddressRequest{},
			want: []string{},
		},
		{
			name: "nil receiver",
			addr: nil,
			want: []string{},
		},
		{
			name: "excess whitespace",
			addr: &AddressRequest{
				StreetAddress: "  123 MAIN ST  ",
				City:          "  NEW YORK  ",
				State:         "  NY  ",
				ZIPCode:       "  10001  ",
			},
			want: []string{"123 MAIN ST", "NEW YORK, NY 10001"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.addr.Lines()
			// Handle both nil and empty slices as equivalent
			if len(got) == 0 && len(tt.want) == 0 {
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Lines() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Example demonstrates using String() and Lines() to format addresses
func ExampleAddressRequest_String() {
	addr := &AddressRequest{
		StreetAddress:    "123 MAIN ST",
		SecondaryAddress: "APT 4B",
		City:             "NEW YORK",
		State:            "NY",
		ZIPCode:          "10001",
		ZIPPlus4:         "1234",
	}

	// Single-line format
	fmt.Println(addr.String())

	// Output:
	// 123 MAIN ST APT 4B, NEW YORK, NY 10001-1234
}

func ExampleAddressRequest_Lines() {
	addr := &AddressRequest{
		StreetAddress: "123 MAIN ST",
		City:          "NEW YORK",
		State:         "NY",
		ZIPCode:       "10001",
	}

	// Two-line format for mailing labels
	for _, line := range addr.Lines() {
		fmt.Println(line)
	}

	// Output:
	// 123 MAIN ST
	// NEW YORK, NY 10001
}

// Benchmark tests to ensure negligible overhead
func BenchmarkAddressRequest_DeliveryLine(b *testing.B) {
	addr := &AddressRequest{
		StreetAddress:    "123 MAIN ST",
		SecondaryAddress: "APT 4B",
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = addr.DeliveryLine()
	}
}

func BenchmarkAddressRequest_LastLine(b *testing.B) {
	addr := &AddressRequest{
		City:     "NEW YORK",
		State:    "NY",
		ZIPCode:  "10001",
		ZIPPlus4: "1234",
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = addr.LastLine()
	}
}

func BenchmarkAddressRequest_String(b *testing.B) {
	addr := &AddressRequest{
		StreetAddress: "123 MAIN ST",
		City:          "NEW YORK",
		State:         "NY",
		ZIPCode:       "10001",
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = addr.String()
	}
}

func BenchmarkAddressRequest_Lines(b *testing.B) {
	addr := &AddressRequest{
		StreetAddress: "123 MAIN ST",
		City:          "NEW YORK",
		State:         "NY",
		ZIPCode:       "10001",
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = addr.Lines()
	}
}
