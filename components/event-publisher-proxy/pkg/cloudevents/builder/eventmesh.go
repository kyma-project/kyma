package builder

import (
	cev2event "github.com/cloudevents/sdk-go/v2/event"

	"github.com/kyma-project/kyma/components/eventing-controller/logger"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/cleaner"

	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/application"
)

// Perform a compile-time check.
var _ CloudEventBuilder = &EventMeshBuilder{}

func NewEventMeshBuilder(prefix string, eventMeshNamespace string, cleaner cleaner.Cleaner,
	applicationLister *application.Lister, logger *logger.Logger) CloudEventBuilder {
	genericBuilder := GenericBuilder{
		typePrefix:        prefix,
		applicationLister: applicationLister,
		logger:            logger,
		cleaner:           cleaner,
	}

	return &EventMeshBuilder{
		genericBuilder:     &genericBuilder,
		eventMeshNamespace: eventMeshNamespace,
	}
}

func (emb *EventMeshBuilder) Build(event cev2event.Event) (*cev2event.Event, error) {
	ceEvent, err := emb.genericBuilder.Build(event)
	if err != nil {
		return nil, err
	}

	// set eventMesh namespace as event source (required by EventMesh)
	ceEvent.SetSource(emb.eventMeshNamespace)

	return ceEvent, err
}
