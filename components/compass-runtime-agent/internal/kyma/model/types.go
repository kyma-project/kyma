package model

type Labels map[string]interface{}

type SpecFormat string

const (
	SpecFormatYAML SpecFormat = "YAML"
	SpecFormatJSON SpecFormat = "JSON"
	SpecFormatXML  SpecFormat = "XML"
)

type EventAPISpecType string

const (
	EventAPISpecTypeAsyncAPI EventAPISpecType = "ASYNC_API"
)

type APISpecType string

const (
	APISpecTypeOdata   APISpecType = "ODATA"
	APISpecTypeOpenAPI APISpecType = "OPEN_API"
)

type DocumentFormat string

const (
	DocumentFormatMarkdown DocumentFormat = "MARKDOWN"
)

// Application contains all associated APIs, EventAPIs and Documents
type Application struct {
	ID                  string
	Name                string
	ProviderDisplayName string
	Description         string
	Labels              Labels
	APIs                []APIDefinitionWithAuth
	EventAPIs           []EventAPIDefinition
	Documents           []Document
	SystemAuthsIDs      []string
}

type APIPackage struct {
	ID                             string
	Name                           string
	Description                    *string
	InstanceAuthRequestInputSchema []byte
	Auth                           *Auth
	APIDefinitions                 []APIDefinitionWithAuth
	EventDefinitions               []EventAPIDefinition
	Documents                      []Document
}

type Auth struct {
	RequestParameters RequestParameters
	Credentials       *Credentials
}

type APIDefinition struct {
	ID          string
	Name        string
	Description string
	TargetUrl   string
	APISpec     *APISpec
}

// APIDefinitionWithAuth contains API data such as URL, credentials and spec
type APIDefinitionWithAuth struct {
	APIDefinition
	Auth
}

// EventAPIDefinition contains Event API details such
type EventAPIDefinition struct {
	ID           string
	Name         string
	Description  string
	EventAPISpec *EventAPISpec
}

// Document contains data of document stored in the Rafter
type Document struct {
	ID            string
	ApplicationID string
	Title         string
	DisplayName   string
	Description   string
	Format        DocumentFormat
	Kind          *string
	Data          []byte
}

// APISpec contains API spec BLOB and its type
type APISpec struct {
	Data   []byte
	Type   APISpecType
	Format SpecFormat
}

// EventAPISpec contains event API spec BLOB and its type
type EventAPISpec struct {
	Data   []byte
	Type   EventAPISpecType
	Format SpecFormat
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
