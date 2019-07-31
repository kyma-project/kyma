package model

// Credentials contains OAuth or Basic Auth configuration.
type Credentials struct {
	// OAuth configuration
	Oauth *Oauth
	// BasicAuth configuration
	Basic *Basic
}

// CredentialsWithCSRF contains OAuth, BasicAuth or Certificates configuration along with optional CSRF data.
type CredentialsWithCSRF struct {
	// OAuth configuration
	Oauth *Oauth
	// BasicAuth configuration
	Basic *Basic

	// Optional CSRF Data
	CSRFInfo *CSRFInfo
}

// Oauth contains data for performing Oauth token request
type Oauth struct {
	// URL to OAuth token provider.
	URL string
	// ClientID to use for authentication.
	ClientID string
	// ClientSecret to use for authentication.
	ClientSecret string
}

// Basic contains user and password for Basic Auth
type Basic struct {
	// Username to use for authentication.
	Username string
	// Password to use for authentication.
	Password string
}

// CSRFInfo contains data for performing CSRF token request
type CSRFInfo struct {
	TokenEndpointURL string
}

// RequestParameters contains additional headers and query parameters
type RequestParameters struct {
	// Additional headers
	Headers *map[string][]string `json:"headers"`
	// Additional query parameters
	QueryParameters *map[string][]string `json:"queryParameters"`
}
