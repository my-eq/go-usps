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

// DiagnosticCode represents the machine-readable identifier for a diagnostic.
type DiagnosticCode string

const (
	// DiagnosticCodeEmptyInput indicates the input string contained no address content.
	DiagnosticCodeEmptyInput DiagnosticCode = "empty_input"
	// DiagnosticCodeInsufficientSegments indicates the parser found fewer than street, city, and state segments.
	DiagnosticCodeInsufficientSegments DiagnosticCode = "insufficient_segments"
	// DiagnosticCodeMissingStreet indicates the primary street segment was absent.
	DiagnosticCodeMissingStreet DiagnosticCode = "missing_street"
	// DiagnosticCodeEmptyStreet indicates the parser could not determine a primary street value.
	DiagnosticCodeEmptyStreet DiagnosticCode = "empty_street"
	// DiagnosticCodeMissingCity indicates that the city component was not present.
	DiagnosticCodeMissingCity DiagnosticCode = "missing_city"
	// DiagnosticCodeMissingStateZIP indicates the state and ZIP portion was missing or malformed.
	DiagnosticCodeMissingStateZIP DiagnosticCode = "missing_state_zip"
	// DiagnosticCodeInvalidStateZIP indicates the state/ZIP portion exists but does not match the expected format.
	DiagnosticCodeInvalidStateZIP DiagnosticCode = "invalid_state_zip"
	// DiagnosticCodeUnknownState indicates the provided state code is not part of the USPS list.
	DiagnosticCodeUnknownState DiagnosticCode = "unknown_state"
	// DiagnosticCodeUnknownSecondary indicates an unrecognized secondary designator was encountered.
	DiagnosticCodeUnknownSecondary DiagnosticCode = "unknown_secondary"
)

// Diagnostic describes an issue encountered while normalizing an address.
type Diagnostic struct {
	Severity DiagnosticSeverity
	Code     DiagnosticCode
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
			Code:     DiagnosticCodeEmptyInput,
			Message:  "address input is empty",
		})
		return address
	}

	segments := splitSegments(normalized)
	if len(segments) < 3 {
		address.Diagnostics = append(address.Diagnostics, Diagnostic{
			Severity: SeverityError,
			Code:     DiagnosticCodeInsufficientSegments,
			Message:  "expected street, city, and state segments separated by commas",
		})
	}

	streetSegment := ""
	citySegments := []string{}
	stateSegment := ""
	secondarySegments := []string{}

	switch len(segments) {
	case 0:
		// already handled by diagnostics
	case 1:
		streetSegment = segments[0]
	case 2:
		streetSegment = segments[0]
		// Check if second segment looks like state+ZIP
		// If not, treat it as city
		if looksLikeStateZip(segments[1]) {
			stateSegment = segments[1]
		} else {
			citySegments = append(citySegments, segments[1])
		}
	default:
		streetSegment = segments[0]
		stateSegment = segments[len(segments)-1]
		middleSegments := segments[1 : len(segments)-1]

		// Check middle segments for secondary address indicators
		for _, seg := range middleSegments {
			if isSecondarySegment(seg) {
				normalized, diag := normalizeSecondarySegment(seg)
				if normalized != "" {
					secondarySegments = append(secondarySegments, normalized)
				}
				if diag != nil {
					address.Diagnostics = append(address.Diagnostics, *diag)
				}
				continue
			}
			citySegments = append(citySegments, seg)
		}
	}

	street, secondary, streetDiags := normalizeStreet(streetSegment)
	address.StreetAddress = street

	// If secondary was found in street segment, use it; otherwise check secondary segments
	allSecondary := make([]string, 0, len(secondarySegments)+1)
	if secondary != "" {
		allSecondary = append(allSecondary, secondary)
	}
	allSecondary = append(allSecondary, secondarySegments...)
	if len(allSecondary) > 0 {
		address.SecondaryAddress = strings.Join(allSecondary, " ")
	}

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

// preprocessInput collapses all whitespace and trims leading/trailing spaces.
func preprocessInput(input string) string {
	fields := strings.Fields(input)
	return strings.Join(fields, " ")
}

// splitSegments splits the input string on commas and returns a slice of non-empty, trimmed segments.
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

