package specification

import "github.com/kyma-project/kyma/components/metadata-service/internal/metadata/serviceapi"

// Events contains specification for events.
type Events struct {
	// Spec contains data of events specification.
	Spec []byte
}

type SpecData struct {
	Id         string
	API        *serviceapi.API
	Events     *Events
	GatewayUrl string
	Docs       []byte
}
