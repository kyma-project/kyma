package model

type Labels map[string]interface{}

// Application contains all associated APIs, and EventAPIs
type Application struct {
	ID                  string
	Name                string
	ProviderDisplayName string
	Description         string
	Labels              Labels
	SystemAuthsIDs      []string
	ApiBundles          []APIBundle
}

type APIBundle struct {
	ID                             string
	Name                           string
	Description                    *string
	InstanceAuthRequestInputSchema *string
	APIDefinitions                 []APIDefinition
	EventDefinitions               []EventAPIDefinition
	DefaultInstanceAuth            *Auth
}

// APIDefinition contains API data such as URL, and credentials
type APIDefinition struct {
	ID          string
	Name        string
	Description string
	TargetUrl   string
	Credentials *Credentials
}

// EventAPIDefinition contains Event API details
type EventAPIDefinition struct {
	ID          string
	Name        string
	Description string
}

// Credentials contains OAuth or BasicAuth configuration along with optional CSRF data.
type Credentials struct {
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

// IsEmpty returns true if additional headers and query parameters contain no data, otherwise false
func (r RequestParameters) IsEmpty() bool {
	return (r.Headers == nil || len(*r.Headers) == 0) && (r.QueryParameters == nil || len(*r.QueryParameters) == 0)
}

// Auth contains authentication data
type Auth struct {
	// Credentials
	Credentials *Credentials
	// Additional request parameters
	RequestParameters *RequestParameters
}
