package parser

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"unicode"

	"github.com/my-eq/go-usps/models"
)

// DiagnosticSeverity represents the level of a diagnostic reported during parsing.
type DiagnosticSeverity string

const (
	// SeverityError indicates that parsing could not produce a USPS-compliant address without intervention.
	SeverityError DiagnosticSeverity = "error"
	// SeverityWarning indicates a recoverable issue that may still yield a valid address but should be reviewed.
	SeverityWarning DiagnosticSeverity = "warning"
)

// Diagnostic describes an issue encountered while normalizing an address.
type Diagnostic struct {
	Severity DiagnosticSeverity
	Code     string
	Message  string
	Span     TextSpan
}

// TextSpan identifies the byte offsets of a diagnostic within the original input.
type TextSpan struct {
	Start int
	End   int
}

// ParsedAddress holds the normalized components expected by USPS Publication 28.
type ParsedAddress struct {
	StreetAddress    string
	SecondaryAddress string
	City             string
	State            string
	ZIPCode          string
	ZIPPlus4         string
	Diagnostics      []Diagnostic
}

// Parse analyzes the provided free-form string and returns a USPS-formatted address with diagnostics.
func Parse(input string) ParsedAddress {
	normalized := preprocessInput(input)
	address := ParsedAddress{}
	address.Diagnostics = make([]Diagnostic, 0)

	if strings.TrimSpace(normalized) == "" {
		address.Diagnostics = append(address.Diagnostics, Diagnostic{
			Severity: SeverityError,
			Code:     "empty_input",
			Message:  "address input is empty",
		})
		return address
	}

	segments := splitSegments(normalized)
	if len(segments) < 3 {
		address.Diagnostics = append(address.Diagnostics, Diagnostic{
			Severity: SeverityError,
			Code:     "insufficient_segments",
			Message:  "expected street, city, and state segments separated by commas",
		})
	}

	streetSegment := ""
	citySegments := []string{}
	stateSegment := ""

	switch len(segments) {
	case 0:
		// already handled by diagnostics
	case 1:
		streetSegment = segments[0]
	case 2:
		streetSegment = segments[0]
		stateSegment = segments[1]
	default:
		streetSegment = segments[0]
		stateSegment = segments[len(segments)-1]
		citySegments = segments[1 : len(segments)-1]
	}

	street, secondary, streetDiags := normalizeStreet(streetSegment)
	address.StreetAddress = street
	address.SecondaryAddress = secondary
	address.Diagnostics = append(address.Diagnostics, streetDiags...)

	city, cityDiags := normalizeCity(citySegments)
	address.City = city
	address.Diagnostics = append(address.Diagnostics, cityDiags...)

	state, zip, zip4, regionDiags := normalizeRegion(stateSegment)
	address.State = state
	address.ZIPCode = zip
	address.ZIPPlus4 = zip4
	address.Diagnostics = append(address.Diagnostics, regionDiags...)

	sortDiagnostics(address.Diagnostics)
	return address
}

// ToAddressRequest converts the parsed address into a models.AddressRequest.
func (p ParsedAddress) ToAddressRequest() models.AddressRequest {
	return models.AddressRequest{
		StreetAddress:    p.StreetAddress,
		SecondaryAddress: p.SecondaryAddress,
		City:             p.City,
		State:            p.State,
		ZIPCode:          p.ZIPCode,
		ZIPPlus4:         p.ZIPPlus4,
	}
}

var whitespace = regexp.MustCompile(`\s+`)

func preprocessInput(input string) string {
	trimmed := strings.TrimSpace(input)
	trimmed = whitespace.ReplaceAllString(trimmed, " ")
	return trimmed
}

func splitSegments(input string) []string {
	if input == "" {
		return nil
	}
	pieces := strings.Split(input, ",")
	segments := make([]string, 0, len(pieces))
	for _, piece := range pieces {
		cleaned := strings.TrimSpace(piece)
		if cleaned != "" {
			segments = append(segments, cleaned)
		}
	}
	return segments
}

