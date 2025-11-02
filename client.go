package usps

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"time"

	"github.com/my-eq/go-usps/models"
)

const (
	// ProductionBaseURL is the base URL for the USPS production API
	ProductionBaseURL = "https://apis.usps.com/addresses/v3"
	// TestingBaseURL is the base URL for the USPS testing API
	TestingBaseURL = "https://apis-tem.usps.com/addresses/v3"
	// DefaultTimeout is the default timeout for HTTP requests
	DefaultTimeout = 30 * time.Second
)

// TokenProvider is an interface for providing OAuth tokens
type TokenProvider interface {
	// GetToken returns the current OAuth token
	GetToken(ctx context.Context) (string, error)
}

// StaticTokenProvider is a simple TokenProvider that returns a fixed token
type StaticTokenProvider struct {
	token string
}

// NewStaticTokenProvider creates a new StaticTokenProvider with the given token
func NewStaticTokenProvider(token string) *StaticTokenProvider {
	return &StaticTokenProvider{token: token}
}

// GetToken returns the static token
func (p *StaticTokenProvider) GetToken(ctx context.Context) (string, error) {
	if p.token == "" {
		return "", fmt.Errorf("token is empty")
	}
	return p.token, nil
}

// Client is the USPS API client
type Client struct {
	baseURL       string
	httpClient    *http.Client
	tokenProvider TokenProvider
}

// Option is a functional option for configuring the Client
type Option func(*Client)

// WithBaseURL sets a custom base URL for the client
func WithBaseURL(baseURL string) Option {
	return func(c *Client) {
		c.baseURL = baseURL
	}
}

// WithHTTPClient sets a custom HTTP client
func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

// WithTimeout sets a custom timeout for the HTTP client
func WithTimeout(timeout time.Duration) Option {
	return func(c *Client) {
		c.httpClient.Timeout = timeout
	}
}

