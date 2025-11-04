package parser

// Lexicon contains USPS Publication 28 lookup tables for address components.
// This follows USPS Pub 28 Appendix C for standard abbreviations.
type Lexicon struct {
	streetSuffixes        map[string]string
	directionals          map[string]string
	secondaryDesignators  map[string]string
	states                map[string]string
}

// newLexicon creates and initializes a new Lexicon with USPS standard abbreviations.
func newLexicon() *Lexicon {
	return &Lexicon{
		streetSuffixes:       initStreetSuffixes(),
		directionals:         initDirectionals(),
		secondaryDesignators: initSecondaryDesignators(),
		states:               initStates(),
	}
}

// NormalizeStreetSuffix returns the USPS standard abbreviation for a street suffix.
func (l *Lexicon) NormalizeStreetSuffix(s string) (string, bool) {
	normalized, ok := l.streetSuffixes[s]
	return normalized, ok
}

// NormalizeDirectional returns the USPS standard abbreviation for a directional.
func (l *Lexicon) NormalizeDirectional(s string) (string, bool) {
	normalized, ok := l.directionals[s]
	return normalized, ok
}

// NormalizeSecondaryDesignator returns the USPS standard abbreviation for a secondary unit.
func (l *Lexicon) NormalizeSecondaryDesignator(s string) (string, bool) {
	normalized, ok := l.secondaryDesignators[s]
	return normalized, ok
}

// NormalizeState returns the two-letter state code.
func (l *Lexicon) NormalizeState(s string) (string, bool) {
	normalized, ok := l.states[s]
	return normalized, ok
}

// initStreetSuffixes initializes the street suffix lookup table.
// Based on USPS Pub 28, Appendix C1.
func initStreetSuffixes() map[string]string {
	suffixes := map[string]string{
		// Common street types
		"ALLEY": "ALY", "ALLEE": "ALY", "ALLY": "ALY", "ALY": "ALY",
		"AVENUE": "AVE", "AV": "AVE", "AVEN": "AVE", "AVENU": "AVE", "AVN": "AVE", "AVNUE": "AVE", "AVE": "AVE",
		"BOULEVARD": "BLVD", "BLVD": "BLVD", "BOUL": "BLVD", "BOULV": "BLVD",
		"CIRCLE": "CIR", "CIR": "CIR", "CIRC": "CIR", "CIRCL": "CIR", "CRCL": "CIR", "CRCLE": "CIR",
		"COURT": "CT", "CT": "CT", "CRT": "CT",
		"DRIVE": "DR", "DR": "DR", "DRIV": "DR", "DRV": "DR",
		"EXPRESSWAY": "EXPY", "EXPY": "EXPY", "EXP": "EXPY", "EXPR": "EXPY", "EXPRESS": "EXPY", "EXPW": "EXPY",
		"HIGHWAY": "HWY", "HWY": "HWY", "HIGHWY": "HWY", "HIWAY": "HWY", "HIWY": "HWY", "HWAY": "HWY",
		"LANE": "LN", "LN": "LN", "LANES": "LN",
		"PARKWAY": "PKWY", "PKWY": "PKWY", "PARKWY": "PKWY", "PKY": "PKWY", "PKWAY": "PKWY",
		"PLACE": "PL", "PL": "PL",
		"PLAZA": "PLZ", "PLZ": "PLZ", "PLZA": "PLZ",
		"ROAD": "RD", "RD": "RD",
		"SQUARE": "SQ", "SQ": "SQ", "SQR": "SQ", "SQRE": "SQ", "SQU": "SQ",
		"STREET": "ST", "ST": "ST", "STR": "ST", "STRT": "ST", "STEET": "ST",
		"TERRACE": "TER", "TER": "TER", "TERR": "TER",
		"TRAIL": "TRL", "TRL": "TRL", "TRAILS": "TRL", "TRK": "TRL",
		"TURNPIKE": "TPKE", "TPKE": "TPKE", "TRNPK": "TPKE", "TURNPK": "TPKE",
		"WAY": "WAY",
	}
	return suffixes
}

// initDirectionals initializes the directional lookup table.
// Based on USPS Pub 28, Appendix C2.
func initDirectionals() map[string]string {
	directionals := map[string]string{
		// Cardinal directions
		"NORTH": "N", "N": "N",
		"SOUTH": "S", "S": "S",
		"EAST": "E", "E": "E",
		"WEST": "W", "W": "W",
		// Intercardinal directions
		"NORTHEAST": "NE", "NE": "NE",
		"NORTHWEST": "NW", "NW": "NW",
		"SOUTHEAST": "SE", "SE": "SE",
		"SOUTHWEST": "SW", "SW": "SW",
	}
	return directionals
}

