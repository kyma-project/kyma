package model

// API is an internal representation of a service's API.
type API struct {
	// TargetUrl points to API.
	TargetUrl string
	// Credentials is a credentials of API.
	Credentials *Credentials
	// Spec contains specification of an API.
	Spec []byte
	// SpecificationUrl is url from where the specification of an API can be acquired - used if Spec is not defined
	SpecificationUrl string
	// ApiType is a type of and API ex. OData, OpenApi
	ApiType string
	// Request Parameters
	RequestParameters *RequestParameters
}

// Credentials contains OAuth configuration.
type Credentials struct {
	// Oauth is OAuth configuration.
	Oauth          *Oauth
	Basic          *Basic
	CertificateGen *CertificateGen
}

type RequestParameters struct {
	Headers         *map[string][]string `json:"headers"`
	QueryParameters *map[string][]string `json:"queryParameters"`
}

type CSRFInfo struct {
	TokenEndpointURL string
}

// Oauth contains details of OAuth configuration.
type Oauth struct {
	// URL to OAuth token provider.
	URL string
	// ClientID to use for authentication.
	ClientID string
	// ClientSecret to use for authentication.
	ClientSecret string
	// Optional CSRF Data
	CSRFInfo *CSRFInfo
}

// Basic contains details of Basic configuration.
type Basic struct {
	// Username to use for authentication.
	Username string
	// Password to use for authentication.
	Password string
	// Optional CSRF Data
	CSRFInfo *CSRFInfo
}

// CertificateGen contains common name of the certificate to generate
type CertificateGen struct {
	CommonName  string
	Certificate string
	// Optional CSRF Data
	CSRFInfo *CSRFInfo
}

// ServiceDefinition is an internal representation of a service.
type ServiceDefinition struct {
	// ID of service
	ID string
	// Name of a service
	Name string
	// External identifier of a service
	Identifier string
	// Provider of a service
	Provider string
	// Description of a service
	Description string
	// Short description of a service
	ShortDescription string
	// Labels of a service
	Labels *map[string]string
	// Api of a service
	Api *API
	// Events of a service
	Events *Events
	// Documentation of service
	Documentation []byte
}

// Events contains specification for events.
type Events struct {
	// Spec contains data of events specification.
	Spec []byte
}
