package usps

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/my-eq/go-usps/models"
)

// mockTokenProvider is a mock implementation of TokenProvider for testing
type mockTokenProvider struct {
	token string
	err   error
}

func (m *mockTokenProvider) GetToken(ctx context.Context) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	return m.token, nil
}

func TestNewClient(t *testing.T) {
	token := "test-token"
	provider := NewStaticTokenProvider(token)

	client := NewClient(provider)

	if client.baseURL != ProductionBaseURL {
		t.Errorf("Expected base URL %s, got %s", ProductionBaseURL, client.baseURL)
	}

	if client.httpClient.Timeout != DefaultTimeout {
		t.Errorf("Expected timeout %v, got %v", DefaultTimeout, client.httpClient.Timeout)
	}
}

func TestNewTestClient(t *testing.T) {
	token := "test-token"
	provider := NewStaticTokenProvider(token)

	client := NewTestClient(provider)

	if client.baseURL != TestingBaseURL {
		t.Errorf("Expected base URL %s, got %s", TestingBaseURL, client.baseURL)
	}
}

func TestNewClientWithOAuth(t *testing.T) {
	client := NewClientWithOAuth("client-id", "client-secret")

	if client.baseURL != ProductionBaseURL {
		t.Errorf("Expected base URL %s, got %s", ProductionBaseURL, client.baseURL)
	}

	// Verify the token provider is an OAuthTokenProvider
	provider, ok := client.tokenProvider.(*OAuthTokenProvider)
	if !ok {
		t.Fatalf("Expected tokenProvider to be *OAuthTokenProvider, got %T", client.tokenProvider)
	}

	if provider.clientID != "client-id" {
		t.Errorf("Expected clientID 'client-id', got '%s'", provider.clientID)
	}
	if provider.clientSecret != "client-secret" {
		t.Errorf("Expected clientSecret 'client-secret', got '%s'", provider.clientSecret)
	}
}

func TestNewTestClientWithOAuth(t *testing.T) {
	client := NewTestClientWithOAuth("test-client-id", "test-client-secret")

	if client.baseURL != TestingBaseURL {
		t.Errorf("Expected base URL %s, got %s", TestingBaseURL, client.baseURL)
	}

	// Verify the token provider is an OAuthTokenProvider
	provider, ok := client.tokenProvider.(*OAuthTokenProvider)
	if !ok {
		t.Fatalf("Expected tokenProvider to be *OAuthTokenProvider, got %T", client.tokenProvider)
	}

	if provider.clientID != "test-client-id" {
		t.Errorf("Expected clientID 'test-client-id', got '%s'", provider.clientID)
	}
	if provider.oauthClient.baseURL != OAuthTestingBaseURL {
		t.Errorf("Expected OAuth baseURL '%s', got '%s'", OAuthTestingBaseURL, provider.oauthClient.baseURL)
	}
}

func TestClientOptions(t *testing.T) {
	token := "test-token"
	provider := NewStaticTokenProvider(token)
	customURL := "https://custom.example.com"
	customTimeout := 60 * time.Second

	client := NewClient(
		provider,
		WithBaseURL(customURL),
		WithTimeout(customTimeout),
	)

	if client.baseURL != customURL {
		t.Errorf("Expected base URL %s, got %s", customURL, client.baseURL)
	}

	if client.httpClient.Timeout != customTimeout {
		t.Errorf("Expected timeout %v, got %v", customTimeout, client.httpClient.Timeout)
	}
}