// extractSecondary returns a normalized secondary string if the segment contains one; otherwise empty string.
// isSecondarySegment checks if a segment contains secondary address indicators
func isSecondarySegment(segment string) bool {
	segmentUpper := strings.ToUpper(strings.TrimSpace(segment))
	segmentClean := strings.ReplaceAll(segmentUpper, ".", "")

	// Special handling for hash sign - it can be followed directly by a number
	if strings.HasPrefix(segmentClean, "#") {
		return true
	}

	secondaryPrefixes := []string{
		"APT", "APARTMENT",
		"UNIT", "SUITE", "STE",
		"ROOM", "RM",
		"FLOOR", "FL",
		"BLDG", "BUILDING",
		"LOT",
	}

	for _, prefix := range secondaryPrefixes {
		if strings.HasPrefix(segmentClean, prefix+" ") ||
			strings.HasPrefix(segmentClean, prefix+"-") ||
			strings.HasPrefix(segmentClean, prefix+"#") ||
			segmentClean == prefix {
			return true
		}
	}

	tokens := strings.Fields(segmentClean)
	if len(tokens) >= 2 {
		remainder := strings.Join(tokens[1:], " ")
		if looksLikeSecondaryValue(remainder) {
			return true
		}
	}

	return false
}

// normalizeSecondarySegment normalizes a standalone secondary address segment
func normalizeSecondarySegment(segment string) (string, *Diagnostic) {
	if segment == "" {
		return "", nil
	}

	segmentUpper := strings.ToUpper(strings.TrimSpace(segment))

	// Remove periods for normalization
	segmentUpper = strings.ReplaceAll(segmentUpper, ".", "")

	// Handle hash sign format (e.g., "#12" or "# 12")
	if strings.HasPrefix(segmentUpper, "#") {
		return segmentUpper, nil
	}

	// Try to match with the secondary pattern
	if matches := secondaryPattern.FindStringSubmatch(segmentUpper); len(matches) == 3 {
		rawDesignator := strings.TrimSpace(matches[1])
		remainder := strings.TrimSpace(matches[2])
		normalizedDesignator, recognized := normalizeSecondaryDesignator(rawDesignator)
		diag := newUnknownSecondaryDiagnostic(rawDesignator, recognized)
		return strings.TrimSpace(normalizedDesignator + " " + remainder), diag
	}

	// If no pattern matches, try to extract designator and number by splitting on whitespace
	parts := strings.Fields(segmentUpper)
	if len(parts) >= 2 {
		// First part might be the designator
		normalizedDesignator, recognized := normalizeSecondaryDesignator(parts[0])
		remainder := strings.Join(parts[1:], " ")
		diag := newUnknownSecondaryDiagnostic(parts[0], recognized)
		return strings.TrimSpace(normalizedDesignator + " " + remainder), diag
	}

	// Return as-is if we can't parse it
	return segmentUpper, nil
}

// splitInlineSecondary returns the portion before the secondary designator, the designator itself,
// and the remainder representing the secondary value. It searches from right to left to prefer the
// last designator occurrence so that street names containing designator words are preserved.
func splitInlineSecondary(segmentUpper string) (primary string, designator string, remainder string, ok bool) {
	matches := secondaryDesignatorPattern.FindAllStringSubmatchIndex(segmentUpper, -1)
	if len(matches) == 0 {
		return segmentUpper, "", "", false
	}

	for i := len(matches) - 1; i >= 0; i-- {
		match := matches[i]
		fullStart, fullEnd := match[0], match[1]
		designatorStart, designatorEnd := match[2], match[3]

		remainderRaw := strings.TrimLeft(segmentUpper[fullEnd:], " .-#")
		remainderRaw = strings.TrimSpace(remainderRaw)
		if remainderRaw == "" {
			continue
		}
		if !looksLikeSecondaryValue(remainderRaw) {
			continue
		}

		primaryPart := strings.TrimSpace(segmentUpper[:fullStart])
		designatorPart := segmentUpper[designatorStart:designatorEnd]
		return primaryPart, designatorPart, remainderRaw, true
	}

	parts := strings.Fields(segmentUpper)
	for i := len(parts) - 2; i >= 0; i-- {
		remainder := strings.Join(parts[i+1:], " ")
		if !looksLikeSecondaryValue(remainder) {
			continue
		}
		primaryPart := strings.Join(parts[:i], " ")
		primaryPart = strings.TrimSpace(primaryPart)
		designatorPart := parts[i]
		return primaryPart, designatorPart, remainder, true
	}

	return segmentUpper, "", "", false
}

