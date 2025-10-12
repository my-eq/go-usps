package models

// TokenRequest is the base object for an OAuth token request
type TokenRequest struct {
	GrantType string `json:"grant_type" url:"grant_type"`
	Scope     string `json:"scope,omitempty" url:"scope,omitempty"`
}

// ClientCredentials represents the OAuth Client Credentials request
type ClientCredentials struct {
	GrantType    string `json:"grant_type" url:"grant_type"`
	ClientID     string `json:"client_id" url:"client_id"`
	ClientSecret string `json:"client_secret" url:"client_secret"`
	Scope        string `json:"scope,omitempty" url:"scope,omitempty"`
}

// RefreshTokenCredentials represents the OAuth Refresh Token request
type RefreshTokenCredentials struct {
	GrantType    string `json:"grant_type" url:"grant_type"`
	ClientID     string `json:"client_id" url:"client_id"`
	ClientSecret string `json:"client_secret" url:"client_secret"`
	RefreshToken string `json:"refresh_token" url:"refresh_token"`
	Scope        string `json:"scope,omitempty" url:"scope,omitempty"`
}

// AuthorizationCodeCredentials represents the OAuth Authorization Code request
type AuthorizationCodeCredentials struct {
	GrantType    string `json:"grant_type" url:"grant_type"`
	ClientID     string `json:"client_id" url:"client_id"`
	ClientSecret string `json:"client_secret" url:"client_secret"`
	Code         string `json:"code" url:"code"`
	RedirectURI  string `json:"redirect_uri" url:"redirect_uri"`
	Scope        string `json:"scope,omitempty" url:"scope,omitempty"`
}

// TokenRevokeRequest represents the token revocation request
type TokenRevokeRequest struct {
	Token         string `json:"token" url:"token"`
	TokenTypeHint string `json:"token_type_hint,omitempty" url:"token_type_hint,omitempty"`
}

// AccessTokenResponse is the base object for an OAuth access token response
type AccessTokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	TokenType   string `json:"token_type"`
	Scope       string `json:"scope,omitempty"`
}

// ProviderAccessTokenResponse is the provider-specific access token response
type ProviderAccessTokenResponse struct {
	AccessToken     string `json:"access_token"`
	ExpiresIn       int    `json:"expires_in"`
	TokenType       string `json:"token_type"`
	Scope           string `json:"scope,omitempty"`
	IssuedAt        int64  `json:"issued_at,omitempty"`
	Status          string `json:"status,omitempty"`
	Issuer          string `json:"issuer,omitempty"`
	ClientID        string `json:"client_id,omitempty"`
	ApplicationName string `json:"application_name,omitempty"`
	APIProducts     string `json:"api_products,omitempty"`
	PublicKey       string `json:"public_key,omitempty"`
}

// TokensResponse is the IETF standard tokens response with access and refresh tokens
type TokensResponse struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
	Scope        string `json:"scope,omitempty"`
	RefreshToken string `json:"refresh_token"`
}

// ProviderTokensResponse is the provider-specific tokens response
type ProviderTokensResponse struct {
	AccessToken           string `json:"access_token"`
	ExpiresIn             int    `json:"expires_in"`
	TokenType             string `json:"token_type"`
	Scope                 string `json:"scope,omitempty"`
	RefreshToken          string `json:"refresh_token"`
	RefreshTokenIssuedAt  int64  `json:"refresh_token_issued_at,omitempty"`
	RefreshTokenExpiresIn int    `json:"refresh_token_expires_in,omitempty"`
	RefreshCount          int    `json:"refresh_count,omitempty"`
	RefreshTokenStatus    string `json:"refresh_token_status,omitempty"`
	IssuedAt              int64  `json:"issued_at,omitempty"`
	Status                string `json:"status,omitempty"`
	Issuer                string `json:"issuer,omitempty"`
	ClientID              string `json:"client_id,omitempty"`
	ApplicationName       string `json:"application_name,omitempty"`
	APIProducts           string `json:"api_products,omitempty"`
	PublicKey             string `json:"public_key,omitempty"`
}

// StandardErrorResponse represents the OAuth standard error response
type StandardErrorResponse struct {
	Error            string `json:"error,omitempty"`
	ErrorDescription string `json:"error_description,omitempty"`
	ErrorURI         string `json:"error_uri,omitempty"`
}