func TestGetAddress_Success(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		if r.URL.Path != "/address" {
			t.Errorf("Expected path /address, got %s", r.URL.Path)
		}

		// Verify authorization header
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-token" {
			t.Errorf("Expected Authorization header 'Bearer test-token', got '%s'", auth)
		}

		// Verify query parameters
		query := r.URL.Query()
		if query.Get("streetAddress") != "123 Main St" {
			t.Errorf("Expected streetAddress '123 Main St', got '%s'", query.Get("streetAddress"))
		}
		if query.Get("city") != "New York" {
			t.Errorf("Expected city 'New York', got '%s'", query.Get("city"))
		}
		if query.Get("state") != "NY" {
			t.Errorf("Expected state 'NY', got '%s'", query.Get("state"))
		}

		// Send response
		response := models.AddressResponse{
			Address: &models.DomesticAddress{
				Address: models.Address{
					StreetAddress: "123 MAIN ST",
				},
				City:     "NEW YORK",
				State:    "NY",
				ZIPCode:  "10001",
				ZIPPlus4: stringPtr("1234"),
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create client
	provider := NewStaticTokenProvider("test-token")
	client := NewClient(provider, WithBaseURL(server.URL))

	// Make request
	ctx := context.Background()
	req := &models.AddressRequest{
		StreetAddress: "123 Main St",
		City:          "New York",
		State:         "NY",
	}

	resp, err := client.GetAddress(ctx, req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify response
	if resp.Address == nil {
		t.Fatal("Expected address in response")
	}
	if resp.Address.StreetAddress != "123 MAIN ST" {
		t.Errorf("Expected street address '123 MAIN ST', got '%s'", resp.Address.StreetAddress)
	}
	if resp.Address.City != "NEW YORK" {
		t.Errorf("Expected city 'NEW YORK', got '%s'", resp.Address.City)
	}
	if resp.Address.State != "NY" {
		t.Errorf("Expected state 'NY', got '%s'", resp.Address.State)
	}
	if resp.Address.ZIPCode != "10001" {
		t.Errorf("Expected ZIP code '10001', got '%s'", resp.Address.ZIPCode)
	}
}

func TestGetAddress_Error(t *testing.T) {
	// Create test server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		errResp := models.ErrorMessage{
			APIVersion: "3.0",
			Error: &models.ErrorInfo{
				Code:    "400",
				Message: "Invalid address",
			},
		}
		_ = json.NewEncoder(w).Encode(errResp)
	}))
	defer server.Close()

	// Create client
	provider := NewStaticTokenProvider("test-token")
	client := NewClient(provider, WithBaseURL(server.URL))

	// Make request
	ctx := context.Background()
	req := &models.AddressRequest{
		StreetAddress: "Invalid",
		State:         "XX",
	}

	_, err := client.GetAddress(ctx, req)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	// Verify it's an APIError
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("Expected APIError, got %T", err)
	}
	if apiErr.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, apiErr.StatusCode)
	}
}

