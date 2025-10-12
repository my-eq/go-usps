package models

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