// looksLikeSecondaryValue determines whether a string appears to be a valid secondary address value.
// It returns true if the value contains any digits, is a short token (length <= 3), or matches
// recognized secondary address keywords such as "PH", "PENTHOUSE", "REAR", etc.
func looksLikeSecondaryValue(value string) bool {
	if value == "" {
		return false
	}

	hasDigit := false
	for _, r := range value {
		if unicode.IsDigit(r) {
			hasDigit = true
			break
		}
	}
	if hasDigit {
		return true
	}

	tokens := strings.Fields(value)
	if len(tokens) == 0 {
		return false
	}

	if len(tokens) == 1 {
		token := tokens[0]
		if len(token) <= 3 {
			return true
		}
		allowed := map[string]struct{}{
			"PH":        {},
			"PENTHOUSE": {},
			"REAR":      {},
			"FRONT":     {},
			"UPPER":     {},
			"LOWER":     {},
			"BSMT":      {},
			"BASEMENT":  {},
			"LOBBY":     {},
		}
		if _, ok := allowed[token]; ok {
			return true
		}
	}

	return false
}

type directionalToken string
type directionalValue string

const (
	directionalNorth     directionalValue = "N"
	directionalSouth     directionalValue = "S"
	directionalEast      directionalValue = "E"
	directionalWest      directionalValue = "W"
	directionalNorthEast directionalValue = "NE"
	directionalNorthWest directionalValue = "NW"
	directionalSouthEast directionalValue = "SE"
	directionalSouthWest directionalValue = "SW"
)

type streetSuffixToken string
type streetSuffixValue string

const (
	streetSuffixAlley     streetSuffixValue = "ALY"
	streetSuffixAvenue    streetSuffixValue = "AVE"
	streetSuffixBoulevard streetSuffixValue = "BLVD"
	streetSuffixCircle    streetSuffixValue = "CIR"
	streetSuffixCourt     streetSuffixValue = "CT"
	streetSuffixDrive     streetSuffixValue = "DR"
	streetSuffixLane      streetSuffixValue = "LN"
	streetSuffixParkway   streetSuffixValue = "PKWY"
	streetSuffixPlace     streetSuffixValue = "PL"
	streetSuffixRoad      streetSuffixValue = "RD"
	streetSuffixSquare    streetSuffixValue = "SQ"
	streetSuffixStreet    streetSuffixValue = "ST"
	streetSuffixTerrace   streetSuffixValue = "TER"
	streetSuffixTrail     streetSuffixValue = "TRL"
	streetSuffixWay       streetSuffixValue = "WAY"
)

type secondaryDesignatorToken string
type secondaryDesignatorValue string

const (
	secondaryDesignatorApartment secondaryDesignatorValue = "APT"
	secondaryDesignatorUnit      secondaryDesignatorValue = "UNIT"
	secondaryDesignatorSuite     secondaryDesignatorValue = "STE"
	secondaryDesignatorRoom      secondaryDesignatorValue = "RM"
	secondaryDesignatorFloor     secondaryDesignatorValue = "FL"
	secondaryDesignatorBuilding  secondaryDesignatorValue = "BLDG"
	secondaryDesignatorLot       secondaryDesignatorValue = "LOT"
)

var secondaryDesignatorTokens = []secondaryDesignatorToken{
	"APT",
	"APARTMENT",
	"UNIT",
	"STE",
	"SUITE",
	"ROOM",
	"RM",
	"FL",
	"FLOOR",
	"BLDG",
	"BUILDING",
	"LOT",
}