func TestGetCityState_Success(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		if r.URL.Path != "/city-state" {
			t.Errorf("Expected path /city-state, got %s", r.URL.Path)
		}

		// Verify query parameters
		query := r.URL.Query()
		if query.Get("ZIPCode") != "10001" {
			t.Errorf("Expected ZIPCode '10001', got '%s'", query.Get("ZIPCode"))
		}

		// Send response
		response := models.CityStateResponse{
			City:    "NEW YORK",
			State:   "NY",
			ZIPCode: "10001",
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create client
	provider := NewStaticTokenProvider("test-token")
	client := NewClient(provider, WithBaseURL(server.URL))

	// Make request
	ctx := context.Background()
	req := &models.CityStateRequest{
		ZIPCode: "10001",
	}

	resp, err := client.GetCityState(ctx, req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify response
	if resp.City != "NEW YORK" {
		t.Errorf("Expected city 'NEW YORK', got '%s'", resp.City)
	}
	if resp.State != "NY" {
		t.Errorf("Expected state 'NY', got '%s'", resp.State)
	}
	if resp.ZIPCode != "10001" {
		t.Errorf("Expected ZIP code '10001', got '%s'", resp.ZIPCode)
	}
}

func TestGetZIPCode_Success(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		if r.URL.Path != "/zipcode" {
			t.Errorf("Expected path /zipcode, got %s", r.URL.Path)
		}

		// Verify query parameters
		query := r.URL.Query()
		if query.Get("streetAddress") != "123 Main St" {
			t.Errorf("Expected streetAddress '123 Main St', got '%s'", query.Get("streetAddress"))
		}
		if query.Get("city") != "New York" {
			t.Errorf("Expected city 'New York', got '%s'", query.Get("city"))
		}
		if query.Get("state") != "NY" {
			t.Errorf("Expected state 'NY', got '%s'", query.Get("state"))
		}

		// Send response
		response := models.ZIPCodeResponse{
			Address: &models.DomesticAddress{
				Address: models.Address{
					StreetAddress: "123 MAIN ST",
				},
				City:     "NEW YORK",
				State:    "NY",
				ZIPCode:  "10001",
				ZIPPlus4: stringPtr("1234"),
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create client
	provider := NewStaticTokenProvider("test-token")
	client := NewClient(provider, WithBaseURL(server.URL))

	// Make request
	ctx := context.Background()
	req := &models.ZIPCodeRequest{
		StreetAddress: "123 Main St",
		City:          "New York",
		State:         "NY",
	}

	resp, err := client.GetZIPCode(ctx, req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify response
	if resp.Address == nil {
		t.Fatal("Expected address in response")
	}
	if resp.Address.ZIPCode != "10001" {
		t.Errorf("Expected ZIP code '10001', got '%s'", resp.Address.ZIPCode)
	}
	if resp.Address.ZIPPlus4 == nil || *resp.Address.ZIPPlus4 != "1234" {
		t.Errorf("Expected ZIP+4 '1234', got %v", resp.Address.ZIPPlus4)
	}
}

func TestStaticTokenProvider(t *testing.T) {
	token := "test-token-123"
	provider := NewStaticTokenProvider(token)

	ctx := context.Background()
	gotToken, err := provider.GetToken(ctx)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if gotToken != token {
		t.Errorf("Expected token '%s', got '%s'", token, gotToken)
	}
}

func TestStaticTokenProvider_EmptyToken(t *testing.T) {
	provider := NewStaticTokenProvider("")

	ctx := context.Background()
	_, err := provider.GetToken(ctx)
	if err == nil {
		t.Fatal("Expected error for empty token, got nil")
	}
}

func TestStructToURLValues(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected map[string]string
	}{
		{
			name: "address request with all fields",
			input: &models.AddressRequest{
				Firm:             "ACME Corp",
				StreetAddress:    "123 Main St",
				SecondaryAddress: "Apt 4B",
				City:             "New York",
				State:            "NY",
				ZIPCode:          "10001",
				ZIPPlus4:         "1234",
			},
			expected: map[string]string{
				"firm":             "ACME Corp",
				"streetAddress":    "123 Main St",
				"secondaryAddress": "Apt 4B",
				"city":             "New York",
				"state":            "NY",
				"ZIPCode":          "10001",
				"ZIPPlus4":         "1234",
			},
		},
		{
			name: "address request with omitted fields",
			input: &models.AddressRequest{
				StreetAddress: "123 Main St",
				State:         "NY",
			},
			expected: map[string]string{
				"streetAddress": "123 Main St",
				"state":         "NY",
			},
		},
		{
			name: "city state request",
			input: &models.CityStateRequest{
				ZIPCode: "10001",
			},
			expected: map[string]string{
				"ZIPCode": "10001",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			values, err := structToURLValues(tt.input)
			if err != nil {
				t.Fatalf("Expected no error, got %v", err)
			}

			for key, expectedValue := range tt.expected {
				gotValue := values.Get(key)
				if gotValue != expectedValue {
					t.Errorf("Expected %s=%s, got %s=%s", key, expectedValue, key, gotValue)
				}
			}

			// Verify no extra keys
			for key := range values {
				if _, ok := tt.expected[key]; !ok {
					t.Errorf("Unexpected key in values: %s", key)
				}
			}
		})
	}
}

func TestAPIError_Error(t *testing.T) {
	err := &APIError{
		StatusCode: 400,
		ErrorMessage: models.ErrorMessage{
			Error: &models.ErrorInfo{
				Message: "Invalid request",
			},
		},
	}

	expected := "USPS API error (status 400): Invalid request"
	if err.Error() != expected {
		t.Errorf("Expected error message '%s', got '%s'", expected, err.Error())
	}
}

func TestGetAddress_WithOptionalFields(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()

		// Verify optional fields are present
		if query.Get("firm") != "ACME Corp" {
			t.Errorf("Expected firm 'ACME Corp', got '%s'", query.Get("firm"))
		}
		if query.Get("secondaryAddress") != "Suite 100" {
			t.Errorf("Expected secondaryAddress 'Suite 100', got '%s'", query.Get("secondaryAddress"))
		}
		if query.Get("urbanization") != "URB LAS GLADIOLAS" {
			t.Errorf("Expected urbanization 'URB LAS GLADIOLAS', got '%s'", query.Get("urbanization"))
		}

		response := models.AddressResponse{
			Firm: "ACME CORP",
			Address: &models.DomesticAddress{
				Address: models.Address{
					StreetAddress:    "123 MAIN ST",
					SecondaryAddress: "STE 100",
				},
				City:         "SAN JUAN",
				State:        "PR",
				ZIPCode:      "00926",
				Urbanization: "URB LAS GLADIOLAS",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	provider := NewStaticTokenProvider("test-token")
	client := NewClient(provider, WithBaseURL(server.URL))

	ctx := context.Background()
	req := &models.AddressRequest{
		Firm:             "ACME Corp",
		StreetAddress:    "123 Main St",
		SecondaryAddress: "Suite 100",
		City:             "San Juan",
		State:            "PR",
		Urbanization:     "URB LAS GLADIOLAS",
	}

	resp, err := client.GetAddress(ctx, req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if resp.Firm != "ACME CORP" {
		t.Errorf("Expected firm 'ACME CORP', got '%s'", resp.Firm)
	}
	if resp.Address.Urbanization != "URB LAS GLADIOLAS" {
		t.Errorf("Expected urbanization 'URB LAS GLADIOLAS', got '%s'", resp.Address.Urbanization)
	}
}

func TestGetAddress_TokenProviderError(t *testing.T) {
	// Create mock provider that returns an error
	provider := &mockTokenProvider{
		err: context.DeadlineExceeded,
	}
	client := NewClient(provider)

	ctx := context.Background()
	req := &models.AddressRequest{
		StreetAddress: "123 Main St",
		State:         "NY",
	}

	_, err := client.GetAddress(ctx, req)
	if err == nil {
		t.Fatal("Expected error from token provider, got nil")
	}
}

func TestGetCityState_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		errResp := models.ErrorMessage{
			Error: &models.ErrorInfo{
				Code:    "404",
				Message: "ZIP code not found",
			},
		}
		_ = json.NewEncoder(w).Encode(errResp)
	}))
	defer server.Close()

	provider := NewStaticTokenProvider("test-token")
	client := NewClient(provider, WithBaseURL(server.URL))

	ctx := context.Background()
	req := &models.CityStateRequest{
		ZIPCode: "99999",
	}

	_, err := client.GetCityState(ctx, req)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("Expected APIError, got %T", err)
	}
	if apiErr.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status code %d, got %d", http.StatusNotFound, apiErr.StatusCode)
	}
}

func TestGetZIPCode_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		errResp := models.ErrorMessage{
			Error: &models.ErrorInfo{
				Code:    "400",
				Message: "Invalid city",
			},
		}
		_ = json.NewEncoder(w).Encode(errResp)
	}))
	defer server.Close()

	provider := NewStaticTokenProvider("test-token")
	client := NewClient(provider, WithBaseURL(server.URL))

	ctx := context.Background()
	req := &models.ZIPCodeRequest{
		StreetAddress: "123 Main St",
		City:          "InvalidCity",
		State:         "NY",
	}

	_, err := client.GetZIPCode(ctx, req)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

