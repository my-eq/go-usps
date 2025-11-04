package parser

import "testing"

func TestLexicon_NormalizeStreetSuffix(t *testing.T) {
	lex := newLexicon()

	tests := []struct {
		input string
		want  string
		found bool
	}{
		{"STREET", "ST", true},
		{"ST", "ST", true},
		{"AVENUE", "AVE", true},
		{"AVE", "AVE", true},
		{"BOULEVARD", "BLVD", true},
		{"BLVD", "BLVD", true},
		{"DRIVE", "DR", true},
		{"DR", "DR", true},
		{"LANE", "LN", true},
		{"LN", "LN", true},
		{"ROAD", "RD", true},
		{"RD", "RD", true},
		{"NOTAREALSTREET", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, found := lex.NormalizeStreetSuffix(tt.input)
			if found != tt.found {
				t.Errorf("found = %v, want %v", found, tt.found)
			}
			if got != tt.want {
				t.Errorf("got = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestLexicon_NormalizeDirectional(t *testing.T) {
	lex := newLexicon()

	tests := []struct {
		input string
		want  string
		found bool
	}{
		{"NORTH", "N", true},
		{"N", "N", true},
		{"SOUTH", "S", true},
		{"S", "S", true},
		{"EAST", "E", true},
		{"E", "E", true},
		{"WEST", "W", true},
		{"W", "W", true},
		{"NORTHEAST", "NE", true},
		{"NE", "NE", true},
		{"NORTHWEST", "NW", true},
		{"NW", "NW", true},
		{"SOUTHEAST", "SE", true},
		{"SE", "SE", true},
		{"SOUTHWEST", "SW", true},
		{"SW", "SW", true},
		{"NOTADIRECTION", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, found := lex.NormalizeDirectional(tt.input)
			if found != tt.found {
				t.Errorf("found = %v, want %v", found, tt.found)
			}
			if got != tt.want {
				t.Errorf("got = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestLexicon_NormalizeSecondaryDesignator(t *testing.T) {
	lex := newLexicon()

	tests := []struct {
		input string
		want  string
		found bool
	}{
		{"APARTMENT", "APT", true},
		{"APT", "APT", true},
		{"SUITE", "STE", true},
		{"STE", "STE", true},
		{"UNIT", "UNIT", true},
		{"FLOOR", "FL", true},
		{"FL", "FL", true},
		{"BUILDING", "BLDG", true},
		{"BLDG", "BLDG", true},
		{"#", "#", true},
		{"NOTADESIGNATOR", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, found := lex.NormalizeSecondaryDesignator(tt.input)
			if found != tt.found {
				t.Errorf("found = %v, want %v", found, tt.found)
			}
			if got != tt.want {
				t.Errorf("got = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestLexicon_NormalizeState(t *testing.T) {
	lex := newLexicon()

	tests := []struct {
		input string
		want  string
		found bool
	}{
		{"NY", "NY", true},
		{"NEW YORK", "NY", true},
		{"CA", "CA", true},
		{"CALIFORNIA", "CA", true},
		{"TX", "TX", true},
		{"TEXAS", "TX", true},
		{"FL", "FL", true},
		{"FLORIDA", "FL", true},
		{"DC", "DC", true},
		{"DISTRICT OF COLUMBIA", "DC", true},
		{"PR", "PR", true},
		{"PUERTO RICO", "PR", true},
		{"NOTASTATE", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, found := lex.NormalizeState(tt.input)
			if found != tt.found {
				t.Errorf("found = %v, want %v", found, tt.found)
			}
			if got != tt.want {
				t.Errorf("got = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNewLexicon(t *testing.T) {
	lex := newLexicon()

	if lex == nil {
		t.Fatal("newLexicon() returned nil")
	}

	if lex.streetSuffixes == nil {
		t.Error("streetSuffixes is nil")
	}

	if lex.directionals == nil {
		t.Error("directionals is nil")
	}

	if lex.secondaryDesignators == nil {
		t.Error("secondaryDesignators is nil")
	}

	if lex.states == nil {
		t.Error("states is nil")
	}

	// Check that maps have entries
	if len(lex.streetSuffixes) == 0 {
		t.Error("streetSuffixes is empty")
	}

	if len(lex.directionals) == 0 {
		t.Error("directionals is empty")
	}

	if len(lex.secondaryDesignators) == 0 {
		t.Error("secondaryDesignators is empty")
	}

	if len(lex.states) == 0 {
		t.Error("states is empty")
	}
}