var (
	// secondaryPattern matches secondary address units such as "APT 5B" or "SUITE #12" by
	// capturing two groups:
	//   1. The secondary designator token (e.g., "APT", "SUITE", "FL", or the literal "#").
	//   2. The trailing unit identifier, preserving spaces, hyphens, and other separators.
	// secondaryDesignatorPattern reuses the same alternation to find designators embedded within
	// the street segment so inline secondaries can be peeled off safely.
	secondaryPattern, secondaryDesignatorPattern = buildSecondaryPatterns(secondaryDesignatorTokens)
	poBoxPattern                                 = regexp.MustCompile(`(?i)^P\s*O\s*BOX\s+(\d+[A-Z0-9]*)$`)
	directionalMap                               = map[directionalToken]directionalValue{
		directionalToken("N"):         directionalNorth,
		directionalToken("NORTH"):     directionalNorth,
		directionalToken("S"):         directionalSouth,
		directionalToken("SOUTH"):     directionalSouth,
		directionalToken("E"):         directionalEast,
		directionalToken("EAST"):      directionalEast,
		directionalToken("W"):         directionalWest,
		directionalToken("WEST"):      directionalWest,
		directionalToken("NE"):        directionalNorthEast,
		directionalToken("NORTHEAST"): directionalNorthEast,
		directionalToken("NW"):        directionalNorthWest,
		directionalToken("NORTHWEST"): directionalNorthWest,
		directionalToken("SE"):        directionalSouthEast,
		directionalToken("SOUTHEAST"): directionalSouthEast,
		directionalToken("SW"):        directionalSouthWest,
		directionalToken("SOUTHWEST"): directionalSouthWest,
	}
	streetSuffixMap = map[streetSuffixToken]streetSuffixValue{
		streetSuffixToken("ALLEY"):     streetSuffixAlley,
		streetSuffixToken("AVENUE"):    streetSuffixAvenue,
		streetSuffixToken("BOULEVARD"): streetSuffixBoulevard,
		streetSuffixToken("CIRCLE"):    streetSuffixCircle,
		streetSuffixToken("COURT"):     streetSuffixCourt,
		streetSuffixToken("DRIVE"):     streetSuffixDrive,
		streetSuffixToken("LANE"):      streetSuffixLane,
		streetSuffixToken("PARKWAY"):   streetSuffixParkway,
		streetSuffixToken("PLACE"):     streetSuffixPlace,
		streetSuffixToken("ROAD"):      streetSuffixRoad,
		streetSuffixToken("SQUARE"):    streetSuffixSquare,
		streetSuffixToken("STREET"):    streetSuffixStreet,
		streetSuffixToken("TERRACE"):   streetSuffixTerrace,
		streetSuffixToken("TRAIL"):     streetSuffixTrail,
		streetSuffixToken("WAY"):       streetSuffixWay,
	}
	secondaryMap = map[secondaryDesignatorToken]secondaryDesignatorValue{
		secondaryDesignatorToken("APT"):       secondaryDesignatorApartment,
		secondaryDesignatorToken("APARTMENT"): secondaryDesignatorApartment,
		secondaryDesignatorToken("UNIT"):      secondaryDesignatorUnit,
		secondaryDesignatorToken("STE"):       secondaryDesignatorSuite,
		secondaryDesignatorToken("SUITE"):     secondaryDesignatorSuite,
		secondaryDesignatorToken("ROOM"):      secondaryDesignatorRoom,
		secondaryDesignatorToken("RM"):        secondaryDesignatorRoom,
		secondaryDesignatorToken("FL"):        secondaryDesignatorFloor,
		secondaryDesignatorToken("FLOOR"):     secondaryDesignatorFloor,
		secondaryDesignatorToken("BLDG"):      secondaryDesignatorBuilding,
		secondaryDesignatorToken("BUILDING"):  secondaryDesignatorBuilding,
		secondaryDesignatorToken("LOT"):       secondaryDesignatorLot,
	}
)

func buildSecondaryPatterns(tokens []secondaryDesignatorToken) (*regexp.Regexp, *regexp.Regexp) {
	parts := make([]string, 0, len(tokens)+1)
	for _, token := range tokens {
		parts = append(parts, regexp.QuoteMeta(string(token)))
	}
	// Include literal "#" to support formats like "SUITE #12" and "#7" without adding it to the normalization map.
	parts = append(parts, regexp.QuoteMeta("#"))

	alternation := strings.Join(parts, "|")
	designatorGroup := fmt.Sprintf("(%s)", alternation)
	pattern := fmt.Sprintf(`(?i)\b(?:%s\b[ .\-#]*)(.+)$`, designatorGroup)
	designatorPattern := fmt.Sprintf(`(?i)\b%s\b`, designatorGroup)
	return regexp.MustCompile(pattern), regexp.MustCompile(designatorPattern)
}

func normalizeStreet(segment string) (street string, secondary string, diags []Diagnostic) {
	if segment == "" {
		return "", "", []Diagnostic{{
			Severity: SeverityError,
			Code:     DiagnosticCodeMissingStreet,
			Message:  "street address segment is missing",
		}}
	}

	segmentUpper := strings.ToUpper(segment)

	if matches := poBoxPattern.FindStringSubmatch(segmentUpper); len(matches) == 2 {
		return fmt.Sprintf("PO BOX %s", matches[1]), "", nil
	}

	secondaryParts := make([]string, 0, 1)
	for {
		primary, designator, remainder, ok := splitInlineSecondary(segmentUpper)
		if !ok {
			break
		}
		normalizedDesignator, recognized := normalizeSecondaryDesignator(designator)
		if diag := newUnknownSecondaryDiagnostic(designator, recognized); diag != nil {
			diags = append(diags, *diag)
		}
		part := strings.TrimSpace(normalizedDesignator + " " + remainder)
		if part == "" {
			break
		}
		secondaryParts = append([]string{part}, secondaryParts...)
		segmentUpper = primary
		if segmentUpper == "" {
			break
		}
	}
	secondary = strings.Join(secondaryParts, " ")
	segmentUpper = strings.TrimSpace(segmentUpper)

	parts := strings.Fields(segmentUpper)
	normalizedParts := make([]string, 0, len(parts))
	// Normalize directionals wherever they appear in the address segment.
	// This allows for valid addresses such as "123 North Main Street" or "East 7th Street".
	for i, part := range parts {
		if normalized, ok := directionalMap[directionalToken(part)]; ok {
			normalizedParts = append(normalizedParts, string(normalized))
			continue
		}
		if normalized, ok := streetSuffixMap[streetSuffixToken(part)]; ok && i == len(parts)-1 {
			normalizedParts = append(normalizedParts, string(normalized))
			continue
		}
		normalizedParts = append(normalizedParts, part)
	}

	street = strings.Join(normalizedParts, " ")
	if street == "" {
		diags = append(diags, Diagnostic{
			Severity: SeverityError,
			Code:     DiagnosticCodeEmptyStreet,
			Message:  "could not determine primary street address",
		})
	}
	return street, secondary, diags
}

