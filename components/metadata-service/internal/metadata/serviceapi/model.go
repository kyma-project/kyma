package serviceapi

// API is an internal representation of a service's API.
type API struct {
	// TargetUrl points to API.
	TargetUrl string
	// Credentials is a credentials of API.
	Credentials *Credentials
	// Spec contains specification of an API.
	Spec []byte
}

// Credentials contains OAuth configuration.
type Credentials struct {
	// Oauth is OAuth configuration.
	Oauth Oauth
	Basic Basic
}

// Oauth contains details of OAuth configuration.
type Oauth struct {
	// URL to OAuth token provider.
	URL string
	// ClientID to use for authentication.
	ClientID string
	// ClientSecret to use for authentication.
	ClientSecret string
}

// Basic contains details of Basic configuration.
type Basic struct {
	// Username to use for authentication.
	Username string
	// Password to use for authentication.
	Password string
}
