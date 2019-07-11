package model

import "github.com/kyma-project/kyma/components/application-gateway/pkg/authorization"

// ServiceDefinition is an internal representation of a service.
type ServiceDefinition struct {
	// ID of service
	ID string
	// Name of a service
	Name string
	// Provider of a service
	Provider string
	// Description of a service
	Description string
	// Api of a service
	Api *API
	// Events of a service
	Events *Events
	// Documentation of service
	Documentation []byte
}

// API is an internal representation of a service's API.
type API struct {
	// TargetUrl points to API.
	TargetUrl string
	// Credentials is a credentials of API.
	Credentials *authorization.Credentials
	// Spec contains specification of an API.
	Spec []byte
	// RequestParameters will be used with request send by the Application Gateway
	RequestParameters *authorization.RequestParameters
}

// Events contains specification for events.
type Events struct {
	// Spec contains data of events specification.
	Spec []byte
}
