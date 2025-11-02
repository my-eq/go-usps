# go-usps

A lightweight, production-grade Go client library for the USPS Addresses 3.0 REST API and OAuth 2.0 API.

[![Go Reference](https://pkg.go.dev/badge/github.com/my-eq/go-usps.svg)](https://pkg.go.dev/github.com/my-eq/go-usps)
[![Go Report Card](https://goreportcard.com/badge/github.com/my-eq/go-usps)](https://goreportcard.com/report/github.com/my-eq/go-usps)
[![CI](https://github.com/my-eq/go-usps/actions/workflows/ci.yml/badge.svg)](https://github.com/my-eq/go-usps/actions/workflows/ci.yml)
[![CodeQL](https://github.com/my-eq/go-usps/actions/workflows/codeql.yml/badge.svg)](https://github.com/my-eq/go-usps/actions/workflows/codeql.yml)
[![Markdown Lint](https://github.com/my-eq/go-usps/actions/workflows/markdown-lint.yml/badge.svg)](https://github.com/my-eq/go-usps/actions/workflows/markdown-lint.yml)

**Enterprise-grade address validation and standardization for Go applications.**

## Why go-usps?

- **🎯 Complete Coverage** - All USPS Addresses 3.0 and OAuth 2.0 endpoints
- **🔒 Automatic OAuth** - Built-in token management with automatic refresh
- **💪 Strongly Typed** - Full type safety based on OpenAPI specification
- **📦 Zero Dependencies** - Only uses Go standard library
- **🏗️ Production Ready** - Powers critical workflows for millions of users
- **🧪 Fully Tested** - 88%+ test coverage with comprehensive test suite

---

## Quick Start

Get started in 60 seconds:

```bash
go get github.com/my-eq/go-usps
```

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/my-eq/go-usps"
    "github.com/my-eq/go-usps/models"
)

func main() {
    // Create client with automatic OAuth (recommended)
    client := usps.NewClientWithOAuth("your-client-id", "your-client-secret")
    
    // Standardize an address
    req := &models.AddressRequest{
        StreetAddress: "123 Main St",
        City:          "New York",
        State:         "NY",
    }
    
    resp, err := client.GetAddress(context.Background(), req)
    if err != nil {
        log.Fatalf("Error: %v", err)
    }
    
    fmt.Printf("Standardized: %s, %s, %s %s\n",
        resp.Address.StreetAddress,
        resp.Address.City,
        resp.Address.State,
        resp.Address.ZIPCode)
}
```

**Get your credentials:** Register at [USPS Developer Portal](https://developers.usps.com)

---

## Core Concepts

### Understanding USPS Address Validation

The USPS Addresses API provides real-time validation and standardization of US domestic addresses.
It ensures addresses are deliverable, corrects common errors, and enriches data with ZIP+4 codes and
delivery point information.

**Key Benefits:**

- **Reduce returns** - Validate shipping addresses before fulfillment
- **Improve deliverability** - Standardize formats to USPS specifications
- **Save costs** - Catch errors before packages are shipped
- **Enhance data quality** - Fill in missing ZIP codes and abbreviations

### Three Core Endpoints

#### 1. Address Standardization (`GetAddress`)

Validates and standardizes complete addresses. Returns the official USPS format with ZIP+4 codes,
delivery point validation, and carrier route information.

```go
req := &models.AddressRequest{
    StreetAddress:    "123 Main St",
    SecondaryAddress: "Apt 4B",  // Optional
    City:             "New York",
    State:            "NY",
}

resp, err := client.GetAddress(ctx, req)
// Returns: standardized address + ZIP+4 + delivery info
```

**Use when:** You have a complete address and need to validate or standardize it.

#### 2. City/State Lookup (`GetCityState`)

Returns the official city and state names for a given ZIP code.

```go
req := &models.CityStateRequest{
    ZIPCode: "10001",
}

resp, err := client.GetCityState(ctx, req)
// Returns: "NEW YORK, NY"
```

**Use when:** You have a ZIP code and need the corresponding city and state.

#### 3. ZIP Code Lookup (`GetZIPCode`)

Returns the ZIP code and ZIP+4 for a given address.

```go
req := &models.ZIPCodeRequest{
    StreetAddress: "123 Main St",
    City:          "New York",
    State:         "NY",
}

resp, err := client.GetZIPCode(ctx, req)
// Returns: ZIP code + ZIP+4
```

**Use when:** You have an address without a ZIP code or need to find the ZIP+4.

### Authentication

All USPS API requests require OAuth 2.0 authentication. This library handles it automatically.

**Recommended approach** (automatic token management):

```go
client := usps.NewClientWithOAuth("client-id", "client-secret")
// Tokens are automatically acquired and refreshed
```

**Alternative** (manual token provider):

```go
tokenProvider := usps.NewStaticTokenProvider("your-access-token")
client := usps.NewClient(tokenProvider)
```

Tokens expire after 8 hours but are automatically refreshed 5 minutes before expiration
when using `NewClientWithOAuth` or `NewOAuthTokenProvider`.

### Error Handling

The library provides structured errors with detailed information:

```go
resp, err := client.GetAddress(ctx, req)
if err != nil {
    if apiErr, ok := err.(*usps.APIError); ok {
        // API-specific error
        fmt.Printf("API Error: %s\n", apiErr.ErrorMessage.Error.Message)
        for _, detail := range apiErr.ErrorMessage.Error.Errors {
            fmt.Printf("  - %s: %s\n", detail.Title, detail.Detail)
        }
    } else {
        // Network or other error
        fmt.Printf("Error: %v\n", err)
    }
    return
}
```

### Environments

The library supports both production and testing environments:

```go
// Production (default)
client := usps.NewClientWithOAuth(clientID, clientSecret)

// Testing
client := usps.NewTestClientWithOAuth(clientID, clientSecret)
```

---

## Usage Examples

### E-commerce: Validate Checkout Addresses

Prevent shipping errors by validating customer addresses during checkout:

```go
func ValidateShippingAddress(street, city, state, zip string) (*models.AddressResponse, error) {
    client := usps.NewClientWithOAuth(os.Getenv("USPS_CLIENT_ID"), 
                                      os.Getenv("USPS_CLIENT_SECRET"))
    
    req := &models.AddressRequest{
        StreetAddress: street,
        City:          city,
        State:         state,
        ZIPCode:       zip,
    }
    
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    resp, err := client.GetAddress(ctx, req)
    if err != nil {
        if apiErr, ok := err.(*usps.APIError); ok {
            return nil, fmt.Errorf("invalid address: %s", apiErr.ErrorMessage.Error.Message)
        }
        return nil, err
    }
    
    return resp, nil
}
```

### Bulk Address Processing

Process a batch of addresses efficiently with concurrent requests:

```go
func ProcessAddresses(addresses []Address) []Result {
    client := usps.NewClientWithOAuth(clientID, clientSecret)
    
    results := make([]Result, len(addresses))
    var wg sync.WaitGroup
    
    // Process up to 10 addresses concurrently
    semaphore := make(chan struct{}, 10)
    
    for i, addr := range addresses {
        wg.Add(1)
        go func(idx int, address Address) {
            defer wg.Done()
            semaphore <- struct{}{}        // Acquire
            defer func() { <-semaphore }() // Release
            
            req := &models.AddressRequest{
                StreetAddress: address.Street,
                City:          address.City,
                State:         address.State,
            }
            
            resp, err := client.GetAddress(context.Background(), req)
            if err != nil {
                results[idx] = Result{Error: err}
                return
            }
            
            results[idx] = Result{
                Standardized: resp.Address,
                ZIPPlus4:     resp.Address.ZIPPlus4, // *string pointer
            }
        }(i, addr)
    }
    
    wg.Wait()
    return results
}
```

### Auto-complete ZIP Codes

Help users by automatically filling in ZIP codes:

```go
func AutoCompleteZIP(street, city, state string) (string, error) {
    client := usps.NewClientWithOAuth(clientID, clientSecret)
    
    req := &models.ZIPCodeRequest{
        StreetAddress: street,
        City:          city,
        State:         state,
    }
    
    resp, err := client.GetZIPCode(context.Background(), req)
    if err != nil {
        return "", err
    }
    
    // Return ZIP+4 format if available
    if resp.ZIPCode.ZIPPlus4 != nil && *resp.ZIPCode.ZIPPlus4 != "" {
        return fmt.Sprintf("%s-%s", resp.ZIPCode.ZIPCode, *resp.ZIPCode.ZIPPlus4), nil
    }
    
    return resp.ZIPCode.ZIPCode, nil
}
```

### Verify Business Addresses

Check if an address is a business location:

```go
func IsBusinessAddress(address *models.AddressRequest) (bool, error) {
    client := usps.NewClientWithOAuth(clientID, clientSecret)
    
    resp, err := client.GetAddress(context.Background(), address)
    if err != nil {
        return false, err
    }
    
    // Check additional info for business indicator
    if resp.AdditionalInfo != nil {
        return resp.AdditionalInfo.Business == "Y", nil
    }
    
    return false, nil
}
```

### Format Addresses for Mailing

Standardize addresses for mail merge or label printing:

```go
func FormatMailingLabel(address *models.AddressRequest) (string, error) {
    client := usps.NewClientWithOAuth(clientID, clientSecret)
    
    resp, err := client.GetAddress(context.Background(), address)
    if err != nil {
        return "", err
    }
    
    // Build formatted address
    var lines []string
    if resp.Firm != "" {
        lines = append(lines, resp.Firm)
    }
    lines = append(lines, resp.Address.StreetAddress)
    if resp.Address.SecondaryAddress != "" {
        lines = append(lines, resp.Address.SecondaryAddress)
    }
    
    cityLine := fmt.Sprintf("%s, %s %s",
        resp.Address.City,
        resp.Address.State,
        resp.Address.ZIPCode)
    
    if resp.Address.ZIPPlus4 != nil && *resp.Address.ZIPPlus4 != "" {
        cityLine = fmt.Sprintf("%s, %s %s-%s",
            resp.Address.City,
            resp.Address.State,
            resp.Address.ZIPCode,
            *resp.Address.ZIPPlus4)
    }
    
    lines = append(lines, cityLine)
    return strings.Join(lines, "\n"), nil
}
```

---

## Advanced Usage

### Custom Token Provider

Implement the `TokenProvider` interface for advanced authentication scenarios like
credential rotation, vault integration, or custom caching:

```go
import (
    "context"
    "fmt"
    
    vault "github.com/hashicorp/vault/api" // Example: HashiCorp Vault client
)

type TokenProvider interface {
    GetToken(ctx context.Context) (string, error)
}

// Example: Vault-backed token provider
type VaultTokenProvider struct {
    vaultClient *vault.Client
    path        string
}

func (p *VaultTokenProvider) GetToken(ctx context.Context) (string, error) {
    secret, err := p.vaultClient.Logical().Read(p.path)
    if err != nil {
        return "", err
    }
    
    token, ok := secret.Data["usps_token"].(string)
    if !ok {
        return "", fmt.Errorf("token not found in vault")
    }
    
    return token, nil
}

// Use with client
client := usps.NewClient(&VaultTokenProvider{
    vaultClient: vaultClient,
    path:        "secret/data/usps",
})
```

### Custom HTTP Client

Configure timeouts, retries, and transport settings:

```go
httpClient := &http.Client{
    Timeout: 60 * time.Second,
    Transport: &http.Transport{
        MaxIdleConns:        100,
        MaxIdleConnsPerHost: 10,
        IdleConnTimeout:     90 * time.Second,
        TLSHandshakeTimeout: 10 * time.Second,
    },
}

client := usps.NewClient(
    tokenProvider,
    usps.WithHTTPClient(httpClient),
)
```

### Retry Logic with Exponential Backoff

Handle transient failures with intelligent retries:

```go
func GetAddressWithRetry(client *usps.Client, req *models.AddressRequest) (*models.AddressResponse, error) {
    maxRetries := 3
    baseDelay := 1 * time.Second
    
    for attempt := 0; attempt <= maxRetries; attempt++ {
        ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
        defer cancel()
        
        resp, err := client.GetAddress(ctx, req)
        if err == nil {
            return resp, nil
        }
        
        // Check if error is retryable
        if apiErr, ok := err.(*usps.APIError); ok {
            // Don't retry 4xx errors (except 429)
            if apiErr.StatusCode >= 400 && apiErr.StatusCode < 500 && apiErr.StatusCode != 429 {
                return nil, err
            }
        }
        
        if attempt < maxRetries {
            delay := baseDelay * time.Duration(1<<uint(attempt)) // Exponential backoff
            time.Sleep(delay)
        }
    }
    
    return nil, fmt.Errorf("max retries exceeded")
}
```

### Circuit Breaker Pattern

Protect your application from cascading failures:

```go
type CircuitBreaker struct {
    client       *usps.Client
    maxFailures  int
    resetTimeout time.Duration
    
    mu            sync.Mutex
    failures      int
    lastFailTime  time.Time
    state         string // "closed", "open", "half-open"
}

func (cb *CircuitBreaker) GetAddress(ctx context.Context, req *models.AddressRequest) (*models.AddressResponse, error) {
    cb.mu.Lock()
    
    // Check if circuit should be reset
    if cb.state == "open" && time.Since(cb.lastFailTime) > cb.resetTimeout {
        cb.state = "half-open"
        cb.failures = 0
    }
    
    if cb.state == "open" {
        cb.mu.Unlock()
        return nil, fmt.Errorf("circuit breaker is open")
    }
    
    cb.mu.Unlock()
    
    // Attempt the request
    resp, err := cb.client.GetAddress(ctx, req)
    
    cb.mu.Lock()
    defer cb.mu.Unlock()
    
    if err != nil {
        cb.failures++
        cb.lastFailTime = time.Now()
        
        if cb.failures >= cb.maxFailures {
            cb.state = "open"
        }
        
        return nil, err
    }
    
    // Success - reset circuit
    cb.failures = 0
    cb.state = "closed"
    
    return resp, nil
}
```

### Request Middleware

Add logging, metrics, or tracing to all requests:

```go
type InstrumentedClient struct {
    client  *usps.Client
    logger  *log.Logger
    metrics MetricsCollector
}

func (ic *InstrumentedClient) GetAddress(ctx context.Context, req *models.AddressRequest) (*models.AddressResponse, error) {
    start := time.Now()
    
    ic.logger.Printf("GetAddress request: %s, %s, %s", req.StreetAddress, req.City, req.State)
    
    resp, err := ic.client.GetAddress(ctx, req)
    
    duration := time.Since(start)
    ic.metrics.RecordDuration("get_address", duration)
    
    if err != nil {
        ic.metrics.IncrementCounter("get_address_errors")
        ic.logger.Printf("GetAddress error: %v (duration: %v)", err, duration)
        return nil, err
    }
    
    ic.metrics.IncrementCounter("get_address_success")
    ic.logger.Printf("GetAddress success (duration: %v)", duration)
    
    return resp, nil
}
```

### Manual OAuth Management

For complex OAuth flows like authorization code with PKCE:

```go
// Step 1: Obtain authorization code (user redirects to USPS)
authURL := "https://apis.usps.com/oauth2/v3/authorize?" +
    "client_id=your-client-id&" +
    "redirect_uri=https://yourapp.com/callback&" +
    "response_type=code&" +
    "scope=addresses"

// Step 2: Exchange code for tokens
oauthClient := usps.NewOAuthClient()

req := &models.AuthorizationCodeCredentials{
    GrantType:    "authorization_code",
    ClientID:     "your-client-id",
    ClientSecret: "your-client-secret",
    Code:         codeFromCallback,
    RedirectURI:  "https://yourapp.com/callback",
}

result, err := oauthClient.PostToken(context.Background(), req)
if err != nil {
    return err
}

tokens := result.(*models.ProviderTokensResponse)

// Step 3: Use access token
tokenProvider := usps.NewStaticTokenProvider(tokens.AccessToken)
client := usps.NewClient(tokenProvider)

// Step 4: Refresh when needed
refreshReq := &models.RefreshTokenCredentials{
    GrantType:    "refresh_token",
    ClientID:     "your-client-id",
    ClientSecret: "your-client-secret",
    RefreshToken: tokens.RefreshToken,
}

newTokens, err := oauthClient.PostToken(context.Background(), refreshReq)
```

### Testing with Mock Responses

Create a custom token provider for testing:

```go
type MockTokenProvider struct{}

func (m *MockTokenProvider) GetToken(ctx context.Context) (string, error) {
    return "mock-token-for-testing", nil
}

// In tests
func TestAddressValidation(t *testing.T) {
    // Use test environment
    client := usps.NewTestClient(&MockTokenProvider{})
    
    // Your test code here
}
```

---

## Advanced Topics

### Distributed Systems Considerations

#### Service-to-Service Authentication

When running in a distributed environment, centralize OAuth token management:

```go
// Token service that manages tokens for all microservices
type TokenService struct {
    client   *usps.OAuthClient
    clientID string
    secret   string
    
    mu    sync.RWMutex
    token string
    expiry time.Time
}

func (ts *TokenService) GetToken(ctx context.Context) (string, error) {
    ts.mu.RLock()
    if time.Now().Before(ts.expiry.Add(-5 * time.Minute)) {
        token := ts.token
        ts.mu.RUnlock()
        return token, nil
    }
    ts.mu.RUnlock()
    
    // Need to refresh
    ts.mu.Lock()
    defer ts.mu.Unlock()
    
    // Double-check after acquiring write lock
    if time.Now().Before(ts.expiry.Add(-5 * time.Minute)) {
        return ts.token, nil
    }
    
    // Fetch new token
    req := &models.ClientCredentials{
        GrantType:    "client_credentials",
        ClientID:     ts.clientID,
        ClientSecret: ts.secret,
    }
    
    result, err := ts.client.PostToken(ctx, req)
    if err != nil {
        return "", err
    }
    
    resp := result.(*models.ProviderAccessTokenResponse)
    ts.token = resp.AccessToken
    ts.expiry = time.Now().Add(time.Duration(resp.ExpiresIn) * time.Second)
    
    return ts.token, nil
}
```

#### Load Balancing and Failover

Distribute requests across multiple client instances:

```go
import (
    "context"
    "sync/atomic"
    
    "github.com/my-eq/go-usps"
    "github.com/my-eq/go-usps/models"
)

type LoadBalancedClient struct {
    clients []*usps.Client
    idx     uint32
}

func (lbc *LoadBalancedClient) GetAddress(ctx context.Context, req *models.AddressRequest) (*models.AddressResponse, error) {
    // Round-robin selection
    idx := atomic.AddUint32(&lbc.idx, 1)
    client := lbc.clients[idx%uint32(len(lbc.clients))]
    
    return client.GetAddress(ctx, req)
}

// Initialize with multiple clients
func NewLoadBalancedClient(tokenProvider usps.TokenProvider, count int) *LoadBalancedClient {
    clients := make([]*usps.Client, count)
    for i := 0; i < count; i++ {
        clients[i] = usps.NewClient(tokenProvider)
    }
    return &LoadBalancedClient{clients: clients}
}
```

### Caching Strategies

#### In-Memory Cache with TTL

Cache validated addresses to reduce API calls:

```go
type CachedClient struct {
    client *usps.Client
    cache  *sync.Map
    ttl    time.Duration
}

type cacheEntry struct {
    response  *models.AddressResponse
    timestamp time.Time
}

func (cc *CachedClient) GetAddress(ctx context.Context, req *models.AddressRequest) (*models.AddressResponse, error) {
    // Create cache key
    key := fmt.Sprintf("%s|%s|%s", req.StreetAddress, req.City, req.State)
    
    // Check cache
    if val, ok := cc.cache.Load(key); ok {
        entry := val.(cacheEntry)
        if time.Since(entry.timestamp) < cc.ttl {
            return entry.response, nil
        }
        cc.cache.Delete(key) // Expired
    }
    
    // Fetch from API
    resp, err := cc.client.GetAddress(ctx, req)
    if err != nil {
        return nil, err
    }
    
    // Store in cache
    cc.cache.Store(key, cacheEntry{
        response:  resp,
        timestamp: time.Now(),
    })
    
    return resp, nil
}
```

#### Redis-backed Cache

For distributed caching across multiple instances:

```go
import (
    "context"
    "encoding/json"
    "time"
    
    "github.com/go-redis/redis/v8" // Example: go-redis client
    "github.com/my-eq/go-usps"
    "github.com/my-eq/go-usps/models"
)

type RedisCachedClient struct {
    client      *usps.Client
    redisClient *redis.Client
    ttl         time.Duration
}

func (rc *RedisCachedClient) GetAddress(ctx context.Context, req *models.AddressRequest) (*models.AddressResponse, error) {
    key := fmt.Sprintf("usps:address:%s:%s:%s", req.StreetAddress, req.City, req.State)
    
    // Try cache first
    cached, err := rc.redisClient.Get(ctx, key).Result()
    if err == nil {
        var resp models.AddressResponse
        if json.Unmarshal([]byte(cached), &resp) == nil {
            return &resp, nil
        }
    }
    
    // Fetch from API
    resp, err := rc.client.GetAddress(ctx, req)
    if err != nil {
        return nil, err
    }
    
    // Store in Redis
    if data, err := json.Marshal(resp); err == nil {
        rc.redisClient.Set(ctx, key, data, rc.ttl)
    }
    
    return resp, nil
}
```

### Rate Limiting

#### Token Bucket Algorithm

Prevent exceeding USPS API rate limits:

```go
import (
    "context"
    "fmt"
    
    "golang.org/x/time/rate"
    "github.com/my-eq/go-usps"
    "github.com/my-eq/go-usps/models"
)

type RateLimitedClient struct {
    client      *usps.Client
    limiter     *rate.Limiter
}

func NewRateLimitedClient(client *usps.Client, requestsPerSecond int) *RateLimitedClient {
    return &RateLimitedClient{
        client:  client,
        limiter: rate.NewLimiter(rate.Limit(requestsPerSecond), requestsPerSecond),
    }
}

func (rlc *RateLimitedClient) GetAddress(ctx context.Context, req *models.AddressRequest) (*models.AddressResponse, error) {
    // Wait for rate limiter
    if err := rlc.limiter.Wait(ctx); err != nil {
        return nil, fmt.Errorf("rate limit: %w", err)
    }
    
    return rlc.client.GetAddress(ctx, req)
}
```

### Observability

#### Structured Logging

Add comprehensive logging for production debugging:

```go
type ObservableClient struct {
    client *usps.Client
    logger *slog.Logger
}

func (oc *ObservableClient) GetAddress(ctx context.Context, req *models.AddressRequest) (*models.AddressResponse, error) {
    requestID := ctx.Value("request_id")
    
    oc.logger.Info("address_validation_start",
        slog.String("request_id", fmt.Sprint(requestID)),
        slog.String("street", req.StreetAddress),
        slog.String("city", req.City),
        slog.String("state", req.State))
    
    start := time.Now()
    resp, err := oc.client.GetAddress(ctx, req)
    duration := time.Since(start)
    
    if err != nil {
        oc.logger.Error("address_validation_failed",
            slog.String("request_id", fmt.Sprint(requestID)),
            slog.Duration("duration", duration),
            slog.String("error", err.Error()))
        return nil, err
    }
    
    oc.logger.Info("address_validation_success",
        slog.String("request_id", fmt.Sprint(requestID)),
        slog.Duration("duration", duration),
        slog.String("zip", resp.Address.ZIPCode))
    
    return resp, nil
}
```

#### Metrics Collection

Track performance and errors with Prometheus:

```go
import (
    "context"
    "time"
    
    "github.com/prometheus/client_golang/prometheus"
    "github.com/my-eq/go-usps"
    "github.com/my-eq/go-usps/models"
)

var (
    requestDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "usps_request_duration_seconds",
            Help: "Duration of USPS API requests",
        },
        []string{"endpoint", "status"},
    )
    
    requestsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "usps_requests_total",
            Help: "Total number of USPS API requests",
        },
        []string{"endpoint", "status"},
    )
)

func init() {
    prometheus.MustRegister(requestDuration)
    prometheus.MustRegister(requestsTotal)
}

type MetricsClient struct {
    client *usps.Client
}

func (mc *MetricsClient) GetAddress(ctx context.Context, req *models.AddressRequest) (*models.AddressResponse, error) {
    start := time.Now()
    
    resp, err := mc.client.GetAddress(ctx, req)
    
    duration := time.Since(start).Seconds()
    status := "success"
    if err != nil {
        status = "error"
    }
    
    requestDuration.WithLabelValues("get_address", status).Observe(duration)
    requestsTotal.WithLabelValues("get_address", status).Inc()
    
    return resp, err
}
```

### Health Checks

Implement health checks for Kubernetes or load balancers:

```go
func USPSHealthCheck(client *usps.Client) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
        defer cancel()
        
        // Use a known good address for health check
        req := &models.AddressRequest{
            StreetAddress: "475 L'Enfant Plaza SW",
            City:          "Washington",
            State:         "DC",
        }
        
        _, err := client.GetAddress(ctx, req)
        if err != nil {
            w.WriteHeader(http.StatusServiceUnavailable)
            json.NewEncoder(w).Encode(map[string]string{
                "status": "unhealthy",
                "error":  err.Error(),
            })
            return
        }
        
        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(map[string]string{
            "status": "healthy",
        })
    }
}
```

### Production Checklist

When deploying to production, ensure you have:

- **✓ Error handling** - Graceful degradation for API failures
- **✓ Timeouts** - Context timeouts on all requests (5-10 seconds recommended)
- **✓ Retries** - Exponential backoff for transient failures
- **✓ Rate limiting** - Respect USPS API limits to avoid 429 errors
- **✓ Caching** - Cache validated addresses (TTL: 24 hours recommended)
- **✓ Monitoring** - Track success rates, latencies, and error rates
- **✓ Circuit breaker** - Prevent cascading failures
- **✓ Logging** - Structured logs with request IDs
- **✓ Health checks** - Endpoint for load balancer health checks
- **✓ Token rotation** - Automatic OAuth token refresh
- **✓ Secrets management** - Never hardcode credentials

---

## API Reference

### Response Fields

#### AddressResponse

```go
type AddressResponse struct {
    Firm           string                 // Business name
    Address        *DomesticAddress       // Standardized address
    AdditionalInfo *AddressAdditionalInfo // Delivery point, carrier route, etc.
    Corrections    []AddressCorrection    // Suggested improvements
    Matches        []AddressMatch         // Match indicators
    Warnings       []string               // Warnings
}
```

#### DomesticAddress

```go
type DomesticAddress struct {
    StreetAddress    string  // Standardized street address
    SecondaryAddress string  // Apartment, suite, etc.
    City             string  // City name
    State            string  // 2-letter state code
    ZIPCode          string  // 5-digit ZIP code
    ZIPPlus4         *string // 4-digit ZIP+4 extension (pointer, may be nil)
    Urbanization     string  // Urbanization code (Puerto Rico)
}
```

#### AddressAdditionalInfo

Rich delivery metadata returned with validated addresses:

```go
type AddressAdditionalInfo struct {
    DeliveryPoint         string // Unique delivery address identifier (2 digits)
    CarrierRoute          string // Carrier route code (4 characters)
    DPVConfirmation       string // Delivery Point Validation: Y, D, S, or N
    DPVCMRA              string // Commercial Mail Receiving Agency: Y or N
    Business             string // Business address indicator: Y or N
    CentralDeliveryPoint string // Central delivery point: Y or N
    Vacant               string // Vacant address indicator: Y or N
}
```

**DPV Confirmation Codes:**

- `Y` - Address is deliverable
- `D` - Address is deliverable but missing secondary (apt, suite)
- `S` - Address is deliverable to building, but not to specific unit
- `N` - Address is not deliverable

### Configuration Options

#### Client Options

```go
// Custom timeout
client := usps.NewClient(tokenProvider, usps.WithTimeout(60 * time.Second))

// Custom HTTP client
client := usps.NewClient(tokenProvider, usps.WithHTTPClient(httpClient))

// Custom base URL (usually for testing)
client := usps.NewClient(tokenProvider, usps.WithBaseURL("https://custom.url"))
```

#### OAuth Provider Options

```go
// Custom scopes
provider := usps.NewOAuthTokenProvider(
    clientID, 
    clientSecret,
    usps.WithOAuthScopes("addresses tracking labels"),
)

// Custom refresh buffer (default: 5 minutes)
provider := usps.NewOAuthTokenProvider(
    clientID,
    clientSecret,
    usps.WithTokenRefreshBuffer(10 * time.Minute),
)

// Enable refresh tokens
provider := usps.NewOAuthTokenProvider(
    clientID,
    clientSecret,
    usps.WithRefreshTokens(true),
)

// Testing environment
provider := usps.NewOAuthTokenProvider(
    clientID,
    clientSecret,
    usps.WithOAuthEnvironment("testing"),
)
```

### Error Types

#### APIError

Returned for USPS API errors (4xx, 5xx responses):

```go
type APIError struct {
    StatusCode   int
    ErrorMessage *ErrorMessage
}

func (e *APIError) Error() string {
    if e.ErrorMessage != nil && e.ErrorMessage.Error != nil {
        return e.ErrorMessage.Error.Message
    }
    return fmt.Sprintf("API error (status %d)", e.StatusCode)
}
```

#### OAuthError

Returned for OAuth authentication errors:

```go
type OAuthError struct {
    StatusCode   int
    ErrorMessage *OAuthErrorMessage
}
```

Common OAuth error codes:

- `invalid_client` - Invalid client credentials
- `invalid_grant` - Invalid authorization code or refresh token
- `invalid_request` - Malformed request
- `unauthorized_client` - Client not authorized for grant type
- `unsupported_grant_type` - Grant type not supported

---

## Additional Resources

### Official Documentation

- [USPS Addresses API v3](https://developers.usps.com/addressesv3) - Complete API specification
- [USPS OAuth 2.0 API](https://developers.usps.com/oauth2v3) - OAuth authentication guide
- [USPS Developer Portal](https://developers.usps.com) - Register for API credentials

### Go Package Documentation

- [pkg.go.dev](https://pkg.go.dev/github.com/my-eq/go-usps) - Full package documentation
- [GitHub Repository](https://github.com/my-eq/go-usps) - Source code and examples

### Getting Help

- **API Issues** - [USPS API Support](https://emailus.usps.com/s/web-tools-inquiry)
- **Library Issues** - [Open an issue](https://github.com/my-eq/go-usps/issues) on GitHub
- **Questions** - Check [existing issues](https://github.com/my-eq/go-usps/issues) or start a discussion

---

## Development

### Running Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run with race detection
go test -race ./...
```

### Integration Tests

Integration tests require valid USPS credentials:

```bash
export USPS_CLIENT_ID="your-client-id"
export USPS_CLIENT_SECRET="your-client-secret"
go test -v ./... -tags=integration
```

### Linting

```bash
# Go linting
go vet ./...
gofmt -l .

# Markdown linting
npx markdownlint-cli2 "**/*.md"
```

---

## Requirements

- **Go 1.19+** - Uses standard library features from Go 1.19
- **USPS API Credentials** - Obtain from [USPS Developer Portal](https://developers.usps.com)

## API Documentation

For complete API documentation, see:

- [USPS Addresses API Documentation](https://developers.usps.com/addressesv3)
- [USPS OAuth 2.0 API Documentation](https://developers.usps.com/oauth)

## License

[MIT License](LICENSE) - See LICENSE file for details.

---

## Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes with tests
4. Run linting and tests (`go test ./... && go vet ./...`)
5. Commit your changes (`git commit -m 'Add amazing feature'`)
6. Push to the branch (`git push origin feature/amazing-feature`)
7. Open a Pull Request

Please ensure your PR:

- ✓ Includes tests for new functionality
- ✓ Maintains or improves test coverage
- ✓ Follows Go best practices and idiomatic style
- ✓ Updates documentation as needed
- ✓ Passes all CI checks