var (
	// secondaryPattern matches secondary address units such as "APT 5B", "SUITE #12", "UNIT 3", etc.
	// 
	// Regex breakdown:
	//   (?i)                : Case-insensitive match
	//   \b                  : Word boundary to ensure unit type is a separate word
	//   (APT|APARTMENT|UNIT|STE|SUITE|RM|ROOM|FL|FLOOR|BLDG|BUILDING|LOT|#)
	//                       : Capture group 1 - matches the unit type (e.g., "APT", "SUITE", "#")
	//   \b[ \-#]*           : Matches optional whitespace, hyphens, or "#" after the unit type
	//   (.+)$               : Capture group 2 - matches the unit identifier (e.g., "5B", "12", "3")
	// 
	// Example matches:
	//   "APT 5B"         => group 1: "APT", group 2: "5B"
	//   "SUITE #12"      => group 1: "SUITE", group 2: "12"
	//   "UNIT-3"         => group 1: "UNIT", group 2: "3"
	//   "#7"             => group 1: "#", group 2: "7"
	secondaryPattern = regexp.MustCompile(`(?i)\b(?:(APT|APARTMENT|UNIT|STE|SUITE|RM|ROOM|FL|FLOOR|BLDG|BUILDING|LOT|#)\b[ \-#]*)(.+)$`)
	poBoxPattern     = regexp.MustCompile(`(?i)^P\s*O\s*BOX\s+(\d+[A-Z0-9]*)$`)
	directionalMap   = map[string]string{
		"N": "N", "NORTH": "N",
		"S": "S", "SOUTH": "S",
		"E": "E", "EAST": "E",
		"W": "W", "WEST": "W",
		"NE": "NE", "NORTHEAST": "NE",
		"NW": "NW", "NORTHWEST": "NW",
		"SE": "SE", "SOUTHEAST": "SE",
		"SW": "SW", "SOUTHWEST": "SW",
	}
	streetSuffixMap = map[string]string{
		"ALLEY":     "ALY",
		"AVENUE":    "AVE",
		"BOULEVARD": "BLVD",
		"CIRCLE":    "CIR",
		"COURT":     "CT",
		"DRIVE":     "DR",
		"LANE":      "LN",
		"PARKWAY":   "PKWY",
		"PLACE":     "PL",
		"ROAD":      "RD",
		"SQUARE":    "SQ",
		"STREET":    "ST",
		"TERRACE":   "TER",
		"TRAIL":     "TRL",
		"WAY":       "WAY",
	}
	secondaryMap = map[string]string{
		"APT":       "APT",
		"APARTMENT": "APT",
		"UNIT":      "UNIT",
		"STE":       "STE",
		"SUITE":     "STE",
		"ROOM":      "RM",
		"RM":        "RM",
		"FL":        "FL",
		"FLOOR":     "FL",
		"BLDG":      "BLDG",
		"BUILDING":  "BLDG",
		"LOT":       "LOT",
		"PO BOX":    "PO BOX",
	}
)

func normalizeStreet(segment string) (street string, secondary string, diags []Diagnostic) {
	if segment == "" {
		return "", "", []Diagnostic{{
			Severity: SeverityError,
			Code:     "missing_street",
			Message:  "street address segment is missing",
		}}
	}

	segmentUpper := strings.ToUpper(segment)

	if matches := poBoxPattern.FindStringSubmatch(segmentUpper); len(matches) == 2 {
		return fmt.Sprintf("PO BOX %s", matches[1]), "", nil
	}

	secondary = ""
	if matches := secondaryPattern.FindStringSubmatch(segmentUpper); len(matches) == 3 {
		rawDesignator := strings.TrimSpace(matches[1])
		remainder := strings.TrimSpace(matches[2])
		normalizedDesignator := normalizeSecondaryDesignator(rawDesignator)
		secondary = strings.TrimSpace(normalizedDesignator + " " + remainder)
		segmentUpper = strings.TrimSpace(segmentUpper[:strings.Index(segmentUpper, matches[0])])
	}

	parts := strings.Fields(segmentUpper)
	normalizedParts := make([]string, 0, len(parts))
	for i, part := range parts {
		if normalized, ok := directionalMap[part]; ok && (i == 1 || i == len(parts)-1) {
			normalizedParts = append(normalizedParts, normalized)
			continue
		}
		if normalized, ok := streetSuffixMap[part]; ok && i == len(parts)-1 {
			normalizedParts = append(normalizedParts, normalized)
			continue
		}
		normalizedParts = append(normalizedParts, part)
	}

	street = strings.Join(normalizedParts, " ")
	if street == "" {
		diags = append(diags, Diagnostic{
			Severity: SeverityError,
			Code:     "empty_street",
			Message:  "could not determine primary street address",
		})
	}
	return street, secondary, diags
}