func TestAPIError_ErrorNoMessage(t *testing.T) {
	err := &APIError{
		StatusCode:   500,
		ErrorMessage: models.ErrorMessage{},
	}

	expected := "USPS API error (status 500)"
	if err.Error() != expected {
		t.Errorf("Expected error message '%s', got '%s'", expected, err.Error())
	}
}

func TestHandleResponse_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	provider := NewStaticTokenProvider("test-token")
	client := NewClient(provider, WithBaseURL(server.URL))

	ctx := context.Background()
	req := &models.CityStateRequest{
		ZIPCode: "10001",
	}

	_, err := client.GetCityState(ctx, req)
	if err == nil {
		t.Fatal("Expected error for invalid JSON, got nil")
	}
}

func TestHandleResponse_InvalidErrorJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("invalid error json"))
	}))
	defer server.Close()

	provider := NewStaticTokenProvider("test-token")
	client := NewClient(provider, WithBaseURL(server.URL))

	ctx := context.Background()
	req := &models.CityStateRequest{
		ZIPCode: "invalid",
	}

	_, err := client.GetCityState(ctx, req)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	// Should still return an error even if we can't parse the error response
	if apiErr, ok := err.(*APIError); ok {
		t.Errorf("Expected generic error, got APIError: %v", apiErr)
	}
}

