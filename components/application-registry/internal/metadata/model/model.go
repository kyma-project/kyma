package model

import (
	"encoding/json"

	"github.com/kyma-project/kyma/components/application-registry/internal/apperrors"
)

// API is an internal representation of a service's API.
type API struct {
	// TargetUrl points to API.
	TargetUrl string
	// Credentials contains credentials of API.
	Credentials *CredentialsWithCSRF
	// Spec contains specification of an API.
	Spec []byte
	// SpecificationUrl is url from where the specification of an API can be acquired - used if Spec is not defined
	SpecificationUrl string
	// ApiType is a type of and API ex. OData, OpenApi
	ApiType string
	// Additional request parameters
	RequestParameters *RequestParameters
	// Specification Credentials contains credentials for fetching API spec.
	SpecificationCredentials *Credentials
	// Additional request parameters to be used when fetching specification
	SpecificationRequestParameters *RequestParameters
}

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
	// Certificates configuration
	CertificateGen *CertificateGen
	// Optional CSRF Data
	CSRFInfo *CSRFInfo
}

// RequestParameters contains additional headers and query parameters
type RequestParameters struct {
	// Additional headers
	Headers *map[string][]string `json:"headers"`
	// Additional query parameters
	QueryParameters *map[string][]string `json:"queryParameters"`
}

// CSRFInfo contains data for performing CSRF token request
type CSRFInfo struct {
	TokenEndpointURL string
}

// Oauth contains data for performing Oauth token request
type Oauth struct {
	// URL to OAuth token provider.
	URL string
	// ClientID to use for authentication.
	ClientID string
	// ClientSecret to use for authentication.
	ClientSecret string
	// Additional request parameters
	RequestParameters *RequestParameters
}

// Basic contains user and password for Basic Auth
type Basic struct {
	// Username to use for authentication.
	Username string
	// Password to use for authentication.
	Password string
}

// CertificateGen contains common name of the certificate to generate
type CertificateGen struct {
	CommonName  string
	Certificate string
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

const (
	requestParametersHeadersKey         = "headers"
	requestParametersQueryParametersKey = "queryParameters"
)

func MapToRequestParameters(data map[string][]byte) (*RequestParameters, apperrors.AppError) {
	requestParameters := &RequestParameters{}

	headersData := data[requestParametersHeadersKey]
	if headersData != nil {
		var headers = &map[string][]string{}
		err := json.Unmarshal(headersData, headers)
		if err != nil {
			return nil, apperrors.Internal("Failed to unmarshal headers, %v", err)
		}

		requestParameters.Headers = headers
	}

	queryParamsData := data[requestParametersQueryParametersKey]
	if queryParamsData != nil {
		var queryParameters = &map[string][]string{}
		err := json.Unmarshal(queryParamsData, queryParameters)
		if err != nil {
			return nil, apperrors.Internal("Failed to unmarshal query parameters, %v", err)
		}

		requestParameters.QueryParameters = queryParameters
	}

	if requestParameters.Headers == nil && requestParameters.QueryParameters == nil {
		return nil, nil
	}

	return requestParameters, nil
}

func RequestParametersToMap(requestParameters *RequestParameters) (map[string][]byte, apperrors.AppError) {
	data := make(map[string][]byte)
	if requestParameters == nil {
		return map[string][]byte{}, nil
	}
	if requestParameters.Headers != nil {
		headers, err := json.Marshal(requestParameters.Headers)
		if err != nil {
			return map[string][]byte{}, apperrors.Internal("Failed to marshall headers from request parameters: %v", err)
		}
		data[requestParametersHeadersKey] = headers
	}
	if requestParameters.QueryParameters != nil {
		queryParameters, err := json.Marshal(requestParameters.QueryParameters)
		if err != nil {
			return map[string][]byte{}, apperrors.Internal("Failed to marshall query parameters from request parameters: %v", err)
		}
		data[requestParametersQueryParametersKey] = queryParameters
	}
	return data, nil
}