func normalizeSecondaryDesignator(designator string) string {
	designator = strings.ToUpper(strings.TrimSpace(designator))
	if mapped, ok := secondaryMap[designator]; ok {
		return mapped
	}
	if mapped, ok := secondaryMap[strings.ReplaceAll(designator, ".", "")]; ok {
		return mapped
	}
	return designator
}

func normalizeCity(segments []string) (string, []Diagnostic) {
	if len(segments) == 0 {
		return "", []Diagnostic{{
			Severity: SeverityWarning,
			Code:     "missing_city",
			Message:  "city component missing; USPS Publication 28 requires a city or acceptable city name",
		}}
	}
	city := strings.ToUpper(strings.Join(segments, " "))
	return city, nil
}

var (
	stateZipPattern = regexp.MustCompile(`(?i)^([A-Z]{2})\s+(\d{5})(?:[\-\s]*(\d{4}))?$`)
)

func normalizeRegion(segment string) (state, zip, zip4 string, diags []Diagnostic) {
	if segment == "" {
		diags = append(diags, Diagnostic{
			Severity: SeverityError,
			Code:     "missing_state_zip",
			Message:  "state and ZIP segment missing",
		})
		return
	}

	segment = strings.ToUpper(strings.TrimSpace(segment))
	matches := stateZipPattern.FindStringSubmatch(segment)
	if len(matches) == 0 {
		hasDigits := strings.IndexFunc(segment, unicode.IsDigit) >= 0
		code := "invalid_state_zip"
		message := "expected two-letter state abbreviation followed by ZIP Code"
		if !hasDigits {
			code = "missing_state_zip"
			message = "state and ZIP Code are required after the city"
		}
		diags = append(diags, Diagnostic{
			Severity: SeverityError,
			Code:     code,
			Message:  message,
		})
		return
	}

	state = matches[1]
	if !isValidState(state) {
		diags = append(diags, Diagnostic{
			Severity: SeverityError,
			Code:     "unknown_state",
			Message:  fmt.Sprintf("state abbreviation %q is not recognized by USPS", state),
		})
	}

	zip = matches[2]
	if len(matches) > 3 {
		zip4 = matches[3]
	}
	return
}

var usStateAbbreviations = map[string]struct{}{
	"AL": {}, "AK": {}, "AZ": {}, "AR": {}, "CA": {}, "CO": {}, "CT": {}, "DE": {}, "DC": {}, "FL": {},
	"GA": {}, "HI": {}, "ID": {}, "IL": {}, "IN": {}, "IA": {}, "KS": {}, "KY": {}, "LA": {}, "ME": {},
	"MD": {}, "MA": {}, "MI": {}, "MN": {}, "MS": {}, "MO": {}, "MT": {}, "NE": {}, "NV": {}, "NH": {},
	"NJ": {}, "NM": {}, "NY": {}, "NC": {}, "ND": {}, "OH": {}, "OK": {}, "OR": {}, "PA": {}, "RI": {},
	"SC": {}, "SD": {}, "TN": {}, "TX": {}, "UT": {}, "VT": {}, "VA": {}, "WA": {}, "WV": {}, "WI": {},
	"WY": {}, "PR": {}, "VI": {}, "GU": {}, "AS": {}, "MP": {},
}

func isValidState(state string) bool {
	_, ok := usStateAbbreviations[state]
	return ok
}

func sortDiagnostics(diags []Diagnostic) {
	sort.SliceStable(diags, func(i, j int) bool {
		if diags[i].Severity != diags[j].Severity {
			return diags[i].Severity == SeverityError
		}
		if diags[i].Code != diags[j].Code {
			return diags[i].Code < diags[j].Code
		}
		if diags[i].Span.Start != diags[j].Span.Start {
			return diags[i].Span.Start < diags[j].Span.Start
		}
		return diags[i].Span.End < diags[j].Span.End
	})
}
