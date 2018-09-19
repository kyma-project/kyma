package metadata

import "github.com/kyma-project/kyma/components/metadata-service/internal/metadata/serviceapi"

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
	// Long description of a service
	ShortDescription string
	// Labels of a service
	Labels *map[string]string
	// Api of a service
	Api *serviceapi.API
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