// normalizeSecondaryDesignator converts secondary address designators (e.g., "Apartment", "Suite")
// to their standard USPS abbreviations. If the designator is recognized, it returns the USPS
// abbreviation; otherwise, it returns the original input.
func normalizeSecondaryDesignator(designator string) (string, bool) {
	designator = strings.ToUpper(strings.TrimSpace(designator))
	if designator == "#" {
		return "#", true
	}
	if mapped, ok := secondaryMap[secondaryDesignatorToken(designator)]; ok {
		return string(mapped), true
	}
	cleaned := strings.ReplaceAll(designator, ".", "")
	if mapped, ok := secondaryMap[secondaryDesignatorToken(cleaned)]; ok {
		return string(mapped), true
	}
	return designator, false
}

func newUnknownSecondaryDiagnostic(designator string, recognized bool) *Diagnostic {
	if recognized {
		return nil
	}
	designator = strings.TrimSpace(strings.ToUpper(designator))
	if designator == "" {
		return nil
	}
	return &Diagnostic{
		Severity: SeverityWarning,
		Code:     DiagnosticCodeUnknownSecondary,
		Message:  fmt.Sprintf("secondary designator %q is not recognized", designator),
	}
}

func normalizeCity(segments []string) (string, []Diagnostic) {
	if len(segments) == 0 {
		return "", []Diagnostic{{
			Severity: SeverityWarning,
			Code:     DiagnosticCodeMissingCity,
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
			Code:     DiagnosticCodeMissingStateZIP,
			Message:  "state and ZIP segment missing",
		})
		return
	}

	segment = strings.ToUpper(strings.TrimSpace(segment))
	matches := stateZipPattern.FindStringSubmatch(segment)
	if len(matches) == 0 {
		hasDigits := strings.IndexFunc(segment, unicode.IsDigit) >= 0
		code := DiagnosticCodeInvalidStateZIP
		message := "expected two-letter state abbreviation followed by ZIP Code"
		if !hasDigits {
			code = DiagnosticCodeMissingStateZIP
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
			Code:     DiagnosticCodeUnknownState,
			Message:  fmt.Sprintf("state abbreviation %q is not recognized by USPS", state),
		})
	}

	zip = matches[2]
	if len(matches) > 3 {
		zip4 = matches[3]
	}
	return
}

// looksLikeStateZip checks if a segment appears to contain state and ZIP code
func looksLikeStateZip(segment string) bool {
	segment = strings.ToUpper(strings.TrimSpace(segment))
	matches := stateZipPattern.FindStringSubmatch(segment)
	return len(matches) > 0
}

var usStateAbbreviations = map[string]struct{}{
	"AL": {}, "AK": {}, "AZ": {}, "AR": {}, "CA": {}, "CO": {}, "CT": {}, "DE": {}, "DC": {}, "FL": {},
	"GA": {}, "HI": {}, "ID": {}, "IL": {}, "IN": {}, "IA": {}, "KS": {}, "KY": {}, "LA": {}, "ME": {},
	"MD": {}, "MA": {}, "MI": {}, "MN": {}, "MS": {}, "MO": {}, "MT": {}, "NE": {}, "NV": {}, "NH": {},
	"NJ": {}, "NM": {}, "NY": {}, "NC": {}, "ND": {}, "OH": {}, "OK": {}, "OR": {}, "PA": {}, "RI": {},
	"SC": {}, "SD": {}, "TN": {}, "TX": {}, "UT": {}, "VT": {}, "VA": {}, "WA": {}, "WV": {}, "WI": {},
	"WY": {}, "PR": {}, "VI": {}, "GU": {}, "AS": {}, "MP": {},
	// Military state codes
	"AA": {}, "AE": {}, "AP": {},
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