// NewClient creates a new USPS API client
func NewClient(tokenProvider TokenProvider, opts ...Option) *Client {
	c := &Client{
		baseURL:       ProductionBaseURL,
		httpClient:    &http.Client{Timeout: DefaultTimeout},
		tokenProvider: tokenProvider,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// NewTestClient creates a new USPS API client configured for the testing environment
func NewTestClient(tokenProvider TokenProvider, opts ...Option) *Client {
	opts = append([]Option{WithBaseURL(TestingBaseURL)}, opts...)
	return NewClient(tokenProvider, opts...)
}

// NewClientWithOAuth creates a new USPS API client with automatic OAuth token management.
// This is a convenience function that creates an OAuthTokenProvider and a Client in one step.
//
// The OAuth provider automatically handles token acquisition, caching, and refresh.
// Additional OAuth options can be passed to customize the provider behavior.
//
// Example:
//
//	client := usps.NewClientWithOAuth("client-id", "client-secret")
//
// Example with options:
//
//	client := usps.NewClientWithOAuth(
//	    "client-id",
//	    "client-secret",
//	    usps.WithOAuthScopes("addresses tracking"),
//	    usps.WithTokenRefreshBuffer(10 * time.Minute),
//	)
func NewClientWithOAuth(clientID, clientSecret string, opts ...OAuthTokenOption) *Client {
	provider := NewOAuthTokenProvider(clientID, clientSecret, opts...)
	return NewClient(provider)
}

// NewTestClientWithOAuth creates a new USPS API client with automatic OAuth token management
// configured for the testing environment. This is a convenience function that combines
// NewOAuthTestTokenProvider and NewTestClient.
//
// Example:
//
//	client := usps.NewTestClientWithOAuth("test-client-id", "test-client-secret")
func NewTestClientWithOAuth(clientID, clientSecret string, opts ...OAuthTokenOption) *Client {
	provider := NewOAuthTestTokenProvider(clientID, clientSecret, opts...)
	return NewTestClient(provider)
}

// doRequest executes an HTTP request and handles the response
func (c *Client) doRequest(ctx context.Context, method, path string, queryParams interface{}) (*http.Response, error) {
	// Build URL with query parameters
	fullURL := c.baseURL + path
	if queryParams != nil {
		values, err := structToURLValues(queryParams)
		if err != nil {
			return nil, fmt.Errorf("failed to encode query parameters: %w", err)
		}
		if len(values) > 0 {
			fullURL += "?" + values.Encode()
		}
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, method, fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Get token and set authorization header
	token, err := c.tokenProvider.GetToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get token: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	return resp, nil
}

// handleResponse processes the HTTP response and unmarshals it into the target
func (c *Client) handleResponse(resp *http.Response, target interface{}) error {
	defer func() { _ = resp.Body.Close() }()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// Handle error responses
	if resp.StatusCode >= 400 {
		var errMsg models.ErrorMessage
		if err := json.Unmarshal(body, &errMsg); err != nil {
			// If we can't parse the error, return a generic error with status code
			return fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
		}
		return &APIError{
			StatusCode:   resp.StatusCode,
			ErrorMessage: errMsg,
		}
	}

	// Unmarshal success response
	if err := json.Unmarshal(body, target); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return nil
}

// structToURLValues converts a struct to url.Values using struct tags
func structToURLValues(s interface{}) (url.Values, error) {
	values := url.Values{}

	v := reflect.ValueOf(s)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return values, fmt.Errorf("expected struct, got %s", v.Kind())
	}

	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i)

		// Get the url tag
		tag := field.Tag.Get("url")
		if tag == "" || tag == "-" {
			continue
		}

		// Parse tag (format: "name,omitempty")
		parts := strings.Split(tag, ",")
		name := parts[0]
		omitEmpty := len(parts) > 1 && parts[1] == "omitempty"

		// Get string value
		var strValue string
		switch value.Kind() {
		case reflect.String:
			strValue = value.String()
		case reflect.Ptr:
			if !value.IsNil() && value.Elem().Kind() == reflect.String {
				strValue = value.Elem().String()
			}
		default:
			continue
		}

		// Add to values if not empty or omitempty is not set
		if strValue != "" || !omitEmpty {
			if strValue != "" {
				values.Add(name, strValue)
			}
		}
	}

	return values, nil
}

// APIError represents an error returned by the USPS API
type APIError struct {
	StatusCode   int
	ErrorMessage models.ErrorMessage
}

// Error implements the error interface
func (e *APIError) Error() string {
	if e.ErrorMessage.Error != nil && e.ErrorMessage.Error.Message != "" {
		return fmt.Sprintf("USPS API error (status %d): %s", e.StatusCode, e.ErrorMessage.Error.Message)
	}
	return fmt.Sprintf("USPS API error (status %d)", e.StatusCode)
}

// GetAddress standardizes a street address
func (c *Client) GetAddress(ctx context.Context, req *models.AddressRequest) (*models.AddressResponse, error) {
	resp, err := c.doRequest(ctx, http.MethodGet, "/address", req)
	if err != nil {
		return nil, err
	}

	var result models.AddressResponse
	if err := c.handleResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetCityState returns the city and state for a given ZIP code
func (c *Client) GetCityState(ctx context.Context, req *models.CityStateRequest) (*models.CityStateResponse, error) {
	resp, err := c.doRequest(ctx, http.MethodGet, "/city-state", req)
	if err != nil {
		return nil, err
	}

	var result models.CityStateResponse
	if err := c.handleResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetZIPCode returns the ZIP code for a given address
func (c *Client) GetZIPCode(ctx context.Context, req *models.ZIPCodeRequest) (*models.ZIPCodeResponse, error) {
	resp, err := c.doRequest(ctx, http.MethodGet, "/zipcode", req)
	if err != nil {
		return nil, err
	}

	var result models.ZIPCodeResponse
	if err := c.handleResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}
