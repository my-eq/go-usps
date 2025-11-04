package models

import "strings"

// AddressRequest represents the parameters for the address standardization endpoint.
type AddressRequest struct {
	Firm             string `url:"firm,omitempty"`
	StreetAddress    string `url:"streetAddress"`
	SecondaryAddress string `url:"secondaryAddress,omitempty"`
	City             string `url:"city,omitempty"`
	State            string `url:"state"`
	Urbanization     string `url:"urbanization,omitempty"`
	ZIPCode          string `url:"ZIPCode,omitempty"`
	ZIPPlus4         string `url:"ZIPPlus4,omitempty"`
}

// DeliveryLine returns the delivery line of the address (street + secondary or firm if no street).
// It joins StreetAddress and SecondaryAddress with a space, trimming redundant whitespace.
// If StreetAddress is empty and Firm is present, it returns Firm.
// Returns an empty string if no meaningful delivery information is present.
func (a *AddressRequest) DeliveryLine() string {
	if a == nil {
		return ""
	}

	// Normalize whitespace in street and secondary
	street := strings.TrimSpace(a.StreetAddress)
	secondary := strings.TrimSpace(a.SecondaryAddress)

	// If we have a street address, combine with secondary if present
	if street != "" {
		if secondary != "" {
			return street + " " + secondary
		}
		return street
	}

	// No street address, use firm if available
	return strings.TrimSpace(a.Firm)
}

// LastLine returns the last line of the address (city, state, ZIP+4).
// Format: "CITY, STATE ZIP" or "CITY, STATE ZIP-PLUS4"
// Omits empty components cleanly, never returns dangling punctuation.
// Returns an empty string if no meaningful location information is present.
func (a *AddressRequest) LastLine() string {
	if a == nil {
		return ""
	}

	var parts []string

	city := strings.TrimSpace(a.City)
	state := strings.TrimSpace(a.State)
	zip := strings.TrimSpace(a.ZIPCode)
	zipPlus4 := strings.TrimSpace(a.ZIPPlus4)

	// Build city, state part
	if city != "" && state != "" {
		parts = append(parts, city+", "+state)
	} else if city != "" {
		parts = append(parts, city)
	} else if state != "" {
		parts = append(parts, state)
	}

	// Build ZIP part
	if zip != "" {
		if zipPlus4 != "" {
			parts = append(parts, zip+"-"+zipPlus4)
		} else {
			parts = append(parts, zip)
		}
	}

	return strings.Join(parts, " ")
}

// String returns a single-line string representation of the address.
// Format: "DeliveryLine, LastLine" or just one if the other is empty.
// Returns an empty string if both delivery and last lines are empty.
func (a *AddressRequest) String() string {
	if a == nil {
		return ""
	}

	delivery := a.DeliveryLine()
	last := a.LastLine()

	if delivery != "" && last != "" {
		return delivery + ", " + last
	}
	if delivery != "" {
		return delivery
	}
	return last
}

// Lines returns a two-line slice representation of the address.
// Returns [DeliveryLine, LastLine], omitting empty lines.
// Returns an empty slice if both lines are empty.
func (a *AddressRequest) Lines() []string {
	if a == nil {
		return []string{}
	}

	delivery := a.DeliveryLine()
	last := a.LastLine()

	var lines []string
	if delivery != "" {
		lines = append(lines, delivery)
	}
	if last != "" {
		lines = append(lines, last)
	}

	return lines
}

// CityStateRequest represents the parameters for the city-state lookup endpoint.
type CityStateRequest struct {
	ZIPCode string `url:"ZIPCode"`
}

// ZIPCodeRequest represents the parameters for the ZIP code lookup endpoint.
type ZIPCodeRequest struct {
	Firm             string `url:"firm,omitempty"`
	StreetAddress    string `url:"streetAddress"`
	SecondaryAddress string `url:"secondaryAddress,omitempty"`
	City             string `url:"city"`
	State            string `url:"state"`
	ZIPCode          string `url:"ZIPCode,omitempty"`
	ZIPPlus4         string `url:"ZIPPlus4,omitempty"`
}