func TestWithHTTPClient(t *testing.T) {
	customClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	provider := NewStaticTokenProvider("test-token")
	client := NewClient(provider, WithHTTPClient(customClient))

	if client.httpClient != customClient {
		t.Error("Expected custom HTTP client to be used")
	}
}

func TestGetAddress_AdditionalInfoAndCorrections(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := models.AddressResponse{
			Address: &models.DomesticAddress{
				Address: models.Address{
					StreetAddress: "123 MAIN ST",
				},
				City:     "NEW YORK",
				State:    "NY",
				ZIPCode:  "10001",
				ZIPPlus4: stringPtr("1234"),
			},
			AdditionalInfo: &models.AddressAdditionalInfo{
				DeliveryPoint:        "12",
				CarrierRoute:         "C001",
				DPVConfirmation:      "Y",
				DPVCMRA:              "N",
				Business:             "N",
				CentralDeliveryPoint: "N",
				Vacant:               "N",
			},
			Corrections: []models.AddressCorrection{
				{Code: "32", Text: "Default address"},
			},
			Matches: []models.AddressMatch{
				{Code: "31", Text: "Single Response - exact match"},
			},
			Warnings: []string{"Warning 1", "Warning 2"},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	provider := NewStaticTokenProvider("test-token")
	client := NewClient(provider, WithBaseURL(server.URL))

	ctx := context.Background()
	req := &models.AddressRequest{
		StreetAddress: "123 Main St",
		City:          "New York",
		State:         "NY",
	}

	resp, err := client.GetAddress(ctx, req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if resp.AdditionalInfo == nil {
		t.Fatal("Expected additional info in response")
	}
	if resp.AdditionalInfo.DPVConfirmation != "Y" {
		t.Errorf("Expected DPVConfirmation 'Y', got '%s'", resp.AdditionalInfo.DPVConfirmation)
	}

	if len(resp.Corrections) != 1 {
		t.Fatalf("Expected 1 correction, got %d", len(resp.Corrections))
	}
	if resp.Corrections[0].Code != "32" {
		t.Errorf("Expected correction code '32', got '%s'", resp.Corrections[0].Code)
	}

	if len(resp.Matches) != 1 {
		t.Fatalf("Expected 1 match, got %d", len(resp.Matches))
	}
	if resp.Matches[0].Code != "31" {
		t.Errorf("Expected match code '31', got '%s'", resp.Matches[0].Code)
	}

	if len(resp.Warnings) != 2 {
		t.Fatalf("Expected 2 warnings, got %d", len(resp.Warnings))
	}
}

func TestStructToURLValues_NonStruct(t *testing.T) {
	// Test with non-struct input
	_, err := structToURLValues("not a struct")
	if err == nil {
		t.Fatal("Expected error for non-struct input, got nil")
	}
}

func TestStructToURLValues_PointerField(t *testing.T) {
	// Test with pointer string field
	value := "test-value"
	type TestStruct struct {
		Field *string `url:"field"`
	}

	input := &TestStruct{
		Field: &value,
	}

	values, err := structToURLValues(input)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if values.Get("field") != "test-value" {
		t.Errorf("Expected field='test-value', got '%s'", values.Get("field"))
	}
}

func TestStructToURLValues_DefaultCase(t *testing.T) {
	// Test with unsupported field type (should be skipped)
	type TestStruct struct {
		IntField int    `url:"intfield"`
		StrField string `url:"strfield"`
	}

	input := &TestStruct{
		IntField: 123,
		StrField: "test",
	}

	values, err := structToURLValues(input)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Int field should be skipped (default case in switch)
	if values.Get("intfield") != "" {
		t.Errorf("Expected intfield to be skipped, got '%s'", values.Get("intfield"))
	}

	// String field should be included
	if values.Get("strfield") != "test" {
		t.Errorf("Expected strfield='test', got '%s'", values.Get("strfield"))
	}
}

func TestStructToURLValues_NoTag(t *testing.T) {
	// Test with field that has no url tag
	type TestStruct struct {
		NoTag   string
		WithTag string `url:"withtag"`
		DashTag string `url:"-"`
	}

	input := &TestStruct{
		NoTag:   "notag",
		WithTag: "withtag",
		DashTag: "dashtag",
	}

	values, err := structToURLValues(input)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// NoTag and DashTag should be skipped
	if values.Get("NoTag") != "" {
		t.Errorf("Expected NoTag to be skipped")
	}
	if values.Get("dashtag") != "" {
		t.Errorf("Expected DashTag to be skipped")
	}

	// WithTag should be included
	if values.Get("withtag") != "withtag" {
		t.Errorf("Expected withtag='withtag', got '%s'", values.Get("withtag"))
	}
}

func TestDoRequest_StructToURLValuesError(t *testing.T) {
	// Test with a non-struct param to trigger structToURLValues error
	provider := NewStaticTokenProvider("test-token")
	client := NewClient(provider)

	ctx := context.Background()

	// Pass a non-struct value as queryParams
	_, err := client.doRequest(ctx, http.MethodGet, "/test", "not a struct")
	if err == nil {
		t.Fatal("Expected error from structToURLValues, got nil")
	}
	if !strings.Contains(err.Error(), "failed to encode query parameters") {
		t.Errorf("Expected error about query parameters, got: %v", err)
	}
}

func TestDoRequest_InvalidURL(t *testing.T) {
	// Create a client with an invalid method that would cause NewRequestWithContext to fail
	provider := NewStaticTokenProvider("test-token")
	client := NewClient(provider, WithBaseURL("http://localhost"))

	ctx := context.Background()

	// Use an invalid HTTP method to trigger the error
	_, err := client.doRequest(ctx, "INVALID\nMETHOD", "/test", nil)
	if err == nil {
		t.Fatal("Expected error for invalid request, got nil")
	}
}

func TestDoRequest_HTTPClientError(t *testing.T) {
	// Create a client with a custom HTTP client that will fail
	provider := NewStaticTokenProvider("test-token")

	// Create a custom HTTP client with a transport that always fails
	customClient := &http.Client{
		Transport: &failingTransport{},
	}

	client := NewClient(provider, WithHTTPClient(customClient))

	ctx := context.Background()
	req := &models.CityStateRequest{
		ZIPCode: "10001",
	}

	_, err := client.GetCityState(ctx, req)
	if err == nil {
		t.Fatal("Expected error from HTTP client, got nil")
	}
}

func TestHandleResponse_ReadBodyError(t *testing.T) {
	// Create a response with a failing body reader
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       &failingReader{},
		Header:     http.Header{},
	}

	provider := NewStaticTokenProvider("test-token")
	client := NewClient(provider)

	var result models.CityStateResponse
	err := client.handleResponse(resp, &result)
	if err == nil {
		t.Fatal("Expected error from failing reader, got nil")
	}
}

func TestGetCityState_DoRequestError(t *testing.T) {
	// Test error propagation from doRequest
	provider := &mockTokenProvider{
		err: context.DeadlineExceeded,
	}
	client := NewClient(provider)

	ctx := context.Background()
	req := &models.CityStateRequest{
		ZIPCode: "10001",
	}

	_, err := client.GetCityState(ctx, req)
	if err == nil {
		t.Fatal("Expected error from doRequest, got nil")
	}
}

func TestGetZIPCode_DoRequestError(t *testing.T) {
	// Test error propagation from doRequest
	provider := &mockTokenProvider{
		err: context.DeadlineExceeded,
	}
	client := NewClient(provider)

	ctx := context.Background()
	req := &models.ZIPCodeRequest{
		StreetAddress: "123 Main St",
		City:          "New York",
		State:         "NY",
	}

	_, err := client.GetZIPCode(ctx, req)
	if err == nil {
		t.Fatal("Expected error from doRequest, got nil")
	}
}

// failingTransport is a custom transport that always fails
type failingTransport struct{}

func (t *failingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return nil, context.DeadlineExceeded
}

// failingReader is a custom reader that always fails
type failingReader struct{}

func (r *failingReader) Read(p []byte) (n int, err error) {
	return 0, context.DeadlineExceeded
}

func (r *failingReader) Close() error {
	return nil
}

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}