// initSecondaryDesignators initializes the secondary unit designator lookup table.
// Based on USPS Pub 28, Appendix C2.
func initSecondaryDesignators() map[string]string {
	designators := map[string]string{
		"APARTMENT": "APT", "APT": "APT", "APTMT": "APT",
		"BASEMENT": "BSMT", "BSMT": "BSMT",
		"BUILDING": "BLDG", "BLDG": "BLDG", "BLD": "BLDG",
		"DEPARTMENT": "DEPT", "DEPT": "DEPT",
		"FLOOR": "FL", "FL": "FL", "FLR": "FL",
		"FRONT": "FRNT", "FRNT": "FRNT",
		"HANGER": "HNGR", "HNGR": "HNGR",
		"KEY": "KEY",
		"LOBBY": "LBBY", "LBBY": "LBBY",
		"LOT": "LOT",
		"LOWER": "LOWR", "LOWR": "LOWR",
		"OFFICE": "OFC", "OFC": "OFC",
		"PENTHOUSE": "PH", "PH": "PH",
		"PIER": "PIER",
		"REAR": "REAR",
		"ROOM": "RM", "RM": "RM",
		"SIDE": "SIDE",
		"SLIP": "SLIP",
		"SPACE": "SPC", "SPC": "SPC",
		"STOP": "STOP",
		"SUITE": "STE", "STE": "STE", "SUIT": "STE",
		"TRAILER": "TRLR", "TRLR": "TRLR",
		"UNIT": "UNIT",
		"UPPER": "UPPR", "UPPR": "UPPR",
		// Common single letter abbreviations
		"#": "#",
	}
	return designators
}

// initStates initializes the state code lookup table.
// Includes both state codes and full state names.
func initStates() map[string]string {
	states := map[string]string{
		// State codes
		"AL": "AL", "AK": "AK", "AZ": "AZ", "AR": "AR", "CA": "CA",
		"CO": "CO", "CT": "CT", "DE": "DE", "FL": "FL", "GA": "GA",
		"HI": "HI", "ID": "ID", "IL": "IL", "IN": "IN", "IA": "IA",
		"KS": "KS", "KY": "KY", "LA": "LA", "ME": "ME", "MD": "MD",
		"MA": "MA", "MI": "MI", "MN": "MN", "MS": "MS", "MO": "MO",
		"MT": "MT", "NE": "NE", "NV": "NV", "NH": "NH", "NJ": "NJ",
		"NM": "NM", "NY": "NY", "NC": "NC", "ND": "ND", "OH": "OH",
		"OK": "OK", "OR": "OR", "PA": "PA", "RI": "RI", "SC": "SC",
		"SD": "SD", "TN": "TN", "TX": "TX", "UT": "UT", "VT": "VT",
		"VA": "VA", "WA": "WA", "WV": "WV", "WI": "WI", "WY": "WY",
		"DC": "DC",
		// Territories
		"AS": "AS", "GU": "GU", "MP": "MP", "PR": "PR", "VI": "VI",
		// Full state names
		"ALABAMA": "AL", "ALASKA": "AK", "ARIZONA": "AZ", "ARKANSAS": "AR",
		"CALIFORNIA": "CA", "COLORADO": "CO", "CONNECTICUT": "CT", "DELAWARE": "DE",
		"FLORIDA": "FL", "GEORGIA": "GA", "HAWAII": "HI", "IDAHO": "ID",
		"ILLINOIS": "IL", "INDIANA": "IN", "IOWA": "IA", "KANSAS": "KS",
		"KENTUCKY": "KY", "LOUISIANA": "LA", "MAINE": "ME", "MARYLAND": "MD",
		"MASSACHUSETTS": "MA", "MICHIGAN": "MI", "MINNESOTA": "MN", "MISSISSIPPI": "MS",
		"MISSOURI": "MO", "MONTANA": "MT", "NEBRASKA": "NE", "NEVADA": "NV",
		"NEW HAMPSHIRE": "NH", "NEW JERSEY": "NJ", "NEW MEXICO": "NM", "NEW YORK": "NY",
		"NORTH CAROLINA": "NC", "NORTH DAKOTA": "ND", "OHIO": "OH", "OKLAHOMA": "OK",
		"OREGON": "OR", "PENNSYLVANIA": "PA", "RHODE ISLAND": "RI", "SOUTH CAROLINA": "SC",
		"SOUTH DAKOTA": "SD", "TENNESSEE": "TN", "TEXAS": "TX", "UTAH": "UT",
		"VERMONT": "VT", "VIRGINIA": "VA", "WASHINGTON": "WA", "WEST VIRGINIA": "WV",
		"WISCONSIN": "WI", "WYOMING": "WY",
		// District and territories
		"DISTRICT OF COLUMBIA": "DC",
		"AMERICAN SAMOA": "AS", "GUAM": "GU", "NORTHERN MARIANA ISLANDS": "MP",
		"PUERTO RICO": "PR", "VIRGIN ISLANDS": "VI",
	}
	return states
}
